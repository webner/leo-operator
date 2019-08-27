package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LeoCertificateSpec defines the desired state of LeoCertificate
// +k8s:openapi-gen=true
type LeoCertificateSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// domain name
	Domain string `json:"domain"`
}

// LeoCertificateStatus defines the observed state of LeoCertificate
// +k8s:openapi-gen=true
type LeoCertificateStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// messages from the operator
	Message string `json:"message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LeoCertificate is the Schema for the leocertificates API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.message"
type LeoCertificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LeoCertificateSpec   `json:"spec,omitempty"`
	Status LeoCertificateStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LeoCertificateList contains a list of LeoCertificate
type LeoCertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LeoCertificate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LeoCertificate{}, &LeoCertificateList{})
}
