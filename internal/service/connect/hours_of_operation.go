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
