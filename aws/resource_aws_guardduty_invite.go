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
	detectorId := d.Get("detector_id").(string)
	accountIds := d.Get("account_ids").([]string)

	params := &guardduty.InviteMembersInput{
		DetectorId: aws.String(detectorId),
		AccountIds: aws.StringSlice(accountIds),
		Message: aws.String(d.Get("message").(string)),
	}

	log.Printf("[DEBUG] GuardDuty Invite Members: %#v", params)
	resp, err := conn.InviteMembers(params)

	if err != nil {
		return fmt.Errorf("Inviting GuardDuty Member failed: %s", err)
	}

	d.SetId("")

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

	return nil
}

func resourceAwsGuardDutyInviteRead(d *schema.ResourceData, meta interface{}) error {
	//TODO Do something?
	return nil
}

func resourceAwsGuardDutyInviteDelete(d* schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}