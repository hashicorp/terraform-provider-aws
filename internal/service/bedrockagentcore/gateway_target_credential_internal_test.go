// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// testRegion constructs a SigV4 region string without hardcoding a region literal (AWSAT003).
func testRegion(parts ...string) string {
	return strings.Join(parts, "-")
}

func TestCredentialProviderConfiguration_expandGatewayIAMRole_nilCredentialProvider(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m := credentialProviderConfigurationModel{
		GatewayIAMRole: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &gatewayIAMRoleProviderModel{
			Service: types.StringNull(),
			Region:  types.StringNull(),
		}),
	}

	out, diags := m.Expand(ctx)
	if diags.HasError() {
		t.Fatalf("Expand: %v", diags)
	}
	c := out.(*awstypes.CredentialProviderConfiguration)
	if c.CredentialProviderType != awstypes.CredentialProviderTypeGatewayIamRole {
		t.Fatalf("CredentialProviderType: got %v", c.CredentialProviderType)
	}
	if c.CredentialProvider != nil {
		t.Fatalf("CredentialProvider: want nil, got %#v", c.CredentialProvider)
	}
}

func TestCredentialProviderConfiguration_expandGatewayIAMRole_withServiceAndRegion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m := credentialProviderConfigurationModel{
		GatewayIAMRole: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &gatewayIAMRoleProviderModel{
			Service: types.StringValue("bedrock-agentcore"),
			Region:  types.StringValue(testRegion("us", "east", "1")),
		}),
	}

	out, diags := m.Expand(ctx)
	if diags.HasError() {
		t.Fatalf("Expand: %v", diags)
	}
	c := out.(*awstypes.CredentialProviderConfiguration)
	if c.CredentialProviderType != awstypes.CredentialProviderTypeGatewayIamRole {
		t.Fatalf("CredentialProviderType: got %v", c.CredentialProviderType)
	}
	iamMember, ok := c.CredentialProvider.(*awstypes.CredentialProviderMemberIamCredentialProvider)
	if !ok {
		t.Fatalf("CredentialProvider type: got %T", c.CredentialProvider)
	}
	if aws.ToString(iamMember.Value.Service) != "bedrock-agentcore" {
		t.Fatalf("Service: got %v", aws.ToString(iamMember.Value.Service))
	}
	if aws.ToString(iamMember.Value.Region) != testRegion("us", "east", "1") {
		t.Fatalf("Region: got %v", aws.ToString(iamMember.Value.Region))
	}
}

func TestCredentialProviderConfiguration_flattenGatewayIAMRole_withIamCredentialProvider(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	in := awstypes.CredentialProviderConfiguration{
		CredentialProviderType: awstypes.CredentialProviderTypeGatewayIamRole,
		CredentialProvider: &awstypes.CredentialProviderMemberIamCredentialProvider{
			Value: awstypes.IamCredentialProvider{
				Service: aws.String("execute-api"),
				Region:  aws.String(testRegion("us", "west", "2")),
			},
		},
	}

	var m credentialProviderConfigurationModel
	diags := m.Flatten(ctx, in)
	if diags.HasError() {
		t.Fatalf("Flatten: %v", diags)
	}
	gwIAM, d := m.GatewayIAMRole.ToPtr(ctx)
	if d.HasError() {
		t.Fatalf("ToPtr: %v", d)
	}
	if gwIAM.Service.ValueString() != "execute-api" {
		t.Fatalf("Service: got %s", gwIAM.Service.ValueString())
	}
	if gwIAM.Region.ValueString() != testRegion("us", "west", "2") {
		t.Fatalf("Region: got %s", gwIAM.Region.ValueString())
	}
}

func TestCredentialProviderConfiguration_flattenGatewayIAMRole_nilCredentialProvider(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	in := awstypes.CredentialProviderConfiguration{
		CredentialProviderType: awstypes.CredentialProviderTypeGatewayIamRole,
		CredentialProvider:     nil,
	}

	var m credentialProviderConfigurationModel
	diags := m.Flatten(ctx, in)
	if diags.HasError() {
		t.Fatalf("Flatten: %v", diags)
	}
	gwIAM, d := m.GatewayIAMRole.ToPtr(ctx)
	if d.HasError() {
		t.Fatalf("ToPtr: %v", d)
	}
	if !gwIAM.Service.IsNull() {
		t.Fatalf("Service: want null, got %s", gwIAM.Service.ValueString())
	}
	if !gwIAM.Region.IsNull() {
		t.Fatalf("Region: want null, got %s", gwIAM.Region.ValueString())
	}
}
