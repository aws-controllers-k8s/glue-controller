# AWS Glue Schema Registry Implementation

Comprehensive guide for managing AWS Glue Schema Registry resources using Kubernetes CRDs.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Resources](#resources)
4. [Features](#features)
5. [Quick Start](#quick-start)
6. [Usage Examples](#usage-examples)
7. [Schema Compatibility](#schema-compatibility)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

## Overview

The AWS Glue Schema Registry implementation provides first-class Kubernetes CRD support for managing schema registries and schemas. This enables declarative schema management with full Kubernetes integration including:

- **Declarative Management**: Define schemas as Kubernetes manifests
- **Cross-Resource References**: Link schemas to registries using CRD references
- **Version Tracking**: Automatic schema version management and tracking
- **Compatibility Enforcement**: Schema evolution with compatibility rules
- **Status Management**: Real-time status updates from AWS
- **Lifecycle Hooks**: Pre/post operation hooks for validation and business logic

## Architecture

### Resource Hierarchy

```
Registry (CRD)
  ├── RegistryARN (AWS Resource)
  └── Schemas (CRD References)
       ├── Schema 1 → SchemaARN (AWS Resource)
       │   └── Versions (AWS Managed)
       ├── Schema 2 → SchemaARN (AWS Resource)
       │   └── Versions (AWS Managed)
       └── Schema N → SchemaARN (AWS Resource)
           └── Versions (AWS Managed)
```

### Controller Components

```
┌─────────────────────────────────────────────────────────┐
│                   Kubernetes API Server                  │
└───────────────────┬─────────────────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
┌───────▼──────┐       ┌───────▼──────┐
│   Registry   │       │    Schema    │
│   Controller │       │   Controller │
└───────┬──────┘       └───────┬──────┘
        │                      │
        │  ┌───────────────────┘
        │  │
        │  │  ┌──────────────────────┐
        │  │  │  Reference Resolver  │
        │  │  │  - Registry lookup   │
        │  │  │  - ARN resolution    │
        │  │  │  - State validation  │
        │  │  └──────────────────────┘
        │  │
        ▼  ▼
┌────────────────────────────────┐
│     AWS Glue API (SDK v2)      │
│  - CreateRegistry              │
│  - CreateSchema                │
│  - GetSchemaByDefinition       │
│  - RegisterSchemaVersion       │
│  - GetSchemaVersion            │
└────────────────────────────────┘
```

## Resources

### Registry CRD

Manages AWS Glue Schema Registry resources.

**Spec Fields**:
- `registryName` (string, required): Name of the registry
- `description` (string, optional): Description of the registry
- `deletionPolicy` (string, optional): Delete or Retain (default: Delete)
- `tags` (map[string]string, optional): Resource tags

**Status Fields**:
- `registryARN` (string): AWS ARN of the registry
- `status` (string): Current status (AVAILABLE, DELETING)
- `ackResourceMetadata`: Standard ACK metadata
- `conditions`: Standard Kubernetes conditions

### Schema CRD

Manages AWS Glue Schema resources within registries.

**Spec Fields**:
- `schemaName` (string, required): Name of the schema
- `registryRef` (object, required): Reference to Registry CRD
  - `name` (string): Name of the Registry CRD
  - `namespace` (string, optional): Namespace (defaults to same namespace)
- `dataFormat` (string, required): AVRO, JSON, or PROTOBUF
- `schemaDefinition` (string, required): Schema definition in the specified format
- `description` (string, optional): Description of the schema
- `compatibility` (string, optional): Compatibility mode
  - NONE, DISABLED, BACKWARD, BACKWARD_TRANSITIVE
  - FORWARD, FORWARD_TRANSITIVE, FULL, FULL_TRANSITIVE
- `deletionPolicy` (string, optional): Delete or Retain (default: Delete)
- `tags` (map[string]string, optional): Resource tags

**Status Fields**:
- `schemaARN` (string): AWS ARN of the schema
- `registryARN` (string): Resolved ARN from RegistryRef
- `latestVersion` (object): Latest schema version metadata
  - `versionNumber` (int64): Version number
  - `schemaVersionID` (string): Version ID
  - `status` (string): Version status (AVAILABLE, PENDING, FAILURE, DELETING)
  - `createdTime` (time): Version creation timestamp
- `schemaVersionID` (string): Current schema version ID
- `schemaVersionStatus` (string): Current version status
- `ackResourceMetadata`: Standard ACK metadata
- `conditions`: Standard Kubernetes conditions including:
  - `ACK.ResourceSynced`: Resource synchronized with AWS
  - `ACK.Terminal`: Terminal error state
  - `VersionPending`: Schema version creation pending
  - `ReferenceNotResolved`: RegistryRef resolution pending

## Features

### 1. Cross-Resource References

Schemas reference registries using Kubernetes-native references:

```yaml
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: my-schema
spec:
  schemaName: my-schema
  registryRef:
    name: my-registry  # References Registry CRD in same namespace
  dataFormat: AVRO
  schemaDefinition: |
    {...}
```

**Reference Resolution**:
- Validates Registry CRD exists and is ready
- Resolves Registry ARN from status
- Handles Registry not ready scenarios with requeue
- Detects terminal states (deletion, errors)

### 2. Schema Version Management

The controller automatically tracks schema versions:

```yaml
status:
  latestVersion:
    versionNumber: 3
    schemaVersionID: "a1b2c3d4-..."
    status: AVAILABLE
    createdTime: "2024-01-15T10:30:00Z"
```

**Version Tracking**:
- Automatic version number tracking
- Version status monitoring (AVAILABLE, PENDING, FAILURE)
- Creation timestamp recording
- Version ID storage for AWS API operations

### 3. Compatibility Enforcement

Schema evolution with compatibility rules:

| Mode | Description | Add Fields | Remove Fields | Change Types |
|------|-------------|------------|---------------|--------------|
| NONE | No checks | ✅ | ✅ | ✅ |
| BACKWARD | New can read old | ✅ (with defaults) | ❌ | ❌ |
| FORWARD | Old can read new | ✅ | ❌ | ❌ |
| FULL | Both directions | ✅ (with defaults) | ❌ | ❌ |
| BACKWARD_TRANSITIVE | All previous versions | ✅ (with defaults) | ❌ | ❌ |
| FORWARD_TRANSITIVE | All previous versions | ✅ | ❌ | ❌ |
| FULL_TRANSITIVE | All previous versions | ✅ (with defaults) | ❌ | ❌ |

### 4. Lifecycle Hooks

Pre and post operation hooks for each CRUD operation:

**Create Hooks**:
- `preCreate`: Validate schema definition before creation
- `postCreate`: Log success, track initial version

**Update Hooks**:
- `preUpdate`: Detect immutable field changes, validate compatibility
- `postUpdate`: Track version changes, log evolution

**Delete Hooks**:
- `preDelete`: Warn about data loss, validate deletion policy
- `postDelete`: Cleanup status, log deletion

### 5. Status Management

Comprehensive status tracking:

**Readiness Determination**:
```go
// Schema is ready if:
// 1. Has SchemaARN
// 2. Has LatestVersion information
// 3. LatestVersion status is AVAILABLE
```

**Custom Conditions**:
- `VersionPending`: Schema version creation in progress
- `ReferenceNotResolved`: Registry reference not yet resolved

**AWS Metadata**:
- Region tracking
- Account ID tracking
- ARN updates

## Quick Start

### Prerequisites

1. AWS Glue Schema Registry enabled in your AWS account
2. ACK Glue controller deployed in Kubernetes cluster
3. IAM permissions for Glue Schema Registry operations

### Step 1: Create a Registry

```bash
kubectl apply -f - <<EOF
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Registry
metadata:
  name: my-registry
spec:
  registryName: my-first-registry
  description: "My first schema registry"
  deletionPolicy: Delete
  tags:
    Environment: development
    Team: data-engineering
EOF
```

### Step 2: Wait for Registry to be Ready

```bash
kubectl wait --for=condition=ACK.ResourceSynced registry/my-registry --timeout=60s
```

### Step 3: Create a Schema

```bash
kubectl apply -f - <<EOF
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
  deletionPolicy: Delete
EOF
```

### Step 4: Check Schema Status

```bash
kubectl get schema user-events -o yaml
```

Expected status:
```yaml
status:
  schemaARN: arn:aws:glue:us-east-1:123456789012:schema/my-first-registry/user-events
  registryARN: arn:aws:glue:us-east-1:123456789012:registry/my-first-registry
  latestVersion:
    versionNumber: 1
    schemaVersionID: "uuid-here"
    status: AVAILABLE
    createdTime: "2024-01-15T10:30:00Z"
  conditions:
  - type: ACK.ResourceSynced
    status: "True"
```

## Usage Examples

### Example 1: Multi-Environment Setup

See `samples/multi-registry.yaml` for complete example with:
- Separate registries for development, staging, production
- Different compatibility modes per environment
- Domain-specific registries (analytics, transactions)

### Example 2: Schema Evolution

See `samples/schema-evolution.yaml` for:
- Initial schema version (v1)
- Evolved schema with new fields (v2)
- Compatibility preservation patterns

### Example 3: Data Format Examples

See individual files for format-specific examples:
- `samples/schema-json.yaml`: JSON Schema with validation
- `samples/schema-protobuf.yaml`: Protocol Buffers
- `samples/registry-with-schemas.yaml`: AVRO format

## Schema Compatibility

### Compatibility Rules

**BACKWARD** (recommended for most use cases):
```yaml
# Version 1
fields: [userId, email]

# Version 2 - BACKWARD compatible
fields: [userId, email, phoneNumber (optional with default)]
```

**FORWARD** (for consumer upgrades before producers):
```yaml
# Version 1
fields: [userId, email, phoneNumber]

# Version 2 - FORWARD compatible
fields: [userId, email]  # phoneNumber removed
```

**FULL** (most restrictive, safest):
```yaml
# Only optional fields with defaults can be added
# No fields can be removed
# No type changes
```

### Compatibility Best Practices

1. **Start with BACKWARD** for most applications
2. **Use FULL_TRANSITIVE** for critical production data
3. **Use NONE only in development**
4. **Test compatibility** before promoting to production
5. **Document breaking changes** clearly

### Handling Incompatible Changes

When you must make breaking changes:

1. **Create new schema** with different name:
```yaml
# Old schema: user-events-v1
# New schema: user-events-v2
```

2. **Run dual-write** period:
   - Write to both old and new schemas
   - Update consumers gradually
   - Deprecate old schema

3. **Use DeletionPolicy: Retain**:
```yaml
spec:
  deletionPolicy: Retain  # Keep schema in AWS even after CRD deletion
```

## Best Practices

### 1. Registry Organization

**By Environment**:
```yaml
# development-registry: NONE compatibility for rapid iteration
# staging-registry: BACKWARD compatibility for validation
# production-registry: FULL_TRANSITIVE compatibility for safety
```

**By Domain**:
```yaml
# analytics-registry: Analytics and metrics data
# transactions-registry: Financial transactions
# events-registry: Application events
```

**By Team**:
```yaml
# data-engineering-registry: Team-specific schemas
# platform-registry: Shared schemas
```

### 2. Naming Conventions

- **Registries**: `{environment}-{purpose}-registry`
  - Example: `production-analytics-registry`

- **Schemas**: `{domain}-{entity}-{events/data}`
  - Example: `user-profile-events`
  - Example: `order-transaction-data`

- **Labels**: Use consistent labels for organization
```yaml
metadata:
  labels:
    environment: production
    domain: analytics
    team: data-engineering
```

### 3. Deletion Policies

- **Development**: `Delete` (ephemeral resources)
- **Staging**: `Retain` (debugging and validation)
- **Production**: `Retain` (never lose schemas)

### 4. Compatibility Settings

- **Development**: `NONE` or `BACKWARD`
- **Staging**: `BACKWARD` or `FULL`
- **Production**: `FULL_TRANSITIVE`

### 5. Version Management

- Always include version information in tags:
```yaml
tags:
  SchemaVersion: v2
  LastModified: "2024-01-15"
```

- Document breaking changes in description:
```yaml
description: "User profile schema - v2: Added phoneNumber and preferences"
```

### 6. Reference Management

- Always use same namespace for related resources:
```yaml
# Registry in 'default' namespace
# Schemas referencing it also in 'default'
```

- Check registry readiness before creating schemas
- Handle reference resolution errors gracefully

## Troubleshooting

### Common Issues

#### 1. Registry Not Ready

**Symptom**:
```
Schema status shows: ReferenceNotResolved condition
```

**Solution**:
```bash
# Check registry status
kubectl get registry my-registry -o yaml

# Wait for registry to be ready
kubectl wait --for=condition=ACK.ResourceSynced registry/my-registry
```

#### 2. Schema Version Pending

**Symptom**:
```yaml
status:
  latestVersion:
    status: PENDING
  conditions:
  - type: VersionPending
    status: "True"
```

**Solution**:
- Wait for AWS to process schema version
- Check AWS Glue console for errors
- Verify schema definition is valid

#### 3. Compatibility Errors

**Symptom**:
```
Schema update rejected due to compatibility violation
```

**Solution**:
1. Review compatibility mode
2. Check what changed in schema definition
3. Add defaults to new fields
4. Consider creating new schema version

#### 4. Cross-Namespace References

**Symptom**:
```
Registry not found in namespace
```

**Solution**:
```yaml
spec:
  registryRef:
    name: my-registry
    namespace: other-namespace  # Explicitly specify namespace
```

### Debugging Commands

```bash
# Check all registries
kubectl get registries

# Check all schemas
kubectl get schemas

# Describe schema for events
kubectl describe schema user-events

# View schema status
kubectl get schema user-events -o jsonpath='{.status}'

# Check controller logs
kubectl logs -n ack-system deployment/ack-glue-controller

# Watch for events
kubectl get events --sort-by='.lastTimestamp' | grep -E 'registry|schema'
```

### Status Conditions Reference

| Condition Type | Status | Meaning |
|---------------|--------|---------|
| ACK.ResourceSynced | True | Resource synchronized with AWS |
| ACK.ResourceSynced | False | Sync pending or failed |
| ACK.Terminal | True | Unrecoverable error |
| VersionPending | True | Schema version creation in progress |
| ReferenceNotResolved | True | Registry reference not yet resolved |

### Error Messages

**"referenced Registry X is not ready yet (no metadata)"**
- Registry CRD exists but AWS resource not created yet
- Wait for registry to complete creation

**"referenced Registry X is not ready yet (no ARN assigned)"**
- Registry created but ARN not yet available
- Normal during initial creation, will auto-resolve

**"referenced Registry X is in a terminal error state"**
- Registry creation failed permanently
- Check registry status and conditions
- Fix registry issues before creating schemas

**"referenced Registry X is being deleted"**
- Registry has deletion timestamp
- Cannot create schemas for deleting registries
- Create new registry or use different one

## Additional Resources

- [AWS Glue Schema Registry Documentation](https://docs.aws.amazon.com/glue/latest/dg/schema-registry.html)
- [ACK Documentation](https://aws-controllers-k8s.github.io/community/)
- [Example Manifests](../samples/)
- [API Reference](../apis/v1alpha1/)

## Support

For issues and questions:
- GitHub Issues: [github.com/aws-controllers-k8s/glue-controller](https://github.com/aws-controllers-k8s/glue-controller)
- ACK Community: [Slack](https://aws-controllers-k8s.slack.com)
