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

type TargetRef struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace,omitempty"`
}

type Strategy struct {
	Mode   string `json:"mode"`   // "count" or "rate"
	Window string `json:"window"` // duration for rate mode
}

type CircuitBreakerSpec struct {
	TargetRef          TargetRef `json:"targetRef"`
	FailureThreshold   int32     `json:"failureThreshold"`
	SuccessThreshold   int32     `json:"successThreshold"`
	TimeoutSeconds     int32     `json:"timeoutSeconds"`
	ResetTimeout       int32     `json:"resetTimeout"`
	Strategy           Strategy  `json:"strategy"`
}

type CircuitBreakerStatus struct {
	State          CircuitBreakerState `json:"state"`
	FailureCount   int32               `json:"failureCount"`
	SuccessCount   int32               `json:"successCount"`
	LastFailure    *metav1.Time        `json:"lastFailure,omitempty"`
	LastSuccess    *metav1.Time        `json:"lastSuccess,omitempty"`
	TargetApplied  bool                `json:"targetApplied"`
	Conditions     []metav1.Condition  `json:"conditions,omitempty"`
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