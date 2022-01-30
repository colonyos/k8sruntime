package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeployContainer(t *testing.T) {
	handler, err := CreateK8sHandler("test")
	assert.Nil(t, err)

	yaml := handler.ComposeDeployment()

	err = handler.CreateDeployment(yaml)
	assert.Nil(t, err)

	err = handler.DeleteDeployment("", false)
	assert.Nil(t, err)
}
