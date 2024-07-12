// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_quick_connect", name="Quick Connect")
// @Tags(identifierAttribute="arn")
func ResourceQuickConnect() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQuickConnectCreate,
		ReadWithoutTimeout:   resourceQuickConnectRead,
		UpdateWithoutTimeout: resourceQuickConnectUpdate,
		DeleteWithoutTimeout: resourceQuickConnectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: verify.SetTagsDiff,
		Schema: map[string]*schema.Schema{
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 250),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"quick_connect_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 127),
			},
			"quick_connect_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"phone_config": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"phone_number": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if v := d.Get("quick_connect_config.0.quick_connect_type").(string); v == connect.QuickConnectTypePhoneNumber {
									return false
								}
								return true
							},
						},
						"queue_config": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"contact_flow_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"queue_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if v := d.Get("quick_connect_config.0.quick_connect_type").(string); v == connect.QuickConnectTypeQueue {
									return false
								}
								return true
							},
						},
						"quick_connect_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(connect.QuickConnectType_Values(), false),
						},
						"user_config": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"contact_flow_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"user_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if v := d.Get("quick_connect_config.0.quick_connect_type").(string); v == connect.QuickConnectTypeUser {
									return false
								}
								return true
							},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceQuickConnectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	name := d.Get(names.AttrName).(string)
	quickConnectConfig := expandQuickConnectConfig(d.Get("quick_connect_config").([]interface{}))
	input := &connect.CreateQuickConnectInput{
		QuickConnectConfig: quickConnectConfig,
		InstanceId:         aws.String(instanceID),
		Name:               aws.String(name),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Connect Quick Connect %s", input)
	output, err := conn.CreateQuickConnectWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Quick Connect (%s): %s", name, err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Quick Connect (%s): empty output", name)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(output.QuickConnectId)))

	return append(diags, resourceQuickConnectRead(ctx, d, meta)...)
}

func resourceQuickConnectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, quickConnectID, err := QuickConnectParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resp, err := conn.DescribeQuickConnectWithContext(ctx, &connect.DescribeQuickConnectInput{
		InstanceId:     aws.String(instanceID),
		QuickConnectId: aws.String(quickConnectID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect Quick Connect (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Quick Connect (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.QuickConnect == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Quick Connect (%s): empty response", d.Id())
	}

	if err := d.Set("quick_connect_config", flattenQuickConnectConfig(resp.QuickConnect.QuickConnectConfig)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrDescription, resp.QuickConnect.Description)
	d.Set(names.AttrName, resp.QuickConnect.Name)
	d.Set(names.AttrARN, resp.QuickConnect.QuickConnectARN)
	d.Set("quick_connect_id", resp.QuickConnect.QuickConnectId)

	setTagsOut(ctx, resp.QuickConnect.Tags)

	return diags
}

func resourceQuickConnectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, quickConnectID, err := QuickConnectParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// QuickConnect has 2 update APIs
	// UpdateQuickConnectNameWithContext: Updates the name and description of a quick connect.
	// UpdateQuickConnectConfigWithContext: Updates the configuration settings for the specified quick connect.

	// updates to name and/or description
	inputNameDesc := &connect.UpdateQuickConnectNameInput{
		InstanceId:     aws.String(instanceID),
		QuickConnectId: aws.String(quickConnectID),
	}

	// Either QuickConnectName or QuickConnectDescription must be specified. Both cannot be null or empty
	if d.HasChanges(names.AttrName, names.AttrDescription) {
		inputNameDesc.Name = aws.String(d.Get(names.AttrName).(string))
		inputNameDesc.Description = aws.String(d.Get(names.AttrDescription).(string))
		_, err = conn.UpdateQuickConnectNameWithContext(ctx, inputNameDesc)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickConnect Name (%s): %s", d.Id(), err)
		}
	}

	// updates to configuration settings
	inputConfig := &connect.UpdateQuickConnectConfigInput{
		InstanceId:     aws.String(instanceID),
		QuickConnectId: aws.String(quickConnectID),
	}

	// QuickConnectConfig is a required field but does not require update if it is unchanged
	if d.HasChange("quick_connect_config") {
		quickConnectConfig := expandQuickConnectConfig(d.Get("quick_connect_config").([]interface{}))
		inputConfig.QuickConnectConfig = quickConnectConfig
		_, err = conn.UpdateQuickConnectConfigWithContext(ctx, inputConfig)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickConnect (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceQuickConnectRead(ctx, d, meta)...)
}

func resourceQuickConnectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, quickConnectID, err := QuickConnectParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.DeleteQuickConnectWithContext(ctx, &connect.DeleteQuickConnectInput{
		InstanceId:     aws.String(instanceID),
		QuickConnectId: aws.String(quickConnectID),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickConnect (%s): %s", d.Id(), err)
	}

	return diags
}

func expandQuickConnectConfig(quickConnectConfig []interface{}) *connect.QuickConnectConfig {
	if len(quickConnectConfig) == 0 || quickConnectConfig[0] == nil {
		return nil
	}

	tfMap, ok := quickConnectConfig[0].(map[string]interface{})
	if !ok {
		return nil
	}

	quickConnectType := tfMap["quick_connect_type"].(string)

	result := &connect.QuickConnectConfig{
		QuickConnectType: aws.String(quickConnectType),
	}

	switch quickConnectType {
	case connect.QuickConnectTypePhoneNumber:
		tpc := tfMap["phone_config"].([]interface{})
		if len(tpc) == 0 || tpc[0] == nil {
			log.Printf("[ERR] 'phone_config' must be set when 'quick_connect_type' is '%s'", quickConnectType)
			return nil
		}
		vpc := tpc[0].(map[string]interface{})
		pc := connect.PhoneNumberQuickConnectConfig{
			PhoneNumber: aws.String(vpc["phone_number"].(string)),
		}
		result.PhoneConfig = &pc

	case connect.QuickConnectTypeQueue:
		tqc := tfMap["queue_config"].([]interface{})
		if len(tqc) == 0 || tqc[0] == nil {
			log.Printf("[ERR] 'queue_config' must be set when 'quick_connect_type' is '%s'", quickConnectType)
			return nil
		}
		vqc := tqc[0].(map[string]interface{})
		qc := connect.QueueQuickConnectConfig{
			ContactFlowId: aws.String(vqc["contact_flow_id"].(string)),
			QueueId:       aws.String(vqc["queue_id"].(string)),
		}
		result.QueueConfig = &qc

	case connect.QuickConnectTypeUser:
		tuc := tfMap["user_config"].([]interface{})
		if len(tuc) == 0 || tuc[0] == nil {
			log.Printf("[ERR] 'user_config' must be set when 'quick_connect_type' is '%s'", quickConnectType)
			return nil
		}
		vuc := tuc[0].(map[string]interface{})
		uc := connect.UserQuickConnectConfig{
			ContactFlowId: aws.String(vuc["contact_flow_id"].(string)),
			UserId:        aws.String(vuc["user_id"].(string)),
		}
		result.UserConfig = &uc

	default:
		log.Printf("[ERR] quick_connect_type is invalid")
		return nil
	}

	return result
}

func flattenQuickConnectConfig(quickConnectConfig *connect.QuickConnectConfig) []interface{} {
	if quickConnectConfig == nil {
		return []interface{}{}
	}

	quickConnectType := aws.StringValue(quickConnectConfig.QuickConnectType)

	values := map[string]interface{}{
		"quick_connect_type": quickConnectType,
	}

	switch quickConnectType {
	case connect.QuickConnectTypePhoneNumber:
		pc := map[string]interface{}{
			"phone_number": aws.StringValue(quickConnectConfig.PhoneConfig.PhoneNumber),
		}
		values["phone_config"] = []interface{}{pc}

	case connect.QuickConnectTypeQueue:
		qc := map[string]interface{}{
			"contact_flow_id": aws.StringValue(quickConnectConfig.QueueConfig.ContactFlowId),
			"queue_id":        aws.StringValue(quickConnectConfig.QueueConfig.QueueId),
		}
		values["queue_config"] = []interface{}{qc}

	case connect.QuickConnectTypeUser:
		uc := map[string]interface{}{
			"contact_flow_id": aws.StringValue(quickConnectConfig.UserConfig.ContactFlowId),
			"user_id":         aws.StringValue(quickConnectConfig.UserConfig.UserId),
		}
		values["user_config"] = []interface{}{uc}

	default:
		log.Printf("[ERR] quick_connect_type is invalid")
		return nil
	}

	return []interface{}{values}
}

func QuickConnectParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:quickConnectID", id)
	}

	return parts[0], parts[1], nil
}
