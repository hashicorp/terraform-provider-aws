// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"context"
	"time"

	r53rcc "github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	timeout    = 60 * time.Second
	minTimeout = 5 * time.Second
)

func waitClusterCreated(ctx context.Context, conn *r53rcc.Client, clusterArn string) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.StatusPending),
		Target:     enum.Slice(awstypes.StatusDeployed),
		Refresh:    statusCluster(ctx, conn, clusterArn),
		Timeout:    timeout,
		MinTimeout: minTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *r53rcc.Client, clusterArn string) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.StatusPendingDeletion),
		Target:         []string{},
		Refresh:        statusCluster(ctx, conn, clusterArn),
		Timeout:        timeout,
		Delay:          minTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitRoutingControlCreated(ctx context.Context, conn *r53rcc.Client, routingControlArn string) (*awstypes.RoutingControl, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.StatusPending),
		Target:     enum.Slice(awstypes.StatusDeployed),
		Refresh:    statusRoutingControl(ctx, conn, routingControlArn),
		Timeout:    timeout,
		MinTimeout: minTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.RoutingControl); ok {
		return output, err
	}

	return nil, err
}

func waitRoutingControlDeleted(ctx context.Context, conn *r53rcc.Client, routingControlArn string) (*awstypes.RoutingControl, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.StatusPendingDeletion),
		Target:         []string{},
		Refresh:        statusRoutingControl(ctx, conn, routingControlArn),
		Timeout:        timeout,
		Delay:          minTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.RoutingControl); ok {
		return output, err
	}

	return nil, err
}

func waitControlPanelCreated(ctx context.Context, conn *r53rcc.Client, controlPanelArn string) (*awstypes.ControlPanel, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.StatusPending),
		Target:     enum.Slice(awstypes.StatusDeployed),
		Refresh:    statusControlPanel(ctx, conn, controlPanelArn),
		Timeout:    timeout,
		MinTimeout: minTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ControlPanel); ok {
		return output, err
	}

	return nil, err
}

func waitControlPanelDeleted(ctx context.Context, conn *r53rcc.Client, controlPanelArn string) (*awstypes.ControlPanel, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.StatusPendingDeletion),
		Target:         []string{},
		Refresh:        statusControlPanel(ctx, conn, controlPanelArn),
		Timeout:        timeout,
		Delay:          minTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ControlPanel); ok {
		return output, err
	}

	return nil, err
}

func waitSafetyRuleCreated(ctx context.Context, conn *r53rcc.Client, safetyRuleArn string) (*r53rcc.DescribeSafetyRuleOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.StatusPending),
		Target:     enum.Slice(awstypes.StatusDeployed),
		Refresh:    statusSafetyRule(ctx, conn, safetyRuleArn),
		Timeout:    timeout,
		MinTimeout: minTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*r53rcc.DescribeSafetyRuleOutput); ok {
		return output, err
	}

	return nil, err
}

func waitSafetyRuleDeleted(ctx context.Context, conn *r53rcc.Client, safetyRuleArn string) (*r53rcc.DescribeSafetyRuleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.StatusPendingDeletion),
		Target:         []string{},
		Refresh:        statusSafetyRule(ctx, conn, safetyRuleArn),
		Timeout:        timeout,
		Delay:          minTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*r53rcc.DescribeSafetyRuleOutput); ok {
		return output, err
	}

	return nil, err
}
