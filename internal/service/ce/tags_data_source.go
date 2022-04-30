package ce

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSourceTags() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTagsRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     schemaCostCategoryRule(),
			},
			"search_string": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringLenBetween(1, 1024),
				ConflictsWith: []string{"sort_by"},
			},
			"sort_by": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"search_string"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(costexplorer.Metric_Values(), false),
						},
						"sort_order": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(costexplorer.SortOrder_Values(), false),
						},
					},
				},
			},
			"tag_key": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"time_period": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"end": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 40),
						},
						"start": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 40),
						},
					},
				},
			},
		},
	}
}

func dataSourceTagsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn

	input := &costexplorer.GetTagsInput{
		TimePeriod: expandTagsTimePeriod(d.Get("time_period").([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := d.GetOk("filter"); ok {
		input.Filter = expandCostExpressions(v.([]interface{}))[0]
	}

	if v, ok := d.GetOk("search_string"); ok {
		input.SearchString = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sort_by"); ok {
		input.SortBy = expandTagsSortBys(v.([]interface{}))
	}

	if v, ok := d.GetOk("tag_key"); ok {
		input.TagKey = aws.String(v.(string))
	}

	resp, err := conn.GetTagsWithContext(ctx, input)

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionReading, ResTags, d.Id(), err)
	}

	d.Set("tags", flex.FlattenStringList(resp.Tags))

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return nil
}

func expandTagsSortBys(tfList []interface{}) []*costexplorer.SortDefinition {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.SortDefinition

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandTagsSortBy(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandTagsSortBy(tfMap map[string]interface{}) *costexplorer.SortDefinition {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.SortDefinition{}
	apiObject.Key = aws.String(tfMap["key"].(string))
	if v, ok := tfMap["sort_order"]; ok {
		apiObject.SortOrder = aws.String(v.(string))
	}

	return apiObject
}

func expandTagsTimePeriod(tfMap map[string]interface{}) *costexplorer.DateInterval {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.DateInterval{}
	apiObject.Start = aws.String(tfMap["start"].(string))
	apiObject.End = aws.String(tfMap["end"].(string))

	return apiObject
}
