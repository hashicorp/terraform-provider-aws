package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceGroupCreate,
		Read:   resourceGroupRead,
		Update: resourceGroupUpdate,
		Delete: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"filter_expression": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).XRayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	input := &xray.CreateGroupInput{
		GroupName:        aws.String(d.Get("group_name").(string)),
		FilterExpression: aws.String(d.Get("filter_expression").(string)),
		Tags:             tags.IgnoreAws().XrayTags(),
	}

	out, err := conn.CreateGroup(input)
	if err != nil {
		return fmt.Errorf("error creating XRay Group: %w", err)
	}

	d.SetId(aws.StringValue(out.Group.GroupARN))

	return resourceGroupRead(d, meta)
}

func resourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).XRayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &xray.GetGroupInput{
		GroupARN: aws.String(d.Id()),
	}

	group, err := conn.GetGroup(input)

	if err != nil {
		if tfawserr.ErrMessageContains(err, xray.ErrCodeInvalidRequestException, "Group not found") {
			log.Printf("[WARN] XRay Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading XRay Group (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(group.Group.GroupARN)
	d.Set("arn", arn)
	d.Set("group_name", group.Group.GroupName)
	d.Set("filter_expression", group.Group.FilterExpression)

	tags, err := tftags.XrayListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for Xray Group (%q): %s", d.Id(), err)
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

func resourceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).XRayConn

	if d.HasChange("filter_expression") {
		input := &xray.UpdateGroupInput{
			GroupARN:         aws.String(d.Id()),
			FilterExpression: aws.String(d.Get("filter_expression").(string)),
		}

		_, err := conn.UpdateGroup(input)
		if err != nil {
			return fmt.Errorf("error updating XRay Group (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.XrayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceGroupRead(d, meta)
}

func resourceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).XRayConn

	log.Printf("[INFO] Deleting XRay Group: %s", d.Id())

	params := &xray.DeleteGroupInput{
		GroupARN: aws.String(d.Id()),
	}
	_, err := conn.DeleteGroup(params)
	if err != nil {
		return fmt.Errorf("error deleting XRay Group (%s): %w", d.Id(), err)
	}

	return nil
}
