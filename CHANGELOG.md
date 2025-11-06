# Changelog

All notable changes to the Circuit Breaker Kubernetes Controller will be documented in this file.

## [v2.0.0] - 2025-11-06

### 🚀 Major Features Added
- **Target Integration**: Added support for Service and HTTPRoute target references
- **Enhanced State Management**: Complete circuit breaker state machine (Closed → Open → Half-Open → Closed)
- **Success Threshold**: Added configurable success threshold for Half-Open → Closed transitions
- **Strategy Configuration**: Added strategy modes (count/rate) with time windows
- **Target Configuration Management**: Automatic annotation management for target resources

### 🎨 Enhanced User Experience
- **Emoji-Enhanced Logging**: Production-ready logs with visual indicators (🔴 🟡 🟢)
- **Comprehensive Status Tracking**: Added successCount, targetApplied, and lastSuccess fields
- **Detailed State Transition Logs**: Clear logging for all state changes with context

### 🔧 Technical Improvements
- **Enhanced CRD Schema**: Complete status fields with proper validation
- **Target Handler Architecture**: Extensible pattern for multiple target types
- **RBAC Enhancements**: Added permissions for Services, HTTPRoutes, and events
- **Helm Chart Improvements**: Complete deployment with proper RBAC and CRD management

### 📦 Deployment & Distribution
- **Docker Hub Integration**: Added build and push scripts for public distribution
- **Helm Chart**: Production-ready Helm chart with configurable values
- **Sample Configurations**: Multiple example configurations for different use cases

### 🧪 Testing & Validation
- **State Transition Testing**: Validated complete circuit breaker lifecycle
- **Target Integration Testing**: Verified Service and HTTPRoute configuration management
- **Production Deployment**: Successfully tested on minikube with real Kubernetes resources

### 📋 Configuration Examples Added
- Enhanced test configurations with Service targets
- Gateway API HTTPRoute integration examples
- Multiple sample circuit breaker configurations

### 🔄 Breaking Changes
- Updated CRD schema with new status fields (requires CRD update)
- Enhanced controller logic with target integration (backward compatible)

### 🐛 Bug Fixes
- Fixed CRD status field definitions for successCount and targetApplied
- Resolved RBAC permissions for target resource management
- Corrected Helm image configuration format

---

## [v1.0.0] - 2025-11-05

### 🎉 Initial Release
- Basic circuit breaker functionality
- Simple state management (Closed/Open)
- Basic CRD and controller implementation
- Minimal logging and status tracking

---

**Migration Guide**: To upgrade from v1.0.0 to v2.0.0, update the CRD and redeploy the controller. Existing circuit breaker resources will be automatically migrated to the new schema.