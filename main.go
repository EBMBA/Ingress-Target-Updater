package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// IPResponse is the struct to unmarshal the IP API response
type IPResponse struct {
	IP string `json:"ip"`
}

var (
	apiURL              = "https://api.ipify.org?format=json"
	targetAnnotationKey = "external-dns.alpha.kubernetes.io/target"
)

func main() {
	ctx := context.Background()
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	publicIP, err := getPublicIP()
	if err != nil {
		fmt.Println("Error getting public IP:", err)
		os.Exit(1)
	}
	fmt.Println("Public IP:", publicIP)

	ingresses, err := clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})

	if err != nil {
		fmt.Println("Error getting ingresses:", err)
		os.Exit(1)
	}

	for _, ingress := range ingresses.Items {
		if !checkIngressAnnotation(clientset, ctx, &ingress, targetAnnotationKey, publicIP) {
			fmt.Println("Ingress", ingress.Name, "does not have the target label or has the wrong value. Setting it...")
			err := setIngressAnnotation(clientset, ctx, &ingress, targetAnnotationKey, publicIP)
			if err != nil {
				fmt.Println("Error setting ingress annotation:", err)
				panic(err.Error())
			}
			fmt.Println("Ingress", ingress.Name, "target label set to", publicIP)
		} else {
			fmt.Println("Ingress", ingress.Name, "already has the target label set to", publicIP)
		}
	}
}

func checkIngressAnnotation(clientset *kubernetes.Clientset, ctx context.Context, ingress *v1.Ingress, annotationKey string, wantedValue string) bool {
	if val, ok := ingress.ObjectMeta.GetAnnotations()[annotationKey]; ok {
		if val == wantedValue {
			return true
		}
	}
	return false
}

func setIngressAnnotation(clientset *kubernetes.Clientset, ctx context.Context, ingress *v1.Ingress, annotationKey string, annotationValue string) error {
	ingress.ObjectMeta.GetAnnotations()[annotationKey] = annotationValue

	ingress, err := clientset.NetworkingV1().Ingresses(ingress.Namespace).Update(ctx, ingress, metav1.UpdateOptions{})
	if err != nil {
		fmt.Println("Error updating ingress:", err)
		return err
	}
	return nil
}

func getPublicIP() (string, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ipResponse IPResponse
	err = json.Unmarshal(body, &ipResponse)
	if err != nil {
		return "", err
	}

	return ipResponse.IP, nil
}
