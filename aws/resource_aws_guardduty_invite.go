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
			"unprocessed_accounts": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func resourceAwsGuardDutyInviteCreate(d* schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn
	detectorId := d.Get("detector_id").(string)

	params := &guardduty.InviteMembersInput{
		DetectorId: aws.String(detectorId),
		AccountIds: aws.StringSlice(d.Get("account_ids").([]string)),
		Message: aws.String(d.Get("message").(string)),
	}

	log.Printf("[DEBUG] GuardDuty Invite Members: %#v", params)
	resp, err := conn.InviteMembers(params)

	if err != nil {
		return fmt.Errorf("Inviting GuardDuty Member failed: %s", err)
	}

	unprocessedAccounts := make(map[string]string, len(resp.UnprocessedAccounts))
	for _, unprocessedAccount := range resp.UnprocessedAccounts {
		unprocessedAccounts[*unprocessedAccount.AccountId] = *unprocessedAccount.Result
	}

	d.SetId(detectorId)
	d.Set("unprocessed_accounts", unprocessedAccounts)

	return nil
}

func resourceAwsGuardDutyInviteRead(d *schema.ResourceData, meta interface{}) error {
	//TODO Do something?
	return nil
}

func resourceAwsGuardDutyInviteDelete(d* schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	params := &guardduty.DeleteInvitationsInput{
		AccountIds: aws.StringSlice(d.Get("account_ids").([]string)),
	}

	log.Printf("[DEBUG] GuardDuty Delete Invitations: %#v", params)
	_, err := conn.DeleteInvitations(params)

	if err != nil {
		return fmt.Errorf("Deleting GuardDuty Invitations '%s' failed: %s", d.Id(), err.Error())
	}

	return nil
}