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
	"encoding/json"
	"fmt"
	"regexp"

	svcapitypes "github.com/aws-controllers-k8s/glue-controller/apis/v1alpha1"
)

const (
	// MaxSchemaDefinitionLength is the maximum length for a schema definition
	// AWS Glue has a limit of 170,000 characters for schema definitions
	MaxSchemaDefinitionLength = 170000

	// SchemaNamePattern is the regex pattern for valid schema names
	// Schema names must start with a letter, contain only alphanumeric characters,
	// hyphens, and underscores
	SchemaNamePattern = `^[a-zA-Z][a-zA-Z0-9_-]*$`
)

var (
	schemaNameRegex = regexp.MustCompile(SchemaNamePattern)

	// ValidDataFormats lists the allowed data formats for schemas
	ValidDataFormats = []string{"AVRO", "JSON", "PROTOBUF"}

	// ValidCompatibilityModes lists the allowed compatibility modes
	ValidCompatibilityModes = []string{
		"NONE",
		"DISABLED",
		"BACKWARD",
		"FORWARD",
		"FULL",
		"BACKWARD_ALL",
		"FORWARD_ALL",
		"FULL_ALL",
	}

	// ValidDeletionPolicies lists the allowed deletion policies
	ValidDeletionPolicies = []string{"Delete", "Retain"}
)

// validateSchemaSpec performs comprehensive validation on a Schema spec
func (rm *resourceManager) validateSchemaSpec(schema *svcapitypes.Schema) error {
	if schema == nil || schema.Spec.SchemaName == nil {
		return fmt.Errorf("schema name is required")
	}

	// Validate schema name format
	if err := validateSchemaName(*schema.Spec.SchemaName); err != nil {
		return err
	}

	// Validate schema definition
	if schema.Spec.SchemaDefinition == nil {
		return fmt.Errorf("schema definition is required")
	}

	if err := validateSchemaDefinition(*schema.Spec.SchemaDefinition); err != nil {
		return err
	}

	// Validate data format
	if schema.Spec.DataFormat == nil {
		return fmt.Errorf("data format is required")
	}

	if err := validateDataFormat(*schema.Spec.DataFormat); err != nil {
		return err
	}

	// Validate schema definition content based on data format
	if err := validateSchemaDefinitionFormat(
		*schema.Spec.SchemaDefinition,
		*schema.Spec.DataFormat,
	); err != nil {
		return err
	}

	// Validate compatibility mode if specified
	if schema.Spec.Compatibility != nil {
		if err := validateCompatibilityMode(*schema.Spec.Compatibility); err != nil {
			return err
		}
	}

	// Validate deletion policy if specified
	if schema.Spec.DeletionPolicy != nil {
		if err := validateDeletionPolicy(*schema.Spec.DeletionPolicy); err != nil {
			return err
		}
	}

	// Validate registry reference if specified
	if schema.Spec.RegistryRef != nil {
		if err := validateRegistryRef(schema.Spec.RegistryRef); err != nil {
			return err
		}
	}

	return nil
}

// validateSchemaName validates that the schema name follows AWS Glue naming rules
func validateSchemaName(name string) error {
	if name == "" {
		return fmt.Errorf("schema name cannot be empty")
	}

	if len(name) > 255 {
		return fmt.Errorf("schema name cannot exceed 255 characters, got %d", len(name))
	}

	if !schemaNameRegex.MatchString(name) {
		return fmt.Errorf(
			"schema name must start with a letter and contain only alphanumeric characters, hyphens, and underscores; got: %s",
			name,
		)
	}

	return nil
}

// validateSchemaDefinition validates the schema definition length and basic structure
func validateSchemaDefinition(definition string) error {
	if definition == "" {
		return fmt.Errorf("schema definition cannot be empty")
	}

	if len(definition) > MaxSchemaDefinitionLength {
		return fmt.Errorf(
			"schema definition exceeds maximum length of %d characters, got %d",
			MaxSchemaDefinitionLength,
			len(definition),
		)
	}

	return nil
}

// validateDataFormat validates that the data format is one of the supported values
func validateDataFormat(format string) error {
	for _, valid := range ValidDataFormats {
		if format == valid {
			return nil
		}
	}

	return fmt.Errorf(
		"invalid data format %q; must be one of: %v",
		format,
		ValidDataFormats,
	)
}

// validateSchemaDefinitionFormat validates that the schema definition is valid
// for the specified data format
func validateSchemaDefinitionFormat(definition, format string) error {
	switch format {
	case "JSON":
		return validateJSONSchema(definition)
	case "AVRO":
		return validateAvroSchema(definition)
	case "PROTOBUF":
		// Protobuf validation is more complex and typically done by AWS
		// We just do basic checks here
		return validateProtobufSchema(definition)
	default:
		// Format already validated by validateDataFormat
		return nil
	}
}

// validateJSONSchema validates that the definition is valid JSON
func validateJSONSchema(definition string) error {
	var js interface{}
	if err := json.Unmarshal([]byte(definition), &js); err != nil {
		return fmt.Errorf("invalid JSON schema definition: %v", err)
	}

	// Additional JSON Schema validation could be added here
	// For now, we just verify it's valid JSON

	return nil
}

// validateAvroSchema validates that the definition is valid Avro schema JSON
func validateAvroSchema(definition string) error {
	// Avro schemas are JSON documents, so first validate as JSON
	var avroSchema map[string]interface{}
	if err := json.Unmarshal([]byte(definition), &avroSchema); err != nil {
		return fmt.Errorf("invalid Avro schema definition (not valid JSON): %v", err)
	}

	// Check for required Avro schema fields
	// Avro schemas must have a "type" field
	if _, hasType := avroSchema["type"]; !hasType {
		return fmt.Errorf("Avro schema must have a 'type' field")
	}

	// For record types, check for name and fields
	if typeVal, ok := avroSchema["type"].(string); ok && typeVal == "record" {
		if _, hasName := avroSchema["name"]; !hasName {
			return fmt.Errorf("Avro record schema must have a 'name' field")
		}
		if _, hasFields := avroSchema["fields"]; !hasFields {
			return fmt.Errorf("Avro record schema must have a 'fields' array")
		}
	}

	return nil
}

// validateProtobufSchema performs basic validation on Protocol Buffer schema
func validateProtobufSchema(definition string) error {
	// Protobuf schemas are text-based .proto files
	// Basic validation: check for common keywords

	if len(definition) == 0 {
		return fmt.Errorf("Protobuf schema cannot be empty")
	}

	// Check for basic Protobuf syntax elements
	// This is a simple check - full validation is done by AWS Glue
	hasMessage := regexp.MustCompile(`message\s+\w+`).MatchString(definition)
	hasService := regexp.MustCompile(`service\s+\w+`).MatchString(definition)
	hasSyntax := regexp.MustCompile(`syntax\s*=\s*"proto[23]"`).MatchString(definition)

	if !hasMessage && !hasService && !hasSyntax {
		return fmt.Errorf(
			"Protobuf schema does not appear to contain valid proto syntax (expected 'message', 'service', or 'syntax' declarations)",
		)
	}

	return nil
}

// validateCompatibilityMode validates the compatibility mode value
func validateCompatibilityMode(mode string) error {
	for _, valid := range ValidCompatibilityModes {
		if mode == valid {
			return nil
		}
	}

	return fmt.Errorf(
		"invalid compatibility mode %q; must be one of: %v",
		mode,
		ValidCompatibilityModes,
	)
}

// validateDeletionPolicy validates the deletion policy value
func validateDeletionPolicy(policy string) error {
	for _, valid := range ValidDeletionPolicies {
		if policy == valid {
			return nil
		}
	}

	return fmt.Errorf(
		"invalid deletion policy %q; must be one of: %v",
		policy,
		ValidDeletionPolicies,
	)
}

// validateRegistryRef validates that a registry reference has required fields
func validateRegistryRef(ref *svcapitypes.RegistryReference) error {
	if ref == nil {
		return nil // Registry ref is optional
	}

	if ref.Name == nil || *ref.Name == "" {
		return fmt.Errorf("registryRef.name is required when registryRef is specified")
	}

	// Validate registry name format (same rules as schema name)
	if err := validateSchemaName(*ref.Name); err != nil {
		return fmt.Errorf("invalid registryRef.name: %v", err)
	}

	return nil
}

// validateSchemaUpdate validates that an update to a schema is allowed
func (rm *resourceManager) validateSchemaUpdate(
	desired *svcapitypes.Schema,
	latest *svcapitypes.Schema,
) error {
	// Validate immutable fields haven't changed
	if desired.Spec.SchemaName != nil && latest.Spec.SchemaName != nil {
		if *desired.Spec.SchemaName != *latest.Spec.SchemaName {
			return fmt.Errorf("schema name is immutable and cannot be changed from %q to %q",
				*latest.Spec.SchemaName,
				*desired.Spec.SchemaName,
			)
		}
	}

	if desired.Spec.DataFormat != nil && latest.Spec.DataFormat != nil {
		if *desired.Spec.DataFormat != *latest.Spec.DataFormat {
			return fmt.Errorf("data format is immutable and cannot be changed from %q to %q",
				*latest.Spec.DataFormat,
				*desired.Spec.DataFormat,
			)
		}
	}

	// Validate registry reference hasn't changed
	if desired.Spec.RegistryRef != nil && latest.Spec.RegistryRef != nil {
		if desired.Spec.RegistryRef.Name != nil && latest.Spec.RegistryRef.Name != nil {
			if *desired.Spec.RegistryRef.Name != *latest.Spec.RegistryRef.Name {
				return fmt.Errorf(
					"registry reference is immutable and cannot be changed from %q to %q",
					*latest.Spec.RegistryRef.Name,
					*desired.Spec.RegistryRef.Name,
				)
			}
		}
	}

	// If schema definition changed, validate the new definition
	if desired.Spec.SchemaDefinition != nil && latest.Spec.SchemaDefinition != nil {
		if *desired.Spec.SchemaDefinition != *latest.Spec.SchemaDefinition {
			// Validate the new definition format
			if desired.Spec.DataFormat != nil {
				if err := validateSchemaDefinitionFormat(
					*desired.Spec.SchemaDefinition,
					*desired.Spec.DataFormat,
				); err != nil {
					return fmt.Errorf("invalid schema definition update: %v", err)
				}
			}
		}
	}

	return nil
}
