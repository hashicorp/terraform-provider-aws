// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package verifiedpermissions

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*itypes.ServicePackageFrameworkDataSource {
	return []*itypes.ServicePackageFrameworkDataSource{
		{
			Factory:  newDataSourcePolicyStore,
			TypeName: "aws_verifiedpermissions_policy_store",
			Name:     "Policy Store",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			},
		},
	}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*itypes.ServicePackageFrameworkResource {
	return []*itypes.ServicePackageFrameworkResource{
		{
			Factory:  newResourceIdentitySource,
			TypeName: "aws_verifiedpermissions_identity_source",
			Name:     "Identity Source",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			},
		},
		{
			Factory:  newResourcePolicy,
			TypeName: "aws_verifiedpermissions_policy",
			Name:     "Policy",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			},
		},
		{
			Factory:  newResourcePolicyStore,
			TypeName: "aws_verifiedpermissions_policy_store",
			Name:     "Policy Store",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			},
		},
		{
			Factory:  newResourcePolicyTemplate,
			TypeName: "aws_verifiedpermissions_policy_template",
			Name:     "Policy Template",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			},
		},
		{
			Factory:  newResourceSchema,
			TypeName: "aws_verifiedpermissions_schema",
			Name:     "Schema",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			},
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*itypes.ServicePackageSDKDataSource {
	return []*itypes.ServicePackageSDKDataSource{}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*itypes.ServicePackageSDKResource {
	return []*itypes.ServicePackageSDKResource{}
}

func (p *servicePackage) ServicePackageName() string {
	return names.VerifiedPermissions
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*verifiedpermissions.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*verifiedpermissions.Options){
		verifiedpermissions.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *verifiedpermissions.Options) {
			if region := config["region"].(string); o.Region != region {
				tflog.Info(ctx, "overriding provider-configured AWS API region", map[string]any{
					"service":         "verifiedpermissions",
					"original_region": o.Region,
					"override_region": region,
				})
				o.Region = region
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return verifiedpermissions.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*verifiedpermissions.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*verifiedpermissions.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *verifiedpermissions.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*verifiedpermissions.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
