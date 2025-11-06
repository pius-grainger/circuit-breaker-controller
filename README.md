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

1. **Build and Load Docker Image**:
   ```bash
   make docker-load
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

```yaml
apiVersion: circuitbreaker.io/v1
kind: CircuitBreaker
metadata:
  name: my-circuit-breaker
spec:
  failureThreshold: 5    # Failures before opening
  timeoutSeconds: 30     # Timeout for half-open state
  resetTimeout: 60       # Time before trying half-open
```

## States

- **Closed**: Normal operation, counting failures
- **Open**: Blocking requests, waiting for reset timeout
- **Half-Open**: Testing if service recovered