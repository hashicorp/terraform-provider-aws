// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package guardduty

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_guardduty_publishing_destination", name="Publishing Destination")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourcePublishingDestination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePublishingDestinationCreate,
		ReadWithoutTimeout:   resourcePublishingDestinationRead,
		UpdateWithoutTimeout: resourcePublishingDestinationUpdate,
		DeleteWithoutTimeout: resourcePublishingDestinationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDestinationARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"destination_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.DestinationTypeS3,
				ValidateDiagFunc: enum.Validate[awstypes.DestinationType](),
			},
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrKMSKeyARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchemaForceNew(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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
		Tags:            getTagsIn(ctx),
	}

	output, err := conn.CreatePublishingDestination(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty Publishing Destination: %s", err)
	}

	destinationID := aws.ToString(output.DestinationId)
	d.SetId(publishingDestinationCreateResourceID(detectorID, destinationID))

	if _, err := waitPublishingDestinationCreated(ctx, conn, detectorID, destinationID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty Publishing Destination (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourcePublishingDestinationRead(ctx, d, meta)...)
}

func resourcePublishingDestinationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.GuardDutyClient(ctx)

	detectorID, destinationID, err := publishingDestinationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	gdo, err := findPublishingDestinationByTwoPartKey(ctx, conn, detectorID, destinationID)
	if retry.NotFound(err) {
		log.Printf("[WARN] GuardDuty Publishing Destination (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Publishing Destination (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, publishingDestinationARN(ctx, c, detectorID, destinationID))
	d.Set(names.AttrDestinationARN, gdo.DestinationProperties.DestinationArn)
	d.Set("destination_id", destinationID)
	d.Set("destination_type", gdo.DestinationType)
	d.Set("detector_id", detectorID)
	d.Set(names.AttrKMSKeyARN, gdo.DestinationProperties.KmsKeyArn)

	setTagsOut(ctx, gdo.Tags)

	return diags
}

func resourcePublishingDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		detectorID, destinationID, err := publishingDestinationParseResourceID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := guardduty.UpdatePublishingDestinationInput{
			DestinationId: aws.String(destinationID),
			DetectorId:    aws.String(detectorID),
			DestinationProperties: &awstypes.DestinationProperties{
				DestinationArn: aws.String(d.Get(names.AttrDestinationARN).(string)),
				KmsKeyArn:      aws.String(d.Get(names.AttrKMSKeyARN).(string)),
			},
		}

		_, err = conn.UpdatePublishingDestination(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GuardDuty Publishing Destination (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePublishingDestinationRead(ctx, d, meta)...)
}

func resourcePublishingDestinationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID, destinationID, err := publishingDestinationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting GuardDuty Publishing Destination: %s", d.Id())
	input := guardduty.DeletePublishingDestinationInput{
		DestinationId: aws.String(destinationID),
		DetectorId:    aws.String(detectorID),
	}
	_, err = conn.DeletePublishingDestination(ctx, &input)

	if errs.IsA[*awstypes.BadRequestException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty Publishing Destination (%s): %s", d.Id(), err)
	}

	return diags
}

const publishingDestinationResourceIDSeparator = ":"

func publishingDestinationCreateResourceID(detectorID, destinationID string) string {
	parts := []string{detectorID, destinationID}
	id := strings.Join(parts, publishingDestinationResourceIDSeparator)

	return id
}

func publishingDestinationParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, publishingDestinationResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected <Detector ID>%[2]s<Publishing Destination ID>", id, publishingDestinationResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findPublishingDestinationByTwoPartKey(ctx context.Context, conn *guardduty.Client, detectorID, destinationID string) (*guardduty.DescribePublishingDestinationOutput, error) {
	input := guardduty.DescribePublishingDestinationInput{
		DestinationId: aws.String(destinationID),
		DetectorId:    aws.String(detectorID),
	}

	return findPublishingDestination(ctx, conn, &input)
}

func findPublishingDestination(ctx context.Context, conn *guardduty.Client, input *guardduty.DescribePublishingDestinationInput) (*guardduty.DescribePublishingDestinationOutput, error) {
	output, err := conn.DescribePublishingDestination(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the one or more input parameters have invalid values.") ||
		errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DestinationProperties == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusPublishingDestination(conn *guardduty.Client, detectorID, destinationID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findPublishingDestinationByTwoPartKey(ctx, conn, detectorID, destinationID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitPublishingDestinationCreated(ctx context.Context, conn *guardduty.Client, detectorID, destinationID string) (*guardduty.DescribePublishingDestinationOutput, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PublishingStatusPendingVerification),
		Target:  enum.Slice(awstypes.PublishingStatusPublishing),
		Refresh: statusPublishingDestination(conn, detectorID, destinationID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*guardduty.DescribePublishingDestinationOutput); ok {
		return output, err
	}

	return nil, err
}

func publishingDestinationARN(ctx context.Context, c *conns.AWSClient, detectorID, destinationID string) string {
	return c.RegionalARN(ctx, "guardduty", "detector/"+detectorID+"/publishingdestination/"+destinationID)
}
