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
	"context"
	"fmt"

	svcapitypes "github.com/aws-controllers-k8s/glue-controller/apis/v1alpha1"
	util "github.com/aws-controllers-k8s/glue-controller/pkg/tags"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
)

// getTags retrieves the resource's associated tags.
func (rm *resourceManager) getTags(
	ctx context.Context,
	resourceARN string,
) (map[string]*string, error) {
	tags, err := util.GetResourceTags(ctx, rm.sdkapi, rm.metrics, resourceARN)
	if err != nil {
		return nil, err
	}

	// util.GetResourceTags already returns map[string]*string, so just return it
	return tags, nil
}

// syncTags keeps the resource's tags in sync.
func (rm *resourceManager) syncTags(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	// Tags are already in map[string]*string format, so no conversion needed
	return util.SyncResourceTags(
		ctx,
		rm.sdkapi,
		rm.metrics,
		string(*latest.ko.Status.ACKResourceMetadata.ARN),
		desired.ko.Spec.Tags,
		latest.ko.Spec.Tags,
		convertToOrderedACKTags,
	)
}

// schemaARN returns the ARN of the Glue schema.
func schemaARN(schema *svcapitypes.Schema) string {
	// Schema ARN format: arn:aws:glue:{region}:{account}:schema/{registry-name}/{schema-name}
	// We can get this from the Status.SchemaARN which is populated after creation
	if schema.Status.SchemaARN != nil {
		return *schema.Status.SchemaARN
	}

	// If not in status yet, construct from metadata and name
	if schema.Status.ACKResourceMetadata != nil &&
		schema.Status.ACKResourceMetadata.Region != nil &&
		schema.Status.ACKResourceMetadata.OwnerAccountID != nil &&
		schema.Status.RegistryARN != nil &&
		schema.Spec.SchemaName != nil {
		// Extract registry name from registry ARN
		// Format: arn:aws:glue:region:account:registry/registry-name
		registryARN := *schema.Status.RegistryARN
		return fmt.Sprintf(
			"arn:aws:glue:%s:%s:schema/%s/%s",
			*schema.Status.ACKResourceMetadata.Region,
			*schema.Status.ACKResourceMetadata.OwnerAccountID,
			registryARN, // This should be just the registry name, not full ARN
			*schema.Spec.SchemaName,
		)
	}

	return ""
}

// preCreate performs pre-creation validation and setup
func (rm *resourceManager) preCreate(
	ctx context.Context,
	r *resource,
) error {
	rlog := ackrtlog.FromContext(ctx)

	// Run comprehensive spec validation
	if err := rm.validateSchemaSpec(r.ko); err != nil {
		return fmt.Errorf("schema validation failed: %v", err)
	}

	// Validate that RegistryRef has been resolved
	if r.ko.Spec.RegistryRef != nil {
		if r.ko.Status.RegistryARN == nil {
			rlog.Info("registryRef not yet resolved, will be resolved by ResolveReferences hook")
			// The ResolveReferences hook should have populated Status.RegistryARN
			// If it's still nil here, something went wrong
			return fmt.Errorf("registryRef must be resolved before creation")
		}
	}

	return nil
}

// postCreate performs post-creation processing including:
// - Logging successful creation
// - Updating status with initial version information
func (rm *resourceManager) postCreate(
	ctx context.Context,
	r *resource,
) error {
	rlog := ackrtlog.FromContext(ctx)

	// Log successful creation with key details
	if r.ko.Spec.SchemaName != nil {
		rlog.Info("schema created successfully",
			"schemaName", *r.ko.Spec.SchemaName,
			"dataFormat", r.ko.Spec.DataFormat,
		)
	}

	// Track initial schema version if available
	if r.ko.Status.LatestVersion != nil && r.ko.Status.LatestVersion.VersionNumber != nil {
		rlog.Info("initial schema version created",
			"versionNumber", *r.ko.Status.LatestVersion.VersionNumber,
		)
	}

	return nil
}

// preUpdate performs pre-update validation
func (rm *resourceManager) preUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
) error {
	// Run comprehensive update validation including immutable field checks
	if err := rm.validateSchemaUpdate(desired.ko, latest.ko); err != nil {
		return fmt.Errorf("schema update validation failed: %v", err)
	}

	// Re-validate the desired spec to ensure all fields are still valid
	if err := rm.validateSchemaSpec(desired.ko); err != nil {
		return fmt.Errorf("desired schema validation failed: %v", err)
	}

	return nil
}

// postUpdate performs post-update processing including:
// - Logging successful updates
// - Tracking schema version changes
// - Validating compatibility mode enforcement
func (rm *resourceManager) postUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
) error {
	rlog := ackrtlog.FromContext(ctx)

	// Log the update with relevant details
	if latest.ko.Spec.SchemaName != nil {
		rlog.Info("schema updated successfully",
			"schemaName", *latest.ko.Spec.SchemaName,
		)
	}

	// Track schema version changes
	desiredVersion := int64(0)
	latestVersion := int64(0)
	if desired.ko.Status.LatestVersion != nil && desired.ko.Status.LatestVersion.VersionNumber != nil {
		desiredVersion = *desired.ko.Status.LatestVersion.VersionNumber
	}
	if latest.ko.Status.LatestVersion != nil && latest.ko.Status.LatestVersion.VersionNumber != nil {
		latestVersion = *latest.ko.Status.LatestVersion.VersionNumber
	}

	if desiredVersion != latestVersion && latestVersion != 0 {
		rlog.Info("schema version updated",
			"previousVersion", desiredVersion,
			"newVersion", latestVersion,
		)
	}

	// Log compatibility mode if changed
	if desired.ko.Spec.Compatibility != nil && latest.ko.Spec.Compatibility != nil {
		if *desired.ko.Spec.Compatibility != *latest.ko.Spec.Compatibility {
			rlog.Info("schema compatibility mode changed",
				"previousMode", *desired.ko.Spec.Compatibility,
				"newMode", *latest.ko.Spec.Compatibility,
			)
		}
	}

	return nil
}

// preDelete performs pre-deletion validation including:
// - Checking deletion policy
// - Logging deletion intent
// - Warning about data loss
func (rm *resourceManager) preDelete(
	ctx context.Context,
	r *resource,
) error {
	rlog := ackrtlog.FromContext(ctx)

	// Log deletion intent with schema details
	if r.ko.Spec.SchemaName != nil {
		rlog.Info("preparing to delete schema",
			"schemaName", *r.ko.Spec.SchemaName,
			"dataFormat", r.ko.Spec.DataFormat,
		)
	}

	// Check deletion policy
	deletionPolicy := "Delete" // default
	if r.ko.Spec.DeletionPolicy != nil {
		deletionPolicy = *r.ko.Spec.DeletionPolicy
	}

	if deletionPolicy == "Retain" {
		// Don't delete from AWS, just remove finalizer
		rlog.Info("deletion policy is Retain, skipping AWS deletion",
			"schemaName", *r.ko.Spec.SchemaName,
		)
		return nil
	}

	// Warn about potential data loss
	if r.ko.Status.LatestVersion != nil && r.ko.Status.LatestVersion.VersionNumber != nil {
		rlog.Info("deleting schema will remove all versions",
			"schemaName", *r.ko.Spec.SchemaName,
			"latestVersion", *r.ko.Status.LatestVersion.VersionNumber,
		)
	}

	// For Schema, there are no dependent resources to check
	// (Schema is a leaf node in the resource hierarchy)
	// However, warn that this schema might be referenced by other systems
	rlog.Info("ensure no external systems depend on this schema before deletion")

	return nil
}

// postDelete performs post-deletion cleanup including:
// - Logging successful deletion
// - Clearing any cached references
func (rm *resourceManager) postDelete(
	ctx context.Context,
	r *resource,
) error {
	rlog := ackrtlog.FromContext(ctx)

	// Log successful deletion
	if r.ko.Spec.SchemaName != nil {
		rlog.Info("schema deleted successfully",
			"schemaName", *r.ko.Spec.SchemaName,
		)
	}

	// Clear any resolved references in status
	// (This ensures clean state if deletion fails and is retried)
	r.ko.Status.RegistryARN = nil
	r.ko.Status.SchemaARN = nil
	r.ko.Status.LatestVersion = nil

	return nil
}

// shouldDeleteFromAWS returns true if the resource should be deleted from AWS
// based on the deletion policy
func (rm *resourceManager) shouldDeleteFromAWS(r *resource) bool {
	// Default to Delete if not specified
	if r.ko.Spec.DeletionPolicy == nil {
		return true
	}
	return *r.ko.Spec.DeletionPolicy == "Delete"
}

// updateConditions updates the resource's conditions based on the operation result
// This is called by the ACK runtime after operations
func (rm *resourceManager) updateConditions(
	r *resource,
	success bool,
	err error,
) (*resource, bool) {
	// Condition management is primarily handled by ACK runtime
	// Custom status updates are handled in status.go
	return r, false
}
