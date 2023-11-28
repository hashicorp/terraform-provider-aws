// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package controltower

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/aws/aws-sdk-go-v2/service/controltower/document"
	"github.com/aws/aws-sdk-go-v2/service/controltower/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_controltower_landing_zone", name="Landing Zone")
func resourceLandingZone() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLandingZoneCreate,
		ReadWithoutTimeout:   resourceLandingZoneRead,
		DeleteWithoutTimeout: resourceLandingZoneDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"manifest": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceLandingZoneCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	manifest := d.Get("manifest").(string)
	version := d.Get("version").(string)
	input := &controltower.CreateLandingZoneInput{
		Manifest: document.NewLazyDocument(manifest),
		Version:  aws.String(version),
		Tags:     getTagsIn(ctx),
	}

	output, err := conn.CreateLandingZone(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Control Tower Landing Zone: %s", err)
	}

	identifier, err := getLandingZoneIdentifierFromARN(*output.Arn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting identifier from ARN: %s", err)
	}

	d.SetId(*aws.String(identifier))

	if _, err := waitLandingZoneOperationSucceeded(ctx, conn, aws.ToString(output.OperationIdentifier), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for ControlTower Landing Zone (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceLandingZoneRead(ctx, d, meta)...)
}

func resourceLandingZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	input := &controltower.GetLandingZoneInput{
		LandingZoneIdentifier: aws.String(d.Id()),
	}

	output, err := conn.GetLandingZone(ctx, input)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Control Tower Landing Zone (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Control Tower Landing Zone: %s", err)
	}

	d.Set("manifest", output.LandingZone.Manifest)
	d.Set("version", output.LandingZone.Version)
	d.Set("arn", output.LandingZone.Arn)

	return nil
}

func resourceLandingZoneDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ControlTowerClient(ctx)

	input := &controltower.DeleteLandingZoneInput{
		LandingZoneIdentifier: aws.String(d.Id()),
	}

	output, err := conn.DeleteLandingZone(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Control Tower Landing Zone: %s", err)
	}

	if _, err := waitLandingZoneOperationSucceeded(ctx, conn, aws.ToString(output.OperationIdentifier), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Control Tower Landing Zone (%s) delete: %s", d.Id(), err)
	}

	return nil
}

// gets the landing zone identifier from the ARN
func getLandingZoneIdentifierFromARN(arnString string) (string, error) {
	arn, err := arn.Parse(arnString)
	if err != nil {
		return "", err
	}
	resourceParts := strings.Split(arn.Resource, "/")
	return resourceParts[len(resourceParts)-1], nil
}

func findLandingZoneOperationDetailsByID(ctx context.Context, conn *controltower.Client, id string) (*types.LandingZoneOperationDetail, error) {
	input := &controltower.GetLandingZoneOperationInput{
		OperationIdentifier: aws.String(id),
	}

	output, err := conn.GetLandingZoneOperation(ctx, input)

	if tfresource.NotFound(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.OperationDetails == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.OperationDetails, nil
}

func statusLandingZoneOperation(ctx context.Context, conn *controltower.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findLandingZoneOperationDetailsByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitLandingZoneOperationSucceeded(ctx context.Context, conn *controltower.Client, id string, timeout time.Duration) (*types.LandingZoneOperationDetail, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(types.LandingZoneOperationStatusInProgress)},
		Target:  []string{string(types.LandingZoneOperationStatusSucceeded)},
		Refresh: statusLandingZoneOperation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.LandingZoneOperationDetail); ok {
		if status := output.Status; status == types.LandingZoneOperationStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}
