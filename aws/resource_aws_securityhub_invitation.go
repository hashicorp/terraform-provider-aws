package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	SecurityHubMemberStatusAssociated = "Associated"
	SecurityHubMemberStatusInvited    = "Invited"
	SecurityHubMemberStatusRemoved    = "Removed"
)

func resourceAwsSecurityHubInvitation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubInvitationCreate,
		Read:   resourceAwsSecurityHubInvitationRead,
		Delete: resourceAwsSecurityHubInvitationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"master_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsSecurityHubInvitationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Creating Security Hub invitation %s", d.Get("account_id").(string))

	resp, err := conn.InviteMembers(&securityhub.InviteMembersInput{
		AccountIds: []*string{aws.String(d.Get("account_id").(string))},
	})

	if err != nil {
		return fmt.Errorf("Error creating Security Hub invitation: %s", err)
	}

	if len(resp.UnprocessedAccounts) > 0 {
		return fmt.Errorf("Error creating Security Hub invitation: UnprocessedAccounts is not empty")
	}

	d.SetId(d.Get("account_id").(string))

	return resourceAwsSecurityHubInvitationRead(d, meta)
}

func resourceAwsSecurityHubInvitationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	log.Printf("[DEBUG] Reading Security Hub member %s", d.Id())
	resp, err := conn.GetMembers(&securityhub.GetMembersInput{
		AccountIds: []*string{aws.String(d.Id())},
	})

	if err != nil {
		if isAWSErr(err, securityhub.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Security Hub member (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if len(resp.Members) == 0 {
		log.Printf("[WARN] Security Hub member (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	member := resp.Members[0]

	// if member.Status != "" {
	// 	log.Printf("[WARN] Security Hub member (%s) not invited/accepted, removing from state", d.Id())
	// 	d.SetId("")
	// 	return nil
	// }

	d.Set("account_id", member.AccountId)
	d.Set("master_id", member.MasterId)

	return nil

}

func resourceAwsSecurityHubInvitationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Deleting Security Hub invitation: %s", d.Id())

	_, err := conn.DeleteMembers(&securityhub.DeleteMembersInput{
		AccountIds: []*string{aws.String(d.Id())},
	})

	if err != nil {
		if isAWSErr(err, securityhub.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Security Hub invitation (%s) not found", d.Id())
			return nil
		}
		return err
	}

	return nil
}
