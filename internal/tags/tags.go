package tags

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	tagsSchema *schema.Schema = &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
	tagsSchemaComputed *schema.Schema = &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Computed: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
	tagsSchemaForceNew *schema.Schema = &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		ForceNew: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
)

// TagsSchema returns the schema to use for tags.
func TagsSchema() *schema.Schema {
	return tagsSchema
}

func TagsSchemaComputed() *schema.Schema {
	return tagsSchemaComputed
}

func TagsSchemaForceNew() *schema.Schema {
	return tagsSchemaForceNew
}
