// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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

// @FrameworkResource("aws_ram_permission", name="Permission")
// @ArnIdentity
// @Tags(identifierAttribute="arn", resourceType="Permission")
// @Testing(importStateIdAttribute="arn")
// @Testing(importIgnore="policy_template")
func newPermissionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &permissionResource{}

	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type permissionResource struct {
	framework.ResourceWithModel[permissionResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *permissionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"default_version": schema.BoolAttribute{
				Computed: true,
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

func (r *permissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan permissionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RAMClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input ram.CreatePermissionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreatePermission(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	// Set unknowns.
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out.Permission, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *permissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state permissionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RAMClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findPermissionByARN(ctx, conn, arn)

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *permissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state permissionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RAMClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		arn := fwflex.StringValueFromFramework(ctx, plan.ARN)
		if err := prunePermissionVersions(ctx, conn, arn); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		input := ram.CreatePermissionVersionInput{
			ClientToken:    aws.String(sdkid.UniqueId()),
			PermissionArn:  aws.String(arn),
			PolicyTemplate: fwflex.StringFromFramework(ctx, plan.PolicyTemplate),
		}
		out, err := conn.CreatePermissionVersion(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out.Permission, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.DefaultVersion = state.DefaultVersion
		plan.Status = state.Status
		plan.Version = state.Version
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *permissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state permissionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RAMClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	input := ram.DeletePermissionInput{
		PermissionArn: aws.String(arn),
	}
	_, err := conn.DeletePermission(ctx, &input)
	if errs.IsA[*awstypes.UnknownResourceException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	if _, err := waitPermissionDeleted(ctx, conn, arn, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

// prunePermissionVersions deletes the oldest version.
//
// Old versions are deleted until there are 4 or less remaining, which means at
// least one more can be created before hitting the maximum of 5.
//
// The default version is never deleted.
func prunePermissionVersions(ctx context.Context, conn *ram.Client, arn string) error {
	versions, err := findPermissionVersionsByARN(ctx, conn, arn)

	if err != nil {
		return err
	}

	if len(versions) < 5 {
		return nil
	}

	oldestVersion := versions[0]
	for _, version := range versions {
		if aws.ToBool(version.DefaultVersion) {
			continue
		}

		if version.CreationTime.Before(aws.ToTime(oldestVersion.CreationTime)) {
			oldestVersion = version
		}
	}

	return deletePermissionVersion(ctx, conn, arn, intflex.StringToInt32Value(oldestVersion.Version))
}

func deletePermissionVersion(ctx context.Context, conn *ram.Client, arn string, versionID int32) error {
	input := ram.DeletePermissionVersionInput{
		PermissionArn:     aws.String(arn),
		PermissionVersion: aws.Int32(versionID),
	}
	_, err := conn.DeletePermissionVersion(ctx, &input)

	if err != nil {
		return fmt.Errorf("deleting RAM Permission (%s) version (%d): %w", arn, versionID, err)
	}

	return nil
}

func findPermissionVersionsByARN(ctx context.Context, conn *ram.Client, arn string) ([]awstypes.ResourceSharePermissionSummary, error) {
	input := ram.ListPermissionVersionsInput{
		PermissionArn: aws.String(arn),
	}

	return findPermissionVersions(ctx, conn, &input)
}

func findPermissionVersions(ctx context.Context, conn *ram.Client, input *ram.ListPermissionVersionsInput) ([]awstypes.ResourceSharePermissionSummary, error) {
	var output []awstypes.ResourceSharePermissionSummary

	pages := ram.NewListPermissionVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.UnknownResourceException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Permissions {
			if !inttypes.IsZero(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func waitPermissionDeleted(ctx context.Context, conn *ram.Client, arn string, timeout time.Duration) (*awstypes.ResourceSharePermissionDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PermissionStatusDeleting),
		Target:  []string{},
		Refresh: statusPermission(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResourceSharePermissionDetail); ok {
		return output, err
	}

	return nil, err
}

func statusPermission(conn *ram.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findPermissionByARN(ctx, conn, arn)

		if retry.NotFound(err) {
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

	output, err := findPermission(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.PermissionStatusDeleted {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, nil
}

func findPermission(ctx context.Context, conn *ram.Client, input *ram.GetPermissionInput) (*awstypes.ResourceSharePermissionDetail, error) {
	output, err := conn.GetPermission(ctx, input)

	if errs.IsA[*awstypes.UnknownResourceException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Permission == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Permission, nil
}

type permissionResourceModel struct {
	framework.WithRegionModel
	ARN            types.String   `tfsdk:"arn"`
	DefaultVersion types.Bool     `tfsdk:"default_version"`
	Name           types.String   `tfsdk:"name"`
	PolicyTemplate types.String   `tfsdk:"policy_template"`
	ResourceType   types.String   `tfsdk:"resource_type"`
	Status         types.String   `tfsdk:"status"`
	Tags           tftags.Map     `tfsdk:"tags"`
	TagsAll        tftags.Map     `tfsdk:"tags_all"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
	Version        types.String   `tfsdk:"version"`
}
