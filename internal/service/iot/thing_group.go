package iot

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceThingGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceThingGroupCreate,
		Read:   resourceThingGroupRead,
		Update: resourceThingGroupUpdate,
		Delete: resourceThingGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"parent_group_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"properties": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attributes": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"metadata": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"creation_date": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"parent_group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"root_to_parent_groups": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"group_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"group_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceThingGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &iot.CreateThingGroupInput{
		ThingGroupName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("parent_group_name"); ok {
		input.ParentGroupName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("properties"); ok {
		input.ThingGroupProperties = expandIotThingsGroupProperties(v.([]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating IoT Thing Group: %s", input)
	out, err := conn.CreateThingGroup(input)
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(out.ThingGroupName))
	return resourceThingGroupRead(d, meta)
}

func resourceThingGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindThingGroupByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Thing Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IoT Thing Group (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.ThingGroupArn)
	d.Set("name", output.ThingGroupName)

	if err := d.Set("metadata", flattenIotThingGroupMetadata(output.ThingGroupMetadata)); err != nil {
		return fmt.Errorf("error setting metadata: %s", err)
	}
	if err := d.Set("properties", flattenIotThingGroupProperties(output.ThingGroupProperties)); err != nil {
		return fmt.Errorf("error setting properties: %s", err)
	}
	d.Set("version", output.Version)

	tags, err := ListTags(conn, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("error listing tags for IoT Thing Group (%s): %w", d.Get("arn").(string), err)
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

func resourceThingGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	input := &iot.UpdateThingGroupInput{
		ThingGroupName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("properties"); ok {
		input.ThingGroupProperties = expandIotThingsGroupProperties(v.([]interface{}))
	}

	_, err := conn.UpdateThingGroup(input)
	if err != nil {
		return err
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceThingGroupRead(d, meta)
}

func resourceThingGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	input := &iot.DeleteThingGroupInput{
		ThingGroupName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting IoT Thing Group: %s", input)

	_, err := conn.DeleteThingGroup(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
			return nil
		}
		return err
	}

	return nil
}

func expandIotThingsGroupProperties(l []interface{}) *iot.ThingGroupProperties {
	m := l[0].(map[string]interface{})

	thingGroupProperties := &iot.ThingGroupProperties{}

	if v, ok := m["attributes"]; ok {
		thingGroupProperties.AttributePayload = &iot.AttributePayload{
			Attributes: flex.ExpandStringMap(v.(map[string]interface{})),
		}
	}

	if v, ok := m["description"]; ok {
		thingGroupProperties.ThingGroupDescription = aws.String(v.(string))
	}

	return thingGroupProperties
}

func flattenIotThingGroupProperties(properties *iot.ThingGroupProperties) []map[string]interface{} {
	if properties == nil {
		return []map[string]interface{}{}
	}

	props := map[string]interface{}{
		"description": aws.StringValue(properties.ThingGroupDescription),
	}

	if properties.AttributePayload != nil {
		props["attributes"] = aws.StringValueMap(properties.AttributePayload.Attributes)
	}

	return []map[string]interface{}{props}
}

func flattenIotThingGroupMetadata(metadata *iot.ThingGroupMetadata) []map[string]interface{} {
	if metadata == nil {
		return []map[string]interface{}{}
	}

	meta := map[string]interface{}{
		"creation_date":         aws.TimeValue(metadata.CreationDate).Unix(),
		"parent_group_name":     aws.StringValue(metadata.ParentGroupName),
		"root_to_parent_groups": expandIotGroupNameAndArnList(metadata.RootToParentThingGroups),
	}

	return []map[string]interface{}{meta}
}

func expandIotGroupNameAndArnList(lgn []*iot.GroupNameAndArn) []*iot.GroupNameAndArn {
	vs := make([]*iot.GroupNameAndArn, 0, len(lgn))
	for _, v := range lgn {
		val, ok := interface{}(v).(iot.GroupNameAndArn)
		if ok {
			vs = append(vs, &iot.GroupNameAndArn{
				GroupName: val.GroupName,
				GroupArn:  val.GroupArn,
			})
		}
	}
	return vs
}
