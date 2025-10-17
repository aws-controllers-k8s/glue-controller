# End-to-End (E2E) Tests for Glue Controller

This directory contains end-to-end tests for the AWS Glue Controller, specifically testing the Schema resource lifecycle and cross-resource references.

## Overview

The E2E tests validate:
- Complete resource lifecycle (create, update, delete)
- Registry reference resolution
- Schema versioning and compatibility
- Status updates and condition management
- Error handling and recovery scenarios

## Prerequisites

### Option 1: Using envtest (Recommended for CI/CD)

1. **Install kubebuilder and envtest binaries**:
   ```bash
   # Install kubebuilder
   os=$(go env GOOS)
   arch=$(go env GOARCH)
   curl -L https://go.kubebuilder.io/dl/latest/${os}/${arch} | tar -xz -C /tmp/
   sudo mv /tmp/kubebuilder_* /usr/local/kubebuilder
   export PATH=$PATH:/usr/local/kubebuilder/bin

   # Setup envtest
   make envtest
   ```

2. **Configure test environment**:
   ```bash
   export KUBEBUILDER_ASSETS="$(setup-envtest use -p path)"
   ```

### Option 2: Using a Real Kubernetes Cluster

1. **Setup test cluster**:
   ```bash
   # Using kind
   kind create cluster --name glue-e2e-test

   # Or using minikube
   minikube start --profile glue-e2e-test
   ```

2. **Install CRDs**:
   ```bash
   make install  # Install Schema and Registry CRDs
   ```

3. **Deploy controller** (optional - for real AWS testing):
   ```bash
   make deploy
   ```

## Running Tests

### Quick Run (skips E2E tests)

```bash
# Run unit and integration tests only
go test ./... -short
```

### Full E2E Test Suite

```bash
# Run all E2E tests
go test ./test/e2e/... -v -timeout 30m

# Run specific test
go test ./test/e2e -v -run TestSchemaLifecycle

# Run with parallel execution
go test ./test/e2e -v -parallel 4
```

### With envtest

```bash
# Set envtest assets path
export KUBEBUILDER_ASSETS="$(setup-envtest use -p path)"

# Run E2E tests
go test ./test/e2e -v
```

### With Real Cluster

```bash
# Ensure KUBECONFIG is set
export KUBECONFIG=~/.kube/config

# Run E2E tests
go test ./test/e2e -v -test.v
```

## Test Structure

### TestSchemaLifecycle

Tests the complete lifecycle of Schema resources:

1. **CreateSchemaWithDefaultRegistry**: Creates a schema using the default AWS Glue registry
2. **CreateSchemaWithRegistryReference**: Creates a schema with explicit registry reference
3. **UpdateSchemaDefinition**: Updates schema definition to create a new version
4. **AttemptImmutableFieldUpdate**: Validates immutable field protection
5. **DeleteSchema**: Tests schema deletion

### TestSchemaReferenceResolution

Tests registry reference resolution scenarios:

1. **SchemaBeforeRegistryReady**: Schema created before referenced registry is ready
2. **MultipleSchemasOneRegistry**: Multiple schemas referencing the same registry

### TestSchemaVersioning

Tests schema version management:

1. **MultipleVersionsWithCompatibility**: Creates multiple schema versions with compatibility checks

## Test Scenarios Covered

### ✅ Happy Path Scenarios
- Schema creation with default registry
- Schema creation with explicit registry reference
- Schema updates creating new versions
- Multiple schemas sharing a registry
- Schema deletion

### ✅ Error Scenarios
- Registry reference not found
- Registry not ready (no metadata)
- Registry not ready (no ARN)
- Registry in terminal state
- Immutable field update attempts

### ✅ Edge Cases
- Schema created before registry becomes ready
- Concurrent schema creation
- Schema version compatibility validation

## Test Configuration

### Environment Variables

- `KUBEBUILDER_ASSETS`: Path to envtest binaries
- `KUBECONFIG`: Path to kubeconfig file for real cluster
- `TEST_TIMEOUT`: Override default test timeout (default: 5m)
- `AWS_REGION`: AWS region for testing (default: us-east-1)

### Test Flags

```bash
# Skip long-running tests
go test ./test/e2e -short

# Verbose output
go test ./test/e2e -v

# Run specific test
go test ./test/e2e -run TestSchemaLifecycle

# Parallel execution
go test ./test/e2e -parallel 4

# Custom timeout
go test ./test/e2e -timeout 30m
```

## Mock vs Real AWS Testing

### Mock Mode (Default)

By default, tests run with mocked AWS SDK clients:
- No actual AWS resources are created
- No AWS credentials required
- Fast execution
- Suitable for CI/CD pipelines

### Real AWS Mode (Optional)

For testing against real AWS Glue:

1. **Configure AWS credentials**:
   ```bash
   export AWS_ACCESS_KEY_ID=your-key
   export AWS_SECRET_ACCESS_KEY=your-secret
   export AWS_REGION=us-east-1
   ```

2. **Run with real AWS flag**:
   ```bash
   go test ./test/e2e -v -tags=integration
   ```

3. **Cleanup** (important!):
   ```bash
   # Tests should cleanup automatically, but verify
   aws glue list-schemas --registry-name test-registry
   ```

## Troubleshooting

### Tests Skipped

If you see "Skipping E2E test" messages:
- Check that envtest binaries are installed
- Verify `KUBEBUILDER_ASSETS` is set
- Or provide a valid `KUBECONFIG` for a test cluster

### Tests Timeout

If tests timeout:
- Increase timeout: `go test ./test/e2e -timeout 30m`
- Check cluster connectivity
- Verify controller is running (for real cluster tests)
- Check logs: `kubectl logs -n glue-system deployment/glue-controller`

### Resource Cleanup Issues

If resources aren't cleaned up:
```bash
# Manual cleanup
kubectl delete schemas --all -n glue-e2e-test
kubectl delete registries --all -n glue-e2e-test
kubectl delete namespace glue-e2e-test
```

### CRD Not Found Errors

```bash
# Reinstall CRDs
make install
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: E2E Tests
on: [push, pull_request]

jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Setup envtest
        run: |
          go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
          echo "KUBEBUILDER_ASSETS=$(setup-envtest use -p path)" >> $GITHUB_ENV

      - name: Run E2E Tests
        run: go test ./test/e2e -v -timeout 30m
```

## Best Practices

1. **Use `-short` flag** for quick feedback during development
2. **Run full E2E suite** before submitting PRs
3. **Check test coverage**: `go test ./test/e2e -cover`
4. **Parallel execution** for faster runs: `-parallel 4`
5. **Cleanup verification**: Always check resources are cleaned up
6. **Logs collection**: Enable verbose mode `-v` when debugging

## Future Enhancements

- [ ] Add webhook validation tests
- [ ] Add performance/load testing scenarios
- [ ] Add chaos engineering tests
- [ ] Add multi-region testing
- [ ] Add AWS service limit testing
- [ ] Add cost estimation tests
- [ ] Add compliance/security tests
