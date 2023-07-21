// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_thing_group", name="Thing Group")
// @Tags(identifierAttribute="arn")
func ResourceThingGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingGroupCreate,
		ReadWithoutTimeout:   resourceThingGroupRead,
		UpdateWithoutTimeout: resourceThingGroupUpdate,
		DeleteWithoutTimeout: resourceThingGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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

func resourceThingGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	name := d.Get("name").(string)
	input := &iot.CreateThingGroupInput{
		Tags:           getTagsIn(ctx),
		ThingGroupName: aws.String(name),
	}

	if v, ok := d.GetOk("parent_group_name"); ok {
		input.ParentGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ThingGroupProperties = expandThingGroupProperties(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreateThingGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Thing Group (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ThingGroupName))

	return append(diags, resourceThingGroupRead(ctx, d, meta)...)
}

func resourceThingGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	output, err := FindThingGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Thing Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Thing Group (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.ThingGroupArn)
	d.Set("name", output.ThingGroupName)

	if output.ThingGroupMetadata != nil {
		if err := d.Set("metadata", []interface{}{flattenThingGroupMetadata(output.ThingGroupMetadata)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting metadata: %s", err)
		}
	} else {
		d.Set("metadata", nil)
	}
	if v := flattenThingGroupProperties(output.ThingGroupProperties); len(v) > 0 {
		if err := d.Set("properties", []interface{}{v}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting properties: %s", err)
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

	return diags
}

func resourceThingGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

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
		_, err := conn.UpdateThingGroupWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IoT Thing Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceThingGroupRead(ctx, d, meta)...)
}

func resourceThingGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	log.Printf("[DEBUG] Deleting IoT Thing Group: %s", d.Id())
	_, err := tfresource.RetryWhen(ctx, thingGroupDeleteTimeout,
		func() (interface{}, error) {
			return conn.DeleteThingGroupWithContext(ctx, &iot.DeleteThingGroupInput{
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
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Thing Group (%s): %s", d.Id(), err)
	}

	return diags
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
