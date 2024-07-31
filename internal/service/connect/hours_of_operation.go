// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"bytes"
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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_hours_of_operation", name="Hours Of Operation")
// @Tags(identifierAttribute="arn")
func ResourceHoursOfOperation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHoursOfOperationCreate,
		ReadWithoutTimeout:   resourceHoursOfOperationRead,
		UpdateWithoutTimeout: resourceHoursOfOperationUpdate,
		DeleteWithoutTimeout: resourceHoursOfOperationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 0,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"day": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(connect.HoursOfOperationDays_Values(), false),
						},
						"end_time": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hours": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"minutes": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						names.AttrStartTime: {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hours": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"minutes": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(m["day"].(string))
					buf.WriteString(fmt.Sprintf("%+v", m["end_time"].([]interface{})))
					buf.WriteString(fmt.Sprintf("%+v", m[names.AttrStartTime].([]interface{})))
					return create.StringHashcode(buf.String())
				},
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 250),
			},
			"hours_of_operation_id": {
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"time_zone": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceHoursOfOperationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	name := d.Get(names.AttrName).(string)
	config := expandConfigs(d.Get("config").(*schema.Set).List())
	input := &connect.CreateHoursOfOperationInput{
		Config:     config,
		InstanceId: aws.String(instanceID),
		Name:       aws.String(name),
		Tags:       getTagsIn(ctx),
		TimeZone:   aws.String(d.Get("time_zone").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Connect Hours of Operation %s", input)
	output, err := conn.CreateHoursOfOperationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Hours of Operation (%s): %s", name, err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Hours of Operation (%s): empty output", name)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(output.HoursOfOperationId)))

	return append(diags, resourceHoursOfOperationRead(ctx, d, meta)...)
}

func resourceHoursOfOperationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, hoursOfOperationID, err := HoursOfOperationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resp, err := conn.DescribeHoursOfOperationWithContext(ctx, &connect.DescribeHoursOfOperationInput{
		HoursOfOperationId: aws.String(hoursOfOperationID),
		InstanceId:         aws.String(instanceID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect Hours of Operation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Hours of Operation (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.HoursOfOperation == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Hours of Operation (%s): empty response", d.Id())
	}

	if err := d.Set("config", flattenConfigs(resp.HoursOfOperation.Config)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrARN, resp.HoursOfOperation.HoursOfOperationArn)
	d.Set("hours_of_operation_id", resp.HoursOfOperation.HoursOfOperationId)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrDescription, resp.HoursOfOperation.Description)
	d.Set(names.AttrName, resp.HoursOfOperation.Name)
	d.Set("time_zone", resp.HoursOfOperation.TimeZone)

	setTagsOut(ctx, resp.HoursOfOperation.Tags)

	return diags
}

func resourceHoursOfOperationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, hoursOfOperationID, err := HoursOfOperationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChanges("config", names.AttrDescription, names.AttrName, "time_zone") {
		_, err = conn.UpdateHoursOfOperationWithContext(ctx, &connect.UpdateHoursOfOperationInput{
			Config:             expandConfigs(d.Get("config").(*schema.Set).List()),
			Description:        aws.String(d.Get(names.AttrDescription).(string)),
			HoursOfOperationId: aws.String(hoursOfOperationID),
			InstanceId:         aws.String(instanceID),
			Name:               aws.String(d.Get(names.AttrName).(string)),
			TimeZone:           aws.String(d.Get("time_zone").(string)),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating HoursOfOperation (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceHoursOfOperationRead(ctx, d, meta)...)
}

func resourceHoursOfOperationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, hoursOfOperationID, err := HoursOfOperationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.DeleteHoursOfOperationWithContext(ctx, &connect.DeleteHoursOfOperationInput{
		HoursOfOperationId: aws.String(hoursOfOperationID),
		InstanceId:         aws.String(instanceID),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting HoursOfOperation (%s): %s", d.Id(), err)
	}

	return diags
}

func expandConfigs(configs []interface{}) []*connect.HoursOfOperationConfig {
	if len(configs) == 0 {
		return nil
	}

	hoursOfOperationConfigs := []*connect.HoursOfOperationConfig{}
	for _, config := range configs {
		data := config.(map[string]interface{})
		hoursOfOperationConfig := &connect.HoursOfOperationConfig{
			Day: aws.String(data["day"].(string)),
		}

		tet := data["end_time"].([]interface{})
		vet := tet[0].(map[string]interface{})
		et := connect.HoursOfOperationTimeSlice{
			Hours:   aws.Int64(int64(vet["hours"].(int))),
			Minutes: aws.Int64(int64(vet["minutes"].(int))),
		}
		hoursOfOperationConfig.EndTime = &et

		tst := data[names.AttrStartTime].([]interface{})
		vst := tst[0].(map[string]interface{})
		st := connect.HoursOfOperationTimeSlice{
			Hours:   aws.Int64(int64(vst["hours"].(int))),
			Minutes: aws.Int64(int64(vst["minutes"].(int))),
		}
		hoursOfOperationConfig.StartTime = &st

		hoursOfOperationConfigs = append(hoursOfOperationConfigs, hoursOfOperationConfig)
	}

	return hoursOfOperationConfigs
}

func flattenConfigs(configs []*connect.HoursOfOperationConfig) []interface{} {
	configsList := []interface{}{}
	for _, config := range configs {
		values := map[string]interface{}{}
		values["day"] = aws.StringValue(config.Day)

		et := map[string]interface{}{
			"hours":   aws.Int64Value(config.EndTime.Hours),
			"minutes": aws.Int64Value(config.EndTime.Minutes),
		}
		values["end_time"] = []interface{}{et}

		st := map[string]interface{}{
			"hours":   aws.Int64Value(config.StartTime.Hours),
			"minutes": aws.Int64Value(config.StartTime.Minutes),
		}
		values[names.AttrStartTime] = []interface{}{st}
		configsList = append(configsList, values)
	}
	return configsList
}

func HoursOfOperationParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:hoursOfOperationID", id)
	}

	return parts[0], parts[1], nil
}
