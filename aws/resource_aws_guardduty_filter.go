package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsGuardDutyFilter() *schema.Resource {
	return &schema.Resource{
		// Create: resourceAwsGuardDutyFilterCreate,
		Read: resourceAwsGuardDutyFilterRead,
		// Update: resourceAwsGuardDutyFilterUpdate,
		Delete: resourceAwsGuardDutyFilterDelete,

		// Importer: &schema.ResourceImporter{
		// 	State: schema.ImportStatePassthrough,
		// },
		Schema: map[string]*schema.Schema{
			// "account_id": { // idk, do we need it
			// 	Type:         schema.TypeString,
			// 	Required:     true,
			// 	ForceNew:     true,
			// 	ValidateFunc: validateAwsAccountId,
			// },
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // perhaps remove here and below, when Update is back
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true, // perhaps remove here and below, when Update is back
			},
			// "tags": { // Must be added back
			// 	Type:     schema.TypeTags, // probably wrong type
			// 	Optional: true,
			// },
			// "findingCriteria": {
			// 	Type:     schema.TypeString, // need to implement a new type
			// 	Optional: true,              // change to required
			// 	ForceNew: true,              // perhaps remove here and below, when Update is back
			// },
			"action": {
				Type:     schema.TypeString, // should have a new type or a validation for NOOP/ARCHIVE
				Optional: true,
				ForceNew: true, // perhaps remove here and below, when Update is back
			},
			"rank": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true, // perhaps remove here and below, when Update is back
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Second),
			Update: schema.DefaultTimeout(60 * time.Second),
		},
	}
}

// func resourceAwsGuardDutyFilterCreate(d *schema.ResourceData, meta interface{}) error {
// 	conn := meta.(*AWSClient).guarddutyconn
// 	accountID := d.Get("account_id").(string)
// 	detectorID := d.Get("detector_id").(string)
//
// 	input := guardduty.CreateMembersInput{
// 		AccountDetails: []*guardduty.AccountDetail{{
// 			AccountId: aws.String(accountID),
// 			Email:     aws.String(d.Get("email").(string)),
// 		}},
// 		DetectorId: aws.String(detectorID),
// 	}
//
// 	log.Printf("[DEBUG] Creating GuardDuty Member: %s", input)
// 	_, err := conn.CreateMembers(&input)
// 	if err != nil {
// 		return fmt.Errorf("Creating GuardDuty Member failed: %s", err.Error())
// 	}
//
// 	d.SetId(fmt.Sprintf("%s:%s", detectorID, accountID))
//
// 	if !d.Get("invite").(bool) {
// 		return resourceAwsGuardDutyFilterRead(d, meta)
// 	}
//
// 	imi := &guardduty.InviteMembersInput{
// 		DetectorId:               aws.String(detectorID),
// 		AccountIds:               []*string{aws.String(accountID)},
// 		DisableEmailNotification: aws.Bool(d.Get("disable_email_notification").(bool)),
// 		Message:                  aws.String(d.Get("invitation_message").(string)),
// 	}
//
// 	log.Printf("[INFO] Inviting GuardDuty Member: %s", input)
// 	_, err = conn.InviteMembers(imi)
// 	if err != nil {
// 		return fmt.Errorf("error inviting GuardDuty Member %q: %s", d.Id(), err)
// 	}
//
// 	err = inviteGuardDutyMemberWaiter(accountID, detectorID, d.Timeout(schema.TimeoutUpdate), conn)
// 	if err != nil {
// 		return fmt.Errorf("error waiting for GuardDuty Member %q invite: %s", d.Id(), err)
// 	}
//
// 	return resourceAwsGuardDutyFilterRead(d, meta)
// }

func resourceAwsGuardDutyFilterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn
	detectorId := d.Get("detectorId").(string)
	filterName := d.Get("filterName").(string)

	input := guardduty.GetFilterInput{
		DetectorId: aws.String(detectorId),
		FilterName: aws.String(filterName),
	}

	log.Printf("[DEBUG] Reading GuardDuty Filter: %s", input)
	filter, err := conn.GetFilter(&input)

	if err != nil {
		if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
			log.Printf("[WARN] GuardDuty detector %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading GuardDuty Filter '%s' failed: %s", filterName, err.Error())
	}

	d.Set("account_id", filter.Action)
	d.Set("account_id", filter.Description)
	d.Set("account_id", filter.Name)
	d.Set("account_id", filter.Rank)
	d.Set("detector_id", d.Id())

	// need to find a way how to fill it interface{}
	// d.Set("account_id", filter.FindingCriteria)

	// FindingCriteria.Criterion
	// Eq
	// Gt
	// Gte
	// Lt
	// Lte
	// Neq

	return nil
}

// func resourceAwsGuardDutyFilterUpdate(d *schema.ResourceData, meta interface{}) error {
// 	conn := meta.(*AWSClient).guarddutyconn
//
// 	accountID, detectorID, err := decodeGuardDutyMemberID(d.Id())
// 	if err != nil {
// 		return err
// 	}
//
// 	if d.HasChange("invite") {
// 		if d.Get("invite").(bool) {
// 			input := &guardduty.InviteMembersInput{
// 				DetectorId:               aws.String(detectorID),
// 				AccountIds:               []*string{aws.String(accountID)},
// 				DisableEmailNotification: aws.Bool(d.Get("disable_email_notification").(bool)),
// 				Message:                  aws.String(d.Get("invitation_message").(string)),
// 			}
//
// 			log.Printf("[INFO] Inviting GuardDuty Member: %s", input)
// 			output, err := conn.InviteMembers(input)
// 			if err != nil {
// 				return fmt.Errorf("error inviting GuardDuty Member %q: %s", d.Id(), err)
// 			}
//
// 			// {"unprocessedAccounts":[{"result":"The request is rejected because the current account has already invited or is already the GuardDuty master of the given member account ID.","accountId":"067819342479"}]}
// 			if len(output.UnprocessedAccounts) > 0 {
// 				return fmt.Errorf("error inviting GuardDuty Member %q: %s", d.Id(), aws.StringValue(output.UnprocessedAccounts[0].Result))
// 			}
// 		} else {
// 			input := &guardduty.DisassociateMembersInput{
// 				AccountIds: []*string{aws.String(accountID)},
// 				DetectorId: aws.String(detectorID),
// 			}
// 			log.Printf("[INFO] Disassociating GuardDuty Member: %s", input)
// 			_, err := conn.DisassociateMembers(input)
// 			if err != nil {
// 				return fmt.Errorf("error disassociating GuardDuty Member %q: %s", d.Id(), err)
// 			}
// 		}
// 	}
//
// 	return resourceAwsGuardDutyFilterRead(d, meta)
// }

func resourceAwsGuardDutyFilterDelete(d *schema.ResourceData, meta interface{}) error {
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
