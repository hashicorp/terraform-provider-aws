// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package ses

import (
	"context"
	"unique"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
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
	return []*inttypes.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*inttypes.ServicePackageSDKDataSource {
	return []*inttypes.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceActiveReceiptRuleSet,
			TypeName: "aws_ses_active_receipt_rule_set",
			Name:     "Active Receipt Rule Set",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  dataSourceDomainIdentity,
			TypeName: "aws_ses_domain_identity",
			Name:     "Domain Identity",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  dataSourceEmailIdentity,
			TypeName: "aws_ses_email_identity",
			Name:     "Email Identity",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{
		{
			Factory:  resourceActiveReceiptRuleSet,
			TypeName: "aws_ses_active_receipt_rule_set",
			Name:     "Active Receipt Rule Set",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceConfigurationSet,
			TypeName: "aws_ses_configuration_set",
			Name:     "Configuration Set",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceDomainDKIM,
			TypeName: "aws_ses_domain_dkim",
			Name:     "Domain DKIM",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceDomainIdentity,
			TypeName: "aws_ses_domain_identity",
			Name:     "Domain Identity",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceDomainIdentityVerification,
			TypeName: "aws_ses_domain_identity_verification",
			Name:     "Domain Identity Verification",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceDomainMailFrom,
			TypeName: "aws_ses_domain_mail_from",
			Name:     "MAIL FROM Domain",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceEmailIdentity,
			TypeName: "aws_ses_email_identity",
			Name:     "Email Identity",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceEventDestination,
			TypeName: "aws_ses_event_destination",
			Name:     "Configuration Set Event Destination",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceIdentityNotificationTopic,
			TypeName: "aws_ses_identity_notification_topic",
			Name:     "Identity Notification Topic",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceIdentityPolicy,
			TypeName: "aws_ses_identity_policy",
			Name:     "Identity Policy",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceReceiptFilter,
			TypeName: "aws_ses_receipt_filter",
			Name:     "Receipt Filter",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceReceiptRule,
			TypeName: "aws_ses_receipt_rule",
			Name:     "Receipt Rule",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceReceiptRuleSet,
			TypeName: "aws_ses_receipt_rule_set",
			Name:     "Receipt Rule Set",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
		{
			Factory:  resourceTemplate,
			TypeName: "aws_ses_template",
			Name:     "Template",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsOverrideEnabled:             true,
				IsValidateOverrideInPartition: true,
			}),
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.SES
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*ses.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*ses.Options){
		ses.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *ses.Options) {
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

	return ses.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*ses.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*ses.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *ses.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*ses.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
