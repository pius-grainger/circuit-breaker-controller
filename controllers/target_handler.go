package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	circuitbreakerv1 "circuit-breaker-controller/api/v1"
)

func (r *CircuitBreakerReconciler) applyTargetConfig(ctx context.Context, cb *circuitbreakerv1.CircuitBreaker, previousState circuitbreakerv1.CircuitBreakerState) error {
	log := log.FromContext(ctx)
	log.Info("🎯 TARGET CONFIG: Starting target configuration", "target", cb.Spec.TargetRef.Kind+"/"+cb.Spec.TargetRef.Name, "state", cb.Status.State)
	
	// Skip if no target reference
	if cb.Spec.TargetRef.APIVersion == "" {
		log.Info("⚠️ TARGET CONFIG: No target reference specified, skipping")
		return nil
	}
	
	// Get target resource
	target := &unstructured.Unstructured{}
	target.SetAPIVersion(cb.Spec.TargetRef.APIVersion)
	target.SetKind(cb.Spec.TargetRef.Kind)
	
	targetNamespace := cb.Spec.TargetRef.Namespace
	if targetNamespace == "" {
		targetNamespace = cb.Namespace
	}
	
	log.Info("🔍 TARGET CONFIG: Fetching target resource", "name", cb.Spec.TargetRef.Name, "namespace", targetNamespace, "kind", cb.Spec.TargetRef.Kind)
	err := r.Get(ctx, types.NamespacedName{
		Name:      cb.Spec.TargetRef.Name,
		Namespace: targetNamespace,
	}, target)
	if err != nil {
		log.Error(err, "❌ TARGET CONFIG: Failed to get target resource", "name", cb.Spec.TargetRef.Name, "namespace", targetNamespace)
		return fmt.Errorf("failed to get target resource: %w", err)
	}
	log.Info("✅ TARGET CONFIG: Successfully fetched target resource")

	// Apply configuration based on circuit breaker state
	switch cb.Status.State {
	case circuitbreakerv1.StateOpen:
		return r.applyOpenConfig(ctx, target, cb)
	case circuitbreakerv1.StateClosed, circuitbreakerv1.StateHalfOpen:
		return r.applyClosedConfig(ctx, target, cb)
	}
	
	return nil
}

func (r *CircuitBreakerReconciler) applyOpenConfig(ctx context.Context, target *unstructured.Unstructured, cb *circuitbreakerv1.CircuitBreaker) error {
	log := log.FromContext(ctx)
	
	// For Service, add circuit breaker annotations
	if target.GetKind() == "Service" && target.GetAPIVersion() == "v1" {
		annotations := target.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		
		// Add circuit breaker annotations
		annotations["circuitbreaker.io/state"] = "open"
		annotations["circuitbreaker.io/managed-by"] = cb.Name
		
		target.SetAnnotations(annotations)
		log.Info("Applied open circuit breaker config to Service", "target", target.GetName())
	} else if target.GetKind() == "HTTPRoute" && target.GetAPIVersion() == "gateway.networking.k8s.io/v1" {
		annotations := target.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		
		// Add circuit breaker annotations
		annotations["circuitbreaker.io/state"] = "open"
		annotations["circuitbreaker.io/managed-by"] = cb.Name
		
		target.SetAnnotations(annotations)
		
		// Modify HTTPRoute to return 503 or route to fallback
		spec, found, err := unstructured.NestedMap(target.Object, "spec")
		if err != nil || !found {
			return fmt.Errorf("failed to get HTTPRoute spec")
		}
		
		// Store original rules if not already stored
		if _, exists := annotations["circuitbreaker.io/original-rules"]; !exists {
			// In a real implementation, you'd serialize the original rules
			annotations["circuitbreaker.io/original-rules"] = "stored"
		}
		
		// Replace rules with circuit breaker response
		rules := []map[string]interface{}{
			{
				"matches": []map[string]interface{}{
					{"path": map[string]interface{}{"type": "PathPrefix", "value": "/"}},
				},
				"filters": []map[string]interface{}{
					{
						"type": "ResponseHeaderModifier",
						"responseHeaderModifier": map[string]interface{}{
							"add": []map[string]interface{}{
								{"name": "X-Circuit-Breaker", "value": "open"},
							},
						},
					},
				},
				"backendRefs": []map[string]interface{}{
					{
						"name": "circuit-breaker-fallback",
						"port": 80,
					},
				},
			},
		}
		
		spec["rules"] = rules
		unstructured.SetNestedMap(target.Object, spec, "spec")
		target.SetAnnotations(annotations)
		
		log.Info("Applied open circuit breaker config to HTTPRoute", "target", target.GetName())
	}
	
	return r.Update(ctx, target)
}

func (r *CircuitBreakerReconciler) applyClosedConfig(ctx context.Context, target *unstructured.Unstructured, cb *circuitbreakerv1.CircuitBreaker) error {
	log := log.FromContext(ctx)
	
	// For Service, restore original configuration
	if target.GetKind() == "Service" && target.GetAPIVersion() == "v1" {
		annotations := target.GetAnnotations()
		if annotations == nil {
			return nil // Nothing to restore
		}
		
		// Remove circuit breaker annotations
		delete(annotations, "circuitbreaker.io/state")
		delete(annotations, "circuitbreaker.io/managed-by")
		
		target.SetAnnotations(annotations)
		log.Info("Applied closed circuit breaker config to Service", "target", target.GetName())
	} else if target.GetKind() == "HTTPRoute" && target.GetAPIVersion() == "gateway.networking.k8s.io/v1" {
		annotations := target.GetAnnotations()
		if annotations == nil {
			return nil // Nothing to restore
		}
		
		// Remove circuit breaker annotations
		delete(annotations, "circuitbreaker.io/state")
		
		// Restore original rules (simplified - in real implementation, deserialize stored rules)
		if _, exists := annotations["circuitbreaker.io/original-rules"]; exists {
			delete(annotations, "circuitbreaker.io/original-rules")
			
			// In a real implementation, you'd restore the original rules here
			// For now, we'll just remove the circuit breaker modifications
			spec, found, err := unstructured.NestedMap(target.Object, "spec")
			if err == nil && found {
				// This is a simplified restoration - in practice you'd store/restore the actual rules
				unstructured.SetNestedMap(target.Object, spec, "spec")
			}
		}
		
		target.SetAnnotations(annotations)
		log.Info("Applied closed circuit breaker config to HTTPRoute", "target", target.GetName())
	}
	
	return r.Update(ctx, target)
}