package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	circuitbreakerv1 "circuit-breaker-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CircuitBreakerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=circuitbreaker.io,resources=circuitbreakers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=circuitbreaker.io,resources=circuitbreakers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

func (r *CircuitBreakerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	// Reduced logging for production
	log.V(1).Info("Reconciling circuit breaker", "name", req.Name, "namespace", req.Namespace)

	var cb circuitbreakerv1.CircuitBreaker
	if err := r.Get(ctx, req.NamespacedName, &cb); err != nil {
		log.Error(err, "Failed to get CircuitBreaker")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Only log status for non-closed states or when verbose logging enabled
	if cb.Status.State != circuitbreakerv1.StateClosed {
		log.Info("Circuit breaker status", "state", cb.Status.State, "failures", cb.Status.FailureCount, "target", cb.Spec.TargetRef.Kind+"/"+cb.Spec.TargetRef.Name)
	} else {
		log.V(1).Info("Circuit breaker status", "state", cb.Status.State, "failures", cb.Status.FailureCount)
	}

	// Initialize status if empty
	originalState := cb.Status.State
	originalFailureCount := cb.Status.FailureCount
	if cb.Status.State == "" {
		cb.Status.State = circuitbreakerv1.StateClosed
	}

	// Circuit breaker logic
	now := metav1.Now()
	stateChanged := false
	previousState := cb.Status.State
	
	log.V(1).Info("Evaluating state transitions", "currentState", cb.Status.State)
	
	switch cb.Status.State {
	case circuitbreakerv1.StateClosed:
		log.V(2).Info("Checking CLOSED state conditions", "failures", cb.Status.FailureCount, "threshold", cb.Spec.FailureThreshold)
		if cb.Status.FailureCount >= cb.Spec.FailureThreshold {
			cb.Status.State = circuitbreakerv1.StateOpen
			cb.Status.LastFailure = &now
			cb.Status.LastTransitionTime = &now
			cb.Status.Reason = "FailureThresholdExceeded"
			cb.Status.Message = fmt.Sprintf("Circuit breaker opened due to %d failures exceeding threshold of %d", cb.Status.FailureCount, cb.Spec.FailureThreshold)
			stateChanged = true
			r.Recorder.Event(&cb, "Warning", "CircuitBreakerOpened", cb.Status.Message)
			log.Info("🔴 CIRCUIT BREAKER OPENED", "name", cb.Name, "failureCount", cb.Status.FailureCount, "threshold", cb.Spec.FailureThreshold)
		} else {
			log.V(2).Info("Circuit breaker remains CLOSED", "failures", cb.Status.FailureCount, "threshold", cb.Spec.FailureThreshold)
		}
	case circuitbreakerv1.StateOpen:
		if cb.Status.LastFailure != nil {
			elapsed := now.Time.Sub(cb.Status.LastFailure.Time)
			log.Info("Checking OPEN state conditions", "elapsed", elapsed.Seconds(), "resetTimeout", cb.Spec.ResetTimeout)
			if elapsed.Seconds() >= float64(cb.Spec.ResetTimeout) {
				cb.Status.State = circuitbreakerv1.StateHalfOpen
				cb.Status.SuccessCount = 0
				cb.Status.LastTransitionTime = &now
				cb.Status.Reason = "ResetTimeoutElapsed"
				cb.Status.Message = fmt.Sprintf("Circuit breaker transitioned to half-open after %v seconds", elapsed.Seconds())
				stateChanged = true
				r.Recorder.Event(&cb, "Normal", "CircuitBreakerHalfOpened", cb.Status.Message)
				log.Info("🟡 CIRCUIT BREAKER HALF-OPENED", "name", cb.Name, "elapsedTime", elapsed.Seconds())
			} else {
				log.V(2).Info("Circuit breaker remains OPEN", "elapsed", elapsed.Seconds(), "resetTimeout", cb.Spec.ResetTimeout)
			}
		}
	case circuitbreakerv1.StateHalfOpen:
		log.Info("Checking HALF-OPEN state conditions", "successes", cb.Status.SuccessCount, "threshold", cb.Spec.SuccessThreshold)
		if cb.Status.SuccessCount >= cb.Spec.SuccessThreshold {
			cb.Status.State = circuitbreakerv1.StateClosed
			cb.Status.FailureCount = 0
			cb.Status.SuccessCount = 0
			cb.Status.LastTransitionTime = &now
			cb.Status.Reason = "SuccessThresholdMet"
			cb.Status.Message = fmt.Sprintf("Circuit breaker closed after %d successful requests", cb.Status.SuccessCount)
			stateChanged = true
			r.Recorder.Event(&cb, "Normal", "CircuitBreakerClosed", cb.Status.Message)
			log.Info("🟢 CIRCUIT BREAKER CLOSED", "name", cb.Name, "successCount", cb.Status.SuccessCount)
		} else if cb.Status.LastFailure != nil {
			elapsed := now.Time.Sub(cb.Status.LastFailure.Time)
			if elapsed.Seconds() >= float64(cb.Spec.TimeoutSeconds) {
				cb.Status.State = circuitbreakerv1.StateOpen
				stateChanged = true
				log.Info("🔴 CIRCUIT BREAKER REOPENED from half-open", "name", cb.Name, "timeout", cb.Spec.TimeoutSeconds)
			}
		}
	}

	// Apply target configuration if state changed
	if stateChanged {
		log.Info("🎯 State changed", "from", previousState, "to", cb.Status.State, "target", cb.Spec.TargetRef.Kind+"/"+cb.Spec.TargetRef.Name)
		if err := r.applyTargetConfig(ctx, &cb, previousState); err != nil {
			log.Error(err, "❌ FAILED to apply target configuration", "target", cb.Spec.TargetRef.Kind+"/"+cb.Spec.TargetRef.Name)
			cb.Status.TargetApplied = false
		} else {
			log.Info("✅ TARGET CONFIGURATION APPLIED SUCCESSFULLY", "target", cb.Spec.TargetRef.Kind+"/"+cb.Spec.TargetRef.Name)
			cb.Status.TargetApplied = true
		}
	} else {
		log.V(2).Info("No state change, skipping target config")
	}

	// Always update if state changed or if this is initialization
	if stateChanged || originalState == "" || cb.Status.FailureCount != originalFailureCount {
		log.V(1).Info("Updating status", "state", cb.Status.State, "failures", cb.Status.FailureCount)
		if err := r.Status().Update(ctx, &cb); err != nil {
			log.Error(err, "❌ FAILED to update circuit breaker status")
			return ctrl.Result{}, err
		}
		log.V(1).Info("Status updated")
	} else {
		log.V(1).Info("No status changes, skipping update")
	}

	nextRequeue := 10 * time.Second
	log.V(1).Info("Reconcile complete", "requeueAfter", nextRequeue, "state", cb.Status.State)
	return ctrl.Result{RequeueAfter: nextRequeue}, nil
}

func (r *CircuitBreakerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&circuitbreakerv1.CircuitBreaker{}).
		Complete(r)
}