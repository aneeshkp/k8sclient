package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var (
	kubeconfig   string
	apiServerURL string
)

func init() {

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&apiServerURL, "master", "",
		"(Deprecated: switch to `--kubeconfig`) The address of the Kubernetes API server. Overrides any value in kubeconfig. "+
			"Only required if out-of-cluster.")
}
func main() {
	flag.Parse()
	// uses the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags(apiServerURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	if err != nil {
		panic(err.Error())
	}

	if err != nil {
		panic(err.Error())
	}
	if config.ServerName == "" {
		fmt.Printf("The cluster server name is %s ", config.Host)
	} else {
		fmt.Printf("The cluster server name is %s ", config.ServerName)
	}

	services, err := kubeClient.CoreV1().Services("").List(v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("")
	fmt.Println("-----------Services are-----------")
	for _, service := range services.Items {
		fmt.Println(service.Name, service.GetName())
	}
	fmt.Printf("There are %d servies in the cluster\n", len(services.Items))

	fmt.Println("")
	fmt.Println("-----------PODS ARE-----------")
	//pods, err := kubeClient.CoreV1().Pods("").List(v1.ListOptions{
	//FieldSelector: "spec.nodeName=aws-node"})
	pods, err := kubeClient.CoreV1().Pods("").List(v1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}
	for _, pod := range pods.Items {
		fmt.Println(pod.Name, pod.Status.PodIP)
	}

}
func getDefaultServer() string {
	if server := os.Getenv("KUBERNETES_MASTER"); len(server) > 0 {
		return server
	}
	return "http://localhost:8080"
}

// loadConfig loads a REST Config as per the rules specified in GetConfig
func loadConfig() (*rest.Config, error) {
	// If a flag is specified with the config location, use that
	fmt.Print(len(kubeconfig))
	if len(kubeconfig) > 0 {
		return clientcmd.BuildConfigFromFlags(apiServerURL, kubeconfig)
	}
	// If an env variable is specified with the config locaiton, use that
	if len(os.Getenv("KUBECONFIG")) > 0 {
		return clientcmd.BuildConfigFromFlags(apiServerURL, os.Getenv("KUBECONFIG"))
	}
	// If no explicit location, try the in-cluster config
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster config, try the default location in the user's home directory
	if usr, err := user.Current(); err == nil {
		if c, err := clientcmd.BuildConfigFromFlags(
			"", filepath.Join(usr.HomeDir, ".kube", "config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not locate a kubeconfig")
}
