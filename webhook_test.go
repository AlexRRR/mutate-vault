package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/api/admission/v1beta1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("v", "5")
}

func getTestWebhookServer() *WebhookServer {
	return &WebhookServer{}
}

func aPodWithoutSecret() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpod",
			Namespace: "testnamespace",
			Labels: map[string]string{
				"VaultPath":     "/secret/dark",
				"inject_secret": "true",
			},
		},
		Spec: corev1.PodSpec{},
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
	// deploymentResource := metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

	// raw, _ := json.Marshal(aDeploymentWithLimits())

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

// inspired by https://github.com/tillkahlbrock/limits-admission-webhook/blob/fe293763304f23835cb3ce49686d9556daae5ea8/main.go
func TestMutation(t *testing.T) {
	deploymentResource := metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	raw, _ := json.Marshal(aPodWithoutSecret())
	aReq := &v1beta1.AdmissionReview{
		Request: &v1beta1.AdmissionRequest{
			Resource:  deploymentResource,
			Name:      "testAdmission",
			Namespace: "testNamespace",
			Object:    runtime.RawExtension{Raw: raw},
		},
	}
	server := getTestWebhookServer()
	server.mutate(aReq)
}
