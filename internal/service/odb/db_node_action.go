// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/actionwait"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwactions "github.com/hashicorp/terraform-provider-aws/internal/framework/actions"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const dbNodeActionPollInterval = 15 * time.Second

// @Action(aws_odb_start_db_node, name="Start DB Node")
func newStartDBNodeAction(context.Context) (action.ActionWithConfigure, error) {
	return &dbNodeAction{operation: "start"}, nil
}

// @Action(aws_odb_stop_db_node, name="Stop DB Node")
func newStopDBNodeAction(context.Context) (action.ActionWithConfigure, error) {
	return &dbNodeAction{operation: "stop"}, nil
}

// @Action(aws_odb_reboot_db_node, name="Reboot DB Node")
func newRebootDBNodeAction(context.Context) (action.ActionWithConfigure, error) {
	return &dbNodeAction{operation: "reboot"}, nil
}

type dbNodeAction struct {
	framework.ActionWithModel[dbNodeActionModel]
	operation string
}

type dbNodeActionModel struct {
	framework.WithRegionModel
	CloudVMClusterID types.String `tfsdk:"cloud_vm_cluster_id"`
	DBNodeID         types.String `tfsdk:"db_node_id"`
	Timeout          types.Int64  `tfsdk:"timeout"`
}

func (a *dbNodeAction) Schema(ctx context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("Performs the %s operation on an existing Oracle Database@AWS DB node and waits for completion.", a.operation),
		Attributes: map[string]schema.Attribute{
			"cloud_vm_cluster_id": schema.StringAttribute{
				Description: "Identifier of the cloud VM cluster containing the DB node.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 64),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_~.-]+$`), "must contain only letters, numbers, underscores, tildes, periods, and hyphens"),
				},
			},
			"db_node_id": schema.StringAttribute{
				Description: "Identifier of the DB node to operate on.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 64),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_~.-]+$`), "must contain only letters, numbers, underscores, tildes, periods, and hyphens"),
				},
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds for the entire operation, including status checks. Defaults to 3600. Must be between 1 and 86400.",
				Optional:    true,
				Validators:  []validator.Int64{int64validator.Between(1, 86400)},
			},
		},
	}
}

func (a *dbNodeAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data dbNodeActionModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}
	if data.CloudVMClusterID.IsUnknown() || data.DBNodeID.IsUnknown() || data.Timeout.IsUnknown() {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("db node action inputs must be known before invocation"))
		return
	}

	// Values unknown during planning must be validated before duration conversion can overflow.
	if timeout := data.Timeout.ValueInt64(); !data.Timeout.IsNull() && (timeout < 1 || timeout > 86400) {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("db node action timeout must be between 1 and 86400 seconds"))
		return
	}

	err := runDBNodeAction(ctx, a.Meta().ODBClient(ctx), a.operation, data.CloudVMClusterID.ValueString(), data.DBNodeID.ValueString(), fwactions.TimeoutOr(data.Timeout, time.Hour), dbNodeActionPollInterval, fwactions.NewSendProgressFunc(resp))
	if err != nil {
		tflog.Error(ctx, "ODB DB node action did not complete", map[string]any{"operation": a.operation, "db_node_id": data.DBNodeID.ValueString()})
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.DBNodeID.ValueString())
	}
}

func runDBNodeAction(ctx context.Context, conn *odb.Client, operation, clusterID, nodeID string, timeout, pollInterval time.Duration, progress fwactions.SendProgressFunc) error {
	if !slices.Contains([]string{"start", "stop", "reboot"}, operation) {
		return fmt.Errorf("unsupported DB node operation %q", operation)
	}
	for name, value := range map[string]string{"cloud_vm_cluster_id": clusterID, "db_node_id": nodeID} {
		if len(value) < 6 || len(value) > 64 || !regexache.MustCompile(`^[a-zA-Z0-9_~.-]+$`).MatchString(value) {
			return fmt.Errorf("invalid %s: must be 6 to 64 letters, numbers, underscores, tildes, periods, or hyphens", name)
		}
	}
	if timeout <= 0 || timeout > 24*time.Hour {
		return errors.New("db node action timeout must be positive and at most 24 hours")
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ctx = tflog.SetField(ctx, "operation", operation)
	ctx = tflog.SetField(ctx, "cloud_vm_cluster_id", clusterID)
	ctx = tflog.SetField(ctx, "db_node_id", nodeID)
	tflog.Info(ctx, "Checking ODB DB node action eligibility")
	progress(ctx, "Checking DB node %s before %s...", nodeID, operation)

	node, err := findDBNodeForAction(ctx, conn, clusterID, nodeID)
	if err != nil {
		return fmt.Errorf("reading DB node before %s: %w", operation, err)
	}
	initial := node.Status
	target := awstypes.DbNodeResourceStatusAvailable
	required := awstypes.DbNodeResourceStatusStopped
	pending := []actionwait.Status{"STOPPED", "STARTING"}
	if operation == "stop" {
		target, required = awstypes.DbNodeResourceStatusStopped, awstypes.DbNodeResourceStatusAvailable
		pending = []actionwait.Status{"AVAILABLE", "STOPPING"}
	}
	if operation == "reboot" {
		required = awstypes.DbNodeResourceStatusAvailable
		pending = []actionwait.Status{"UPDATING", "STOPPING", "STOPPED", "STARTING", "WAITING_FOR_REBOOT"}
	}
	if operation != "reboot" && initial == target {
		tflog.Info(ctx, "ODB DB node is already in the requested state", map[string]any{names.AttrStatus: string(initial)})
		progress(ctx, "DB node %s is already %s; no operation was sent.", nodeID, target)
		return nil
	}
	if initial != required {
		return fmt.Errorf("cannot %s DB node %s in state %s; required state is %s", operation, nodeID, initial, required)
	}

	progress(ctx, "Submitting DB node %s operation...", operation)
	tflog.Info(ctx, "Submitting ODB DB node action")
	// These APIs have no idempotency token. Retrying an ambiguous response can
	// repeat a reboot, so only read requests retain the provider's retry policy.
	noRetry := func(options *odb.Options) { options.RetryMaxAttempts = 1 }
	var status awstypes.DbNodeResourceStatus
	var returnedID string
	switch operation {
	case "start":
		out, err := conn.StartDbNode(ctx, &odb.StartDbNodeInput{CloudVmClusterId: aws.String(clusterID), DbNodeId: aws.String(nodeID)}, noRetry)
		if err != nil {
			return fmt.Errorf("starting DB node %s (request was not retried; check node status before invoking again): %w", nodeID, err)
		}
		status, returnedID = out.Status, aws.ToString(out.DbNodeId)
	case "stop":
		out, err := conn.StopDbNode(ctx, &odb.StopDbNodeInput{CloudVmClusterId: aws.String(clusterID), DbNodeId: aws.String(nodeID)}, noRetry)
		if err != nil {
			return fmt.Errorf("stopping DB node %s (request was not retried; check node status before invoking again): %w", nodeID, err)
		}
		status, returnedID = out.Status, aws.ToString(out.DbNodeId)
	case "reboot":
		out, err := conn.RebootDbNode(ctx, &odb.RebootDbNodeInput{CloudVmClusterId: aws.String(clusterID), DbNodeId: aws.String(nodeID)}, noRetry)
		if err != nil {
			return fmt.Errorf("rebooting DB node %s (request was not retried; check node status before invoking again): %w", nodeID, err)
		}
		status, returnedID = out.Status, aws.ToString(out.DbNodeId)
	}
	if returnedID != nodeID {
		return fmt.Errorf("db node %s response did not identify the requested node; check node status before invoking again", operation)
	}
	if status != "" && status != target && !slices.Contains(pending, actionwait.Status(status)) {
		return fmt.Errorf("db node %s returned unexpected state %s after %s", nodeID, status, operation)
	}
	tflog.Info(ctx, "ODB DB node action accepted", map[string]any{names.AttrStatus: string(status)})
	progress(ctx, "DB node %s request accepted; waiting for %s...", operation, target)

	transitionObserved := status != "" && status != awstypes.DbNodeResourceStatusAvailable
	successes := 1
	if operation == "reboot" {
		successes = 2
	}
	lastStatus := status
	_, err = actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[*awstypes.DbNode], error) {
		node, err := findDBNodeForAction(ctx, conn, clusterID, nodeID)
		if err != nil {
			return actionwait.FetchResult[*awstypes.DbNode]{}, err
		}
		lastStatus = node.Status
		tflog.Debug(ctx, "Polled ODB DB node status", map[string]any{names.AttrStatus: string(node.Status)})
		status := actionwait.Status(node.Status)
		if operation == "reboot" {
			if slices.Contains([]awstypes.DbNodeResourceStatus{awstypes.DbNodeResourceStatusUpdating, awstypes.DbNodeResourceStatusStopping, awstypes.DbNodeResourceStatusStopped, awstypes.DbNodeResourceStatusStarting}, node.Status) {
				transitionObserved = true
			}
			// AVAILABLE can be the pre-reboot state. Do not report success until
			// the operation response or a read has shown a lifecycle transition.
			if node.Status == awstypes.DbNodeResourceStatusAvailable && !transitionObserved {
				status = "WAITING_FOR_REBOOT"
			}
		}
		return actionwait.FetchResult[*awstypes.DbNode]{Status: status, Value: node}, nil
	}, actionwait.Options[*awstypes.DbNode]{
		Timeout:            timeout,
		Interval:           actionwait.FixedInterval(pollInterval),
		ProgressInterval:   30 * time.Second,
		SuccessStates:      []actionwait.Status{actionwait.Status(target)},
		TransitionalStates: pending,
		FailureStates:      []actionwait.Status{"FAILED", "TERMINATED", "TERMINATING"},
		ConsecutiveSuccess: successes,
		ProgressSink: func(_ actionwait.FetchResult[any], _ actionwait.ProgressMeta) {
			progress(ctx, "Waiting for DB node %s after %s: current state %s, target %s.", nodeID, operation, lastStatus, target)
		},
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("db node %s %s timed out after %s (last state %s); the service operation may still be running; check node status before invoking again: %w", nodeID, operation, timeout, lastStatus, err)
		}
		return fmt.Errorf("waiting for DB node %s after %s (last state %s): %w", nodeID, operation, lastStatus, err)
	}

	tflog.Info(ctx, "ODB DB node action completed", map[string]any{names.AttrStatus: string(target)})
	progress(ctx, "DB node %s %s completed successfully; node is %s.", nodeID, operation, target)
	return nil
}

func findDBNodeForAction(ctx context.Context, conn *odb.Client, clusterID, nodeID string) (*awstypes.DbNode, error) {
	out, err := conn.GetDbNode(ctx, &odb.GetDbNodeInput{CloudVmClusterId: aws.String(clusterID), DbNodeId: aws.String(nodeID)})
	if err != nil {
		return nil, fmt.Errorf("reading DB node %s: %w", nodeID, err)
	}
	if out == nil || out.DbNode == nil {
		return nil, fmt.Errorf("reading DB node %s: empty GetDbNode response", nodeID)
	}
	if aws.ToString(out.DbNode.DbNodeId) != nodeID {
		return nil, fmt.Errorf("reading DB node %s: GetDbNode returned a different node", nodeID)
	}
	return out.DbNode, nil
}
