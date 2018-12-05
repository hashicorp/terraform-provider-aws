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
		},
	}
}

func resourceAwsSecurityHubMemberCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Creating Security Hub member %s", d.Get("rest_api_id").(string))

	resp, err := conn.CreateMembers(&securityhub.CreateMembersInput{
		AccountDetails: []*securityhub.AccountDetails{
			{
				AccountId: aws.String(d.Get("account_id").(string)),
				Email:     aws.String(d.Get("email").(string)),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("Error creating Security Hub member: %s", err)
	}

	if len(resp.UnprocessedAccounts) > 0 {
		return fmt.Errorf("Error creating Security Hub member: UnprocessedAccounts is not empty")
	}

	d.SetId(d.Get("account_id").(string))

	return resourceAwsSecurityHubMemberRead(d, meta)
}

func resourceAwsSecurityHubMemberRead(d *schema.ResourceData, meta interface{}) error {
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

	d.Set("account_id", member.AccountId)
	d.Set("email", member.Email)

	return nil
}

func resourceAwsSecurityHubMemberDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Deleting Security Hub member: %s", d.Id())

	_, err := conn.DeleteMembers(&securityhub.DeleteMembersInput{
		AccountIds: []*string{aws.String(d.Id())},
	})

	if err != nil {
		if isAWSErr(err, securityhub.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Security Hub member (%s) not found", d.Id())
			return nil
		}
		return err
	}

	return nil
}
