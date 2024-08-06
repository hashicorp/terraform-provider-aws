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

// @SDKResource("aws_guardduty_ipset", name="IP Set")
// @Tags(identifierAttribute="arn")
func ResourceIPSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPSetCreate,
		ReadWithoutTimeout:   resourceIPSetRead,
		UpdateWithoutTimeout: resourceIPSetUpdate,
		DeleteWithoutTimeout: resourceIPSetDelete,

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
				ValidateDiagFunc: enum.Validate[awstypes.IpSetFormat](),
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

func resourceIPSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID := d.Get("detector_id").(string)
	input := &guardduty.CreateIPSetInput{
		DetectorId: aws.String(detectorID),
		Name:       aws.String(d.Get(names.AttrName).(string)),
		Format:     awstypes.IpSetFormat(d.Get(names.AttrFormat).(string)),
		Location:   aws.String(d.Get(names.AttrLocation).(string)),
		Activate:   aws.Bool(d.Get("activate").(bool)),
		Tags:       getTagsIn(ctx),
	}

	resp, err := conn.CreateIPSet(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty IPSet (%s): %s", d.Get(names.AttrName).(string), err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.IpSetStatusActivating, awstypes.IpSetStatusDeactivating),
		Target:     enum.Slice(awstypes.IpSetStatusActive, awstypes.IpSetStatusInactive),
		Refresh:    ipsetRefreshStatusFunc(ctx, conn, *resp.IpSetId, detectorID),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty IPSet (%s): waiting for completion: %s", d.Get(names.AttrName).(string), err)
	}

	d.SetId(fmt.Sprintf("%s:%s", detectorID, *resp.IpSetId))

	return append(diags, resourceIPSetRead(ctx, d, meta)...)
}

func resourceIPSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	ipSetId, detectorId, err := DecodeIPSetID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty IPSet (%s): %s", d.Id(), err)
	}
	input := &guardduty.GetIPSetInput{
		DetectorId: aws.String(detectorId),
		IpSetId:    aws.String(ipSetId),
	}

	resp, err := conn.GetIPSet(ctx, input)
	if err != nil {
		if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
			log.Printf("[WARN] GuardDuty IPSet (%s) not found, removing from state", ipSetId)
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty IPSet (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "guardduty",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("detector/%s/ipset/%s", detectorId, ipSetId),
	}.String()
	d.Set(names.AttrARN, arn)

	d.Set("detector_id", detectorId)
	d.Set(names.AttrFormat, resp.Format)
	d.Set(names.AttrLocation, resp.Location)
	d.Set(names.AttrName, resp.Name)
	d.Set("activate", resp.Status == awstypes.IpSetStatusActive)

	setTagsOut(ctx, resp.Tags)

	return diags
}

func resourceIPSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	ipSetId, detectorId, err := DecodeIPSetID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty IPSet (%s): %s", d.Id(), err)
	}

	if d.HasChanges("activate", names.AttrLocation, names.AttrName) {
		input := &guardduty.UpdateIPSetInput{
			DetectorId: aws.String(detectorId),
			IpSetId:    aws.String(ipSetId),
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

		_, err = conn.UpdateIPSet(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GuardDuty IPSet (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceIPSetRead(ctx, d, meta)...)
}

func resourceIPSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	ipSetId, detectorId, err := DecodeIPSetID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty IPSet (%s): %s", d.Id(), err)
	}
	input := &guardduty.DeleteIPSetInput{
		DetectorId: aws.String(detectorId),
		IpSetId:    aws.String(ipSetId),
	}

	_, err = conn.DeleteIPSet(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty IPSet (%s): %s", d.Id(), err)
	}

	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.IpSetStatusActive,
			awstypes.IpSetStatusActivating,
			awstypes.IpSetStatusInactive,
			awstypes.IpSetStatusDeactivating,
			awstypes.IpSetStatusDeletePending,
		),
		Target:     enum.Slice(awstypes.IpSetStatusDeleted),
		Refresh:    ipsetRefreshStatusFunc(ctx, conn, ipSetId, detectorId),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty IPSet (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func ipsetRefreshStatusFunc(ctx context.Context, conn *guardduty.Client, ipSetID, detectorID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &guardduty.GetIPSetInput{
			DetectorId: aws.String(detectorID),
			IpSetId:    aws.String(ipSetID),
		}
		resp, err := conn.GetIPSet(ctx, input)
		if err != nil {
			return nil, "failed", err
		}
		return resp, string(resp.Status), nil
	}
}

func DecodeIPSetID(id string) (ipsetID, detectorID string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = fmt.Errorf("GuardDuty IPSet ID must be of the form <Detector ID>:<IPSet ID>, was provided: %s", id)
		return
	}
	ipsetID = parts[1]
	detectorID = parts[0]
	return
}
