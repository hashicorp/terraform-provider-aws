package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSecurityHubAcceptInvitation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubAcceptInvitationCreate,
		Read:   resourceAwsSecurityHubAcceptInvitationRead,
		Delete: resourceAwsSecurityHubAcceptInvitationDelete,

		Schema: map[string]*schema.Schema{
			"master_account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsSecurityHubAcceptInvitationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	log.Print("[DEBUG] Retrieving list of invitations")

	resp, err := conn.ListInvitations(&securityhub.ListInvitationsInput{})

	if err != nil {
		return fmt.Errorf("Error retrieving Security Hub invitations list: %s", err)
	}

	masterAccountId := d.Get("master_account_id").(string)

	log.Printf("[DEBUG] Accepting invitation to Security Hub from %s", masterAccountId)

	for i := range resp.Invitations {
		if *resp.Invitations[i].AccountId == masterAccountId {
			_, err := conn.AcceptInvitation(&securityhub.AcceptInvitationInput{
				MasterId:     aws.String(masterAccountId),
				InvitationId: resp.Invitations[i].InvitationId,
			})

			if err != nil {
				return fmt.Errorf("Error accepting invite to Security Hub from %s: %s", masterAccountId, err)
			}

			d.SetId(*resp.Invitations[i].InvitationId)
		}
	}

	return nil
}

func resourceAwsSecurityHubAcceptInvitationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	log.Print("[DEBUG] Retrieving list of invitations")

	resp, err := conn.ListInvitations(&securityhub.ListInvitationsInput{})

	if err != nil {
		return fmt.Errorf("Error retrieving Security Hub invitations list: %s", err)
	}

	for i := range resp.Invitations {
		if *resp.Invitations[i].AccountId == d.Get("master_account_id").(string) {
			d.SetId(*resp.Invitations[i].InvitationId)
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceAwsSecurityHubAcceptInvitationDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Will not delete Security Hub invitation. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}
