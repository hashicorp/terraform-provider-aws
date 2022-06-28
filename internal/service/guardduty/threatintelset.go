package guardduty

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceThreatintelset() *schema.Resource {
	return &schema.Resource{
		Create: resourceThreatintelsetCreate,
		Read:   resourceThreatintelsetRead,
		Update: resourceThreatintelsetUpdate,
		Delete: resourceThreatintelsetDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					guardduty.ThreatIntelSetFormatTxt,
					guardduty.ThreatIntelSetFormatStix,
					guardduty.ThreatIntelSetFormatOtxCsv,
					guardduty.ThreatIntelSetFormatAlienVault,
					guardduty.ThreatIntelSetFormatProofPoint,
					guardduty.ThreatIntelSetFormatFireEye,
				}, false),
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
			},
			"activate": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"tags": tftags.TagsSchema(),

			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceThreatintelsetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GuardDutyConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	detectorID := d.Get("detector_id").(string)
	input := &guardduty.CreateThreatIntelSetInput{
		DetectorId: aws.String(detectorID),
		Name:       aws.String(d.Get("name").(string)),
		Format:     aws.String(d.Get("format").(string)),
		Location:   aws.String(d.Get("location").(string)),
		Activate:   aws.Bool(d.Get("activate").(bool)),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.CreateThreatIntelSet(input)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{guardduty.ThreatIntelSetStatusActivating, guardduty.ThreatIntelSetStatusDeactivating},
		Target:     []string{guardduty.ThreatIntelSetStatusActive, guardduty.ThreatIntelSetStatusInactive},
		Refresh:    threatintelsetRefreshStatusFunc(conn, *resp.ThreatIntelSetId, detectorID),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for GuardDuty ThreatIntelSet status to be \"%s\" or \"%s\": %s",
			guardduty.ThreatIntelSetStatusActive, guardduty.ThreatIntelSetStatusInactive, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", detectorID, *resp.ThreatIntelSetId))
	return resourceThreatintelsetRead(d, meta)
}

func resourceThreatintelsetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GuardDutyConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	threatIntelSetId, detectorId, err := DecodeThreatintelsetID(d.Id())
	if err != nil {
		return err
	}
	input := &guardduty.GetThreatIntelSetInput{
		DetectorId:       aws.String(detectorId),
		ThreatIntelSetId: aws.String(threatIntelSetId),
	}

	resp, err := conn.GetThreatIntelSet(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
			log.Printf("[WARN] GuardDuty ThreatIntelSet %q not found, removing from state", threatIntelSetId)
			d.SetId("")
			return nil
		}
		return err
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "guardduty",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("detector/%s/threatintelset/%s", detectorId, threatIntelSetId),
	}.String()
	d.Set("arn", arn)

	d.Set("detector_id", detectorId)
	d.Set("format", resp.Format)
	d.Set("location", resp.Location)
	d.Set("name", resp.Name)
	d.Set("activate", aws.StringValue(resp.Status) == guardduty.ThreatIntelSetStatusActive)

	tags := KeyValueTags(resp.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceThreatintelsetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GuardDutyConn

	threatIntelSetID, detectorId, err := DecodeThreatintelsetID(d.Id())
	if err != nil {
		return err
	}

	if d.HasChanges("activate", "location", "name") {
		input := &guardduty.UpdateThreatIntelSetInput{
			DetectorId:       aws.String(detectorId),
			ThreatIntelSetId: aws.String(threatIntelSetID),
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

		_, err = conn.UpdateThreatIntelSet(input)
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating GuardDuty Threat Intel Set (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceThreatintelsetRead(d, meta)
}

func resourceThreatintelsetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GuardDutyConn

	threatIntelSetID, detectorId, err := DecodeThreatintelsetID(d.Id())
	if err != nil {
		return err
	}
	input := &guardduty.DeleteThreatIntelSetInput{
		DetectorId:       aws.String(detectorId),
		ThreatIntelSetId: aws.String(threatIntelSetID),
	}

	_, err = conn.DeleteThreatIntelSet(input)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			guardduty.ThreatIntelSetStatusActive,
			guardduty.ThreatIntelSetStatusActivating,
			guardduty.ThreatIntelSetStatusInactive,
			guardduty.ThreatIntelSetStatusDeactivating,
			guardduty.ThreatIntelSetStatusDeletePending,
		},
		Target:     []string{guardduty.ThreatIntelSetStatusDeleted},
		Refresh:    threatintelsetRefreshStatusFunc(conn, threatIntelSetID, detectorId),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for GuardDuty ThreatIntelSet status to be \"%s\": %s", guardduty.ThreatIntelSetStatusDeleted, err)
	}

	return nil
}

func threatintelsetRefreshStatusFunc(conn *guardduty.GuardDuty, threatIntelSetID, detectorID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &guardduty.GetThreatIntelSetInput{
			DetectorId:       aws.String(detectorID),
			ThreatIntelSetId: aws.String(threatIntelSetID),
		}
		resp, err := conn.GetThreatIntelSet(input)
		if err != nil {
			return nil, "failed", err
		}
		return resp, *resp.Status, nil
	}
}

func DecodeThreatintelsetID(id string) (threatIntelSetID, detectorID string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = fmt.Errorf("GuardDuty ThreatIntelSet ID must be of the form <Detector ID>:<ThreatIntelSet ID>, was provided: %s", id)
		return
	}
	threatIntelSetID = parts[1]
	detectorID = parts[0]
	return
}
