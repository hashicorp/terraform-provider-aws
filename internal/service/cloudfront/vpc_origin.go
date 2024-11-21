package cloudfront

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
	"time"
)

// @FrameworkResource("aws_cloudfront_vpc_origin", name="VPC Origin")
func newCloudfrontVPCOriginResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &cloudfrontVPCOriginResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type cloudfrontVPCOriginResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *cloudfrontVPCOriginResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_cloudfront_vpc_origin"
}

func (r *cloudfrontVPCOriginResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrCreatedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrVPCOriginID: schema.StringAttribute{
				Computed: true,
			},
			names.AttrLastModifiedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				// TODO: Add validation
			},
			// TODO: VPCOriginEndpointConfig
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *cloudfrontVPCOriginResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *cloudfrontVPCOriginResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}
