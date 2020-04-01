package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VMISpec defines which resource is targeted for generation
type VMISpec struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// ServiceRefSpec defines the service for the route to reference
type ServiceRefSpec struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// ServiceSpec defines the details of the service to be generated
type ServiceSpec struct {
	metav1.ObjectMeta `json:",omitempty,inline"`
	Selector          map[string]string `json:"selector,omitempty"`
	Port              int32             `json:"port,omitempty"`
	TargetPort        int32             `json:"targetPort,omitempty"`
}

// RouteSpec defines the details of the routes to be generated
type RouteSpec struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Host              string         `json:"host,omitempty"`
	ServiceRef        ServiceRefSpec `json:"serviceRef,omitempty"`
	GenerateEndpoint  bool           `json:"generateEndpoint,omitempty"`
}

// GenerateSpec defines the gvr and wanted metadata to be used for generating new objects
type GenerateSpec struct {
	VMI      VMISpec       `json:"vmi"`
	Services []ServiceSpec `json:"services,omitempty"`
	Routes   []RouteSpec   `json:"routes,omitempty"`
}

// KubevirtAddonSpec defines the desired state of KubevirtAddon
type KubevirtAddonSpec struct {
	Generate *GenerateSpec `json:"generate,omitempty"`
}

// KubevirtAddonStatus defines the observed state of KubevirtAddon
type KubevirtAddonStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KubevirtAddon is the Schema for the kubevirtaddons API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=kubevirtaddons,scope=Namespaced
type KubevirtAddon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubevirtAddonSpec   `json:"spec,omitempty"`
	Status KubevirtAddonStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KubevirtAddonList contains a list of KubevirtAddon
type KubevirtAddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubevirtAddon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubevirtAddon{}, &KubevirtAddonList{})
}
