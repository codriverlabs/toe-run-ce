# Collector RBAC Configuration

## Overview

The collector service requires minimal RBAC permissions for token validation and basic Kubernetes API access. It operates with a restricted service account to minimize security exposure.

## Service Account

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: toe-collector
  namespace: toe-system
```

## Role Permissions (Namespace-Scoped)

The collector uses a namespace-scoped Role instead of ClusterRole to limit its permissions.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: toe-system
  name: toe-collector-role
rules:
# Token validation - ServiceAccount token review
- apiGroups: ["authentication.k8s.io"]
  resources: ["tokenreviews"]
  verbs: ["create"]

# ConfigMap access for configuration
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch"]

# Secret access for TLS certificates
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get"]
  resourceNames: ["toe-collector-certs"]
```

## RoleBinding

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: toe-collector-rolebinding
  namespace: toe-system
subjects:
- kind: ServiceAccount
  name: toe-collector
  namespace: toe-system
roleRef:
  kind: Role
  name: toe-collector-role
  apiGroup: rbac.authorization.k8s.io
```

## Complete RBAC Manifest

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: toe-collector
  namespace: toe-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: toe-system
  name: toe-collector-role
rules:
# Token validation
- apiGroups: ["authentication.k8s.io"]
  resources: ["tokenreviews"]
  verbs: ["create"]

# Configuration access
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch"]

# TLS certificate access
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get"]
  resourceNames: ["toe-collector-certs"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: toe-collector-rolebinding
  namespace: toe-system
subjects:
- kind: ServiceAccount
  name: toe-collector
  namespace: toe-system
roleRef:
  kind: Role
  name: toe-collector-role
  apiGroup: rbac.authorization.k8s.io
```

## Security Architecture

### Authentication Flow

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  PowerTool      │    │    Collector     │    │  Kubernetes     │
│  (Client)       │    │    Service       │    │  API Server     │
│                 │    │                  │    │                 │
│ 1. Send Token   │───▶│ 2. Validate      │───▶│ 3. TokenReview  │
│                 │    │    Token         │    │                 │
│                 │    │                  │◀───│ 4. Validation   │
│                 │◀───│ 5. Accept/Reject │    │    Result       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Permission Justification

| Permission | Purpose | Risk Level | Mitigation |
|------------|---------|------------|------------|
| tokenreviews/create | Validate PowerTool tokens | Low | Namespace-scoped, no data access |
| configmaps/get,list,watch | Read collector configuration | Low | Read-only, specific ConfigMaps |
| secrets/get | Access TLS certificates | Medium | Restricted to specific secret name |

## TLS Configuration

### Certificate Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: toe-collector-certs
  namespace: toe-system
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-certificate>
  tls.key: <base64-encoded-private-key>
```

### Certificate Management

The collector requires TLS certificates for secure communication:

1. **Self-Signed Certificates** (Development):
   ```bash
   # Generate self-signed certificate
   openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
     -keyout tls.key -out tls.crt \
     -subj "/CN=toe-collector.toe-system.svc.cluster.local"
   
   # Create secret
   kubectl create secret tls toe-collector-certs \
     --cert=tls.crt --key=tls.key -n toe-system
   ```

2. **cert-manager Integration** (Production):
   ```yaml
   apiVersion: cert-manager.io/v1
   kind: Certificate
   metadata:
     name: toe-collector-cert
     namespace: toe-system
   spec:
     secretName: toe-collector-certs
     dnsNames:
     - toe-collector.toe-system.svc.cluster.local
     issuerRef:
       name: cluster-issuer
       kind: ClusterIssuer
   ```

## Network Security

### Service Configuration

```yaml
apiVersion: v1
kind: Service
metadata:
  name: toe-collector
  namespace: toe-system
spec:
  selector:
    app: toe-collector
  ports:
  - port: 8443
    targetPort: 8443
    protocol: TCP
    name: https
  type: ClusterIP  # Internal access only
```

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: toe-collector-netpol
  namespace: toe-system
spec:
  podSelector:
    matchLabels:
      app: toe-collector
  policyTypes:
  - Ingress
  - Egress
  ingress:
  # Allow PowerTool connections
  - from:
    - namespaceSelector: {}  # All namespaces
    ports:
    - protocol: TCP
      port: 8443
  egress:
  # Allow Kubernetes API access
  - to: []
    ports:
    - protocol: TCP
      port: 443
    - protocol: TCP
      port: 6443
```

## Security Considerations

### Minimal Attack Surface

1. **Namespace Isolation**:
   - Role instead of ClusterRole
   - Limited to toe-system namespace
   - No cross-namespace access

2. **Resource Restrictions**:
   - No PowerTool/PowerToolConfig access
   - No pod manipulation capabilities
   - No secret access beyond TLS certs

3. **Network Isolation**:
   - ClusterIP service (internal only)
   - Network policies restrict traffic
   - TLS-only communication

### Token Security

1. **Token Validation**:
   - Uses Kubernetes TokenReview API
   - Validates token authenticity
   - Checks token expiration

2. **Token Scope**:
   - Tokens are scoped to specific PowerTool
   - Time-limited based on tool duration
   - Cannot be reused across jobs

## Monitoring and Auditing

### Security Events

Monitor for these collector security events:

```bash
# Authentication failures
kubectl logs -n toe-system deployment/toe-collector | grep "authentication failed"

# Unauthorized access attempts
kubectl logs -n toe-system deployment/toe-collector | grep "unauthorized"

# TLS certificate issues
kubectl logs -n toe-system deployment/toe-collector | grep "tls\|certificate"
```

### Audit Queries

```bash
# Check collector permissions
kubectl auth can-i --list --as=system:serviceaccount:toe-system:toe-collector -n toe-system

# Verify token review access
kubectl auth can-i create tokenreviews --as=system:serviceaccount:toe-system:toe-collector

# Check secret access
kubectl auth can-i get secrets/toe-collector-certs --as=system:serviceaccount:toe-system:toe-collector -n toe-system
```

### Health Checks

```bash
# Test collector endpoint
curl -k https://toe-collector.toe-system.svc.cluster.local:8443/health

# Check certificate validity
openssl s_client -connect toe-collector.toe-system.svc.cluster.local:8443 -servername toe-collector.toe-system.svc.cluster.local

# Verify service account token
kubectl get serviceaccount toe-collector -n toe-system -o yaml
```

## Troubleshooting

### Common Issues

1. **Token Validation Failures**:
   ```
   Error: failed to validate token: tokenreviews.authentication.k8s.io is forbidden
   ```
   - Check Role includes tokenreviews create permission
   - Verify RoleBinding is correct

2. **TLS Certificate Issues**:
   ```
   Error: failed to load TLS certificate: secret "toe-collector-certs" not found
   ```
   - Ensure secret exists in toe-system namespace
   - Check secret has correct tls.crt and tls.key data

3. **ConfigMap Access Denied**:
   ```
   Error: configmaps "tools-configuration" is forbidden
   ```
   - Verify Role includes configmaps get permission
   - Check RoleBinding namespace matches

### Validation Commands

```bash
# Test collector RBAC
kubectl auth can-i create tokenreviews --as=system:serviceaccount:toe-system:toe-collector
kubectl auth can-i get configmaps --as=system:serviceaccount:toe-system:toe-collector -n toe-system
kubectl auth can-i get secrets/toe-collector-certs --as=system:serviceaccount:toe-system:toe-collector -n toe-system

# Check effective permissions
kubectl describe rolebinding toe-collector-rolebinding -n toe-system
kubectl describe role toe-collector-role -n toe-system
```
