// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_registry_record", name="Registry Record")
// @IdentityAttribute("registry_id")
// @IdentityAttribute("record_id")
// @ImportIDHandler(registryRecordImportID)
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator="randomWithPrefixAndUnderscore(t)")
// @Testing(importStateIdAttribute="record_id")
// @Testing(preCheck="testAccPreCheckRegistries")
func newRegistryRecordResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &registryRecordResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type registryRecordResource struct {
	framework.ResourceWithModel[registryRecordResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *registryRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"descriptor_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DescriptorType](),
				Required:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"record_arn": framework.ARNAttributeComputedOnly(),
			"record_id":  framework.IDAttribute(),
			"registry_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *registryRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan registryRecordResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input bedrockagentcorecontrol.CreateRegistryRecordInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	out, err := conn.CreateRegistryRecord(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	recordARN, registryID := aws.ToString(out.RecordArn), fwflex.StringValueFromFramework(ctx, plan.RegistryID)
	created, err := waitRegistryRecordCreated(ctx, conn, recordARN, registryID, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordARN)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, created, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *registryRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state registryRecordResourceModel
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

func (r *registryRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state registryRecordResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		registryID, recordID := fwflex.StringValueFromFramework(ctx, plan.RegistryID), fwflex.StringValueFromFramework(ctx, plan.RecordID)
		var input bedrockagentcorecontrol.UpdateRegistryRecordInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateRegistryRecord(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
			return
		}

		if _, err := waitRegistryRecordUpdated(ctx, conn, registryID, recordID, r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *registryRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state registryRecordResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	registryID, recordID := fwflex.StringValueFromFramework(ctx, state.RegistryID), fwflex.StringValueFromFramework(ctx, state.RecordID)
	input := bedrockagentcorecontrol.DeleteRegistryRecordInput{
		RecordId:   aws.String(recordID),
		RegistryId: aws.String(registryID),
	}

	_, err := conn.DeleteRegistryRecord(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
		return
	}

	if _, err := waitRegistryRecordDeleted(ctx, conn, registryID, recordID, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
		return
	}
}

func (r *registryRecordResource) flatten(ctx context.Context, registryRecord *bedrockagentcorecontrol.GetRegistryRecordOutput, data *registryRecordResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, registryRecord, data)...)
	return diags
}

func findRegistryRecordByTwoPartKey(ctx context.Context, conn *bedrockagentcorecontrol.Client, registryID, recordID string) (*bedrockagentcorecontrol.GetRegistryRecordOutput, error) {
	input := bedrockagentcorecontrol.GetRegistryRecordInput{
		RecordId:   aws.String(recordID),
		RegistryId: aws.String(registryID),
	}

	return findRegistryRecord(ctx, conn, &input)
}

func findRegistryRecord(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetRegistryRecordInput) (*bedrockagentcorecontrol.GetRegistryRecordOutput, error) {
	out, err := conn.GetRegistryRecord(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

func statusRegistryRecord(conn *bedrockagentcorecontrol.Client, registryID, recordID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findRegistryRecordByTwoPartKey(ctx, conn, registryID, recordID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func waitRegistryRecordCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, registryID, recordID string, timeout time.Duration) (*bedrockagentcorecontrol.GetRegistryRecordOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RegistryRecordStatusCreating),
		Target:                    enum.Slice(awstypes.RegistryRecordStatusApproved),
		Refresh:                   statusRegistryRecord(conn, registryID, recordID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetRegistryRecordOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitRegistryRecordUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, registryID, recordID string, timeout time.Duration) (*bedrockagentcorecontrol.GetRegistryRecordOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RegistryRecordStatusUpdating),
		Target:                    enum.Slice(awstypes.RegistryRecordStatusApproved),
		Refresh:                   statusRegistryRecord(conn, registryID, recordID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetRegistryRecordOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitRegistryRecordDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, registryID, recordID string, timeout time.Duration) (*bedrockagentcorecontrol.GetRegistryRecordOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"DELETING"},
		Target:  []string{},
		Refresh: statusRegistryRecord(conn, registryID, recordID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetRegistryRecordOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

var (
	_ inttypes.ImportIDParser = registryRecordImportID{}
)

type registryRecordImportID struct{}

func (registryRecordImportID) Parse(id string) (string, map[string]any, error) {
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

type registryRecordResourceModel struct {
	framework.WithRegionModel
	Description    types.String                                `tfsdk:"description"`
	DescriptorType fwtypes.StringEnum[awstypes.DescriptorType] `tfsdk:"descriptor_type"`
	Name           types.String                                `tfsdk:"name"`
	RecordARN      types.String                                `tfsdk:"record_arn"`
	RecordID       types.String                                `tfsdk:"record_id"`
	RegistryID     types.String                                `tfsdk:"registry_id"`
	Timeouts       timeouts.Value                              `tfsdk:"timeouts"`
}
