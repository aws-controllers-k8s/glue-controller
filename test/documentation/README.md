# Documentation Tests

This directory contains tests that verify the accuracy and completeness of documentation, including:

- Example YAML manifests
- API documentation
- Code examples
- Configuration examples
- Use case scenarios

## Purpose

Documentation tests ensure that:

1. **Examples are valid**: All YAML examples can be parsed correctly
2. **Examples are complete**: All important API fields are demonstrated
3. **Examples are current**: Examples match the current API version
4. **Use cases are covered**: Common scenarios are documented with working examples

## Running Documentation Tests

### Run all documentation tests
```bash
go test -v ./test/documentation
```

### Run specific test
```bash
go test -v ./test/documentation -run TestSchemaExamples_BasicSchema
```

### Run with coverage
```bash
go test -v -cover ./test/documentation
```

## Test Categories

### 1. Basic Examples (`TestSchemaExamples_BasicSchema`)
Tests the most basic schema example:
- Required fields (schemaName, dataFormat, schemaDefinition)
- Default configuration
- Simple AVRO schema

### 2. Registry Integration (`TestSchemaExamples_WithRegistry`)
Tests schema with registry reference:
- RegistryRef configuration
- Cross-resource references

### 3. Data Format Examples
Tests for each supported format:
- **AVRO**: `TestSchemaExamples_BasicSchema`, `TestSchemaExamples_NestedAVRO`
- **JSON Schema**: `TestSchemaExamples_JSONSchema`
- **Protobuf**: `TestSchemaExamples_ProtobufSchema`

### 4. Compatibility Modes (`TestSchemaExamples_CompatibilityModes`)
Tests all compatibility mode examples:
- BACKWARD
- FORWARD
- FULL
- BACKWARD_ALL
- FORWARD_ALL
- FULL_ALL
- DISABLED

### 5. Complex Examples (`TestSchemaExamples_NestedAVRO`)
Tests advanced schema patterns:
- Nested records
- Arrays of records
- Complex type hierarchies

### 6. File-Based Tests (`TestSchemaExamples_FromFile`)
Tests examples from the `examples/` directory:
- Loads all `schema*.yaml` files
- Validates each file can be parsed
- Verifies basic structure

### 7. API Documentation (`TestSchemaExamples_APIFieldDocumentation`)
Tests that all API fields are documented:
- Required fields coverage
- Optional fields coverage
- Comprehensive example

### 8. Error Scenarios (`TestSchemaExamples_ErrorScenarios`)
Tests documented error cases:
- Missing required fields
- Invalid field values
- Edge cases

### 9. Use Cases (`TestSchemaExamples_UseCases`)
Tests real-world scenario examples:
- Event streaming
- Data lake
- API payload validation

## Adding New Documentation Tests

When adding new examples to documentation:

1. **Create the test**:
   ```go
   func TestSchemaExamples_NewFeature(t *testing.T) {
       example := `
   apiVersion: glue.services.k8s.aws/v1alpha1
   kind: Schema
   ...
   `
       var schema svcapitypes.Schema
       err := yaml.Unmarshal([]byte(example), &schema)
       require.NoError(t, err)

       // Add assertions
   }
   ```

2. **Add the example file** (optional):
   - Create `examples/schema-new-feature.yaml`
   - File will be automatically tested by `TestSchemaExamples_FromFile`

3. **Update documentation**:
   - Add example to README or user guide
   - Reference the test in code comments

## Test Structure

```
test/documentation/
├── README.md                    # This file
├── schema_examples_test.go      # Schema documentation tests
└── registry_examples_test.go    # Registry documentation tests (future)
```

## Example YAML Structure

All Schema examples should follow this structure:

```yaml
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: example-name              # Required: Unique name
  namespace: default              # Optional: Defaults to default
spec:
  schemaName: example-name        # Required: AWS Glue schema name
  dataFormat: AVRO|JSON|PROTOBUF  # Required: Schema format
  schemaDefinition: |             # Required: Schema definition
    ...
  compatibility: BACKWARD         # Optional: Compatibility mode
  registryRef:                    # Optional: Registry reference
    name: registry-name
  description: "..."              # Optional: Description
  tags:                           # Optional: Tags
    - key: Environment
      value: Production
```

## Validation Levels

Documentation tests perform validation at multiple levels:

1. **Syntax Level**: YAML can be parsed without errors
2. **Structure Level**: Required fields are present
3. **Semantic Level**: Field values are valid and meaningful
4. **Integration Level**: References to other resources are correct

## Continuous Integration

These tests run as part of the CI pipeline to ensure:

- Documentation stays synchronized with code
- Examples work with current API version
- New features are properly documented
- Breaking changes are caught early

## Troubleshooting

### Test fails with "Example YAML should be valid"
- Check YAML syntax (indentation, quotes)
- Verify API version matches current version
- Ensure all required fields are present

### Test fails with "Examples directory not found"
- Tests skip if `examples/` directory doesn't exist
- This is expected for in-progress development
- Create `examples/` directory and add example files

### Test fails after API changes
- Update examples to match new API version
- Update field names if they changed
- Add tests for new fields

## Best Practices

1. **Keep examples simple**: Start with minimal working example
2. **Build complexity gradually**: Show advanced features separately
3. **Test edge cases**: Include boundary conditions
4. **Document assumptions**: Explain prerequisites and requirements
5. **Provide context**: Explain when to use each example

## Future Enhancements

Planned improvements for documentation tests:

- [ ] Validate schema definitions are syntactically correct
- [ ] Test compatibility between schema versions
- [ ] Validate cross-references between resources
- [ ] Test migration scenarios
- [ ] Add performance benchmarks for complex schemas
- [ ] Generate documentation from test examples
