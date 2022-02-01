package test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/stretchr/testify/assert"
)

const serverPrvKey = "09545df1812e252a2a853cca29d7eace4a3fe2baad334e3b7141a98d43c31e7b"
const ColoniesServerHost = "localhost"
const ColoniesServerPort = 8080

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

func DeleteColony(t *testing.T, client *client.ColoniesClient, colonyID string) {
	err := client.DeleteColony(colonyID, serverPrvKey)
	assert.Nil(t, err)
}
