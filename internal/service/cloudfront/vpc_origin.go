package cloudfront

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrLastModifiedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},

			names.AttrTags: tftags.TagsAttribute(),
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
						// TODO: User should be able to just specify an array, not object internals.
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"items": schema.SetAttribute{
									CustomType:  fwtypes.SetOfStringType,
									Required:    true,
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

	data.ARN = fwflex.StringToFramework(ctx, output.VpcOrigin.Arn)
	data.CreatedTime = fwflex.TimeToFramework(ctx, output.VpcOrigin.CreatedTime)
	data.Id = fwflex.StringToFramework(ctx, output.VpcOrigin.Id)
	data.LastModifiedTime = fwflex.TimeToFramework(ctx, output.VpcOrigin.LastModifiedTime)
	data.Status = fwflex.StringToFramework(ctx, output.VpcOrigin.Status)
	data.ETag = fwflex.StringToFramework(ctx, output.ETag)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *cloudfrontVPCOriginResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vpcOriginModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	input := &cloudfront.GetVpcOriginInput{
		Id: aws.String(data.Id.ValueString()),
	}

	output, err := conn.GetVpcOrigin(ctx, input)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront VPC Origin (%s)", data.Id.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.VpcOrigin, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ARN = fwflex.StringToFramework(ctx, output.VpcOrigin.Arn)
	data.CreatedTime = fwflex.TimeToFramework(ctx, output.VpcOrigin.CreatedTime)
	data.Id = fwflex.StringToFramework(ctx, output.VpcOrigin.Id)
	data.LastModifiedTime = fwflex.TimeToFramework(ctx, output.VpcOrigin.LastModifiedTime)
	data.Status = fwflex.StringToFramework(ctx, output.VpcOrigin.Status)
	data.ETag = fwflex.StringToFramework(ctx, output.ETag)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *cloudfrontVPCOriginResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old vpcOriginModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	var new vpcOriginModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	input := &cloudfront.UpdateVpcOriginInput{
		Id:      aws.String(new.Id.ValueString()),
		IfMatch: aws.String(old.ETag.ValueString()),
	}

	response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.UpdateVpcOrigin(ctx, input)

	// TODO: Handle "IllegalUpdate" error
	// TODO: Add Timeouts

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Updating Cloudfront VPC Origin (%s)", old.Id.ValueString()), err.Error())
		return
	}

	new.CreatedTime = fwflex.TimeToFramework(ctx, output.VpcOrigin.CreatedTime)
	new.LastModifiedTime = fwflex.TimeToFramework(ctx, output.VpcOrigin.LastModifiedTime)
	new.Status = fwflex.StringToFramework(ctx, output.VpcOrigin.Status)
	new.ETag = fwflex.StringToFramework(ctx, output.ETag)

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)

}

func (r *cloudfrontVPCOriginResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data vpcOriginModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	input := &cloudfront.DeleteVpcOriginInput{
		Id:      aws.String(data.Id.ValueString()),
		IfMatch: aws.String(data.ETag.ValueString()),
	}

	_, err := conn.DeleteVpcOrigin(ctx, input)

	// TODO: Handle "IllegalDelete" error
	// TODO: Add Timeouts

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudFront VPC Origin (%s)", data.Id.ValueString()), err.Error())
		return
	}
}

type vpcOriginModel struct {
	ARN                     types.String                                        `tfsdk:"arn"`
	CreatedTime             timetypes.RFC3339                                   `tfsdk:"created_time"`
	Id                      types.String                                        `tfsdk:"id"`
	ETag                    types.String                                        `tfsdk:"etag"`
	LastModifiedTime        timetypes.RFC3339                                   `tfsdk:"last_modified_time"`
	Status                  types.String                                        `tfsdk:"status"`
	VpcOriginEndpointConfig fwtypes.ObjectValueOf[vpcOriginEndpointConfigModel] `tfsdk:"vpc_origin_endpoint_config"`
	Tags                    tftags.Map                                          `tfsdk:"tags"`
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
