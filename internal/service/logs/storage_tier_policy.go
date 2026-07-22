// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_log_storage_tier_policy", name="Storage Tier Policy")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(serialize=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs;cloudwatchlogs.GetStorageTierPolicyOutput")
func newStorageTierPolicyResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &storageTierPolicyResource{}

	return r, nil
}

type storageTierPolicyResource struct {
	framework.ResourceWithModel[storageTierPolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *storageTierPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CloudWatch Logs account-level storage tier policy. When set to `INTELLIGENT_TIERING`, CloudWatch Logs automatically moves log data to the most cost-effective storage tier based on access frequency.",
		Attributes: map[string]schema.Attribute{
			"storage_tier": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.StorageTier](),
				Required:    true,
				Description: "The storage tier to set for the account. Valid values are `STANDARD` or `INTELLIGENT_TIERING`.",
			},
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
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

type storageTierPolicyResourceModel struct {
	StorageTier fwtypes.StringEnum[awstypes.StorageTier] `tfsdk:"storage_tier"`
	ID          types.String                             `tfsdk:"id"`
	Region      types.String                             `tfsdk:"region"`
	Timeouts    timeouts.Value                           `tfsdk:"timeouts"`
}

func (r *storageTierPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LogsClient(ctx)

	var plan storageTierPolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input cloudwatchlogs.PutStorageTierPolicyInput
	input.StorageTier = plan.StorageTier.ValueEnum()

	_, err := conn.PutStorageTierPolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "creating CloudWatch Logs Storage Tier Policy")
		return
	}

	plan.ID = types.StringValue(r.Meta().Region(ctx))
	plan.Region = types.StringValue(r.Meta().Region(ctx))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *storageTierPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LogsClient(ctx)

	var state storageTierPolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findStorageTierPolicy(ctx, conn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "reading CloudWatch Logs Storage Tier Policy")
		return
	}

	state.StorageTier = fwtypes.StringEnumValue(out.StorageTier)
	state.Region = types.StringValue(r.Meta().Region(ctx))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *storageTierPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LogsClient(ctx)

	var plan storageTierPolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input cloudwatchlogs.PutStorageTierPolicyInput
	input.StorageTier = plan.StorageTier.ValueEnum()

	_, err := conn.PutStorageTierPolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "updating CloudWatch Logs Storage Tier Policy")
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *storageTierPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LogsClient(ctx)

	var state storageTierPolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// Reset to STANDARD (default state) since there's no dedicated delete API
	var input cloudwatchlogs.PutStorageTierPolicyInput
	input.StorageTier = awstypes.StorageTierStandard

	_, err := conn.PutStorageTierPolicy(ctx, &input)
	if err != nil && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
		smerr.AddError(ctx, &resp.Diagnostics, err, "deleting CloudWatch Logs Storage Tier Policy")
		return
	}
}

func findStorageTierPolicy(ctx context.Context, conn *cloudwatchlogs.Client) (*cloudwatchlogs.GetStorageTierPolicyOutput, error) {
	input := &cloudwatchlogs.GetStorageTierPolicyInput{}

	out, err := conn.GetStorageTierPolicy(ctx, input)
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}
