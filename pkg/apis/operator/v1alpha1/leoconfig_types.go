package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LeoConfigSpec defines the desired state of LeoConfig
// +k8s:openapi-gen=true
type LeoConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Production bool        `json:"production"`
	Account    LeoAccount  `json:"account,omitempty"`
	Provider   LeoProvider `json:"provider"`
}

// LeoAccount defines the email address used for registration
type LeoAccount struct {
	Email string `json:"email,omitempty"`
}

// LeoProvider defines which provider is used the solve the challenge
type LeoProvider struct {
	AcmeDNS LeoProviderAcmeDNS `json:"acmeDNS,omitempty"`
}

// LeoProviderAcmeDNS is the config for the acme dns service
type LeoProviderAcmeDNS struct {
	URL string `json:"url,omitempty"`
}

// LeoConfigStatus defines the observed state of LeoConfig
// +k8s:openapi-gen=true
type LeoConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LeoConfig is the Schema for the leoconfigs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type LeoConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LeoConfigSpec   `json:"spec,omitempty"`
	Status LeoConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LeoConfigList contains a list of LeoConfig
type LeoConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LeoConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LeoConfig{}, &LeoConfigList{})
}
