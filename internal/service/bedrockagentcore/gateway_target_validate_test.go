// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// testAccAgentCoreRuntimeInvocationsMCPURL returns the regional bedrock-agentcore invocations URL shape used for
// Agent Core Runtime MCP gateway targets (URL-encoded runtime ARN in the path). Matches AWS docs / PR guidance;
// not used against real AWS in these unit tests.
func testAccAgentCoreRuntimeInvocationsMCPURL() string {
	const accountID = "123456789012"

	r := testRegion("us", "east", "1")
	rawARN := "arn:aws:bedrock-agentcore:" + r + ":" + accountID + ":runtime/example-aBcDeF1234"
	encoded := strings.ReplaceAll(strings.ReplaceAll(rawARN, ":", "%3A"), "/", "%2F")

	return "https://bedrock-agentcore." + r + ".amazonaws.com/runtimes/" + encoded + "/invocations?qualifier=DEFAULT"
}

func TestValidateGatewayIAMRoleCredentialConfiguration_regionWithoutService(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	data := &gatewayTargetResourceModel{
		CredentialProviderConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &credentialProviderConfigurationModel{
			ApiKey: fwtypes.NewListNestedObjectValueOfNull[apiKeyCredentialProviderModel](ctx),
			OAuth:  fwtypes.NewListNestedObjectValueOfNull[oauthCredentialProviderModel](ctx),
			GatewayIAMRole: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &gatewayIAMRoleProviderModel{
				Service: types.StringNull(),
				Region:  types.StringValue(testRegion("us", "east", "1")),
			}),
		}),
	}

	diags := validateGatewayIAMRoleCredentialConfiguration(ctx, data)
	if !diags.HasError() {
		t.Fatal("expected error when region is set without service")
	}
}

func TestValidateGatewayIAMRoleCredentialConfiguration_mcpServerRequiresService(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	data := &gatewayTargetResourceModel{
		CredentialProviderConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &credentialProviderConfigurationModel{
			ApiKey: fwtypes.NewListNestedObjectValueOfNull[apiKeyCredentialProviderModel](ctx),
			OAuth:  fwtypes.NewListNestedObjectValueOfNull[oauthCredentialProviderModel](ctx),
			GatewayIAMRole: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &gatewayIAMRoleProviderModel{
				Service: types.StringNull(),
				Region:  types.StringNull(),
			}),
		}),
		TargetConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &targetConfigurationModel{
			MCP: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &mcpConfigurationModel{
				ApiGateway:    fwtypes.NewListNestedObjectValueOfNull[mcpApiGatewayConfigurationModel](ctx),
				Lambda:        fwtypes.NewListNestedObjectValueOfNull[mcpLambdaConfigurationModel](ctx),
				SmithyModel:   fwtypes.NewListNestedObjectValueOfNull[apiSchemaConfigurationModel](ctx),
				OpenApiSchema: fwtypes.NewListNestedObjectValueOfNull[apiSchemaConfigurationModel](ctx),
				MCPServer: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &mcpServerConfigurationModel{
					Endpoint: types.StringValue(testAccAgentCoreRuntimeInvocationsMCPURL()),
				}),
			}),
		}),
	}

	diags := validateGatewayIAMRoleCredentialConfiguration(ctx, data)
	if !diags.HasError() {
		t.Fatal("expected error for MCP server target with gateway_iam_role and no service")
	}
	msg := diags[0].Detail()
	if !strings.Contains(msg, "MCP server") {
		t.Fatalf("unexpected diagnostic: %s", msg)
	}
}

func TestValidateGatewayIAMRoleCredentialConfiguration_mcpServerWithService(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	data := &gatewayTargetResourceModel{
		CredentialProviderConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &credentialProviderConfigurationModel{
			ApiKey: fwtypes.NewListNestedObjectValueOfNull[apiKeyCredentialProviderModel](ctx),
			OAuth:  fwtypes.NewListNestedObjectValueOfNull[oauthCredentialProviderModel](ctx),
			GatewayIAMRole: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &gatewayIAMRoleProviderModel{
				Service: types.StringValue("bedrock-agentcore"),
				Region:  types.StringValue(testRegion("us", "east", "1")),
			}),
		}),
		TargetConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &targetConfigurationModel{
			MCP: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &mcpConfigurationModel{
				ApiGateway:    fwtypes.NewListNestedObjectValueOfNull[mcpApiGatewayConfigurationModel](ctx),
				Lambda:        fwtypes.NewListNestedObjectValueOfNull[mcpLambdaConfigurationModel](ctx),
				SmithyModel:   fwtypes.NewListNestedObjectValueOfNull[apiSchemaConfigurationModel](ctx),
				OpenApiSchema: fwtypes.NewListNestedObjectValueOfNull[apiSchemaConfigurationModel](ctx),
				MCPServer: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &mcpServerConfigurationModel{
					Endpoint: types.StringValue(testAccAgentCoreRuntimeInvocationsMCPURL()),
				}),
			}),
		}),
	}

	diags := validateGatewayIAMRoleCredentialConfiguration(ctx, data)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
}

func TestValidateGatewayIAMRoleCredentialConfiguration_nonMcpEmptyGatewayIAM(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	data := &gatewayTargetResourceModel{
		CredentialProviderConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &credentialProviderConfigurationModel{
			ApiKey: fwtypes.NewListNestedObjectValueOfNull[apiKeyCredentialProviderModel](ctx),
			OAuth:  fwtypes.NewListNestedObjectValueOfNull[oauthCredentialProviderModel](ctx),
			GatewayIAMRole: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &gatewayIAMRoleProviderModel{
				Service: types.StringNull(),
				Region:  types.StringNull(),
			}),
		}),
		TargetConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &targetConfigurationModel{
			MCP: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &mcpConfigurationModel{
				ApiGateway:    fwtypes.NewListNestedObjectValueOfNull[mcpApiGatewayConfigurationModel](ctx),
				Lambda:        fwtypes.NewListNestedObjectValueOfNull[mcpLambdaConfigurationModel](ctx),
				SmithyModel:   fwtypes.NewListNestedObjectValueOfNull[apiSchemaConfigurationModel](ctx),
				OpenApiSchema: fwtypes.NewListNestedObjectValueOfNull[apiSchemaConfigurationModel](ctx),
				MCPServer:     fwtypes.NewListNestedObjectValueOfNull[mcpServerConfigurationModel](ctx),
			}),
		}),
	}

	diags := validateGatewayIAMRoleCredentialConfiguration(ctx, data)
	if diags.HasError() {
		t.Fatalf("unexpected error for non-MCP target with gateway_iam_role {}: %v", diags)
	}
}
