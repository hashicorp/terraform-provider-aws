// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package apigateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*itypes.ServicePackageFrameworkDataSource {
	return []*itypes.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*itypes.ServicePackageFrameworkResource {
	return []*itypes.ServicePackageFrameworkResource{
		{
			Factory:  newResourceAccount,
			TypeName: "aws_api_gateway_account",
			Name:     "Account",
		},
		{
			Factory:  newDomainNameAccessAssociationResource,
			TypeName: "aws_api_gateway_domain_name_access_association",
			Name:     "Domain Name Access Association",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*itypes.ServicePackageSDKDataSource {
	return []*itypes.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceAPIKey,
			TypeName: "aws_api_gateway_api_key",
			Name:     "API Key",
			Tags:     &itypes.ServicePackageResourceTags{},
		},
		{
			Factory:  dataSourceAuthorizer,
			TypeName: "aws_api_gateway_authorizer",
			Name:     "Authorizer",
		},
		{
			Factory:  dataSourceAuthorizers,
			TypeName: "aws_api_gateway_authorizers",
			Name:     "Authorizers",
		},
		{
			Factory:  dataSourceDomainName,
			TypeName: "aws_api_gateway_domain_name",
			Name:     "Domain Name",
			Tags:     &itypes.ServicePackageResourceTags{},
		},
		{
			Factory:  dataSourceExport,
			TypeName: "aws_api_gateway_export",
			Name:     "Export",
		},
		{
			Factory:  dataSourceResource,
			TypeName: "aws_api_gateway_resource",
			Name:     "Resource",
		},
		{
			Factory:  dataSourceRestAPI,
			TypeName: "aws_api_gateway_rest_api",
			Name:     "REST API",
			Tags:     &itypes.ServicePackageResourceTags{},
		},
		{
			Factory:  dataSourceSDK,
			TypeName: "aws_api_gateway_sdk",
			Name:     "SDK",
		},
		{
			Factory:  dataSourceVPCLink,
			TypeName: "aws_api_gateway_vpc_link",
			Name:     "VPC Link",
			Tags:     &itypes.ServicePackageResourceTags{},
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*itypes.ServicePackageSDKResource {
	return []*itypes.ServicePackageSDKResource{
		{
			Factory:  resourceAPIKey,
			TypeName: "aws_api_gateway_api_key",
			Name:     "API Key",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceAuthorizer,
			TypeName: "aws_api_gateway_authorizer",
			Name:     "Authorizer",
		},
		{
			Factory:  resourceBasePathMapping,
			TypeName: "aws_api_gateway_base_path_mapping",
			Name:     "Base Path Mapping",
		},
		{
			Factory:  resourceClientCertificate,
			TypeName: "aws_api_gateway_client_certificate",
			Name:     "Client Certificate",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceDeployment,
			TypeName: "aws_api_gateway_deployment",
			Name:     "Deployment",
		},
		{
			Factory:  resourceDocumentationPart,
			TypeName: "aws_api_gateway_documentation_part",
			Name:     "Documentation Part",
		},
		{
			Factory:  resourceDocumentationVersion,
			TypeName: "aws_api_gateway_documentation_version",
			Name:     "Documentation Version",
		},
		{
			Factory:  resourceDomainName,
			TypeName: "aws_api_gateway_domain_name",
			Name:     "Domain Name",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceGatewayResponse,
			TypeName: "aws_api_gateway_gateway_response",
			Name:     "Gateway Response",
		},
		{
			Factory:  resourceIntegration,
			TypeName: "aws_api_gateway_integration",
			Name:     "Integration",
		},
		{
			Factory:  resourceIntegrationResponse,
			TypeName: "aws_api_gateway_integration_response",
			Name:     "Integration Response",
		},
		{
			Factory:  resourceMethod,
			TypeName: "aws_api_gateway_method",
			Name:     "Method",
		},
		{
			Factory:  resourceMethodResponse,
			TypeName: "aws_api_gateway_method_response",
			Name:     "Method Response",
		},
		{
			Factory:  resourceMethodSettings,
			TypeName: "aws_api_gateway_method_settings",
			Name:     "Method Settings",
		},
		{
			Factory:  resourceModel,
			TypeName: "aws_api_gateway_model",
			Name:     "Model",
		},
		{
			Factory:  resourceRequestValidator,
			TypeName: "aws_api_gateway_request_validator",
			Name:     "Request Validator",
		},
		{
			Factory:  resourceResource,
			TypeName: "aws_api_gateway_resource",
			Name:     "Resource",
		},
		{
			Factory:  resourceRestAPI,
			TypeName: "aws_api_gateway_rest_api",
			Name:     "REST API",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceRestAPIPolicy,
			TypeName: "aws_api_gateway_rest_api_policy",
			Name:     "REST API Policy",
		},
		{
			Factory:  resourceStage,
			TypeName: "aws_api_gateway_stage",
			Name:     "Stage",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceUsagePlan,
			TypeName: "aws_api_gateway_usage_plan",
			Name:     "Usage Plan",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceUsagePlanKey,
			TypeName: "aws_api_gateway_usage_plan_key",
			Name:     "Usage Plan Key",
		},
		{
			Factory:  resourceVPCLink,
			TypeName: "aws_api_gateway_vpc_link",
			Name:     "VPC Link",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.APIGateway
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*apigateway.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*apigateway.Options){
		apigateway.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *apigateway.Options) {
			if region := config["region"].(string); o.Region != region {
				tflog.Info(ctx, "overriding provider-configured AWS API region", map[string]any{
					"service":         "apigateway",
					"original_region": o.Region,
					"override_region": region,
				})
				o.Region = region
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return apigateway.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*apigateway.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*apigateway.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *apigateway.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*apigateway.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
