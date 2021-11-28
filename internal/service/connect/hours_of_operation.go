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

func ResourceHoursOfOperation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHoursOfOperationCreate,
		ReadContext:   resourceHoursOfOperationRead,
		UpdateContext: resourceHoursOfOperationUpdate,
		DeleteContext: resourceHoursOfOperationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(connectHoursOfOperationCreatedTimeout),
			Delete: schema.DefaultTimeout(connectHoursOfOperationDeletedTimeout),
		},
		Schema: map[string]*schema.Schema{
			"config": {
				Type:     schema.TypeList,
				Required: true,
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
						"start_time": {
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hours_of_operation_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hours_of_operation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"time_zone": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}


func expandConfigs(configs []interface{}) ([]*connect.HoursOfOperationConfig, error) {
	if len(configs) == 0 {
		return nil, nil
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

		tst := data["start_time"].([]interface{})
		vst := tst[0].(map[string]interface{})
		st := connect.HoursOfOperationTimeSlice{
			Hours:   aws.Int64(int64(vst["hours"].(int))),
			Minutes: aws.Int64(int64(vst["minutes"].(int))),
		}
		hoursOfOperationConfig.StartTime = &st

		hoursOfOperationConfigs = append(hoursOfOperationConfigs, hoursOfOperationConfig)
	}

	return hoursOfOperationConfigs, nil
}

func flattenConfigs(configs []*connect.HoursOfOperationConfig, d *schema.ResourceData) []interface{} {
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
		values["start_time"] = []interface{}{st}
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
