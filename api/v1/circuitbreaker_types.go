package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "Closed"
	StateOpen     CircuitBreakerState = "Open"
	StateHalfOpen CircuitBreakerState = "HalfOpen"
)

type CircuitBreakerSpec struct {
	FailureThreshold int32 `json:"failureThreshold"`
	TimeoutSeconds   int32 `json:"timeoutSeconds"`
	ResetTimeout     int32 `json:"resetTimeout"`
}

type CircuitBreakerStatus struct {
	State        CircuitBreakerState `json:"state"`
	FailureCount int32               `json:"failureCount"`
	LastFailure  *metav1.Time        `json:"lastFailure,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type CircuitBreaker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CircuitBreakerSpec   `json:"spec,omitempty"`
	Status CircuitBreakerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type CircuitBreakerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CircuitBreaker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CircuitBreaker{}, &CircuitBreakerList{})
}