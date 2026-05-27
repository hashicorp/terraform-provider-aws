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

// @SDKResource("aws_guardduty_ipset", name="IP Set")
// @Tags(identifierAttribute="arn")
// @Testing(serialize=true)
// @Testing(preCheck="testAccPreCheckDetectorNotExists")
func resourceIPSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPSetCreate,
		ReadWithoutTimeout:   resourceIPSetRead,
		UpdateWithoutTimeout: resourceIPSetUpdate,
		DeleteWithoutTimeout: resourceIPSetDelete,

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
				ValidateDiagFunc: enum.Validate[awstypes.IpSetFormat](),
			},
			"ip_set_id": {
				Type:     schema.TypeString,
				Computed: true,
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
		},
	}
}

func resourceIPSetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID := d.Get("detector_id").(string)
	name := d.Get(names.AttrName).(string)
	input := guardduty.CreateIPSetInput{
		Activate:   aws.Bool(d.Get("activate").(bool)),
		DetectorId: aws.String(detectorID),
		Format:     awstypes.IpSetFormat(d.Get(names.AttrFormat).(string)),
		Location:   aws.String(d.Get(names.AttrLocation).(string)),
		Name:       aws.String(name),
		Tags:       getTagsIn(ctx),
	}

	output, err := conn.CreateIPSet(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty IPSet (%s): %s", name, err)
	}

	ipSetID := aws.ToString(output.IpSetId)
	d.SetId(ipSetCreateResourceID(detectorID, ipSetID))

	if _, err := waitIPSetCreated(ctx, conn, detectorID, ipSetID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty IPSet (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceIPSetRead(ctx, d, meta)...)
}

func resourceIPSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.GuardDutyClient(ctx)

	detectorID, ipSetID, err := ipSetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findIPSetByTwoPartKey(ctx, conn, detectorID, ipSetID)
	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] GuardDuty IPSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty IPSet (%s): %s", d.Id(), err)
	}

	d.Set("activate", output.Status == awstypes.IpSetStatusActive)
	d.Set(names.AttrARN, ipSetARN(ctx, c, detectorID, ipSetID))
	d.Set("detector_id", detectorID)
	d.Set(names.AttrFormat, output.Format)
	d.Set("ip_set_id", ipSetID)
	d.Set(names.AttrLocation, output.Location)
	d.Set(names.AttrName, output.Name)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceIPSetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		detectorID, ipSetID, err := ipSetParseResourceID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := guardduty.UpdateIPSetInput{
			DetectorId: aws.String(detectorID),
			IpSetId:    aws.String(ipSetID),
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

		_, err = conn.UpdateIPSet(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GuardDuty IPSet (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceIPSetRead(ctx, d, meta)...)
}

func resourceIPSetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID, ipSetID, err := ipSetParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	input := guardduty.DeleteIPSetInput{
		DetectorId: aws.String(detectorID),
		IpSetId:    aws.String(ipSetID),
	}
	_, err = conn.DeleteIPSet(ctx, &input)
	if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected since no such resource found.") {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty IPSet (%s): %s", d.Id(), err)
	}

	if _, err := waitIPSetDeleted(ctx, conn, detectorID, ipSetID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty IPSet (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const ipSetResourceIDSeparator = ":"

func ipSetCreateResourceID(detectorID, ipSetID string) string {
	parts := []string{detectorID, ipSetID}
	id := strings.Join(parts, ipSetResourceIDSeparator)

	return id
}

func ipSetParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, ipSetResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected <Detector ID>%[2]s<IPSet ID>", id, ipSetResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findIPSetByTwoPartKey(ctx context.Context, conn *guardduty.Client, detectorID, ipSetID string) (*guardduty.GetIPSetOutput, error) {
	input := guardduty.GetIPSetInput{
		DetectorId: aws.String(detectorID),
		IpSetId:    aws.String(ipSetID),
	}

	output, err := findIPSet(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.IpSetStatusDeleted {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, nil
}

func findIPSet(ctx context.Context, conn *guardduty.Client, input *guardduty.GetIPSetInput) (*guardduty.GetIPSetOutput, error) {
	output, err := conn.GetIPSet(ctx, input)

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

func statusIPSet(conn *guardduty.Client, detectorID, ipSetID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findIPSetByTwoPartKey(ctx, conn, detectorID, ipSetID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitIPSetCreated(ctx context.Context, conn *guardduty.Client, detectorID, ipSetID string) (*guardduty.GetIPSetOutput, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.IpSetStatusActivating, awstypes.IpSetStatusDeactivating),
		Target:     enum.Slice(awstypes.IpSetStatusActive, awstypes.IpSetStatusInactive),
		Refresh:    statusIPSet(conn, detectorID, ipSetID),
		Timeout:    timeout,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*guardduty.GetIPSetOutput); ok {
		return output, err
	}

	return nil, err
}

func waitIPSetDeleted(ctx context.Context, conn *guardduty.Client, detectorID, ipSetID string) (*guardduty.GetIPSetOutput, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.IpSetStatusActive,
			awstypes.IpSetStatusActivating,
			awstypes.IpSetStatusInactive,
			awstypes.IpSetStatusDeactivating,
			awstypes.IpSetStatusDeletePending,
		),
		Target:     []string{},
		Refresh:    statusIPSet(conn, detectorID, ipSetID),
		Timeout:    timeout,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*guardduty.GetIPSetOutput); ok {
		return output, err
	}

	return nil, err
}

func ipSetARN(ctx context.Context, c *conns.AWSClient, detectorID, ipSetID string) string {
	return c.RegionalARN(ctx, "guardduty", "detector/"+detectorID+"/ipset/"+ipSetID)
}
