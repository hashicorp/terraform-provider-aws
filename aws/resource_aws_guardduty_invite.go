package aws

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"log"
)

func resourceAwsGuardDutyInvite() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyInviteCreate,
		Read: resourceAwsGuardDutyInviteRead,
		Delete: resourceAwsGuardDutyInviteDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account_ids": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
			},
			"message": {
				Type:         schema.TypeString,
				Required:     false,
			},
		},
	}
}

func resourceAwsGuardDutyInviteCreate(d* schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	params := &guardduty.InviteMembersInput{
		DetectorId: aws.String(d.Get("detector_id").(string)),
		Message: aws.String(d.Get("message").(string)),
	}

	accountIds := d.Get("account_ids").([]string)
	params.AccountIds = aws.StringSlice(accountIds)

	log.Printf("[DEBUG] GuardDuty Invite Members: %#v", params)
	resp, err := conn.InviteMembers(params)

	if err != nil {
		return fmt.Errorf("Inviting GuardDuty Member failed: %s", err)
	}

	for _, accountID := range accountIds {
		for _, unprocessedAccount := range resp.UnprocessedAccounts {
			if accountID == *unprocessedAccount.AccountId {
				d.Set("unprocessed_reason", unprocessedAccount.Result)
			}
		}

		err := resourceAwsGuardDutyMemberRead(d, meta)

		if err != nil {
			return err
		}
	}
	d.SetId()

	return nil
}

func resourceAwsGuardDutyInviteRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsGuardDutyInviteDelete(d* schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}