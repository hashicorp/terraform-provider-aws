// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_transfer_web_app", name="Web App")
// @Tags(identifierAttribute="arn")
func newResourceWebApp(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceWebApp{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameWebApp = "Web App"
)

type resourceWebApp struct {
	framework.ResourceWithModel[resourceWebAppModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceWebApp) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_endpoint": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(1024),
				},
			},
			names.AttrARN:     framework.ARNAttributeComputedOnly(),
			names.AttrID:      framework.IDAttribute(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"web_app_endpoint_policy": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.WebAppEndpointPolicy](),
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.WebAppEndpointPolicy](),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"web_app_units": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[webAppUnitsModel](ctx),
				Optional:   true,
				Computed:   true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"provisioned": types.Int64Type,
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"identity_provider_details": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[identityProviderDetailsModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"identity_center_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[identityCenterConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"application_arn": schema.StringAttribute{
										Computed: true,
									},
									"instance_arn": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(10, 1024),
											stringvalidator.RegexMatches(regexache.MustCompile(`^arn:[\w-]+:sso:::instance/(sso)?ins-[a-zA-Z0-9-.]{16}$`), ""),
										},
									},
									names.AttrRole: schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(20, 2048),
											stringvalidator.RegexMatches(regexache.MustCompile(`^arn:.*role/\S+$`), ""),
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: false,
				Delete: true,
			}),
		},
	}
}

func (r *resourceWebApp) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().TransferClient(ctx)

	var plan resourceWebAppModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input transfer.CreateWebAppInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateWebApp(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Transfer, create.ErrActionCreating, ResNameWebApp, "", err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Transfer, create.ErrActionCreating, ResNameWebApp, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.WebAppId = flex.StringToFramework(ctx, out.WebAppId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitWebAppCreated(ctx, conn, plan.WebAppId.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Transfer, create.ErrActionWaitingForCreation, ResNameWebApp, plan.WebAppId.String(), err),
			err.Error(),
		)
		return
	}

	rout, _ := findWebAppByID(ctx, conn, plan.WebAppId.ValueString())
	resp.Diagnostics.Append(flex.Flatten(ctx, rout, &plan, flex.WithFieldNamePrefix("Described"))...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceWebApp) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().TransferClient(ctx)

	var state resourceWebAppModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findWebAppByID(ctx, conn, state.WebAppId.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Transfer, create.ErrActionReading, ResNameWebApp, state.WebAppId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("Described"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceWebApp) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	needUpdate := false
	conn := r.Meta().TransferClient(ctx)

	var plan, state resourceWebAppModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := transfer.UpdateWebAppInput{
			WebAppId: state.WebAppId.ValueStringPointer(),
		}

		if !state.AccessEndpoint.Equal(plan.AccessEndpoint) {
			if v := plan.AccessEndpoint.ValueStringPointer(); v != nil && aws.ToString(v) != "" {
				input.AccessEndpoint = v
			}
			needUpdate = true
		}
		if !state.IdentityProviderDetails.Equal(plan.IdentityProviderDetails) {
			if v, diags := plan.IdentityProviderDetails.ToPtr(ctx); v != nil && !diags.HasError() {
				if v, diags := v.IdentityCenterConfig.ToPtr(ctx); v != nil && !diags.HasError() {
					input.IdentityProviderDetails = &awstypes.UpdateWebAppIdentityProviderDetailsMemberIdentityCenterConfig{
						Value: awstypes.UpdateWebAppIdentityCenterConfig{
							Role: v.Role.ValueStringPointer(),
						},
					}
					needUpdate = true
				}
			}
		}
		if !state.WebAppUnits.Equal(plan.WebAppUnits) {
			if v, diags := plan.WebAppUnits.ToPtr(ctx); v != nil && !diags.HasError() {
				if v, diags := plan.WebAppUnits.ToPtr(ctx); v != nil && !diags.HasError() {
					input.WebAppUnits = &awstypes.WebAppUnitsMemberProvisioned{
						Value: flex.Int32ValueFromFrameworkInt64(ctx, v.Provisioned),
					}
					needUpdate = true
				}
			}
		}

		if needUpdate {
			out, err := conn.UpdateWebApp(ctx, &input)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.Transfer, create.ErrActionUpdating, ResNameWebApp, plan.WebAppId.String(), err),
					err.Error(),
				)
				return
			}
			if out == nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.Transfer, create.ErrActionUpdating, ResNameWebApp, plan.WebAppId.String(), nil),
					errors.New("empty output").Error(),
				)
				return
			}
		}

		if !state.Tags.Equal(plan.Tags) {
			if err := updateTags(ctx, conn, plan.ARN.ValueString(), state.Tags, plan.Tags); err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.Transfer, create.ErrActionUpdating, ResNameWebApp, plan.WebAppId.String(), err),
					err.Error(),
				)
				return
			}
		}

		if resp.Diagnostics.HasError() {
			return
		}
	}

	rout, _ := findWebAppByID(ctx, conn, plan.WebAppId.ValueString())
	resp.Diagnostics.Append(flex.Flatten(ctx, rout, &plan, flex.WithFieldNamePrefix("Described"))...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceWebApp) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().TransferClient(ctx)

	var state resourceWebAppModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := transfer.DeleteWebAppInput{
		WebAppId: state.WebAppId.ValueStringPointer(),
	}

	_, err := conn.DeleteWebApp(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Transfer, create.ErrActionDeleting, ResNameWebApp, state.WebAppId.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitWebAppDeleted(ctx, conn, state.WebAppId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Transfer, create.ErrActionWaitingForDeletion, ResNameWebApp, state.WebAppId.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceWebApp) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

const (
	statusNormal = "Normal"
)

func waitWebAppCreated(ctx context.Context, conn *transfer.Client, id string, timeout time.Duration) (*awstypes.DescribedWebApp, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusWebApp(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DescribedWebApp); ok {
		return out, err
	}

	return nil, err
}

func waitWebAppDeleted(ctx context.Context, conn *transfer.Client, id string, timeout time.Duration) (*awstypes.DescribedWebApp, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusNormal},
		Target:  []string{},
		Refresh: statusWebApp(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DescribedWebApp); ok {
		return out, err
	}

	return nil, err
}

func statusWebApp(ctx context.Context, conn *transfer.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findWebAppByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, statusNormal, nil
	}
}

func findWebAppByID(ctx context.Context, conn *transfer.Client, id string) (*awstypes.DescribedWebApp, error) {
	input := transfer.DescribeWebAppInput{
		WebAppId: aws.String(id),
	}

	out, err := conn.DescribeWebApp(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.WebApp == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.WebApp, nil
}

type resourceWebAppModel struct {
	framework.WithRegionModel
	AccessEndpoint          types.String                                                  `tfsdk:"access_endpoint"`
	ARN                     types.String                                                  `tfsdk:"arn"`
	IdentityProviderDetails fwtypes.ListNestedObjectValueOf[identityProviderDetailsModel] `tfsdk:"identity_provider_details"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                `tfsdk:"timeouts"`
	WebAppEndpointPolicy    fwtypes.StringEnum[awstypes.WebAppEndpointPolicy]             `tfsdk:"web_app_endpoint_policy"`
	WebAppUnits             fwtypes.ListNestedObjectValueOf[webAppUnitsModel]             `tfsdk:"web_app_units"`
	WebAppId                types.String                                                  `tfsdk:"id"`
}

type identityProviderDetailsModel struct {
	IdentityCenterConfig fwtypes.ListNestedObjectValueOf[identityCenterConfigModel] `tfsdk:"identity_center_config"`
}

type identityCenterConfigModel struct {
	ApplicationArn types.String `tfsdk:"application_arn"`
	InstanceArn    types.String `tfsdk:"instance_arn"`
	Role           types.String `tfsdk:"role"`
}

type webAppUnitsModel struct {
	Provisioned types.Int64 `tfsdk:"provisioned"`
}

func sweepWebApps(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := transfer.ListWebAppsInput{}
	conn := client.TransferClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := transfer.NewListWebAppsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.WebApps {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceWebApp, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.WebAppId))),
			)
		}
	}

	return sweepResources, nil
}

var (
	_ flex.Expander  = webAppUnitsModel{}
	_ flex.Flattener = &webAppUnitsModel{}
	_ flex.Expander  = identityProviderDetailsModel{}
	_ flex.Flattener = &identityProviderDetailsModel{}
)

func (m webAppUnitsModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var v awstypes.WebAppUnits

	switch {
	case !m.Provisioned.IsNull():
		var apiObject awstypes.WebAppUnitsMemberProvisioned
		apiObject.Value = aws.ToInt32(flex.Int32FromFrameworkInt64(ctx, &m.Provisioned))
		v = &apiObject
	}

	return v, diags
}

func (m *webAppUnitsModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.WebAppUnitsMemberProvisioned:
		m.Provisioned = flex.Int32ToFrameworkInt64(ctx, &t.Value)
	}
	return diags
}

func (m identityProviderDetailsModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var v awstypes.WebAppIdentityProviderDetails

	switch {
	case !m.IdentityCenterConfig.IsNull():
		data, d := m.IdentityCenterConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var apiObject awstypes.WebAppIdentityProviderDetailsMemberIdentityCenterConfig
		diags.Append(flex.Expand(ctx, data, &apiObject.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		v = &apiObject
	}

	return v, diags
}

func (m *identityProviderDetailsModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.DescribedWebAppIdentityProviderDetailsMemberIdentityCenterConfig:
		var data identityCenterConfigModel
		diags.Append(flex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		m.IdentityCenterConfig = fwtypes.NewListNestedObjectValueOfPtrMust[identityCenterConfigModel](ctx, &data)
	default:
		diags.AddError("Interface Conversion Error", fmt.Sprintf("cannot flatten %T into %T", v, m))
	}
	return diags
}
