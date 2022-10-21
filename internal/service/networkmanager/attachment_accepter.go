package networkmanager

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// AttachmentAccepter does not require AttachmentType. However, querying attachments for status updates requires knowing tyupe
// To facilitate querying and waiters on specific attachment types, attachment_type set to required

func ResourceAttachmentAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAttachmentAccepterCreate,
		ReadWithoutTimeout:   resourceAttachmentAccepterRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"attachment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"attachment_policy_rule_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			// querying attachments requires knowing the type ahead of time
			// therefore type is required in provider, though not on the API
			"attachment_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(networkmanager.AttachmentType_Values(), false),
			},
			"core_network_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"edge_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"segment_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAttachmentAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn

	var state string
	attachmentID := d.Get("attachment_id").(string)
	aType := d.Get("attachment_type")

	switch aType {
	case "VPC":
		vpcAttachment, err := FindVPCAttachmentByID(ctx, conn, attachmentID)

		if err != nil {
			return diag.Errorf("reading Network Manager VPC Attachment (%s): %s", attachmentID, err)
		}

		state = aws.StringValue(vpcAttachment.Attachment.State)

		d.SetId(attachmentID)

	case "SITE_TO_SITE_VPN":
		vpnAttachment, err := FindVpnAttachmentByID(ctx, conn, attachmentID)

		if err != nil {
			return diag.Errorf("reading Network Manager VPC Attachment (%s): %s", attachmentID, err)
		}

		state = aws.StringValue(vpnAttachment.Attachment.State)

		d.SetId(attachmentID)

	default:
		return diag.Errorf("Unsupported Network Manager Attachment type: %s", aType)
	}

	if state == networkmanager.AttachmentStatePendingAttachmentAcceptance || state == networkmanager.AttachmentStatePendingTagAcceptance {
		input := &networkmanager.AcceptAttachmentInput{
			AttachmentId: aws.String(attachmentID),
		}

		_, err := conn.AcceptAttachmentWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("accepting Network Manager Attachment (%s): %s", attachmentID, err)
		}

		switch aType {
		case "VPC":
			if _, err := waitVPCAttachmentCreated(ctx, conn, attachmentID, d.Timeout(schema.TimeoutCreate)); err != nil {
				return diag.Errorf("waiting for Network Manager VPC Attachment (%s) create: %s", attachmentID, err)
			}

		case "SITE_TO_SITE_VPN":
			if _, err := waitVpnAttachmentAvailable(ctx, conn, attachmentID, d.Timeout(schema.TimeoutCreate)); err != nil {
				return diag.Errorf("waiting for Network Manager VPN Attachment (%s) create: %s", attachmentID, err)
			}
		}
	}

	return resourceAttachmentAccepterRead(ctx, d, meta)
}

func resourceAttachmentAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn

	switch aType := d.Get("attachment_type"); aType {
	case "VPC":
		vpcAttachment, err := FindVPCAttachmentByID(ctx, conn, d.Id())

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] Network Manager VPC Attachment %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		if err != nil {
			return diag.Errorf("reading Network Manager VPC Attachment (%s): %s", d.Id(), err)
		}

		a := vpcAttachment.Attachment
		d.Set("attachment_policy_rule_number", a.AttachmentPolicyRuleNumber)
		d.Set("core_network_arn", a.CoreNetworkArn)
		d.Set("core_network_id", a.CoreNetworkId)
		d.Set("edge_location", a.EdgeLocation)
		d.Set("owner_account_id", a.OwnerAccountId)
		d.Set("resource_arn", a.ResourceArn)
		d.Set("segment_name", a.SegmentName)
		d.Set("state", a.State)

	case "SITE_TO_SITE_VPN":
		vpnAttachment, err := FindVpnAttachmentByID(ctx, conn, d.Id())

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] Network Manager VPC Attachment %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		if err != nil {
			return diag.Errorf("reading Network Manager VPN Attachment (%s): %s", d.Id(), err)
		}

		a := vpnAttachment.Attachment
		d.Set("attachment_policy_rule_number", a.AttachmentPolicyRuleNumber)
		d.Set("core_network_arn", a.CoreNetworkArn)
		d.Set("core_network_id", a.CoreNetworkId)
		d.Set("edge_location", a.EdgeLocation)
		d.Set("owner_account_id", a.OwnerAccountId)
		d.Set("resource_arn", a.ResourceArn)
		d.Set("segment_name", a.SegmentName)
		d.Set("state", a.State)
	}

	return nil
}
