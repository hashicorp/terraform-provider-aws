package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsRamResourceShareAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRamResourceShareAccepterCreate,
		Read:   resourceAwsRamResourceShareAccepterRead,
		Delete: resourceAwsRamResourceShareAccepterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"receiver_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"sender_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"share_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"share_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"client_token": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceAwsRamResourceShareAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	requestInput := &ram.AcceptResourceShareInvitationInput{
		ResourceShareInvitationArn: aws.String(d.Get("arn").(string)),
	}

	if v, ok := d.GetOk("client_token"); ok {
		requestInput.ClientToken = aws.String(v.(string))
	}

	log.Println("[DEBUG] Accept RAM resource share invitation request:", requestInput)
	requestOutput, err := conn.AcceptResourceShareInvitation(requestInput)
	if err != nil {
		return fmt.Errorf("Error accepting RAM resource share invitation: %s", err)
	}

	d.SetId(aws.StringValue(requestOutput.ResourceShareInvitation.ResourceShareInvitationArn))

	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareInvitationStatusPending},
		Target:  []string{ram.ResourceShareInvitationStatusAccepted},
		Refresh: resourceAwsRamResourceShareAccepterStateRefreshFunc(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for RAM resource share invitation (%s) state: %s", d.Id(), err)
	}

	return resourceAwsRamResourceShareAccepterRead(d, meta)
}

func resourceAwsRamResourceShareAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	request := &ram.GetResourceShareInvitationsInput{
		ResourceShareInvitationArns: []*string{aws.String(d.Id())},
	}

	output, err := conn.GetResourceShareInvitations(request)
	if err != nil {
		if isAWSErr(err, ram.ErrCodeUnknownResourceException, "") {
			log.Printf("[WARN] No RAM resource share invitation by ARN (%s) found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading RAM resource share invitation %s: %s", d.Id(), err)
	}

	if len(output.ResourceShareInvitations) == 0 {
		log.Printf("[WARN] No RAM resource share invitation by ARN (%s) found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	invitation := output.ResourceShareInvitations[0]

	d.Set("status", invitation.Status)
	d.Set("receiver_account_id", invitation.ReceiverAccountId)
	d.Set("sender_account_id", invitation.SenderAccountId)
	d.Set("share_arn", invitation.ResourceShareArn)
	d.Set("share_name", invitation.ResourceShareName)

	return nil
}

func resourceAwsRamResourceShareAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Will not delete resource share invitation. Terraform will remove this invitation from the state file. However, resources may remain.")
	return nil
}

func resourceAwsRamResourceShareAccepterStateRefreshFunc(conn *ram.RAM, invitationArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		request := &ram.GetResourceShareInvitationsInput{
			ResourceShareInvitationArns: []*string{aws.String(invitationArn)},
		}

		output, err := conn.GetResourceShareInvitations(request)

		if err != nil {
			return nil, "Unable to get resource share invitations", err
		}

		if len(output.ResourceShareInvitations) == 0 {
			return nil, "Resource share invitation not found", nil
		}

		invitation := output.ResourceShareInvitations[0]

		return invitation, aws.StringValue(invitation.Status), nil
	}
}
