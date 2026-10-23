// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatchomni_dashboard", name="Dashboard")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("space_id")
// @IdentityAttribute("dashboard_id")
// @ImportIDHandler("dashboardImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types;awstypes;awstypes.OmniDashboard")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(importStateIdAttributes="space_id;dashboard_id", importStateIdAttributesSep="flex.ResourceIdSeparator")
func newDashboardResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &dashboardResource{}, nil
}

type dashboardResource struct {
	framework.ResourceWithModel[dashboardResourceModel]
	framework.WithImportByIdentity
}

func (r *dashboardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"body": schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Required:   true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1048576),
				},
			},
			"dashboard_id": framework.IDAttribute(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
				},
				PlanModifiers: []planmodifier.String{
					// UpdateOmniDashboard leaves an omitted description unchanged and rejects an empty one,
					// so removing the description requires replacement.
					stringplanmodifier.RequiresReplaceIf(func(ctx context.Context, request planmodifier.StringRequest, response *stringplanmodifier.RequiresReplaceIfFuncResponse) {
						response.RequiresReplace = request.PlanValue.IsNull() && !request.StateValue.IsNull()
					}, "Removing the description requires replacement.", "Removing the description requires replacement."),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z_.@~()-]+$`), "must contain only letters, numbers, and the characters _ . @ ~ ( ) -"),
				},
			},
			"space_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`), "must be a space ID"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *dashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dashboardResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input cloudwatchomni.CreateOmniDashboardInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	// CreateOmniDashboard takes a tag map, but the tag operations live in the cloudwatch package and use a slice.
	if tags := keyValueTags(ctx, getTagsIn(ctx)).Map(); len(tags) > 0 {
		input.Tags = tags
	}

	out, err := conn.CreateOmniDashboard(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out.OmniDashboard, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *dashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dashboardResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	spaceID, dashboardID := fwflex.StringValueFromFramework(ctx, state.SpaceID), fwflex.StringValueFromFramework(ctx, state.DashboardID)
	out, err := findDashboardByTwoPartKey(ctx, conn, spaceID, dashboardID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, dashboardID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *dashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state dashboardResourceModel
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
		dashboardID := fwflex.StringValueFromFramework(ctx, plan.DashboardID)
		var input cloudwatchomni.UpdateOmniDashboardInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateOmniDashboard(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, dashboardID)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out.OmniDashboard, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *dashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dashboardResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	spaceID, dashboardID := fwflex.StringValueFromFramework(ctx, state.SpaceID), fwflex.StringValueFromFramework(ctx, state.DashboardID)
	input := cloudwatchomni.DeleteOmniDashboardInput{
		DashboardId: aws.String(dashboardID),
		SpaceId:     aws.String(spaceID),
	}
	_, err := conn.DeleteOmniDashboard(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		// A missing space is reported as AccessDeniedException.
		if errs.IsA[*awstypes.AccessDeniedException](err) {
			_, findErr := findDashboardByTwoPartKey(ctx, conn, spaceID, dashboardID)
			if retry.NotFound(findErr) {
				return
			}
			if findErr != nil {
				smerr.AddError(ctx, &resp.Diagnostics, findErr, smerr.ID, dashboardID)
				return
			}
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, dashboardID)
		return
	}
}

func (r *dashboardResource) flatten(ctx context.Context, dashboard *awstypes.OmniDashboard, data *dashboardResourceModel) diag.Diagnostics {
	return fwflex.Flatten(ctx, dashboard, data)
}

var (
	_ inttypes.ImportIDParser = dashboardImportID{}
)

type dashboardImportID struct{}

func (dashboardImportID) Parse(id string) (string, map[string]any, error) {
	const (
		importIDPartCount = 2
	)
	parts, err := intflex.ExpandResourceId(id, importIDPartCount, false)
	if err != nil {
		return "", nil, fmt.Errorf("unexpected format for ID (%[1]s), expected space_id%[2]sdashboard_id: %w", id, intflex.ResourceIdSeparator, err)
	}

	result := map[string]any{
		"dashboard_id": parts[1],
		"space_id":     parts[0],
	}

	return id, result, nil
}

// findDashboardByTwoPartKey returns the dashboard, or a NotFoundError if the dashboard or its space doesn't exist.
func findDashboardByTwoPartKey(ctx context.Context, conn *cloudwatchomni.Client, spaceID, dashboardID string) (*awstypes.OmniDashboard, error) {
	input := cloudwatchomni.GetOmniDashboardInput{
		DashboardId: aws.String(dashboardID),
		SpaceId:     aws.String(spaceID),
	}

	out, err := conn.GetOmniDashboard(ctx, &input)
	switch {
	case errs.IsA[*awstypes.ResourceNotFoundException](err):
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	case errs.IsA[*awstypes.AccessDeniedException](err):
		// The service returns AccessDeniedException instead of ResourceNotFoundException for a missing space,
		// so only treat the error as not found once the space's absence is confirmed.
		_, spaceErr := findSpaceByID(ctx, conn, spaceID)
		if retry.NotFound(spaceErr) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		if spaceErr != nil {
			return nil, smarterr.NewError(spaceErr)
		}

		return nil, smarterr.NewError(err)
	case err != nil:
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.OmniDashboard == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.OmniDashboard, nil
}

type dashboardResourceModel struct {
	framework.WithRegionModel
	ARN         types.String         `tfsdk:"arn"`
	Body        jsontypes.Normalized `tfsdk:"body"`
	DashboardID types.String         `tfsdk:"dashboard_id"`
	Description types.String         `tfsdk:"description"`
	Name        types.String         `tfsdk:"name"`
	SpaceID     types.String         `tfsdk:"space_id"`
	Tags        tftags.Map           `tfsdk:"tags"`
	TagsAll     tftags.Map           `tfsdk:"tags_all"`
}
