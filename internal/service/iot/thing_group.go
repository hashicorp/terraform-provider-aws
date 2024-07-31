// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_thing_group", name="Thing Group")
// @Tags(identifierAttribute="arn")
func resourceThingGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingGroupCreate,
		ReadWithoutTimeout:   resourceThingGroupRead,
		UpdateWithoutTimeout: resourceThingGroupUpdate,
		DeleteWithoutTimeout: resourceThingGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCreationDate: {
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
									names.AttrGroupName: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
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
			names.AttrProperties: {
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
									names.AttrAttributes: {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceThingGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &iot.CreateThingGroupInput{
		Tags:           getTagsIn(ctx),
		ThingGroupName: aws.String(name),
	}

	if v, ok := d.GetOk("parent_group_name"); ok {
		input.ParentGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrProperties); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ThingGroupProperties = expandThingGroupProperties(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreateThingGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Thing Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ThingGroupName))

	return append(diags, resourceThingGroupRead(ctx, d, meta)...)
}

func resourceThingGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	output, err := findThingGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Thing Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Thing Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.ThingGroupArn)
	d.Set(names.AttrName, output.ThingGroupName)

	if output.ThingGroupMetadata != nil {
		if err := d.Set("metadata", []interface{}{flattenThingGroupMetadata(output.ThingGroupMetadata)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting metadata: %s", err)
		}
	} else {
		d.Set("metadata", nil)
	}
	if v := flattenThingGroupProperties(output.ThingGroupProperties); len(v) > 0 {
		if err := d.Set(names.AttrProperties, []interface{}{v}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting properties: %s", err)
		}
	} else {
		d.Set(names.AttrProperties, nil)
	}

	if output.ThingGroupMetadata != nil {
		d.Set("parent_group_name", output.ThingGroupMetadata.ParentGroupName)
	} else {
		d.Set("parent_group_name", nil)
	}
	d.Set(names.AttrVersion, output.Version)

	return diags
}

func resourceThingGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &iot.UpdateThingGroupInput{
			ExpectedVersion: aws.Int64(int64(d.Get(names.AttrVersion).(int))),
			ThingGroupName:  aws.String(d.Get(names.AttrName).(string)),
		}

		if v, ok := d.GetOk(names.AttrProperties); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ThingGroupProperties = expandThingGroupProperties(v.([]interface{})[0].(map[string]interface{}))
		} else {
			input.ThingGroupProperties = &awstypes.ThingGroupProperties{}
		}

		// https://docs.aws.amazon.com/iot/latest/apireference/API_AttributePayload.html#API_AttributePayload_Contents:
		// "To remove an attribute, call UpdateThing with an empty attribute value."
		if input.ThingGroupProperties.AttributePayload == nil {
			input.ThingGroupProperties.AttributePayload = &awstypes.AttributePayload{
				Attributes: map[string]string{},
			}
		}

		_, err := conn.UpdateThingGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IoT Thing Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceThingGroupRead(ctx, d, meta)...)
}

func resourceThingGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	log.Printf("[DEBUG] Deleting IoT Thing Group: %s", d.Id())
	const (
		timeout = 1 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[*awstypes.InvalidRequestException](ctx, timeout,
		func() (interface{}, error) {
			return conn.DeleteThingGroup(ctx, &iot.DeleteThingGroupInput{
				ThingGroupName: aws.String(d.Id()),
			})
		})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Thing Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findThingGroupByName(ctx context.Context, conn *iot.Client, name string) (*iot.DescribeThingGroupOutput, error) {
	input := &iot.DescribeThingGroupInput{
		ThingGroupName: aws.String(name),
	}

	output, err := conn.DescribeThingGroup(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandThingGroupProperties(tfMap map[string]interface{}) *awstypes.ThingGroupProperties {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ThingGroupProperties{}

	if v, ok := tfMap["attribute_payload"].([]interface{}); ok && len(v) > 0 {
		apiObject.AttributePayload = expandAttributePayload(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
		apiObject.ThingGroupDescription = aws.String(v)
	}

	return apiObject
}

func expandAttributePayload(tfMap map[string]interface{}) *awstypes.AttributePayload {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AttributePayload{}

	if v, ok := tfMap[names.AttrAttributes].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.Attributes = flex.ExpandStringValueMap(v)
	}

	return apiObject
}

func flattenThingGroupMetadata(apiObject *awstypes.ThingGroupMetadata) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CreationDate; v != nil {
		tfMap[names.AttrCreationDate] = aws.ToTime(v).Format(time.RFC3339)
	}

	if v := apiObject.ParentGroupName; v != nil {
		tfMap["parent_group_name"] = aws.ToString(v)
	}

	if v := apiObject.RootToParentThingGroups; v != nil {
		tfMap["root_to_parent_groups"] = flattenGroupNameAndARNs(v)
	}

	return tfMap
}

func flattenGroupNameAndARN(apiObject awstypes.GroupNameAndArn) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.GroupArn; v != nil {
		tfMap["group_arn"] = aws.ToString(v)
	}

	if v := apiObject.GroupName; v != nil {
		tfMap[names.AttrGroupName] = aws.ToString(v)
	}

	return tfMap
}

func flattenGroupNameAndARNs(apiObjects []awstypes.GroupNameAndArn) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenGroupNameAndARN(apiObject))
	}

	return tfList
}

func flattenThingGroupProperties(apiObject *awstypes.ThingGroupProperties) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := flattenAttributePayload(apiObject.AttributePayload); len(v) > 0 {
		tfMap["attribute_payload"] = []interface{}{v}
	}

	if v := apiObject.ThingGroupDescription; v != nil {
		tfMap[names.AttrDescription] = aws.ToString(v)
	}

	return tfMap
}

func flattenAttributePayload(apiObject *awstypes.AttributePayload) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Attributes; v != nil {
		tfMap[names.AttrAttributes] = aws.StringMap(v)
	}

	return tfMap
}
