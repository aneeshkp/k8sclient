package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var (
	kubeconfig   string
	apiServerURL string
)

type Clients struct {
	kubeClient *kubernetes.Clientset
}

func init() {

	flag.StringVar(&kubeconfig, "kubeconfig", "./kubeconfig/config", "Path to a kubeconfig. Only required if out-of-cluster.")
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
	clients := &Clients{
		kubeClient: kubeClient,
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

	services, err := kubeClient.CoreV1().Services("").List(metav1.ListOptions{})
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
	pods, err := kubeClient.CoreV1().Pods("").List(metav1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}

	pod := clients.createPod("puttur")
	time.Sleep(10 * time.Second)
	clients.printPodLogs(*pod)
	for _, pod := range pods.Items {
		fmt.Println(pod.Name, pod.Status.PodIP)
	}

}

func (c *Clients) createNameSpace(ns string) error {

	nsSpec := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	_, err := c.kubeClient.Core().Namespaces().Create(nsSpec)
	return err
}

// create pod
func (c *Clients) deletePod() {
	listOptions := metav1.ListOptions{
		LabelSelector: "app=weather-report",
	}
	podList, _ := c.kubeClient.CoreV1().Pods("").List(listOptions)
	if len(podList.Items) != 0 {
		fmt.Printf("Expected no items in podList, got %d", len(podList.Items))
	}
}
func (c *Clients) createPod(city string) *corev1.Pod {

	//kubeClient.CoreV1().Pods("").Create()

	url := fmt.Sprintf("http://wttr.in/%s?%d", city, 1)
	labels := map[string]string{
		"app":  "weather-report",
		"city": city,
		"days": "1",
	}

	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weather-report-" + strconv.Itoa(time.Now().Nanosecond()),
			Namespace: "default",
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "weather",
					Image:   "tutum/curl",
					Command: []string{"sh", "-c", "curl -s " + url + " && sleep 3600"},
				},
			},
		},
	}
	if c.kubeClient == nil {
		fmt.Printf("Client is nil")
	}

	newpod, _ := c.kubeClient.CoreV1().Pods("default").Create(pod)

	//fmt.Printf("Pod created: %s", resp)
	return newpod
}
func (c *Clients) printPodLogs(pod corev1.Pod) {
	podLogOpts := corev1.PodLogOptions{}

	req := c.kubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream()
	if err != nil {
		fmt.Println("error in opening stream")
	}
	if podLogs == nil {
		fmt.Println("error in opening stream")
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		fmt.Println("error in copy information from podLogs to buf")
	}
	str := buf.String()

	fmt.Println(str)
}
func createDeployment() {

}
func createDaemonSet() {

}

func createServiceAccount() {

}
func createRBAC() {

}

func createLabel() {

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
