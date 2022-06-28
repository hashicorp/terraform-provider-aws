package ram

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceResourceShare() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceResourceShareRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"resource_owner": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ram.ResourceOwner_Values(), false),
			},

			"resource_share_status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ram.ResourceShareStatus_Values(), false),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"owning_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tftags.TagsSchemaComputed(),

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceResourceShareRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RAMConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	owner := d.Get("resource_owner").(string)

	filters, filtersOk := d.GetOk("filter")

	params := &ram.GetResourceSharesInput{
		Name:          aws.String(name),
		ResourceOwner: aws.String(owner),
	}

	if v, ok := d.GetOk("resource_share_status"); ok {
		params.ResourceShareStatus = aws.String(v.(string))
	}

	if filtersOk {
		params.TagFilters = buildTagFilters(filters.(*schema.Set))
	}

	for {
		resp, err := conn.GetResourceShares(params)

		if err != nil {
			return fmt.Errorf("Error retrieving resource share: empty response for: %s", params)
		}

		if len(resp.ResourceShares) > 1 {
			return fmt.Errorf("Multiple resource shares found for: %s", name)
		}

		if resp == nil || len(resp.ResourceShares) == 0 {
			return fmt.Errorf("No matching resource found: %w", err)
		}

		for _, r := range resp.ResourceShares {
			if aws.StringValue(r.Name) == name {
				d.SetId(aws.StringValue(r.ResourceShareArn))
				d.Set("arn", r.ResourceShareArn)
				d.Set("owning_account_id", r.OwningAccountId)
				d.Set("status", r.Status)

				if err := d.Set("tags", KeyValueTags(r.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
					return fmt.Errorf("error setting tags: %w", err)
				}

				break
			}
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return nil
}

func buildTagFilters(set *schema.Set) []*ram.TagFilter {
	var filters []*ram.TagFilter

	for _, v := range set.List() {
		m := v.(map[string]interface{})
		var filterValues []*string
		for _, e := range m["values"].([]interface{}) {
			filterValues = append(filterValues, aws.String(e.(string)))
		}
		filters = append(filters, &ram.TagFilter{
			TagKey:    aws.String(m["name"].(string)),
			TagValues: filterValues,
		})
	}

	return filters
}
