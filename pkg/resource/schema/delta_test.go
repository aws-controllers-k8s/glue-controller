// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
// http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"

	svcapitypes "github.com/aws-controllers-k8s/glue-controller/apis/v1alpha1"
)

// TestNewResourceDelta tests delta comparison for schema resources
func TestNewResourceDelta(t *testing.T) {
	tests := []struct {
		name          string
		desired       *resource
		latest        *resource
		expectedDiffs []string
		noDifferences bool
	}{
		{
			name: "no differences",
			desired: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName:       ptrString("test-schema"),
						DataFormat:       ptrString("AVRO"),
						SchemaDefinition: ptrString(`{"type": "record"}`),
						Compatibility:    ptrString("BACKWARD"),
					},
				},
			},
			latest: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName:       ptrString("test-schema"),
						DataFormat:       ptrString("AVRO"),
						SchemaDefinition: ptrString(`{"type": "record"}`),
						Compatibility:    ptrString("BACKWARD"),
					},
				},
			},
			noDifferences: true,
		},
		{
			name: "immutable field change - SchemaName",
			desired: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("new-name"),
						DataFormat: ptrString("AVRO"),
					},
				},
			},
			latest: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("old-name"),
						DataFormat: ptrString("AVRO"),
					},
				},
			},
			expectedDiffs: []string{"Spec.SchemaName"},
		},
		{
			name: "immutable field change - DataFormat",
			desired: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("test-schema"),
						DataFormat: ptrString("JSON"),
					},
				},
			},
			latest: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("test-schema"),
						DataFormat: ptrString("AVRO"),
					},
				},
			},
			expectedDiffs: []string{"Spec.DataFormat"},
		},
		{
			name: "mutable field change - SchemaDefinition",
			desired: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName:       ptrString("test-schema"),
						DataFormat:       ptrString("AVRO"),
						SchemaDefinition: ptrString(`{"type": "record", "fields": [{"name": "newField", "type": "string"}]}`),
					},
				},
			},
			latest: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName:       ptrString("test-schema"),
						DataFormat:       ptrString("AVRO"),
						SchemaDefinition: ptrString(`{"type": "record"}`),
					},
				},
			},
			expectedDiffs: []string{"Spec.SchemaDefinition"},
		},
		{
			name: "mutable field change - Compatibility",
			desired: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName:    ptrString("test-schema"),
						DataFormat:    ptrString("AVRO"),
						Compatibility: ptrString("FULL"),
					},
				},
			},
			latest: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName:    ptrString("test-schema"),
						DataFormat:    ptrString("AVRO"),
						Compatibility: ptrString("BACKWARD"),
					},
				},
			},
			expectedDiffs: []string{"Spec.Compatibility"},
		},
		{
			name: "mutable field change - Description",
			desired: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName:  ptrString("test-schema"),
						DataFormat:  ptrString("AVRO"),
						Description: ptrString("Updated description"),
					},
				},
			},
			latest: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName:  ptrString("test-schema"),
						DataFormat:  ptrString("AVRO"),
						Description: ptrString("Old description"),
					},
				},
			},
			expectedDiffs: []string{"Spec.Description"},
		},
		{
			name: "tag changes",
			desired: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("test-schema"),
						DataFormat: ptrString("AVRO"),
						Tags: map[string]*string{
							"Environment": ptrString("production"),
							"Team":        ptrString("data-engineering"),
						},
					},
				},
			},
			latest: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("test-schema"),
						DataFormat: ptrString("AVRO"),
						Tags: map[string]*string{
							"Environment": ptrString("development"),
						},
					},
				},
			},
			expectedDiffs: []string{"Spec.Tags"},
		},
		{
			name: "status changes",
			desired: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("test-schema"),
						DataFormat: ptrString("AVRO"),
					},
					Status: svcapitypes.Schema_SDK_Status{
						SchemaARN: ptrString("arn:aws:glue:us-east-1:123456789012:schema/registry/test-schema"),
					},
				},
			},
			latest: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("test-schema"),
						DataFormat: ptrString("AVRO"),
					},
					Status: svcapitypes.Schema_SDK_Status{
						SchemaARN: ptrString("arn:aws:glue:us-east-1:123456789012:schema/registry/old-schema"),
					},
				},
			},
			expectedDiffs: []string{"Status.SchemaARN"},
		},
		{
			name: "version metadata changes",
			desired: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("test-schema"),
						DataFormat: ptrString("AVRO"),
					},
					Status: svcapitypes.Schema_SDK_Status{
						LatestVersion: &svcapitypes.SchemaVersionMetadata{
							VersionNumber: ptrInt64(2),
							Status:        ptrString("AVAILABLE"),
						},
					},
				},
			},
			latest: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("test-schema"),
						DataFormat: ptrString("AVRO"),
					},
					Status: svcapitypes.Schema_SDK_Status{
						LatestVersion: &svcapitypes.SchemaVersionMetadata{
							VersionNumber: ptrInt64(1),
							Status:        ptrString("AVAILABLE"),
						},
					},
				},
			},
			expectedDiffs: []string{"Status.LatestVersion.VersionNumber"},
		},
		{
			name: "registry reference change",
			desired: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("test-schema"),
						DataFormat: ptrString("AVRO"),
						RegistryRef: &svcapitypes.RegistryReference{
							Name: ptrString("new-registry"),
						},
					},
				},
			},
			latest: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("test-schema"),
						DataFormat: ptrString("AVRO"),
						RegistryRef: &svcapitypes.RegistryReference{
							Name: ptrString("old-registry"),
						},
					},
				},
			},
			expectedDiffs: []string{"Spec.RegistryRef"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := newResourceDelta(tt.desired, tt.latest)

			if tt.noDifferences {
				assert.Empty(t, delta.Differences, "Expected no differences")
			} else {
				assert.NotEmpty(t, delta.Differences, "Expected differences")

				for _, expectedDiff := range tt.expectedDiffs {
					found := false
					for _, diff := range delta.Differences {
						if diff.Path.Contains(expectedDiff) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected difference not found: %s", expectedDiff)
				}
			}
		})
	}
}

// TestCompareImmutableFields tests immutable field comparison
func TestCompareImmutableFields(t *testing.T) {
	tests := []struct {
		name          string
		desired       *resource
		latest        *resource
		expectChanges bool
		changedFields []string
	}{
		{
			name: "no changes to immutable fields",
			desired: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("test-schema"),
						DataFormat: ptrString("AVRO"),
						RegistryRef: &svcapitypes.RegistryReference{
							Name: ptrString("test-registry"),
						},
					},
				},
			},
			latest: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("test-schema"),
						DataFormat: ptrString("AVRO"),
						RegistryRef: &svcapitypes.RegistryReference{
							Name: ptrString("test-registry"),
						},
					},
				},
			},
			expectChanges: false,
		},
		{
			name: "SchemaName changed",
			desired: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("new-name"),
					},
				},
			},
			latest: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						SchemaName: ptrString("old-name"),
					},
				},
			},
			expectChanges: true,
			changedFields: []string{"Spec.SchemaName"},
		},
		{
			name: "DataFormat changed",
			desired: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						DataFormat: ptrString("JSON"),
					},
				},
			},
			latest: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						DataFormat: ptrString("AVRO"),
					},
				},
			},
			expectChanges: true,
			changedFields: []string{"Spec.DataFormat"},
		},
		{
			name: "RegistryRef changed",
			desired: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						RegistryRef: &svcapitypes.RegistryReference{
							Name: ptrString("registry-2"),
						},
					},
				},
			},
			latest: &resource{
				ko: &svcapitypes.Schema{
					Spec: svcapitypes.SchemaSpec{
						RegistryRef: &svcapitypes.RegistryReference{
							Name: ptrString("registry-1"),
						},
					},
				},
			},
			expectChanges: true,
			changedFields: []string{"Spec.RegistryRef"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := newResourceDelta(tt.desired, tt.latest)

			hasImmutableChanges := false
			for _, diff := range delta.Differences {
				for _, field := range tt.changedFields {
					if diff.Path.Contains(field) {
						hasImmutableChanges = true
						break
					}
				}
			}

			assert.Equal(t, tt.expectChanges, hasImmutableChanges)
		})
	}
}

// Helper functions
func ptrInt64(i int64) *int64 {
	return &i
}
