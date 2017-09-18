package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsServiceCatalogPortfolio() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogPortfolioCreate,
		Read:   resourceAwsServiceCatalogPortfolioRead,
		Update: resourceAwsServiceCatalogPortfolioUpdate,
		Delete: resourceAwsServiceCatalogPortfolioDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"provider_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
		},
	}
}

func resourceAwsServiceCatalogPortfolioCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.CreatePortfolioInput{}
	if v, ok := d.GetOk("name"); ok {
		now := time.Now()
		input.IdempotencyToken = aws.String(fmt.Sprintf("%d", now.UnixNano()))
		input.DisplayName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provider_name"); ok {
		input.ProviderName = aws.String(v.(string))
	}

	resp, err := conn.CreatePortfolio(&input)
	if err != nil {
		return fmt.Errorf("Creating ServiceCatalog portfolio failed: %s", err.Error())
	}
	d.SetId(*resp.PortfolioDetail.Id)

	return resourceAwsServiceCatalogPortfolioRead(d, meta)
}

func resourceAwsServiceCatalogPortfolioRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.DescribePortfolioInput{}
	input.Id = aws.String(d.Id())

	resp, err := conn.DescribePortfolio(&input)
	if err != nil {
		return fmt.Errorf("Reading ServiceCatalog portfolio '%s' failed: %s", *input.Id, err.Error())
	}
	portfolioDetail := resp.PortfolioDetail

	d.Set("description", portfolioDetail.Description)
	d.Set("display_name", portfolioDetail.DisplayName)
	d.Set("id", d.Id())
	d.Set("provider_name", portfolioDetail.ProviderName)
	return nil
}

func resourceAwsServiceCatalogPortfolioUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.UpdatePortfolioInput{}
	input.Id = aws.String(d.Id())

	if v, ok := d.GetOk("name"); ok {
		input.DisplayName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provider_name"); ok {
		input.ProviderName = aws.String(v.(string))
	}

	_, err := conn.UpdatePortfolio(&input)
	if err != nil {
		return fmt.Errorf("Updating ServiceCatalog portfolio '%s' failed: %s", *input.Id, err.Error())
	}
	return resourceAwsServiceCatalogPortfolioRead(d, meta)
}

func resourceAwsServiceCatalogPortfolioDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.DeletePortfolioInput{}
	input.Id = aws.String(d.Id())

	_, err := conn.DeletePortfolio(&input)
	if err != nil {
		return fmt.Errorf("Deleting ServiceCatalog portfolio '%s' failed: %s", *input.Id, err.Error())
	}
	return nil
}
