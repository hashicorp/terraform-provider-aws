// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/action/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/actionvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/actionwait"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwactions "github.com/hashicorp/terraform-provider-aws/internal/framework/actions"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action("aws_elasticache_apply_service_update", name="Apply Service Update")
func newApplyServiceUpdateAction(_ context.Context) (action.ActionWithConfigure, error) {
	var r applyServiceUpdateAction
	r.SetDefaultInvokeTimeout(1 * time.Hour)

	return &r, nil
}

type applyServiceUpdateAction struct {
	framework.ActionWithModel[applyServiceUpdateModel]
	framework.ActionWithTimeouts
}

func (a *applyServiceUpdateAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cache_cluster_id": schema.StringAttribute{
				Optional: true,
			},
			"replication_group_id": schema.StringAttribute{
				Optional: true,
			},
			"service_update_name": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx),
		},
	}
}

func (a *applyServiceUpdateAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config applyServiceUpdateModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout := a.InvokeTimeout(ctx, config.Timeouts)

	var input elasticache.BatchApplyUpdateActionInput
	resp.Diagnostics.Append(flex.Expand(ctx, config, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.CacheClusterIds = flex.StringSliceValueFromFramework(ctx, config.CacheClusterID)
	input.ReplicationGroupIds = flex.StringSliceValueFromFramework(ctx, config.ReplicationGroupID)

	var (
		targetType string
		targetKey  string
		targetID   string
	)
	if !config.CacheClusterID.IsNull() {
		targetType = "Cache Cluster"
		targetKey = "cache_cluster_id"
		targetID = config.CacheClusterID.ValueString()
	} else if !config.ReplicationGroupID.IsNull() {
		targetType = "Replication Group"
		targetKey = "replication_group_id"
		targetID = config.ReplicationGroupID.ValueString()
	}

	ctx = tflog.SetField(ctx, "service_update_name", config.ServiceUpdateName.ValueString())
	ctx = tflog.SetField(ctx, targetKey, targetID)

	tflog.Info(ctx, "Applying ElastiCache service update")

	// Send initial progress update
	cb := fwactions.NewSendProgressFunc(resp)
	cb(ctx, "Applying ElastiCache service update %q for %s %q...", config.ServiceUpdateName.ValueString(), targetType, targetID)

	conn := a.Meta().ElastiCacheClient(ctx)

	output, err := conn.BatchApplyUpdateAction(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Apply Service Update",
			fmt.Sprintf("Could not apply ElastiCache service update %q for %s %q: %s", config.ServiceUpdateName.ValueString(), targetType, targetID, err),
		)
		return
	}
	if len(output.UnprocessedUpdateActions) > 0 {
		uprocessedAction := output.UnprocessedUpdateActions[0]
		resp.Diagnostics.AddError(
			"Failed to Apply Service Update",
			fmt.Sprintf("Could not apply ElastiCache service update %q for %s %q.\n\n", config.ServiceUpdateName.ValueString(), targetType, targetID)+
				fmt.Sprintf("Error Type: %s\nError Message: %s", aws.ToString(uprocessedAction.ErrorType), aws.ToString(uprocessedAction.ErrorMessage)),
		)
		return
	}

	cb(ctx, "ElastiCache service update %q for %s %q started, waiting for completion...", config.ServiceUpdateName.ValueString(), targetType, targetID)

	_, err = actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[*awstypes.UpdateAction], error) {
		input := elasticache.DescribeUpdateActionsInput{
			ServiceUpdateName: config.ServiceUpdateName.ValueStringPointer(),
		}
		input.CacheClusterIds = flex.StringSliceValueFromFramework(ctx, config.CacheClusterID)
		input.ReplicationGroupIds = flex.StringSliceValueFromFramework(ctx, config.ReplicationGroupID)

		output, err := findServiceUpdateAction(ctx, conn, &input)
		if err != nil {
			return actionwait.FetchResult[*awstypes.UpdateAction]{}, fmt.Errorf("getting update status: %w", err)
		}
		return actionwait.FetchResult[*awstypes.UpdateAction]{
			Status: actionwait.Status(output.UpdateActionStatus),
		}, nil
	}, actionwait.Options[*awstypes.UpdateAction]{
		Timeout:          timeout,
		Interval:         actionwait.FixedInterval(actionwait.DefaultPollInterval),
		ProgressInterval: 60 * time.Second,
		SuccessStates:    []actionwait.Status{actionwait.Status(awstypes.UpdateActionStatusComplete)},
		TransitionalStates: []actionwait.Status{
			actionwait.Status(awstypes.UpdateActionStatusWaitingToStart),
			actionwait.Status(awstypes.UpdateActionStatusInProgress),
		},
		ProgressSink: func(fr actionwait.FetchResult[any], meta actionwait.ProgressMeta) {
			cb(ctx, "ElastiCache service update %q for %s %q is currently %q, continuing to wait for completion...", config.ServiceUpdateName.ValueString(), targetType, targetID, fr.Status)
		},
	})
	if err != nil {
		if errs.IsA[*actionwait.TimeoutError](err) {
			resp.Diagnostics.AddError(
				"Timeout Waiting for Service Update to Complete",
				fmt.Sprintf("ElastiCache service update %q for %s %q did not complete within %s: %s", config.ServiceUpdateName.ValueString(), targetType, targetID, timeout, err),
			)
		} else if errs.IsA[*actionwait.UnexpectedStateError](err) {
			resp.Diagnostics.AddError(
				"Unexpected Service Update State",
				fmt.Sprintf("ElastiCache service update %q for %s %q entered unexpected state: %s", config.ServiceUpdateName.ValueString(), targetType, targetID, err),
			)
		} else {
			resp.Diagnostics.AddError(
				"Failed While Waiting for Service Update to Complete",
				fmt.Sprintf("ElastiCache service update %q for %s %q: %s", config.ServiceUpdateName.ValueString(), targetType, targetID, err),
			)
		}
		return
	}

	// Final success message
	cb(ctx, "ElastiCache service update %q for %s %q applied successfully", config.ServiceUpdateName.ValueString(), targetType, targetID)

	tflog.Info(ctx, "ElastiCache service update applied successfully")
}

func (a *applyServiceUpdateAction) ConfigValidators(context.Context) []action.ConfigValidator {
	return []action.ConfigValidator{
		actionvalidator.Conflicting(
			path.MatchRoot("cache_cluster_id"),
			path.MatchRoot("replication_group_id"),
		),
	}
}

type applyServiceUpdateModel struct {
	framework.WithRegionModel
	CacheClusterID     types.String   `tfsdk:"cache_cluster_id" autoflex:"-"`
	ReplicationGroupID types.String   `tfsdk:"replication_group_id" autoflex:"-"`
	ServiceUpdateName  types.String   `tfsdk:"service_update_name"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}
