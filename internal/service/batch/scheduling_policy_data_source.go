package batch

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceSchedulingPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSchedulingPolicyRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"fair_share_policy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compute_reservation": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"share_decay_seconds": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"share_distribution": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"share_identifier": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"weight_factor": {
										Type:     schema.TypeFloat,
										Computed: true,
									},
								},
							},
							Set: func(v interface{}) int {
								var buf bytes.Buffer
								m := v.(map[string]interface{})
								buf.WriteString(m["share_identifier"].(string))
								if v, ok := m["weight_factor"]; ok {
									buf.WriteString(fmt.Sprintf("%s-", v))
								}
								return create.StringHashcode(buf.String())
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}
