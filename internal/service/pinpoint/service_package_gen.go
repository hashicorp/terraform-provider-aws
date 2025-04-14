// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package pinpoint

import (
	"context"
	"unique"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpoint"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*inttypes.ServicePackageFrameworkDataSource {
	return []*inttypes.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	return []*inttypes.ServicePackageFrameworkResource{
		{
			Factory:  newResourceEmailTemplate,
			TypeName: "aws_pinpoint_email_template",
			Name:     "Email Template",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			}),
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*inttypes.ServicePackageSDKDataSource {
	return []*inttypes.ServicePackageSDKDataSource{}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{
		{
			Factory:  resourceADMChannel,
			TypeName: "aws_pinpoint_adm_channel",
			Name:     "ADM Channel",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceAPNSChannel,
			TypeName: "aws_pinpoint_apns_channel",
			Name:     "APNS Channel",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceAPNSSandboxChannel,
			TypeName: "aws_pinpoint_apns_sandbox_channel",
			Name:     "APNS Sandbox Channel",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceAPNSVoIPChannel,
			TypeName: "aws_pinpoint_apns_voip_channel",
			Name:     "APNS VoIP Channel",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceAPNSVoIPSandboxChannel,
			TypeName: "aws_pinpoint_apns_voip_sandbox_channel",
			Name:     "APNS VoIP Sandbox Channel",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceApp,
			TypeName: "aws_pinpoint_app",
			Name:     "App",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			}),
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceBaiduChannel,
			TypeName: "aws_pinpoint_baidu_channel",
			Name:     "Baidu Channel",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceEmailChannel,
			TypeName: "aws_pinpoint_email_channel",
			Name:     "Email Channel",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceEventStream,
			TypeName: "aws_pinpoint_event_stream",
			Name:     "Event Stream",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceGCMChannel,
			TypeName: "aws_pinpoint_gcm_channel",
			Name:     "GCM Channel",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceSMSChannel,
			TypeName: "aws_pinpoint_sms_channel",
			Name:     "SMS Channel",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:                      false,
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.Pinpoint
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*pinpoint.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*pinpoint.Options){
		pinpoint.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *pinpoint.Options) {
			if region := config[names.AttrRegion].(string); o.Region != region {
				tflog.Info(ctx, "overriding provider-configured AWS API region", map[string]any{
					"service":         p.ServicePackageName(),
					"original_region": o.Region,
					"override_region": region,
				})
				o.Region = region
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return pinpoint.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*pinpoint.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*pinpoint.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *pinpoint.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*pinpoint.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
