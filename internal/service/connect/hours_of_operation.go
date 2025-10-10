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

// @SDKResource("aws_connect_hours_of_operation", name="Hours Of Operation")
// @Tags(identifierAttribute="arn")
func resourceHoursOfOperation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHoursOfOperationCreate,
		ReadWithoutTimeout:   resourceHoursOfOperationRead,
		UpdateWithoutTimeout: resourceHoursOfOperationUpdate,
		DeleteWithoutTimeout: resourceHoursOfOperationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.HoursOfOperationDays](),
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

func resourceHoursOfOperationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	name := d.Get(names.AttrName).(string)
	input := &connect.CreateHoursOfOperationInput{
		Config:     expandHoursOfOperationConfigs(d.Get("config").(*schema.Set).List()),
		InstanceId: aws.String(instanceID),
		Name:       aws.String(name),
		Tags:       getTagsIn(ctx),
		TimeZone:   aws.String(d.Get("time_zone").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateHoursOfOperation(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Hours Of Operation (%s): %s", name, err)
	}

	id := hoursOfOperationCreateResourceID(instanceID, aws.ToString(output.HoursOfOperationId))
	d.SetId(id)

	return append(diags, resourceHoursOfOperationRead(ctx, d, meta)...)
}

func resourceHoursOfOperationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, hoursOfOperationID, err := hoursOfOperationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	hoursOfOperation, err := findHoursOfOperationByTwoPartKey(ctx, conn, instanceID, hoursOfOperationID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Hours Of Operation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Hours Of Operation (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, hoursOfOperation.HoursOfOperationArn)
	if err := d.Set("config", flattenHoursOfOperationConfigs(hoursOfOperation.Config)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set(names.AttrDescription, hoursOfOperation.Description)
	d.Set("hours_of_operation_id", hoursOfOperation.HoursOfOperationId)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrName, hoursOfOperation.Name)
	d.Set("time_zone", hoursOfOperation.TimeZone)

	setTagsOut(ctx, hoursOfOperation.Tags)

	return diags
}

func resourceHoursOfOperationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, hoursOfOperationID, err := hoursOfOperationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChanges("config", names.AttrDescription, names.AttrName, "time_zone") {
		input := &connect.UpdateHoursOfOperationInput{
			Config:             expandHoursOfOperationConfigs(d.Get("config").(*schema.Set).List()),
			Description:        aws.String(d.Get(names.AttrDescription).(string)),
			HoursOfOperationId: aws.String(hoursOfOperationID),
			InstanceId:         aws.String(instanceID),
			Name:               aws.String(d.Get(names.AttrName).(string)),
			TimeZone:           aws.String(d.Get("time_zone").(string)),
		}

		_, err = conn.UpdateHoursOfOperation(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating HoursOfOperation (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceHoursOfOperationRead(ctx, d, meta)...)
}

func resourceHoursOfOperationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, hoursOfOperationID, err := hoursOfOperationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Connect Hours Of Operation: %s", d.Id())
	input := connect.DeleteHoursOfOperationInput{
		HoursOfOperationId: aws.String(hoursOfOperationID),
		InstanceId:         aws.String(instanceID),
	}
	_, err = conn.DeleteHoursOfOperation(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect Hours Of Operation (%s): %s", d.Id(), err)
	}

	return diags
}

const hoursOfOperationResourceIDSeparator = ":"

func hoursOfOperationCreateResourceID(instanceID, hoursOfOperationID string) string {
	parts := []string{instanceID, hoursOfOperationID}
	id := strings.Join(parts, hoursOfOperationResourceIDSeparator)

	return id
}

func hoursOfOperationParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, hoursOfOperationResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected instanceID%[2]shoursOfOperationID", id, hoursOfOperationResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findHoursOfOperationByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, hoursOfOperationID string) (*awstypes.HoursOfOperation, error) {
	input := &connect.DescribeHoursOfOperationInput{
		HoursOfOperationId: aws.String(hoursOfOperationID),
		InstanceId:         aws.String(instanceID),
	}

	return findHoursOfOperation(ctx, conn, input)
}

func findHoursOfOperation(ctx context.Context, conn *connect.Client, input *connect.DescribeHoursOfOperationInput) (*awstypes.HoursOfOperation, error) {
	output, err := conn.DescribeHoursOfOperation(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.HoursOfOperation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.HoursOfOperation, nil
}

func expandHoursOfOperationConfigs(tfList []any) []awstypes.HoursOfOperationConfig {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := []awstypes.HoursOfOperationConfig{}

	for _, config := range tfList {
		tfMap := config.(map[string]any)
		apiObject := awstypes.HoursOfOperationConfig{
			Day: awstypes.HoursOfOperationDays(tfMap["day"].(string)),
		}

		if v, ok := tfMap["end_time"].([]any); ok && len(v) > 0 && v[0] != nil {
			tfMap := v[0].(map[string]any)

			apiObject.EndTime = &awstypes.HoursOfOperationTimeSlice{
				Hours:   aws.Int32(int32(tfMap["hours"].(int))),
				Minutes: aws.Int32(int32(tfMap["minutes"].(int))),
			}
		}

		if v, ok := tfMap[names.AttrStartTime].([]any); ok && len(v) > 0 && v[0] != nil {
			tfMap := v[0].(map[string]any)

			apiObject.StartTime = &awstypes.HoursOfOperationTimeSlice{
				Hours:   aws.Int32(int32(tfMap["hours"].(int))),
				Minutes: aws.Int32(int32(tfMap["minutes"].(int))),
			}
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenHoursOfOperationConfigs(apiObjects []awstypes.HoursOfOperationConfig) []any {
	tfList := []any{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"day": apiObject.Day,
		}

		if v := apiObject.EndTime; v != nil {
			tfMap["end_time"] = []any{map[string]any{
				"hours":   aws.ToInt32(v.Hours),
				"minutes": aws.ToInt32(v.Minutes),
			}}
		}

		if v := apiObject.StartTime; v != nil {
			tfMap[names.AttrStartTime] = []any{map[string]any{
				"hours":   aws.ToInt32(v.Hours),
				"minutes": aws.ToInt32(v.Minutes),
			}}
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
