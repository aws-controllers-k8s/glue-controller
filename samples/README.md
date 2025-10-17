# AWS Glue Controller Examples

This directory contains example Kubernetes manifests for AWS Glue resources.

## Table of Contents

- [Registry Examples](#registry-examples)
- [Schema Examples](#schema-examples)
- [Multi-Registry Setup](#multi-registry-setup)
- [Schema Evolution](#schema-evolution)

## Registry Examples

### Basic Registry

**File**: [`registry-basic.yaml`](registry-basic.yaml)

Simple registry example demonstrating:
- Basic registry configuration
- Description and deletion policy
- Tag usage for organization

```bash
kubectl apply -f registry-basic.yaml
kubectl get registry my-basic-registry
```

### Registry with Schemas

**File**: [`registry-with-schemas.yaml`](registry-with-schemas.yaml)

Complete example showing:
- Registry with specific compatibility mode
- Schema with cross-resource reference (RegistryRef)
- AVRO schema definition
- Compatibility and deletion policies

```bash
kubectl apply -f registry-with-schemas.yaml
kubectl wait --for=condition=ACK.ResourceSynced registry/production-registry
kubectl get schema user-events-schema
```

## Schema Examples

### JSON Schema Example

**File**: [`schema-json.yaml`](schema-json.yaml)

Demonstrates JSON Schema format with:
- JSON Schema validation syntax
- Required and optional fields
- Pattern validation
- Enum constraints
- FORWARD compatibility mode

**Use Case**: Product catalog events with structured validation

```bash
kubectl apply -f schema-json.yaml
kubectl describe schema product-catalog-json
```

### Protocol Buffers Example

**File**: [`schema-protobuf.yaml`](schema-protobuf.yaml)

Demonstrates Protocol Buffers format with:
- Proto3 syntax
- Message definitions
- Nested types
- Enumerations
- Maps
- BACKWARD_TRANSITIVE compatibility

**Use Case**: Order processing events with strong typing

```bash
kubectl apply -f schema-protobuf.yaml
kubectl get schema order-events-protobuf -o yaml
```

## Multi-Registry Setup

**File**: [`multi-registry.yaml`](multi-registry.yaml)

Comprehensive multi-environment example demonstrating:
- **Environment Separation**: Development, Staging, Production registries
- **Domain Separation**: Analytics registry for specific data domain
- **Different Compatibility Modes**: NONE (dev), BACKWARD (staging), FULL_TRANSITIVE (prod)
- **Deletion Policies**: Delete (dev), Retain (staging/prod)
- **Organization Best Practices**: Naming conventions, labels, tags

**Registries**:
1. `development-registry` - Rapid iteration with NONE compatibility
2. `staging-registry` - Validation with BACKWARD compatibility
3. `production-registry` - Strict safety with FULL_TRANSITIVE
4. `analytics-registry` - Domain-specific for analytics data

**Schemas**:
1. `dev-experiment-schema` → development-registry (NONE compatibility)
2. `prod-transaction-schema` → production-registry (FULL_TRANSITIVE)
3. `analytics-pageview-schema` → analytics-registry (FORWARD)

```bash
# Apply all resources
kubectl apply -f multi-registry.yaml

# Check registries by environment
kubectl get registries -l environment=production

# Check schemas by domain
kubectl get schemas -l domain=analytics

# View registry organization
kubectl get registries --show-labels
```

## Schema Evolution

**File**: [`schema-evolution.yaml`](schema-evolution.yaml)

Demonstrates schema evolution patterns:
- **Initial Schema (v1)**: Basic customer profile fields
- **Evolved Schema (v2)**: Added optional fields with defaults
- **Compatibility Preservation**: BACKWARD compatibility maintained
- **Evolution Best Practices**: Documentation and versioning

**Schema Versions**:
1. `customer-profile-v1`: Initial schema with core fields
2. `customer-profile-v2`: Enhanced with phoneNumber and preferences

**Key Concepts**:
- Adding optional fields with defaults (BACKWARD compatible)
- Version tagging and documentation
- Field-level documentation
- Schema evolution notes and guidelines

```bash
# Apply evolution example
kubectl apply -f schema-evolution.yaml

# Check both versions
kubectl get schemas -l evolution-example=true

# View version 2 details
kubectl get schema customer-profile-v2 -o yaml
```

## Usage Patterns

### Creating Resources

```bash
# Apply single file
kubectl apply -f registry-basic.yaml

# Apply all examples
kubectl apply -f .

# Apply with namespace
kubectl apply -f registry-basic.yaml -n my-namespace
```

### Checking Status

```bash
# List all registries
kubectl get registries

# List all schemas
kubectl get schemas

# Detailed status
kubectl describe registry my-registry
kubectl describe schema my-schema

# JSON/YAML output
kubectl get schema my-schema -o yaml
kubectl get schema my-schema -o jsonpath='{.status.latestVersion}'
```

### Waiting for Resources

```bash
# Wait for registry to be ready
kubectl wait --for=condition=ACK.ResourceSynced registry/my-registry --timeout=60s

# Wait for schema to be ready
kubectl wait --for=condition=ACK.ResourceSynced schema/my-schema --timeout=60s
```

### Cleaning Up

```bash
# Delete specific resource
kubectl delete -f registry-basic.yaml

# Delete all examples
kubectl delete -f .

# Delete by label
kubectl delete schemas -l environment=development
```

## Compatibility Mode Comparison

| Mode | Add Optional Fields | Remove Fields | Change Types | Use Case |
|------|-------------------|---------------|--------------|----------|
| **NONE** | ✅ | ✅ | ✅ | Development only |
| **BACKWARD** | ✅ (with defaults) | ❌ | ❌ | Most common, new code reads old data |
| **FORWARD** | ✅ | ❌ | ❌ | Old code reads new data |
| **FULL** | ✅ (with defaults) | ❌ | ❌ | Both directions, safest |
| **BACKWARD_TRANSITIVE** | ✅ (with defaults) | ❌ | ❌ | Compatible with ALL previous versions |
| **FORWARD_TRANSITIVE** | ✅ | ❌ | ❌ | Compatible with ALL previous versions |
| **FULL_TRANSITIVE** | ✅ (with defaults) | ❌ | ❌ | Production critical data |

## Data Format Support

### AVRO
- Strongly typed
- Compact binary format
- Schema evolution support
- Recommended for high-throughput data pipelines

**Example**: `registry-with-schemas.yaml`

### JSON Schema
- Human-readable
- Web-friendly
- Flexible validation
- Good for APIs and less strict schemas

**Example**: `schema-json.yaml`

### Protocol Buffers
- Efficient serialization
- Strong typing
- Cross-language support
- Best for microservices

**Example**: `schema-protobuf.yaml`

## Best Practices Demonstrated

### 1. Environment Organization
```yaml
# Development: Rapid iteration
deletionPolicy: Delete
compatibility: NONE

# Staging: Validation
deletionPolicy: Retain
compatibility: BACKWARD

# Production: Safety
deletionPolicy: Retain
compatibility: FULL_TRANSITIVE
```

### 2. Naming Conventions
```yaml
# Registries: {environment}-{purpose}-registry
name: production-analytics-registry

# Schemas: {domain}-{entity}-{type}
name: user-profile-events
```

### 3. Labels and Tags
```yaml
metadata:
  labels:
    environment: production
    domain: analytics
    team: data-engineering
spec:
  tags:
    Environment: production
    CriticalData: "true"
```

### 4. Cross-Resource References
```yaml
spec:
  registryRef:
    name: production-registry  # Same namespace
    # namespace: other-ns      # Cross-namespace if needed
```

### 5. Schema Versioning
```yaml
# Document versions in tags
tags:
  SchemaVersion: v2
  LastModified: "2024-01-15"

# Use version in description
description: "Customer profile schema - v2: Added phoneNumber"
```

## Troubleshooting Examples

### Registry Not Ready
```bash
# Check registry status
kubectl get registry my-registry -o yaml | grep -A 10 conditions

# Wait for registry
kubectl wait --for=condition=ACK.ResourceSynced registry/my-registry
```

### Schema Reference Issues
```bash
# Check if registry exists
kubectl get registry production-registry

# Check schema status
kubectl describe schema my-schema | grep -A 5 Conditions
```

### Compatibility Errors
```bash
# View schema events
kubectl get events --field-selector involvedObject.name=my-schema

# Check schema definition
kubectl get schema my-schema -o jsonpath='{.spec.schemaDefinition}'
```

## Additional Resources

- [Schema Registry Documentation](../docs/SCHEMA_REGISTRY.md)
- [AWS Glue Schema Registry](https://docs.aws.amazon.com/glue/latest/dg/schema-registry.html)
- [ACK Documentation](https://aws-controllers-k8s.github.io/community/)

## Contributing

To add new examples:
1. Create YAML file with descriptive name
2. Add comprehensive comments
3. Update this README with description
4. Test example in real cluster
5. Submit pull request
