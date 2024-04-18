// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"context"
	"time"

	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	timeout    = 60 * time.Second
	minTimeout = 5 * time.Second
)

func waitClusterCreated(ctx context.Context, conn *r53rcc.Route53RecoveryControlConfig, clusterArn string) (*r53rcc.DescribeClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    statusCluster(ctx, conn, clusterArn),
		Timeout:    timeout,
		MinTimeout: minTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*r53rcc.DescribeClusterOutput); ok {
		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *r53rcc.Route53RecoveryControlConfig, clusterArn string) (*r53rcc.DescribeClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        statusCluster(ctx, conn, clusterArn),
		Timeout:        timeout,
		Delay:          minTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*r53rcc.DescribeClusterOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRoutingControlCreated(ctx context.Context, conn *r53rcc.Route53RecoveryControlConfig, routingControlArn string) (*r53rcc.DescribeRoutingControlOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    statusRoutingControl(ctx, conn, routingControlArn),
		Timeout:    timeout,
		MinTimeout: minTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*r53rcc.DescribeRoutingControlOutput); ok {
		return output, err
	}

	return nil, err
}

func waitRoutingControlDeleted(ctx context.Context, conn *r53rcc.Route53RecoveryControlConfig, routingControlArn string) (*r53rcc.DescribeRoutingControlOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        statusRoutingControl(ctx, conn, routingControlArn),
		Timeout:        timeout,
		Delay:          minTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*r53rcc.DescribeRoutingControlOutput); ok {
		return output, err
	}

	return nil, err
}

func waitControlPanelCreated(ctx context.Context, conn *r53rcc.Route53RecoveryControlConfig, controlPanelArn string) (*r53rcc.DescribeControlPanelOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
		Refresh:    statusControlPanel(ctx, conn, controlPanelArn),
		Timeout:    timeout,
		MinTimeout: minTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*r53rcc.DescribeControlPanelOutput); ok {
		return output, err
	}

	return nil, err
}

func waitControlPanelDeleted(ctx context.Context, conn *r53rcc.Route53RecoveryControlConfig, controlPanelArn string) (*r53rcc.DescribeControlPanelOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
		Target:         []string{},
		Refresh:        statusControlPanel(ctx, conn, controlPanelArn),
		Timeout:        timeout,
		Delay:          minTimeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*r53rcc.DescribeControlPanelOutput); ok {
		return output, err
	}

	return nil, err
}

func waitSafetyRuleCreated(ctx context.Context, conn *r53rcc.Route53RecoveryControlConfig, safetyRuleArn string) (*r53rcc.DescribeSafetyRuleOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    []string{r53rcc.StatusPending},
		Target:     []string{r53rcc.StatusDeployed},
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

func waitSafetyRuleDeleted(ctx context.Context, conn *r53rcc.Route53RecoveryControlConfig, safetyRuleArn string) (*r53rcc.DescribeSafetyRuleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{r53rcc.StatusPendingDeletion},
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
