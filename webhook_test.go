package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getTestWebhookServer() *WebhookServer {
	return &WebhookServer{}
}

func aDeploymentWithLimits() *v1.Deployment {
	return &v1.Deployment{
		Spec: v1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceMemory: resource.Quantity{},
							}}}}}}},
	}

}

func TestServerHandleAdmissionRequest(t *testing.T) {

	tests := []struct {
		name            string
		admissionReview admissionv1beta1.AdmissionReview
	}{
		{
			name: "Call webhook correctly",
			admissionReview: admissionv1beta1.AdmissionReview{
				Request: &admissionv1beta1.AdmissionRequest{
					UID: "requestUID",
				},
			},
		},
	}
	deploymentResource := metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

	raw, _ := json.Marshal(aDeploymentWithLimits())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytesIn, _ := json.Marshal(tt.admissionReview)
			req, _ := http.NewRequest("POST", "", bytes.NewBuffer(bytesIn))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			server := getTestWebhookServer()
			handler := http.HandlerFunc(server.serve)
			handler.ServeHTTP(rr, req)
		})
	}

}
