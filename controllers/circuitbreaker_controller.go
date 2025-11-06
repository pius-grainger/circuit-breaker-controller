package controllers

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	circuitbreakerv1 "circuit-breaker-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CircuitBreakerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=circuitbreaker.io,resources=circuitbreakers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=circuitbreaker.io,resources=circuitbreakers/status,verbs=get;update;patch

func (r *CircuitBreakerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling CircuitBreaker", "name", req.Name, "namespace", req.Namespace)

	var cb circuitbreakerv1.CircuitBreaker
	if err := r.Get(ctx, req.NamespacedName, &cb); err != nil {
		log.Error(err, "Failed to get CircuitBreaker")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Current status", "state", cb.Status.State, "failureCount", cb.Status.FailureCount, "threshold", cb.Spec.FailureThreshold)

	// Initialize status if empty
	originalState := cb.Status.State
	originalFailureCount := cb.Status.FailureCount
	if cb.Status.State == "" {
		cb.Status.State = circuitbreakerv1.StateClosed
	}

	// Circuit breaker logic
	now := metav1.Now()
	stateChanged := false
	
	switch cb.Status.State {
	case circuitbreakerv1.StateClosed:
		if cb.Status.FailureCount >= cb.Spec.FailureThreshold {
			cb.Status.State = circuitbreakerv1.StateOpen
			cb.Status.LastFailure = &now
			stateChanged = true
			log.Info("Circuit breaker opened", "name", cb.Name, "failureCount", cb.Status.FailureCount)
		}
	case circuitbreakerv1.StateOpen:
		if cb.Status.LastFailure != nil {
			elapsed := now.Time.Sub(cb.Status.LastFailure.Time)
			if elapsed.Seconds() >= float64(cb.Spec.ResetTimeout) {
				cb.Status.State = circuitbreakerv1.StateHalfOpen
				stateChanged = true
				log.Info("Circuit breaker half-opened", "name", cb.Name)
			}
		}
	case circuitbreakerv1.StateHalfOpen:
		// In real implementation, this would check for successful requests
		// For demo, auto-close after timeout
		if cb.Status.LastFailure != nil {
			elapsed := now.Time.Sub(cb.Status.LastFailure.Time)
			if elapsed.Seconds() >= float64(cb.Spec.ResetTimeout+cb.Spec.TimeoutSeconds) {
				cb.Status.State = circuitbreakerv1.StateClosed
				cb.Status.FailureCount = 0
				stateChanged = true
				log.Info("Circuit breaker closed", "name", cb.Name)
			}
		}
	}

	// Always update if state changed or if this is initialization
	if stateChanged || originalState == "" || cb.Status.FailureCount != originalFailureCount {

		if err := r.Status().Update(ctx, &cb); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func (r *CircuitBreakerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&circuitbreakerv1.CircuitBreaker{}).
		Complete(r)
}