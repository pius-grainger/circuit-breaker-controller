# Circuit Breaker Kubernetes Controller

A production-ready Kubernetes controller that manages circuit breaker resources with advanced target integration and comprehensive observability.

## Features

- **Three States**: Closed, Open, Half-Open with automatic transitions
- **Target Integration**: Service and HTTPRoute resource management
- **Configurable Thresholds**: Failure and success count settings
- **Strategy Support**: Count-based and rate-based failure detection
- **Production Ready**: Comprehensive status tracking, events, and observability
- **Helm Deployment**: Production-grade Helm chart with HPA, PDB, and security contexts

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

### Using DockerHub Image (Recommended)
```bash
# Install with production-ready configuration
helm install circuit-breaker-controller ./helm/circuit-breaker-controller

# Or with custom values
helm install circuit-breaker-controller ./helm/circuit-breaker-controller \
  --set autoscaling.enabled=true \
  --set podDisruptionBudget.enabled=true
```

### Development Installation
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
  name: basic-example
  annotations:
    description: "Basic circuit breaker without target integration"
spec:
  # Core thresholds
  failureThreshold: 5    # Failures before opening
  successThreshold: 2    # Successes to close from half-open
  
  # Timeouts (seconds)
  timeoutSeconds: 30     # Timeout for half-open state
  resetTimeout: 60       # Time before trying half-open
  
  # Strategy configuration
  strategy:
    mode: count          # "count" or "rate"
    window: 1m           # Window for rate mode (rate mode only)
```

### Service Integration
```yaml
apiVersion: circuitbreaker.io/v1
kind: CircuitBreaker
metadata:
  name: payment-service-cb
  namespace: production
  labels:
    app: payment-service
    environment: production
spec:
  # Target Service resource
  targetRef:
    apiVersion: v1
    kind: Service
    name: payment-service
    namespace: production
  
  # Production-ready thresholds
  failureThreshold: 10     # Higher threshold for production
  successThreshold: 3      # More successes required to close
  timeoutSeconds: 45       # Longer timeout for complex operations
  resetTimeout: 120        # 2 minutes before retry
  
  # Rate-based strategy for high-traffic services
  strategy:
    mode: rate
    window: 5m             # 5-minute window for rate calculation
```

### Gateway API Integration
```yaml
apiVersion: circuitbreaker.io/v1
kind: CircuitBreaker
metadata:
  name: api-gateway-cb
  namespace: gateway-system
  labels:
    component: api-gateway
    tier: edge
spec:
  # Target HTTPRoute resource
  targetRef:
    apiVersion: gateway.networking.k8s.io/v1
    kind: HTTPRoute
    name: api-routes
    namespace: gateway-system
  
  # Edge service configuration
  failureThreshold: 15     # Higher threshold for edge services
  successThreshold: 5      # More conservative recovery
  timeoutSeconds: 60       # Longer timeout for downstream services
  resetTimeout: 300        # 5 minutes before retry
  
  # Rate-based strategy for API gateway
  strategy:
    mode: rate
    window: 10m            # 10-minute window for API rate tracking
```

## State Machine

- **Closed**: Normal operation, counting failures. Transitions to Open when `failureCount >= failureThreshold`
- **Open**: Blocking requests, waiting for reset timeout. Transitions to Half-Open after `resetTimeout` seconds
- **Half-Open**: Testing if service recovered. Transitions to Closed when `successCount >= successThreshold`, or back to Open on any failure

```
┌─────────┐    failureCount >= failureThreshold    ┌──────┐
│ Closed  │ ────────────────────────────────────► │ Open │
│         │                                       │      │
└─────────┘                                       └──────┘
     ▲                                               │
     │                                               │ resetTimeout
     │ successCount >= successThreshold              │ elapsed
     │                                               ▼
┌─────────┐                                    ┌───────────┐
│         │ ◄──── any failure ──────────────── │ Half-Open │
└─────────┘                                    └───────────┘
```

## Strategies

### Count-based Strategy (Default)
```yaml
strategy:
  mode: count
  window: 1m    # Not used in count mode
```
Counts consecutive failures. Simple and predictable for most use cases.

### Rate-based Strategy
```yaml
strategy:
  mode: rate
  window: 5m    # Time window for rate calculation
```
Calculates failure rate within a time window. Better for high-traffic services with variable load.

## Status Fields

The controller maintains comprehensive status information:

```yaml
status:
  state: Closed                    # Current state: Closed/Open/HalfOpen
  failureCount: 0                  # Current failure count
  successCount: 0                  # Current success count (in Half-Open)
  lastFailure: "2025-11-06T12:00:00Z"      # Timestamp of last failure
  lastSuccess: "2025-11-06T11:30:00Z"      # Timestamp of last success
  lastTransitionTime: "2025-11-06T10:00:00Z"  # When state last changed
  reason: "Initialized"            # Reason for current state
  message: "Circuit breaker initialized in closed state"
  targetApplied: false             # Whether target configuration is applied
  conditions:                      # Standard Kubernetes conditions
  - type: "Ready"
    status: "True"
    lastTransitionTime: "2025-11-06T10:00:00Z"
    reason: "CircuitBreakerReady"
    message: "Circuit breaker is ready and operational"
```

## Observability

### Events
The controller emits Kubernetes Events for state transitions:
- `CircuitBreakerOpened`: When failures exceed threshold
- `CircuitBreakerHalfOpened`: When reset timeout elapses
- `CircuitBreakerClosed`: When successes meet threshold

### Monitoring Commands
```bash
# View all circuit breakers with status overview
kubectl get circuitbreakers
# Output:
# NAME       STATE      FAILURES   LASTTRANSITION         AGE
# checkout   HalfOpen   3          2025-11-07T11:45:00Z   5m
# payment    Closed     0                                 2h
# gateway    Open       15         2025-11-07T11:40:00Z   1h

# View circuit breakers across all namespaces
kubectl get circuitbreakers -A

# Watch circuit breaker status changes
kubectl get circuitbreaker my-circuit-breaker -w

# View events for state transitions
kubectl get events --field-selector involvedObject.name=my-circuit-breaker

# Check detailed status and configuration
kubectl describe circuitbreaker my-circuit-breaker

# Watch controller logs for debugging
kubectl logs -l app.kubernetes.io/name=circuit-breaker-controller -f
```

## Helm Chart Configuration

The Helm chart supports production-ready configurations:

```yaml
# values.yaml
image:
  repository: piuschungath/circuit-breaker-controller
  tag: "v3"

# Production settings
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5
  targetCPUUtilizationPercentage: 70

podDisruptionBudget:
  enabled: true
  minAvailable: 1

# Security contexts
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 65532
  fsGroup: 65532

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true

# Resource management
resources:
  limits:
    cpu: 500m
    memory: 128Mi
  requests:
    cpu: 10m
    memory: 64Mi

# Leader election for HA
leaderElection:
  enabled: true
```

## Sample Configurations

The repository includes organized sample configurations:

- **`config/samples/basic/`**: Simple circuit breaker examples
- **`config/samples/service-integration/`**: Service target integration examples
- **`config/samples/gateway-integration/`**: Gateway API HTTPRoute examples

```bash
# Apply basic example
kubectl apply -f config/samples/basic/complete_example.yaml

# Apply production service integration
kubectl apply -f config/samples/service-integration/production_example.yaml

# Apply gateway integration
kubectl apply -f config/samples/gateway-integration/advanced_gateway.yaml
```