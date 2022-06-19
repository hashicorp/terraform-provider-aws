package kendra

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceIndex() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity_units": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query_capacity_units": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"storage_capacity_units": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"document_metadata_configuration_updates": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"relevance": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"duration": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"freshness": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"importance": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"rank_order": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"values_importance_map": {
										Type:     schema.TypeMap,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeInt},
									},
								},
							},
						},
						"search": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"displayable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"facetable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"searchable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"sortable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"edition": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"error_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`[a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,35}`),
					"Starts with an alphanumeric character. Subsequently, can contain alphanumeric characters and hyphens. Fixed length of 36.",
				),
			},
			"index_statistics": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"faq_statistics": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"indexed_question_answers_count": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"text_document_statistics": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"indexed_text_bytes": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"indexed_text_documents_count": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_side_encryption_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_context_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_group_resolution_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_group_resolution_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"user_token_configurations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"json_token_type_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"group_attribute_field": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"user_name_attribute_field": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"jwt_token_type_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"claim_regex": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"group_attribute_field": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"issuer": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"key_location": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"secrets_manager_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"url": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"user_name_attribute_field": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}
