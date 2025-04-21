// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_quick_connect", name="Quick Connect")
// @Tags(identifierAttribute="arn")
func resourceQuickConnect() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQuickConnectCreate,
		ReadWithoutTimeout:   resourceQuickConnectRead,
		UpdateWithoutTimeout: resourceQuickConnectUpdate,
		DeleteWithoutTimeout: resourceQuickConnectDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 250),
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
								if v := awstypes.QuickConnectType(d.Get("quick_connect_config.0.quick_connect_type").(string)); v == awstypes.QuickConnectTypePhoneNumber {
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
								if v := awstypes.QuickConnectType(d.Get("quick_connect_config.0.quick_connect_type").(string)); v == awstypes.QuickConnectTypeQueue {
									return false
								}
								return true
							},
						},
						"quick_connect_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.QuickConnectType](),
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
								if v := awstypes.QuickConnectType(d.Get("quick_connect_config.0.quick_connect_type").(string)); v == awstypes.QuickConnectTypeUser {
									return false
								}
								return true
							},
						},
					},
				},
			},
			"quick_connect_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceQuickConnectCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	name := d.Get(names.AttrName).(string)
	input := &connect.CreateQuickConnectInput{
		InstanceId:         aws.String(instanceID),
		Name:               aws.String(name),
		QuickConnectConfig: expandQuickConnectConfig(d.Get("quick_connect_config").([]any)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateQuickConnect(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Quick Connect (%s): %s", name, err)
	}

	id := quickConnectCreateResourceID(instanceID, aws.ToString(output.QuickConnectId))
	d.SetId(id)

	return append(diags, resourceQuickConnectRead(ctx, d, meta)...)
}

func resourceQuickConnectRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, quickConnectID, err := quickConnectParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	quickConnect, err := findQuickConnectByTwoPartKey(ctx, conn, instanceID, quickConnectID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Quick Connect (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Quick Connect (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, quickConnect.QuickConnectARN)
	d.Set(names.AttrDescription, quickConnect.Description)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrName, quickConnect.Name)
	if err := d.Set("quick_connect_config", flattenQuickConnectConfig(quickConnect.QuickConnectConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting quick_connect_config: %s", err)
	}
	d.Set("quick_connect_id", quickConnect.QuickConnectId)

	setTagsOut(ctx, quickConnect.Tags)

	return diags
}

func resourceQuickConnectUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, quickConnectID, err := quickConnectParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// QuickConnect has 2 update APIs
	// UpdateQuickConnectNameWithContext: Updates the name and description of a quick connect.
	// UpdateQuickConnectConfigWithContext: Updates the configuration settings for the specified quick connect.

	// Either QuickConnectName or QuickConnectDescription must be specified. Both cannot be null or empty
	if d.HasChanges(names.AttrName, names.AttrDescription) {
		// updates to name and/or description
		input := &connect.UpdateQuickConnectNameInput{
			Description:    aws.String(d.Get(names.AttrDescription).(string)),
			InstanceId:     aws.String(instanceID),
			Name:           aws.String(d.Get(names.AttrName).(string)),
			QuickConnectId: aws.String(quickConnectID),
		}

		_, err = conn.UpdateQuickConnectName(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Quick Connect(%s) Name: %s", d.Id(), err)
		}
	}

	// QuickConnectConfig is a required field but does not require update if it is unchanged
	if d.HasChange("quick_connect_config") {
		// updates to configuration settings
		input := &connect.UpdateQuickConnectConfigInput{
			InstanceId:         aws.String(instanceID),
			QuickConnectConfig: expandQuickConnectConfig(d.Get("quick_connect_config").([]any)),
			QuickConnectId:     aws.String(quickConnectID),
		}

		_, err = conn.UpdateQuickConnectConfig(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Quick Connect (%s) Config: %s", d.Id(), err)
		}
	}

	return append(diags, resourceQuickConnectRead(ctx, d, meta)...)
}

func resourceQuickConnectDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, quickConnectID, err := quickConnectParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Connect Quick Connect: %s", d.Id())
	input := connect.DeleteQuickConnectInput{
		InstanceId:     aws.String(instanceID),
		QuickConnectId: aws.String(quickConnectID),
	}
	_, err = conn.DeleteQuickConnect(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect Quick Connect (%s): %s", d.Id(), err)
	}

	return diags
}

const quickConnectResourceIDSeparator = ":"

func quickConnectCreateResourceID(instanceID, routingProfileID string) string {
	parts := []string{instanceID, routingProfileID}
	id := strings.Join(parts, quickConnectResourceIDSeparator)

	return id
}

func quickConnectParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, quickConnectResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected instanceID%[2]squickConnectID", id, quickConnectResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findQuickConnectByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, quickConnectID string) (*awstypes.QuickConnect, error) {
	input := &connect.DescribeQuickConnectInput{
		InstanceId:     aws.String(instanceID),
		QuickConnectId: aws.String(quickConnectID),
	}

	return findQuickConnect(ctx, conn, input)
}

func findQuickConnect(ctx context.Context, conn *connect.Client, input *connect.DescribeQuickConnectInput) (*awstypes.QuickConnect, error) {
	output, err := conn.DescribeQuickConnect(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.QuickConnect == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.QuickConnect, nil
}

func expandQuickConnectConfig(tfList []any) *awstypes.QuickConnectConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	quickConnectType := awstypes.QuickConnectType(tfMap["quick_connect_type"].(string))
	apiObject := &awstypes.QuickConnectConfig{
		QuickConnectType: quickConnectType,
	}

	switch quickConnectType {
	case awstypes.QuickConnectTypePhoneNumber:
		v := tfMap["phone_config"].([]any)
		if len(v) == 0 || v[0] == nil {
			log.Printf("[ERR] 'phone_config' must be set when 'quick_connect_type' is '%s'", quickConnectType)
			return nil
		}

		tfMap := v[0].(map[string]any)
		apiObject.PhoneConfig = &awstypes.PhoneNumberQuickConnectConfig{
			PhoneNumber: aws.String(tfMap["phone_number"].(string)),
		}

	case awstypes.QuickConnectTypeQueue:
		v := tfMap["queue_config"].([]any)
		if len(v) == 0 || v[0] == nil {
			log.Printf("[ERR] 'queue_config' must be set when 'quick_connect_type' is '%s'", quickConnectType)
			return nil
		}

		tfMap := v[0].(map[string]any)
		apiObject.QueueConfig = &awstypes.QueueQuickConnectConfig{
			ContactFlowId: aws.String(tfMap["contact_flow_id"].(string)),
			QueueId:       aws.String(tfMap["queue_id"].(string)),
		}

	case awstypes.QuickConnectTypeUser:
		v := tfMap["user_config"].([]any)
		if len(v) == 0 || v[0] == nil {
			log.Printf("[ERR] 'user_config' must be set when 'quick_connect_type' is '%s'", quickConnectType)
			return nil
		}

		tfMap := v[0].(map[string]any)
		apiObject.UserConfig = &awstypes.UserQuickConnectConfig{
			ContactFlowId: aws.String(tfMap["contact_flow_id"].(string)),
			UserId:        aws.String(tfMap["user_id"].(string)),
		}

	default:
		log.Printf("[ERR] quick_connect_type is invalid")
		return nil
	}

	return apiObject
}

func flattenQuickConnectConfig(apiObject *awstypes.QuickConnectConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	quickConnectType := apiObject.QuickConnectType
	tfMap := map[string]any{
		"quick_connect_type": quickConnectType,
	}

	switch quickConnectType {
	case awstypes.QuickConnectTypePhoneNumber:
		tfMap["phone_config"] = []any{map[string]any{
			"phone_number": aws.ToString(apiObject.PhoneConfig.PhoneNumber),
		}}

	case awstypes.QuickConnectTypeQueue:
		tfMap["queue_config"] = []any{map[string]any{
			"contact_flow_id": aws.ToString(apiObject.QueueConfig.ContactFlowId),
			"queue_id":        aws.ToString(apiObject.QueueConfig.QueueId),
		}}

	case awstypes.QuickConnectTypeUser:
		tfMap["user_config"] = []any{map[string]any{
			"contact_flow_id": aws.ToString(apiObject.UserConfig.ContactFlowId),
			"user_id":         aws.ToString(apiObject.UserConfig.UserId),
		}}

	default:
		log.Printf("[ERR] quick_connect_type is invalid")
		return nil
	}

	return []any{tfMap}
}
