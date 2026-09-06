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
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestCapacityProviderConfigurationRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	input := bedrockagentcorecontrol.CreateCapacityProviderInput{
		Name: aws.String("test_capacity"), Description: aws.String(names.AttrDescription),
		PermissionsConfiguration: &awstypes.PermissionsConfiguration{CapacityProviderOperatorRoleArn: aws.String("arn:aws:iam::123456789012:role/operator")},
		ComputeConfiguration: &awstypes.ComputeConfigurationMemberEc2Configuration{Value: awstypes.Ec2Configuration{
			LaunchTemplateSource: &awstypes.LaunchTemplateSourceMemberLaunchParameters{Value: awstypes.LaunchParameters{
				OperatingSystem:      awstypes.OperatingSystemLinuxArm64,
				InstanceRequirements: &awstypes.InstanceRequirements{AllowedInstanceTypes: []string{"t4g.nano"}},
				InstanceProfileArn:   aws.String("arn:aws:iam::123456789012:instance-profile/test"),
				SshKeyName:           aws.String("test-key"), Monitoring: awstypes.MonitoringBasic,
				PropagatedTags:                   map[string]string{"Name": "test"},
				CapacityReservationSpecification: &awstypes.CapacityReservationSpecification{CapacityReservationPreference: awstypes.CapacityReservationPreferenceOpen, CapacityReservationTarget: &awstypes.CapacityReservationTarget{CapacityReservationId: aws.String("cr-12345678")}},
				EphemeralVolumes:                 []awstypes.EphemeralBlockDeviceMapping{{DeviceName: aws.String("/dev/sdb"), VirtualName: aws.String("ephemeral0")}},
				LicenseSpecifications:            []awstypes.LicenseSpecification{{LicenseConfigurationArn: aws.String("arn:aws:license-manager:us-east-1:123456789012:license-configuration:lic-test")}},
			}},
			VpcConfiguration:       &awstypes.VpcConfiguration{Subnets: []string{"subnet-12345678"}, SecurityGroups: []string{"sg-12345678"}},
			LifecycleConfiguration: &awstypes.InstanceLifecycleConfiguration{IdleInstanceTimeout: aws.Int32(900), MaxLifetime: aws.Int32(28800)},
			RootVolume:             &awstypes.RootVolumeConfiguration{Encrypted: aws.Bool(true), FreeSpaceGiB: aws.Int32(8), Iops: aws.Int32(3000), KmsKeyId: aws.String("alias/test"), Throughput: aws.Int32(125), VolumeType: awstypes.EbsVolumeTypeGp3},
			Volumes:                []awstypes.VolumeConfiguration{&awstypes.VolumeConfigurationMemberEbsConfiguration{Value: awstypes.EbsVolumeConfiguration{Name: aws.String("data"), SizeGiB: aws.Int32(1), Encrypted: aws.Bool(true), Iops: aws.Int32(3000), KmsKeyId: aws.String("alias/test"), SnapshotId: aws.String("snap-12345678"), Throughput: aws.Int32(125), VolumeType: awstypes.EbsVolumeTypeGp3}}},
		}},
	}
	var model capacityProviderResourceModel
	if diags := flex.Flatten(ctx, input, &model, flex.WithFieldNamePrefix("CapacityProvider")); diags.HasError() {
		t.Fatal(diags)
	}
	var got bedrockagentcorecontrol.CreateCapacityProviderInput
	if diags := flex.Expand(ctx, model, &got, flex.WithFieldNamePrefix("CapacityProvider")); diags.HasError() {
		t.Fatal(diags)
	}
	if !reflect.DeepEqual(input, got) {
		t.Errorf("capacity provider configuration changed during round trip: expected %#v, got %#v", input, got)
	}
}
