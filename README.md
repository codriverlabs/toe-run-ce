# TOE (Tactical Operations Engine) - Community Edition

A Kubernetes operator for running performance profiling and diagnostic tools on target pods in a secure, controlled manner.

## Description

TOE provides a declarative way to run profiling tools like `perf`, `strace`, and other diagnostic utilities against running pods in Kubernetes clusters. It uses a secure architecture with separate controller and collector components to ensure proper isolation and access control.

## Key Features

- **Declarative Profiling**: Define profiling jobs using Kubernetes custom resources
- **Security First**: Secure token-based authentication between components
- **Multiple Output Modes**: Support for ephemeral, PVC, and collector-based output storage
- **Tool Management**: Centralized configuration of available profiling tools
- **RBAC Integration**: Full Kubernetes RBAC support for access control

## Getting Started

### Prerequisites
- go version v1.24.0+
- docker version 25.0+
- kubectl version v1.33.0+
- Access to a Kubernetes v1.33+ cluster

Check your versions:
```sh
go version
docker --version
kubectl version --client
```

### Quick Install

Install using the pre-built YAML manifests:

```sh
kubectl apply -f https://github.com/codriverlabs/toe-run-ce/releases/latest/download/toe-operator-v1.1.0-public-preview.yaml
```

Or using Helm:

```sh
helm install toe-operator https://github.com/codriverlabs/toe-run-ce/releases/latest/download/toe-operator-v1.1.0-public-preview.tgz
```

### Building from Source

**Build and push your image:**

```sh
make docker-build docker-push IMG=<your-registry>/toe:tag
```

**Install the CRDs:**

```sh
make install
```

**Deploy the operator:**

```sh
make deploy IMG=<your-registry>/toe:tag
```

### Using Private Container Registries

For private registries like ECR, configure image pull secrets:

```sh
# Configure ECR access for target namespace
./helper_scripts/setup-namespace-docker-secrets.sh default 123456789012.dkr.ecr.us-west-2.amazonaws.com us-west-2
```

### Example Usage

Create a PowerTool to profile a running application:

```yaml
apiVersion: codriverlabs.ai.toe.run/v1alpha1
kind: PowerTool
metadata:
  name: profile-my-app
spec:
  targets:
    labelSelector:
      matchLabels:
        app: my-application
  tool:
    name: "aperf"
    duration: "30s"
  output:
    mode: "ephemeral"
```

Apply the configuration:

```sh
kubectl apply -f examples/powertool-aperf-ephemeral.yaml
```

## Architecture

TOE consists of two main components:

- **Controller**: Manages PowerTool resources and orchestrates profiling jobs
- **Collector**: Securely receives and stores profiling data from tools

## Documentation

- [Security Model](docs/security/README.md)
- [TLS Setup](docs/tls-setup.md)
- [Examples](examples/README.md)
- [Roadmap](ROADMAP.md)

## Contributing

We welcome contributions! Please see our contributing guidelines for more information.

## Uninstall

```sh
kubectl delete -k config/samples/
make uninstall
make undeploy
```

## License

See [LICENSE.md](LICENSE.md) for license information.
