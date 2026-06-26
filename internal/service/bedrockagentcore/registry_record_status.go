// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_registry_record_status", name="Registry Record Status")
// @IdentityAttribute("registry_id")
// @IdentityAttribute("record_id")
// @ImportIDHandler(registryRecordStatusImportID)
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="record_id")
// @Testing(preCheck="testAccPreCheckRegistries")
// @Testing(importStateIdFunc="testAccRegistryRecordStatusImportStateIDFunc")
// @Testing(hasExistsFunction=false)
// @Testing(checkDestroyNoop=true)
func newRegistryRecordStatusResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &registryRecordStatusResource{}
	return r, nil
}

type registryRecordStatusResource struct {
	framework.ResourceWithModel[registryRecordStatusResourceModel]
	framework.WithNoOpDelete
	framework.WithImportByIdentity
}

func (r *registryRecordStatusResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"record_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"registry_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(enum.Slice(awstypes.RegistryRecordStatusApproved, awstypes.RegistryRecordStatusRejected, awstypes.RegistryRecordStatusDeprecated)...),
				},
			},
			names.AttrStatusReason: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 255),
				},
			},
		},
	}
}

func (r *registryRecordStatusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan registryRecordStatusResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	recordID := fwflex.StringValueFromFramework(ctx, plan.RecordID)
	var input bedrockagentcorecontrol.UpdateRegistryRecordStatusInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateRegistryRecordStatus(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *registryRecordStatusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state registryRecordStatusResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	registryID, recordID := fwflex.StringValueFromFramework(ctx, state.RegistryID), fwflex.StringValueFromFramework(ctx, state.RecordID)
	out, err := findRegistryRecordByTwoPartKey(ctx, conn, registryID, recordID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *registryRecordStatusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan registryRecordStatusResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	recordID := fwflex.StringValueFromFramework(ctx, plan.RecordID)
	var input bedrockagentcorecontrol.UpdateRegistryRecordStatusInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateRegistryRecordStatus(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *registryRecordStatusResource) flatten(ctx context.Context, registryRecord *bedrockagentcorecontrol.GetRegistryRecordOutput, data *registryRecordStatusResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, registryRecord, data)...)
	return diags
}

type registryRecordStatusResourceModel struct {
	framework.WithRegionModel
	RecordID     types.String `tfsdk:"record_id"`
	RegistryID   types.String `tfsdk:"registry_id"`
	Status       types.String `tfsdk:"status"`
	StatusReason types.String `tfsdk:"status_reason"`
}

var (
	_ inttypes.ImportIDParser = registryRecordStatusImportID{}
)

type registryRecordStatusImportID struct{}

func (registryRecordStatusImportID) Parse(id string) (string, map[string]any, error) {
	const (
		registryRecordIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(id, registryRecordIDParts, true)

	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"registry_id": parts[0],
		"record_id":   parts[1],
	}

	return id, result, nil
}
