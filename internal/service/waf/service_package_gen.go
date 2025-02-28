// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package waf

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
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
	return []*itypes.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*itypes.ServicePackageSDKDataSource {
	return []*itypes.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceIPSet,
			TypeName: "aws_waf_ipset",
			Name:     "IPSet",
		},
		{
			Factory:  dataSourceRateBasedRule,
			TypeName: "aws_waf_rate_based_rule",
			Name:     "Rate Based Rule",
		},
		{
			Factory:  dataSourceRule,
			TypeName: "aws_waf_rule",
			Name:     "Rule",
		},
		{
			Factory:  dataSourceSubscribedRuleGroup,
			TypeName: "aws_waf_subscribed_rule_group",
			Name:     "Subscribed Rule Group",
		},
		{
			Factory:  dataSourceWebACL,
			TypeName: "aws_waf_web_acl",
			Name:     "Web ACL",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*itypes.ServicePackageSDKResource {
	return []*itypes.ServicePackageSDKResource{
		{
			Factory:  resourceByteMatchSet,
			TypeName: "aws_waf_byte_match_set",
			Name:     "ByteMatchSet",
		},
		{
			Factory:  resourceGeoMatchSet,
			TypeName: "aws_waf_geo_match_set",
			Name:     "GeoMatchSet",
		},
		{
			Factory:  resourceIPSet,
			TypeName: "aws_waf_ipset",
			Name:     "IPSet",
		},
		{
			Factory:  resourceRateBasedRule,
			TypeName: "aws_waf_rate_based_rule",
			Name:     "Rate Based Rule",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceRegexMatchSet,
			TypeName: "aws_waf_regex_match_set",
			Name:     "Regex Match Set",
		},
		{
			Factory:  resourceRegexPatternSet,
			TypeName: "aws_waf_regex_pattern_set",
			Name:     "Regex Pattern Set",
		},
		{
			Factory:  resourceRule,
			TypeName: "aws_waf_rule",
			Name:     "Rule",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceRuleGroup,
			TypeName: "aws_waf_rule_group",
			Name:     "Rule Group",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceSizeConstraintSet,
			TypeName: "aws_waf_size_constraint_set",
			Name:     "Size Constraint Set",
		},
		{
			Factory:  resourceSQLInjectionMatchSet,
			TypeName: "aws_waf_sql_injection_match_set",
			Name:     "SqlInjectionMatchSet",
		},
		{
			Factory:  resourceWebACL,
			TypeName: "aws_waf_web_acl",
			Name:     "Web ACL",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  resourceXSSMatchSet,
			TypeName: "aws_waf_xss_match_set",
			Name:     "XSS Match Set",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.WAF
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*waf.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*waf.Options){
		waf.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *waf.Options) {
			if region := config["region"].(string); o.Region != region {
				tflog.Info(ctx, "overriding provider-configured AWS API region", map[string]any{
					"service":         "waf",
					"original_region": o.Region,
					"override_region": region,
				})
				o.Region = region
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return waf.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*waf.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*waf.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *waf.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*waf.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
