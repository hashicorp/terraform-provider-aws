// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_opensearch_application_capability", name="Application Capability")
func newApplicationCapabilityResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &applicationCapabilityResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	applicationCapabilityIDPartCount = 2
)

type applicationCapabilityResource struct {
	framework.ResourceWithModel[applicationCapabilityResourceModel]
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *applicationCapabilityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrApplicationID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"capability_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 30),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z0-9-]+$`),
						"must contain only alphanumeric characters and hyphens",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CapabilityStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"ai_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[aiConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *applicationCapabilityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan applicationCapabilityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchClient(ctx)

	applicationID := fwflex.StringValueFromFramework(ctx, plan.ApplicationID)
	capabilityName := fwflex.StringValueFromFramework(ctx, plan.CapabilityName)

	// AIConfig is the only CapabilityBaseRequestConfig union member the SDK exposes today.
	aiConfig, diags := plan.AIConfig.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if aiConfig == nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("ai_config"),
			"Missing capability configuration block",
			"the ai_config block must be provided",
		)
		return
	}

	input := opensearch.RegisterCapabilityInput{
		ApplicationId:  aws.String(applicationID),
		CapabilityName: aws.String(capabilityName),
		CapabilityConfig: &awstypes.CapabilityBaseRequestConfigMemberAiConfig{
			Value: awstypes.AIConfig{},
		},
	}

	_, err := conn.RegisterCapability(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("registering OpenSearch Application Capability (%s/%s)", applicationID, capabilityName), err.Error())
		return
	}

	out, err := waitApplicationCapabilityCreated(ctx, conn, applicationID, capabilityName, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for OpenSearch Application Capability (%s/%s) create", applicationID, capabilityName), err.Error())
		return
	}

	plan.Status = fwtypes.StringEnumValue(out.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *applicationCapabilityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state applicationCapabilityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchClient(ctx)

	applicationID := fwflex.StringValueFromFramework(ctx, state.ApplicationID)
	capabilityName := fwflex.StringValueFromFramework(ctx, state.CapabilityName)

	out, err := findApplicationCapabilityByTwoPartKey(ctx, conn, applicationID, capabilityName)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading OpenSearch Application Capability (%s/%s)", applicationID, capabilityName), err.Error())
		return
	}

	state.Status = fwtypes.StringEnumValue(out.Status)

	// Repopulate the configuration block from the returned union member. This is
	// required for import, where state has no block to begin with.
	switch out.CapabilityConfig.(type) {
	case *awstypes.CapabilityExtendedResponseConfigMemberAiConfig:
		aiConfig, diags := fwtypes.NewListNestedObjectValueOfPtr(ctx, &aiConfigModel{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.AIConfig = aiConfig
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationCapabilityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state applicationCapabilityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchClient(ctx)

	applicationID := fwflex.StringValueFromFramework(ctx, state.ApplicationID)
	capabilityName := fwflex.StringValueFromFramework(ctx, state.CapabilityName)

	input := opensearch.DeregisterCapabilityInput{
		ApplicationId:  aws.String(applicationID),
		CapabilityName: aws.String(capabilityName),
	}

	_, err := conn.DeregisterCapability(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deregistering OpenSearch Application Capability (%s/%s)", applicationID, capabilityName), err.Error())
		return
	}

	if _, err := waitApplicationCapabilityDeleted(ctx, conn, applicationID, capabilityName, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for OpenSearch Application Capability (%s/%s) delete", applicationID, capabilityName), err.Error())
		return
	}
}

func (r *applicationCapabilityResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(request.ID, applicationCapabilityIDPartCount, false)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrApplicationID), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("capability_name"), parts[1])...)
}

func waitApplicationCapabilityCreated(ctx context.Context, conn *opensearch.Client, applicationID, capabilityName string, timeout time.Duration) (*opensearch.GetCapabilityOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CapabilityStatusCreating, awstypes.CapabilityStatusUpdating),
		Target:  enum.Slice(awstypes.CapabilityStatusActive),
		Refresh: statusApplicationCapability(conn, applicationID, capabilityName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*opensearch.GetCapabilityOutput); ok {
		return out, err
	}

	return nil, err
}

func waitApplicationCapabilityDeleted(ctx context.Context, conn *opensearch.Client, applicationID, capabilityName string, timeout time.Duration) (*opensearch.GetCapabilityOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CapabilityStatusDeleting, awstypes.CapabilityStatusActive),
		Target:  []string{},
		Refresh: statusApplicationCapability(conn, applicationID, capabilityName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*opensearch.GetCapabilityOutput); ok {
		return out, err
	}

	return nil, err
}

func statusApplicationCapability(conn *opensearch.Client, applicationID, capabilityName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findApplicationCapabilityByTwoPartKey(ctx, conn, applicationID, capabilityName)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findApplicationCapabilityByTwoPartKey(ctx context.Context, conn *opensearch.Client, applicationID, capabilityName string) (*opensearch.GetCapabilityOutput, error) {
	input := opensearch.GetCapabilityInput{
		ApplicationId:  aws.String(applicationID),
		CapabilityName: aws.String(capabilityName),
	}

	output, err := conn.GetCapability(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type applicationCapabilityResourceModel struct {
	framework.WithRegionModel
	AIConfig       fwtypes.ListNestedObjectValueOf[aiConfigModel] `tfsdk:"ai_config"`
	ApplicationID  types.String                                   `tfsdk:"application_id"`
	CapabilityName types.String                                   `tfsdk:"capability_name"`
	Status         fwtypes.StringEnum[awstypes.CapabilityStatus]  `tfsdk:"status"`
	Timeouts       timeouts.Value                                 `tfsdk:"timeouts"`
}

type aiConfigModel struct{}
