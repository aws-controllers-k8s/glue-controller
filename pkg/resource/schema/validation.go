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
	"strings"

	svcapitypes "github.com/aws-controllers-k8s/glue-controller/apis/v1alpha1"
)

// ValidateSchemaSpec validates the complete Schema specification
func ValidateSchemaSpec(schema *svcapitypes.Schema) error {
	// Validate required fields
	if schema.Spec.SchemaName == nil || *schema.Spec.SchemaName == "" {
		return fmt.Errorf("schemaName is required")
	}

	if schema.Spec.DataFormat == nil || *schema.Spec.DataFormat == "" {
		return fmt.Errorf("dataFormat is required")
	}

	if schema.Spec.SchemaDefinition == nil || *schema.Spec.SchemaDefinition == "" {
		return fmt.Errorf("schemaDefinition is required")
	}

	// Validate data format
	validFormats := map[string]bool{
		"AVRO":     true,
		"JSON":     true,
		"PROTOBUF": true,
	}
	if !validFormats[*schema.Spec.DataFormat] {
		return fmt.Errorf("invalid dataFormat: %s, must be one of: AVRO, JSON, PROTOBUF", *schema.Spec.DataFormat)
	}

	// Validate compatibility mode if provided
	if schema.Spec.Compatibility != nil {
		if err := ValidateCompatibilityMode(*schema.Spec.Compatibility); err != nil {
			return fmt.Errorf("invalid compatibility mode: %w", err)
		}
	}

	// Validate schema definition based on format
	switch *schema.Spec.DataFormat {
	case "AVRO":
		if err := ValidateAVROSchema(*schema.Spec.SchemaDefinition); err != nil {
			return fmt.Errorf("invalid AVRO schema definition: %w", err)
		}
	case "JSON":
		if err := ValidateJSONSchema(*schema.Spec.SchemaDefinition); err != nil {
			return fmt.Errorf("invalid JSON schema definition: %w", err)
		}
	case "PROTOBUF":
		if err := ValidateProtobufSchema(*schema.Spec.SchemaDefinition); err != nil {
			return fmt.Errorf("invalid Protobuf schema definition: %w", err)
		}
	}

	return nil
}

// ValidateAVROSchema validates an AVRO schema definition
func ValidateAVROSchema(definition string) error {
	// Parse JSON
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(definition), &schema); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Check required fields
	schemaType, ok := schema["type"]
	if !ok {
		return fmt.Errorf("missing type field")
	}

	// Validate based on type
	typeStr, ok := schemaType.(string)
	if !ok {
		return fmt.Errorf("type must be a string")
	}

	switch typeStr {
	case "record":
		return validateAVRORecord(schema)
	case "enum":
		return validateAVROEnum(schema)
	case "array":
		return validateAVROArray(schema)
	case "map":
		return validateAVROMap(schema)
	case "fixed":
		return validateAVROFixed(schema)
	case "string", "int", "long", "float", "double", "boolean", "null", "bytes":
		// Primitive types are valid
		return nil
	default:
		return fmt.Errorf("unknown AVRO type: %s", typeStr)
	}
}

func validateAVRORecord(schema map[string]interface{}) error {
	// Record must have name and fields
	if _, ok := schema["name"]; !ok {
		return fmt.Errorf("missing name field for record type")
	}

	fields, ok := schema["fields"]
	if !ok {
		// Empty fields is valid
		return nil
	}

	// Validate fields array
	fieldsArray, ok := fields.([]interface{})
	if !ok {
		return fmt.Errorf("fields must be an array")
	}

	for i, field := range fieldsArray {
		fieldMap, ok := field.(map[string]interface{})
		if !ok {
			return fmt.Errorf("field %d must be an object", i)
		}

		if _, ok := fieldMap["name"]; !ok {
			return fmt.Errorf("field %d missing name", i)
		}

		if _, ok := fieldMap["type"]; !ok {
			return fmt.Errorf("field %d missing type", i)
		}
	}

	return nil
}

func validateAVROEnum(schema map[string]interface{}) error {
	if _, ok := schema["name"]; !ok {
		return fmt.Errorf("missing name field for enum type")
	}

	symbols, ok := schema["symbols"]
	if !ok {
		return fmt.Errorf("missing symbols field for enum type")
	}

	symbolsArray, ok := symbols.([]interface{})
	if !ok {
		return fmt.Errorf("symbols must be an array")
	}

	if len(symbolsArray) == 0 {
		return fmt.Errorf("symbols array cannot be empty")
	}

	return nil
}

func validateAVROArray(schema map[string]interface{}) error {
	if _, ok := schema["items"]; !ok {
		return fmt.Errorf("missing items field for array type")
	}
	return nil
}

func validateAVROMap(schema map[string]interface{}) error {
	if _, ok := schema["values"]; !ok {
		return fmt.Errorf("missing values field for map type")
	}
	return nil
}

func validateAVROFixed(schema map[string]interface{}) error {
	if _, ok := schema["name"]; !ok {
		return fmt.Errorf("missing name field for fixed type")
	}

	if _, ok := schema["size"]; !ok {
		return fmt.Errorf("missing size field for fixed type")
	}

	return nil
}

// ValidateJSONSchema validates a JSON schema definition
func ValidateJSONSchema(definition string) error {
	// Parse JSON
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(definition), &schema); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// JSON Schema is very flexible, so we just validate it's valid JSON
	// and optionally check for common fields
	if schemaVersion, ok := schema["$schema"]; ok {
		schemaVersionStr, ok := schemaVersion.(string)
		if !ok {
			return fmt.Errorf("$schema must be a string")
		}
		// Validate it's a known JSON Schema version
		if !strings.Contains(schemaVersionStr, "json-schema.org") {
			return fmt.Errorf("unknown JSON Schema version: %s", schemaVersionStr)
		}
	}

	return nil
}

// ValidateProtobufSchema validates a Protobuf schema definition
func ValidateProtobufSchema(definition string) error {
	// Basic Protobuf validation
	lines := strings.Split(definition, "\n")

	// Check for syntax declaration
	hasSyntax := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "syntax") {
			hasSyntax = true
			// Extract syntax version
			syntaxRegex := regexp.MustCompile(`syntax\s*=\s*"([^"]+)"`)
			matches := syntaxRegex.FindStringSubmatch(trimmed)
			if len(matches) < 2 {
				return fmt.Errorf("invalid syntax declaration format")
			}
			if matches[1] != "proto3" && matches[1] != "proto2" {
				return fmt.Errorf("unsupported syntax: %s (expected proto3 or proto2)", matches[1])
			}
			// Prefer proto3
			if matches[1] == "proto2" {
				return fmt.Errorf("proto2 syntax is not recommended, use proto3")
			}
			break
		}
	}

	if !hasSyntax {
		return fmt.Errorf("missing syntax declaration (syntax = \"proto3\")")
	}

	// Check for at least one message or enum definition
	hasDefinition := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "message ") || strings.HasPrefix(trimmed, "enum ") {
			hasDefinition = true
			break
		}
	}

	if !hasDefinition {
		return fmt.Errorf("schema must contain at least one message or enum definition")
	}

	// Basic field number validation
	fieldRegex := regexp.MustCompile(`^\s*\w+\s+\w+\s*=\s*(\d+)\s*;`)
	inEnum := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track enum context
		if strings.HasPrefix(trimmed, "enum ") {
			inEnum = true
			continue
		}
		if inEnum && strings.HasPrefix(trimmed, "}") {
			inEnum = false
			continue
		}

		if fieldRegex.MatchString(line) {
			matches := fieldRegex.FindStringSubmatch(line)
			if len(matches) < 2 {
				continue
			}
			// Field numbers are validated but we don't enforce specific ranges here
			// AWS Glue will do the final validation
		} else if strings.Contains(line, "=") && !strings.Contains(line, "syntax") && !strings.Contains(line, "option") && !inEnum {
			// Line looks like a field but doesn't match the pattern
			// Skip enum values (they use = but are not field definitions)
			if !strings.HasPrefix(trimmed, "//") && trimmed != "" {
				if !strings.HasPrefix(trimmed, "}") && !strings.HasPrefix(trimmed, "message") && !strings.HasPrefix(trimmed, "enum") {
					return fmt.Errorf("invalid field definition: %s (missing field number?)", line)
				}
			}
		}
	}

	return nil
}

// ValidateCompatibilityMode validates a schema compatibility mode
func ValidateCompatibilityMode(mode string) error {
	validModes := map[string]bool{
		"BACKWARD":     true,
		"FORWARD":      true,
		"FULL":         true,
		"DISABLED":     true,
		"BACKWARD_ALL": true,
		"FORWARD_ALL":  true,
		"FULL_ALL":     true,
		"":             true, // Empty is valid (uses default)
	}

	if !validModes[mode] {
		return fmt.Errorf("invalid compatibility mode: %s, must be one of: BACKWARD, FORWARD, FULL, DISABLED, BACKWARD_ALL, FORWARD_ALL, FULL_ALL", mode)
	}

	return nil
}

// ValidateSchemaUpdate validates updates to a Schema resource
// Some fields are immutable after creation
func ValidateSchemaUpdate(oldSchema, newSchema *svcapitypes.Schema) error {
	// SchemaName is immutable
	if oldSchema.Spec.SchemaName != nil && newSchema.Spec.SchemaName != nil {
		if *oldSchema.Spec.SchemaName != *newSchema.Spec.SchemaName {
			return fmt.Errorf("schemaName is immutable and cannot be changed")
		}
	}

	// DataFormat is immutable
	if oldSchema.Spec.DataFormat != nil && newSchema.Spec.DataFormat != nil {
		if *oldSchema.Spec.DataFormat != *newSchema.Spec.DataFormat {
			return fmt.Errorf("dataFormat is immutable and cannot be changed")
		}
	}

	// RegistryRef is immutable
	if oldSchema.Spec.RegistryRef != nil && newSchema.Spec.RegistryRef != nil {
		if oldSchema.Spec.RegistryRef.Name != nil && newSchema.Spec.RegistryRef.Name != nil {
			if *oldSchema.Spec.RegistryRef.Name != *newSchema.Spec.RegistryRef.Name {
				return fmt.Errorf("registryRef is immutable and cannot be changed")
			}
		}
	}

	// Validate the new schema spec
	return ValidateSchemaSpec(newSchema)
}
