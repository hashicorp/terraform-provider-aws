// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workmail

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workmail/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_workmail_organization", name="Organization")
func newResourceOrganization(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceOrganization{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameOrganization = "Organization"
)

type resourceOrganization struct {
	framework.ResourceWithModel[resourceOrganizationModel]
	framework.WithTimeouts
	framework.WithImportByID
}

// Schema leaving out directory_id, kms_key_arn, domains, and enable_interoperability until later
func (r *resourceOrganization) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(), // OrganizationId
			names.AttrAlias: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceOrganization) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var plan resourceOrganizationModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input workmail.CreateOrganizationInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())

	out, err := conn.CreateOrganization(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Alias.String())
		return
	}
	if out == nil || out.OrganizationId == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Alias.String())
		return
	}

	found, err := FindOrganizationByID(ctx, conn, *out.OrganizationId)
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, *out.OrganizationId)
		return
	}

	// add missing values, arn is not returned after create
	plan.ID = fwflex.StringToFramework(ctx, found.OrganizationId)
	plan.ARN = fwflex.StringToFramework(ctx, found.ARN)

	smerr.EnrichAppend(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitOrganizationCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Alias.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceOrganization) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var state resourceOrganizationModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindOrganizationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	// resource in delete state
	if aws.ToString(out.State) == statusDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	// add missing values
	state.ID = fwflex.StringToFramework(ctx, out.OrganizationId)
	state.ARN = fwflex.StringToFramework(ctx, out.ARN)

	smerr.EnrichAppend(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceOrganization) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var state resourceOrganizationModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindOrganizationByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	// already deleted
	if aws.ToString(out.State) == statusDeleted {
		return
	}

	input := workmail.DeleteOrganizationInput{
		OrganizationId: state.ID.ValueStringPointer(),
		ClientToken:    aws.String(sdkid.UniqueId()),
	}

	_, err = conn.DeleteOrganization(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitOrganizationDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

const (
	statusRequested = "Requested"
	statusCreating  = "Creating"
	statusDeleting  = "Deleting"
	statusActive    = "Active"
	statusDeleted   = "Deleted"
)

func waitOrganizationCreated(ctx context.Context, conn *workmail.Client, id string, timeout time.Duration) (*workmail.DescribeOrganizationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusRequested, statusCreating},
		Target:                    []string{statusActive},
		Refresh:                   statusOrganization(ctx, conn, id),
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
		Pending: []string{statusDeleting},
		Target:  []string{statusDeleted},
		Refresh: statusOrganization(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*workmail.DescribeOrganizationOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusOrganization(ctx context.Context, conn *workmail.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := FindOrganizationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, aws.ToString(out.State), nil
	}
}

func FindOrganizationByID(ctx context.Context, conn *workmail.Client, id string) (*workmail.DescribeOrganizationOutput, error) {
	input := workmail.DescribeOrganizationInput{
		OrganizationId: aws.String(id),
	}

	out, err := conn.DescribeOrganization(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.OrganizationId == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out, nil
}

type resourceOrganizationModel struct {
	framework.WithRegionModel
	ARN      types.String   `tfsdk:"arn"`
	ID       types.String   `tfsdk:"id"` // from organisationId
	Alias    types.String   `tfsdk:"alias"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func sweepOrganizations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := workmail.ListOrganizationsInput{}
	conn := client.WorkMailClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := workmail.NewListOrganizationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.OrganizationSummaries {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceOrganization, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.OrganizationId))),
			)
		}
	}

	return sweepResources, nil
}
