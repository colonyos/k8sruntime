package k8s

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
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
	handler, err := CreateK8sHandler("test")
	assert.Nil(t, err)

	coloniesServerHost := "10.0.0.240"
	coloniesServerPort := "8080"
	port, err := strconv.Atoi(coloniesServerPort)
	assert.Nil(t, err)

	client := client.CreateColoniesClient(coloniesServerHost, port, true)

	runtimeType := "fibonacci_solver"
	name := "fibonacci"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 1
	mem := 1024
	gpu := ""
	gpus := 0

	containerImage := "johan/fibonacci"
	cmdStr := "go run solver.go"
	colonyID := "6007729ab9a8985b3a3d2da67f255ba13632c4670fe5c218981d77c55f7b3cab"
	colonyPrvKey := "67590823e9a5745ad2aa3acd0038b691bb2e8fa01fb6ad9594020d696e5b9eaf"

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
	yaml := handler.ComposeDeployment(name, containerImage, cmdStr, colonyID, runtimePrvKey, coloniesServerHost, coloniesServerPort)
	fmt.Println(yaml)

	err = handler.CreateDeployment(yaml)
	assert.Nil(t, err)

	names, err := handler.GetDeployments()
	assert.Nil(t, err)
	assert.Equal(t, names[0], "fibonacci-deployment")

	time.Sleep(5 * time.Second)

	err = handler.DeleteDeployment("fibonacci-deployment", false)
	assert.Nil(t, err)
}
