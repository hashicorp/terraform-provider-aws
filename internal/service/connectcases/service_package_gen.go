// Code generated by internal/generate/servicepackages/main.go; DO NOT EDIT.

package connectcases

import (
	"context"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	connectcases_sdkv2 "github.com/aws/aws-sdk-go-v2/service/connectcases"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  ResourceContactCase,
			TypeName: "aws_connectcases_contact_case",
			Name:     "Connect Cases Contact Case",
		},
		{
			Factory:  ResourceDomain,
			TypeName: "aws_connectcases_domain",
			Name:     "Connect Cases Domain",
		},
		{
			Factory:  ResourceField,
			TypeName: "aws_connectcases_field",
			Name:     "Connect Cases Field",
		},
		{
			Factory:  ResourceLayout,
			TypeName: "aws_connectcases_layout",
			Name:     "Connect Cases Layout",
		},
		{
			Factory:  ResourceRelatedItem,
			TypeName: "aws_connectcases_related_item",
			Name:     "Connect Cases Related Item",
		},
		{
			Factory:  ResourceTemplate,
			TypeName: "aws_connectcases_template",
			Name:     "Connect Cases Template",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.ConnectCases
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*connectcases_sdkv2.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws_sdkv2.Config))

	return connectcases_sdkv2.NewFromConfig(cfg, func(o *connectcases_sdkv2.Options) {
		if endpoint := config["endpoint"].(string); endpoint != "" {
			o.BaseEndpoint = aws_sdkv2.String(endpoint)
		}
	}), nil
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
