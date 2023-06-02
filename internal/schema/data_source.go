package schema

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// Adapted from https://github.com/hashicorp/terraform-provider-google/google/datasource_helpers.go. Thanks!

// dataSourceSchemaFromResourceSchema is a recursive func that
// converts an existing Resource schema to a Datasource schema.
// All schema elements are copied, but certain attributes are ignored or changed:
// - all attributes have Computed = true
// - all attributes have ForceNew, Required = false
// - Validation funcs and attributes (e.g. MaxItems) are not copied
func DataSourceSchemaFromResourceSchema(rs map[string]*schema.Schema) map[string]*schema.Schema {
	ds := make(map[string]*schema.Schema, len(rs))

	for k, v := range rs {
		ds[k] = DataSourcePropertyFromResourceProperty(v)
	}

	return ds
}

func DataSourcePropertyFromResourceProperty(rs *schema.Schema) *schema.Schema {
	ds := &schema.Schema{
		Computed:    true,
		Description: rs.Description,
		Type:        rs.Type,
	}

	switch rs.Type {
	case schema.TypeSet:
		ds.Set = rs.Set
		fallthrough
	case schema.TypeList, schema.TypeMap:
		// List & Set types are generally used for 2 cases:
		// - a list/set of simple primitive values (e.g. list of strings)
		// - a sub resource
		// Maps are usually used for maps of simple primitives
		switch elem := rs.Elem.(type) {
		case *schema.Resource:
			// handle the case where the Element is a sub-resource
			ds.Elem = &schema.Resource{
				Schema: DataSourceSchemaFromResourceSchema(elem.Schema),
			}
		case *schema.Schema:
			// handle simple primitive case
			ds.Elem = &schema.Schema{Type: elem.Type}
		}
	}

	return ds
}
