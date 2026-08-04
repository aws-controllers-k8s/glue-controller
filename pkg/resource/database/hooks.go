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

package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/glue/types"

	svcapitypes "github.com/aws-controllers-k8s/glue-controller/apis/v1alpha1"
	util "github.com/aws-controllers-k8s/glue-controller/pkg/tags"
)

// buildDatabaseInput constructs a DatabaseInput struct from the resource spec,
// used for both CreateDatabase and UpdateDatabase API calls.
func (rm *resourceManager) buildDatabaseInput(
	r *resource,
) (*svcsdktypes.DatabaseInput, error) {
	databaseInput := &svcsdktypes.DatabaseInput{}
	if r.ko.Spec.Name != nil {
		databaseInput.Name = r.ko.Spec.Name
	}
	if r.ko.Spec.Description != nil {
		databaseInput.Description = r.ko.Spec.Description
	}
	if r.ko.Spec.LocationURI != nil {
		databaseInput.LocationUri = r.ko.Spec.LocationURI
	}
	if r.ko.Spec.Parameters != nil {
		databaseInput.Parameters = aws.ToStringMap(r.ko.Spec.Parameters)
	}
	if r.ko.Spec.TableDefaultPermissions != nil {
		perms := []svcsdktypes.PrincipalPermissions{}
		for _, p := range r.ko.Spec.TableDefaultPermissions {
			elem := svcsdktypes.PrincipalPermissions{}
			if p.Permissions != nil {
				pp := []svcsdktypes.Permission{}
				for _, perm := range p.Permissions {
					pp = append(pp, svcsdktypes.Permission(*perm))
				}
				elem.Permissions = pp
			}
			if p.Principal != nil {
				elem.Principal = &svcsdktypes.DataLakePrincipal{
					DataLakePrincipalIdentifier: p.Principal.DataLakePrincipalIdentifier,
				}
			}
			perms = append(perms, elem)
		}
		databaseInput.CreateTableDefaultPermissions = perms
	}
	if r.ko.Spec.FederatedDatabase != nil {
		databaseInput.FederatedDatabase = &svcsdktypes.FederatedDatabase{
			ConnectionName: r.ko.Spec.FederatedDatabase.ConnectionName,
			Identifier:     r.ko.Spec.FederatedDatabase.Identifier,
		}
	}
	return databaseInput, nil
}

// getTags retrieves the resource's associated tags.
func (rm *resourceManager) getTags(
	ctx context.Context,
	resourceARN string,
) (map[string]*string, error) {
	return util.GetResourceTags(ctx, rm.sdkapi, rm.metrics, resourceARN)
}

// syncTags keeps the resource's tags in sync.
func (rm *resourceManager) syncTags(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	return util.SyncResourceTags(ctx, rm.sdkapi, rm.metrics, string(*latest.ko.Status.ACKResourceMetadata.ARN), desired.ko.Spec.Tags, latest.ko.Spec.Tags, convertToOrderedACKTags)
}

// databaseARN returns the ARN of the Glue database with the given name.
func databaseARN(database *svcapitypes.Database) string {
	// TODO(a-hilaly): I know there could be other partitions, but I'm
	// not sure how to determine at this level of abstraction. Probably
	// something the SDK/runtime should handle. For now, we'll just use
	// the `aws` partition.
	if database.Status.ACKResourceMetadata == nil ||
		database.Status.ACKResourceMetadata.Region == nil ||
		database.Status.ACKResourceMetadata.OwnerAccountID == nil ||
		database.Spec.Name == nil {
		return ""
	}
	return fmt.Sprintf("arn:aws:glue:%s:%s:database/%s", *database.Status.ACKResourceMetadata.Region, *database.Status.ACKResourceMetadata.OwnerAccountID, *database.Spec.Name)
}
