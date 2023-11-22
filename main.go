package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	flags "github.com/jessevdk/go-flags"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// IPResponse is the struct to unmarshal the IP API response
type IPResponse struct {
	IP string `json:"ip"`
}

type Arguments struct {
	FilterLabelKey      string `short:"k" long:"filter-label-key" description:"Filter ingresses by label key"`
	FilterLabelValue    string `short:"v" long:"filter-label-value" description:"Filter ingresses by label value"`
	ApiURL              string `short:"a" long:"api-url" description:"IP API URL"`
	TargetAnnotationKey string `short:"t" long:"target-annotation-key" description:"Target annotation key"`
}

var args = Arguments{
	ApiURL:              "https://api.ipify.org?format=json",
	TargetAnnotationKey: "external-dns.alpha.kubernetes.io/target",
	FilterLabelKey:      "",
	FilterLabelValue:    "",
}

func main() {

	_, err := flags.Parse(&args)
	if err != nil {
		fmt.Println("Error parsing arguments:", err)
		os.Exit(1)
	}

	numberOfChanges := 0
	ctx := context.Background()
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("Error getting cluster config:", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Error getting clientset:", err)
		os.Exit(1)
	}

	publicIP, err := getPublicIP()
	if err != nil {
		fmt.Println("Error getting public IP:", err)
		os.Exit(1)
	}
	fmt.Println("Public IP:", publicIP)

	ingresses := &v1.IngressList{}

	if args.FilterLabelKey != "" && args.FilterLabelValue != "" {
		fmt.Println("Filtering ingresses by label", args.FilterLabelKey, "=", args.FilterLabelValue)
		ingresses, err = clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", args.FilterLabelKey, args.FilterLabelValue),
		})
	} else {
		ingresses, err = clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		fmt.Println("Error getting ingresses:", err)
		os.Exit(1)
	}

	if len(ingresses.Items) == 0 {
		fmt.Println("No ingresses found")
		os.Exit(0)
	}

	for _, ingress := range ingresses.Items {
		if !checkIngressAnnotation(clientset, ctx, &ingress, args.TargetAnnotationKey, publicIP) {
			fmt.Println("Ingress", ingress.Name, "does not have the target label or has the wrong value. Setting it...")
			err := setIngressAnnotation(clientset, ctx, &ingress, args.TargetAnnotationKey, publicIP)
			if err != nil {
				fmt.Println("Error setting ingress annotation:", err)
				panic(err.Error())
			}
			fmt.Println("Ingress", ingress.Name, "target label set to", publicIP)
			numberOfChanges++
		} else {
			fmt.Println("Ingress", ingress.Name, "already has the target label set to", publicIP)
		}
	}

	fmt.Println("Number of changes:", numberOfChanges)
	os.Exit(0)
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
	resp, err := http.Get(args.ApiURL)
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
