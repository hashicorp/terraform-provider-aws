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
			"tags": tagsSchemaForceNew(),
			"finding_criteria": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true, // perhaps remove here and below, when Update is back
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"criterion": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"field": {
										Type:     schema.TypeString,
										Required: true,
										// ValidateFunc: validation.StringInSlice([]string{
										// 	"region"
										// }, false),
									},
									"condition": {
										Type:     schema.TypeString,
										Required: true,
									},
									"values": {
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
			"action": {
				Type:     schema.TypeString, // should have a new type or a validation for NOOP/ARCHIVE
				Optional: true,
				ForceNew: true, // perhaps remove here and below, when Update is back
			},
			"rank": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true, // perhaps remove here and below, when Update is back
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Second),
			Update: schema.DefaultTimeout(60 * time.Second),
		},
	}
}

//noinspection GoMissingReturn
func buildFindingCriteria(findingCriteria map[string]interface{}) *guardduty.FindingCriteria {
	// 	criteriaMap := map[string][]string{
	// 		"confidence": {"equals", "not_equals", "greater_than", "greater_than_or_equal", "less_than", "less_than_or_equal"},
	// 		"id":         {"equals", "not_equals", "greater_than", "greater_than_or_equal", "less_than", "less_than_or_equal"},
	// 		"account_id": {"equals", "not_equals"},
	// 		"region":     {"equals", "not_equals"},
	// 		"resource.accessKeyDetails.accessKeyId":                                          {"equals", "not_equals"},
	// 		"resource.accessKeyDetails.principalId":                                          {"equals", "not_equals"},
	// 		"resource.accessKeyDetails.userName":                                             {"equals", "not_equals"},
	// 		"resource.accessKeyDetails.userType":                                             {"equals", "not_equals"},
	// 		"resource.instanceDetails.iamInstanceProfile.id":                                 {"equals", "not_equals"},
	// 		"resource.instanceDetails.imageId":                                               {"equals", "not_equals"},
	// 		"resource.instanceDetails.instanceId":                                            {"equals", "not_equals"},
	// 		"resource.instanceDetails.networkInterfaces.ipv6Addresses":                       {"equals", "not_equals"},
	// 		"resource.instanceDetails.networkInterfaces.privateIpAddresses.privateIpAddress": {"equals", "not_equals"},
	// 		"resource.instanceDetails.networkInterfaces.publicDnsName":                       {"equals", "not_equals"},
	// 		"resource.instanceDetails.networkInterfaces.publicIp":                            {"equals", "not_equals"},
	// 		"resource.instanceDetails.networkInterfaces.securityGroups.groupId":              {"equals", "not_equals"},
	// 		"resource.instanceDetails.networkInterfaces.securityGroups.groupName":            {"equals", "not_equals"},
	// 		"resource.instanceDetails.networkInterfaces.subnetId":                            {"equals", "not_equals"},
	// 		"resource.instanceDetails.networkInterfaces.vpcId":                               {"equals", "not_equals"},
	// 		"resource.instanceDetails.tags.key":                                              {"equals", "not_equals"},
	// 		"resource.instanceDetails.tags.value":                                            {"equals", "not_equals"},
	// 		"resource.resourceType":                                                          {"equals", "not_equals"},
	// 		"service.action.actionType":                                                      {"equals", "not_equals"},
	// 		"service.action.awsApiCallAction.api":                                            {"equals", "not_equals"},
	// 		"service.action.awsApiCallAction.callerType":                                     {"equals", "not_equals"},
	// 		"service.action.awsApiCallAction.remoteIpDetails.city.cityName":                  {"equals", "not_equals"},
	// 		"service.action.awsApiCallAction.remoteIpDetails.country.countryName":            {"equals", "not_equals"},
	// 		"service.action.awsApiCallAction.remoteIpDetails.ipAddressV4":                    {"equals", "not_equals"},
	// 		"service.action.awsApiCallAction.remoteIpDetails.organization.asn":               {"equals", "not_equals"},
	// 		"service.action.awsApiCallAction.remoteIpDetails.organization.asnOrg":            {"equals", "not_equals"},
	// 		"service.action.awsApiCallAction.serviceName":                                    {"equals", "not_equals"},
	// 		"service.action.dnsRequestAction.domain":                                         {"equals", "not_equals"},
	// 		"service.action.networkConnectionAction.blocked":                                 {"equals", "not_equals"},
	// 		"service.action.networkConnectionAction.connectionDirection":                     {"equals", "not_equals"},
	// 		"service.action.networkConnectionAction.localPortDetails.port":                   {"equals", "not_equals"},
	// 		"service.action.networkConnectionAction.protocol":                                {"equals", "not_equals"},
	// 		"service.action.networkConnectionAction.remoteIpDetails.city.cityName":           {"equals", "not_equals"},
	// 		"service.action.networkConnectionAction.remoteIpDetails.country.countryName":     {"equals", "not_equals"},
	// 		"service.action.networkConnectionAction.remoteIpDetails.ipAddressV4":             {"equals", "not_equals"},
	// 		"service.action.networkConnectionAction.remoteIpDetails.organization.asn":        {"equals", "not_equals"},
	// 		"service.action.networkConnectionAction.remoteIpDetails.organization.asnOrg":     {"equals", "not_equals"},
	// 		"service.action.networkConnectionAction.remotePortDetails.port":                  {"equals", "not_equals"},
	// 		"service.additionalInfo.threatListName":                                          {"equals", "not_equals"},
	// 		"service.archived":                                                               {"equals", "not_equals"},
	// 		"service.resourceRole":                                                           {"equals", "not_equals"},
	// 		"severity":                                                                       {"equals", "not_equals"},
	// 		"type":                                                                           {"equals", "not_equals"},
	// 		"updatedAt":                                                                      {"equals", "not_equals"},
	// 	}
	//
	inputFindingCriteria := findingCriteria["criterion"].(*schema.Set).List() //[0].(map[string]interface{})
	criteria := map[string]*guardduty.Condition{}
	for _, criterion := range inputFindingCriteria {
		typedCriterion := criterion.(map[string]interface{})
		log.Printf("[DEBUG!!!!!!!!!!] Criterion info: %#v", criterion)

		values := make([]string, len(typedCriterion["values"].([]interface{})))
		for i, v := range typedCriterion["values"].([]interface{}) {
			values[i] = string(v.(string))
		}

		criteria[typedCriterion["field"].(string)] = &guardduty.Condition{
			Equals: aws.StringSlice(values),
		}
	}
	log.Printf("[DEBUG] Creating FindingCriteria map: %#v", findingCriteria)
	log.Printf("[DEBUG] Creating FindingCriteria's criteria map: %#v", criteria)

	return &guardduty.FindingCriteria{Criterion: criteria}
}

func resourceAwsGuardDutyFilterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := guardduty.CreateFilterInput{
		DetectorId:  aws.String(d.Get("detector_id").(string)),
		Name:        aws.String(d.Get("name").(string)),
		Description: aws.String(d.Get("description").(string)),
		Rank:        aws.Int64(int64(d.Get("rank").(int))),
	}

	// building `FindingCriteria`
	findingCriteria := d.Get("finding_criteria").([]interface{})[0].(map[string]interface{})
	buildFindingCriteria(findingCriteria)
	input.FindingCriteria = buildFindingCriteria(findingCriteria)

	tagsInterface := d.Get("tags").(map[string]interface{})
	if len(tagsInterface) > 0 {
		tags := make(map[string]*string, len(tagsInterface))
		for i, v := range tagsInterface {
			tags[i] = aws.String(v.(string))
		}

		input.Tags = tags
	}

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
	detectorId := d.Get("detector_id").(string)
	name := d.Get("name").(string)

	input := guardduty.GetFilterInput{
		DetectorId: aws.String(detectorId),
		FilterName: aws.String(name),
	}

	log.Printf("[DEBUG] Reading GuardDuty Filter: %s", input)
	filter, err := conn.GetFilter(&input)

	if err != nil {
		if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
			log.Printf("[WARN] GuardDuty detector %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading GuardDuty Filter '%s' failed: %s", name, err.Error())
	}

	d.Set("action", filter.Action) // Make sure I really want to set all these attrs
	d.Set("description", filter.Description)
	d.Set("rank", filter.Rank)
	d.Set("name", d.Id())

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

	detectorId := d.Get("detector_id").(string)
	name := d.Get("name").(string)

	input := guardduty.DeleteFilterInput{
		FilterName: aws.String(name),
		DetectorId: aws.String(detectorId),
	}

	log.Printf("[DEBUG] Delete GuardDuty Filter: %s", input)

	_, err := conn.DeleteFilter(&input)
	if err != nil {
		return fmt.Errorf("Deleting GuardDuty Filter '%s' failed: %s", d.Id(), err.Error())
	}
	return nil
}
