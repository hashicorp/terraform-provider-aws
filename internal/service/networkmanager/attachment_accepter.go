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
)

// AttachmentAccepter is not specific to AttachmentType. However, querying attachments for status updates is
// To facilitate querying and waiters on specific attachment types, attachment_type required

func ResourceAttachmentAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: ResourceAttachmentAccepterCreate,
		ReadWithoutTimeout:   schema.NoopContext,
		DeleteWithoutTimeout: ResourceAttachmentAccepterDelete,

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
				ValidateFunc: validation.StringInSlice([]string{networkmanager.AttachmentTypeVpc}, false),
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

func ResourceAttachmentAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn

	attachmentId := d.Get("attachment_id").(string)
	attachmentType := d.Get("attachment_type").(string)
	attachment := &networkmanager.Attachment{}
	accepted := false
	var state string

	if attachmentType == networkmanager.AttachmentTypeVpc {
		output, err := FindVPCAttachmentByID(ctx, conn, attachmentId)

		if err != nil {
			return diag.Errorf("error finding Network Manager VPC Attachment: %s", err)
		}

		state = aws.StringValue(output.Attachment.State)
		attachment = output.Attachment
	}

	if state == networkmanager.AttachmentStateAvailable {
		accepted = true
		log.Printf("[WARN] Attachment (%s) already accepted, importing attributes into state without accepting.", attachmentId)
	}

	if !accepted {
		input := &networkmanager.AcceptAttachmentInput{
			AttachmentId: aws.String(attachmentId),
		}

		log.Printf("[DEBUG] Accepting Network Manager Attachment: %s", input)
		a, err := conn.AcceptAttachmentWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("error accepting Network Manager Attachment: %s", err)
		}

		attachment = a.Attachment
	}

	d.SetId(attachmentId)

	if attachmentType == networkmanager.AttachmentTypeVpc {
		if _, err := WaitVPCAttachmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			d.SetId("")
			return diag.Errorf("error waiting for Network Manager VPC Attachment (%s) create: %s", d.Id(), err)
		}
	}

	d.Set("core_network_id", attachment.CoreNetworkId)
	d.Set("state", attachment.State)
	d.Set("core_network_arn", attachment.CoreNetworkArn)
	d.Set("attachment_policy_rule_number", attachment.AttachmentPolicyRuleNumber)
	d.Set("edge_location", attachment.EdgeLocation)
	d.Set("owner_account_id", attachment.OwnerAccountId)
	d.Set("resource_arn", attachment.ResourceArn)
	d.Set("segment_name", attachment.SegmentName)

	return nil
}

func ResourceAttachmentAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[WARN] Attachment (%s) not deleted, removing from state.", d.Id())

	return nil
}
