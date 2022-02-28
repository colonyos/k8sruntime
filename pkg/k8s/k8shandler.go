package k8s

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type K8sHandler struct {
	client    dynamic.Interface
	clientset *kubernetes.Clientset
	namespace string
}

type ContainerSpec struct {
	Args           []string
	Name           string
	ContainerImage string
}

func CreateK8sHandler(namespace string) (*K8sHandler, error) {
	handler := &K8sHandler{}
	handler.namespace = namespace

	var err error
	handler.client, handler.clientset, err = handler.setupK8sClient()
	if err != nil {
		return nil, err
	}

	return handler, nil
}

func (handler *K8sHandler) setupK8sClient() (dynamic.Interface, *kubernetes.Clientset, error) {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)

	return client, clientset, nil
}

func (handler K8sHandler) parseCmd(cmdStr string) string {
	cmdArr := strings.Split(cmdStr, " ")

	cmdArrStr := "["
	start := true
	for _, c := range cmdArr {
		if start {
			cmdArrStr += "\"" + c + "\","
			start = false
		} else {
			cmdArrStr += " \"" + c + "\","
		}
	}
	cmdArrStr = cmdArrStr[:len(cmdArrStr)-1]
	cmdArrStr += "]"

	return cmdArrStr
}

func (handler K8sHandler) arrayToString(array []string) string {
	str := "["
	for _, a := range array {
		str += "\"" + a + "\", "
	}
	str = str[:len(str)-2]
	str += "]"
	return str
}

func (handler K8sHandler) ComposeDeployment(name string, containerImage string, cmdStr string, args []string, colonyID string, cores int, mem int, gpu int, runtimePrvKey string, coloniesServerHost string, coloniesServerPort string) string {
	cpuStr := strconv.Itoa(1000*cores) + "m"
	memStr := strconv.Itoa(mem) + "Mi"

	var command []string
	command = append(command, cmdStr)
	command = append(command, args...)
	commandStr := handler.arrayToString(command)

	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ` + name + `-deployment
  labels:
    app: ` + name + `
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ` + name + `
  template:
    metadata:
      labels:
        app: ` + name + `
    spec:
      containers:
      - name: ` + name + `
        image: ` + containerImage + `
        resources:
          requests:
            memory: "` + memStr + `"
            cpu: "` + cpuStr + `"
          limits:
            memory: "` + memStr + `" 
            cpu: "` + cpuStr + `"
        command: ` + commandStr + `
        env:
        - name: COLONYID
          value: "` + colonyID + `" 
        - name: RUNTIME_PRVKEY
          value: "` + runtimePrvKey + `"
        - name: COLONIES_SERVER_HOST
          value: "` + coloniesServerHost + `"
        - name: COLONIES_SERVER_PORT
          value: "` + coloniesServerPort + `"
`
	return yaml
}

func (handler K8sHandler) CreateDeployment(yamlFile string) error {
	deployment := &unstructured.Unstructured{}

	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err := dec.Decode([]byte(yamlFile), nil, deployment)
	if err != nil {
		return err
	}

	resource := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	_, err = handler.client.Resource(resource).Namespace(handler.namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (handler K8sHandler) DeleteDeployment(deploymentName string, force bool) error {
	var deleteOptions metav1.DeleteOptions
	deletePolicy := metav1.DeletePropagationForeground
	if force {
		gracePeriod := int64(0)
		deleteOptions = metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriod,
			PropagationPolicy:  &deletePolicy,
		}
	} else {
		deleteOptions = metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}
	}

	resource := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	err := handler.client.Resource(resource).Namespace(handler.namespace).Delete(context.TODO(), deploymentName, deleteOptions)

	return err
}

func (handler K8sHandler) GetDeployments() ([]string, error) {
	var names []string
	listOptions := metav1.ListOptions{}

	resource := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	deployments, err := handler.client.Resource(resource).Namespace(handler.namespace).List(context.TODO(), listOptions)
	if err != nil {
		return names, err
	}

	for _, d := range deployments.Items {
		metadata := d.Object["metadata"].(map[string]interface{})
		name := metadata["name"].(string)
		names = append(names, name)
	}

	return names, err
}
