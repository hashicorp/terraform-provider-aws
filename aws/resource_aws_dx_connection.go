package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/directconnect/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/directconnect/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsDxConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxConnectionCreate,
		Read:   resourceAwsDxConnectionRead,
		Update: resourceAwsDxConnectionUpdate,
		Delete: resourceAwsDxConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bandwidth": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateDxConnectionBandWidth(),
			},
			"has_logical_redundancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsDxConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &directconnect.CreateConnectionInput{
		Bandwidth:      aws.String(d.Get("bandwidth").(string)),
		ConnectionName: aws.String(name),
		Location:       aws.String(d.Get("location").(string)),
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().DirectconnectTags()
	}

	log.Printf("[DEBUG] Creating Direct Connect connection: %s", input)
	output, err := conn.CreateConnection(input)

	if err != nil {
		return fmt.Errorf("error creating Direct Connect connection (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.ConnectionId))

	return resourceAwsDxConnectionRead(d, meta)
}

func resourceAwsDxConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	connection, err := finder.ConnectionByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Direct Connect connection (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    aws.StringValue(connection.Region),
		Service:   "directconnect",
		AccountID: aws.StringValue(connection.OwnerAccount),
		Resource:  fmt.Sprintf("dxcon/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("aws_device", connection.AwsDeviceV2)
	d.Set("bandwidth", connection.Bandwidth)
	d.Set("has_logical_redundancy", connection.HasLogicalRedundancy)
	d.Set("jumbo_frame_capable", connection.JumboFrameCapable)
	d.Set("location", connection.Location)
	d.Set("name", connection.ConnectionName)

	tags, err := keyvaluetags.DirectconnectListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Direct Connect connection (%s): %w", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsDxConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.DirectconnectUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Direct Connect connection (%s) tags: %w", arn, err)
		}
	}

	return resourceAwsDxConnectionRead(d, meta)
}

func resourceAwsDxConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	log.Printf("[DEBUG] Deleting Direct Connect connection: %s", d.Id())
	_, err := conn.DeleteConnection(&directconnect.DeleteConnectionInput{
		ConnectionId: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "Could not find Connection with ID") {
		return nil
	}

	_, err = waiter.ConnectionDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Direct Connect connection (%s) delete: %w", d.Id(), err)
	}

	return nil
}
