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
	detectorID := d.Get("detector_id").(string)
	accountIDs := d.Get("account_ids").([]string)

	input := guardduty.InviteMembersInput{
		DetectorId: aws.String(detectorID),
		AccountIds: aws.StringSlice(accountIDs),
		Message: aws.String(d.Get("message").(string)),
	}

	log.Printf("[DEBUG] Inviting GuardDuty Member: %s", input)
	imo, err := conn.InviteMembers(&input)
	if err != nil {
		return fmt.Errorf("Inviting GuardDuty Member failed: %s", err.Error())
	}

	if imo.UnprocessedAccounts != nil || len(imo.UnprocessedAccounts) > 0 {
		for _, unprocessedAccount := range imo.UnprocessedAccounts {
			log.Printf("[WARN] GuardDuty Members %q not processed: %s", unprocessedAccount.AccountId, unprocessedAccount.Result)
		}
	}

	for _, accountID := range accountIDs {
		d.SetId(fmt.Sprintf("%s:%s", detectorID, accountID))
		err := resourceAwsGuardDutyMemberRead(d, meta)
		if err != nil {
			return err
		}
	}

	return nil
}