// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53profiles

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53profiles"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53profiles/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_route53profiles_profile", name="Profile")
// @Tags("identifierAttribute=arn")
func newResourceProfile(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceProfile{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultReadTimeout(30 * time.Minute)
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
			"owner_id": schema.StringAttribute{
				Computed: true,
			},
			"share_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ShareStatus](),
				Computed:   true,
			},
			"status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ProfileStatus](),
				Computed:   true,
			},
			"status_message": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
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

	var tags []awstypes.Tag

	// Even tough the tags are Map based, CreateProfile expects a slice of tags
	for k, v := range getTagsIn(ctx) {
		k := k
		v := v
		tags = append(tags, awstypes.Tag{Key: &k, Value: &v})
	}

	// Additional fields
	input.ClientToken = aws.String(id.UniqueId())
	input.Tags = tags

	output, err := conn.CreateProfile(ctx, input)
	name := data.Name

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating route53 profile: (%s)", name), err.Error())
		return
	}

	data.ARN = fwflex.StringToFramework(ctx, output.Profile.Arn)
	data.ID = fwflex.StringToFramework(ctx, output.Profile.Id)
	data.OwnerId = fwflex.StringToFramework(ctx, output.Profile.OwnerId)
	data.ShareStatus = fwtypes.StringEnumValue(output.Profile.ShareStatus)
	data.Status = fwtypes.StringEnumValue(output.Profile.Status)
	data.StatusMessage = fwflex.StringToFramework(ctx, output.Profile.StatusMessage)

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

	out, err := findProfileByID(ctx, conn, state.ID.ValueString())

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

func (r *resourceProfile) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var oldProfile, newProfile resourceProfileData
	resp.Diagnostics.Append(req.State.Get(ctx, &oldProfile)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &newProfile)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update is only called when `tags` are updated
	// Set unknown values on newProfile to oldProfile
	newProfile.OwnerId = oldProfile.OwnerId
	newProfile.ShareStatus = oldProfile.ShareStatus
	newProfile.Status = oldProfile.Status
	newProfile.StatusMessage = oldProfile.StatusMessage

	resp.Diagnostics.Append(resp.State.Set(ctx, &newProfile)...)
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
		out, err := findProfileByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func (r *resourceProfile) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func findProfileByID(ctx context.Context, conn *route53profiles.Client, id string) (*awstypes.Profile, error) {
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
	ARN           types.String                               `tfsdk:"arn"`
	ID            types.String                               `tfsdk:"id"`
	Name          types.String                               `tfsdk:"name"`
	OwnerId       types.String                               `tfsdk:"owner_id"`
	ShareStatus   fwtypes.StringEnum[awstypes.ShareStatus]   `tfsdk:"share_status"`
	Status        fwtypes.StringEnum[awstypes.ProfileStatus] `tfsdk:"status"`
	StatusMessage types.String                               `tfsdk:"status_message"`
	Tags          types.Map                                  `tfsdk:"tags"`
	TagsAll       types.Map                                  `tfsdk:"tags_all"`
	Timeouts      timeouts.Value                             `tfsdk:"timeouts"`
}
