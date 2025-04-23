// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_guardduty_detector", name="Detector")
// @Tags(identifierAttribute="arn")
func ResourceDetector() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDetectorCreate,
		ReadWithoutTimeout:   resourceDetectorRead,
		UpdateWithoutTimeout: resourceDetectorUpdate,
		DeleteWithoutTimeout: resourceDetectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			// finding_publishing_frequency is marked as Computed:true since
			// GuardDuty member accounts inherit setting from master account
			"finding_publishing_frequency": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceDetectorCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	input := &guardduty.CreateDetectorInput{
		Enable: aws.Bool(d.Get("enable").(bool)),
		Tags:   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("finding_publishing_frequency"); ok {
		input.FindingPublishingFrequency = awstypes.FindingPublishingFrequency(v.(string))
	}

	output, err := conn.CreateDetector(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty Detector: %s", err)
	}

	d.SetId(aws.ToString(output.DetectorId))

	return append(diags, resourceDetectorRead(ctx, d, meta)...)
}

func resourceDetectorRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	gdo, err := FindDetectorByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GuardDuty Detector (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Detector (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, meta.(*conns.AWSClient).AccountID(ctx))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Region:    meta.(*conns.AWSClient).Region(ctx),
		Service:   "guardduty",
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("detector/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("enable", gdo.Status == awstypes.DetectorStatusEnabled)
	d.Set("finding_publishing_frequency", gdo.FindingPublishingFrequency)

	setTagsOut(ctx, gdo.Tags)

	return diags
}

func resourceDetectorUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &guardduty.UpdateDetectorInput{
			DetectorId:                 aws.String(d.Id()),
			Enable:                     aws.Bool(d.Get("enable").(bool)),
			FindingPublishingFrequency: awstypes.FindingPublishingFrequency(d.Get("finding_publishing_frequency").(string)),
		}

		_, err := conn.UpdateDetector(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GuardDuty Detector (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDetectorRead(ctx, d, meta)...)
}

func resourceDetectorDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	log.Printf("[DEBUG] Deleting GuardDuty Detector: %s", d.Id())
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.BadRequestException](ctx, membershipPropagationTimeout, func() (any, error) {
		return conn.DeleteDetector(ctx, &guardduty.DeleteDetectorInput{
			DetectorId: aws.String(d.Id()),
		})
	}, "cannot delete detector while it has invited or associated members")

	if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty Detector (%s): %s", d.Id(), err)
	}

	return diags
}

func FindDetectorByID(ctx context.Context, conn *guardduty.Client, id string) (*guardduty.GetDetectorOutput, error) {
	input := &guardduty.GetDetectorInput{
		DetectorId: aws.String(id),
	}

	output, err := conn.GetDetector(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
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

// FindDetector returns the ID of the current account's active GuardDuty detector.
func FindDetector(ctx context.Context, conn *guardduty.Client) (*string, error) {
	output, err := findDetectors(ctx, conn)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDetectors(ctx context.Context, conn *guardduty.Client) ([]string, error) {
	input := &guardduty.ListDetectorsInput{}
	var output []string

	pages := guardduty.NewListDetectorsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.DetectorIds...)
	}

	return output, nil
}
