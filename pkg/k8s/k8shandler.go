package k8s

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

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
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
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

func (handler K8sHandler) ComposeDeployment() string {
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fibonacci-deployment
  labels:
    app: fibonacci
spec:
  replicas: 1
  selector:
    matchLabels:
      app: fibonacci
  template:
    metadata:
      labels:
        app: fibonacci
    spec:
      containers:
      - name: fibonacci
        image: johan/fibonacci
        command:
            - "go"
            - "run"
            - "solver.go"
        env:
        - name: COLONYID
          value: "6007729ab9a8985b3a3d2da67f255ba13632c4670fe5c218981d77c55f7b3cab"
        - name: RUNTIME_PRVKEY
          value: "2a8647f61c18eb0fe05b33ee1bbe6c7b946bcc763b29f9a3601ea85cb5f7b6eb"
        - name: COLONIES_SERVER_HOST
          value: "10.0.0.240"
        - name: COLONIES_SERVER_PORT
          value: "8080"
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

	deploymentRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	_, err = handler.client.Resource(deploymentRes).Namespace(handler.namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	fmt.Println(deploymentRes)

	return nil
}

func (handler K8sHandler) DeleteDeployment(deploymentName string, force bool) error {
	deploymentResDep := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

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

	err := handler.client.Resource(deploymentResDep).Namespace(handler.namespace).Delete(context.TODO(), "fibonacci-deployment", deleteOptions)

	return err
}
