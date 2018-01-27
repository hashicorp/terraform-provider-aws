package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsGuardDutyIpset() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyIpsetCreate,
		Read:   resourceAwsGuardDutyIpsetRead,
		Update: resourceAwsGuardDutyIpsetUpdate,
		Delete: resourceAwsGuardDutyIpsetDelete,

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
			"format": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateGuardDutyIpsetFormat,
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
			},
			"activate": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceAwsGuardDutyIpsetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	detectorID := d.Get("detector_id").(string)
	input := &guardduty.CreateIPSetInput{
		DetectorId: aws.String(detectorID),
		Name:       aws.String(d.Get("name").(string)),
		Format:     aws.String(d.Get("format").(string)),
		Location:   aws.String(d.Get("location").(string)),
		Activate:   aws.Bool(d.Get("activate").(bool)),
	}

	resp, err := conn.CreateIPSet(input)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{guardduty.IpSetStatusActivating},
		Target:     []string{guardduty.IpSetStatusActive},
		Refresh:    guardDutyIpsetRefreshStatusFunc(conn, *resp.IpSetId, detectorID),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("[WARN] Error waiting for GuardDuty IpSet status to be \"%s\": %s", guardduty.IpSetStatusActive, err)
	}

	d.SetId(*resp.IpSetId)
	return resourceAwsGuardDutyIpsetRead(d, meta)
}

func resourceAwsGuardDutyIpsetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := &guardduty.GetIPSetInput{
		DetectorId: aws.String(d.Get("detector_id").(string)),
		IpSetId:    aws.String(d.Id()),
	}

	resp, err := conn.GetIPSet(input)
	if err != nil {
		if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
			log.Printf("[WARN] GuardDuty IpSet %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("format", resp.Format)
	d.Set("location", resp.Location)
	d.Set("name", resp.Name)
	return nil
}

func resourceAwsGuardDutyIpsetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	detectorID := d.Get("detector_id").(string)
	input := &guardduty.UpdateIPSetInput{
		DetectorId: aws.String(detectorID),
		IpSetId:    aws.String(d.Id()),
	}

	if d.HasChange("name") {
		input.Name = aws.String(d.Get("name").(string))
	}
	if d.HasChange("location") {
		input.Location = aws.String(d.Get("location").(string))
	}
	if d.HasChange("activate") {
		input.Activate = aws.Bool(d.Get("activate").(bool))
	}

	_, err := conn.UpdateIPSet(input)
	if err != nil {
		return err
	}

	return resourceAwsGuardDutyIpsetRead(d, meta)
}

func resourceAwsGuardDutyIpsetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	detectorID := d.Get("detector_id").(string)
	input := &guardduty.DeleteIPSetInput{
		DetectorId: aws.String(detectorID),
		IpSetId:    aws.String(d.Id()),
	}

	_, err := conn.DeleteIPSet(input)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			guardduty.IpSetStatusActive,
			guardduty.IpSetStatusInactive,
			guardduty.IpSetStatusDeactivating,
			guardduty.IpSetStatusDeletePending,
		},
		Target:     []string{guardduty.IpSetStatusDeleted},
		Refresh:    guardDutyIpsetRefreshStatusFunc(conn, d.Id(), detectorID),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("[WARN] Error waiting for GuardDuty IpSet status to be \"%s\": %s", guardduty.IpSetStatusDeleted, err)
	}

	return nil
}

func guardDutyIpsetRefreshStatusFunc(conn *guardduty.GuardDuty, ipSetID, detectorID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &guardduty.GetIPSetInput{
			DetectorId: aws.String(detectorID),
			IpSetId:    aws.String(ipSetID),
		}
		resp, err := conn.GetIPSet(input)
		if err != nil {
			return nil, "failed", err
		}
		return resp, *resp.Status, nil
	}
}
