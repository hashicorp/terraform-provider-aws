// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	"github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_prometheus_alert_manager_definition", name="Alert Manager Definition")
func resourceAlertManagerDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAlertManagerDefinitionCreate,
		ReadWithoutTimeout:   resourceAlertManagerDefinitionRead,
		UpdateWithoutTimeout: resourceAlertManagerDefinitionUpdate,
		DeleteWithoutTimeout: resourceAlertManagerDefinitionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"definition": {
				Type:     schema.TypeString,
				Required: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAlertManagerDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AMPClient(ctx)

	workspaceID := d.Get("workspace_id").(string)
	input := &amp.CreateAlertManagerDefinitionInput{
		Data:        []byte(d.Get("definition").(string)),
		WorkspaceId: aws.String(workspaceID),
	}

	_, err := conn.CreateAlertManagerDefinition(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Prometheus Alert Manager Definition (%s): %s", workspaceID, err)
	}

	d.SetId(workspaceID)

	if _, err := waitAlertManagerDefinitionCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Alert Manager Definition (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceAlertManagerDefinitionRead(ctx, d, meta)...)
}

func resourceAlertManagerDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AMPClient(ctx)

	amd, err := findAlertManagerDefinitionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Prometheus Alert Manager Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Prometheus Alert Manager Definition (%s): %s", d.Id(), err)
	}

	d.Set("definition", string(amd.Data))
	d.Set("workspace_id", d.Id())

	return diags
}

func resourceAlertManagerDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AMPClient(ctx)

	input := &amp.PutAlertManagerDefinitionInput{
		Data:        []byte(d.Get("definition").(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	}

	_, err := conn.PutAlertManagerDefinition(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Prometheus Alert Manager Definition (%s): %s", d.Id(), err)
	}

	if _, err := waitAlertManagerDefinitionUpdated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Alert Manager Definition (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceAlertManagerDefinitionRead(ctx, d, meta)...)
}

func resourceAlertManagerDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AMPClient(ctx)

	log.Printf("[DEBUG] Deleting Prometheus Alert Manager Definition: (%s)", d.Id())
	_, err := conn.DeleteAlertManagerDefinition(ctx, &amp.DeleteAlertManagerDefinitionInput{
		WorkspaceId: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Prometheus Alert Manager Definition (%s): %s", d.Id(), err)
	}

	if _, err := waitAlertManagerDefinitionDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Alert Manager Definition (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findAlertManagerDefinitionByID(ctx context.Context, conn *amp.Client, id string) (*types.AlertManagerDefinitionDescription, error) {
	input := &amp.DescribeAlertManagerDefinitionInput{
		WorkspaceId: aws.String(id),
	}

	output, err := conn.DescribeAlertManagerDefinition(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AlertManagerDefinition == nil || output.AlertManagerDefinition.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AlertManagerDefinition, nil
}

func statusAlertManagerDefinition(ctx context.Context, conn *amp.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAlertManagerDefinitionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.StatusCode), nil
	}
}

func waitAlertManagerDefinitionCreated(ctx context.Context, conn *amp.Client, id string) (*types.AlertManagerDefinitionDescription, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.AlertManagerDefinitionStatusCodeCreating),
		Target:  enum.Slice(types.AlertManagerDefinitionStatusCodeActive),
		Refresh: statusAlertManagerDefinition(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AlertManagerDefinitionDescription); ok {
		if statusCode := output.Status.StatusCode; statusCode == types.AlertManagerDefinitionStatusCodeCreationFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitAlertManagerDefinitionUpdated(ctx context.Context, conn *amp.Client, id string) (*types.AlertManagerDefinitionDescription, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.AlertManagerDefinitionStatusCodeUpdating),
		Target:  enum.Slice(types.AlertManagerDefinitionStatusCodeActive),
		Refresh: statusAlertManagerDefinition(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AlertManagerDefinitionDescription); ok {
		if statusCode := output.Status.StatusCode; statusCode == types.AlertManagerDefinitionStatusCodeUpdateFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitAlertManagerDefinitionDeleted(ctx context.Context, conn *amp.Client, id string) (*types.AlertManagerDefinitionDescription, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.AlertManagerDefinitionStatusCodeDeleting),
		Target:  []string{},
		Refresh: statusAlertManagerDefinition(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AlertManagerDefinitionDescription); ok {
		return output, err
	}

	return nil, err
}
