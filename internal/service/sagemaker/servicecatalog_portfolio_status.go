package sagemaker

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceServicecatalogPortfolioStatus() *schema.Resource {
	return &schema.Resource{
		Create: resourceServicecatalogPortfolioStatusPut,
		Read:   resourceServicecatalogPortfolioStatusRead,
		Update: resourceServicecatalogPortfolioStatusPut,
		Delete: schema.Noop,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"status": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(sagemaker.SagemakerServicecatalogStatus_Values(), false),
			},
		},
	}
}

func resourceServicecatalogPortfolioStatusPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	status := d.Get("status").(string)
	var err error
	if status == sagemaker.SagemakerServicecatalogStatusEnabled {
		_, err = conn.EnableSagemakerServicecatalogPortfolio(&sagemaker.EnableSagemakerServicecatalogPortfolioInput{})
	} else {
		_, err = conn.DisableSagemakerServicecatalogPortfolio(&sagemaker.DisableSagemakerServicecatalogPortfolioInput{})
	}

	if err != nil {
		return fmt.Errorf("setting SageMaker Servicecatalog Portfolio Status: %w", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	return resourceServicecatalogPortfolioStatusRead(d, meta)
}

func resourceServicecatalogPortfolioStatusRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	resp, err := conn.GetSagemakerServicecatalogPortfolioStatus(&sagemaker.GetSagemakerServicecatalogPortfolioStatusInput{})
	if err != nil {
		return fmt.Errorf("Getting SageMaker Servicecatalog Portfolio Status: %w", err)
	}

	d.Set("status", resp.Status)

	return nil
}
