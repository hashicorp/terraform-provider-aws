// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_site_to_site_vpn_attachment", name="Site To Site VPN Attachment")
// @Tags(identifierAttribute="arn")
func resourceSiteToSiteVPNAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSiteToSiteVPNAttachmentCreate,
		ReadWithoutTimeout:   resourceSiteToSiteVPNAttachmentRead,
		UpdateWithoutTimeout: resourceSiteToSiteVPNAttachmentUpdate,
		DeleteWithoutTimeout: resourceSiteToSiteVPNAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attachment_policy_rule_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"attachment_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"edge_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"segment_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpn_connection_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:[^:]{1,63}:ec2:[^:]{0,63}:[^:]{0,63}:vpn-connection\/vpn-[0-9a-f]{8,17}$`), "Must be valid VPN ARN"),
			},
		},
	}
}

func resourceSiteToSiteVPNAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	coreNetworkID := d.Get("core_network_id").(string)
	vpnConnectionARN := d.Get("vpn_connection_arn").(string)
	input := &networkmanager.CreateSiteToSiteVpnAttachmentInput{
		CoreNetworkId:    aws.String(coreNetworkID),
		Tags:             getTagsIn(ctx),
		VpnConnectionArn: aws.String(vpnConnectionARN),
	}

	output, err := conn.CreateSiteToSiteVpnAttachment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Site To Site VPN (%s) Attachment (%s): %s", vpnConnectionARN, coreNetworkID, err)
	}

	d.SetId(aws.ToString(output.SiteToSiteVpnAttachment.Attachment.AttachmentId))

	if _, err := waitSiteToSiteVPNAttachmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Site To Site VPN Attachment (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceSiteToSiteVPNAttachmentRead(ctx, d, meta)...)
}

func resourceSiteToSiteVPNAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	vpnAttachment, err := findSiteToSiteVPNAttachmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Site To Site VPN Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Site To Site VPN Attachment (%s): %s", d.Id(), err)
	}

	attachment := vpnAttachment.Attachment
	d.Set(names.AttrARN, attachmentARN(ctx, meta.(*conns.AWSClient), d.Id()))
	d.Set("attachment_policy_rule_number", attachment.AttachmentPolicyRuleNumber)
	d.Set("attachment_type", attachment.AttachmentType)
	d.Set("core_network_arn", attachment.CoreNetworkArn)
	d.Set("core_network_id", attachment.CoreNetworkId)
	d.Set("edge_location", attachment.EdgeLocation)
	d.Set(names.AttrOwnerAccountID, attachment.OwnerAccountId)
	d.Set(names.AttrResourceARN, attachment.ResourceArn)
	d.Set("segment_name", attachment.SegmentName)
	d.Set(names.AttrState, attachment.State)
	d.Set("vpn_connection_arn", attachment.ResourceArn)

	setTagsOut(ctx, attachment.Tags)

	return diags
}

func resourceSiteToSiteVPNAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceSiteToSiteVPNAttachmentRead(ctx, d, meta)
}

func resourceSiteToSiteVPNAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	// If ResourceAttachmentAccepter is used, then VPN Attachment state
	// is never updated from StatePendingAttachmentAcceptance and the delete fails
	output, err := findSiteToSiteVPNAttachmentByID(ctx, conn, d.Id())

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Site To Site VPN Attachment (%s): %s", d.Id(), err)
	}

	if state := output.Attachment.State; state == awstypes.AttachmentStatePendingAttachmentAcceptance || state == awstypes.AttachmentStatePendingTagAcceptance {
		return sdkdiag.AppendErrorf(diags, "cannot delete Network Manager Site To Site VPN Attachment (%s) in state: %s", d.Id(), state)
	}

	log.Printf("[DEBUG] Deleting Network Manager Site To Site VPN Attachment: %s", d.Id())
	_, err = conn.DeleteAttachment(ctx, &networkmanager.DeleteAttachmentInput{
		AttachmentId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Network Manager Site To Site VPN Attachment (%s): %s", d.Id(), err)
	}

	if _, err := waitSiteToSiteVPNAttachmentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Site To Site VPN Attachment (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findSiteToSiteVPNAttachmentByID(ctx context.Context, conn *networkmanager.Client, id string) (*awstypes.SiteToSiteVpnAttachment, error) {
	input := &networkmanager.GetSiteToSiteVpnAttachmentInput{
		AttachmentId: aws.String(id),
	}

	output, err := conn.GetSiteToSiteVpnAttachment(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.SiteToSiteVpnAttachment == nil || output.SiteToSiteVpnAttachment.Attachment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SiteToSiteVpnAttachment, nil
}

func statusSiteToSiteVPNAttachment(ctx context.Context, conn *networkmanager.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findSiteToSiteVPNAttachmentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Attachment.State), nil
	}
}

func waitSiteToSiteVPNAttachmentCreated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.SiteToSiteVpnAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AttachmentStateCreating, awstypes.AttachmentStatePendingNetworkUpdate),
		Target:                    enum.Slice(awstypes.AttachmentStateAvailable, awstypes.AttachmentStatePendingAttachmentAcceptance),
		Timeout:                   timeout,
		Refresh:                   statusSiteToSiteVPNAttachment(ctx, conn, id),
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SiteToSiteVpnAttachment); ok {
		tfresource.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitSiteToSiteVPNAttachmentDeleted(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.SiteToSiteVpnAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.AttachmentStateDeleting),
		Target:         []string{},
		Timeout:        timeout,
		Refresh:        statusSiteToSiteVPNAttachment(ctx, conn, id),
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SiteToSiteVpnAttachment); ok {
		tfresource.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitSiteToSiteVPNAttachmentAvailable(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.SiteToSiteVpnAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStateCreating, awstypes.AttachmentStatePendingAttachmentAcceptance, awstypes.AttachmentStatePendingNetworkUpdate),
		Target:  enum.Slice(awstypes.AttachmentStateAvailable),
		Timeout: timeout,
		Refresh: statusSiteToSiteVPNAttachment(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SiteToSiteVpnAttachment); ok {
		tfresource.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}
