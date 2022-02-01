package colony

import (
	"testing"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/k8s/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRuntime(t *testing.T) {
	client := client.CreateColoniesClient(test.ColoniesServerHost, test.ColoniesServerPort, true)
	colonyID, colonyPrvKey := test.CreateColony(t, client)

	kubeCRT := CreateKubeColonyRT("test", test.ColoniesServerHost, test.ColoniesServerPort, colonyID, colonyPrvKey)
	runtimeID, err := kubeCRT.registerRuntime()
	assert.Nil(t, err)

	count, err := kubeCRT.nrOfRuntimes()
	assert.Nil(t, err)
	assert.Equal(t, count, 1)

	err = kubeCRT.unregisterRuntime(runtimeID)
	assert.Nil(t, err)

	count, err = kubeCRT.nrOfRuntimes()
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	test.DeleteColony(t, client, colonyID)
}
