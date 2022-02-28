package test

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/stretchr/testify/assert"
)

const serverPrvKey = "09545df1812e252a2a853cca29d7eace4a3fe2baad334e3b7141a98d43c31e7b"
const ColoniesServerHost = "10.0.0.240"
const ColoniesServerPort = 50080

func CreateColony(t *testing.T, client *client.ColoniesClient) (string, string) {
	crypto := crypto.CreateCrypto()
	colonyPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID, err := crypto.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	return colonyID, colonyPrvKey
}

func CreateRuntime(t *testing.T, client *client.ColoniesClient, colonyID string, colonyPrvKey string) (string, string) {
	crypto := crypto.CreateCrypto()

	runtimePrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	runtimeID, err := crypto.GenerateID(runtimePrvKey)
	assert.Nil(t, err)

	runtimeType := "test_runtime_type"
	name := "test_runtime"
	cpu := ""
	cores := 0
	mem := 0
	gpu := ""
	gpus := 0

	runtime := core.CreateRuntime(runtimeID, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus, time.Now(), time.Now())
	addedRuntime, err := client.AddRuntime(runtime, colonyPrvKey)
	assert.Nil(t, err)
	assert.True(t, runtime.Equals(addedRuntime))
	err = client.ApproveRuntime(runtime.ID, colonyPrvKey)
	assert.Nil(t, err)

	return runtimeID, runtimePrvKey
}

func DeleteColony(t *testing.T, client *client.ColoniesClient, colonyID string) {
	err := client.DeleteColony(colonyID, serverPrvKey)
	assert.Nil(t, err)
}
