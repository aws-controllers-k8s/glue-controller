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
	"context"
	"testing"

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/aws-controllers-k8s/glue-controller/apis/v1alpha1"
)

// TestIsSchemaReady tests the schema readiness determination
func TestIsSchemaReady(t *testing.T) {
	tests := []struct {
		name     string
		resource *resource
		expected bool
	}{
		{
			name: "fully ready schema",
			resource: &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{
						SchemaARN: ptrString("arn:aws:glue:us-east-1:123456789012:schema/registry/test-schema"),
						LatestVersion: &svcapitypes.SchemaVersionMetadata{
							VersionNumber: ptrInt64(1),
							Status:        ptrString("AVAILABLE"),
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "no SchemaARN",
			resource: &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{
						SchemaARN: nil,
						LatestVersion: &svcapitypes.SchemaVersionMetadata{
							VersionNumber: ptrInt64(1),
							Status:        ptrString("AVAILABLE"),
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "empty SchemaARN",
			resource: &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{
						SchemaARN: ptrString(""),
						LatestVersion: &svcapitypes.SchemaVersionMetadata{
							VersionNumber: ptrInt64(1),
							Status:        ptrString("AVAILABLE"),
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "no LatestVersion",
			resource: &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{
						SchemaARN:     ptrString("arn:aws:glue:us-east-1:123456789012:schema/registry/test-schema"),
						LatestVersion: nil,
					},
				},
			},
			expected: false,
		},
		{
			name: "no version status",
			resource: &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{
						SchemaARN: ptrString("arn:aws:glue:us-east-1:123456789012:schema/registry/test-schema"),
						LatestVersion: &svcapitypes.SchemaVersionMetadata{
							VersionNumber: ptrInt64(1),
							Status:        nil,
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "version status PENDING",
			resource: &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{
						SchemaARN: ptrString("arn:aws:glue:us-east-1:123456789012:schema/registry/test-schema"),
						LatestVersion: &svcapitypes.SchemaVersionMetadata{
							VersionNumber: ptrInt64(1),
							Status:        ptrString("PENDING"),
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "version status FAILURE",
			resource: &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{
						SchemaARN: ptrString("arn:aws:glue:us-east-1:123456789012:schema/registry/test-schema"),
						LatestVersion: &svcapitypes.SchemaVersionMetadata{
							VersionNumber: ptrInt64(1),
							Status:        ptrString("FAILURE"),
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "version status DELETING",
			resource: &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{
						SchemaARN: ptrString("arn:aws:glue:us-east-1:123456789012:schema/registry/test-schema"),
						LatestVersion: &svcapitypes.SchemaVersionMetadata{
							VersionNumber: ptrInt64(1),
							Status:        ptrString("DELETING"),
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := &resourceManager{}
			result := rm.isSchemaReady(tt.resource)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestUpdateVersionStatus tests version status updates
func TestUpdateVersionStatus(t *testing.T) {
	tests := []struct {
		name                       string
		initialResource            *resource
		versionNumber              int64
		versionID                  string
		status                     string
		createdTime                *metav1.Time
		expectedVersion            *svcapitypes.SchemaVersionMetadata
		shouldHavePendingCondition bool
	}{
		{
			name: "update to AVAILABLE status",
			initialResource: &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{},
				},
			},
			versionNumber: 1,
			versionID:     "version-id-1",
			status:        "AVAILABLE",
			createdTime:   &metav1.Time{},
			expectedVersion: &svcapitypes.SchemaVersionMetadata{
				VersionNumber:   ptrInt64(1),
				SchemaVersionID: ptrString("version-id-1"),
				Status:          ptrString("AVAILABLE"),
			},
			shouldHavePendingCondition: false,
		},
		{
			name: "update to PENDING status",
			initialResource: &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{},
				},
			},
			versionNumber: 2,
			versionID:     "version-id-2",
			status:        "PENDING",
			expectedVersion: &svcapitypes.SchemaVersionMetadata{
				VersionNumber:   ptrInt64(2),
				SchemaVersionID: ptrString("version-id-2"),
				Status:          ptrString("PENDING"),
			},
			shouldHavePendingCondition: true,
		},
		{
			name: "update existing version",
			initialResource: &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{
						LatestVersion: &svcapitypes.SchemaVersionMetadata{
							VersionNumber:   ptrInt64(1),
							SchemaVersionID: ptrString("old-version-id"),
							Status:          ptrString("PENDING"),
						},
					},
				},
			},
			versionNumber: 2,
			versionID:     "new-version-id",
			status:        "AVAILABLE",
			expectedVersion: &svcapitypes.SchemaVersionMetadata{
				VersionNumber:   ptrInt64(2),
				SchemaVersionID: ptrString("new-version-id"),
				Status:          ptrString("AVAILABLE"),
			},
			shouldHavePendingCondition: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			rm := &resourceManager{}

			rm.updateVersionStatus(ctx, tt.initialResource, tt.versionNumber, tt.versionID, tt.status, tt.createdTime)

			// Verify LatestVersion was updated
			assert.NotNil(t, tt.initialResource.ko.Status.LatestVersion)
			assert.Equal(t, *tt.expectedVersion.VersionNumber, *tt.initialResource.ko.Status.LatestVersion.VersionNumber)
			assert.Equal(t, *tt.expectedVersion.SchemaVersionID, *tt.initialResource.ko.Status.LatestVersion.SchemaVersionID)
			assert.Equal(t, *tt.expectedVersion.Status, *tt.initialResource.ko.Status.LatestVersion.Status)

			// Verify conditions
			hasPendingCondition := false
			for _, condition := range tt.initialResource.ko.Status.Conditions {
				if condition.Type == "VersionPending" {
					hasPendingCondition = true
					if tt.shouldHavePendingCondition {
						assert.Equal(t, "True", string(condition.Status))
					}
				}
			}
			assert.Equal(t, tt.shouldHavePendingCondition, hasPendingCondition)
		})
	}
}

// TestUpdateStatusFromAWSResponse tests AWS response status updates
func TestUpdateStatusFromAWSResponse(t *testing.T) {
	tests := []struct {
		name              string
		initialResource   *resource
		awsRegion         string
		awsAccountID      string
		expectedRegion    string
		expectedAccountID string
		shouldUpdateARN   bool
	}{
		{
			name: "update with new metadata",
			initialResource: &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{
						ACKResourceMetadata: nil,
					},
				},
			},
			awsRegion:         "us-east-1",
			awsAccountID:      "123456789012",
			expectedRegion:    "us-east-1",
			expectedAccountID: "123456789012",
		},
		{
			name: "update with existing metadata",
			initialResource: &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{
						ACKResourceMetadata: &ackv1alpha1.ResourceMetadata{
							Region:         ptrAWSRegion("us-west-2"),
							OwnerAccountID: ptrAWSAccountID("999999999999"),
						},
					},
				},
			},
			awsRegion:         "us-east-1",
			awsAccountID:      "123456789012",
			expectedRegion:    "us-east-1",    // Should update
			expectedAccountID: "123456789012", // Should update
		},
		{
			name: "update ARN from SchemaARN",
			initialResource: &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{
						SchemaARN: ptrString("arn:aws:glue:us-east-1:123456789012:schema/registry/test-schema"),
					},
				},
			},
			awsRegion:       "us-east-1",
			awsAccountID:    "123456789012",
			shouldUpdateARN: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			rm := &resourceManager{
				awsRegion:    ackv1alpha1.AWSRegion(tt.awsRegion),
				awsAccountID: ackv1alpha1.AWSAccountID(tt.awsAccountID),
			}

			err := rm.updateStatusFromAWSResponse(ctx, tt.initialResource)
			assert.NoError(t, err)

			// Verify ACKResourceMetadata was created/updated
			assert.NotNil(t, tt.initialResource.ko.Status.ACKResourceMetadata)
			assert.NotNil(t, tt.initialResource.ko.Status.ACKResourceMetadata.Region)
			assert.Equal(t, tt.expectedRegion, string(*tt.initialResource.ko.Status.ACKResourceMetadata.Region))
			assert.NotNil(t, tt.initialResource.ko.Status.ACKResourceMetadata.OwnerAccountID)
			assert.Equal(t, tt.expectedAccountID, string(*tt.initialResource.ko.Status.ACKResourceMetadata.OwnerAccountID))

			// Verify ARN was updated if SchemaARN was present
			if tt.shouldUpdateARN {
				assert.NotNil(t, tt.initialResource.ko.Status.ACKResourceMetadata.ARN)
				assert.Equal(t, *tt.initialResource.ko.Status.SchemaARN, string(*tt.initialResource.ko.Status.ACKResourceMetadata.ARN))
			}
		})
	}
}

// TestConditionManagement tests custom condition management
func TestConditionManagement(t *testing.T) {
	tests := []struct {
		name              string
		initialConditions []*ackv1alpha1.Condition
		operation         string
		expectedCondition string
		expectedStatus    string
	}{
		{
			name:              "set VersionPending condition",
			initialConditions: []*ackv1alpha1.Condition{},
			operation:         "setVersionPending",
			expectedCondition: "VersionPending",
			expectedStatus:    "True",
		},
		{
			name: "clear VersionPending condition",
			initialConditions: []*ackv1alpha1.Condition{
				{
					Type:   "VersionPending",
					Status: "True",
				},
			},
			operation:         "clearVersionPending",
			expectedCondition: "VersionPending",
			expectedStatus:    "False",
		},
		{
			name:              "set ReferenceNotResolved condition",
			initialConditions: []*ackv1alpha1.Condition{},
			operation:         "setReferenceNotResolved",
			expectedCondition: "ReferenceNotResolved",
			expectedStatus:    "True",
		},
		{
			name: "clear ReferenceNotResolved condition",
			initialConditions: []*ackv1alpha1.Condition{
				{
					Type:   "ReferenceNotResolved",
					Status: "True",
				},
			},
			operation:         "clearReferenceNotResolved",
			expectedCondition: "ReferenceNotResolved",
			expectedStatus:    "False",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			rm := &resourceManager{}
			r := &resource{
				ko: &svcapitypes.Schema{
					Status: svcapitypes.Schema_SDK_Status{
						Conditions: tt.initialConditions,
					},
				},
			}

			// Execute operation
			switch tt.operation {
			case "setVersionPending":
				rm.setVersionPendingCondition(ctx, r)
			case "clearVersionPending":
				rm.clearVersionPendingCondition(ctx, r)
			case "setReferenceNotResolved":
				rm.setReferenceNotResolvedCondition(ctx, r, "test-registry")
			case "clearReferenceNotResolved":
				rm.clearReferenceNotResolvedCondition(ctx, r)
			}

			// Verify condition was set/cleared
			found := false
			for _, condition := range r.ko.Status.Conditions {
				if condition.Type == ackv1alpha1.ConditionType(tt.expectedCondition) {
					found = true
					assert.Equal(t, tt.expectedStatus, string(condition.Status))
				}
			}

			if tt.expectedStatus == "True" {
				assert.True(t, found, "Expected condition not found")
			}
			// For "False" status, condition might be removed entirely
		})
	}
}

// Helper functions
func ptrAWSRegion(s string) *ackv1alpha1.AWSRegion {
	r := ackv1alpha1.AWSRegion(s)
	return &r
}

func ptrAWSAccountID(s string) *ackv1alpha1.AWSAccountID {
	id := ackv1alpha1.AWSAccountID(s)
	return &id
}
