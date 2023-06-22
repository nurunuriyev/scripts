package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// DiscordWebhookURL should be set to your Discord webhook URL.
const DiscordWebhookURL = "https://discord.com/api/webhooks/1121414192130429049/xSX0IvxscjESuOxvp3EyjSRkWtjwL5TsMiMkCP5j3IMMXniH7UX2cfCsdO6isj7gG-3a"

func main() {
	var config *rest.Config
	var err error

	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, pod := range pods.Items {
		if pod.Namespace == "default" {
			continue
		}
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				fmt.Println(containerStatus.Name, containerStatus.State.Waiting.Reason, containerStatus.State.Waiting.Message)
				sendAlertToDiscord(containerStatus.Name, containerStatus.State.Waiting.Reason, containerStatus.State.Waiting.Message)

			}

		}
	}
}

func sendAlertToDiscord(podName, status, reason string) {
	content := fmt.Sprintf("Pod **%s** is not running.\nReason: `%s`\nMessage: `%s`", podName, status, reason)

	payload := map[string]interface{}{
		"content": content,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error preparing Discord payload:", err)
		return
	}

	resp, err := http.Post(DiscordWebhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Println("Error sending alert to Discord:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		fmt.Printf("Error: Discord returned %d status code\n", resp.StatusCode)
	}
}
