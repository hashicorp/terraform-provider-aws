package networkmanager

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceAttachmentAcceptor() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: ResourceAttachmentAcceptorCreate,
		ReadWithoutTimeout:   schema.NoopContext,
		DeleteWithoutTimeout: ResourceAttachmentAcceptorDelete,

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

func ResourceAttachmentAcceptorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn

	attachmentId := d.Get("attachment_id").(string)

	input := &networkmanager.AcceptAttachmentInput{
		AttachmentId: aws.String(attachmentId),
	}

	log.Printf("[DEBUG] Accepting Network Manager VPC Attachment: %s", input)
	a, err := conn.AcceptAttachmentWithContext(ctx, input)

	if err != nil {
		output, err := FindVpcAttachmentByID(ctx, conn, attachmentId)

		state := aws.StringValue(output.Attachment.State)

		if state == networkmanager.AttachmentStateAvailable {
			log.Printf("[WARN] Attachment (%s) already accepted, importing attributes into state without accepting.", d.Id())
			a.Attachment = output.Attachment
		} else {
			return diag.Errorf("error accepting Network Manager VPC Attachment: %s", err)
		}
	}

	d.SetId(attachmentId)
	if _, err := waitVpcAttachmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		d.SetId("")
		return diag.Errorf("error waiting for Network Manager VPC Attachment (%s) create: %s", d.Id(), err)
	}

	d.Set("core_network_id", a.Attachment.CoreNetworkId)
	d.Set("state", a.Attachment.State)
	d.Set("core_network_arn", a.Attachment.CoreNetworkArn)
	d.Set("attachment_policy_rule_number", a.Attachment.AttachmentPolicyRuleNumber)
	d.Set("attachment_type", a.Attachment.AttachmentType)
	d.Set("edge_location", a.Attachment.EdgeLocation)
	d.Set("owner_account_id", a.Attachment.OwnerAccountId)
	d.Set("resource_arn", a.Attachment.ResourceArn)
	d.Set("segment_name", a.Attachment.SegmentName)

	return nil
}

func ResourceAttachmentAcceptorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[WARN] Attachment (%s) not deleted, removing from state.", d.Id())

	return nil
}

func statusAttachmentAcceptorState(ctx context.Context, conn *networkmanager.NetworkManager, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVpcAttachmentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Attachment.State), nil
	}
}

func waitAttachmentAcceptorCreated(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.VpcAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.AttachmentStatePendingAttachmentAcceptance, networkmanager.AttachmentStatePendingNetworkUpdate},
		Target:  []string{networkmanager.AttachmentStateAvailable},
		Timeout: timeout,
		Refresh: statusVpcAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}
