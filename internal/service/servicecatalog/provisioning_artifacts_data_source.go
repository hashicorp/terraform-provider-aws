package servicecatalog

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceProvisioningArtifacts() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceProvisioningArtifactsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(ConstraintReadTimeout),
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Default:      AcceptLanguageEnglish,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(AcceptLanguage_Values(), false),
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"provisioning_artifact_details": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"created_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"guidance": {
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
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceProvisioningArtifactsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceCatalogConn()

	input := &servicecatalog.ListProvisioningArtifactsInput{
		ProductId:      aws.String(d.Get("product_id").(string)),
		AcceptLanguage: aws.String(d.Get("accept_language").(string)),
	}

	output, err := conn.ListProvisioningArtifacts(input)

	if err != nil {
		return fmt.Errorf("error describing provisioning artifact: %w", err)
	}
	if output == nil {
		return fmt.Errorf("no provisioning artifacts found matching criteria; try different search")
	}
	if err := d.Set("provisioning_artifact_details", flattenProvisioningArtifactDetails(output.ProvisioningArtifactDetails)); err != nil {
		return fmt.Errorf("error setting provisioning artifact details: %w", err)
	}

	d.SetId(d.Get("product_id").(string))
	d.Set("accept_language", d.Get("accept_language").(string))

	return nil
}

func flattenProvisioningArtifactDetails(apiObjects []*servicecatalog.ProvisioningArtifactDetail) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}
		tfList = append(tfList, flattenProvisioningArtifactDetail(apiObject))
	}

	return tfList
}

func flattenProvisioningArtifactDetail(apiObject *servicecatalog.ProvisioningArtifactDetail) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Active != nil {
		tfMap["active"] = aws.BoolValue(apiObject.Active)
	}
	if apiObject.CreatedTime != nil {
		tfMap["created_time"] = aws.TimeValue(apiObject.CreatedTime).String()
	}
	if apiObject.Description != nil {
		tfMap["description"] = aws.StringValue(apiObject.Description)
	}
	if apiObject.Guidance != nil {
		tfMap["guidance"] = aws.StringValue(apiObject.Guidance)
	}
	if apiObject.Id != nil {
		tfMap["id"] = aws.StringValue(apiObject.Id)
	}
	if apiObject.Name != nil {
		tfMap["name"] = aws.StringValue(apiObject.Name)
	}
	if apiObject.Type != nil {
		tfMap["type"] = aws.StringValue(apiObject.Type)
	}

	return tfMap
}
