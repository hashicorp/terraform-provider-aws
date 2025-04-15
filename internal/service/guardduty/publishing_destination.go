// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_guardduty_publishing_destination", name="Publishing Destination")
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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.DestinationTypeS3,
				ValidateDiagFunc: enum.Validate[awstypes.DestinationType](),
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

func resourcePublishingDestinationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID := d.Get("detector_id").(string)
	input := guardduty.CreatePublishingDestinationInput{
		DetectorId: aws.String(detectorID),
		DestinationProperties: &awstypes.DestinationProperties{
			DestinationArn: aws.String(d.Get(names.AttrDestinationARN).(string)),
			KmsKeyArn:      aws.String(d.Get(names.AttrKMSKeyARN).(string)),
		},
		DestinationType: awstypes.DestinationType(d.Get("destination_type").(string)),
	}

	output, err := conn.CreatePublishingDestination(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty Publishing Destination: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", d.Get("detector_id"), aws.ToString(output.DestinationId)))

	_, err = waitPublishingDestinationCreated(ctx, conn, aws.ToString(output.DestinationId), detectorID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty PublishingDestination status to be \"%s\": %s",
			string(awstypes.PublishingStatusPublishing), err)
	}

	return append(diags, resourcePublishingDestinationRead(ctx, d, meta)...)
}

func resourcePublishingDestinationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	destinationId, detectorId, err := DecodePublishDestinationID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Publishing Destination (%s): %s", d.Id(), err)
	}

	input := &guardduty.DescribePublishingDestinationInput{
		DetectorId:    aws.String(detectorId),
		DestinationId: aws.String(destinationId),
	}

	gdo, err := conn.DescribePublishingDestination(ctx, input)
	if err != nil {
		if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the one or more input parameters have invalid values.") {
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

func resourcePublishingDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	destinationId, detectorId, err := DecodePublishDestinationID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Publishing Destination (%s): %s", d.Id(), err)
	}

	input := guardduty.UpdatePublishingDestinationInput{
		DestinationId: aws.String(destinationId),
		DetectorId:    aws.String(detectorId),
		DestinationProperties: &awstypes.DestinationProperties{
			DestinationArn: aws.String(d.Get(names.AttrDestinationARN).(string)),
			KmsKeyArn:      aws.String(d.Get(names.AttrKMSKeyARN).(string)),
		},
	}

	if _, err = conn.UpdatePublishingDestination(ctx, &input); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Publishing Destination (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePublishingDestinationRead(ctx, d, meta)...)
}

func resourcePublishingDestinationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	destinationId, detectorId, err := DecodePublishDestinationID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty Publishing Destination (%s): %s", d.Id(), err)
	}

	input := guardduty.DeletePublishingDestinationInput{
		DestinationId: aws.String(destinationId),
		DetectorId:    aws.String(detectorId),
	}

	log.Printf("[DEBUG] Delete GuardDuty Publishing Destination: %+v", input)
	_, err = conn.DeletePublishingDestination(ctx, &input)

	if errs.IsA[*awstypes.BadRequestException](err) {
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
