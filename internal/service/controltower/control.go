// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package controltower

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/aws/aws-sdk-go-v2/service/controltower/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_controltower_control", name="Control")
func resourceControl() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceControlCreate,
		ReadWithoutTimeout:   resourceControlRead,
		DeleteWithoutTimeout: resourceControlDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"control_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"target_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceControlCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	controlIdentifier := d.Get("control_identifier").(string)
	targetIdentifier := d.Get("target_identifier").(string)
	id := errs.Must(flex.FlattenResourceId([]string{targetIdentifier, controlIdentifier}, controlResourceIDPartCount, false))
	input := &controltower.EnableControlInput{
		ControlIdentifier: aws.String(controlIdentifier),
		TargetIdentifier:  aws.String(targetIdentifier),
	}

	output, err := conn.EnableControl(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ControlTower Control (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationIdentifier), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ControlTower Control (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceControlRead(ctx, d, meta)...)
}

func resourceControlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), controlResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	targetIdentifier, controlIdentifier := parts[0], parts[1]
	output, err := findEnabledControlByTwoPartKey(ctx, conn, targetIdentifier, controlIdentifier)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ControlTower Control %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ControlTower Control (%s): %s", d.Id(), err)
	}

	d.Set("control_identifier", output.ControlIdentifier)
	d.Set("target_identifier", targetIdentifier)

	return diags
}

func resourceControlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), controlResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	targetIdentifier, controlIdentifier := parts[0], parts[1]

	log.Printf("[DEBUG] Deleting ControlTower Control: %s", d.Id())
	output, err := conn.DisableControl(ctx, &controltower.DisableControlInput{
		ControlIdentifier: aws.String(controlIdentifier),
		TargetIdentifier:  aws.String(targetIdentifier),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ControlTower Control (%s): %s", d.Id(), err)
	}

	if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationIdentifier), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ControlTower Control (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const (
	controlResourceIDPartCount = 2
)

func findEnabledControlByTwoPartKey(ctx context.Context, conn *controltower.Client, targetIdentifier, controlIdentifier string) (*types.EnabledControlSummary, error) {
	input := &controltower.ListEnabledControlsInput{
		TargetIdentifier: aws.String(targetIdentifier),
	}

	return findEnabledControl(ctx, conn, input, func(v *types.EnabledControlSummary) bool {
		return aws.ToString(v.ControlIdentifier) == controlIdentifier
	})
}

func findEnabledControl(ctx context.Context, conn *controltower.Client, input *controltower.ListEnabledControlsInput, filter tfslices.Predicate[*types.EnabledControlSummary]) (*types.EnabledControlSummary, error) {
	output, err := findEnabledControls(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findEnabledControls(ctx context.Context, conn *controltower.Client, input *controltower.ListEnabledControlsInput, filter tfslices.Predicate[*types.EnabledControlSummary]) ([]*types.EnabledControlSummary, error) {
	var output []*types.EnabledControlSummary

	pages := controltower.NewListEnabledControlsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.EnabledControls {
			v := v
			if v := &v; filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findControlOperationByID(ctx context.Context, conn *controltower.Client, id string) (*types.ControlOperation, error) {
	input := &controltower.GetControlOperationInput{
		OperationIdentifier: aws.String(id),
	}

	output, err := conn.GetControlOperation(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ControlOperation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ControlOperation, nil
}

func statusControlOperation(ctx context.Context, conn *controltower.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findControlOperationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString((*string)(&output.Status)), nil
	}
}

func waitOperationSucceeded(ctx context.Context, conn *controltower.Client, id string, timeout time.Duration) (*types.ControlOperation, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ControlOperationStatusInProgress),
		Target:  enum.Slice(types.ControlOperationStatusSucceeded),
		Refresh: statusControlOperation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*controltower.GetControlOperationOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ControlOperation.StatusMessage)))

		return output.ControlOperation, err
	}

	return nil, err
}
