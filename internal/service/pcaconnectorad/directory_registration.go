// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pcaconnectorad

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pcaconnectorad"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pcaconnectorad/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Directory Registration")
// @Tags(identifierAttribute="arn")
func newResourceDirectoryRegistration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDirectoryRegistration{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDirectoryRegistration = "Directory Registration"
)

type resourceDirectoryRegistration struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDirectoryRegistration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_pcaconnectorad_directory_registration"
}

func (r *resourceDirectoryRegistration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"directory_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id":              framework.IDAttribute(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceDirectoryRegistration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().PCAConnectorADClient(ctx)

	var plan resourceDirectoryRegistrationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &pcaconnectorad.CreateDirectoryRegistrationInput{
		DirectoryId: aws.String(plan.DirectoryID.ValueString()),
		Tags:        getTagsIn(ctx),
	}

	out, err := conn.CreateDirectoryRegistration(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PCAConnectorAD, create.ErrActionCreating, ResNameDirectoryRegistration, plan.DirectoryID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.DirectoryRegistrationArn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PCAConnectorAD, create.ErrActionCreating, ResNameDirectoryRegistration, plan.DirectoryID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.DirectoryRegistrationArn)
	plan.ID = flex.StringToFramework(ctx, out.DirectoryRegistrationArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitDirectoryRegistrationCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PCAConnectorAD, create.ErrActionWaitingForCreation, ResNameDirectoryRegistration, plan.DirectoryID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDirectoryRegistration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().PCAConnectorADClient(ctx)

	var state resourceDirectoryRegistrationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDirectoryRegistrationByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PCAConnectorAD, create.ErrActionSetting, ResNameDirectoryRegistration, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.DirectoryID = flex.StringToFramework(ctx, out.DirectoryId)
	state.ID = flex.StringToFramework(ctx, out.Arn)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDirectoryRegistration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceDirectoryRegistration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().PCAConnectorADClient(ctx)

	var state resourceDirectoryRegistrationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &pcaconnectorad.DeleteDirectoryRegistrationInput{
		DirectoryRegistrationArn: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteDirectoryRegistration(ctx, in)

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PCAConnectorAD, create.ErrActionDeleting, ResNameDirectoryRegistration, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitDirectoryRegistrationDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PCAConnectorAD, create.ErrActionWaitingForDeletion, ResNameDirectoryRegistration, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceDirectoryRegistration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *resourceDirectoryRegistration) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func waitDirectoryRegistrationCreated(ctx context.Context, conn *pcaconnectorad.Client, id string, timeout time.Duration) (*awstypes.DirectoryRegistration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.DirectoryRegistrationStatusCreating)},
		Target:                    []string{string(awstypes.DirectoryRegistrationStatusActive)},
		Refresh:                   statusDirectoryRegistration(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DirectoryRegistration); ok {
		return out, err
	}

	return nil, err
}

func waitDirectoryRegistrationDeleted(ctx context.Context, conn *pcaconnectorad.Client, id string, timeout time.Duration) (*awstypes.DirectoryRegistration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(awstypes.DirectoryRegistrationStatusDeleting), string(awstypes.DirectoryRegistrationStatusActive)},
		Target:  []string{},
		Refresh: statusDirectoryRegistration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DirectoryRegistration); ok {
		return out, err
	}

	return nil, err
}

func statusDirectoryRegistration(ctx context.Context, conn *pcaconnectorad.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findDirectoryRegistrationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.Status)), nil
	}
}

func findDirectoryRegistrationByID(ctx context.Context, conn *pcaconnectorad.Client, id string) (*awstypes.DirectoryRegistration, error) {
	in := &pcaconnectorad.GetDirectoryRegistrationInput{
		DirectoryRegistrationArn: aws.String(id),
	}

	out, err := conn.GetDirectoryRegistration(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.DirectoryRegistration == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.DirectoryRegistration, nil
}

type resourceDirectoryRegistrationData struct {
	ARN         types.String   `tfsdk:"arn"`
	DirectoryID types.String   `tfsdk:"directory_id"`
	ID          types.String   `tfsdk:"id"`
	Tags        types.Map      `tfsdk:"tags"`
	TagsAll     types.Map      `tfsdk:"tags_all"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}
