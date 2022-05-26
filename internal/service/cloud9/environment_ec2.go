package cloud9

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEnvironmentEC2() *schema.Resource {
	return &schema.Resource{
		Create: resourceEnvironmentEC2Create,
		Read:   resourceEnvironmentEC2Read,
		Update: resourceEnvironmentEC2Update,
		Delete: resourceEnvironmentEC2Delete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automatic_stop_time_minutes": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtMost(20160),
			},
			"connection_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      cloud9.ConnectionTypeConnectSsh,
				ValidateFunc: validation.StringInSlice(cloud9.ConnectionType_Values(), false),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 200),
			},
			"image_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"amazonlinux-1-x86_64",
					"amazonlinux-2-x86_64",
					"ubuntu-18.04-x86_64",
					"resolve:ssm:/aws/service/cloud9/amis/amazonlinux-1-x86_64",
					"resolve:ssm:/aws/service/cloud9/amis/amazonlinux-2-x86_64",
					"resolve:ssm:/aws/service/cloud9/amis/ubuntu-18.04-x86_64",
				}, false),
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 60),
			},
			"owner_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEnvironmentEC2Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Cloud9Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &cloud9.CreateEnvironmentEC2Input{
		ClientRequestToken: aws.String(resource.UniqueId()),
		ConnectionType:     aws.String(d.Get("connection_type").(string)),
		InstanceType:       aws.String(d.Get("instance_type").(string)),
		Name:               aws.String(name),
		Tags:               Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("automatic_stop_time_minutes"); ok {
		input.AutomaticStopTimeMinutes = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("image_id"); ok {
		input.ImageId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("owner_arn"); ok {
		input.OwnerArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("subnet_id"); ok {
		input.SubnetId = aws.String(v.(string))
	}

	log.Printf("[INFO] Creating Cloud9 EC2 Environment: %s", input)
	var output *cloud9.CreateEnvironmentEC2Output
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.CreateEnvironmentEC2(input)

		if err != nil {
			// NotFoundException: User arn:aws:iam::*******:user/****** does not exist.
			if tfawserr.ErrMessageContains(err, cloud9.ErrCodeNotFoundException, "User") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateEnvironmentEC2(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Cloud9 EC2 Environment (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.EnvironmentId))

	_, err = waitEnvironmentReady(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Cloud9 EC2 Environment (%s) create: %w", d.Id(), err)
	}

	return resourceEnvironmentEC2Read(d, meta)
}

func resourceEnvironmentEC2Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Cloud9Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	env, err := FindEnvironmentByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cloud9 EC2 Environment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Cloud9 EC2 Environment (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(env.Arn)
	d.Set("arn", arn)
	d.Set("connection_type", env.ConnectionType)
	d.Set("description", env.Description)
	d.Set("name", env.Name)
	d.Set("owner_arn", env.OwnerArn)
	d.Set("type", env.Type)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Cloud9 EC2 Environment (%s): %w", arn, err)
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

func resourceEnvironmentEC2Update(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Cloud9Conn

	if d.HasChangesExcept("tags_all", "tags") {
		input := cloud9.UpdateEnvironmentInput{
			Description:   aws.String(d.Get("description").(string)),
			EnvironmentId: aws.String(d.Id()),
			Name:          aws.String(d.Get("name").(string)),
		}

		log.Printf("[INFO] Updating Cloud9 EC2 Environment: %s", input)
		_, err := conn.UpdateEnvironment(&input)

		if err != nil {
			return fmt.Errorf("error updating Cloud9 EC2 Environment (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		arn := d.Get("arn").(string)

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Cloud9 EC2 Environment (%s) tags: %w", arn, err)
		}
	}

	return resourceEnvironmentEC2Read(d, meta)
}

func resourceEnvironmentEC2Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Cloud9Conn

	log.Printf("[INFO] Deleting Cloud9 EC2 Environment: %s", d.Id())
	_, err := conn.DeleteEnvironment(&cloud9.DeleteEnvironmentInput{
		EnvironmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloud9.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Cloud9 EC2 Environment (%s): %w", d.Id(), err)
	}

	_, err = waitEnvironmentDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Cloud9 EC2 Environment (%s) delete: %w", d.Id(), err)
	}

	return nil
}
