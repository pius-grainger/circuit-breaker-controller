# Circuit Breaker Kubernetes Controller

A minimal Kubernetes controller that manages circuit breaker resources.

## Features

- **Three States**: Closed, Open, Half-Open
- **Configurable Thresholds**: Failure count and timeout settings
- **Automatic Recovery**: Transitions between states based on time and failure count

## Quick Start

1. **Install CRD**:
   ```bash
   make install
   ```

2. **Run Controller** (in separate terminal):
   ```bash
   make run
   ```

3. **Create Sample Resource**:
   ```bash
   make sample
   ```

## Helm Installation

1. **Build Docker Image**:
   ```bash
   make docker-build
   ```

2. **Install with Helm**:
   ```bash
   make helm-install
   ```

3. **Uninstall**:
   ```bash
   make helm-uninstall
   ```

## Configuration

### Basic Circuit Breaker
```yaml
apiVersion: circuitbreaker.io/v1
kind: CircuitBreaker
metadata:
  name: my-circuit-breaker
spec:
  failureThreshold: 5    # Failures before opening
  successThreshold: 2    # Successes to close from half-open
  timeoutSeconds: 30     # Timeout for half-open state
  resetTimeout: 60       # Time before trying half-open
  strategy:
    mode: count          # "count" or "rate"
    window: 1m           # Window for rate mode
```

### Gateway API Integration
```yaml
apiVersion: circuitbreaker.io/v1
kind: CircuitBreaker
metadata:
  name: checkout-circuit-breaker
spec:
  targetRef:
    apiVersion: gateway.networking.k8s.io/v1
    kind: HTTPRoute
    name: checkout-route
    namespace: web
  failureThreshold: 5
  successThreshold: 2
  timeoutSeconds: 30
  resetTimeout: 60
  strategy:
    mode: count
    window: 1m
```

## States

- **Closed**: Normal operation, counting failures
- **Open**: Blocking requests, waiting for reset timeout
- **Half-Open**: Testing if service recovered