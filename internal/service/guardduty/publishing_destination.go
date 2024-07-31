// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_guardduty_publishing_destination")
func ResourcePublishingDestination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePublishingDestinationCreate,
		ReadWithoutTimeout:   resourcePublishingDestinationRead,
		UpdateWithoutTimeout: resourcePublishingDestinationUpdate,
		DeleteWithoutTimeout: resourcePublishingDestinationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      guardduty.DestinationTypeS3,
				ValidateFunc: validation.StringInSlice(guardduty.DestinationType_Values(), false),
			},
			names.AttrDestinationARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrKMSKeyARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourcePublishingDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	detectorID := d.Get("detector_id").(string)
	input := guardduty.CreatePublishingDestinationInput{
		DetectorId: aws.String(detectorID),
		DestinationProperties: &guardduty.DestinationProperties{
			DestinationArn: aws.String(d.Get(names.AttrDestinationARN).(string)),
			KmsKeyArn:      aws.String(d.Get(names.AttrKMSKeyARN).(string)),
		},
		DestinationType: aws.String(d.Get("destination_type").(string)),
	}

	output, err := conn.CreatePublishingDestinationWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty Publishing Destination: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", d.Get("detector_id"), aws.StringValue(output.DestinationId)))

	_, err = waitPublishingDestinationCreated(ctx, conn, aws.StringValue(output.DestinationId), detectorID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty PublishingDestination status to be \"%s\": %s",
			guardduty.PublishingStatusPublishing, err)
	}

	return append(diags, resourcePublishingDestinationRead(ctx, d, meta)...)
}

func resourcePublishingDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	destinationId, detectorId, err := DecodePublishDestinationID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Publishing Destination (%s): %s", d.Id(), err)
	}

	input := &guardduty.DescribePublishingDestinationInput{
		DetectorId:    aws.String(detectorId),
		DestinationId: aws.String(destinationId),
	}

	gdo, err := conn.DescribePublishingDestinationWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the one or more input parameters have invalid values.") {
			log.Printf("[WARN] GuardDuty Publishing Destination: %q not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Publishing Destination (%s): %s", d.Id(), err)
	}

	d.Set("detector_id", detectorId)
	d.Set("destination_type", gdo.DestinationType)
	d.Set(names.AttrKMSKeyARN, gdo.DestinationProperties.KmsKeyArn)
	d.Set(names.AttrDestinationARN, gdo.DestinationProperties.DestinationArn)
	return diags
}

func resourcePublishingDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	destinationId, detectorId, err := DecodePublishDestinationID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Publishing Destination (%s): %s", d.Id(), err)
	}

	input := guardduty.UpdatePublishingDestinationInput{
		DestinationId: aws.String(destinationId),
		DetectorId:    aws.String(detectorId),
		DestinationProperties: &guardduty.DestinationProperties{
			DestinationArn: aws.String(d.Get(names.AttrDestinationARN).(string)),
			KmsKeyArn:      aws.String(d.Get(names.AttrKMSKeyARN).(string)),
		},
	}

	if _, err = conn.UpdatePublishingDestinationWithContext(ctx, &input); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Publishing Destination (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePublishingDestinationRead(ctx, d, meta)...)
}

func resourcePublishingDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	destinationId, detectorId, err := DecodePublishDestinationID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty Publishing Destination (%s): %s", d.Id(), err)
	}

	input := guardduty.DeletePublishingDestinationInput{
		DestinationId: aws.String(destinationId),
		DetectorId:    aws.String(detectorId),
	}

	log.Printf("[DEBUG] Delete GuardDuty Publishing Destination: %s", input)
	_, err = conn.DeletePublishingDestinationWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, guardduty.ErrCodeBadRequestException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty Publishing Destination (%s): %s", d.Id(), err)
	}

	return diags
}

func DecodePublishDestinationID(id string) (destinationID, detectorID string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = fmt.Errorf("GuardDuty Publishing Destination ID must be of the form <Detector ID>:<Publishing Destination ID>, was provided: %s", id)
		return
	}
	destinationID = parts[1]
	detectorID = parts[0]
	return
}
