# Circuit Breaker Samples

This directory contains example configurations for the Circuit Breaker Kubernetes Controller.

## Directory Structure

### 📁 `basic/`
Basic circuit breaker configurations without target integration.
- `sample_circuitbreaker.yaml` - Simple circuit breaker example
- `simple_test.yaml` - Basic test configuration
- `complete_example.yaml` - Complete example with all fields and status

### 📁 `service-integration/`
Circuit breaker configurations with Kubernetes Service integration.
- `enhanced_test.yaml` - Enhanced circuit breaker with Service target
- `enhanced-service.yaml` - Sample Service and Deployment for testing
- `service_test.yaml` - Service integration test configuration
- `production_example.yaml` - Production-ready configuration with rate strategy

### 📁 `gateway-integration/`
Circuit breaker configurations with Gateway API integration.
- `gateway_circuitbreaker.yaml` - Circuit breaker with HTTPRoute target
- `test_httproute.yaml` - Sample HTTPRoute for Gateway API testing
- `advanced_gateway.yaml` - Advanced API gateway configuration with rate strategy

## Quick Start

1. **Basic Circuit Breaker**:
   ```bash
   kubectl apply -f config/samples/basic/
   ```

2. **Service Integration**:
   ```bash
   kubectl apply -f config/samples/service-integration/
   ```

3. **Gateway API Integration** (requires Gateway API CRDs):
   ```bash
   kubectl apply -f config/samples/gateway-integration/
   ```

## Configuration Examples

### Basic Circuit Breaker
```yaml
apiVersion: circuitbreaker.io/v1
kind: CircuitBreaker
metadata:
  name: basic-circuit-breaker
spec:
  failureThreshold: 5
  successThreshold: 2
  timeoutSeconds: 30
  resetTimeout: 60
  strategy:
    mode: count
    window: 1m
```

### Service Integration
```yaml
apiVersion: circuitbreaker.io/v1
kind: CircuitBreaker
metadata:
  name: service-circuit-breaker
spec:
  targetRef:
    apiVersion: v1
    kind: Service
    name: my-service
    namespace: default
  failureThreshold: 3
  successThreshold: 2
  timeoutSeconds: 10
  resetTimeout: 20
  strategy:
    mode: count
    window: 1m
```

### Gateway API Integration
```yaml
apiVersion: circuitbreaker.io/v1
kind: CircuitBreaker
metadata:
  name: gateway-circuit-breaker
spec:
  targetRef:
    apiVersion: gateway.networking.k8s.io/v1
    kind: HTTPRoute
    name: my-route
    namespace: default
  failureThreshold: 5
  successThreshold: 2
  timeoutSeconds: 30
  resetTimeout: 60
  strategy:
    mode: rate
    window: 5m
```