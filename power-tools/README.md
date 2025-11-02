# Power Tools

This directory contains the source code and configurations for various power tools used by the TOE operator.

## Directory Structure

Each power tool has its own directory with the following structure:
```
power-tools/
├── aperf/
│   ├── config/
│   │   └── powertoolconfig-aperf.yaml    # Kubernetes PowerToolConfig
│   ├── Dockerfile                         # Container image definition
│   └── entrypoint.sh                      # Tool entrypoint script
├── chaos/
│   ├── config/
│   │   └── powertoolconfig-chaos.yaml
│   ├── Dockerfile
│   └── *.sh                               # Chaos scripts
├── tcpdump/
│   ├── config/
│   │   └── powertoolconfig-tcpdump.yaml
│   ├── Dockerfile
│   └── entrypoint.sh
└── common/
    └── send-profile.sh                    # Shared utility scripts
```

## Available Power Tools

### aperf
- **Base Image**: Amazon Linux 2023 Minimal
- **Tool**: AWS aperf v0.1.18-alpha (AWS performance profiler)
- **Usage**: Profiles applications using Linux perf via PID attachment
- **Architecture**: Supports x86_64 and aarch64
- **Security**: Requires root access (runAsRoot: true) for kernel perf events
- **Config**: `aperf/config/powertoolconfig-aperf.yaml`

### chaos
- **Base Image**: Alpine Linux
- **Tool**: Custom chaos engineering scripts
- **Usage**: Injects various types of chaos (CPU, memory, network, storage)
- **Config**: `chaos/config/powertoolconfig-chaos.yaml`

### tcpdump
- **Base Image**: Alpine Linux
- **Tool**: tcpdump network packet analyzer
- **Usage**: Captures network traffic from target pods
- **Config**: `tcpdump/config/powertoolconfig-tcpdump.yaml`

## Building Images

```bash
# Build aperf profiler image
docker build -t localhost:32000/codriverlabs/toe/aperf:latest power-tools/aperf/
docker push localhost:32000/codriverlabs/toe/aperf:latest

# Build chaos tool image
docker build -t localhost:32000/codriverlabs/toe/chaos:latest power-tools/chaos/
docker push localhost:32000/codriverlabs/toe/chaos:latest

# Build tcpdump image
docker build -t localhost:32000/codriverlabs/toe/tcpdump:latest power-tools/tcpdump/
docker push localhost:32000/codriverlabs/toe/tcpdump:latest
```

## Deploying PowerToolConfigs

Deploy the PowerToolConfig to make the tool available in your cluster:

```bash
# Deploy aperf configuration
kubectl apply -f power-tools/aperf/config/powertoolconfig-aperf.yaml

# Deploy chaos configuration
kubectl apply -f power-tools/chaos/config/powertoolconfig-chaos.yaml

# Deploy tcpdump configuration
kubectl apply -f power-tools/tcpdump/config/powertoolconfig-tcpdump.yaml
```

## Using Power Tools

After deploying PowerToolConfigs, use PowerTool CRDs to execute tools against target pods.
See `examples/` directory for usage examples.

## Environment Variables

### aperf
- `DURATION`: Profiling duration in seconds (default: 30)
- `PROFILE_TYPE`: Profile type - cpu, memory, etc. (default: cpu)
- `TARGET_PID`: Process ID to profile (default: 1)
- `OUTPUT_PATH`: Output file path (default: /tmp/profile.json)
- `TARGET_CONTAINER_NAME`: Name of the target container

### chaos
- `CHAOS_TYPE`: Type of chaos to inject (cpu, memory, network, storage, process)
- `DURATION`: Duration of chaos injection
- `INTENSITY`: Intensity level of chaos
- `TARGET_CONTAINER_NAME`: Name of the target container

### tcpdump
- `DURATION`: Capture duration
- `INTERFACE`: Network interface to capture (default: eth0)
- `FILTER`: BPF filter expression
- `OUTPUT_PATH`: Output file path
- `TARGET_CONTAINER_NAME`: Name of the target container
