// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package workmail

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workmail/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_workmail_organization", name="Organization")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("organization_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workmail;workmail.DescribeOrganizationOutput")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importIgnore="delete_directory")
// @Testing(importStateIdAttribute="organization_id")
func newOrganizationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &organizationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameOrganization = "Organization"
)

type organizationResource struct {
	framework.ResourceWithModel[organizationResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *organizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"completed_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"default_mail_domain": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"delete_directory": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"directory_id": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"directory_type": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"interoperability_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrKMSKeyARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"migration_admin": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_alias": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_id": framework.IDAttribute(),
			names.AttrState: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var plan organizationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input workmail.CreateOrganizationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Alias = plan.OrganizationAlias.ValueStringPointer()
	input.EnableInteroperability = plan.InteroperabilityEnabled.ValueBool()

	out, err := conn.CreateOrganization(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.OrganizationAlias.String())
		return
	}
	if out == nil || out.OrganizationId == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.OrganizationAlias.String())
		return
	}

	plan.OrganizationId = flex.StringToFramework(ctx, out.OrganizationId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitOrganizationCreated(ctx, conn, plan.OrganizationId.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.OrganizationId.String())
		return
	}

	findOutput, err := findOrganizationByID(ctx, conn, plan.OrganizationId.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.OrganizationId.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, findOutput, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	if err := createTags(ctx, conn, *findOutput.ARN, getTagsIn(ctx)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, *findOutput.ARN)
		return
	}

	plan.OrganizationAlias = types.StringPointerValue(findOutput.Alias)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var state organizationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findOrganizationByID(ctx, conn, state.OrganizationId.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.OrganizationId.String())
		return
	}

	if aws.ToString(out.State) == "Deleted" {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(&retry.NotFoundError{Message: fmt.Sprintf("WorkMail Organization %s is in Deleted state", state.OrganizationId.ValueString())}))
		resp.State.RemoveResource(ctx)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	state.OrganizationAlias = types.StringPointerValue(out.Alias)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state organizationResourceModel

	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// delete_directory is a local-only flag, no AWS API call needed, copy only this value into state
	state.DeleteDirectory = plan.DeleteDirectory
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var state organizationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := workmail.DeleteOrganizationInput{
		DeleteDirectory:                 state.DeleteDirectory.ValueBool(),
		DeleteIdentityCenterApplication: true,
		OrganizationId:                  state.OrganizationId.ValueStringPointer(),
	}

	_, err := conn.DeleteOrganization(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		} else if errs.IsA[*awstypes.OrganizationStateException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.OrganizationId.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitOrganizationDeleted(ctx, conn, state.OrganizationId.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.OrganizationId.String())
		return
	}
}

const (
	statusActive   = "Active"
	statusDeleting = "Deleting"
	statusDeleted  = "Deleted"
)

func waitOrganizationCreated(ctx context.Context, conn *workmail.Client, id string, timeout time.Duration) (*workmail.DescribeOrganizationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusActive},
		Refresh:                   statusOrganization(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*workmail.DescribeOrganizationOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitOrganizationDeleted(ctx context.Context, conn *workmail.Client, id string, timeout time.Duration) (*workmail.DescribeOrganizationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusActive, statusDeleting},
		Target:  []string{statusDeleted},
		Refresh: statusOrganization(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*workmail.DescribeOrganizationOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusOrganization(conn *workmail.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findOrganizationByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, aws.ToString(out.State), nil
	}
}

func findOrganizationByID(ctx context.Context, conn *workmail.Client, id string) (*workmail.DescribeOrganizationOutput, error) {
	input := workmail.DescribeOrganizationInput{
		OrganizationId: aws.String(id),
	}

	out, err := conn.DescribeOrganization(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.OrganizationId == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type organizationResourceModel struct {
	framework.WithRegionModel
	ARN                     types.String      `tfsdk:"arn"`
	CompletedDate           timetypes.RFC3339 `tfsdk:"completed_date"`
	DefaultMailDomain       types.String      `tfsdk:"default_mail_domain"`
	DeleteDirectory         types.Bool        `tfsdk:"delete_directory"`
	DirectoryId             types.String      `tfsdk:"directory_id"`
	DirectoryType           types.String      `tfsdk:"directory_type"`
	InteroperabilityEnabled types.Bool        `tfsdk:"interoperability_enabled"`
	KmsKeyARN               fwtypes.ARN       `tfsdk:"kms_key_arn"`
	MigrationAdmin          types.String      `tfsdk:"migration_admin"`
	OrganizationAlias       types.String      `tfsdk:"organization_alias"`
	OrganizationId          types.String      `tfsdk:"organization_id"`
	State                   types.String      `tfsdk:"state"`
	Tags                    tftags.Map        `tfsdk:"tags"`
	TagsAll                 tftags.Map        `tfsdk:"tags_all"`
	Timeouts                timeouts.Value    `tfsdk:"timeouts"`
}
