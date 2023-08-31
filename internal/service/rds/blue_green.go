// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"
	"time"

	rds_sdkv2 "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

type cleanupWaiterFunc func(context.Context, ...tfresource.OptionsFunc) //nolint:unused // WIP

type cleanupWaiterErrFunc func(context.Context, ...tfresource.OptionsFunc) error //nolint:unused // WIP

type blueGreenOrchestrator struct {
	conn           *rds_sdkv2.Client
	cleanupWaiters []cleanupWaiterFunc //nolint:unused // WIP
}

func newBlueGreenOrchestrator(conn *rds_sdkv2.Client) *blueGreenOrchestrator {
	return &blueGreenOrchestrator{
		conn: conn,
	}
}

func (o *blueGreenOrchestrator) cleanUp(ctx context.Context) { //nolint:unused // WIP
	if len(o.cleanupWaiters) == 0 {
		return
	}

	waiter, waiters := o.cleanupWaiters[0], o.cleanupWaiters[1:]
	waiter(ctx)
	for _, waiter := range waiters {
		// Skip the delay for subsequent waiters. Since we're waiting for all of the waiters
		// to complete, we don't need to run them concurrently, saving on network traffic.
		waiter(ctx, tfresource.WithDelay(0))
	}
}

func (o *blueGreenOrchestrator) createDeployment(ctx context.Context, input *rds_sdkv2.CreateBlueGreenDeploymentInput) (*types.BlueGreenDeployment, error) {
	createOut, err := o.conn.CreateBlueGreenDeployment(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("creating Blue/Green Deployment: %s", err)
	}
	dep := createOut.BlueGreenDeployment
	return dep, nil
}

func (o *blueGreenOrchestrator) waitForDeploymentAvailable(ctx context.Context, identifier string, timeout time.Duration) (*types.BlueGreenDeployment, error) {
	dep, err := waitBlueGreenDeploymentAvailable(ctx, o.conn, identifier, timeout)
	if err != nil {
		return nil, fmt.Errorf("creating Blue/Green Deployment: %s", err)
	}
	return dep, nil
}

func (o *blueGreenOrchestrator) switchover(ctx context.Context, identifier string, timeout time.Duration) (*types.BlueGreenDeployment, error) {
	input := &rds_sdkv2.SwitchoverBlueGreenDeploymentInput{
		BlueGreenDeploymentIdentifier: aws.String(identifier),
	}
	_, err := tfresource.RetryWhen(ctx, 10*time.Minute,
		func() (interface{}, error) {
			return o.conn.SwitchoverBlueGreenDeployment(ctx, input)
		},
		func(err error) (bool, error) {
			return errs.IsA[*types.InvalidBlueGreenDeploymentStateFault](err), err
		},
	)
	if err != nil {
		return nil, fmt.Errorf("switching over Blue/Green Deployment: %s", err)
	}

	dep, err := waitBlueGreenDeploymentSwitchoverCompleted(ctx, o.conn, identifier, timeout)
	if err != nil {
		return nil, fmt.Errorf("switching over Blue/Green Deployment: waiting for completion: %s", err)
	}
	return dep, nil
}

type instanceHandler struct {
	conn *rds_sdkv2.Client
}

func newInstanceHandler(conn *rds_sdkv2.Client) *instanceHandler {
	return &instanceHandler{
		conn: conn,
	}
}

func (h *instanceHandler) precondition(ctx context.Context, d *schema.ResourceData) error {
	needsPreConditions := false
	input := &rds_sdkv2.ModifyDBInstanceInput{
		ApplyImmediately:     true,
		DBInstanceIdentifier: aws.String(d.Get("identifier").(string)),
	}

	// Backups must be enabled for Blue/Green Deployments. Enable them first.
	o, n := d.GetChange("backup_retention_period")
	if o.(int) == 0 && n.(int) > 0 {
		needsPreConditions = true
		input.BackupRetentionPeriod = aws.Int32(int32(d.Get("backup_retention_period").(int)))
	}

	if d.HasChange("deletion_protection") {
		needsPreConditions = true
		input.DeletionProtection = aws.Bool(d.Get("deletion_protection").(bool))
	}

	if needsPreConditions {
		err := dbInstanceModify(ctx, h.conn, d.Id(), input, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("setting pre-conditions: %s", err)
		}
	}
	return nil
}

func (h *instanceHandler) createBlueGreenInput(d *schema.ResourceData) *rds_sdkv2.CreateBlueGreenDeploymentInput {
	input := &rds_sdkv2.CreateBlueGreenDeploymentInput{
		BlueGreenDeploymentName: aws.String(d.Get("identifier").(string)),
		Source:                  aws.String(d.Get("arn").(string)),
	}

	if d.HasChange("engine_version") {
		input.TargetEngineVersion = aws.String(d.Get("engine_version").(string))
	}
	if d.HasChange("parameter_group_name") {
		input.TargetDBParameterGroupName = aws.String(d.Get("parameter_group_name").(string))
	}

	return input
}

func (h *instanceHandler) modifyTarget(ctx context.Context, identifier string, d *schema.ResourceData, timeout time.Duration, operation string) error {
	modifyInput := &rds_sdkv2.ModifyDBInstanceInput{
		ApplyImmediately:     true,
		DBInstanceIdentifier: aws.String(identifier),
	}

	needsModify := dbInstancePopulateModify(modifyInput, d)

	if needsModify {
		log.Printf("[DEBUG] %s: Updating Green environment", operation)

		err := dbInstanceModify(ctx, h.conn, d.Id(), modifyInput, timeout)
		if err != nil {
			return fmt.Errorf("updating Green environment: %s", err)
		}
	}

	return nil
}
