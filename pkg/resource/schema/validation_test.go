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
	"testing"

	svcapitypes "github.com/aws-controllers-k8s/glue-controller/apis/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestValidateSchemaDefinition tests AVRO schema definition validation
func TestValidateSchemaDefinition_AVRO(t *testing.T) {
	tests := []struct {
		name       string
		definition string
		wantErr    bool
		errMsg     string
	}{
		{
			name: "ValidAVRORecord",
			definition: `{
				"type": "record",
				"name": "User",
				"fields": [
					{"name": "id", "type": "string"},
					{"name": "email", "type": "string"}
				]
			}`,
			wantErr: false,
		},
		{
			name: "ValidAVROEnum",
			definition: `{
				"type": "enum",
				"name": "Status",
				"symbols": ["ACTIVE", "INACTIVE", "PENDING"]
			}`,
			wantErr: false,
		},
		{
			name: "ValidAVROArray",
			definition: `{
				"type": "record",
				"name": "Container",
				"fields": [
					{"name": "items", "type": {"type": "array", "items": "string"}}
				]
			}`,
			wantErr: false,
		},
		{
			name: "ValidAVROUnion",
			definition: `{
				"type": "record",
				"name": "Optional",
				"fields": [
					{"name": "value", "type": ["null", "string"]}
				]
			}`,
			wantErr: false,
		},
		{
			name: "ValidAVRONested",
			definition: `{
				"type": "record",
				"name": "Order",
				"fields": [
					{"name": "id", "type": "string"},
					{"name": "customer", "type": {
						"type": "record",
						"name": "Customer",
						"fields": [
							{"name": "name", "type": "string"},
							{"name": "email", "type": "string"}
						]
					}}
				]
			}`,
			wantErr: false,
		},
		{
			name:       "InvalidJSON",
			definition: `{invalid json`,
			wantErr:    true,
			errMsg:     "invalid JSON",
		},
		{
			name:       "MissingTypeField",
			definition: `{"name": "User", "fields": []}`,
			wantErr:    true,
			errMsg:     "missing type field",
		},
		{
			name:       "MissingNameField",
			definition: `{"type": "record", "fields": []}`,
			wantErr:    true,
			errMsg:     "missing name field",
		},
		{
			name:       "InvalidFieldType",
			definition: `{"type": "record", "name": "User", "fields": [{"name": "id"}]}`,
			wantErr:    true,
			errMsg:     "field 0 missing type",
		},
		{
			name:       "EmptyFields",
			definition: `{"type": "record", "name": "User", "fields": []}`,
			wantErr:    false, // Valid empty record
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAVROSchema(tt.definition)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidateSchemaDefinition_JSON tests JSON schema definition validation
func TestValidateSchemaDefinition_JSON(t *testing.T) {
	tests := []struct {
		name       string
		definition string
		wantErr    bool
		errMsg     string
	}{
		{
			name: "ValidJSONSchema",
			definition: `{
				"$schema": "http://json-schema.org/draft-07/schema#",
				"type": "object",
				"properties": {
					"id": {"type": "string"},
					"email": {"type": "string", "format": "email"}
				},
				"required": ["id"]
			}`,
			wantErr: false,
		},
		{
			name: "ValidJSONSchemaWithNesting",
			definition: `{
				"$schema": "http://json-schema.org/draft-07/schema#",
				"type": "object",
				"properties": {
					"user": {
						"type": "object",
						"properties": {
							"name": {"type": "string"},
							"age": {"type": "integer"}
						}
					}
				}
			}`,
			wantErr: false,
		},
		{
			name: "ValidJSONSchemaWithArray",
			definition: `{
				"$schema": "http://json-schema.org/draft-07/schema#",
				"type": "object",
				"properties": {
					"tags": {
						"type": "array",
						"items": {"type": "string"}
					}
				}
			}`,
			wantErr: false,
		},
		{
			name:       "InvalidJSON",
			definition: `{invalid`,
			wantErr:    true,
			errMsg:     "invalid JSON",
		},
		{
			name:       "EmptySchema",
			definition: `{}`,
			wantErr:    false, // Valid minimal schema
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSONSchema(tt.definition)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidateSchemaDefinition_Protobuf tests Protobuf schema definition validation
func TestValidateSchemaDefinition_Protobuf(t *testing.T) {
	tests := []struct {
		name       string
		definition string
		wantErr    bool
		errMsg     string
	}{
		{
			name: "ValidProtobuf",
			definition: `
				syntax = "proto3";
				package test;

				message User {
					string id = 1;
					string email = 2;
				}
			`,
			wantErr: false,
		},
		{
			name: "ValidProtobufWithEnum",
			definition: `
				syntax = "proto3";
				package test;

				enum Status {
					UNKNOWN = 0;
					ACTIVE = 1;
					INACTIVE = 2;
				}

				message User {
					string id = 1;
					Status status = 2;
				}
			`,
			wantErr: false,
		},
		{
			name: "ValidProtobufWithNested",
			definition: `
				syntax = "proto3";
				package test;

				message Order {
					string id = 1;
					Customer customer = 2;

					message Customer {
						string name = 1;
						string email = 2;
					}
				}
			`,
			wantErr: false,
		},
		{
			name:       "MissingSyntax",
			definition: `message User { string id = 1; }`,
			wantErr:    true,
			errMsg:     "missing syntax declaration",
		},
		{
			name:       "InvalidSyntax",
			definition: `syntax = "proto2"; message User { string id = 1; }`,
			wantErr:    true,
			errMsg:     "proto2 syntax is not recommended",
		},
		{
			name:       "MissingFieldNumber",
			definition: `syntax = "proto3"; message User { string id; }`,
			wantErr:    true,
			errMsg:     "schema must contain at least one message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProtobufSchema(tt.definition)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidateCompatibilityMode tests compatibility mode validation
func TestValidateCompatibilityMode(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		wantErr bool
	}{
		{
			name:    "ValidBACKWARD",
			mode:    "BACKWARD",
			wantErr: false,
		},
		{
			name:    "ValidFORWARD",
			mode:    "FORWARD",
			wantErr: false,
		},
		{
			name:    "ValidFULL",
			mode:    "FULL",
			wantErr: false,
		},
		{
			name:    "ValidDISABLED",
			mode:    "DISABLED",
			wantErr: false,
		},
		{
			name:    "ValidBACKWARD_ALL",
			mode:    "BACKWARD_ALL",
			wantErr: false,
		},
		{
			name:    "ValidFORWARD_ALL",
			mode:    "FORWARD_ALL",
			wantErr: false,
		},
		{
			name:    "ValidFULL_ALL",
			mode:    "FULL_ALL",
			wantErr: false,
		},
		{
			name:    "InvalidMode",
			mode:    "UNKNOWN",
			wantErr: true,
		},
		{
			name:    "EmptyMode",
			mode:    "",
			wantErr: false, // Empty is valid (uses default)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCompatibilityMode(tt.mode)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidateSchemaSpec tests complete Schema spec validation
func TestValidateSchemaSpec(t *testing.T) {
	tests := []struct {
		name    string
		schema  *svcapitypes.Schema
		wantErr bool
		errMsg  string
	}{
		{
			name: "ValidSchemaWithDefaultRegistry",
			schema: &svcapitypes.Schema{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-schema",
					Namespace: "default",
				},
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": [{"name": "id", "type": "string"}]}`),
				},
			},
			wantErr: false,
		},
		{
			name: "ValidSchemaWithRegistryReference",
			schema: &svcapitypes.Schema{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-schema",
					Namespace: "default",
				},
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": [{"name": "id", "type": "string"}]}`),
					RegistryRef: &svcapitypes.RegistryReference{
						Name: ptrString("custom-registry"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "ValidSchemaWithCompatibility",
			schema: &svcapitypes.Schema{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-schema",
					Namespace: "default",
				},
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": [{"name": "id", "type": "string"}]}`),
					Compatibility:    ptrString("BACKWARD"),
				},
			},
			wantErr: false,
		},
		{
			name: "MissingSchemaName",
			schema: &svcapitypes.Schema{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-schema",
					Namespace: "default",
				},
				Spec: svcapitypes.SchemaSpec{
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": []}`),
				},
			},
			wantErr: true,
			errMsg:  "schemaName is required",
		},
		{
			name: "MissingDataFormat",
			schema: &svcapitypes.Schema{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-schema",
					Namespace: "default",
				},
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": []}`),
				},
			},
			wantErr: true,
			errMsg:  "dataFormat is required",
		},
		{
			name: "MissingSchemaDefinition",
			schema: &svcapitypes.Schema{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-schema",
					Namespace: "default",
				},
				Spec: svcapitypes.SchemaSpec{
					SchemaName: ptrString("test-schema"),
					DataFormat: ptrString("AVRO"),
				},
			},
			wantErr: true,
			errMsg:  "schemaDefinition is required",
		},
		{
			name: "InvalidDataFormat",
			schema: &svcapitypes.Schema{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-schema",
					Namespace: "default",
				},
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					DataFormat:       ptrString("INVALID"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": []}`),
				},
			},
			wantErr: true,
			errMsg:  "invalid dataFormat",
		},
		{
			name: "InvalidCompatibilityMode",
			schema: &svcapitypes.Schema{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-schema",
					Namespace: "default",
				},
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": []}`),
					Compatibility:    ptrString("INVALID"),
				},
			},
			wantErr: true,
			errMsg:  "invalid compatibility mode",
		},
		{
			name: "InvalidAVROSchemaDefinition",
			schema: &svcapitypes.Schema{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-schema",
					Namespace: "default",
				},
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{invalid json`),
				},
			},
			wantErr: true,
			errMsg:  "invalid AVRO schema definition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSchemaSpec(tt.schema)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidateSchemaUpdate tests immutable field validation
func TestValidateSchemaUpdate(t *testing.T) {
	tests := []struct {
		name      string
		oldSchema *svcapitypes.Schema
		newSchema *svcapitypes.Schema
		wantErr   bool
		errMsg    string
	}{
		{
			name: "ValidUpdate_SchemaDefinitionOnly",
			oldSchema: &svcapitypes.Schema{
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": [{"name": "id", "type": "string"}]}`),
				},
			},
			newSchema: &svcapitypes.Schema{
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": [{"name": "id", "type": "string"}, {"name": "name", "type": "string"}]}`),
				},
			},
			wantErr: false,
		},
		{
			name: "ValidUpdate_CompatibilityChange",
			oldSchema: &svcapitypes.Schema{
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": []}`),
					Compatibility:    ptrString("BACKWARD"),
				},
			},
			newSchema: &svcapitypes.Schema{
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": []}`),
					Compatibility:    ptrString("FORWARD"),
				},
			},
			wantErr: false,
		},
		{
			name: "InvalidUpdate_SchemaNameChange",
			oldSchema: &svcapitypes.Schema{
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("old-name"),
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": []}`),
				},
			},
			newSchema: &svcapitypes.Schema{
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("new-name"),
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": []}`),
				},
			},
			wantErr: true,
			errMsg:  "schemaName is immutable",
		},
		{
			name: "InvalidUpdate_DataFormatChange",
			oldSchema: &svcapitypes.Schema{
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": []}`),
				},
			},
			newSchema: &svcapitypes.Schema{
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					DataFormat:       ptrString("JSON"),
					SchemaDefinition: ptrString(`{"type": "object"}`),
				},
			},
			wantErr: true,
			errMsg:  "dataFormat is immutable",
		},
		{
			name: "InvalidUpdate_RegistryRefChange",
			oldSchema: &svcapitypes.Schema{
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": []}`),
					RegistryRef: &svcapitypes.RegistryReference{
						Name: ptrString("old-registry"),
					},
				},
			},
			newSchema: &svcapitypes.Schema{
				Spec: svcapitypes.SchemaSpec{
					SchemaName:       ptrString("test-schema"),
					DataFormat:       ptrString("AVRO"),
					SchemaDefinition: ptrString(`{"type": "record", "name": "Test", "fields": []}`),
					RegistryRef: &svcapitypes.RegistryReference{
						Name: ptrString("new-registry"),
					},
				},
			},
			wantErr: true,
			errMsg:  "registryRef is immutable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSchemaUpdate(tt.oldSchema, tt.newSchema)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
