package colony

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/kolony/pkg/k8s"
)

type KubeColonyRT struct {
	coloniesServerHost string
	coloniesServerPort int
	client             *client.ColoniesClient
	colonyID           string
	colonyPrvKey       string
	targetColonyID     string
	targetColonyPrvKey string
	name               string
	runtimeID          string
	runtimePrvKey      string
	stop               chan bool
	namespace          string
	k8sHandler         *k8s.K8sHandler
}

func CreateKubeColonyRT(name,
	coloniesServerHost string,
	coloniesServerPort int,
	colonyID string,
	colonyPrvKey string,
	targetColonyID string,
	targetColonyPrvKey string,
	namespace string) (*KubeColonyRT, error) {

	kubeCRT := &KubeColonyRT{}
	kubeCRT.coloniesServerHost = coloniesServerHost
	kubeCRT.coloniesServerPort = coloniesServerPort
	kubeCRT.client = client.CreateColoniesClient(coloniesServerHost, coloniesServerPort, true) // TODO: insecure
	kubeCRT.colonyID = colonyID
	kubeCRT.colonyPrvKey = colonyPrvKey
	kubeCRT.targetColonyID = targetColonyID
	kubeCRT.targetColonyPrvKey = targetColonyPrvKey
	kubeCRT.name = name
	kubeCRT.stop = make(chan bool, 1)
	kubeCRT.namespace = namespace

	k8sHandler, err := k8s.CreateK8sHandler(namespace)
	if err != nil {
		return kubeCRT, err
	}
	kubeCRT.k8sHandler = k8sHandler

	runtimeID, runtimePrvKey, err := kubeCRT.registerRuntime(kubeCRT.name, "kube_runtime", kubeCRT.colonyID, kubeCRT.colonyPrvKey, 0, 0, 0)
	if err != nil {
		return kubeCRT, nil
	}
	kubeCRT.runtimeID = runtimeID
	kubeCRT.runtimePrvKey = runtimePrvKey

	return kubeCRT, nil
}

func (kubeCRT *KubeColonyRT) ServeForEver() error {
	for {
		assignedProcess, err := kubeCRT.client.AssignProcess(kubeCRT.colonyID, kubeCRT.runtimePrvKey)

		if err == nil {
			err := kubeCRT.deploy(assignedProcess)
			if err != nil {
				fmt.Println(err)
			}
		}

		select {
		case stopNow := <-kubeCRT.stop:
			if stopNow {
				return nil
			}
		default:
			time.Sleep(1000 * time.Millisecond)
		}
	}
}

func (kubeCRT *KubeColonyRT) deploy(process *core.Process) error {
	name := "temp_name" // XXX TODO
	runtimeID, runtimePrvKey, err := kubeCRT.registerRuntime(name, name, kubeCRT.targetColonyID, kubeCRT.targetColonyPrvKey, process.ProcessSpec.Conditions.Cores, process.ProcessSpec.Conditions.Mem, process.ProcessSpec.Conditions.GPUs)
	if err != nil {
		return err
	}

	yaml := kubeCRT.k8sHandler.ComposeDeployment(runtimeID[0:15]+"-"+name, process.ProcessSpec.Image, process.ProcessSpec.Cmd, process.ProcessSpec.Args, kubeCRT.targetColonyID, process.ProcessSpec.Conditions.Cores, process.ProcessSpec.Conditions.Mem, process.ProcessSpec.Conditions.GPUs, runtimePrvKey, kubeCRT.coloniesServerHost, strconv.Itoa(kubeCRT.coloniesServerPort))

	err = kubeCRT.k8sHandler.CreateDeployment(yaml)
	if err != nil {
		return err
	}

	kubeCRT.client.CloseSuccessful(process.ID, kubeCRT.runtimePrvKey)
	return nil
}

func (kubeCRT *KubeColonyRT) registerRuntime(name, runtimeType string, colonyID string, colonyPrvKey string, cores int, mem int, gpus int) (string, string, error) {
	crypto := crypto.CreateCrypto()
	runtimePrvKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return "", "", err
	}

	runtimeID, err := crypto.GenerateID(runtimePrvKey)
	if err != nil {
		return "", "", err
	}

	cpu := ""
	gpu := ""

	runtime := core.CreateRuntime(runtimeID, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus, time.Now(), time.Now())

	_, err = kubeCRT.client.AddRuntime(runtime, colonyPrvKey)
	if err != nil {
		return runtimeID, runtimePrvKey, err
	}

	err = kubeCRT.client.ApproveRuntime(runtime.ID, colonyPrvKey)
	if err != nil {
		return runtimeID, runtimePrvKey, err
	}

	return runtimeID, runtimePrvKey, nil
}

func (kubeCRT *KubeColonyRT) unregisterRuntime() error {
	return kubeCRT.client.DeleteRuntime(kubeCRT.runtimeID, kubeCRT.colonyPrvKey)
}

func (kubeCRT *KubeColonyRT) isRegistered() (bool, error) {
	runtimes, err := kubeCRT.client.GetRuntimes(kubeCRT.colonyID, kubeCRT.colonyPrvKey)
	if err != nil {
		return false, err
	}

	for _, runtime := range runtimes {
		if runtime.ID == kubeCRT.runtimeID {
			return true, nil
		}
	}

	return false, nil
}

func (kubeCRT *KubeColonyRT) RemoveAllDeployments() error {
	deploymentNames, err := kubeCRT.k8sHandler.GetDeployments()
	if err != nil {
		return err
	}

	runtimes, err := kubeCRT.client.GetRuntimes(kubeCRT.targetColonyID, kubeCRT.targetColonyPrvKey)

	for _, deploymentName := range deploymentNames {
		s := strings.Split(deploymentName, "-")
		name := s[0]
		for _, runtime := range runtimes {
			if runtime.ID[0:len(name)] == name {
				err := kubeCRT.k8sHandler.DeleteDeployment(deploymentName, false)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (kubeCRT *KubeColonyRT) Destroy() error {
	kubeCRT.stop <- true
	return kubeCRT.unregisterRuntime()
}
