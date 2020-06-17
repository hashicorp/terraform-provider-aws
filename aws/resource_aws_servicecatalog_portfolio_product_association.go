package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"strings"
	"time"
)

func resourceAwsServiceCatalogPortfolioProductAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogPortfolioProductAssociationCreate,
		Read:   resourceAwsServiceCatalogPortfolioProductAssociationRead,
		Update: resourceAwsServiceCatalogPortfolioProductAssociationUpdate,
		Delete: resourceAwsServiceCatalogPortfolioProductAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsServiceCatalogPortfolioProductAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	_, portfolioId, productId := resourceAwsServiceCatalogPortfolioProductAssociationRequiredParameters(d)
	input := servicecatalog.AssociateProductWithPortfolioInput{
		PortfolioId: aws.String(portfolioId),
		ProductId:   aws.String(productId),
	}
	conn := meta.(*AWSClient).scconn
	_, err := conn.AssociateProductWithPortfolio(&input)
	if err != nil {
		return fmt.Errorf("creating Service Catalog Product(%s)/Portfolio(%s) Association failed: %s",
			productId, portfolioId, err.Error())
	}
	return resourceAwsServiceCatalogPortfolioProductAssociationRead(d, meta)
}

func resourceAwsServiceCatalogPortfolioProductAssociationRead(d *schema.ResourceData, meta interface{}) error {
	id, portfolioId, productId := resourceAwsServiceCatalogPortfolioProductAssociationRequiredParameters(d)
	input := servicecatalog.ListPortfoliosForProductInput{
		ProductId: aws.String(productId),
	}
	conn := meta.(*AWSClient).scconn
	var pageToken = ""
	isFound := false
	for {
		pageOfDetails, nextPageToken, err := resourceAwsServiceCatalogPortfolioProductAssociationListPortfoliosForProductPage(conn, input, &pageToken)
		if err != nil {
			return err
		}
		for _, portfolioDetail := range pageOfDetails {
			if *portfolioDetail.Id == portfolioId {
				isFound = true
				d.SetId(id)
				break
			}
		}
		if nextPageToken == nil || isFound {
			break
		}
		pageToken = *nextPageToken
	}
	if !isFound {
		log.Printf("[WARN] Service Catalog Product(%s)/Portfolio(%s) Association not found, removing from state",
			productId, portfolioId)
		d.SetId("")
	}
	d.Set("product_id", productId)
	d.Set("portfolio_id", portfolioId)
	return nil
}

func resourceAwsServiceCatalogPortfolioProductAssociationListPortfoliosForProductPage(conn *servicecatalog.ServiceCatalog, input servicecatalog.ListPortfoliosForProductInput, nextPageToken *string) ([]*servicecatalog.PortfolioDetail, *string, error) {
	input.PageToken = nextPageToken
	var page, err = conn.ListPortfoliosForProduct(&input)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieving Service Catalog Associations for Product/Portfolios: %s", err.Error())
	}
	portfolioDetails := page.PortfolioDetails
	return portfolioDetails, page.NextPageToken, nil
}

func resourceAwsServiceCatalogPortfolioProductAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	const productIdKey = "product_id"
	const portfolioIdKey = "portfolio_id"
	if d.HasChange(productIdKey) || d.HasChange(portfolioIdKey) {
		oldProductId, newProductId := d.GetChange(productIdKey)
		oldPortfolioId, newPortfolioId := d.GetChange(portfolioIdKey)
		d.Set(productIdKey, oldProductId)
		d.Set(portfolioIdKey, oldPortfolioId)
		resourceAwsServiceCatalogPortfolioProductAssociationDelete(d, meta)
		d.Set(productIdKey, newProductId)
		d.Set(portfolioIdKey, newPortfolioId)
		resourceAwsServiceCatalogPortfolioProductAssociationCreate(d, meta)
	}
	return resourceAwsServiceCatalogPortfolioProductAssociationRead(d, meta)
}

func resourceAwsServiceCatalogPortfolioProductAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	_, portfolioId, productId := resourceAwsServiceCatalogPortfolioProductAssociationRequiredParameters(d)
	input := servicecatalog.DisassociateProductFromPortfolioInput{
		PortfolioId: aws.String(portfolioId),
		ProductId:   aws.String(productId),
	}
	conn := meta.(*AWSClient).scconn
	_, err := conn.DisassociateProductFromPortfolio(&input)
	if err != nil {
		return fmt.Errorf("deleting Service Catalog Product(%s)/Portfolio(%s) Association failed: %s",
			productId, portfolioId, err.Error())
	}
	return nil
}

func resourceAwsServiceCatalogPortfolioProductAssociationRequiredParameters(d *schema.ResourceData) (string, string, string) {
	if productId, ok := d.GetOk("product_id"); ok {
		portfolioId := d.Get("portfolio_id").(string)
		id := portfolioId + "--" + productId.(string)
		return id, portfolioId, productId.(string)
	}
	return parseServiceCatalogPortfolioProductAssociationResourceId(d.Id())
}

func parseServiceCatalogPortfolioProductAssociationResourceId(id string) (string, string, string) {
	s := strings.SplitN(id, "--", 2)
	portfolioId := s[0]
	productId := s[1]
	return id, portfolioId, productId
}
