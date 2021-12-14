package connect

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceHoursOfOperation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHoursOfOperationRead,
		Schema: map[string]*schema.Schema{
			"config": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"day": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"end_time": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hours": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"minutes": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"start_time": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hours": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"minutes": {
										Type:     schema.TypeInt,
										Computed: true,
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
					buf.WriteString(fmt.Sprintf("%+v", m["start_time"].([]interface{})))
					return create.StringHashcode(buf.String())
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hours_of_operation_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hours_of_operation_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"hours_of_operation_id", "name"},
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "hours_of_operation_id"},
			},
			"tags": tftags.TagsSchemaComputed(),
			"time_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

