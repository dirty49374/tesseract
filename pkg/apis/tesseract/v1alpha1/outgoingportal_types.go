package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OutgoingPortalSpec defines the desired state of OutgoingPortal
// +k8s:openapi-gen=true
type OutgoingPortalSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	ServiceFDQN string `json:"serviceFDQN,omitempty"`
	ServicePort int32  `json:"servicePort,omitempty"`
	Gateway     string `json:"gateway,omitempty"`
}

// OutgoingPortalStatus defines the observed state of OutgoingPortal
// +k8s:openapi-gen=true
type OutgoingPortalStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OutgoingPortal is the Schema for the outgoingportals API
// +k8s:openapi-gen=true
type OutgoingPortal struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OutgoingPortalSpec   `json:"spec,omitempty"`
	Status OutgoingPortalStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OutgoingPortalList contains a list of OutgoingPortal
type OutgoingPortalList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OutgoingPortal `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OutgoingPortal{}, &OutgoingPortalList{})
}
