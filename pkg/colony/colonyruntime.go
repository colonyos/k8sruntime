package colony

import (
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
)

type KubeColonyRT struct {
	client       *client.ColoniesClient
	colonyID     string
	colonyPrvKey string
	name         string
}

func CreateKubeColonyRT(name, coloniesServerHost string, coloniesServerPort int, colonyID string, colonyPrvKey string) *KubeColonyRT {
	kubeCRT := &KubeColonyRT{}
	kubeCRT.client = client.CreateColoniesClient("localhost", 8080, true) // TODO: insecure
	kubeCRT.colonyID = colonyID
	kubeCRT.colonyPrvKey = colonyPrvKey
	kubeCRT.name = name

	return kubeCRT
}

func (kubeCRT *KubeColonyRT) registerRuntime() (string, error) {
	crypto := crypto.CreateCrypto()
	runtimePrvKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return "", err
	}

	runtimeID, err := crypto.GenerateID(runtimePrvKey)
	if err != nil {
		return "", err
	}

	runtimeType := "kube_runtime"
	cpu := ""
	cores := 0
	mem := 0
	gpu := ""
	gpus := 0

	runtime := core.CreateRuntime(runtimeID, runtimeType, kubeCRT.name, kubeCRT.colonyID, cpu, cores, mem, gpu, gpus)

	_, err = kubeCRT.client.AddRuntime(runtime, kubeCRT.colonyPrvKey)
	if err != nil {
		return runtimeID, err
	}

	err = kubeCRT.client.ApproveRuntime(runtime.ID, kubeCRT.colonyPrvKey)
	if err != nil {
		return runtimeID, err
	}

	return runtimeID, nil
}

func (kubeCRT *KubeColonyRT) unregisterRuntime(runtimeID string) error {
	return kubeCRT.client.DeleteRuntime(runtimeID, kubeCRT.colonyPrvKey)
}

func (kubeCRT *KubeColonyRT) nrOfRuntimes() (int, error) {
	runtimes, err := kubeCRT.client.GetRuntimes(kubeCRT.colonyID, kubeCRT.colonyPrvKey)
	if err != nil {
		return -1, err
	}

	return len(runtimes), nil
}
