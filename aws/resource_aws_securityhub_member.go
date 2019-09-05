package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSecurityHubMember() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubMemberCreate,
		Read:   resourceAwsSecurityHubMemberRead,
		Delete: resourceAwsSecurityHubMemberDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"master_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"member_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsSecurityHubMemberCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	accountId := d.Get("account_id").(string)

	log.Printf("[DEBUG] Adding %s to Security Hub", accountId)

	_, err := conn.CreateMembers(&securityhub.CreateMembersInput{
		AccountDetails: []*securityhub.AccountDetails{
			{
				AccountId: aws.String(accountId),
				Email:     aws.String(d.Get("email").(string)),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("Error adding %s to Security Hub: %s", accountId, err)
	}

	d.SetId(accountId)

	return resourceAwsSecurityHubMemberRead(d, meta)
}

func resourceAwsSecurityHubMemberRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	log.Printf("[DEBUG] Retrieving members of Security Hub for account %s", d.Id())
	resp, err := conn.GetMembers(&securityhub.GetMembersInput{
		AccountIds: []*string{
			aws.String(d.Id()),
		},
	})

	if err != nil {
		return fmt.Errorf("Error retrieving members of Security Hub for account %s: %s", d.Id(), err)
	}

	// This means that this account is not associated anymore
	if len(resp.Members) == 0 {
		d.SetId("")
		return nil
	}

	member := resp.Members[0]

	d.Set("master_id", member.MasterId)
	d.Set("member_status", member.MemberStatus)

	return nil
}

func resourceAwsSecurityHubMemberDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Deleting %s from Security Hub", d.Id())

	_, err := conn.DeleteMembers(&securityhub.DeleteMembersInput{
		AccountIds: []*string{
			aws.String(d.Id()),
		},
	})

	if err != nil {
		return fmt.Errorf("Error deleting %s from Security Hub: %s", d.Id(), err)
	}

	return nil
}
