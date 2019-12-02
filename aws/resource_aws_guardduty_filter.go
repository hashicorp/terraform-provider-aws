package aws

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	//"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsGuardDutyFilter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyFilterCreate,
		Read:   resourceAwsGuardDutyFilterRead,
		Update: resourceAwsGuardDutyFilterUpdate,
		Delete: resourceAwsGuardDutyFilterDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"auto_archive": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"rank": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"filters": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"filter_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"filter_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			// "tags": tagsSchema(),
		},
	}
}

func resourceAwsGuardDutyFilterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	var autoArchive string
	if d.Get("auto_archive").(bool) {
		autoArchive = "ARCHIVE"
	} else {
		autoArchive = "NOOP"
	}

	rank := int64(d.Get("rank").(int))
	var findingCriteria = resourceAwsGuardDutyCreateFindingCriteria(d.Get("filters").([]interface{}))

	input := &guardduty.CreateFilterInput{
		DetectorId:      aws.String(d.Get("detector_id").(string)),
		Name:            aws.String(d.Get("name").(string)),
		Description:     aws.String(d.Get("description").(string)),
		Action:          aws.String(autoArchive),
		Rank:            aws.Int64(rank),
		FindingCriteria: findingCriteria,
		// Tags:            tags,
	}

	resp, err := conn.CreateFilter(input)
	if err != nil {
		fmt.Printf("[ERROR] Create filter returned an error: %s", err)
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"modifying"},
		Target:     []string{"available"},
		Refresh:    guardDutyFilterRefreshStatusFunc(conn, *resp.Name, d.Get("detector_id").(string)),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("[ERROR] Timeout waiting for GuardDuty CreateFilter: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", d.Get("detector_id"), d.Get("name")))
	return resourceAwsGuardDutyFilterRead(d, meta)
}

func resourceAwsGuardDutyFilterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := &guardduty.GetFilterInput{
		DetectorId: aws.String(d.Get("detector_id").(string)),
		FilterName: aws.String(d.Get("name").(string)),
	}

	resp, err := conn.GetFilter(input)
	if err != nil {
		log.Printf("[ERROR] The filter name or detector Id was not found: %s", err)
		return err
	}

	d.Set("detector_id", d.Get("detector_id"))
	d.Set("name", resp.Name)
	d.Set("description", resp.Description)
	d.Set("action", resp.Action)
	d.Set("rank", resp.Rank)
	d.Set("criteria", resp.FindingCriteria)

	return nil
}

func resourceAwsGuardDutyFilterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := &guardduty.UpdateFilterInput{
		DetectorId: aws.String(d.Get("detector_id").(string)),
		FilterName: aws.String(d.Get("name").(string)),
	}

	if d.HasChange("detector_id") {
		input.DetectorId = aws.String(d.Get("detector_id").(string))
	}
	if d.HasChange("name") {
		input.FilterName = aws.String(d.Get("name").(string))
	}
	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}
	if d.HasChange("auto_archive") {
		var autoArchive string
		if d.Get("auto_archive").(bool) {
			autoArchive = "ARCHIVE"
		} else {
			autoArchive = "NOOP"
		}
		input.Action = aws.String(autoArchive)
	}
	if d.HasChange("rank") {
		rank := int64(d.Get("rank").(int))
		input.Rank = aws.Int64(rank)
		//input.Rank = aws.Int64(d.Get("rank").(int64))
	}
	if d.HasChange("filters") {
		var findingCriteria = resourceAwsGuardDutyCreateFindingCriteria(d.Get("filters").([]interface{}))
		input.FindingCriteria = findingCriteria
	}

	_, err := conn.UpdateFilter(input)
	if err != nil {
		log.Printf("[ERROR] The GuardDuty filter failed to update: %s", err)
		return err
	}

	return resourceAwsGuardDutyFilterRead(d, meta)
}

func resourceAwsGuardDutyFilterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := &guardduty.DeleteFilterInput{
		DetectorId: aws.String(d.Get("detector_id").(string)),
		FilterName: aws.String(d.Get("name").(string)),
	}

	_, err := conn.DeleteFilter(input)
	if err != nil {
		log.Printf("[ERROR] The GuardDuty filter failed to delete: %s", err)
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"modifying"},
		Target:     []string{"deleted"},
		Refresh:    guardDutyFilterRefreshStatusFunc(conn, d.Get("name").(string), d.Get("detector_id").(string)),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("[ERROR] Timeout waiting for GuardDuty DeleteFilter: %s", err)
	}

	return nil
}

func guardDutyFilterRefreshStatusFunc(conn *guardduty.GuardDuty, filterName string, detectorID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &guardduty.GetFilterInput{
			DetectorId: aws.String(detectorID),
			FilterName: aws.String(filterName),
		}
		emptyResp := &guardduty.GetFilterOutput{}
		resp, err := conn.GetFilter(input)
		if err != nil {
			log.Printf("[TEST] Got an error in guardDutyFilterRefreshStatusFunc")
			// log.Printf("%s", err)
			// log.Printf("%s", resp)
			// log.Printf("%s", err.Code())
			// log.Printf("%s", err.Message())
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "InternalServerErrorException" {
				// Bucket without policy
				log.Printf("[TEST] InternalServerErrorException")
				return emptyResp, "deleted", nil
				// return results, nil
			}
			// if resp == "InternalServerErrorException: The request is rejected since no such resource found." {
			// 	log.Printf("[TEST] InternalServerErrorException")
			// }
			// if isAWSErr(err, guardduty.InternalServerErrorException, "The request is rejected since no such resource found.") {
			// 	log.Printf("[TEST] InternalServerErrorException")
			// 	return nil, "deleted", nil
			// }
			return nil, "failed", err
		}
		if resp == nil {
			return nil, "deleted", nil
		}
		return resp, "available", nil
	}
}

func resourceAwsGuardDutyCreateFindingCriteria(l []interface{}) *guardduty.FindingCriteria {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	criterion := make(map[string]*guardduty.Condition)
	for _, element := range l {
		m := element.(map[string]interface{})
		var filterKey = m["filter_key"].(string)
		var filterValue = m["filter_value"].(string)
		var filterType = m["filter_type"].(string)
		guardduty_condition := &guardduty.Condition{}
		switch filterType {
		case "Equals":
			guardduty_condition.Equals = aws.StringSlice([]string{filterValue})
		case "GreaterThan":
			filterValueInt64, _ := strconv.ParseInt(filterValue, 10, 32)
			guardduty_condition.GreaterThan = aws.Int64(filterValueInt64)
		case "GreaterThanOrEqual":
			filterValueInt64, _ := strconv.ParseInt(filterValue, 10, 32)
			guardduty_condition.GreaterThanOrEqual = aws.Int64(filterValueInt64)
		case "LessThan":
			filterValueInt64, _ := strconv.ParseInt(filterValue, 10, 32)
			guardduty_condition.LessThan = aws.Int64(filterValueInt64)
		case "LessThanOrEqual":
			filterValueInt64, _ := strconv.ParseInt(filterValue, 10, 32)
			guardduty_condition.LessThanOrEqual = aws.Int64(filterValueInt64)
		case "NotEquals":
			guardduty_condition.NotEquals = aws.StringSlice([]string{filterValue})
		default:
			log.Printf("[ERROR] Invalid filter type used: %s", filterValue)
		}
		criterion[filterKey] = guardduty_condition
	}

	findingCriteria := &guardduty.FindingCriteria{
		Criterion: criterion,
	}

	return findingCriteria
}
