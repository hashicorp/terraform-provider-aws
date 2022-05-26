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

func ResourceIPSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceIPSetCreate,
		Read:   resourceIPSetRead,
		Update: resourceIPSetUpdate,
		Delete: resourceIPSetDelete,

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
					guardduty.IpSetFormatTxt,
					guardduty.IpSetFormatStix,
					guardduty.IpSetFormatOtxCsv,
					guardduty.IpSetFormatAlienVault,
					guardduty.IpSetFormatProofPoint,
					guardduty.IpSetFormatFireEye,
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

func resourceIPSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GuardDutyConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	detectorID := d.Get("detector_id").(string)
	input := &guardduty.CreateIPSetInput{
		DetectorId: aws.String(detectorID),
		Name:       aws.String(d.Get("name").(string)),
		Format:     aws.String(d.Get("format").(string)),
		Location:   aws.String(d.Get("location").(string)),
		Activate:   aws.Bool(d.Get("activate").(bool)),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.CreateIPSet(input)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{guardduty.IpSetStatusActivating, guardduty.IpSetStatusDeactivating},
		Target:     []string{guardduty.IpSetStatusActive, guardduty.IpSetStatusInactive},
		Refresh:    ipsetRefreshStatusFunc(conn, *resp.IpSetId, detectorID),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for GuardDuty IpSet status to be \"%s\" or \"%s\": %s", guardduty.IpSetStatusActive, guardduty.IpSetStatusInactive, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", detectorID, *resp.IpSetId))
	return resourceIPSetRead(d, meta)
}

func resourceIPSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GuardDutyConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ipSetId, detectorId, err := DecodeIPSetID(d.Id())
	if err != nil {
		return err
	}
	input := &guardduty.GetIPSetInput{
		DetectorId: aws.String(detectorId),
		IpSetId:    aws.String(ipSetId),
	}

	resp, err := conn.GetIPSet(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
			log.Printf("[WARN] GuardDuty IpSet %q not found, removing from state", ipSetId)
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
		Resource:  fmt.Sprintf("detector/%s/ipset/%s", detectorId, ipSetId),
	}.String()
	d.Set("arn", arn)

	d.Set("detector_id", detectorId)
	d.Set("format", resp.Format)
	d.Set("location", resp.Location)
	d.Set("name", resp.Name)
	d.Set("activate", aws.StringValue(resp.Status) == guardduty.IpSetStatusActive)

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

func resourceIPSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GuardDutyConn

	ipSetId, detectorId, err := DecodeIPSetID(d.Id())
	if err != nil {
		return err
	}

	if d.HasChanges("activate", "location", "name") {
		input := &guardduty.UpdateIPSetInput{
			DetectorId: aws.String(detectorId),
			IpSetId:    aws.String(ipSetId),
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

		_, err = conn.UpdateIPSet(input)
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating GuardDuty IP Set (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceIPSetRead(d, meta)
}

func resourceIPSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GuardDutyConn

	ipSetId, detectorId, err := DecodeIPSetID(d.Id())
	if err != nil {
		return err
	}
	input := &guardduty.DeleteIPSetInput{
		DetectorId: aws.String(detectorId),
		IpSetId:    aws.String(ipSetId),
	}

	_, err = conn.DeleteIPSet(input)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			guardduty.IpSetStatusActive,
			guardduty.IpSetStatusActivating,
			guardduty.IpSetStatusInactive,
			guardduty.IpSetStatusDeactivating,
			guardduty.IpSetStatusDeletePending,
		},
		Target:     []string{guardduty.IpSetStatusDeleted},
		Refresh:    ipsetRefreshStatusFunc(conn, ipSetId, detectorId),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for GuardDuty IpSet status to be \"%s\": %s", guardduty.IpSetStatusDeleted, err)
	}

	return nil
}

func ipsetRefreshStatusFunc(conn *guardduty.GuardDuty, ipSetID, detectorID string) resource.StateRefreshFunc {
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

func DecodeIPSetID(id string) (ipsetID, detectorID string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = fmt.Errorf("GuardDuty IPSet ID must be of the form <Detector ID>:<IPSet ID>, was provided: %s", id)
		return
	}
	ipsetID = parts[1]
	detectorID = parts[0]
	return
}
