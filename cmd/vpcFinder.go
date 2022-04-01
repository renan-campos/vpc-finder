package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	awsIMDSv1Server = "http://169.254.169.254"
)

func main() {
	var imdsServer string

	namespace, found := os.LookupEnv("NAMESPACE")
	if !found {
		fmt.Fprintf(os.Stderr,
			"NAMESPACE environment variabled not found\n")
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		imdsServer = os.Args[1]
	} else {
		imdsServer = awsIMDSv1Server
	}
	fmt.Printf("Using IMDSv1 server: %s\n", imdsServer)

	fmt.Println("Setting up k8s client")
	k8sClient, err := setupK8sClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set up k8s client: %s", err)
		os.Exit(1)
	}

	// A mac address for one of the interfaces on the instance is needed
	// to query for the VPC CIDR block.
	fmt.Println("Determining mac address on instance")
	mac := httpGet(fmt.Sprintf("%s/%s", imdsServer, "latest/meta-data/mac"))
	if mac == "" {
		fmt.Fprintf(os.Stderr,
			"Failed for find mac address of instance\n")
		os.Exit(1)
	}
	fmt.Printf("mac address found: %s\n", mac)

	fmt.Println("Determining VPC CIDR")
	cidr := httpGet(
		fmt.Sprintf("%s/%s/%s/%s",
			imdsServer, "latest/meta-data/network/interfaces/macs", mac, "vpc-ipv4-cidr-block"))
	if cidr == "" {
		fmt.Fprintf(os.Stderr,
			"Failed to find VPC CIDR\n")
		os.Exit(1)
	}
	fmt.Printf("VPC CIDR: %s\n", cidr)

	fmt.Println("Creating Config Map with AWS data")
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "aws-data",
		},
		Data: map[string]string{
			"vpc-cidr": cidr,
		},
	}
	err = createOrUpdateConfigMap(k8sClient, &configMap, namespace)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Failed to create configmap: %s", err)
		os.Exit(1)
	}
	fmt.Println("vpc-finder successfully completed!")
}

func setupK8sClient() (*kubernetes.Clientset, error) {
	// To run locally:
	// Point the environment variable LOCAL_RUN to your kubeconfig
	localKubeconfig, found := os.LookupEnv("LOCAL_RUN")
	if found {
		config, err := clientcmd.BuildConfigFromFlags("", localKubeconfig)
		if err != nil {
			panic(err.Error())
		}

		return kubernetes.NewForConfig(config)
	}

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	return kubernetes.NewForConfig(config)
}

func createOrUpdateConfigMap(k8sClient *kubernetes.Clientset, configMap *corev1.ConfigMap, namespace string) error {
	_, err := k8sClient.CoreV1().ConfigMaps(namespace).Create(
		context.Background(), configMap, metav1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			_, err = k8sClient.CoreV1().ConfigMaps(namespace).Update(
				context.Background(), configMap, metav1.UpdateOptions{})
		}
	}
	return err
}

func httpGet(address string) string {
	for tries := 0; tries < 3; tries++ {
		if tries > 0 {
			fmt.Println("\tRetrying...")
		}

		resp, err := http.Get(address)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Failed to GET %s: %s\n",
				address, err)
			continue
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Failed to read body of GET %s: %s\n",
				address, err)
			continue
		}
		return string(body)
	}
	return ""
}
