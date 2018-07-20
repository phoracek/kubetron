package admission

import (
	"fmt"

	"github.com/golang/glog"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"
)

const (
	// Used not to overwrite last patch created by Kubernetes
	lastAppliedConfigPath = "/metadata/annotations/kubectl.kubernetes.io~1last-applied-configuration"
)

// AdmissionHook is implementation of generic-admission-server MutatingAdmissionHook interface
type AdmissionHook struct{}

// Initialize is called once when generic-addmission-server starts
func (ah *AdmissionHook) Initialize(kubeClientConfig *restclient.Config, stopCh <-chan struct{}) error {
	return nil
}

// MutatingResource registers MutatingAdmissionHook resource in Kubernetes, it is called once when generic-admission-server starts
func (ah *AdmissionHook) MutatingResource() (schema.GroupVersionResource, string) {
	return schema.GroupVersionResource{
			Group:    "kubetron.network.kubevirt.io",
			Version:  "v1alpha1",
			Resource: "admission",
		},
		"NetworkAdmission"
}

// Admit is called per each API request touching selected resources. Resource selector is defined in MutatingWebhookConfiguration as a part of Kubetron manifest and handles only Pods
func (ah *AdmissionHook) Admit(req *admissionv1beta1.AdmissionRequest) *admissionv1beta1.AdmissionResponse {
	resp := &admissionv1beta1.AdmissionResponse{}
	resp.UID = req.UID

	requestName := fmt.Sprintf("%s %s %s/%s", req.Operation, req.Kind, req.Namespace, req.Name)
	glog.V(2).Infof("[%s] Processing request", requestName)
	glog.V(6).Infof("[%s] Input: %s", requestName, string(req.Object.Raw))

	// Only handle Pod CREATE and DELETE calls, ignore the rest
	if req.Operation == admissionv1beta1.Create {
		ah.handleAdmissionRequestToCreate(requestName, req, resp)
	} else if req.Operation == admissionv1beta1.Delete {
		ah.handleAdmissionRequestToDelete(requestName, req, resp)
	} else {
		ah.ignoreAdmissionRequest(requestName, req, resp)
	}

	return resp
}

// handleAdmissionRequestToCreate makes sure that if a Pod requests networks, respective LSPs will be created and side-container requesting Kubetron Device Plugin resource will be added
func (ah *AdmissionHook) handleAdmissionRequestToCreate(requestName string, req *admissionv1beta1.AdmissionRequest, resp *admissionv1beta1.AdmissionResponse) {
	resp.Allowed = true
}

// handleAdmissionRequestToDelete reads networksSpec of give Pod, if the Pod has some ports assigned, this methods make sure they their respective LSP will be removed from OVN NB
func (ah *AdmissionHook) handleAdmissionRequestToDelete(requestName string, req *admissionv1beta1.AdmissionRequest, resp *admissionv1beta1.AdmissionResponse) {
	resp.Allowed = true
}

// ignoreAdmissionRequest ignores AdmissionRequest's contents and just allows it
func (ah *AdmissionHook) ignoreAdmissionRequest(requestName string, req *admissionv1beta1.AdmissionRequest, resp *admissionv1beta1.AdmissionResponse) {
	resp.Allowed = true
	glog.V(2).Infof("[%s] Skipping", requestName)
}
