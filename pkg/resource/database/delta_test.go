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
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"

	svcapitypes "github.com/aws-controllers-k8s/glue-controller/apis/v1alpha1"
)

func newDatabase(description, locationURI *string) *resource {
	return &resource{ko: &svcapitypes.Database{
		Spec: svcapitypes.DatabaseSpec{
			Name:        aws.String("rl_db"),
			Description: description,
			LocationURI: locationURI,
			TargetDatabase: &svcapitypes.DatabaseIdentifier{
				CatalogID:    aws.String("111122223333"),
				DatabaseName: aws.String("producer_db"),
			},
		},
	}}
}

func TestNewResourceDelta_DescriptionAndLocationURI(t *testing.T) {
	tests := []struct {
		name            string
		desired, latest *resource
		wantDiff        bool
	}{
		{
			// GetDatabase returns "" for both fields on a resource link.
			name:    "unset in spec, empty from AWS",
			desired: newDatabase(nil, nil),
			latest:  newDatabase(aws.String(""), aws.String("")),
		},
		{
			name:     "unset in spec, value in AWS",
			desired:  newDatabase(nil, nil),
			latest:   newDatabase(aws.String("old"), aws.String("s3://bucket/old")),
			wantDiff: true,
		},
		{
			name:     "values differ",
			desired:  newDatabase(aws.String("new"), aws.String("s3://bucket/new")),
			latest:   newDatabase(aws.String("old"), aws.String("s3://bucket/old")),
			wantDiff: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := newResourceDelta(tt.desired, tt.latest)
			for _, path := range []string{"Spec.Description", "Spec.LocationURI"} {
				if got := delta.DifferentAt(path); got != tt.wantDiff {
					t.Errorf("DifferentAt(%q) = %v, want %v", path, got, tt.wantDiff)
				}
			}
		})
	}
}
