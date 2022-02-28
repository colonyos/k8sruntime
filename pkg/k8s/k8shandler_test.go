package k8s

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/kolony/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestParseCmd(t *testing.T) {
	handler, err := CreateK8sHandler("test")
	assert.Nil(t, err)

	cmdArrStr := handler.parseCmd("go run solver.go")
	assert.Equal(t, "[\"go\", \"run\", \"solver.go\"]", cmdArrStr)

	cmdArrStr = handler.parseCmd("go")
	assert.Equal(t, "[\"go\"]", cmdArrStr)
}

func TestDeployContainer(t *testing.T) {
	client := client.CreateColoniesClient(test.ColoniesServerHost, test.ColoniesServerPort, true)
	colonyID, colonyPrvKey := test.CreateColony(t, client)

	handler, err := CreateK8sHandler("test")
	assert.Nil(t, err)

	runtimeType := "fibonacci_solver"
	name := "fibonacci"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 1
	mem := 1024
	gpu := ""
	gpus := 0

	containerImage := "colonyos/fibonacci"
	cmdStr := "go"
	args := []string{"run", "solver.go"}

	// Register a new runtime
	crypto := crypto.CreateCrypto()
	runtimePrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	runtimeID, err := crypto.GenerateID(runtimePrvKey)
	assert.Nil(t, err)

	runtime := core.CreateRuntime(runtimeID, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus)
	_, err = client.AddRuntime(runtime, colonyPrvKey)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime.ID, colonyPrvKey)
	assert.Nil(t, err)

	// Deploy container to K8s
	yaml := handler.ComposeDeployment(name, containerImage, cmdStr, args, colonyID, cores, mem, gpus, runtimePrvKey, test.ColoniesServerHost, strconv.Itoa(test.ColoniesServerPort))
	fmt.Println(yaml)

	err = handler.CreateDeployment(yaml)
	assert.Nil(t, err)

	names, err := handler.GetDeployments()
	assert.Nil(t, err)
	assert.Equal(t, names[0], "fibonacci-deployment")

	time.Sleep(5 * time.Second)

	err = handler.DeleteDeployment("fibonacci-deployment", false)
	assert.Nil(t, err)

	err = client.DeleteRuntime(runtime.ID, colonyPrvKey)
	assert.Nil(t, err)

	test.DeleteColony(t, client, colonyID)
}
