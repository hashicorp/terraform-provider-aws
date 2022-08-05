package iot

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
			"metadata": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"creation_date": {
							Type:     schema.TypeString,
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
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"parent_group_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"properties": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_payload": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"attributes": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	thingGroupDeleteTimeout = 1 * time.Minute
)

func resourceThingGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &iot.CreateThingGroupInput{
		ThingGroupName: aws.String(name),
	}

	if v, ok := d.GetOk("parent_group_name"); ok {
		input.ParentGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ThingGroupProperties = expandThingGroupProperties(v.([]interface{})[0].(map[string]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating IoT Thing Group: %s", input)
	output, err := conn.CreateThingGroup(input)

	if err != nil {
		return fmt.Errorf("error creating IoT Thing Group (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.ThingGroupName))

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

	if output.ThingGroupMetadata != nil {
		if err := d.Set("metadata", []interface{}{flattenThingGroupMetadata(output.ThingGroupMetadata)}); err != nil {
			return fmt.Errorf("error setting metadata: %w", err)
		}
	} else {
		d.Set("metadata", nil)
	}
	if v := flattenThingGroupProperties(output.ThingGroupProperties); len(v) > 0 {
		if err := d.Set("properties", []interface{}{v}); err != nil {
			return fmt.Errorf("error setting properties: %w", err)
		}
	} else {
		d.Set("properties", nil)
	}

	if output.ThingGroupMetadata != nil {
		d.Set("parent_group_name", output.ThingGroupMetadata.ParentGroupName)
	} else {
		d.Set("parent_group_name", nil)
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

	if d.HasChangesExcept("tags", "tags_all") {
		input := &iot.UpdateThingGroupInput{
			ExpectedVersion: aws.Int64(int64(d.Get("version").(int))),
			ThingGroupName:  aws.String(d.Get("name").(string)),
		}

		if v, ok := d.GetOk("properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ThingGroupProperties = expandThingGroupProperties(v.([]interface{})[0].(map[string]interface{}))
		} else {
			input.ThingGroupProperties = &iot.ThingGroupProperties{}
		}

		// https://docs.aws.amazon.com/iot/latest/apireference/API_AttributePayload.html#API_AttributePayload_Contents:
		// "To remove an attribute, call UpdateThing with an empty attribute value."
		if input.ThingGroupProperties.AttributePayload == nil {
			input.ThingGroupProperties.AttributePayload = &iot.AttributePayload{
				Attributes: map[string]*string{},
			}
		}

		log.Printf("[DEBUG] Updating IoT Thing Group: %s", input)
		_, err := conn.UpdateThingGroup(input)

		if err != nil {
			return fmt.Errorf("error updating IoT Thing Group (%s): %w", d.Id(), err)
		}
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

	log.Printf("[DEBUG] Deleting IoT Thing Group: %s", d.Id())
	_, err := tfresource.RetryWhen(thingGroupDeleteTimeout,
		func() (interface{}, error) {
			return conn.DeleteThingGroup(&iot.DeleteThingGroupInput{
				ThingGroupName: aws.String(d.Id()),
			})
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, iot.ErrCodeInvalidRequestException, "there are still child groups attached") {
				return true, err
			}

			return false, err
		})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting IoT Thing Group (%s): %w", d.Id(), err)
	}

	return nil
}

func expandThingGroupProperties(tfMap map[string]interface{}) *iot.ThingGroupProperties {
	if tfMap == nil {
		return nil
	}

	apiObject := &iot.ThingGroupProperties{}

	if v, ok := tfMap["attribute_payload"].([]interface{}); ok && len(v) > 0 {
		apiObject.AttributePayload = expandAttributePayload(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["description"].(string); ok && v != "" {
		apiObject.ThingGroupDescription = aws.String(v)
	}

	return apiObject
}

func expandAttributePayload(tfMap map[string]interface{}) *iot.AttributePayload {
	if tfMap == nil {
		return nil
	}

	apiObject := &iot.AttributePayload{}

	if v, ok := tfMap["attributes"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.Attributes = flex.ExpandStringMap(v)
	}

	return apiObject
}

func flattenThingGroupMetadata(apiObject *iot.ThingGroupMetadata) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CreationDate; v != nil {
		tfMap["creation_date"] = aws.TimeValue(v).Format(time.RFC3339)
	}

	if v := apiObject.ParentGroupName; v != nil {
		tfMap["parent_group_name"] = aws.StringValue(v)
	}

	if v := apiObject.RootToParentThingGroups; v != nil {
		tfMap["root_to_parent_groups"] = flattenGroupNameAndARNs(v)
	}

	return tfMap
}

func flattenGroupNameAndARN(apiObject *iot.GroupNameAndArn) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.GroupArn; v != nil {
		tfMap["group_arn"] = aws.StringValue(v)
	}

	if v := apiObject.GroupName; v != nil {
		tfMap["group_name"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenGroupNameAndARNs(apiObjects []*iot.GroupNameAndArn) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenGroupNameAndARN(apiObject))
	}

	return tfList
}

func flattenThingGroupProperties(apiObject *iot.ThingGroupProperties) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := flattenAttributePayload(apiObject.AttributePayload); len(v) > 0 {
		tfMap["attribute_payload"] = []interface{}{v}
	}

	if v := apiObject.ThingGroupDescription; v != nil {
		tfMap["description"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenAttributePayload(apiObject *iot.AttributePayload) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Attributes; v != nil {
		tfMap["attributes"] = aws.StringValueMap(v)
	}

	return tfMap
}
