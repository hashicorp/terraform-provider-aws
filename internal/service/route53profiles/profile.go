// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53profiles

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53profiles"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53profiles/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource struct with schema method
// 4. Create, read, update, delete methods (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_route53profiles_profile", name="Profile")
func newResourceProfile(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceProfile{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameProfile = "Profile"
)

type resourceProfile struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[resourceProfileData]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceProfile) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_route53profiles_profile"
}

func (r *resourceProfile) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"id":  framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
				Read:   true,
			}),
		},
	}
}

func (r *resourceProfile) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().Route53ProfilesClient(ctx)

	var data resourceProfileData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &route53profiles.CreateProfileInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)

	// Additional fields
	input.ClientToken = aws.String(id.UniqueId())

	output, err := conn.CreateProfile(ctx, input)
	name := data.Name

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating route53 profile: (%s)", name), err.Error())
		return
	}

	data.ARN = fwflex.StringToFramework(ctx, output.Profile.Arn)
	data.ID = fwflex.StringToFramework(ctx, output.Profile.Id)

	if _, err := waitProfileCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for route53 profile (%s) created", name), err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceProfile) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().Route53ProfilesClient(ctx)

	var state resourceProfileData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindProfileByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionSetting, ResNameProfile, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceProfile) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().Route53ProfilesClient(ctx)

	var state resourceProfileData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &route53profiles.DeleteProfileInput{
		ProfileId: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteProfile(ctx, in)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionDeleting, ResNameProfile, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitProfileDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Profiles, create.ErrActionWaitingForDeletion, ResNameProfile, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceProfile) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitProfileCreated(ctx context.Context, conn *route53profiles.Client, id string, timeout time.Duration) (*route53profiles.GetProfileOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ProfileStatusCreating),
		Target:                    enum.Slice(awstypes.ProfileStatusComplete),
		Refresh:                   statusProfile(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*route53profiles.GetProfileOutput); ok {
		return out, err
	}

	return nil, err
}

func waitProfileDeleted(ctx context.Context, conn *route53profiles.Client, id string, timeout time.Duration) (*route53profiles.DeleteProfileOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ProfileStatusDeleting),
		Target:  []string{},
		Refresh: statusProfile(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*route53profiles.DeleteProfileOutput); ok {
		return out, err
	}

	return nil, err
}

func statusProfile(ctx context.Context, conn *route53profiles.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindProfileByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func FindProfileByID(ctx context.Context, conn *route53profiles.Client, id string) (*awstypes.Profile, error) {
	in := &route53profiles.GetProfileInput{
		ProfileId: aws.String(id),
	}

	out, err := conn.GetProfile(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Profile == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Profile, nil
}

type resourceProfileData struct {
	ARN      types.String   `tfsdk:"arn"`
	ID       types.String   `tfsdk:"id"`
	Name     types.String   `tfsdk:"name"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
