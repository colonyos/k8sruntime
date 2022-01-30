package colony

import (
	"fmt"

	"github.com/colonyos/colonies/pkg/client"
)

type K8sColonyRuntime struct {
}

func CreateK8sColonyRuntime() *K8sColonyRuntime {
	runtime := &K8sColonyRuntime{}

	client := client.CreateColoniesClient("localhost", 8080, true)
	fmt.Println(client)
	return runtime
}
