package emr

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func DataSourceReleaseLabels() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReleaseLabelsRead,
		Schema: map[string]*schema.Schema{
			"filters": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"application": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"release_labels": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func dataSourceReleaseLabelsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EMRConn

	input := &emr.ListReleaseLabelsInput{}

	if v, ok := d.GetOk("filters"); ok && len(v.([]interface{})) > 0 {
		input.Filters = expandReleaseLabelsFilters(v.([]interface{}))
	}

	out, err := conn.ListReleaseLabels(input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading EMR Release Label: %w", err))
	}

	if len(out.ReleaseLabels) == 0 {
		return diag.Errorf("no EMR release labels found")
	}

	d.SetId(strings.Join(aws.StringValueSlice(out.ReleaseLabels), ","))
	d.Set("release_labels", flex.FlattenStringSet(out.ReleaseLabels))

	return nil
}

func expandReleaseLabelsFilters(filters []interface{}) *emr.ReleaseLabelFilter {
	if len(filters) == 0 || filters[0] == nil {
		return nil
	}

	m := filters[0].(map[string]interface{})
	app := &emr.ReleaseLabelFilter{}

	if v, ok := m["application"].(string); ok && v != "" {
		app.Application = aws.String(v)
	}

	if v, ok := m["prefix"].(string); ok && v != "" {
		app.Prefix = aws.String(v)
	}

	return app
}
