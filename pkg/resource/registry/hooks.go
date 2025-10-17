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

package registry

import (
	"context"
	"fmt"

	svcapitypes "github.com/aws-controllers-k8s/glue-controller/apis/v1alpha1"
	util "github.com/aws-controllers-k8s/glue-controller/pkg/tags"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/glue"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
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

// registryARN returns the ARN of the Glue registry with the given name.
func registryARN(registry *svcapitypes.Registry) string {
	return fmt.Sprintf(
		"arn:aws:glue:%s:%s:registry/%s",
		*registry.Status.ACKResourceMetadata.Region,
		*registry.Status.ACKResourceMetadata.OwnerAccountID,
		*registry.Spec.RegistryName,
	)
}

// preCreate performs pre-creation validation and setup
func (rm *resourceManager) preCreate(
	ctx context.Context,
	r *resource,
) error {
	// No specific pre-create hooks needed for Registry
	return nil
}

// postCreate performs post-creation processing
func (rm *resourceManager) postCreate(
	ctx context.Context,
	r *resource,
) error {
	// Sync tags after creation if needed
	return nil
}

// preUpdate performs pre-update validation
func (rm *resourceManager) preUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
) error {
	// Validate that RegistryName hasn't changed (immutable field)
	if desired.ko.Spec.RegistryName != nil && latest.ko.Spec.RegistryName != nil {
		if *desired.ko.Spec.RegistryName != *latest.ko.Spec.RegistryName {
			return fmt.Errorf("registry name is immutable and cannot be changed")
		}
	}
	return nil
}

// postUpdate performs post-update processing
func (rm *resourceManager) postUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
) error {
	// No specific post-update hooks needed for Registry
	return nil
}

// preDelete performs pre-deletion validation including:
// - Checking deletion policy
// - Verifying no dependent schemas exist (if DeletionPolicy is Delete)
func (rm *resourceManager) preDelete(
	ctx context.Context,
	r *resource,
) error {
	rlog := ackrtlog.FromContext(ctx)

	// Check deletion policy
	deletionPolicy := "Delete" // default
	if r.ko.Spec.DeletionPolicy != nil {
		deletionPolicy = *r.ko.Spec.DeletionPolicy
	}

	if deletionPolicy == "Retain" {
		// Don't delete from AWS, just remove finalizer
		rlog.Info("deletion policy is Retain, skipping AWS deletion")
		return nil
	}

	// For Delete policy, check for dependent schemas
	if r.ko.Status.DependentSchemaCount != nil && *r.ko.Status.DependentSchemaCount > 0 {
		return fmt.Errorf(
			"cannot delete registry with %d dependent schemas; delete schemas first or set deletionPolicy to Retain",
			*r.ko.Status.DependentSchemaCount,
		)
	}

	return nil
}

// postDelete performs post-deletion cleanup
func (rm *resourceManager) postDelete(
	ctx context.Context,
	r *resource,
) error {
	// No specific post-delete hooks needed for Registry
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

// countDependentSchemas returns the number of schemas associated with this registry
func (rm *resourceManager) countDependentSchemas(
	ctx context.Context,
	r *resource,
) (int64, error) {
	rlog := ackrtlog.FromContext(ctx)

	// Get the registry ARN - need to check both status and construct from metadata
	var registryARNStr string
	if r.ko.Status.RegistryARN != nil {
		registryARNStr = *r.ko.Status.RegistryARN
	} else if r.ko.Status.ACKResourceMetadata != nil &&
		r.ko.Status.ACKResourceMetadata.Region != nil &&
		r.ko.Status.ACKResourceMetadata.OwnerAccountID != nil &&
		r.ko.Spec.RegistryName != nil {
		// Construct ARN if not in status yet
		registryARNStr = registryARN(r.ko)
	} else {
		// If we don't have enough information, there can't be any schemas yet
		return 0, nil
	}

	// List schemas for this registry
	input := &svcsdk.ListSchemasInput{
		RegistryId: &svcsdktypes.RegistryId{
			RegistryArn: &registryARNStr,
		},
		MaxResults: aws.Int32(100), // AWS maximum per page
	}

	var count int64 = 0
	paginator := svcsdk.NewListSchemasPaginator(rm.sdkapi, input)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			rlog.Info("failed to list schemas for registry", "error", err)
			return 0, err
		}

		if resp.Schemas != nil {
			count += int64(len(resp.Schemas))
		}
	}

	return count, nil
}
