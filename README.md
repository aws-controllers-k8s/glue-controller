# ACK service controller for AWS Glue

This repository contains source code for the AWS Controllers for Kubernetes
(ACK) service controller for AWS Glue.

The AWS Glue controller manages AWS Glue resources including:

- **Jobs**: ETL jobs for data processing
- **Schema Registry**: Schema registries for data governance
- **Schemas**: Schema definitions with version management and compatibility enforcement

Please [log issues][ack-issues] and feedback on the main AWS Controllers for
Kubernetes Github project.

[ack-issues]: https://github.com/aws/aws-controllers-k8s/issues

## Features

### AWS Glue Jobs

Manage ETL jobs declaratively with Kubernetes CRDs. Configure job properties,
execution parameters, and monitor job runs directly from Kubernetes.

### AWS Glue Schema Registry

Full support for AWS Glue Schema Registry with advanced features:

- **Declarative Schema Management**: Define schemas as Kubernetes manifests
- **Cross-Resource References**: Link schemas to registries using native Kubernetes references
- **Version Tracking**: Automatic schema version management and status tracking
- **Compatibility Enforcement**: Schema evolution with configurable compatibility rules (BACKWARD, FORWARD, FULL, etc.)
- **Multi-Format Support**: AVRO, JSON Schema, and Protocol Buffers
- **Lifecycle Management**: Pre/post operation hooks for validation and business logic

See [Schema Registry Documentation](docs/SCHEMA_REGISTRY.md) for comprehensive usage guide.

## Quick Start

### Prerequisites

- Kubernetes cluster (1.19+)
- AWS credentials configured
- ACK Glue controller deployed

### Example: Create a Schema Registry

```yaml
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Registry
metadata:
  name: my-registry
spec:
  registryName: my-first-registry
  description: "My first schema registry"
  tags:
    Environment: development
```

### Example: Create a Schema

```yaml
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: user-events
spec:
  schemaName: user-events
  registryRef:
    name: my-registry
  dataFormat: AVRO
  schemaDefinition: |
    {
      "type": "record",
      "name": "UserEvent",
      "fields": [
        {"name": "userId", "type": "string"},
        {"name": "eventType", "type": "string"},
        {"name": "timestamp", "type": "long"}
      ]
    }
  compatibility: BACKWARD
```

## Documentation

- [Schema Registry Guide](docs/SCHEMA_REGISTRY.md) - Comprehensive guide for Schema Registry features
- [Examples](samples/) - Example manifests for common use cases
- [API Reference](apis/v1alpha1/) - CRD API documentation

## Resources

- [AWS Glue Documentation](https://docs.aws.amazon.com/glue/)
- [AWS Glue Schema Registry](https://docs.aws.amazon.com/glue/latest/dg/schema-registry.html)
- [ACK Documentation](https://aws-controllers-k8s.github.io/community/)

## Contributing

We welcome community contributions and pull requests.

See our [contribution guide](/CONTRIBUTING.md) for more information on how to
report issues, set up a development environment, and submit code.

We adhere to the [Amazon Open Source Code of Conduct][coc].

You can also learn more about our [Governance](/GOVERNANCE.md) structure.

[coc]: https://aws.github.io/code-of-conduct

## License

This project is [licensed](/LICENSE) under the Apache-2.0 License.
