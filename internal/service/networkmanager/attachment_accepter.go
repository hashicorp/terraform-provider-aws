package networkmanager

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					networkmanager.AttachmentTypeVpc,
				}, false),
				// Implement Values() function for validation as more types are onboarded to provider
				// networkmanager.AttachmentType_Values(), false),
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

	if attachmentType := d.Get("attachment_type").(string); attachmentType != networkmanager.AttachmentTypeVpc {
		return diag.Errorf("unsupported Network Manager Attachment type: %s", attachmentType)
	}

	attachmentID := d.Get("attachment_id").(string)
	vpcAttachment, err := FindVPCAttachmentByID(ctx, conn, attachmentID)

	if err != nil {
		return diag.Errorf("reading Network Manager VPC Attachment (%s): %s", attachmentID, err)
	}

	if state := aws.StringValue(vpcAttachment.Attachment.State); state == networkmanager.AttachmentStatePendingAttachmentAcceptance || state == networkmanager.AttachmentStatePendingTagAcceptance {
		input := &networkmanager.AcceptAttachmentInput{
			AttachmentId: aws.String(attachmentID),
		}

		_, err := conn.AcceptAttachmentWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("accepting Network Manager Attachment (%s): %s", attachmentID, err)
		}

		if _, err := waitVPCAttachmentCreated(ctx, conn, attachmentID, d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.Errorf("waiting for Network Manager VPC Attachment (%s) create: %s", attachmentID, err)
		}
	}

	d.SetId(attachmentID)

	return resourceAttachmentAccepterRead(ctx, d, meta)
}

func resourceAttachmentAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn

	vpcAttachment, err := FindVPCAttachmentByID(ctx, conn, d.Id())

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

	return nil
}
