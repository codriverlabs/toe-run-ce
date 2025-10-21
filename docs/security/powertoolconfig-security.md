# PowerToolConfig Security Configuration

## Overview

PowerToolConfig CRDs are the exclusive mechanism for defining security contexts for power tools. This document details how security is configured, enforced, and managed through PowerToolConfig resources.

## Security Model

### Administrative Control

Only cluster administrators can create and modify PowerToolConfig resources:

```yaml
# RBAC for PowerToolConfig management
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: powertoolconfig-admin
rules:
- apiGroups: ["codriverlabs.ai.toe.run"]
  resources: ["powertoolconfigs"]
  verbs: ["create", "update", "patch", "delete", "get", "list", "watch"]
```

### User Restrictions

Regular users cannot modify security settings:

```yaml
# RBAC for PowerTool users (no PowerToolConfig access)
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: powertool-user
rules:
- apiGroups: ["codriverlabs.ai.toe.run"]
  resources: ["powertools"]
  verbs: ["create", "update", "patch", "delete", "get", "list", "watch"]
# Note: No powertoolconfigs permissions
```

## Security Configuration

### SecuritySpec Structure

```yaml
apiVersion: codriverlabs.ai.toe.run/v1alpha1
kind: PowerToolConfig
metadata:
  name: aperf-config
  namespace: toe-system
spec:
  name: "aperf"
  image: "localhost:32000/toe/aperf:v1.0.5"
  security:
    # Privilege settings
    allowPrivileged: false      # Never allow privileged containers
    allowHostPID: false         # Prevent host PID namespace access
    
    # Capability management
    capabilities:
      add: ["SYS_PTRACE"]       # Minimal required capabilities
      drop: ["ALL"]             # Drop all default capabilities first
```

### Security Fields

| Field | Purpose | Security Impact | Recommendations |
|-------|---------|-----------------|-----------------|
| `allowPrivileged` | Controls privileged container execution | **HIGH** | Always `false` unless absolutely required |
| `allowHostPID` | Access to host PID namespace | **HIGH** | Always `false` for security isolation |
| `capabilities.add` | Linux capabilities to grant | **MEDIUM** | Minimal set required for tool function |
| `capabilities.drop` | Linux capabilities to remove | **LOW** | Use `["ALL"]` then add specific capabilities |

## Security Profiles

### Minimal Profile (Recommended)

For tools that need basic system access:

```yaml
apiVersion: codriverlabs.ai.toe.run/v1alpha1
kind: PowerToolConfig
metadata:
  name: basic-tool-config
spec:
  name: "basic-tool"
  image: "registry/basic-tool:latest"
  security:
    allowPrivileged: false
    allowHostPID: false
    capabilities:
      drop: ["ALL"]
      add: []  # No additional capabilities
```

### Profiling Profile

For performance profiling tools:

```yaml
apiVersion: codriverlabs.ai.toe.run/v1alpha1
kind: PowerToolConfig
metadata:
  name: profiler-config
spec:
  name: "aperf"
  image: "registry/aperf:latest"
  security:
    allowPrivileged: false
    allowHostPID: false
    capabilities:
      drop: ["ALL"]
      add: ["SYS_PTRACE"]  # Required for process tracing
```

### System Analysis Profile

For tools requiring deeper system access:

```yaml
apiVersion: codriverlabs.ai.toe.run/v1alpha1
kind: PowerToolConfig
metadata:
  name: system-analyzer-config
spec:
  name: "system-analyzer"
  image: "registry/system-analyzer:latest"
  security:
    allowPrivileged: false
    allowHostPID: false
    capabilities:
      drop: ["ALL"]
      add: 
        - "SYS_PTRACE"    # Process tracing
        - "SYS_ADMIN"     # System administration
        - "NET_ADMIN"     # Network administration
```

### High-Privilege Profile (Use with Extreme Caution)

For tools requiring extensive system access:

```yaml
apiVersion: codriverlabs.ai.toe.run/v1alpha1
kind: PowerToolConfig
metadata:
  name: privileged-tool-config
  annotations:
    security.toe.run/risk-level: "high"
    security.toe.run/justification: "Required for kernel debugging"
    security.toe.run/approved-by: "security-team@company.com"
spec:
  name: "kernel-debugger"
  image: "registry/kernel-debugger:latest"
  security:
    allowPrivileged: true   # ⚠️ HIGH RISK
    allowHostPID: true      # ⚠️ HIGH RISK
    capabilities:
      # Note: Privileged containers get all capabilities
```

## Capability Reference

### Common Capabilities for Power Tools

| Capability | Purpose | Risk Level | Use Cases |
|------------|---------|------------|-----------|
| `SYS_PTRACE` | Process tracing and debugging | Medium | Profilers, debuggers |
| `SYS_ADMIN` | System administration tasks | High | System analyzers, network tools |
| `NET_ADMIN` | Network administration | Medium | Network profilers, packet capture |
| `DAC_OVERRIDE` | Bypass file permission checks | High | File system analyzers |
| `SETUID` | Set user ID | High | Identity management tools |
| `SETGID` | Set group ID | High | Identity management tools |

### Dangerous Capabilities (Avoid)

| Capability | Risk | Why Dangerous |
|------------|------|---------------|
| `SYS_MODULE` | Extreme | Can load kernel modules |
| `SYS_RAWIO` | Extreme | Direct hardware access |
| `SYS_BOOT` | Extreme | Can reboot system |
| `MAC_ADMIN` | High | Modify MAC policies |
| `MAC_OVERRIDE` | High | Bypass MAC policies |

## Security Enforcement

### Controller Enforcement

The PowerTool controller enforces security policies:

1. **Lookup Phase**: Controller finds PowerToolConfig by tool name
2. **Validation Phase**: Validates security configuration exists
3. **Application Phase**: Applies ONLY PowerToolConfig security settings
4. **Rejection Phase**: Ignores any security settings in PowerTool

```go
// Controller enforcement logic (simplified)
func (r *PowerToolReconciler) createEphemeralContainer(job *PowerTool) error {
    // Get admin-defined security
    toolConfig := r.getToolConfig(job.Spec.Tool.Name)
    
    // Apply ONLY PowerToolConfig security
    securityContext := toolConfig.Spec.Security
    
    // User security settings are IGNORED
    // job.Spec.Security is not used
    
    return r.createContainer(toolConfig.Spec.Image, securityContext)
}
```

### Admission Control (Optional)

For additional security, implement admission controllers:

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionWebhook
metadata:
  name: powertool-security-validator
webhooks:
- name: validate-powertool-security
  rules:
  - operations: ["CREATE", "UPDATE"]
    apiGroups: ["codriverlabs.ai.toe.run"]
    resources: ["powertools"]
  admissionReviewVersions: ["v1"]
  clientConfig:
    service:
      name: powertool-admission-webhook
      namespace: toe-system
      path: "/validate-powertool"
```

## Security Best Practices

### For Administrators

1. **Principle of Least Privilege**:
   ```yaml
   # Good: Minimal capabilities
   capabilities:
     drop: ["ALL"]
     add: ["SYS_PTRACE"]
   
   # Bad: Excessive capabilities
   capabilities:
     add: ["SYS_ADMIN", "NET_ADMIN", "DAC_OVERRIDE"]
   ```

2. **Regular Security Reviews**:
   ```bash
   # Audit all PowerToolConfigs
   kubectl get powertoolconfigs -A -o yaml | grep -A 10 security:
   
   # Check for privileged tools
   kubectl get powertoolconfigs -A -o jsonpath='{range .items[*]}{.metadata.name}: {.spec.security.allowPrivileged}{"\n"}{end}'
   ```

3. **Documentation and Approval**:
   ```yaml
   metadata:
     annotations:
       security.toe.run/risk-assessment: "completed-2024-01-15"
       security.toe.run/approved-by: "security-team@company.com"
       security.toe.run/review-date: "2024-07-15"
   ```

### For Tool Developers

1. **Minimal Capability Requests**:
   - Document exactly why each capability is needed
   - Provide alternatives with fewer privileges
   - Test with minimal capability sets

2. **Security Documentation**:
   ```yaml
   metadata:
     annotations:
       tool.toe.run/capabilities-required: "SYS_PTRACE for process tracing"
       tool.toe.run/security-impact: "Can read process memory of target containers"
       tool.toe.run/alternatives: "Use non-privileged mode with reduced functionality"
   ```

### For Users

1. **Use Existing Configurations**:
   ```bash
   # List available tools
   kubectl get powertoolconfigs -n toe-system
   
   # Check tool security profile
   kubectl describe powertoolconfig aperf-config -n toe-system
   ```

2. **Request New Tools Properly**:
   - Submit security justification
   - Provide capability requirements
   - Follow approval process

## Monitoring and Auditing

### Security Monitoring

```bash
# Monitor PowerToolConfig changes
kubectl get events --field-selector involvedObject.kind=PowerToolConfig

# Check for privileged tool usage
kubectl get powertoolconfigs -A -o json | jq '.items[] | select(.spec.security.allowPrivileged == true) | .metadata.name'

# Audit capability usage
kubectl get powertoolconfigs -A -o json | jq '.items[] | {name: .metadata.name, capabilities: .spec.security.capabilities.add}'
```

### Compliance Reporting

```bash
# Generate security report
cat << 'EOF' > security-report.sh
#!/bin/bash
echo "PowerToolConfig Security Report - $(date)"
echo "=================================="
echo
echo "Total PowerToolConfigs:"
kubectl get powertoolconfigs -A --no-headers | wc -l
echo
echo "Privileged Tools:"
kubectl get powertoolconfigs -A -o json | jq -r '.items[] | select(.spec.security.allowPrivileged == true) | .metadata.name'
echo
echo "High-Risk Capabilities:"
kubectl get powertoolconfigs -A -o json | jq -r '.items[] | select(.spec.security.capabilities.add[]? | contains("SYS_ADMIN", "SYS_MODULE", "SYS_RAWIO")) | .metadata.name'
EOF
chmod +x security-report.sh
```

## Troubleshooting

### Common Security Issues

1. **Tool Not Found**:
   ```
   Error: PowerToolConfig not found for tool: mytool
   ```
   - Create PowerToolConfig with matching name
   - Check namespace (toe-system recommended)

2. **Insufficient Capabilities**:
   ```
   Error: Operation not permitted
   ```
   - Review tool documentation for required capabilities
   - Add minimal required capabilities to PowerToolConfig

3. **Privileged Access Denied**:
   ```
   Error: Privileged containers are not allowed
   ```
   - Check if tool truly needs privileged access
   - Consider alternative approaches with specific capabilities

### Validation Commands

```bash
# Check PowerToolConfig security
kubectl get powertoolconfig aperf-config -o jsonpath='{.spec.security}'

# Validate capability syntax
kubectl apply --dry-run=client -f powertoolconfig.yaml

# Test tool execution
kubectl apply -f powertool-test.yaml
kubectl logs -f powertool-test
```
