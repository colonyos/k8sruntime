package colony

import (
	"fmt"
	"testing"
)

func TestCreateK8sColonyRuntime(t *testing.T) {
	runtime := CreateK8sColonyRuntime()
	fmt.Println(runtime)
}
