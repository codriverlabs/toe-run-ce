# Security Documentation

This directory contains comprehensive security documentation for the toe-k8s-operator system.

## Files Overview

- `security-model.md` - Overall security architecture and model
- `controller-rbac.md` - RBAC permissions for the PowerTool controller
- `collector-rbac.md` - RBAC permissions for the collector service
- `powertoolconfig-security.md` - Security configuration through PowerToolConfig CRDs
- `powertool-rbac-roles.md` - Built-in RBAC roles for PowerTool resources
- `rbac-deployment.md` - Advanced RBAC deployment patterns and examples

## Security Principles

1. **Least Privilege** - Components have minimal required permissions
2. **Separation of Concerns** - Security configuration separated from execution
3. **Administrative Control** - Only admins can define security contexts
4. **Defense in Depth** - Multiple layers of security controls
5. **Role-Based Access** - Granular permissions through built-in RBAC roles

## Quick Reference

| Component | Security Control | File |
|-----------|------------------|------|
| PowerTool Controller | RBAC, Service Account | `controller-rbac.md` |
| Collector Service | RBAC, TLS, Authentication | `collector-rbac.md` |
| Tool Security | Capabilities, Privileges | `powertoolconfig-security.md` |
| User Access Control | Built-in RBAC Roles | `powertool-rbac-roles.md` |
| Advanced RBAC | Custom Patterns, Examples | `rbac-deployment.md` |
| Overall Model | Architecture, Flow | `security-model.md` |

## RBAC Quick Start

For most deployments, use the built-in RBAC roles:

- **Admin Role**: `powertool-admin-role` - Full access to PowerTool and PowerToolConfig
- **Editor Role**: `powertool-editor-role` - Can create profiling jobs, read-only tool configs  
- **Viewer Role**: `powertool-viewer-role` - Read-only access for monitoring and auditing

See `powertool-rbac-roles.md` for detailed usage examples.
