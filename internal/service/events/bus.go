package events

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBus() *schema.Resource {
	return &schema.Resource{
		Create: resourceBusCreate,
		Read:   resourceBusRead,
		Update: resourceBusUpdate,
		Delete: resourceBusDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validCustomEventBusName,
			},
			"event_source_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validSourceName,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBusCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	eventBusName := d.Get("name").(string)
	input := &eventbridge.CreateEventBusInput{
		Name: aws.String(eventBusName),
	}

	if v, ok := d.GetOk("event_source_name"); ok {
		input.EventSourceName = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating EventBridge event bus: %v", input)

	output, err := conn.CreateEventBus(input)

	// Some partitions may not support tag-on-create
	if input.Tags != nil && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] EventBridge Bus (%s) create failed (%s) with tags. Trying create without tags.", eventBusName, err)
		input.Tags = nil
		output, err = conn.CreateEventBus(input)
	}

	if err != nil {
		return fmt.Errorf("Creating EventBridge event bus (%s) failed: %w", eventBusName, err)
	}

	d.SetId(eventBusName)

	log.Printf("[INFO] EventBridge event bus (%s) created", d.Id())

	// Post-create tagging supported in some partitions
	if input.Tags == nil && len(tags) > 0 {
		err := UpdateTags(conn, aws.StringValue(output.EventBusArn), nil, tags)

		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] error adding tags after create for EventBridge Bus (%s): %s", d.Id(), err)
			return resourceBusRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("error creating EventBridge Bus (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceBusRead(d, meta)
}

func resourceBusRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &eventbridge.DescribeEventBusInput{
		Name: aws.String(d.Id()),
	}

	output, err := conn.DescribeEventBus(input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] EventBridge event bus (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading EventBridge event bus: %w", err)
	}

	d.Set("arn", output.Arn)
	d.Set("name", output.Name)

	tags, err := ListTags(conn, aws.StringValue(output.Arn))

	// ISO partitions may not support tagging, giving error
	if verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] Unable to list tags for EventBridge Bus %s: %s", d.Id(), err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing tags for EventBridge event bus (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceBusUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(conn, arn, o, n)

		if verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] Unable to update tags for EventBridge Bus %s: %s", d.Id(), err)
			return resourceBusRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("error updating EventBridge Bus tags: %w", err)
		}
	}

	return resourceBusRead(d, meta)
}

func resourceBusDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn
	log.Printf("[INFO] Deleting EventBridge event bus (%s)", d.Id())
	_, err := conn.DeleteEventBus(&eventbridge.DeleteEventBusInput{
		Name: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] EventBridge event bus (%s) not found", d.Id())
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting EventBridge event bus (%s): %w", d.Id(), err)
	}
	log.Printf("[INFO] EventBridge event bus (%s) deleted", d.Id())

	return nil
}
