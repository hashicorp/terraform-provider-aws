package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsGuardDutyMember() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyMemberCreate,
		Read:   resourceAwsGuardDutyMemberRead,
		Delete: resourceAwsGuardDutyMemberDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"relationship_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invite": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"disable_email_notification": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"invitation_message": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Second),
		},
	}
}

func resourceAwsGuardDutyMemberCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn
	accountID := d.Get("account_id").(string)
	detectorID := d.Get("detector_id").(string)

	input := guardduty.CreateMembersInput{
		AccountDetails: []*guardduty.AccountDetail{{
			AccountId: aws.String(accountID),
			Email:     aws.String(d.Get("email").(string)),
		}},
		DetectorId: aws.String(detectorID),
	}

	log.Printf("[DEBUG] Creating GuardDuty Member: %s", input)
	_, err := conn.CreateMembers(&input)
	if err != nil {
		return fmt.Errorf("Creating GuardDuty Member failed: %s", err.Error())
	}

	d.SetId(fmt.Sprintf("%s:%s", detectorID, accountID))

	if !d.Get("invite").(bool) {
		return resourceAwsGuardDutyMemberRead(d, meta)
	}

	imi := &guardduty.InviteMembersInput{
		DetectorId:               aws.String(detectorID),
		AccountIds:               []*string{aws.String(accountID)},
		DisableEmailNotification: aws.Bool(d.Get("disable_email_notification").(bool)),
		Message:                  aws.String(d.Get("invitation_message").(string)),
	}

	_, err = conn.InviteMembers(imi)
	if err != nil {
		return fmt.Errorf("Inviting GuardDuty Member failed: %s", err)
	}

	// wait until e-mail verification finishes
	if err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		input := guardduty.GetMembersInput{
			DetectorId: imi.DetectorId,
			AccountIds: imi.AccountIds,
		}

		log.Printf("[DEBUG] Reading GuardDuty Member: %s", input)
		gmo, err := conn.GetMembers(&input)

		if err != nil {
			if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
				log.Printf("[WARN] GuardDuty detector %q not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error reading GuardDuty Member %q: %s", d.Id(), err))
		}

		if gmo == nil || len(gmo.Members) == 0 {
			return resource.RetryableError(fmt.Errorf("error reading GuardDuty Member %q: member missing from response", d.Id()))
		}

		member := gmo.Members[0]
		status := aws.StringValue(member.RelationshipStatus)

		if status == "Disabled" || status == "Enabled" || status == "Invited" {
			return nil
		}

		if status == "Created" || status == "EmailVerificationInProgress" {
			return resource.RetryableError(fmt.Errorf("Expected member to be invited but was in state: %s", status))
		}

		return resource.NonRetryableError(fmt.Errorf("error inviting GuardDuty Member %q: invalid status: %s", d.Id(), status))
	}); err != nil {
		return err
	}

	return resourceAwsGuardDutyMemberRead(d, meta)
}

func resourceAwsGuardDutyMemberRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	accountID, detectorID, err := decodeGuardDutyMemberID(d.Id())
	if err != nil {
		return err
	}

	input := guardduty.GetMembersInput{
		AccountIds: []*string{aws.String(accountID)},
		DetectorId: aws.String(detectorID),
	}

	log.Printf("[DEBUG] Reading GuardDuty Member: %s", input)
	gmo, err := conn.GetMembers(&input)
	if err != nil {
		if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
			log.Printf("[WARN] GuardDuty detector %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading GuardDuty Member '%s' failed: %s", d.Id(), err.Error())
	}

	if gmo.Members == nil || (len(gmo.Members) < 1) {
		log.Printf("[WARN] GuardDuty Member %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	member := gmo.Members[0]
	d.Set("account_id", member.AccountId)
	d.Set("detector_id", detectorID)
	d.Set("email", member.Email)

	status := aws.StringValue(member.RelationshipStatus)
	d.Set("relationship_status", status)

	// https://docs.aws.amazon.com/guardduty/latest/ug/list-members.html
	d.Set("invite", false)
	if status == "Disabled" || status == "Enabled" || status == "Invited" || status == "EmailVerificationInProgress" {
		d.Set("invite", true)
	}

	return nil
}

func resourceAwsGuardDutyMemberDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	accountID, detectorID, err := decodeGuardDutyMemberID(d.Id())
	if err != nil {
		return err
	}

	input := guardduty.DeleteMembersInput{
		AccountIds: []*string{aws.String(accountID)},
		DetectorId: aws.String(detectorID),
	}

	log.Printf("[DEBUG] Delete GuardDuty Member: %s", input)
	_, err = conn.DeleteMembers(&input)
	if err != nil {
		return fmt.Errorf("Deleting GuardDuty Member '%s' failed: %s", d.Id(), err.Error())
	}
	return nil
}

func decodeGuardDutyMemberID(id string) (accountID, detectorID string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = fmt.Errorf("GuardDuty Member ID must be of the form <Detector ID>:<Member AWS Account ID>, was provided: %s", id)
		return
	}
	accountID = parts[1]
	detectorID = parts[0]
	return
}
