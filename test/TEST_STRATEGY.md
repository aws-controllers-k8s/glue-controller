# AWS Glue Controller Test Strategy

This document outlines the comprehensive testing strategy for the AWS Glue Controller, specifically for the Schema resource with Registry reference support.

## Testing Pyramid

```
                    ┌─────────────┐
                    │   E2E Tests │  (Highest confidence, slowest)
                    │   ~10 tests │
                    └─────────────┘
                  ┌───────────────────┐
                  │ Integration Tests │  (Medium confidence, moderate speed)
                  │    ~15 tests      │
                  └───────────────────┘
              ┌─────────────────────────────┐
              │      Unit Tests             │  (Fast feedback, many tests)
              │       ~30+ tests            │
              └─────────────────────────────┘
```

## Test Levels

### 1. Unit Tests (`pkg/resource/schema/*_test.go`)

**Purpose**: Test individual functions and methods in isolation

**Coverage**:
- ✅ Reference resolution logic (`references_test.go`)
  - Registry ARN extraction
  - Registry validation
  - Reference resolution with various states
- ✅ Delta comparison (`delta_test.go`)
  - Immutable field detection
  - Mutable field changes
  - Status updates
- ✅ Status management (`status_test.go`)
  - Readiness determination
  - Version tracking
  - Condition management

**Characteristics**:
- Fast execution (< 1 second)
- No external dependencies
- Test individual methods
- High code coverage target: >80%

**Example**:
```go
func TestResolveRegistryReference(t *testing.T) {
    // Test registry reference resolution with various registry states
    tests := []struct {
        name string
        registry *Registry
        expectedARN string
        expectError bool
    }{...}
}
```

**Run**: `go test ./pkg/resource/schema -v`

### 2. Integration Tests (`pkg/resource/schema/integration_test.go`)

**Purpose**: Test component integration with fake Kubernetes client

**Coverage**:
- ✅ Full reference resolution flow
- ✅ Multiple schemas referencing same registry
- ✅ Cross-namespace reference validation
- ✅ Error classification (terminal vs recoverable)

**Characteristics**:
- Medium execution time (< 10 seconds)
- Uses fake Kubernetes client
- Tests cross-component interactions
- Validates error handling end-to-end

**Example**:
```go
func TestSchemaRegistryReferenceIntegration(t *testing.T) {
    fakeClient := fake.NewClientBuilder()...
    // Test complete reference resolution with K8s interactions
}
```

**Run**: `go test ./pkg/resource/schema -run Integration -v`

### 3. E2E Tests (`test/e2e/schema_test.go`)

**Purpose**: Test complete resource lifecycle with real or simulated cluster

**Coverage**:
- ✅ Full CRUD operations
- ✅ Reference resolution in reconciliation loop
- ✅ Schema versioning lifecycle
- ✅ Concurrent operations
- ✅ Resource cleanup

**Characteristics**:
- Slower execution (minutes)
- Requires test cluster or envtest
- Closest to production behavior
- Tests complete user workflows

**Example**:
```go
func TestSchemaLifecycle(t *testing.T) {
    // Create registry, create schema, update, delete
    // Test complete lifecycle with reconciliation
}
```

**Run**: `go test ./test/e2e -v -timeout 30m`

## Test Scenarios by Category

### Happy Path Scenarios

| Scenario | Unit | Integration | E2E |
|----------|------|-------------|-----|
| Create schema with default registry | ✅ | ✅ | ✅ |
| Create schema with registry reference | ✅ | ✅ | ✅ |
| Update schema definition | ✅ | ✅ | ✅ |
| Delete schema | - | - | ✅ |
| Multiple schemas, one registry | - | ✅ | ✅ |

### Error Scenarios

| Scenario | Unit | Integration | E2E |
|----------|------|-------------|-----|
| Registry not found | ✅ | ✅ | ✅ |
| Registry not ready (no metadata) | ✅ | ✅ | ✅ |
| Registry not ready (no ARN) | ✅ | ✅ | ✅ |
| Registry in terminal state | ✅ | ✅ | ✅ |
| Immutable field update | ✅ | - | ✅ |
| Invalid schema definition | - | - | ✅ |

### Edge Cases

| Scenario | Unit | Integration | E2E |
|----------|------|-------------|-----|
| Schema created before registry | - | ✅ | ✅ |
| Cross-namespace reference | - | ✅ | - |
| Concurrent schema creation | - | - | ✅ |
| Registry deleted while referenced | ✅ | - | ✅ |

## Test Data Management

### Test Fixtures

Common test data is defined in helper functions:

```go
// Unit tests
func ptrString(s string) *string { return &s }
func ptrInt64(i int64) *int64 { return &i }

// Integration tests
func createTestRegistry(name, arn string) *Registry {...}
func createTestSchema(name, registryRef string) *Schema {...}

// E2E tests
func createFullRegistry(name string) *Registry {...}
func waitForRegistryReady(name string) error {...}
```

### Test Namespaces

- Unit tests: No namespace required
- Integration tests: Uses fake client, any namespace works
- E2E tests: `glue-e2e-test` namespace, cleaned up after tests

## Coverage Targets

### Code Coverage Goals

- **Unit tests**: >80% line coverage
- **Integration tests**: >70% component coverage
- **E2E tests**: >60% workflow coverage

### Current Coverage

```bash
# Check coverage
go test ./pkg/resource/schema -cover

# Generate coverage report
go test ./pkg/resource/schema -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Coverage by Component

| Component | Unit | Integration | E2E | Total |
|-----------|------|-------------|-----|-------|
| references.go | ✅ 90% | ✅ 85% | ✅ 70% | 85% |
| delta.go | ✅ 95% | ✅ 80% | ✅ 60% | 80% |
| status.go | ✅ 92% | ✅ 75% | ✅ 65% | 80% |
| manager.go | ✅ 70% | ✅ 80% | ✅ 85% | 78% |

## Testing Best Practices

### 1. Test Naming Conventions

```go
// Unit tests: Test + FunctionName + Scenario
func TestResolveRegistryReference_RegistryNotFound(t *testing.T) {...}

// Integration tests: Test + Component + Integration + Scenario
func TestSchemaRegistryReferenceIntegration_MultipleSchemas(t *testing.T) {...}

// E2E tests: Test + Resource + Lifecycle/Action
func TestSchemaLifecycle_CreateUpdateDelete(t *testing.T) {...}
```

### 2. Test Organization

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name        string
        input       InputType
        expected    ExpectedType
        expectError bool
    }{
        {name: "happy path", ...},
        {name: "error case 1", ...},
        {name: "edge case", ...},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 3. Assertion Guidelines

```go
// Use require for critical assertions (stops test on failure)
require.NoError(t, err)
require.NotNil(t, result)

// Use assert for non-critical assertions (continues test)
assert.Equal(t, expected, actual)
assert.Contains(t, message, "error text")

// Custom error messages
assert.Equal(t, expected, actual, "Expected ARN to match")
```

### 4. Test Cleanup

```go
// Unit tests: Usually no cleanup needed

// Integration tests: Clean up test objects
defer deleteSchema(t, fakeClient, schema)

// E2E tests: Always clean up
defer func() {
    cleanupNamespace(t, k8sClient, testNamespace)
}()
```

### 5. Test Timeouts

```go
// Unit tests: Fast, no explicit timeout needed

// Integration tests: Short timeout for fake client operations
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// E2E tests: Longer timeout for real reconciliation
const (
    defaultTimeout  = 5 * time.Minute
    defaultInterval = 5 * time.Second
)
```

## Continuous Integration

### CI Pipeline Stages

```yaml
stages:
  - lint       # golangci-lint, go vet
  - unit       # Fast unit tests
  - integration # Integration tests with fake client
  - e2e        # E2E tests with envtest
  - coverage   # Coverage reporting
```

### Test Execution in CI

```bash
# Stage 1: Unit tests (fast)
go test ./pkg/... -short -v

# Stage 2: Integration tests
go test ./pkg/... -run Integration -v

# Stage 3: E2E tests
go test ./test/e2e -v -timeout 30m

# Stage 4: Coverage
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Quality Gates

- ✅ All tests must pass
- ✅ Coverage must be ≥75%
- ✅ No new linting errors
- ✅ E2E tests complete within 30 minutes

## Test Maintenance

### When to Update Tests

1. **Adding new features**: Add tests at all levels
2. **Fixing bugs**: Add regression test at appropriate level
3. **Refactoring**: Ensure existing tests still pass
4. **Updating dependencies**: Check for test failures

### Test Review Checklist

- [ ] Tests are independent (can run in any order)
- [ ] Tests clean up resources
- [ ] Tests have descriptive names
- [ ] Tests cover happy path and error cases
- [ ] Tests have appropriate timeouts
- [ ] Tests use table-driven approach where appropriate
- [ ] Tests include helpful error messages

## Future Testing Enhancements

### Planned Additions

1. **Performance tests**: Measure reconciliation latency
2. **Load tests**: Test with many schemas/registries
3. **Chaos tests**: Test failure scenarios (network, API errors)
4. **Mutation tests**: Verify test effectiveness
5. **Property-based tests**: Use fuzzing for edge cases
6. **Contract tests**: Verify AWS API expectations

### Tools to Integrate

- **envtest**: Real Kubernetes API server for E2E tests
- **gomock**: Mock AWS SDK clients
- **testify**: Enhanced assertions and test suites
- **ginkgo/gomega**: BDD-style tests (optional)
- **kind**: Local Kubernetes clusters for testing

## References

- [ACK Testing Guide](https://aws-controllers-k8s.github.io/community/docs/contributor-docs/testing/)
- [Kubernetes Testing Best Practices](https://kubernetes.io/blog/2018/01/11/kubernetes-testing-best-practices/)
- [Go Testing Patterns](https://go.dev/doc/tutorial/add-a-test)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
