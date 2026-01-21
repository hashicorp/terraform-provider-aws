// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ram_permission", name="Permission")
// @ArnIdentity(identityDuplicateAttributes="id")
// @Tags(identifierAttribute="arn")
// @Testing(importStateIdAttribute="arn")
// @Testing(importIgnore="policy_template")
func newPermissionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePermission{}

	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type resourcePermission struct {
	framework.ResourceWithModel[resourcePermissionModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *resourcePermission) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"default_version": schema.BoolAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrARN)),
			names.AttrLastUpdatedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 36),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^\S[\w.-]*$`),
						"value must contain letters and numbers",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_template": schema.StringAttribute{
				Required: true,
			},
			names.AttrResourceType: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVersion: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Delete: true,
			}),
		},
	}
}

func (r *resourcePermission) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RAMClient(ctx)

	var plan resourcePermissionModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	input := ram.CreatePermissionInput{
		ClientToken: aws.String(sdkid.UniqueId()),
	}
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreatePermission(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.Permission == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.Permission, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	plan.setID()
	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourcePermission) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RAMClient(ctx)

	var state resourcePermissionModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPermissionByARN(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.String())
		return
	}
	setTagsOut(ctx, out.Tags)

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourcePermission) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().RAMClient(ctx)

	var plan, state resourcePermissionModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		if err := permissionPruneVersions(ctx, conn, plan.ARN.ValueString()); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.String())
			return
		}

		input := ram.CreatePermissionVersionInput{
			ClientToken:    aws.String(sdkid.UniqueId()),
			PermissionArn:  plan.ARN.ValueStringPointer(),
			PolicyTemplate: plan.PolicyTemplate.ValueStringPointer(),
		}

		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.CreatePermissionVersion(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.String())
			return
		}
		if out == nil || out.Permission == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ARN.String())
			return
		}

		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.Permission, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourcePermission) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RAMClient(ctx)

	var state resourcePermissionModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := ram.DeletePermissionInput{
		PermissionArn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeletePermission(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.UnknownResourceException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.String())
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitPermissionDeleted(ctx, conn, state.ARN.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.String())
		return
	}
}

func waitPermissionDeleted(ctx context.Context, conn *ram.Client, arn string, timeout time.Duration) (*awstypes.ResourceSharePermissionDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PermissionStatusDeleting),
		Target:  enum.Slice(awstypes.PermissionStatusDeleted),
		Refresh: statusPermission(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResourceSharePermissionDetail); ok {
		return output, err
	}

	return nil, err
}

func statusPermission(ctx context.Context, conn *ram.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findPermissionByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findPermissionByARN(ctx context.Context, conn *ram.Client, arn string) (*awstypes.ResourceSharePermissionDetail, error) {
	input := ram.GetPermissionInput{
		PermissionArn: &arn,
	}

	out, err := conn.GetPermission(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.UnknownResourceException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Permission == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.Permission, nil
}

type resourcePermissionModel struct {
	framework.WithRegionModel
	ARN             types.String      `tfsdk:"arn"`
	CreationTime    timetypes.RFC3339 `tfsdk:"creation_time"`
	ID              types.String      `tfsdk:"id"`
	DefaultVersion  types.Bool        `tfsdk:"default_version"`
	LastUpdatedTime timetypes.RFC3339 `tfsdk:"last_updated_time"`
	Name            types.String      `tfsdk:"name"`
	PolicyTemplate  types.String      `tfsdk:"policy_template"`
	ResourceType    types.String      `tfsdk:"resource_type"`
	Tags            tftags.Map        `tfsdk:"tags"`
	TagsAll         tftags.Map        `tfsdk:"tags_all"`
	Timeouts        timeouts.Value    `tfsdk:"timeouts"`
	Status          types.String      `tfsdk:"status"`
	Version         types.String      `tfsdk:"version"`
}

func (data *resourcePermissionModel) setID() {
	data.ID = data.ARN
}

// policyPruneVersions deletes the oldest version.
//
// Old versions are deleted until there are 4 or less remaining, which means at
// least one more can be created before hitting the maximum of 5.
//
// The default version is never deleted.
func permissionPruneVersions(ctx context.Context, conn *ram.Client, arn string) error {
	versions, err := findPermissionVersionsByARN(ctx, conn, arn)

	if err != nil {
		return err
	}

	if len(versions) < 5 {
		return nil
	}

	oldestVersion := versions[0]
	for _, version := range versions {
		if *version.DefaultVersion {
			continue
		}

		if version.CreationTime.Before(aws.ToTime(oldestVersion.CreationTime)) {
			oldestVersion = version
		}
	}

	versionInt, err := strconv.Atoi(aws.ToString(oldestVersion.Version))
	if err != nil {
		return fmt.Errorf("failed to parse version '%s' to int: %w", aws.ToString(oldestVersion.Version), err)
	}
	return permissionDeleteVersion(ctx, conn, arn, int32(versionInt))
}

func permissionDeleteVersion(ctx context.Context, conn *ram.Client, arn string, versionID int32) error {
	input := &ram.DeletePermissionVersionInput{
		PermissionArn:     aws.String(arn),
		PermissionVersion: aws.Int32(versionID),
	}

	_, err := conn.DeletePermissionVersion(ctx, input)

	if err != nil {
		return fmt.Errorf("deleting RAM Permission (%s) version (%d): %w", arn, versionID, err)
	}

	return nil
}

func findPermissionVersionsByARN(ctx context.Context, conn *ram.Client, arn string) ([]awstypes.ResourceSharePermissionSummary, error) {
	input := &ram.ListPermissionVersionsInput{
		PermissionArn: aws.String(arn),
	}
	var output []awstypes.ResourceSharePermissionSummary

	pages := ram.NewListPermissionVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.UnknownResourceException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Permissions {
			if !reflect.ValueOf(v).IsZero() {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
