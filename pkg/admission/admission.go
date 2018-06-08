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
	// This annotation is used by user to request list of networks delimited by comma
	networksAnnotationName = "kubetron.network.kubevirt.io/networks"
	// This annotation is populated by Admission Controller and contains information later used by Device Plugin
	networksSpecAnnotationName = "kubetron.network.kubevirt.io/networksSpec"
	// Used not to overwrite last patch created by Kubernetes
	lastAppliedConfigPath = "/metadata/annotations/kubectl.kubernetes.io~1last-applied-configuration"
	// Letters used to generate random interface suffix
	letters = "abcdefghijklmnopqrstuvwxyz"
)

// AdmissionHook is implementation of generic-admission-server MutatingAdmissionHook interface
type AdmissionHook struct {
	// Full URL of OVN Manager (e.g. Neutron or oVirt OVN provider)
	ProviderURL string
	// Name of resource exposed by Kubetron's Device Plugin
	ResourceName string
	// Kubernetes client instance
	client *kubernetes.Clientset
	// Client to access OVN Manager
	providerClient *providerClient
}

// Initialize is called once when generic-addmission-server starts
func (ah *AdmissionHook) Initialize(kubeClientConfig *restclient.Config, stopCh <-chan struct{}) error {
	if ah.ProviderURL == "" {
		glog.Fatal(fmt.Errorf("Provider URL was not set"))
	}
	if ah.ResourceName == "" {
		glog.Fatal(fmt.Errorf("Resource name was not set"))
	}

	// Initialize Kubernetes client, configuration is passed from generic-admission-server
	client, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize Kubernetes client: %v", err)
	}
	ah.client = client

	// Initialize OVN Manager client
	ah.providerClient = NewProviderClient(ah.ProviderURL)

	return nil
}

// MutatingResource registers MutatingAdmissionHook resource in Kubernetes, it is called once when generic-admission-server starts
func (ah *AdmissionHook) MutatingResource() (schema.GroupVersionResource, string) {
	return schema.GroupVersionResource{
			Group:    "kubetron.network.kubevirt.io",
			Version:  "v1alpha1",
			Resource: "admission",
		},
		"Admission"
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
	// Parse Pod object from request
	pod := v1.Pod{}
	err := json.Unmarshal(req.Object.Raw, &pod)
	if err != nil {
		setResponseError(resp, requestName, "Failed to read Pod: %v", err)
		return
	}

	// Read Pod's networks annotation, if it is missing (the Pod does not want any extra networks), request left unprocessed and just allowed
	annotations := pod.ObjectMeta.GetAnnotations()
	networksAnnotation, networkAnnotationFound := annotations[networksAnnotationName]
	if !networkAnnotationFound {
		glog.V(2).Infof("[%s] Skipping: Required annotation not present", requestName)
		resp.Allowed = true
		return
	}

	// Get list of desired networks, networks annotation contains a string with a list of networks delimited by comma
	networks := make([]string, 0)
	for _, rawNetwork := range strings.Split(networksAnnotation, ",") {
		networks = append(networks, strings.Trim(rawNetwork, " "))
	}
	glog.V(4).Infof("[%s] Networks: %s", requestName, networks)


	// Get map of available networks (keys) and their respective IDs (values)
	providerNetworkByName, err := ah.providerClient.ListNetworkByName()
	if err != nil {
		setResponseError(resp, requestName, "Failed to list provider networks: %v", err)
		return
	}

	// Verify that all requested networks exist and are available
	for _, network := range networks {
		if _, ok := providerNetworkByName[network]; !ok {
			setResponseError(resp, requestName, "Network %s was not found", network)
			return
		}
	}

	// initializedPod will be later updated by Admission
	initializedPod := pod.DeepCopy()

	// dhclientInterfaces is a list that keeps track of all assigned interfaces that will require DHCP client to obtain an IP address
	dhclientInterfaces := make([]string, 0)

	resources := map[v1.ResourceName]resource.Quantity{
		v1.ResourceName(ah.ResourceName): resource.MustParse("1"),
	}

	// networksSpec will be later saved as the Pod's annotation, it keeps ports' details that are later used by Device Plugin to complete attachment
	networksSpec := make(map[string]spec.NetworkSpec)

	// Create port per each network request, such ports are only in OVN NB database, no actual interfaces are created at this point
	// TODO: cleanup if fails
	for _, network := range networks {
		macAddress := generateRandomMac()
		portName := generatePortName(network)

		// Create the port on OVN NB
		portID, hasFixedIPs, err := ah.providerClient.CreateNetworkPort(providerNetworkByName[network].ID, portName, macAddress)
		if err != nil {
			setResponseError(resp, requestName, "Error creating port: %v", err)
			return
		}

		// TODO: if network has physnet, add it to list for resource request
		if providerNetworkByName[network].Physnet != "" {
			// TODO: get namespace from kubetron config
			resources[v1.ResourceName("kubetron.network.kubevirt.io/" + providerNetworkByName[network].Physnet)] = resource.MustParse("1")
			dhclientInterfaces = append(dhclientInterfaces, portName)
		}

		// If selected network has a subnet assigned, fixed IPs will be assigned to the port, add such interfaces to dhclientInterfaces list so we later call DHCP client on them
		if hasFixedIPs {
			dhclientInterfaces = append(dhclientInterfaces, portName)
		}

		// NetworkSpec is later used by Device Plugin to create an interface with PortName and MacAddress and mark it with PortID on OVS so it will be mapped to already created OVN port
		networksSpec[network] = spec.NetworkSpec{
			MacAddress: macAddress,
			PortName:   portName,
			PortID:     portID,
		}
	}

	// Marshal networksSpec into JSON bytes and save it as the Pod's annotation
	networksSpecJSON, err := json.Marshal(networksSpec)
	if err != nil {
		setResponseError(resp, requestName, "Failed to marshal networksSpec: %v", err)
		return
	}
	initializedPod.ObjectMeta.Annotations[networksSpecAnnotationName] = string(networksSpecJSON)

	// Add a side-container to the Pod. This side-container will request Kubetron resource, with this request, Kubernetes scheduler will later place this Pod on a Node with Kubetron Device Plugin installed. This side-container also runs DHCP server on ports with assigned subnet
	resourceContainer := v1.Container{
		Name:  "kubetron-request-sidecart",
		Image: "phoracek/kubetron-sidecar",
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList(resources),
		},
		SecurityContext: &v1.SecurityContext{
			// This side-container must be privileged in order to change Pod's IPs
			Privileged: newTrue(),
		},
		Args: dhclientInterfaces,
	}
	initializedPod.Spec.Containers = append(pod.Spec.Containers, resourceContainer)

	// Marshal initialized Pod specification into JSON bytes, so it can be later used to create a patch
	newData, err := json.Marshal(initializedPod)
	if err != nil {
		setResponseError(resp, requestName, "Failed to encode processed request: %v", err)
		return
	}

	// Create patch that will add the annotation and side-container to original Pod specification
	patchBytes, err := createPatch(req.Object.Raw, newData)
	if err != nil {
		setResponseError(resp, requestName, "Error creating patch: %v", err)
		return
	}

	// Save the patch to AdmissionResponse
	if string(patchBytes) != "[]" {
		glog.V(2).Infof("[%s] Patching", requestName)
		glog.V(6).Infof("[%s] Patch: %s", requestName, string(patchBytes))
		resp.Patch = patchBytes
		resp.PatchType = func() *admissionv1beta1.PatchType { // TODO: could i use it directly? new()
			pt := admissionv1beta1.PatchTypeJSONPatch
			return &pt
		}()
	}

	resp.Allowed = true
}

// handleAdmissionRequestToDelete reads networksSpec of give Pod, if the Pod has some ports assigned, this methods make sure they their respective LSP will be removed from OVN NB
func (ah *AdmissionHook) handleAdmissionRequestToDelete(requestName string, req *admissionv1beta1.AdmissionRequest, resp *admissionv1beta1.AdmissionResponse) {
	// Read current spec of the to-be-removed Pod
	pod, err := ah.client.CoreV1().Pods(req.Namespace).Get(req.Name, metav1.GetOptions{})
	if err != nil {
		setResponseError(resp, requestName, "Failed to obtain Pod: %v", err)
		return
	}

	// Read Pod's networksSpec annotation, if not found, don't process the request, just allow it
	annotations := pod.ObjectMeta.GetAnnotations()
	networksSpecAnnotation, networksSpecAnnotationFound := annotations[networksSpecAnnotationName]
	if !networksSpecAnnotationFound {
		glog.V(2).Infof("[%s] Skipping: Required annotation not present.", req.Operation, requestName)
		resp.Allowed = true
		return
	}

	// Parse networksSpec in order to obtain ports' IDs
	networksSpec := spec.NetworksSpec{}
	err = json.Unmarshal([]byte(networksSpecAnnotation), &networksSpec)
	if err != nil {
		setResponseError(resp, requestName, "Failed to read networksSpec: %v", err)
		return
	}
	glog.V(4).Infof("[%s] Network spec: %s", requestName, networksSpecAnnotation)

	// Remove assigned Pod's LSPs from OVN
	for _, spec := range networksSpec {
		err := ah.providerClient.DeleteNetworkPort(spec.PortID)
		if err != nil {
			setResponseError(resp, requestName, "Error creating port: %v", err)
			return
		}
	}
	glog.V(2).Infof("[%s] Successfully created ports", requestName)

	resp.Allowed = true
}

// ignoreAdmissionRequest ignores AdmissionRequest's contents and just allows it
func (ah *AdmissionHook) ignoreAdmissionRequest(requestName string, req *admissionv1beta1.AdmissionRequest, resp *admissionv1beta1.AdmissionResponse) {
	resp.Allowed = true
	glog.V(2).Infof("[%s] Skipping", requestName)
}

// createPatch creates a RFC 6902 patch between old and new
func createPatch(old []byte, new []byte) ([]byte, error) {
	patch, err := jsonpatch.CreatePatch(old, new)
	if err != nil {
		return nil, fmt.Errorf("failed while creating patch: %v", err)
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
		return nil, fmt.Errorf("failed while marshalling patch: %v", err)
	}

	return patchBytes, nil
}

// setResponseError is a helper that denies AdmissionRequest and populates AddmissionResponse with an error message
// TODO: Log request name
func setResponseError(resp *admissionv1beta1.AdmissionResponse, requestName string, message string, args ...interface{}) {
	glog.Errorf("[%s] %s", requestName, fmt.Sprintf(message, args...))
	resp.Allowed = false
	resp.Result = &metav1.Status{
		Status: metav1.StatusFailure, Code: http.StatusInternalServerError, Reason: metav1.StatusReasonInternalError,
		Message: fmt.Sprintf(message, args...),
	}
}

// generateRandomMac returns random local unicast MAC address
func generateRandomMac() string {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	buf[0] = (buf[0] | 2) & 0xfe // make the address local and unicast
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
}

// generatePortName builds name of a port in format ${NETWORK_NAME}-${RANDOM_SUFFIX}, length of the name is set to 15 characters so it fits max length of an interface name
// TODO: keep prefix and suffix length in a constant
func generatePortName(networkName string) string {
	prefixLen := min(len(networkName), 8)
	suffixLen := 13 - prefixLen
	suffix := make([]byte, suffixLen)
	for i := range suffix {
		suffix[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("%s-%s", networkName[0:prefixLen], suffix)
}

// min returns the lower of two int values
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// newTrue is used to create pointer to true
func newTrue() *bool {
	b := true
	return &b
}
