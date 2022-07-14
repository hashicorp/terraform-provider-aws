package mediastore

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceContainer() *schema.Resource {
	return &schema.Resource{
		Create: resourceContainerCreate,
		Read:   resourceContainerRead,
		Update: resourceContainerUpdate,
		Delete: resourceContainerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\w+$`), "must contain alphanumeric characters or underscores"),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceContainerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).MediaStoreConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &mediastore.CreateContainerInput{
		ContainerName: aws.String(d.Get("name").(string)),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.CreateContainer(input)
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(resp.Container.Name))

	stateConf := &resource.StateChangeConf{
		Pending:    []string{mediastore.ContainerStatusCreating},
		Target:     []string{mediastore.ContainerStatusActive},
		Refresh:    containerRefreshStatusFunc(conn, d.Id()),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return resourceContainerRead(d, meta)
}

func resourceContainerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).MediaStoreConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &mediastore.DescribeContainerInput{
		ContainerName: aws.String(d.Id()),
	}
	resp, err := conn.DescribeContainer(input)
	if tfawserr.ErrCodeEquals(err, mediastore.ErrCodeContainerNotFoundException) {
		log.Printf("[WARN] No Container found: %s, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error describing media store container %s: %s", d.Id(), err)
	}

	arn := aws.StringValue(resp.Container.ARN)
	d.Set("arn", arn)
	d.Set("name", resp.Container.Name)
	d.Set("endpoint", resp.Container.Endpoint)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for media store container (%s): %s", arn, err)
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

func resourceContainerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).MediaStoreConn

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating media store container (%s) tags: %s", arn, err)
		}
	}

	return resourceContainerRead(d, meta)
}

func resourceContainerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).MediaStoreConn

	input := &mediastore.DeleteContainerInput{
		ContainerName: aws.String(d.Id()),
	}
	_, err := conn.DeleteContainer(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, mediastore.ErrCodeContainerNotFoundException) {
			return nil
		}
		return err
	}

	dcinput := &mediastore.DescribeContainerInput{
		ContainerName: aws.String(d.Id()),
	}
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DescribeContainer(dcinput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, mediastore.ErrCodeContainerNotFoundException) {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		return resource.RetryableError(fmt.Errorf("Media Store Container (%s) still exists", d.Id()))
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DescribeContainer(dcinput)
	}
	if err != nil {
		return fmt.Errorf("error waiting for Media Store Container (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func containerRefreshStatusFunc(conn *mediastore.MediaStore, cn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &mediastore.DescribeContainerInput{
			ContainerName: aws.String(cn),
		}
		resp, err := conn.DescribeContainer(input)
		if err != nil {
			return nil, "failed", err
		}
		return resp, *resp.Container.Status, nil
	}
}
