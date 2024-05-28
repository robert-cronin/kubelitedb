# KubeLiteDB

**KubeLiteDB** is an open-source project that provides a Kubernetes Custom Resource Definition (CRD) and operator for managing SQLite instances within a Kubernetes cluster. This project enables users to deploy and manage lightweight, ephemeral SQLite databases easily, leveraging Kubernetes' scalability and orchestration capabilities.

## Features

- **Custom Resource Definition (CRD)**: Define and manage SQLite instances using Kubernetes-native APIs.
- **Lightweight Databases**: Perfect for development, testing, edge computing, and microservices.
- **Ephemeral and Persistent Storage**: Support for both temporary and persistent storage configurations.
- **Easy Deployment**: Simplify database provisioning and management in your Kubernetes clusters.
- **Scalability**: Deploy multiple isolated SQLite instances with ease.
- **Auto Backups**: Configure the SQLite instances to automatically backup using cronjobs.

## Use Cases

- Development and testing environments
- Edge computing and IoT data collection
- Microservices and serverless architectures
- Offline-first applications
- Local analytics and reporting

## Getting Started

1. **Install the CRD**

   ```sh
   kubectl apply -f artifacts/crd.yaml
   ```

2. **Deploy a SQLite Instance**

   ```sh
   kubectl apply -f artifacts/example-sqlite-instance.yaml
   ```

## Contributing

We welcome contributions from the community. Please read our [contributing guide](CONTRIBUTING.md) to get started.

## License

This project is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for more information.