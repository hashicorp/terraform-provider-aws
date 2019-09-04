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
		Create: resourceAwsGuardDutyFilterCreate,
		Read:   resourceAwsGuardDutyFilterRead,
		// Update: resourceAwsGuardDutyFilterUpdate,
		Delete: resourceAwsGuardDutyFilterDelete,

		// Importer: &schema.ResourceImporter{
		// 	State: schema.ImportStatePassthrough,
		// },
		Schema: map[string]*schema.Schema{ // TODO: add validations
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
			"finding_criteria": {
				Type:     schema.TypeList, // Probably need to use FindingCriteria type
				MaxItems: 1,
				Required: true, // change to required
				ForceNew: true, // perhaps remove here and below, when Update is back
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"criterion": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"condition": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"greater_than": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"greater_than_or_equal": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"less_than": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"less_than_or_equal": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"not_equals": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
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

func resourceAwsGuardDutyFilterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := guardduty.CreateFilterInput{
		DetectorId:  aws.String(d.Get("detector_id").(string)),
		Name:        aws.String(d.Get("name").(string)),
		Description: aws.String(d.Get("description").(string)),
		Rank:        aws.Int64(int64(d.Get("rank").(int))),
	}

	// building FindingCriteria
	findingCriteria := d.Get("finding_criteria").([]interface{})[0].(map[string]interface{})
	criterion := findingCriteria["criterion"].(*schema.Set).List()[0].(map[string]interface{})
	condition := criterion["condition"].(*schema.Set).List()[0].(map[string]interface{})

	interfaceForEquals := condition["equals"].([]interface{})

	equals := make([]string, len(interfaceForEquals))
	for i, v := range interfaceForEquals {
		new[i] = string(v.(string))
	}

	interfaceForNotEquals := condition["equals"].([]interface{})

	notEquals := make([]string, len(interfaceForNotEquals))
	for i, v := range interfaceForNotEquals {
		new[i] = string(v.(string))
	}

	input.FindingCriteria = &guardduty.FindingCriteria{
		Criterion: map[string]*guardduty.Condition{ // with star or without, that't the question!
			"condition": &guardduty.Condition{
				Equals:             aws.StringSlice(equals),
				GreaterThan:        aws.Int64(int64(condition["greater_than"].(int))),
				GreaterThanOrEqual: aws.Int64(int64(condition["greater_than_or_equal"].(int))),
				LessThan:           aws.Int64(int64(condition["less_than"].(int))),
				LessThanOrEqual:    aws.Int64(int64(condition["less_than_or_equal"].(int))),
				NotEquals:          aws.StringSlice(notEquals), //aws.StringSlice([]string(condition["equals"].([]interface{}))),
			},
		},
	}
	log.Printf("[DEBUG] Creating FindingCriteria map: %#v", findingCriteria)

	// Setting the default value for `action`
	action := "NOOP"

	if len(d.Get("action").(string)) > 0 {
		action = d.Get("action").(string)
	}

	input.Action = aws.String(action)

	log.Printf("[DEBUG] Creating GuardDuty Filter: %s", input)
	output, err := conn.CreateFilter(&input)
	if err != nil {
		return fmt.Errorf("Creating GuardDuty Filter failed: %s", err.Error())
	}
	d.SetId(*output.Name)

	return resourceAwsGuardDutyFilterRead(d, meta)
}

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

	d.Set("action", filter.Action)
	d.Set("description", filter.Description)
	d.Set("rank", filter.Rank)
	d.Set("name", d.Id())

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
