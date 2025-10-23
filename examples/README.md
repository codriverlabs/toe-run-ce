# PowerTool Examples

This directory contains example configurations for PowerTool and PowerToolConfig CRDs.

## PowerToolConfig Examples

PowerToolConfig defines the configuration for individual power tools:

- `powertoolconfig-aperf.yaml` - Configuration for the aperf performance profiling tool

## PowerTool Examples

PowerTool defines the execution of power tools against target pods:

### Basic Examples
- `powertool-aperf-ephemeral.yaml` - Simple ephemeral execution
- `powertool-aperf-pvc.yaml` - Output to persistent volume
- `powertool-collector.yaml` - Output to collector service

### Advanced Examples
- `powertool-scheduled.yaml` - Scheduled execution with cron
- `powertool-conflict-test.yaml` - Conflict resolution testing
- `test-token-config.yaml` - Custom token configuration

## Legacy Examples (Deprecated)

The following files use the old ProfileJob API and have been removed:
- `aperf-ephemeral-example.yaml` (replaced by `powertool-aperf-ephemeral.yaml`)
- `aperf-pvc-example.yaml` (replaced by `powertool-aperf-pvc.yaml`)
- `conflict-test-example.yaml` (replaced by `powertool-conflict-test.yaml`)

## Usage

1. **Deploy PowerToolConfig first:**
   ```bash
   kubectl apply -f powertoolconfig-aperf.yaml
   ```

2. **Deploy PowerTool:**
   ```bash
   kubectl apply -f powertool-aperf-ephemeral.yaml
   ```

3. **Check status:**
   ```bash
   kubectl get powertools
   kubectl describe powertool aperf-ephemeral-example
   ```

## Migration Guide

To migrate from ProfileJob to PowerTool:

1. Change `apiVersion` from `toe.run/v1alpha1` to `codriverlabs.ai.toe.run/v1alpha1`
2. Change `kind` from `ProfileJob` to `PowerTool`
3. Rename `spec.profiler` to `spec.tool`
4. Update `spec.tool.tool` to `spec.tool.name`
5. Create corresponding PowerToolConfig for tool definitions
