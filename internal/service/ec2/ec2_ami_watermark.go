// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ami_watermark", name="AMI Watermark")
func newAMIWatermarkResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &amiWatermarkResource{}, nil
}

type amiWatermarkResource struct {
	framework.ResourceWithModel[amiWatermarkResourceModel]
	framework.WithNoUpdate
}

func (r *amiWatermarkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"image_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"watermark_key": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"watermark_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 128),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z0-9()\[\] ./\-'@_]+$`),
						"must contain only alphanumeric characters, parentheses (()), square brackets ([]), spaces, periods (.), slashes (/), dashes (-), single quotes ('), at-signs (@), or underscores (_)",
					),
				},
			},
		},
	}
}

func (r *amiWatermarkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data amiWatermarkResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	imageID := data.ImageID.ValueString()
	watermarkName := data.WatermarkName.ValueString()
	input := ec2.AttachImageWatermarkInput{
		ImageId:       aws.String(imageID),
		WatermarkName: aws.String(watermarkName),
	}

	output, err := conn.AttachImageWatermark(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, imageID)
		return
	}

	watermarkKey := aws.ToString(output.WatermarkKey)
	data.ID = fwflex.StringValueToFramework(ctx, watermarkKey)
	data.WatermarkKey = types.StringValue(watermarkKey)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, data))
}

func (r *amiWatermarkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data amiWatermarkResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	_, err := findImageWatermark(ctx, conn, data.ImageID.ValueString(), data.WatermarkKey.ValueString())

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func (r *amiWatermarkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data amiWatermarkResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	imageID := data.ImageID.ValueString()
	watermarkKey := data.WatermarkKey.ValueString()
	input := ec2.DetachImageWatermarkInput{
		ImageId:      aws.String(imageID),
		WatermarkKey: aws.String(watermarkKey),
	}

	_, err := conn.DetachImageWatermark(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAMIIDNotFound, errCodeInvalidAMIIDUnavailable) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, imageID)
		return
	}
}

func (r *amiWatermarkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	const amiWatermarkIDParts = 2
	parts, err := intflex.ExpandResourceId(req.ID, amiWatermarkIDParts, false)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("image_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("watermark_key"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), parts[1])...)
}

type amiWatermarkResourceModel struct {
	framework.WithRegionModel
	ID            types.String `tfsdk:"id"`
	ImageID       types.String `tfsdk:"image_id"`
	WatermarkKey  types.String `tfsdk:"watermark_key"`
	WatermarkName types.String `tfsdk:"watermark_name"`
}
