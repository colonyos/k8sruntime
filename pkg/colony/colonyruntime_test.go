package colony

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/kolony/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRuntime(t *testing.T) {
	client := client.CreateColoniesClient(test.ColoniesServerHost, test.ColoniesServerPort, true)
	colonyID, colonyPrvKey := test.CreateColony(t, client)
	targetColonyID, targetColonyPrvKey := test.CreateColony(t, client)

	kubeCRT, err := CreateKubeColonyRT("test", test.ColoniesServerHost, test.ColoniesServerPort, colonyID, colonyPrvKey, targetColonyID, targetColonyPrvKey, "test")
	assert.Nil(t, err)

	registered, err := kubeCRT.isRegistered()
	assert.Nil(t, err)
	assert.True(t, registered)

	err = kubeCRT.Destroy()
	assert.Nil(t, err)

	registered, err = kubeCRT.isRegistered()
	assert.Nil(t, err)
	assert.False(t, registered)

	test.DeleteColony(t, client, colonyID)
	test.DeleteColony(t, client, targetColonyID)
}

func TestServe(t *testing.T) {
	client := client.CreateColoniesClient(test.ColoniesServerHost, test.ColoniesServerPort, true)
	colonyID, colonyPrvKey := test.CreateColony(t, client)
	targetColonyID, targetColonyPrvKey := test.CreateColony(t, client)
	_, runtimePrvKey := test.CreateRuntime(t, client, colonyID, colonyPrvKey)

	kubeCRT, err := CreateKubeColonyRT("test", test.ColoniesServerHost, test.ColoniesServerPort, colonyID, colonyPrvKey, targetColonyID, targetColonyPrvKey, "test")
	assert.Nil(t, err)

	go func() {
		err := kubeCRT.ServeForEver()
		assert.Nil(t, err)
	}()

	json := `
{
    "conditions": {
        "runtimetype": "kube_runtime"
    },
    "env": {
        "name": "fibonacci",
        "container_image": "johan/fibonacci",
        "cmd": "go run solver.go",
		"cores": "2",
		"mem": "3000",
		"gpus": "0"
    }
}
`
	processSpec, err := core.ConvertJSONToProcessSpec(json)
	assert.Nil(t, err)

	processSpec.Conditions.ColonyID = colonyID

	_, err = client.SubmitProcessSpec(processSpec, runtimePrvKey)
	assert.Nil(t, err)

	time.Sleep(2000 * time.Millisecond)

	err = kubeCRT.RemoveAllDeployments()
	time.Sleep(2000 * time.Millisecond)
	assert.Nil(t, err)

	err = kubeCRT.Destroy()
	assert.Nil(t, err)

	test.DeleteColony(t, client, colonyID)
	test.DeleteColony(t, client, targetColonyID)
}
