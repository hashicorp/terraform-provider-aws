package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceQuickConnect() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceQuickConnectCreate,
		ReadContext:   resourceQuickConnectRead,
		UpdateContext: resourceQuickConnectUpdate,
		DeleteContext: resourceQuickConnectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(connectQuickConnectCreatedTimeout),
			Delete: schema.DefaultTimeout(connectQuickConnectDeletedTimeout),
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 250),
			},
			"quick_connect_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"quick_connect_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
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
