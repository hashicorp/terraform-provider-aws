package aws

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsGuardDutyFilter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyFilterCreate,
		Read:   resourceAwsGuardDutyFilterRead,
		Update: resourceAwsGuardDutyFilterUpdate,
		Delete: resourceAwsGuardDutyFilterDelete,

		// Importer: &schema.ResourceImporter{
		// 	State: schema.ImportStatePassthrough,
		// },
		Schema: map[string]*schema.Schema{
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchemaForceNew(),
			"finding_criteria": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"criterion": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"field": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(criteriaFields(), false),
									},
									"condition": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"equals",
											"not_equals",
											"greater_than",
											"greater_than_or_equal",
											"less_than",
											"less_than_or_equal",
										}, false),
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
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"NOOP",
					"ARCHIVE",
				}, false),
			},
			"rank": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Second),
			Update: schema.DefaultTimeout(60 * time.Second),
		},
	}
}

func criteriaFields() []string {
	criteria := make([]string, 0, len(criteriaMap()))
	for criterion := range criteriaMap() {
		criteria = append(criteria, criterion)
	}
	return criteria
}

func criteriaMap() map[string][]string {
	return map[string][]string{
		"confidence": {"equals", "not_equals", "greater_than", "greater_than_or_equal", "less_than", "less_than_or_equal"},
		"id":         {"equals", "not_equals", "greater_than", "greater_than_or_equal", "less_than", "less_than_or_equal"},
		"account_id": {"equals", "not_equals"},
		"region":     {"equals", "not_equals"},
		"resource.accessKeyDetails.accessKeyId":                                          {"equals", "not_equals"},
		"resource.accessKeyDetails.principalId":                                          {"equals", "not_equals"},
		"resource.accessKeyDetails.userName":                                             {"equals", "not_equals"},
		"resource.accessKeyDetails.userType":                                             {"equals", "not_equals"},
		"resource.instanceDetails.iamInstanceProfile.id":                                 {"equals", "not_equals"},
		"resource.instanceDetails.imageId":                                               {"equals", "not_equals"},
		"resource.instanceDetails.instanceId":                                            {"equals", "not_equals"},
		"resource.instanceDetails.networkInterfaces.ipv6Addresses":                       {"equals", "not_equals"},
		"resource.instanceDetails.networkInterfaces.privateIpAddresses.privateIpAddress": {"equals", "not_equals"},
		"resource.instanceDetails.networkInterfaces.publicDnsName":                       {"equals", "not_equals"},
		"resource.instanceDetails.networkInterfaces.publicIp":                            {"equals", "not_equals"},
		"resource.instanceDetails.networkInterfaces.securityGroups.groupId":              {"equals", "not_equals"},
		"resource.instanceDetails.networkInterfaces.securityGroups.groupName":            {"equals", "not_equals"},
		"resource.instanceDetails.networkInterfaces.subnetId":                            {"equals", "not_equals"},
		"resource.instanceDetails.networkInterfaces.vpcId":                               {"equals", "not_equals"},
		"resource.instanceDetails.tags.key":                                              {"equals", "not_equals"},
		"resource.instanceDetails.tags.value":                                            {"equals", "not_equals"},
		"resource.resourceType":                                                          {"equals", "not_equals"},
		"service.action.actionType":                                                      {"equals", "not_equals"},
		"service.action.awsApiCallAction.api":                                            {"equals", "not_equals"},
		"service.action.awsApiCallAction.callerType":                                     {"equals", "not_equals"},
		"service.action.awsApiCallAction.remoteIpDetails.city.cityName":                  {"equals", "not_equals"},
		"service.action.awsApiCallAction.remoteIpDetails.country.countryName":            {"equals", "not_equals"},
		"service.action.awsApiCallAction.remoteIpDetails.ipAddressV4":                    {"equals", "not_equals"},
		"service.action.awsApiCallAction.remoteIpDetails.organization.asn":               {"equals", "not_equals"},
		"service.action.awsApiCallAction.remoteIpDetails.organization.asnOrg":            {"equals", "not_equals"},
		"service.action.awsApiCallAction.serviceName":                                    {"equals", "not_equals"},
		"service.action.dnsRequestAction.domain":                                         {"equals", "not_equals"},
		"service.action.networkConnectionAction.blocked":                                 {"equals", "not_equals"},
		"service.action.networkConnectionAction.connectionDirection":                     {"equals", "not_equals"},
		"service.action.networkConnectionAction.localPortDetails.port":                   {"equals", "not_equals"},
		"service.action.networkConnectionAction.protocol":                                {"equals", "not_equals"},
		"service.action.networkConnectionAction.remoteIpDetails.city.cityName":           {"equals", "not_equals"},
		"service.action.networkConnectionAction.remoteIpDetails.country.countryName":     {"equals", "not_equals"},
		"service.action.networkConnectionAction.remoteIpDetails.ipAddressV4":             {"equals", "not_equals"},
		"service.action.networkConnectionAction.remoteIpDetails.organization.asn":        {"equals", "not_equals"},
		"service.action.networkConnectionAction.remoteIpDetails.organization.asnOrg":     {"equals", "not_equals"},
		"service.action.networkConnectionAction.remotePortDetails.port":                  {"equals", "not_equals"},
		"service.additionalInfo.threatListName":                                          {"equals", "not_equals"},
		"service.archived":                                                               {"equals", "not_equals"},
		"service.resourceRole":                                                           {"equals", "not_equals"},
		"severity":                                                                       {"equals", "not_equals"},
		"type":                                                                           {"equals", "not_equals"},
		"updatedAt":                                                                      {"equals", "not_equals"},
	}
}

func conditionAllowedForCriterion(criterion map[string]interface{}) bool {
	availableConditions := criteriaMap()[criterion["field"].(string)]
	conditionToCheck := criterion["condition"].(string)

	for _, availableCondition := range availableConditions {
		if availableCondition == conditionToCheck {
			return true
		}
	}
	return false
}

func buildFindingCriteria(findingCriteria map[string]interface{}) (*guardduty.FindingCriteria, error) {
	inputFindingCriteria := findingCriteria["criterion"].(*schema.Set).List()
	criteria := map[string]*guardduty.Condition{}
	for _, criterion := range inputFindingCriteria {
		typedCriterion := criterion.(map[string]interface{})
		log.Printf("[DEBUG!!!!!!!!!!] Criterion info: %#v", criterion)

		if !conditionAllowedForCriterion(typedCriterion) {
			return nil, fmt.Errorf("The condition is not supported for the given field. Supported conditions are: %v", criteriaMap()[typedCriterion["field"].(string)])
		}

		switch typedCriterion["condition"].(string) {
		case "equals":
			criteria[typedCriterion["field"].(string)] = &guardduty.Condition{
				Equals: aws.StringSlice(conditionValueToStrings(typedCriterion["values"].([]interface{}))),
			}
		case "greater_than":
			criteria[typedCriterion["field"].(string)] = &guardduty.Condition{
				GreaterThan: aws.Int64(conditionValueToInt(typedCriterion["values"].([]interface{})).(int64)),
			}
		case "greater_than_or_equals":
			criteria[typedCriterion["field"].(string)] = &guardduty.Condition{
				GreaterThanOrEqual: aws.Int64(conditionValueToInt(typedCriterion["values"].([]interface{})).(int64)),
			}
		case "less_than":
			criteria[typedCriterion["field"].(string)] = &guardduty.Condition{
				LessThan: aws.Int64(conditionValueToInt(typedCriterion["values"].([]interface{})).(int64)),
			}
		case "less_than_or_equals":
			criteria[typedCriterion["field"].(string)] = &guardduty.Condition{
				LessThanOrEqual: aws.Int64(conditionValueToInt(typedCriterion["values"].([]interface{})).(int64)),
			}
		case "not_equals":
			criteria[typedCriterion["field"].(string)] = &guardduty.Condition{
				NotEquals: aws.StringSlice(conditionValueToStrings(typedCriterion["values"].([]interface{}))),
			}
		}

	}
	log.Printf("[DEBUG] Creating FindingCriteria map: %#v", findingCriteria)
	log.Printf("[DEBUG] Creating FindingCriteria's criteria map: %#v", criteria)

	return &guardduty.FindingCriteria{Criterion: criteria}, nil
}

func conditionValueToStrings(untypedValues []interface{}) []string {
	values := make([]string, len(untypedValues))
	for i, v := range untypedValues {
		values[i] = string(v.(string))
	}
	return values
}

func conditionValueToInt(untypedValues []interface{}) interface{} {
	if len(untypedValues) != 1 {
		return fmt.Errorf("Exactly one value must be given for conditions like less_ or greater_than. Instead given: %v", untypedValues)
	}

	untypedValue := untypedValues[0]
	typedValue, err := strconv.ParseInt(untypedValue.(string), 10, 64)
	if err != nil {
		return fmt.Errorf("Parsing condition value failed: %s", err.Error())
	}

	return typedValue
}

func resourceAwsGuardDutyFilterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := guardduty.CreateFilterInput{
		Action:      aws.String(d.Get("action").(string)),
		Description: aws.String(d.Get("description").(string)),
		DetectorId:  aws.String(d.Get("detector_id").(string)),
		Name:        aws.String(d.Get("name").(string)),
		Rank:        aws.Int64(int64(d.Get("rank").(int))),
	}

	// building `FindingCriteria`
	findingCriteria := d.Get("finding_criteria").([]interface{})[0].(map[string]interface{})

	var err error
	input.FindingCriteria, err = buildFindingCriteria(findingCriteria)
	if err != nil {
		return err
	}

	tagsInterface := d.Get("tags").(map[string]interface{})
	if len(tagsInterface) > 0 {
		tags := make(map[string]*string, len(tagsInterface))
		for i, v := range tagsInterface {
			tags[i] = aws.String(v.(string))
		}

		input.Tags = tags
	}

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

func resourceAwsGuardDutyFilterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := guardduty.UpdateFilterInput{
		Action:      aws.String(d.Get("action").(string)),
		Description: aws.String(d.Get("description").(string)),
		DetectorId:  aws.String(d.Get("detector_id").(string)),
		FilterName:  aws.String(d.Get("name").(string)),
		Rank:        aws.Int64(int64(d.Get("rank").(int))),
	}

	// building `FindingCriteria`
	findingCriteria := d.Get("finding_criteria").([]interface{})[0].(map[string]interface{})

	var err error
	input.FindingCriteria, err = buildFindingCriteria(findingCriteria)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Updating GuardDuty Filter: %s", input)

	_, err = conn.UpdateFilter(&input)
	if err != nil {
		return fmt.Errorf("Updating GuardDuty Filter with ID %s failed: %s", d.Id(), err.Error())
	}

	return resourceAwsGuardDutyFilterRead(d, meta)
}

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
