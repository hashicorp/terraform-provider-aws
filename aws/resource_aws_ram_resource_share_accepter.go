package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsRamResourceShareAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRamResourceShareAccepterCreate,
		Read:   resourceAwsRamResourceShareAccepterRead,
		Delete: resourceAwsRamResourceShareAccepterDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"share_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"invitation_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"share_id": {
				Type:     schema.TypeString,
				Computed: true,
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

			"share_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceAwsRamResourceShareAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	shareARN := d.Get("share_arn").(string)

	invitation, err := resourceAwsRamResourceShareGetInvitation(conn, shareARN, ram.ResourceShareInvitationStatusPending)

	if err != nil {
		return err
	}

	if invitation == nil || aws.StringValue(invitation.ResourceShareInvitationArn) == "" {
		return fmt.Errorf(
			"No RAM Resource Share (%s) invitation found\n\n"+
				"NOTE: If both AWS accounts are in the same AWS Organization and RAM Sharing with AWS Organizations is enabled, this resource is not necessary",
			shareARN)
	}

	input := &ram.AcceptResourceShareInvitationInput{
		ClientToken:                aws.String(resource.UniqueId()),
		ResourceShareInvitationArn: invitation.ResourceShareInvitationArn,
	}

	log.Printf("[DEBUG] Accept RAM resource share invitation request: %s", input)
	output, err := conn.AcceptResourceShareInvitation(input)

	if err != nil {
		return fmt.Errorf("Error accepting RAM resource share invitation: %s", err)
	}

	d.SetId(shareARN)

	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareInvitationStatusPending},
		Target:  []string{ram.ResourceShareInvitationStatusAccepted},
		Refresh: resourceAwsRamResourceShareAccepterStateRefreshFunc(
			conn,
			aws.StringValue(output.ResourceShareInvitation.ResourceShareInvitationArn)),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	_, err = stateConf.WaitForState()

	if err != nil {
		return fmt.Errorf("Error waiting for RAM resource share (%s) state: %s", d.Id(), err)
	}

	return resourceAwsRamResourceShareAccepterRead(d, meta)
}

func resourceAwsRamResourceShareAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	invitation, err := resourceAwsRamResourceShareGetInvitation(conn, d.Id(), ram.ResourceShareInvitationStatusAccepted)

	if err == nil && invitation == nil {
		log.Printf("[WARN] No RAM resource share invitation by ARN (%s) found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	d.Set("status", invitation.Status)
	d.Set("receiver_account_id", invitation.ReceiverAccountId)
	d.Set("sender_account_id", invitation.SenderAccountId)
	d.Set("share_arn", invitation.ResourceShareArn)
	d.Set("invitation_arn", invitation.ResourceShareInvitationArn)
	d.Set("share_id", resourceAwsRamResourceShareGetIDFromARN(d.Id()))
	d.Set("share_name", invitation.ResourceShareName)

	listInput := &ram.ListResourcesInput{
		MaxResults:        aws.Int64(int64(500)),
		ResourceOwner:     aws.String(ram.ResourceOwnerOtherAccounts),
		ResourceShareArns: aws.StringSlice([]string{d.Id()}),
		Principal:         invitation.SenderAccountId,
	}

	var resourceARNs []*string
	err = conn.ListResourcesPages(listInput, func(page *ram.ListResourcesOutput, lastPage bool) bool {
		for _, resource := range page.Resources {
			resourceARNs = append(resourceARNs, resource.Arn)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("Error reading RAM resource share resources %s: %s", d.Id(), err)
	}

	if err := d.Set("resources", flattenStringList(resourceARNs)); err != nil {
		return fmt.Errorf("unable to set resources: %s", err)
	}

	return nil
}

func resourceAwsRamResourceShareAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	receiverAccountID := d.Get("receiver_account_id").(string)

	if receiverAccountID == "" {
		return fmt.Errorf("The receiver account ID is required to leave a resource share")
	}

	input := &ram.DisassociateResourceShareInput{
		ClientToken:      aws.String(resource.UniqueId()),
		ResourceShareArn: aws.String(d.Id()),
		Principals:       []*string{aws.String(receiverAccountID)},
	}
	log.Printf("[DEBUG] Leave RAM resource share request: %s", input)

	_, err := conn.DisassociateResourceShare(input)

	if err != nil {
		return fmt.Errorf("Error leaving RAM resource share: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociated},
		Target:  []string{ram.ResourceShareAssociationStatusDisassociated},
		Refresh: resourceAwsRamResourceShareStateRefreshFunc(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutDelete),
	}

	if _, err := stateConf.WaitForState(); err != nil {
		if isAWSErr(err, ram.ErrCodeUnknownResourceException, "") {
			// what we want
			return nil
		}
		return fmt.Errorf("Error waiting for RAM resource share (%s) state: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsRamResourceShareGetInvitation(conn *ram.RAM, resourceShareARN, status string) (*ram.ResourceShareInvitation, error) {
	input := &ram.GetResourceShareInvitationsInput{
		ResourceShareArns: []*string{aws.String(resourceShareARN)},
	}

	var invitation *ram.ResourceShareInvitation
	err := conn.GetResourceShareInvitationsPages(input, func(page *ram.GetResourceShareInvitationsOutput, lastPage bool) bool {
		for _, rsi := range page.ResourceShareInvitations {
			if aws.StringValue(rsi.Status) == status {
				invitation = rsi
				return false
			}
		}

		return !lastPage
	})

	if invitation == nil {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("Error reading RAM resource share invitation %s: %s", resourceShareARN, err)
	}

	return invitation, nil
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

func resourceAwsRamResourceShareGetIDFromARN(arn string) string {
	return strings.Replace(arn[strings.LastIndex(arn, ":")+1:], "resource-share/", "rs-", -1)
}
