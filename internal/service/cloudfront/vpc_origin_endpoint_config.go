package cloudfront

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
	"time"
)

// @FrameworkResource("aws_cloudfront_vpc_origin_endpoint_config", name="VPC Origin Endpoint Config")
func newCloudfrontVPCOriginEndpointConfigResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &cloudfrontVPCOriginEndpointConfigResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type cloudfrontVPCOriginEndpointConfigResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *cloudfrontVPCOriginEndpointConfigResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrHTTPPort: schema.NumberAttribute{
				Required: true,
			},
			names.AttrHTTPSPort: schema.NumberAttribute{
				Required: true,
			},
			// TODO: OriginProtocolPolicy
			// TODO: OriginSslProtocols
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *cloudfrontVPCOriginEndpointConfigResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginEndpointConfigResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginEndpointConfigResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginEndpointConfigResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginEndpointConfigResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_cloudfront_vpc_origin_endpoint_config"
}
