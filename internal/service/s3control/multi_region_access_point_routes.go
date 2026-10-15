// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3control_multi_region_access_point_routes", name="Multi-Region Access Point Routes")
// @IdentityAttribute("mrap")
// @Testing(hasNoPreExistingResource=true)
// @Testing(checkDestroyNoop=true)
// @Testing(altRegionTfVars=true)
// @Testing(identityRegionOverrideTest=false)
// @Testing(importStateIdAttribute="mrap")
func newMultiRegionAccessPointRoutesResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &multiRegionAccessPointRoutesResource{}, nil
}

type multiRegionAccessPointRoutesResource struct {
	framework.ResourceWithModel[multiRegionAccessPointRoutesResourceModel]
	framework.WithImportByIdentity
	framework.WithNoOpDelete
}

func (r *multiRegionAccessPointRoutesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAccountID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
			},
			"mrap": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"route": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[multiRegionAccessPointRouteModel](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrBucket: schema.StringAttribute{
							Required: true,
						},
						names.AttrRegion: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								fwvalidators.AWSRegion(),
							},
						},
						"traffic_dial_percentage": schema.Int64Attribute{
							Required: true,
							Validators: []validator.Int64{
								int64validator.Between(0, 100),
							},
						},
					},
				},
			},
		},
	}
}

func (r *multiRegionAccessPointRoutesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data multiRegionAccessPointRoutesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.AccountID.IsUnknown() {
		data.AccountID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	conn := r.Meta().S3ControlClient(ctx)

	var input s3control.SubmitMultiRegionAccessPointRoutesInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, &input, fwflex.WithFieldNameSuffix("Updates"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.SubmitMultiRegionAccessPointRoutes(ctx, &input)

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating S3 Multi-Region Access Point Routes (%s)", data.Mrap.ValueString()), err.Error())
		return
	}

	output, err := findMultiRegionAccessPointRoutesByTwoPartKey(ctx, conn, data.AccountID.ValueString(), data.Mrap.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading S3 Multi-Region Access Point Routes (%s)", data.Mrap.ValueString()), err.Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *multiRegionAccessPointRoutesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data multiRegionAccessPointRoutesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.AccountID.IsNull() || data.AccountID.ValueString() == "" {
		data.AccountID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	conn := r.Meta().S3ControlClient(ctx)

	output, err := findMultiRegionAccessPointRoutesByTwoPartKey(ctx, conn, data.AccountID.ValueString(), data.Mrap.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading S3 Multi-Region Access Point Routes (%s)", data.Mrap.ValueString()), err.Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *multiRegionAccessPointRoutesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data multiRegionAccessPointRoutesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	var input s3control.SubmitMultiRegionAccessPointRoutesInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, &input, fwflex.WithFieldNameSuffix("Updates"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.SubmitMultiRegionAccessPointRoutes(ctx, &input)

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("updating S3 Multi-Region Access Point Routes (%s)", data.Mrap.ValueString()), err.Error())
		return
	}

	output, err := findMultiRegionAccessPointRoutesByTwoPartKey(ctx, conn, data.AccountID.ValueString(), data.Mrap.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading S3 Multi-Region Access Point Routes (%s)", data.Mrap.ValueString()), err.Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findMultiRegionAccessPointRoutesByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, mrap string) (*s3control.GetMultiRegionAccessPointRoutesOutput, error) {
	input := s3control.GetMultiRegionAccessPointRoutesInput{
		AccountId: aws.String(accountID),
		Mrap:      aws.String(mrap),
	}

	output, err := conn.GetMultiRegionAccessPointRoutes(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchMultiRegionAccessPoint) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Routes) == 0 {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type multiRegionAccessPointRoutesResourceModel struct {
	framework.WithRegionModel
	AccountID types.String                                                     `tfsdk:"account_id"`
	Mrap      types.String                                                     `tfsdk:"mrap"`
	Route     fwtypes.SetNestedObjectValueOf[multiRegionAccessPointRouteModel] `tfsdk:"route"`
}

type multiRegionAccessPointRouteModel struct {
	Bucket                types.String `tfsdk:"bucket"`
	Region                types.String `tfsdk:"region"`
	TrafficDialPercentage types.Int64  `tfsdk:"traffic_dial_percentage"`
}
