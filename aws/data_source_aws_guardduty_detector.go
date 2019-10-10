package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsGuarddutyDetector() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsGuarddutyDetectorRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"service_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"finding_publishing_frequency": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsGuarddutyDetectorRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	detectorId := d.Get("id").(string)

	if detectorId == "" {
		log.Print("[DEBUG] No Guardduty ID has been defined, try to lookup using find (There should be only one)")
		input := &guardduty.ListDetectorsInput{}

		resp, err := conn.ListDetectors(input)
		if err != nil {
			return err
		}

		if resp == nil || len(resp.DetectorIds) == 0 {
			return fmt.Errorf("no detectors found")
		}
		if len(resp.DetectorIds) > 1 {
			return fmt.Errorf("multiple detectors found; Amazon AWS API behavior has changed; this is a bug in the provider an should be reported")
		}

		detectorId = aws.StringValue(resp.DetectorIds[0])
	}

	getInput := &guardduty.GetDetectorInput{
		DetectorId: aws.String(detectorId),
	}

	getResp, err := conn.GetDetector(getInput)
	if err != nil {
		return err
	}

	if getResp == nil {
		return fmt.Errorf("cannot receive detector details")
	}

	status := getResp.Status
	enabled := false
	if *status == "ENABLED" {
		enabled = true
	} else {
		enabled = false
	}

	serviceRole := getResp.ServiceRole
	frequency := getResp.FindingPublishingFrequency

	log.Printf("[DEBUG] Setting AWS Guardduty Detector ID to %s.", detectorId)
	d.SetId(detectorId)
	log.Printf("[DEBUG] Setting AWS Guardduty enabled to %t.", enabled)
	d.Set("enabled", enabled)
	log.Printf("[DEBUG] Setting AWS Guardduty Service Role ARN to %s.", *serviceRole)
	d.Set("service_role_arn", *serviceRole)
	log.Printf("[DEBUG] Setting AWS Guardduty Log Publishing Frequency to %s.", *frequency)
	d.Set("finding_publishing_frequency", *frequency)

	return nil
}
