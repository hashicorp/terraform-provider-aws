package events

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceArchive() *schema.Resource {
	return &schema.Resource{
		Create: resourceArchiveCreate,
		Read:   resourceArchiveRead,
		Update: resourceArchiveUpdate,
		Delete: resourceArchiveDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validArchiveName,
			},
			"event_source_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"event_pattern": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateEventPatternValue(),
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v.(string))
					return json
				},
			},
			"retention_days": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceArchiveCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	input, err := buildCreateArchiveInputStruct(d)

	if err != nil {
		return fmt.Errorf("Creating EventBridge Archive parameters failed: %w", err)
	}

	log.Printf("[DEBUG] Creating EventBridge Archive: %s", input)

	_, err = conn.CreateArchive(input)
	if err != nil {
		return fmt.Errorf("Creating EventBridge Archive failed: %w", err)
	}

	d.SetId(d.Get("name").(string))

	log.Printf("[INFO] EventBridge Archive (%s) created", d.Id())

	return resourceArchiveRead(d, meta)
}

func resourceArchiveRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn
	input := &eventbridge.DescribeArchiveInput{
		ArchiveName: aws.String(d.Id()),
	}

	out, err := conn.DescribeArchive(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] EventBridge archive (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error reading EventBridge archive (%s): %w", d.Id(), err)
	}

	d.Set("name", out.ArchiveName)
	d.Set("description", out.Description)
	d.Set("event_pattern", out.EventPattern)
	d.Set("event_source_arn", out.EventSourceArn)
	d.Set("arn", out.ArchiveArn)
	d.Set("retention_days", out.RetentionDays)

	return nil
}

func resourceArchiveUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	input, err := buildUpdateArchiveInputStruct(d)

	if err != nil {
		return fmt.Errorf("Creating EventBridge Archive parameters failed: %w", err)
	}

	log.Printf("[DEBUG] Updating EventBridge Archive: %s", input)
	_, err = conn.UpdateArchive(input)
	if err != nil {
		return fmt.Errorf("error updating EventBridge Archive (%s): %w", d.Id(), err)
	}

	return resourceArchiveRead(d, meta)
}

func resourceArchiveDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	input := &eventbridge.DeleteArchiveInput{
		ArchiveName: aws.String(d.Get("name").(string)),
	}

	_, err := conn.DeleteArchive(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting EventBridge Archive (%s): %w", d.Id(), err)
	}

	return nil
}

func buildCreateArchiveInputStruct(d *schema.ResourceData) (*eventbridge.CreateArchiveInput, error) {
	input := eventbridge.CreateArchiveInput{
		ArchiveName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("event_pattern"); ok {
		pattern, err := structure.NormalizeJsonString(v)
		if err != nil {
			return nil, fmt.Errorf("event pattern contains an invalid JSON: %w", err)
		}
		input.EventPattern = aws.String(pattern)
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("event_source_arn"); ok {
		input.EventSourceArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("retention_days"); ok {
		input.RetentionDays = aws.Int64(int64(v.(int)))
	}

	return &input, nil
}

func buildUpdateArchiveInputStruct(d *schema.ResourceData) (*eventbridge.UpdateArchiveInput, error) {
	input := eventbridge.UpdateArchiveInput{
		ArchiveName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("event_pattern"); ok {
		pattern, err := structure.NormalizeJsonString(v)
		if err != nil {
			return nil, fmt.Errorf("event pattern contains an invalid JSON: %w", err)
		}
		input.EventPattern = aws.String(pattern)
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("retention_days"); ok {
		input.RetentionDays = aws.Int64(int64(v.(int)))
	}

	return &input, nil
}
