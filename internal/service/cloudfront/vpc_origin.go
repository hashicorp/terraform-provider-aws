package cloudfront

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudfront_vpc_origin", name="VPC Origin")
func newCloudfrontVPCOriginResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &cloudfrontVPCOriginResource{}
	return r, nil
}

type cloudfrontVPCOriginResource struct {
	framework.ResourceWithConfigure
}

func (r *cloudfrontVPCOriginResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_cloudfront_vpc_origin"
}

func (r *cloudfrontVPCOriginResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrLastModifiedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},

			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrVPCOriginEndpointConfig: schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[vpcOriginEndpointConfigModel](ctx),
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Attributes: map[string]schema.Attribute{
					"origin_arn": schema.StringAttribute{
						Required:   true,
						CustomType: fwtypes.ARNType,
					},
					"http_port": schema.Int32Attribute{
						Required: true,
						Validators: []validator.Int32{
							int32validator.Between(1, 65535),
						},
					},
					"https_port": schema.Int32Attribute{
						Required: true,
						Validators: []validator.Int32{
							int32validator.Between(1, 65535),
						},
					},
					names.AttrName: schema.StringAttribute{
						Required: true,
					},
					names.AttrOriginProtocolPolicy: schema.StringAttribute{
						Required:   true,
						CustomType: fwtypes.StringEnumType[awstypes.OriginProtocolPolicy](),
					},
				},
				Blocks: map[string]schema.Block{
					names.AttrOriginSSLProtocols: schema.ListNestedBlock{
						CustomType: fwtypes.NewListNestedObjectTypeOf[originSSLProtocolsModel](ctx),
						Validators: []validator.List{
							listvalidator.IsRequired(),
							listvalidator.SizeAtLeast(1),
							listvalidator.SizeAtMost(1),
						},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"items": schema.SetAttribute{
									CustomType:  fwtypes.SetOfStringType,
									Optional:    true,
									ElementType: types.StringType,
								},
								"quantity": schema.Int64Attribute{
									Required: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *cloudfrontVPCOriginResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vpcOriginModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)
	var input cloudfront.CreateVpcOriginInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)

	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.Tags.Items = tags
	}

	output, err := conn.CreateVpcOrigin(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("Creating VPC Cloudfront Origin", err.Error())
		return
	}

	// Set values for unknowns.
	data.ARN = fwflex.StringToFramework(ctx, output.VpcOrigin.Arn)
	data.CreatedTime = fwflex.TimeToFramework(ctx, output.VpcOrigin.CreatedTime)
	data.Id = fwflex.StringToFramework(ctx, output.VpcOrigin.Id)
	data.LastModifiedTime = fwflex.TimeToFramework(ctx, output.VpcOrigin.LastModifiedTime)
	data.Status = fwflex.StringToFramework(ctx, output.VpcOrigin.Status)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
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

type vpcOriginModel struct {
	ARN                     types.String                                        `tfsdk:"arn"`
	CreatedTime             timetypes.RFC3339                                   `tfsdk:"created_time"`
	Id                      types.String                                        `tfsdk:"id"`
	LastModifiedTime        timetypes.RFC3339                                   `tfsdk:"last_modified_time"`
	Status                  types.String                                        `tfsdk:"status"`
	VpcOriginEndpointConfig fwtypes.ObjectValueOf[vpcOriginEndpointConfigModel] `tfsdk:"vpc_origin_endpoint_config"`
	Tags                    tftags.Map                                          `tfsdk:"tags"`
	TagsAll                 tftags.Map                                          `tfsdk:"tags_all"`
}

type vpcOriginEndpointConfigModel struct {
	Arn                  types.String                                             `tfsdk:"origin_arn"`
	HTTPPort             types.Int32                                              `tfsdk:"http_port"`
	HTTPSPort            types.Int32                                              `tfsdk:"https_port"`
	Name                 types.String                                             `tfsdk:"name"`
	OriginProtocolPolicy fwtypes.StringEnum[awstypes.OriginProtocolPolicy]        `tfsdk:"origin_protocol_policy"`
	OriginSslProtocols   fwtypes.ListNestedObjectValueOf[originSSLProtocolsModel] `tfsdk:"origin_ssl_protocols"`
}

type originSSLProtocolsModel struct {
	Items    fwtypes.SetValueOf[types.String] `tfsdk:"items"`
	Quantity types.Int64                      `tfsdk:"quantity"`
}
