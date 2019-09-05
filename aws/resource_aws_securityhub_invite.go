package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSecurityHubInvite() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubInviteCreate,
		Read:   resourceAwsSecurityHubInviteRead,
		Delete: resourceAwsSecurityHubInviteDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsSecurityHubInviteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	accountId := d.Get("account_id").(string)

	log.Printf("[DEBUG] Inviting %s to Security Hub", accountId)

	_, err := conn.InviteMembers(&securityhub.InviteMembersInput{
		AccountIds: []*string{
			aws.String(accountId),
		},
	})

	if err != nil {
		return fmt.Errorf("Error inviting %s to Security Hub: %s", accountId, err)
	}

	d.SetId(accountId)

	return resourceAwsSecurityHubInviteRead(d, meta)
}

func resourceAwsSecurityHubInviteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	log.Printf("[DEBUG] Retrieving members of Security Hub for account %s", d.Id())
	membersResp, err := conn.GetMembers(&securityhub.GetMembersInput{
		AccountIds: []*string{
			aws.String(d.Id()),
		},
	})

	if err != nil {
		return fmt.Errorf("Error retrieving members of Security Hub for account %s: %s", d.Id(), err)
	}

	// This means that this account is not associated anymore
	if len(membersResp.Members) == 0 {
		d.SetId("")
	}

	return nil
}

func resourceAwsSecurityHubInviteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Deleting invitation for %s from Security Hub", d.Id())

	_, err := conn.DeleteInvitations(&securityhub.DeleteInvitationsInput{
		AccountIds: []*string{
			aws.String(d.Id()),
		},
	})

	if err != nil {
		return fmt.Errorf("Error deleting invitation for %s from Security Hub: %s", d.Id(), err)
	}

	return nil
}
