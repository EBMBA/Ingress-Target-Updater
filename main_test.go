package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	CoreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetPublicIP(t *testing.T) {
	// Create a test server to mock the API response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"IP": "127.0.0.1"}`))
	}))
	defer server.Close()

	// Set the test server URL as the API URL
	apiURL = server.URL

	// Call the function being tested
	ip, err := getPublicIP()

	// Check if the function returned an error
	if err != nil {
		t.Errorf("getPublicIP returned an unexpected error: %v", err)
	}

	// Check if the returned IP is correct
	expectedIP := "127.0.0.1"
	if ip != expectedIP {
		t.Errorf("getPublicIP returned an unexpected IP. Expected: %s, Got: %s", expectedIP, ip)
	}
}

func TestGetPublicIP_Error(t *testing.T) {
	// Create a test server to mock the API response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Set the test server URL as the API URL
	apiURL = server.URL

	// Call the function being tested
	ip, err := getPublicIP()

	// Check if the function returned an error
	if err == nil {
		t.Error("getPublicIP did not return an expected error")
	}

	// Check if the returned IP is empty
	if ip != "" {
		t.Errorf("getPublicIP returned an unexpected IP. Expected: '', Got: %s", ip)
	}
}
func TestSetIngressAnnotation(t *testing.T) {
	// Create a context
	ctx := context.TODO()
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Create a namespace object
	namespace := &CoreV1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}

	clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})

	// Create a sample ingress object
	ingress := &v1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "test-namespace",
			Annotations: map[string]string{
				"external-dns.alpha.kubernetes.io/target": "127.0.0.1",
			},
		},
	}

	// Create the ingress object
	_, err := clientset.NetworkingV1().Ingresses(ingress.Namespace).Create(ctx, ingress, metav1.CreateOptions{})

	if err != nil {
		t.Errorf("Failed to create ingress: %v", err)
	}

	// Get the ingress object from the API
	ingress, err = clientset.NetworkingV1().Ingresses(ingress.Namespace).Get(ctx, ingress.Name, metav1.GetOptions{})

	if err != nil {
		t.Errorf("Failed to get ingress: %v", err)
	}

	// Set the annotation key and value
	annotationKey := "external-dns.alpha.kubernetes.io/target"
	annotationValue := "10.0.0.1"
	ingress.ObjectMeta.Annotations[annotationKey] = annotationValue

	// Update the ingress object
	updatedIngress, err := clientset.NetworkingV1().Ingresses(ingress.Namespace).Update(ctx, ingress, metav1.UpdateOptions{})
	if err != nil {
		t.Errorf("Failed to update ingress: %v", err)
	}

	// Verify the updated annotation value
	updatedValue, ok := updatedIngress.ObjectMeta.GetAnnotations()[annotationKey]
	if !ok {
		t.Errorf("Annotation key '%s' not found in updated ingress", annotationKey)
	}

	if updatedValue != annotationValue {
		t.Errorf("Updated annotation value is incorrect. Expected: %s, Got: %s", annotationValue, updatedValue)
	}
}
