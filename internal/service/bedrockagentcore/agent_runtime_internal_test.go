// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

func TestAgentRuntimeCapacityProviderRoundTrip(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		input bedrockagentcorecontrol.CreateAgentRuntimeInput
	}{
		{name: "microVM", input: bedrockagentcorecontrol.CreateAgentRuntimeInput{NetworkConfiguration: &awstypes.NetworkConfiguration{NetworkMode: awstypes.NetworkModePublic}}},
		{name: "instances", input: bedrockagentcorecontrol.CreateAgentRuntimeInput{
			CapacityProviderConfiguration: &awstypes.CapacityProviderConfiguration{CapacityProviderArn: aws.String("arn:aws:bedrock-agentcore:eu-west-1:123456789012:capacity-provider/example-abc1234567")},
			FilesystemConfigurations:      []awstypes.FilesystemConfiguration{&awstypes.FilesystemConfigurationMemberCapacityProviderVolume{Value: awstypes.CapacityProviderVolumeConfiguration{VolumeName: aws.String("data"), MountPath: aws.String("/mnt/data")}}},
		}},
		{name: "efs", input: bedrockagentcorecontrol.CreateAgentRuntimeInput{
			FilesystemConfigurations: []awstypes.FilesystemConfiguration{&awstypes.FilesystemConfigurationMemberEfsAccessPoint{Value: awstypes.EfsAccessPointConfiguration{AccessPointArn: aws.String("arn:aws:elasticfilesystem:eu-west-1:123456789012:access-point/fsap-1234567890abcdef0"), MountPath: aws.String("/mnt/efs")}}},
		}},
		{name: "s3Files", input: bedrockagentcorecontrol.CreateAgentRuntimeInput{
			FilesystemConfigurations: []awstypes.FilesystemConfiguration{&awstypes.FilesystemConfigurationMemberS3FilesAccessPoint{Value: awstypes.S3FilesAccessPointConfiguration{AccessPointArn: aws.String("arn:aws:s3:eu-west-1:123456789012:accesspoint/example"), MountPath: aws.String("/mnt/s3")}}},
		}},
		{name: "sessionStorage", input: bedrockagentcorecontrol.CreateAgentRuntimeInput{
			FilesystemConfigurations: []awstypes.FilesystemConfiguration{&awstypes.FilesystemConfigurationMemberSessionStorage{Value: awstypes.SessionStorageConfiguration{MountPath: aws.String("/mnt/session")}}},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			var model agentRuntimeResourceModel
			if diags := flex.Flatten(ctx, tc.input, &model); diags.HasError() {
				t.Fatal(diags)
			}
			var got bedrockagentcorecontrol.CreateAgentRuntimeInput
			if diags := flex.Expand(ctx, model, &got); diags.HasError() {
				t.Fatal(diags)
			}
			if !reflect.DeepEqual(tc.input, got) {
				t.Fatalf("runtime configuration changed during round trip: expected %#v, got %#v", tc.input, got)
			}
		})
	}
}
