// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestFlattenServiceVolumeConfigurations(t *testing.T) {
	t.Parallel()

	configurations := []awstypes.ServiceVolumeConfiguration{
		{
			Name: aws.String("s3files-volume"),
		},
		{
			Name: aws.String("ebs-volume"),
			ManagedEBSVolume: &awstypes.ServiceManagedEBSVolumeConfiguration{
				RoleArn: aws.String("arn:aws:iam::123456789012:role/ecs-volume-role"), // lintignore:AWSAT005
			},
		},
	}

	result := flattenServiceVolumeConfigurations(context.Background(), configurations)

	if got, want := len(result), 1; got != want {
		t.Fatalf("expected %d EBS volume configuration, got %d", want, got)
	}

	configuration := result[0].(map[string]any)
	if got, want := configuration[names.AttrName], "ebs-volume"; got != want {
		t.Errorf("expected volume name %q, got %q", want, got)
	}
	if _, ok := configuration["managed_ebs_volume"]; !ok {
		t.Error("expected managed EBS volume configuration")
	}
}
