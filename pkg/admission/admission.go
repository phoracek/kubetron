// TODO: improve logging
// TODO: add documentation
// TODO: cleanup and refactoring
// TODO: expose env vars with KUBETRON_NETWORK_NAME="interface_name" to containers
package admission

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/mattbaird/jsonpatch"
	"github.com/phoracek/kubetron/pkg/spec"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const (
	networksAnnotationName     = "kubetron.network.kubevirt.io/networks"
	networksSpecAnnotationName = "kubetron.network.kubevirt.io/networksSpec"
	lastAppliedConfigPath      = "/metadata/annotations/kubectl.kubernetes.io~1last-applied-configuration"
	letters                    = "abcdefghijklmnopqrstuvwxyz"
)

type AdmissionHook struct {
	ProviderURL    string
	ResourceName   string
	client         *kubernetes.Clientset
	providerClient *providerClient
}

func (ah *AdmissionHook) Initialize(kubeClientConfig *restclient.Config, stopCh <-chan struct{}) error {
	if ah.ProviderURL == "" {
		return fmt.Errorf("provider-url was not set")
	}

	client, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return fmt.Errorf("failed to intialise kubernetes clientset: %v", err)
	}
	ah.client = client

	ah.providerClient = NewProviderClient(ah.ProviderURL)

	glog.Info("Webhook Initialization Complete.")
	return nil
}

func (ah *AdmissionHook) MutatingResource() (schema.GroupVersionResource, string) {
	return schema.GroupVersionResource{
			Group:    "kubetron.network.kubevirt.io",
			Version:  "v1alpha1",
			Resource: "admission",
		},
		"Admission"
}

func (ah *AdmissionHook) Admit(req *admissionv1beta1.AdmissionRequest) *admissionv1beta1.AdmissionResponse {
	if req.Operation == admissionv1beta1.Create {
		return ah.admitCreate(req)
	} else if req.Operation == admissionv1beta1.Delete {
		return ah.admitDelete(req)
	} else {
		return ah.admitIgnore(req)
	}
}

func (ah *AdmissionHook) admitCreate(req *admissionv1beta1.AdmissionRequest) *admissionv1beta1.AdmissionResponse {
	resp := &admissionv1beta1.AdmissionResponse{}
	resp.UID = req.UID
	requestName := fmt.Sprintf("%s %s/%s", req.Kind, req.Namespace, req.Name)

	pod := v1.Pod{}
	err := json.Unmarshal(req.Object.Raw, &pod)
	if err != nil {
		return errorResponse(resp, "Failed to read Pod: %v", err)
	}

	annotations := pod.ObjectMeta.GetAnnotations()
	networksAnnotation, networkAnnotationFound := annotations[networksAnnotationName]
	if !networkAnnotationFound {
		glog.V(2).Infof("Skipping %s request for %s: Required annotation not present.", req.Operation, requestName)
		resp.Allowed = true
		return resp
	}

	networks := make([]string, 0)
	for _, rawNetwork := range strings.Split(networksAnnotation, ",") {
		networks = append(networks, strings.Trim(rawNetwork, " "))
	}
	// TODO: v6 log for networks

	glog.V(2).Infof("Processing %s request for %s", req.Operation, requestName)

	glog.V(6).Infof("Input for %s: %s", requestName, string(req.Object.Raw))

	providerNetworkIDsByNames, err := ah.providerClient.ListNetworkIDsByNames()
	if err != nil {
		return errorResponse(resp, "Failed to list provider networks: %v", err)
	}
	for _, network := range networks {
		if _, ok := providerNetworkIDsByNames[network]; !ok {
			return errorResponse(resp, "Network %s was not found", network)
		}
	}

	initializedPod := pod.DeepCopy()

	// TODO: only first one was checked
	dhclientInterfaces := make([]string, 0)

	// TODO: use struct instead of map, use networkspec type
	// TODO: cleanup if fails
	networksSpec := make(map[string]spec.NetworkSpec)
	for _, network := range networks {
		macAddress := generateRandomMac()
		portName := generatePortName(network)
		portID, hasFixedIPs, err := ah.providerClient.CreateNetworkPort(providerNetworkIDsByNames[network], portName, macAddress)
		if err != nil {
			return errorResponse(resp, "Error creating port: %v", err)
		}

		if hasFixedIPs {
			dhclientInterfaces = append(dhclientInterfaces, portName)
		}

		networksSpec[network] = spec.NetworkSpec{
			MacAddress: macAddress,
			PortName:   portName,
			PortID:     portID,
		}
	}
	networksSpecJSON, err := json.Marshal(networksSpec)
	if err != nil {
		return errorResponse(resp, "Failed to marshal networksSpec: %v", err)
	}
	initializedPod.ObjectMeta.Annotations[networksSpecAnnotationName] = string(networksSpecJSON)

	// TODO: configure readiness on sidecar
	resourceContainer := v1.Container{
		Name:  "kubetron-request-sidecart",
		Image: "phoracek/kubetron-sidecar",
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceName(ah.ResourceName): resource.MustParse("1"),
			},
		},
		SecurityContext: &v1.SecurityContext{
			Privileged: newTrue(),
		},
		Args: dhclientInterfaces,
	}
	initializedPod.Spec.Containers = append(pod.Spec.Containers, resourceContainer)

	newData, err := json.Marshal(initializedPod)
	if err != nil {
		return errorResponse(resp, "Failed to encode processed request: %v", err)
	}

	patchBytes, err := createPatch(req.Object.Raw, newData)
	if err != nil {
		return errorResponse(resp, "Error creating patch: %v", err)
	}

	if string(patchBytes) != "[]" {
		glog.V(2).Infof("Patching %s", requestName)
		glog.V(4).Infof("Patch for %s: %s", requestName, string(patchBytes))
		resp.Patch = patchBytes
		resp.PatchType = func() *admissionv1beta1.PatchType { // TODO: could i use it directly? new()
			pt := admissionv1beta1.PatchTypeJSONPatch
			return &pt
		}()
	}

	resp.Allowed = true
	return resp
}

// TODO: move duplicated code to functions
func (ah *AdmissionHook) admitDelete(req *admissionv1beta1.AdmissionRequest) *admissionv1beta1.AdmissionResponse {
	resp := &admissionv1beta1.AdmissionResponse{}
	resp.UID = req.UID
	requestName := fmt.Sprintf("%s %s/%s", req.Kind, req.Namespace, req.Name)

	pod, err := ah.client.CoreV1().Pods(req.Namespace).Get(req.Name, metav1.GetOptions{})
	if err != nil {
		return errorResponse(resp, "Failed to obtain Pod: %v", err)
	}

	glog.V(2).Infof("Processing %s request for %s", req.Operation, requestName)

	glog.V(6).Infof("Input for %s: %s", requestName, string(req.Object.Raw))

	annotations := pod.ObjectMeta.GetAnnotations()
	networksSpecAnnotation, networksSpecAnnotationFound := annotations[networksSpecAnnotationName]
	if !networksSpecAnnotationFound {
		glog.V(2).Infof("Skipping %s request for %s: Required annotation not present.", req.Operation, requestName)
		resp.Allowed = true
		return resp
	}

	networksSpec := spec.NetworksSpec{}
	err = json.Unmarshal([]byte(networksSpecAnnotation), &networksSpec)
	if err != nil {
		return errorResponse(resp, "Failed to read networksSpec: %v", err)
	}
	glog.V(2).Infof("Network spec for request %s: %s", requestName, networksSpecAnnotation)

	for _, spec := range networksSpec {
		err := ah.providerClient.DeleteNetworkPort(spec.PortID)
		if err != nil {
			return errorResponse(resp, "Error creating port: %v", err)
		}
	}
	glog.V(2).Infof("Successfully created ports for request %s", requestName)

	resp.Allowed = true
	return resp
}

func (ah *AdmissionHook) admitIgnore(req *admissionv1beta1.AdmissionRequest) *admissionv1beta1.AdmissionResponse {
	resp := &admissionv1beta1.AdmissionResponse{}
	resp.UID = req.UID
	requestName := fmt.Sprintf("%s %s/%s", req.Kind, req.Namespace, req.Name)
	glog.V(2).Infof("Skipping %s request for %s", req.Operation, requestName)
	resp.Allowed = true
	return resp
}

func createPatch(old []byte, new []byte) ([]byte, error) {
	patch, err := jsonpatch.CreatePatch(old, new)
	if err != nil {
		return nil, fmt.Errorf("error calculating patch: %v", err)
	}

	allowedOps := []jsonpatch.JsonPatchOperation{}
	for _, op := range patch {
		// Don't patch the lastAppliedConfig created by kubectl
		if op.Path == lastAppliedConfigPath {
			continue
		}
		allowedOps = append(allowedOps, op)
	}

	patchBytes, err := json.Marshal(allowedOps)
	if err != nil {
		return nil, fmt.Errorf("error marshalling patch: %v", err)
	}
	return patchBytes, nil
}

func errorResponse(resp *admissionv1beta1.AdmissionResponse, message string, args ...interface{}) *admissionv1beta1.AdmissionResponse {
	glog.Errorf(message, args...)
	resp.Allowed = false
	resp.Result = &metav1.Status{
		Status: metav1.StatusFailure, Code: http.StatusInternalServerError, Reason: metav1.StatusReasonInternalError,
		Message: fmt.Sprintf(message, args...),
	}
	return resp
}

func generateRandomMac() string {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	buf[0] = (buf[0] | 2) & 0xfe
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
}

func generatePortName(networkName string) string {
	prefixLen := min(len(networkName), 8)
	suffixLen := 13 - prefixLen
	suffix := make([]byte, suffixLen)
	for i := range suffix {
		suffix[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("%s-%s", networkName[0:prefixLen], suffix)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func newTrue() *bool {
	b := true
	return &b
}
