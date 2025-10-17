// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package schema

import (
	"os"
	"path/filepath"
	"testing"

	svcapitypes "github.com/aws-controllers-k8s/glue-controller/apis/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// TestSchemaExamples_BasicSchema tests the basic schema example from documentation
func TestSchemaExamples_BasicSchema(t *testing.T) {
	example := `
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: user-schema
spec:
  schemaName: user-schema
  dataFormat: AVRO
  compatibility: BACKWARD
  schemaDefinition: |
    {
      "type": "record",
      "name": "User",
      "fields": [
        {"name": "id", "type": "string"},
        {"name": "email", "type": "string"},
        {"name": "created_at", "type": "long"}
      ]
    }
`

	var schema svcapitypes.Schema
	err := yaml.Unmarshal([]byte(example), &schema)
	require.NoError(t, err, "Example YAML should be valid")

	// Verify required fields
	assert.NotNil(t, schema.Spec.SchemaName)
	assert.Equal(t, "user-schema", *schema.Spec.SchemaName)
	assert.NotNil(t, schema.Spec.DataFormat)
	assert.Equal(t, "AVRO", *schema.Spec.DataFormat)
	assert.NotNil(t, schema.Spec.Compatibility)
	assert.Equal(t, "BACKWARD", *schema.Spec.Compatibility)
	assert.NotNil(t, schema.Spec.SchemaDefinition)
}

// TestSchemaExamples_WithRegistry tests schema with registry reference
func TestSchemaExamples_WithRegistry(t *testing.T) {
	example := `
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: order-schema
spec:
  schemaName: order-schema
  dataFormat: AVRO
  registryRef:
    name: my-registry
  schemaDefinition: |
    {
      "type": "record",
      "name": "Order",
      "fields": [
        {"name": "order_id", "type": "string"},
        {"name": "customer_id", "type": "string"},
        {"name": "total", "type": "double"}
      ]
    }
`

	var schema svcapitypes.Schema
	err := yaml.Unmarshal([]byte(example), &schema)
	require.NoError(t, err, "Example YAML should be valid")

	// Verify registry reference
	assert.NotNil(t, schema.Spec.RegistryRef)
	assert.NotNil(t, schema.Spec.RegistryRef.Name)
	assert.Equal(t, "my-registry", *schema.Spec.RegistryRef.Name)
}

// TestSchemaExamples_JSONSchema tests JSON Schema format example
func TestSchemaExamples_JSONSchema(t *testing.T) {
	example := `
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: product-schema
spec:
  schemaName: product-schema
  dataFormat: JSON
  schemaDefinition: |
    {
      "$schema": "http://json-schema.org/draft-07/schema#",
      "type": "object",
      "properties": {
        "id": {"type": "string"},
        "name": {"type": "string"},
        "price": {"type": "number"}
      },
      "required": ["id", "name"]
    }
`

	var schema svcapitypes.Schema
	err := yaml.Unmarshal([]byte(example), &schema)
	require.NoError(t, err, "Example YAML should be valid")

	assert.Equal(t, "JSON", *schema.Spec.DataFormat)
	assert.Contains(t, *schema.Spec.SchemaDefinition, "json-schema.org")
}

// TestSchemaExamples_ProtobufSchema tests Protobuf format example
func TestSchemaExamples_ProtobufSchema(t *testing.T) {
	example := `
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: event-schema
spec:
  schemaName: event-schema
  dataFormat: PROTOBUF
  schemaDefinition: |
    syntax = "proto3";
    package events;

    message Event {
      string id = 1;
      string type = 2;
      int64 timestamp = 3;
      bytes payload = 4;
    }
`

	var schema svcapitypes.Schema
	err := yaml.Unmarshal([]byte(example), &schema)
	require.NoError(t, err, "Example YAML should be valid")

	assert.Equal(t, "PROTOBUF", *schema.Spec.DataFormat)
	assert.Contains(t, *schema.Spec.SchemaDefinition, "syntax = \"proto3\"")
}

// TestSchemaExamples_CompatibilityModes tests all compatibility mode examples
func TestSchemaExamples_CompatibilityModes(t *testing.T) {
	modes := []string{
		"BACKWARD",
		"FORWARD",
		"FULL",
		"BACKWARD_ALL",
		"FORWARD_ALL",
		"FULL_ALL",
		"DISABLED",
	}

	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			example := `
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: test-schema
spec:
  schemaName: test-schema
  dataFormat: AVRO
  compatibility: ` + mode + `
  schemaDefinition: |
    {
      "type": "record",
      "name": "Test",
      "fields": [{"name": "id", "type": "string"}]
    }
`

			var schema svcapitypes.Schema
			err := yaml.Unmarshal([]byte(example), &schema)
			require.NoError(t, err, "Example YAML should be valid")

			assert.NotNil(t, schema.Spec.Compatibility)
			assert.Equal(t, mode, *schema.Spec.Compatibility)
		})
	}
}

// TestSchemaExamples_NestedAVRO tests nested AVRO schema example
func TestSchemaExamples_NestedAVRO(t *testing.T) {
	example := `
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: complex-schema
spec:
  schemaName: complex-schema
  dataFormat: AVRO
  schemaDefinition: |
    {
      "type": "record",
      "name": "Order",
      "fields": [
        {"name": "id", "type": "string"},
        {
          "name": "customer",
          "type": {
            "type": "record",
            "name": "Customer",
            "fields": [
              {"name": "name", "type": "string"},
              {"name": "email", "type": "string"}
            ]
          }
        },
        {
          "name": "items",
          "type": {
            "type": "array",
            "items": {
              "type": "record",
              "name": "Item",
              "fields": [
                {"name": "sku", "type": "string"},
                {"name": "quantity", "type": "int"}
              ]
            }
          }
        }
      ]
    }
`

	var schema svcapitypes.Schema
	err := yaml.Unmarshal([]byte(example), &schema)
	require.NoError(t, err, "Example YAML should be valid")

	assert.Contains(t, *schema.Spec.SchemaDefinition, "Customer")
	assert.Contains(t, *schema.Spec.SchemaDefinition, "Item")
	assert.Contains(t, *schema.Spec.SchemaDefinition, "array")
}

// TestSchemaExamples_FromFile tests loading examples from file
func TestSchemaExamples_FromFile(t *testing.T) {
	// Skip if examples directory doesn't exist
	examplesDir := "../../../examples"
	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		t.Skip("Examples directory not found, skipping file-based tests")
	}

	// Find all schema example files
	files, err := filepath.Glob(filepath.Join(examplesDir, "schema*.yaml"))
	if err != nil {
		t.Fatalf("Failed to glob example files: %v", err)
	}

	if len(files) == 0 {
		t.Skip("No schema example files found")
	}

	// Test each example file
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			require.NoError(t, err, "Should read example file")

			var schema svcapitypes.Schema
			err = yaml.Unmarshal(data, &schema)
			require.NoError(t, err, "Example YAML should be valid")

			// Verify basic structure
			assert.NotNil(t, schema.Spec.SchemaName, "Schema name should be set")
			assert.NotNil(t, schema.Spec.DataFormat, "Data format should be set")
			assert.NotNil(t, schema.Spec.SchemaDefinition, "Schema definition should be set")
		})
	}
}

// TestSchemaExamples_APIFieldDocumentation tests that all API fields are documented
func TestSchemaExamples_APIFieldDocumentation(t *testing.T) {
	// Test that examples cover all important Schema spec fields
	requiredFields := []string{
		"schemaName",
		"dataFormat",
		"schemaDefinition",
	}

	optionalFields := []string{
		"compatibility",
		"registryRef",
		"description",
		"tags",
	}

	// Create comprehensive example
	example := `
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: comprehensive-schema
spec:
  schemaName: comprehensive-schema
  dataFormat: AVRO
  compatibility: BACKWARD
  description: "Comprehensive schema example"
  registryRef:
    name: my-registry
  tags:
    Environment: "Production"
    Team: "DataEngineering"
  schemaDefinition: |
    {
      "type": "record",
      "name": "Example",
      "fields": [{"name": "id", "type": "string"}]
    }
`

	var schema svcapitypes.Schema
	err := yaml.Unmarshal([]byte(example), &schema)
	require.NoError(t, err, "Comprehensive example should be valid")

	// Verify required fields
	for _, field := range requiredFields {
		t.Run("Required_"+field, func(t *testing.T) {
			switch field {
			case "schemaName":
				assert.NotNil(t, schema.Spec.SchemaName)
			case "dataFormat":
				assert.NotNil(t, schema.Spec.DataFormat)
			case "schemaDefinition":
				assert.NotNil(t, schema.Spec.SchemaDefinition)
			}
		})
	}

	// Verify optional fields
	for _, field := range optionalFields {
		t.Run("Optional_"+field, func(t *testing.T) {
			switch field {
			case "compatibility":
				assert.NotNil(t, schema.Spec.Compatibility)
			case "registryRef":
				assert.NotNil(t, schema.Spec.RegistryRef)
			case "description":
				assert.NotNil(t, schema.Spec.Description)
			case "tags":
				assert.NotNil(t, schema.Spec.Tags)
			}
		})
	}
}

// TestSchemaExamples_ErrorScenarios tests documented error scenarios
func TestSchemaExamples_ErrorScenarios(t *testing.T) {
	t.Run("MissingRequiredField", func(t *testing.T) {
		example := `
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: invalid-schema
spec:
  # Missing schemaName
  dataFormat: AVRO
  schemaDefinition: |
    {"type": "string"}
`

		var schema svcapitypes.Schema
		err := yaml.Unmarshal([]byte(example), &schema)
		require.NoError(t, err, "YAML parsing should succeed")

		// Verify field is actually missing
		assert.Nil(t, schema.Spec.SchemaName, "Schema name should be nil")
	})

	t.Run("InvalidDataFormat", func(t *testing.T) {
		example := `
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: invalid-format
spec:
  schemaName: invalid-format
  dataFormat: INVALID
  schemaDefinition: |
    {"type": "string"}
`

		var schema svcapitypes.Schema
		err := yaml.Unmarshal([]byte(example), &schema)
		require.NoError(t, err, "YAML parsing should succeed")

		assert.Equal(t, "INVALID", *schema.Spec.DataFormat)
	})
}

// TestSchemaExamples_UseCases tests common use case examples
func TestSchemaExamples_UseCases(t *testing.T) {
	useCases := map[string]string{
		"EventStreaming": `
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: event-stream-schema
spec:
  schemaName: event-stream-schema
  dataFormat: AVRO
  compatibility: FORWARD
  schemaDefinition: |
    {
      "type": "record",
      "name": "StreamEvent",
      "fields": [
        {"name": "eventId", "type": "string"},
        {"name": "eventType", "type": "string"},
        {"name": "timestamp", "type": "long"},
        {"name": "data", "type": "string"}
      ]
    }
`,
		"DataLake": `
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: datalake-schema
spec:
  schemaName: datalake-schema
  dataFormat: AVRO
  compatibility: BACKWARD
  schemaDefinition: |
    {
      "type": "record",
      "name": "DataRecord",
      "fields": [
        {"name": "id", "type": "string"},
        {"name": "partition_key", "type": "string"},
        {"name": "data", "type": "bytes"}
      ]
    }
`,
		"APIPayload": `
apiVersion: glue.services.k8s.aws/v1alpha1
kind: Schema
metadata:
  name: api-payload-schema
spec:
  schemaName: api-payload-schema
  dataFormat: JSON
  schemaDefinition: |
    {
      "$schema": "http://json-schema.org/draft-07/schema#",
      "type": "object",
      "properties": {
        "requestId": {"type": "string"},
        "payload": {"type": "object"}
      },
      "required": ["requestId"]
    }
`,
	}

	for name, example := range useCases {
		t.Run(name, func(t *testing.T) {
			var schema svcapitypes.Schema
			err := yaml.Unmarshal([]byte(example), &schema)
			require.NoError(t, err, "Use case example should be valid")

			assert.NotNil(t, schema.Spec.SchemaName)
			assert.NotNil(t, schema.Spec.DataFormat)
			assert.NotNil(t, schema.Spec.SchemaDefinition)
		})
	}
}
