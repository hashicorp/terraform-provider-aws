// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package bedrockagent

import (
	"context"
	"unique"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*inttypes.ServicePackageFrameworkDataSource {
	return []*inttypes.ServicePackageFrameworkDataSource{
		{
			Factory:  newDataSourceAgentVersions,
			TypeName: "aws_bedrockagent_agent_versions",
			Name:     "Agent Versions",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			}),
		},
	}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	return []*inttypes.ServicePackageFrameworkResource{
		{
			Factory:  newAgentResource,
			TypeName: "aws_bedrockagent_agent",
			Name:     "Agent",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: "agent_arn",
			}),
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			}),
		},
		{
			Factory:  newAgentActionGroupResource,
			TypeName: "aws_bedrockagent_agent_action_group",
			Name:     "Agent Action Group",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			}),
		},
		{
			Factory:  newAgentAliasResource,
			TypeName: "aws_bedrockagent_agent_alias",
			Name:     "Agent Alias",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: "agent_alias_arn",
			}),
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			}),
		},
		{
			Factory:  newAgentCollaboratorResource,
			TypeName: "aws_bedrockagent_agent_collaborator",
			Name:     "Agent Collaborator",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			}),
		},
		{
			Factory:  newAgentKnowledgeBaseAssociationResource,
			TypeName: "aws_bedrockagent_agent_knowledge_base_association",
			Name:     "Agent Knowledge Base Association",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			}),
		},
		{
			Factory:  newDataSourceResource,
			TypeName: "aws_bedrockagent_data_source",
			Name:     "Data Source",
			Region: unique.Make(inttypes.ServicePackageResourceRegion{
				IsGlobal:          false,
				IsOverrideEnabled: false,
			}),
		},
		{
			Factory:  newKnowledgeBaseResource,
			TypeName: "aws_bedrockagent_knowledge_base",
			Name:     "Knowledge Base",
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
	return []*inttypes.ServicePackageSDKResource{}
}

func (p *servicePackage) ServicePackageName() string {
	return names.BedrockAgent
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*bedrockagent.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*bedrockagent.Options){
		bedrockagent.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *bedrockagent.Options) {
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

	return bedrockagent.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*bedrockagent.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*bedrockagent.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *bedrockagent.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*bedrockagent.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
