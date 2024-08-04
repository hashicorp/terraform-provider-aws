// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_guardduty_threatintelset", name="Threat Intel Set")
// @Tags(identifierAttribute="arn")
func ResourceThreatIntelSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThreatIntelSetCreate,
		ReadWithoutTimeout:   resourceThreatIntelSetRead,
		UpdateWithoutTimeout: resourceThreatIntelSetUpdate,
		DeleteWithoutTimeout: resourceThreatIntelSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrFormat: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ThreatIntelSetFormat](),
			},
			names.AttrLocation: {
				Type:     schema.TypeString,
				Required: true,
			},
			"activate": {
				Type:     schema.TypeBool,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceThreatIntelSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID := d.Get("detector_id").(string)
	name := d.Get(names.AttrName).(string)
	input := &guardduty.CreateThreatIntelSetInput{
		DetectorId: aws.String(detectorID),
		Name:       aws.String(name),
		Format:     awstypes.ThreatIntelSetFormat(d.Get(names.AttrFormat).(string)),
		Location:   aws.String(d.Get(names.AttrLocation).(string)),
		Activate:   aws.Bool(d.Get("activate").(bool)),
		Tags:       getTagsIn(ctx),
	}

	resp, err := conn.CreateThreatIntelSet(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty Threat Intel Set (%s): %s", name, err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ThreatIntelSetStatusActivating, awstypes.ThreatIntelSetStatusDeactivating),
		Target:     enum.Slice(awstypes.ThreatIntelSetStatusActive, awstypes.ThreatIntelSetStatusInactive),
		Refresh:    threatintelsetRefreshStatusFunc(ctx, conn, *resp.ThreatIntelSetId, detectorID),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty Threat Intel Set (%s): waiting for completion: %s", name, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", detectorID, aws.ToString(resp.ThreatIntelSetId)))

	return append(diags, resourceThreatIntelSetRead(ctx, d, meta)...)
}

func resourceThreatIntelSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	threatIntelSetId, detectorId, err := DecodeThreatIntelSetID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Threat Intel Set (%s): %s", d.Id(), err)
	}
	input := &guardduty.GetThreatIntelSetInput{
		DetectorId:       aws.String(detectorId),
		ThreatIntelSetId: aws.String(threatIntelSetId),
	}

	resp, err := conn.GetThreatIntelSet(ctx, input)
	if err != nil {
		if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
			log.Printf("[WARN] GuardDuty ThreatIntelSet %q not found, removing from state", threatIntelSetId)
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Threat Intel Set (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "guardduty",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("detector/%s/threatintelset/%s", detectorId, threatIntelSetId),
	}.String()
	d.Set(names.AttrARN, arn)

	d.Set("detector_id", detectorId)
	d.Set(names.AttrFormat, resp.Format)
	d.Set(names.AttrLocation, resp.Location)
	d.Set(names.AttrName, resp.Name)
	d.Set("activate", resp.Status == awstypes.ThreatIntelSetStatusActive)

	setTagsOut(ctx, resp.Tags)

	return diags
}

func resourceThreatIntelSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	threatIntelSetID, detectorId, err := DecodeThreatIntelSetID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Threat Intel Set (%s): %s", d.Id(), err)
	}

	if d.HasChanges("activate", names.AttrLocation, names.AttrName) {
		input := &guardduty.UpdateThreatIntelSetInput{
			DetectorId:       aws.String(detectorId),
			ThreatIntelSetId: aws.String(threatIntelSetID),
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}
		if d.HasChange(names.AttrLocation) {
			input.Location = aws.String(d.Get(names.AttrLocation).(string))
		}
		if d.HasChange("activate") {
			input.Activate = aws.Bool(d.Get("activate").(bool))
		}

		if _, err = conn.UpdateThreatIntelSet(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GuardDuty Threat Intel Set (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceThreatIntelSetRead(ctx, d, meta)...)
}

func resourceThreatIntelSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	threatIntelSetID, detectorId, err := DecodeThreatIntelSetID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty Threat Intel Set (%s): %s", d.Id(), err)
	}
	input := &guardduty.DeleteThreatIntelSetInput{
		DetectorId:       aws.String(detectorId),
		ThreatIntelSetId: aws.String(threatIntelSetID),
	}

	_, err = conn.DeleteThreatIntelSet(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty Threat Intel Set (%s): %s", d.Id(), err)
	}

	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.ThreatIntelSetStatusActive,
			awstypes.ThreatIntelSetStatusActivating,
			awstypes.ThreatIntelSetStatusInactive,
			awstypes.ThreatIntelSetStatusDeactivating,
			awstypes.ThreatIntelSetStatusDeletePending,
		),
		Target:     enum.Slice(awstypes.ThreatIntelSetStatusDeleted),
		Refresh:    threatintelsetRefreshStatusFunc(ctx, conn, threatIntelSetID, detectorId),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty ThreatIntelSet status to be \"%s\": %s", string(awstypes.ThreatIntelSetStatusDeleted), err)
	}

	return diags
}

func threatintelsetRefreshStatusFunc(ctx context.Context, conn *guardduty.Client, threatIntelSetID, detectorID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &guardduty.GetThreatIntelSetInput{
			DetectorId:       aws.String(detectorID),
			ThreatIntelSetId: aws.String(threatIntelSetID),
		}
		resp, err := conn.GetThreatIntelSet(ctx, input)
		if err != nil {
			return nil, "failed", err
		}
		return resp, string(resp.Status), nil
	}
}

func DecodeThreatIntelSetID(id string) (threatIntelSetID, detectorID string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = fmt.Errorf("GuardDuty ThreatIntelSet ID must be of the form <Detector ID>:<ThreatIntelSet ID>, was provided: %s", id)
		return
	}
	threatIntelSetID = parts[1]
	detectorID = parts[0]
	return
}
