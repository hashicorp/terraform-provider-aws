package servicecatalog

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceLaunchPaths() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceLaunchPathsRead,

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "en",
				ValidateFunc: validation.StringInSlice(AcceptLanguage_Values(), false),
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"summaries": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"constraint_summaries": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"description": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"path_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tags": tftags.TagsSchemaComputed(),
					},
				},
			},
		},
	}
}

func dataSourceLaunchPathsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceCatalogConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	summaries, err := WaitLaunchPathsReady(conn, d.Get("accept_language").(string), d.Get("product_id").(string))

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Launch Paths: %w", err)
	}

	if err := d.Set("summaries", flattenServiceCatalogLaunchPathSummaries(summaries, ignoreTagsConfig)); err != nil {
		return fmt.Errorf("error setting summaries: %w", err)
	}

	d.SetId(d.Get("product_id").(string))

	return nil
}

func flattenServiceCatalogLaunchPathSummary(apiObject *servicecatalog.LaunchPathSummary, ignoreTagsConfig *tftags.IgnoreConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if len(apiObject.ConstraintSummaries) > 0 {
		tfMap["constraint_summaries"] = flattenServiceCatalogConstraintSummaries(apiObject.ConstraintSummaries)
	}

	if apiObject.Id != nil {
		tfMap["path_id"] = aws.StringValue(apiObject.Id)
	}

	if apiObject.Name != nil {
		tfMap["name"] = aws.StringValue(apiObject.Name)
	}

	tags := KeyValueTags(apiObject.Tags)

	tfMap["tags"] = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()

	return tfMap
}

func flattenServiceCatalogLaunchPathSummaries(apiObjects []*servicecatalog.LaunchPathSummary, ignoreTagsConfig *tftags.IgnoreConfig) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenServiceCatalogLaunchPathSummary(apiObject, ignoreTagsConfig))
	}

	return tfList
}

func flattenServiceCatalogConstraintSummary(apiObject *servicecatalog.ConstraintSummary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Description != nil {
		tfMap["description"] = aws.StringValue(apiObject.Description)
	}

	if apiObject.Type != nil {
		tfMap["type"] = aws.StringValue(apiObject.Type)
	}

	return tfMap
}

func flattenServiceCatalogConstraintSummaries(apiObjects []*servicecatalog.ConstraintSummary) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenServiceCatalogConstraintSummary(apiObject))
	}

	return tfList
}
