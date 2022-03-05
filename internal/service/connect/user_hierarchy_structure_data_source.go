package connect

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceUserHierarchyStructure() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"hierarchy_structure": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"level_one": func() *schema.Schema {
							schema := connectUserHierarchyLevelDataSourceSchema()
							return schema
						}(),
						"level_two": func() *schema.Schema {
							schema := connectUserHierarchyLevelDataSourceSchema()
							return schema
						}(),
						"level_three": func() *schema.Schema {
							schema := connectUserHierarchyLevelDataSourceSchema()
							return schema
						}(),
						"level_four": func() *schema.Schema {
							schema := connectUserHierarchyLevelDataSourceSchema()
							return schema
						}(),
						"level_five": func() *schema.Schema {
							schema := connectUserHierarchyLevelDataSourceSchema()
							return schema
						}(),
					},
				},
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
		},
	}
}

// Each level shares the same schema
func connectUserHierarchyLevelDataSourceSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

