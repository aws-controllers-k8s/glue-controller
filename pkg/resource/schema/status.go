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

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/aws-controllers-k8s/glue-controller/apis/v1alpha1"
)

// isSchemaReady determines if the schema is fully ready for use
func (rm *resourceManager) isSchemaReady(r *resource) bool {
	// Schema is ready if:
	// 1. It has a SchemaARN in status
	// 2. It has latest version information
	// 3. The latest version is AVAILABLE

	if r.ko.Status.SchemaARN == nil || *r.ko.Status.SchemaARN == "" {
		return false
	}

	if r.ko.Status.LatestVersion == nil {
		return false
	}

	if r.ko.Status.LatestVersion.Status == nil {
		return false
	}

	// Check version status
	versionStatus := *r.ko.Status.LatestVersion.Status
	return versionStatus == "AVAILABLE"
}

// setReadyCondition sets or clears the Ready condition
func setReadyCondition(ko *svcapitypes.Schema, ready bool, err error) {
	var condition *ackv1alpha1.Condition

	if ready {
		condition = &ackv1alpha1.Condition{
			Type:   ackv1alpha1.ConditionTypeResourceSynced,
			Status: corev1.ConditionTrue,
		}
	} else {
		message := "Schema is not ready"
		if err != nil {
			message = err.Error()
		}
		condition = &ackv1alpha1.Condition{
			Type:    ackv1alpha1.ConditionTypeResourceSynced,
			Status:  corev1.ConditionFalse,
			Message: &message,
		}
	}

	// Update or add the condition
	if ko.Status.Conditions == nil {
		ko.Status.Conditions = []*ackv1alpha1.Condition{}
	}

	// Find and update existing condition or append new one
	updated := false
	for i, existing := range ko.Status.Conditions {
		if existing.Type == condition.Type {
			ko.Status.Conditions[i] = condition
			updated = true
			break
		}
	}

	if !updated {
		ko.Status.Conditions = append(ko.Status.Conditions, condition)
	}
}

// updateStatusFromAWSResponse updates the resource status from an AWS API response
// This is called after successful Read, Create, or Update operations
func (rm *resourceManager) updateStatusFromAWSResponse(
	ctx context.Context,
	r *resource,
) error {
	rlog := ackrtlog.FromContext(ctx)

	// Ensure Status.ACKResourceMetadata exists
	if r.ko.Status.ACKResourceMetadata == nil {
		r.ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}

	// Update ACKResourceMetadata with region and account
	if r.ko.Status.ACKResourceMetadata.Region == nil {
		region := ackv1alpha1.AWSRegion(rm.awsRegion)
		r.ko.Status.ACKResourceMetadata.Region = &region
	}

	if r.ko.Status.ACKResourceMetadata.OwnerAccountID == nil {
		accountID := ackv1alpha1.AWSAccountID(rm.awsAccountID)
		r.ko.Status.ACKResourceMetadata.OwnerAccountID = &accountID
	}

	// Update ARN if available
	if r.ko.Status.SchemaARN != nil {
		arn := ackv1alpha1.AWSResourceName(*r.ko.Status.SchemaARN)
		r.ko.Status.ACKResourceMetadata.ARN = &arn
	}

	// Log status update
	if r.ko.Status.LatestVersion != nil && r.ko.Status.LatestVersion.VersionNumber != nil {
		rlog.Info("updated schema status from AWS",
			"schemaARN", r.ko.Status.SchemaARN,
			"latestVersion", *r.ko.Status.LatestVersion.VersionNumber,
		)
	}

	return nil
}

// setVersionPendingCondition sets a condition indicating a new version is being created
func (rm *resourceManager) setVersionPendingCondition(
	ctx context.Context,
	r *resource,
) {
	rlog := ackrtlog.FromContext(ctx)

	message := "New schema version is being created"
	condition := &ackv1alpha1.Condition{
		Type:    "VersionPending",
		Status:  corev1.ConditionTrue,
		Message: &message,
	}

	// Add condition to status
	if r.ko.Status.Conditions == nil {
		r.ko.Status.Conditions = []*ackv1alpha1.Condition{}
	}

	// Update or add condition
	updated := false
	for i, existing := range r.ko.Status.Conditions {
		if existing.Type == condition.Type {
			r.ko.Status.Conditions[i] = condition
			updated = true
			break
		}
	}

	if !updated {
		r.ko.Status.Conditions = append(r.ko.Status.Conditions, condition)
	}

	rlog.Info("set version pending condition")
}

// clearVersionPendingCondition removes the version pending condition
func (rm *resourceManager) clearVersionPendingCondition(
	ctx context.Context,
	r *resource,
) {
	if r.ko.Status.Conditions == nil {
		return
	}

	// Remove VersionPending condition
	newConditions := []*ackv1alpha1.Condition{}
	for _, condition := range r.ko.Status.Conditions {
		if condition.Type != "VersionPending" {
			newConditions = append(newConditions, condition)
		}
	}

	r.ko.Status.Conditions = newConditions
}

// updateVersionStatus updates the version-related status fields
func (rm *resourceManager) updateVersionStatus(
	ctx context.Context,
	r *resource,
	versionNumber int64,
	versionID string,
	status string,
	createdTime *metav1.Time,
) {
	rlog := ackrtlog.FromContext(ctx)

	// Initialize LatestVersion if needed
	if r.ko.Status.LatestVersion == nil {
		r.ko.Status.LatestVersion = &svcapitypes.SchemaVersionMetadata{}
	}

	// Update version fields
	r.ko.Status.LatestVersion.VersionNumber = &versionNumber
	r.ko.Status.LatestVersion.SchemaVersionID = &versionID
	r.ko.Status.LatestVersion.Status = &status
	if createdTime != nil {
		r.ko.Status.LatestVersion.CreatedTime = createdTime
	}

	rlog.Info("updated version status",
		"versionNumber", versionNumber,
		"versionID", versionID,
		"versionStatus", status,
	)

	// Clear version pending condition if version is available
	if status == "AVAILABLE" {
		rm.clearVersionPendingCondition(ctx, r)
	} else if status == "PENDING" {
		rm.setVersionPendingCondition(ctx, r)
	}
}

// setReferenceNotResolvedCondition sets a condition indicating reference resolution is pending
func (rm *resourceManager) setReferenceNotResolvedCondition(
	ctx context.Context,
	r *resource,
	refName string,
) {
	rlog := ackrtlog.FromContext(ctx)
	message := fmt.Sprintf("Waiting for Registry reference '%s' to be resolved", refName)
	condition := &ackv1alpha1.Condition{
		Type:    "ReferenceNotResolved",
		Status:  corev1.ConditionTrue,
		Message: &message,
	}

	if r.ko.Status.Conditions == nil {
		r.ko.Status.Conditions = []*ackv1alpha1.Condition{}
	}

	// Update or add condition
	updated := false
	for i, existing := range r.ko.Status.Conditions {
		if existing.Type == condition.Type {
			r.ko.Status.Conditions[i] = condition
			updated = true
			break
		}
	}

	if !updated {
		r.ko.Status.Conditions = append(r.ko.Status.Conditions, condition)
	}

	rlog.Info("set reference not resolved condition",
		"registryRef", refName,
	)
}

// clearReferenceNotResolvedCondition clears the reference resolution pending condition
func (rm *resourceManager) clearReferenceNotResolvedCondition(
	ctx context.Context,
	r *resource,
) {
	rlog := ackrtlog.FromContext(ctx)
	if r.ko.Status.Conditions == nil {
		return
	}

	cleared := false
	newConditions := []*ackv1alpha1.Condition{}
	for _, condition := range r.ko.Status.Conditions {
		if condition.Type != "ReferenceNotResolved" {
			newConditions = append(newConditions, condition)
		} else {
			cleared = true
		}
	}

	r.ko.Status.Conditions = newConditions

	if cleared {
		rlog.Info("cleared reference not resolved condition")
	}
}
