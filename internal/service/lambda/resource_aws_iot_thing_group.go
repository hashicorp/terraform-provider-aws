package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"log"
)

func resourceAwsIotThingGroup() *schema.Resource{
	return &schema.Resource{
		Create: resourceAwsIotThingGroupCreate,
		Read: resourceAwsIotThingGroupRead,
		Update: resourceAwsIotThingGroupUpdate,
		Delete: resourceAwsIotThingGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
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
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attributes": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
			"default_client_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsIotThingGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn
	input := &iot.CreateThingGroupInput{
		ThingGroupName: aws.String(d.Get("name").(string)),
		Tags:           keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().IotTags(),
	}

	if v, ok := d.GetOk("parent_group_name"); ok {
		input.ParentGroupName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("properties"); ok {
		input.ThingGroupProperties = expandIotThingsGroupProperties(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating IoT Thing Group: %s", input)
	out, err := conn.CreateThingGroup(input)
	if err != nil {
		return err
	}

	d.SetId(*out.ThingGroupName)
	return resourceAwsIotThingGroupRead(d, meta)
}

func resourceAwsIotThingGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	input := &iot.DescribeThingGroupInput{
		ThingGroupName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading IoT Thing Group: %s", input)
	out, err := conn.DescribeThingGroup(input)

	if err != nil {
		if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] IoT Thing Group %q not found, removing from state", d.Id())
			d.SetId("")
		}
		return err
	}

	log.Printf("[DEBUG] Received IoT Thing Group: %s", out)

	d.Set("arn", out.ThingGroupArn)
	d.Set("name", out.ThingGroupName)
	if err := d.Set("metadata", flattenIotThingGroupMetadata(out.ThingGroupMetadata)); err != nil {
		return fmt.Errorf("error setting metadata #{err}")
	}
	if err := d.Set("properties", flattenIotThingGroupProperties(out.ThingGroupProperties)); err != nil {
		return fmt.Errorf("error setting properties #{err}")
	}
	d.Set("version", out.Version)

	return nil
}

func resourceAwsIotThingGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	input := &iot.UpdateThingGroupInput{
		ThingGroupName: aws.String(d.Get("name").(string)),
	}

	if d.HasChange("properties") {
		if v, ok := d.GetOk("properties"); ok {
			input.ThingGroupProperties = expandIotThingsGroupProperties(v.([]interface{}))
		}
	}

	_, err := conn.UpdateThingGroup(input)
	if err != nil {
		return err
	}

	return resourceAwsIotThingGroupRead(d, meta)
}

func resourceAwsIotThingGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	input := &iot.DeleteThingGroupInput{
		ThingGroupName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting IoT Thing Group: %s", input)

	_, err := conn.DeleteThingGroup(input)
	if err != nil {
		if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
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
			Attributes: stringMapToPointers(v.(map[string]interface{})),
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
		"attributes": aws.StringValueMap(properties.AttributePayload.Attributes),
		"description": aws.StringValue(properties.ThingGroupDescription),
	}

	return []map[string]interface{}{props}
}

func flattenIotThingGroupMetadata(metadata *iot.ThingGroupMetadata) []map[string]interface{} {
	if metadata == nil {
		return []map[string]interface{}{}
	}

	meta := map[string]interface{}{
		"creation_date": aws.TimeValue(metadata.CreationDate),
		"parent_group_name": aws.StringValue(metadata.ParentGroupName),
		"root_to_parent_groups": expandIotGroupNameAndArnList(metadata.RootToParentThingGroups),
	}

	return []map[string]interface{}{meta}
}

func expandIotGroupNameAndArnList(lgn []*iot.GroupNameAndArn) []*iot.GroupNameAndArn {
	vs := make([]*iot.GroupNameAndArn, 0, len(lgn))
	for _, v := range lgn {
		val, ok := interface{}(v).(iot.GroupNameAndArn)
		if ok && &val != nil {
			vs = append(vs, &iot.GroupNameAndArn{
				GroupName: val.GroupName,
				GroupArn: val.GroupArn,
			})
		}
	}
	return vs
}