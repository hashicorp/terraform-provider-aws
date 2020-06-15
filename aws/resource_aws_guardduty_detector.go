package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/guardduty/waiter"
)

func resourceAwsGuardDutyDetector() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyDetectorCreate,
		Read:   resourceAwsGuardDutyDetectorRead,
		Update: resourceAwsGuardDutyDetectorUpdate,
		Delete: resourceAwsGuardDutyDetectorDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// finding_publishing_frequency is marked as Computed:true since
			// GuardDuty member accounts inherit setting from master account
			"finding_publishing_frequency": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceAwsGuardDutyDetectorCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := guardduty.CreateDetectorInput{
		Enable: aws.Bool(d.Get("enable").(bool)),
	}

	if v, ok := d.GetOk("finding_publishing_frequency"); ok {
		input.FindingPublishingFrequency = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating GuardDuty Detector: %s", input)
	output, err := conn.CreateDetector(&input)
	if err != nil {
		return fmt.Errorf("Creating GuardDuty Detector failed: %s", err.Error())
	}
	d.SetId(*output.DetectorId)

	return resourceAwsGuardDutyDetectorRead(d, meta)
}

func resourceAwsGuardDutyDetectorRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn
	input := guardduty.GetDetectorInput{
		DetectorId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading GuardDuty Detector: %s", input)
	gdo, err := conn.GetDetector(&input)
	if err != nil {
		if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
			log.Printf("[WARN] GuardDuty detector %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading GuardDuty Detector '%s' failed: %s", d.Id(), err.Error())
	}

	d.Set("account_id", meta.(*AWSClient).accountid)
	d.Set("enable", *gdo.Status == guardduty.DetectorStatusEnabled)
	d.Set("finding_publishing_frequency", gdo.FindingPublishingFrequency)

	return nil
}

func resourceAwsGuardDutyDetectorUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := guardduty.UpdateDetectorInput{
		DetectorId:                 aws.String(d.Id()),
		Enable:                     aws.Bool(d.Get("enable").(bool)),
		FindingPublishingFrequency: aws.String(d.Get("finding_publishing_frequency").(string)),
	}

	log.Printf("[DEBUG] Update GuardDuty Detector: %s", input)
	_, err := conn.UpdateDetector(&input)
	if err != nil {
		return fmt.Errorf("Updating GuardDuty Detector '%s' failed: %s", d.Id(), err.Error())
	}

	return resourceAwsGuardDutyDetectorRead(d, meta)
}

func resourceAwsGuardDutyDetectorDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := &guardduty.DeleteDetectorInput{
		DetectorId: aws.String(d.Id()),
	}

	err := resource.Retry(waiter.MembershipPropagationTimeout, func() *resource.RetryError {
		_, err := conn.DeleteDetector(input)

		if isAWSErr(err, guardduty.ErrCodeBadRequestException, "cannot delete detector while it has invited or associated members") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteDetector(input)
	}

	if err != nil {
		return fmt.Errorf("error deleting GuardDuty Detector (%s): %w", d.Id(), err)
	}

	return nil
}
