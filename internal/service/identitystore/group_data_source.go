package identitystore

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGroupRead,

		Schema: map[string]*schema.Schema{
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"filter": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"attribute_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexp.MustCompile(`^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$`), "must match ([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}"),
				),
			},

			"identity_store_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]*$`), "must match [a-zA-Z0-9-]"),
				),
			},
		},
	}
}

const (
	DSNameGroup = "Group Data Source"
)

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreConn

	identityStoreId := d.Get("identity_store_id").(string)

	// Filters has been marked as deprecated in favour of GetGroupId, which
	// allows only a single filter. Keep using it to maintain backwards
	// compatibility of the data source.

	input := &identitystore.ListGroupsInput{
		IdentityStoreId: aws.String(identityStoreId),
		Filters:         expandFilters(d.Get("filter").(*schema.Set).List()),
	}

	var results []types.Group

	paginator := identitystore.NewListGroupsPaginator(conn, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return create.DiagError(names.IdentityStore, create.ErrActionReading, DSNameGroup, identityStoreId, err)
		}

		for _, group := range page.Groups {
			if v, ok := d.GetOk("group_id"); ok && v.(string) != aws.ToString(group.GroupId) {
				continue
			}

			results = append(results, group)
		}
	}

	if len(results) == 0 {
		return diag.Errorf("no Identity Store Group found matching criteria\n%v; try different search", input.Filters)
	}

	if len(results) > 1 {
		return diag.Errorf("multiple Identity Store Groups found matching criteria\n%v; try different search", input.Filters)
	}

	group := results[0]

	d.SetId(aws.ToString(group.GroupId))
	d.Set("display_name", group.DisplayName)
	d.Set("group_id", group.GroupId)

	return nil
}

func expandFilters(l []interface{}) []types.Filter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	filters := make([]types.Filter, 0, len(l))
	for _, v := range l {
		tfMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		filter := types.Filter{}

		if v, ok := tfMap["attribute_path"].(string); ok && v != "" {
			filter.AttributePath = aws.String(v)
		}

		if v, ok := tfMap["attribute_value"].(string); ok && v != "" {
			filter.AttributeValue = aws.String(v)
		}

		filters = append(filters, filter)
	}

	return filters
}
