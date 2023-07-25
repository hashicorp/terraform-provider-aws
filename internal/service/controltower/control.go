// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package controltower

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/controltower"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_controltower_control")
func ResourceControl() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).ControlTowerConn(ctx)

	controlIdentifier := d.Get("control_identifier").(string)
	targetIdentifier := d.Get("target_identifier").(string)
	id := ControlCreateResourceID(targetIdentifier, controlIdentifier)
	input := &controltower.EnableControlInput{
		ControlIdentifier: aws.String(controlIdentifier),
		TargetIdentifier:  aws.String(targetIdentifier),
	}

	output, err := conn.EnableControlWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating ControlTower Control (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitOperationSucceeded(ctx, conn, aws.StringValue(output.OperationIdentifier), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for ControlTower Control (%s) create: %s", d.Id(), err)
	}

	return resourceControlRead(ctx, d, meta)
}

func resourceControlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ControlTowerConn(ctx)

	targetIdentifier, controlIdentifier, err := ControlParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	output, err := FindEnabledControlByTwoPartKey(ctx, conn, targetIdentifier, controlIdentifier)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ControlTower Control %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading ControlTower Control (%s): %s", d.Id(), err)
	}

	d.Set("control_identifier", output.ControlIdentifier)
	d.Set("target_identifier", targetIdentifier)

	return nil
}

func resourceControlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ControlTowerConn(ctx)

	targetIdentifier, controlIdentifier, err := ControlParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting ControlTower Control: %s", d.Id())
	output, err := conn.DisableControlWithContext(ctx, &controltower.DisableControlInput{
		ControlIdentifier: aws.String(controlIdentifier),
		TargetIdentifier:  aws.String(targetIdentifier),
	})

	if err != nil {
		return diag.Errorf("deleting ControlTower Control (%s): %s", d.Id(), err)
	}

	if _, err := waitOperationSucceeded(ctx, conn, aws.StringValue(output.OperationIdentifier), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for ControlTower Control (%s) delete: %s", d.Id(), err)
	}

	return nil
}

const controlResourceIDSeparator = ","

func ControlCreateResourceID(targetIdentifier, controlIdentifier string) string {
	parts := []string{targetIdentifier, controlIdentifier}
	id := strings.Join(parts, controlResourceIDSeparator)

	return id
}

func ControlParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, controlResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TargetIdentifier%[2]sControlIdentifier", id, controlResourceIDSeparator)
}

func FindEnabledControlByTwoPartKey(ctx context.Context, conn *controltower.ControlTower, targetIdentifier, controlIdentifier string) (*controltower.EnabledControlSummary, error) {
	input := &controltower.ListEnabledControlsInput{
		TargetIdentifier: aws.String(targetIdentifier),
	}
	var output *controltower.EnabledControlSummary

	err := conn.ListEnabledControlsPagesWithContext(ctx, input, func(page *controltower.ListEnabledControlsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.EnabledControls {
			if v == nil {
				continue
			}

			if aws.StringValue(v.ControlIdentifier) == controlIdentifier {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, controltower.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findControlOperationByID(ctx context.Context, conn *controltower.ControlTower, id string) (*controltower.ControlOperation, error) {
	input := &controltower.GetControlOperationInput{
		OperationIdentifier: aws.String(id),
	}

	output, err := conn.GetControlOperationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, controltower.ErrCodeResourceNotFoundException) {
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

func statusControlOperation(ctx context.Context, conn *controltower.ControlTower, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findControlOperationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitOperationSucceeded(ctx context.Context, conn *controltower.ControlTower, id string, timeout time.Duration) (*controltower.ControlOperation, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{controltower.ControlOperationStatusInProgress},
		Target:  []string{controltower.ControlOperationStatusSucceeded},
		Refresh: statusControlOperation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*controltower.ControlOperation); ok {
		if status := aws.StringValue(output.Status); status == controltower.ControlOperationStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}
