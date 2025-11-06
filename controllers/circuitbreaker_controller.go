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
	log.Info("=== RECONCILING CIRCUIT BREAKER ===", "name", req.Name, "namespace", req.Namespace)

	var cb circuitbreakerv1.CircuitBreaker
	if err := r.Get(ctx, req.NamespacedName, &cb); err != nil {
		log.Error(err, "Failed to get CircuitBreaker")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Current circuit breaker status", 
		"state", cb.Status.State, 
		"failureCount", cb.Status.FailureCount, 
		"successCount", cb.Status.SuccessCount,
		"failureThreshold", cb.Spec.FailureThreshold,
		"successThreshold", cb.Spec.SuccessThreshold,
		"targetRef", cb.Spec.TargetRef.Kind+"/"+cb.Spec.TargetRef.Name)

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
	
	log.Info("Evaluating circuit breaker state transitions", "currentState", cb.Status.State)
	
	switch cb.Status.State {
	case circuitbreakerv1.StateClosed:
		log.Info("Checking CLOSED state conditions", "failures", cb.Status.FailureCount, "threshold", cb.Spec.FailureThreshold)
		if cb.Status.FailureCount >= cb.Spec.FailureThreshold {
			cb.Status.State = circuitbreakerv1.StateOpen
			cb.Status.LastFailure = &now
			stateChanged = true
			log.Info("🔴 CIRCUIT BREAKER OPENED", "name", cb.Name, "failureCount", cb.Status.FailureCount, "threshold", cb.Spec.FailureThreshold)
		} else {
			log.Info("Circuit breaker remains CLOSED", "failures", cb.Status.FailureCount, "threshold", cb.Spec.FailureThreshold)
		}
	case circuitbreakerv1.StateOpen:
		if cb.Status.LastFailure != nil {
			elapsed := now.Time.Sub(cb.Status.LastFailure.Time)
			log.Info("Checking OPEN state conditions", "elapsed", elapsed.Seconds(), "resetTimeout", cb.Spec.ResetTimeout)
			if elapsed.Seconds() >= float64(cb.Spec.ResetTimeout) {
				cb.Status.State = circuitbreakerv1.StateHalfOpen
				cb.Status.SuccessCount = 0
				stateChanged = true
				log.Info("🟡 CIRCUIT BREAKER HALF-OPENED", "name", cb.Name, "elapsedTime", elapsed.Seconds())
			} else {
				log.Info("Circuit breaker remains OPEN", "elapsed", elapsed.Seconds(), "resetTimeout", cb.Spec.ResetTimeout)
			}
		}
	case circuitbreakerv1.StateHalfOpen:
		log.Info("Checking HALF-OPEN state conditions", "successes", cb.Status.SuccessCount, "threshold", cb.Spec.SuccessThreshold)
		if cb.Status.SuccessCount >= cb.Spec.SuccessThreshold {
			cb.Status.State = circuitbreakerv1.StateClosed
			cb.Status.FailureCount = 0
			cb.Status.SuccessCount = 0
			stateChanged = true
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
		log.Info("🎯 APPLYING TARGET CONFIGURATION", "previousState", previousState, "newState", cb.Status.State, "target", cb.Spec.TargetRef.Kind+"/"+cb.Spec.TargetRef.Name)
		if err := r.applyTargetConfig(ctx, &cb, previousState); err != nil {
			log.Error(err, "❌ FAILED to apply target configuration", "target", cb.Spec.TargetRef.Kind+"/"+cb.Spec.TargetRef.Name)
			cb.Status.TargetApplied = false
		} else {
			log.Info("✅ TARGET CONFIGURATION APPLIED SUCCESSFULLY", "target", cb.Spec.TargetRef.Kind+"/"+cb.Spec.TargetRef.Name)
			cb.Status.TargetApplied = true
		}
	} else {
		log.Info("No state change detected, skipping target configuration")
	}

	// Always update if state changed or if this is initialization
	if stateChanged || originalState == "" || cb.Status.FailureCount != originalFailureCount {
		log.Info("💾 UPDATING CIRCUIT BREAKER STATUS", "state", cb.Status.State, "failureCount", cb.Status.FailureCount, "targetApplied", cb.Status.TargetApplied)
		if err := r.Status().Update(ctx, &cb); err != nil {
			log.Error(err, "❌ FAILED to update circuit breaker status")
			return ctrl.Result{}, err
		}
		log.Info("✅ STATUS UPDATED SUCCESSFULLY")
	} else {
		log.Info("No status changes detected, skipping update")
	}

	nextRequeue := 10 * time.Second
	log.Info("⏰ RECONCILE COMPLETE", "requeueAfter", nextRequeue, "finalState", cb.Status.State)
	log.Info("=== END RECONCILE ===")
	return ctrl.Result{RequeueAfter: nextRequeue}, nil
}

func (r *CircuitBreakerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&circuitbreakerv1.CircuitBreaker{}).
		Complete(r)
}