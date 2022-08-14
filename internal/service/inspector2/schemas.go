package inspector2

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dateFilter() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"end_inclusive": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.IsRFC3339Time,
				},
				"start_inclusive": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.IsRFC3339Time,
				},
			},
		},
	}
}

func numberFilter() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"lower_inclusive": {
					Type:     schema.TypeFloat,
					Optional: true,
				},
				"upper_inclusive": {
					Type:     schema.TypeFloat,
					Optional: true,
				},
			},
		},
	}
}

func portRangeFilter() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"begin_inclusive": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"end_inclusive": {
					Type:     schema.TypeInt,
					Optional: true,
				},
			},
		},
	}
}

func mapFilter() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"EQUALS"}, false),
				},
				"key": {
					Type:     schema.TypeString,
					Required: true,
				},
				"value": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func packageFilter() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"architecture": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"comparison": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice([]string{"EQUALS", "NOT_EQUALS", "PREFIX"}, false),
							},
							"value": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"epoch": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"lower_inclusive": {
								Type:     schema.TypeFloat,
								Optional: true,
							},
							"upper_inclusive": {
								Type:     schema.TypeFloat,
								Optional: true,
							},
						},
					},
				},
				"name": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"comparison": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice([]string{"EQUALS", "NOT_EQUALS", "PREFIX"}, false),
							},
							"value": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"release": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"comparison": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice([]string{"EQUALS", "NOT_EQUALS", "PREFIX"}, false),
							},
							"value": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"source_layer_hash": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"comparison": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice([]string{"EQUALS", "NOT_EQUALS", "PREFIX"}, false),
							},
							"value": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"version": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"comparison": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice([]string{"EQUALS", "NOT_EQUALS", "PREFIX"}, false),
							},
							"value": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
			},
		},
	}
}

func stringFilter() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"EQUALS", "NOT_EQUALS", "PREFIX"}, false),
				},
				"value": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func stringFilterWithValues(values []string) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"EQUALS", "NOT_EQUALS", "PREFIX"}, false),
				},
				"value": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(values, false),
				},
			},
		},
	}
}
