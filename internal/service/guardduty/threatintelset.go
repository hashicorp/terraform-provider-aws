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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_guardduty_threatintelset", name="Threat Intel Set")
// @Tags(identifierAttribute="arn")
// @Testing(serialize=true)
// @Testing(preCheck="testAccPreCheckDetectorNotExists")
func resourceThreatIntelSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThreatIntelSetCreate,
		ReadWithoutTimeout:   resourceThreatIntelSetRead,
		UpdateWithoutTimeout: resourceThreatIntelSetUpdate,
		DeleteWithoutTimeout: resourceThreatIntelSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"activate": {
				Type:     schema.TypeBool,
				Required: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"threat_intel_set_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceThreatIntelSetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID := d.Get("detector_id").(string)
	name := d.Get(names.AttrName).(string)
	input := guardduty.CreateThreatIntelSetInput{
		Activate:   aws.Bool(d.Get("activate").(bool)),
		DetectorId: aws.String(detectorID),
		Format:     awstypes.ThreatIntelSetFormat(d.Get(names.AttrFormat).(string)),
		Location:   aws.String(d.Get(names.AttrLocation).(string)),
		Name:       aws.String(name),
		Tags:       getTagsIn(ctx),
	}

	output, err := conn.CreateThreatIntelSet(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty Threat Intel Set (%s): %s", name, err)
	}

	threatIntelSetID := aws.ToString(output.ThreatIntelSetId)
	d.SetId(threatIntelSetCreateResourceID(detectorID, threatIntelSetID))

	if _, err := waitThreatIntelSetCreated(ctx, conn, detectorID, threatIntelSetID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty Threat Intel Set (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceThreatIntelSetRead(ctx, d, meta)...)
}

func resourceThreatIntelSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.GuardDutyClient(ctx)

	detectorID, threatIntelSetID, err := threatIntelSetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findThreatIntelSetByTwoPartKey(ctx, conn, detectorID, threatIntelSetID)
	if retry.NotFound(err) {
		log.Printf("[WARN] GuardDuty Threat Intel Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Threat Intel Set (%s): %s", d.Id(), err)
	}

	d.Set("activate", output.Status == awstypes.ThreatIntelSetStatusActive)
	d.Set(names.AttrARN, threatIntelSetARN(ctx, c, detectorID, threatIntelSetID))
	d.Set("detector_id", detectorID)
	d.Set(names.AttrFormat, output.Format)
	d.Set(names.AttrLocation, output.Location)
	d.Set(names.AttrName, output.Name)
	d.Set("threat_intel_set_id", threatIntelSetID)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceThreatIntelSetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		detectorID, threatIntelSetID, err := threatIntelSetParseResourceID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := guardduty.UpdateThreatIntelSetInput{
			DetectorId:       aws.String(detectorID),
			ThreatIntelSetId: aws.String(threatIntelSetID),
		}

		if d.HasChange("activate") {
			input.Activate = aws.Bool(d.Get("activate").(bool))
		}
		if d.HasChange(names.AttrLocation) {
			input.Location = aws.String(d.Get(names.AttrLocation).(string))
		}
		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		_, err = conn.UpdateThreatIntelSet(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GuardDuty Threat Intel Set (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceThreatIntelSetRead(ctx, d, meta)...)
}

func resourceThreatIntelSetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID, threatIntelSetID, err := threatIntelSetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := guardduty.DeleteThreatIntelSetInput{
		DetectorId:       aws.String(detectorID),
		ThreatIntelSetId: aws.String(threatIntelSetID),
	}
	_, err = conn.DeleteThreatIntelSet(ctx, &input)
	if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected since no such resource found.") {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty Threat Intel Set (%s): %s", d.Id(), err)
	}

	if _, err := waitThreatIntelSetDeleted(ctx, conn, detectorID, threatIntelSetID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty Threat Intel Set (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const threatIntelSetResourceIDSeparator = ":"

func threatIntelSetCreateResourceID(detectorID, threatIntelSetID string) string {
	parts := []string{detectorID, threatIntelSetID}
	id := strings.Join(parts, threatIntelSetResourceIDSeparator)

	return id
}

func threatIntelSetParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, threatIntelSetResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected <Detector ID>%[2]s<ThreatIntelSet ID>", id, threatIntelSetResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findThreatIntelSetByTwoPartKey(ctx context.Context, conn *guardduty.Client, detectorID, threatIntelSetID string) (*guardduty.GetThreatIntelSetOutput, error) {
	input := guardduty.GetThreatIntelSetInput{
		DetectorId:       aws.String(detectorID),
		ThreatIntelSetId: aws.String(threatIntelSetID),
	}

	output, err := findThreatIntelSet(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.ThreatIntelSetStatusDeleted {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, nil
}

func findThreatIntelSet(ctx context.Context, conn *guardduty.Client, input *guardduty.GetThreatIntelSetInput) (*guardduty.GetThreatIntelSetOutput, error) {
	output, err := conn.GetThreatIntelSet(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusThreatIntelSet(conn *guardduty.Client, detectorID, threatIntelSetID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findThreatIntelSetByTwoPartKey(ctx, conn, detectorID, threatIntelSetID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitThreatIntelSetCreated(ctx context.Context, conn *guardduty.Client, detectorID, threatIntelSetID string) (*guardduty.GetThreatIntelSetOutput, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ThreatIntelSetStatusActivating, awstypes.ThreatIntelSetStatusDeactivating),
		Target:     enum.Slice(awstypes.ThreatIntelSetStatusActive, awstypes.ThreatIntelSetStatusInactive),
		Refresh:    statusThreatIntelSet(conn, detectorID, threatIntelSetID),
		Timeout:    timeout,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*guardduty.GetThreatIntelSetOutput); ok {
		return output, err
	}

	return nil, err
}

func waitThreatIntelSetDeleted(ctx context.Context, conn *guardduty.Client, detectorID, threatIntelSetID string) (*guardduty.GetThreatIntelSetOutput, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.ThreatIntelSetStatusActive,
			awstypes.ThreatIntelSetStatusActivating,
			awstypes.ThreatIntelSetStatusInactive,
			awstypes.ThreatIntelSetStatusDeactivating,
			awstypes.ThreatIntelSetStatusDeletePending,
		),
		Target:     []string{},
		Refresh:    statusThreatIntelSet(conn, detectorID, threatIntelSetID),
		Timeout:    timeout,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*guardduty.GetThreatIntelSetOutput); ok {
		return output, err
	}

	return nil, err
}

func threatIntelSetARN(ctx context.Context, c *conns.AWSClient, detectorID, threatIntelSetID string) string {
	return c.RegionalARN(ctx, "guardduty", "detector/"+detectorID+"/threatintelset/"+threatIntelSetID)
}
