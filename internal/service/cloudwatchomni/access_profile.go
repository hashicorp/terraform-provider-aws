// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

// @FrameworkResource("aws_cloudwatchomni_access_profile", name="Access Profile")
// @IdentityAttribute("space_id")
// @IdentityAttribute("profile_id")
// @ImportIDHandler("accessProfileImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types;awstypes;awstypes.AccessProfile")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(importStateIdFunc="testAccAccessProfileImportStateIDFunc")
// @Testing(importStateIdAttribute="profile_id")
func newAccessProfileResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &accessProfileResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)

	return r, nil
}

type accessProfileResource struct {
	framework.ResourceWithModel[accessProfileResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *accessProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"assume_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AssumeStatus](),
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"profile_id": framework.IDAttribute(),
			"profile_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AccessProfileType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *accessProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accessProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input cloudwatchomni.CreateAccessProfileInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	out, err := conn.CreateAccessProfile(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	spaceID, profileID := fwflex.StringValueFromFramework(ctx, plan.SpaceID), aws.ToString(out.AccessProfile.ProfileId)
	profile, err := waitAccessProfileCreated(ctx, conn, spaceID, profileID, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		// Taint the resource.
		resp.State.SetAttribute(ctx, path.Root("profile_id"), profileID)
		resp.State.SetAttribute(ctx, path.Root("space_id"), spaceID)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, profileID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, profile, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *accessProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state accessProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	spaceID, profileID := fwflex.StringValueFromFramework(ctx, state.SpaceID), fwflex.StringValueFromFramework(ctx, state.ProfileID)
	out, err := findAccessProfileByTwoPartKey(ctx, conn, spaceID, profileID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, profileID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *accessProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state accessProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		profileID := fwflex.StringValueFromFramework(ctx, plan.ProfileID)
		var input cloudwatchomni.UpdateAccessProfileInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		// An omitted description leaves the existing value unchanged, so clear it explicitly.
		if plan.Description.IsNull() && !state.Description.IsNull() {
			input.Description = aws.String("")
		}

		_, err := conn.UpdateAccessProfile(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, profileID)
			return
		}

		// The UpdateAccessProfile response omits some fields (e.g. profileType), so re-read the profile.
		spaceID := fwflex.StringValueFromFramework(ctx, plan.SpaceID)
		profile, err := findAccessProfileByTwoPartKey(ctx, conn, spaceID, profileID)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, profileID)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, profile, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		// Computed attributes without UseStateForUnknown are unknown in the plan.
		plan.AssumeStatus = state.AssumeStatus
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *accessProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state accessProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	spaceID, profileID := fwflex.StringValueFromFramework(ctx, state.SpaceID), fwflex.StringValueFromFramework(ctx, state.ProfileID)
	input := cloudwatchomni.DeleteAccessProfileInput{
		ProfileId: aws.String(profileID),
		SpaceId:   aws.String(spaceID),
	}
	_, err := conn.DeleteAccessProfile(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, profileID)
		return
	}

	const (
		accessProfileDeletedTimeout = 5 * time.Minute
	)
	_, err = tfresource.RetryUntilNotFound(ctx, accessProfileDeletedTimeout, func(ctx context.Context) (any, error) {
		return findAccessProfileByTwoPartKey(ctx, conn, spaceID, profileID)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, profileID)
		return
	}
}

func (r *accessProfileResource) flatten(ctx context.Context, profile *awstypes.AccessProfile, data *accessProfileResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	// A cleared description may be returned as an empty string; treat it as absent.
	if aws.ToString(profile.Description) == "" {
		profile.Description = nil
	}
	diags.Append(fwflex.Flatten(ctx, profile, data)...)
	return diags
}

func findAccessProfileByTwoPartKey(ctx context.Context, conn *cloudwatchomni.Client, spaceID, profileID string) (*awstypes.AccessProfile, error) {
	input := cloudwatchomni.GetAccessProfileInput{
		ProfileId: aws.String(profileID),
		SpaceId:   aws.String(spaceID),
	}

	return findAccessProfile(ctx, conn, &input)
}

func findAccessProfile(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.GetAccessProfileInput) (*awstypes.AccessProfile, error) {
	out, err := conn.GetAccessProfile(ctx, input)
	switch {
	case errs.IsA[*awstypes.ResourceNotFoundException](err):
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	// GetAccessProfile reports a missing parent space as AccessDeniedException,
	// so confirm the space's absence before treating the profile as not found.
	case errs.IsA[*awstypes.AccessDeniedException](err):
		_, serr := findSpaceByID(ctx, conn, aws.ToString(input.SpaceId))
		if retry.NotFound(serr) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		if serr != nil {
			return nil, smarterr.NewError(serr)
		}
		return nil, smarterr.NewError(err)
	case err != nil:
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.AccessProfile == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.AccessProfile, nil
}

// waitAccessProfileCreated waits for the service to populate AssumeStatus, which is set asynchronously after creation.
func waitAccessProfileCreated(ctx context.Context, conn *cloudwatchomni.Client, spaceID, profileID string, timeout time.Duration) (*awstypes.AccessProfile, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{""},
		Target:  enum.Values[awstypes.AssumeStatus](),
		Refresh: statusAccessProfile(conn, spaceID, profileID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AccessProfile); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusAccessProfile(conn *cloudwatchomni.Client, spaceID, profileID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findAccessProfileByTwoPartKey(ctx, conn, spaceID, profileID)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.AssumeStatus), nil
	}
}

type accessProfileResourceModel struct {
	framework.WithRegionModel
	ARN          types.String                                   `tfsdk:"arn"`
	AssumeStatus fwtypes.StringEnum[awstypes.AssumeStatus]      `tfsdk:"assume_status"`
	Description  types.String                                   `tfsdk:"description"`
	Name         types.String                                   `tfsdk:"name"`
	ProfileID    types.String                                   `tfsdk:"profile_id"`
	ProfileType  fwtypes.StringEnum[awstypes.AccessProfileType] `tfsdk:"profile_type"`
	SpaceID      types.String                                   `tfsdk:"space_id"`
	Timeouts     timeouts.Value                                 `tfsdk:"timeouts"`
}

var _ inttypes.ImportIDParser = accessProfileImportID{}

type accessProfileImportID struct{}

func (accessProfileImportID) Parse(id string) (string, map[string]any, error) {
	const (
		accessProfileImportIDPartCount = 2
	)
	parts, err := intflex.ExpandResourceId(id, accessProfileImportIDPartCount, false)
	if err != nil {
		return "", nil, fmt.Errorf("unexpected format for ID (%[1]s), expected space_id%[2]sprofile_id: %w", id, intflex.ResourceIdSeparator, err)
	}

	result := map[string]any{
		"profile_id": parts[1],
		"space_id":   parts[0],
	}

	return id, result, nil
}
