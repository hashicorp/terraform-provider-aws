package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"time"
)

func resourceAwsServiceCatalogPortfolioProductAssociation() *schema.Resource {
	return &schema.Resource{
		Create: createResource,
		Read: readResource,
		Update: updateResource,
		Delete: deleteResource,
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
				Type: schema.TypeString,
				Required: true,
			},
			"product_id": {
				Type: schema.TypeString,
				Required: true,
			},
		},
	}
}

func createResource(d *schema.ResourceData, meta interface{}) error {
	productId, portfolioId := requiredParameters(d)
	input := servicecatalog.AssociateProductWithPortfolioInput{
		PortfolioId: aws.String(portfolioId.(string)),
		ProductId: aws.String(productId.(string)),
	}
	conn := meta.(*AWSClient).scconn
	_, err := conn.AssociateProductWithPortfolio(&input)
	if err != nil {
		return fmt.Errorf("creating Service Catalog Product(%s)/Portfolio(%s) Association failed: %s",
			productId, portfolioId, err.Error())
	}
	return readResource(d, meta)
}

func readResource(d *schema.ResourceData, meta interface{}) error {
	productId, portfolioId := requiredParameters(d)
	input := servicecatalog.ListPortfoliosForProductInput{
		ProductId: aws.String(productId.(string)),
	}
	conn := meta.(*AWSClient).scconn

	//TODO - move to a function to allow recursive or repeated calls
	// Fetch a Page
	//TODO - fixed - page size
	//TODO - optional - page token
	var products, err = conn.ListPortfoliosForProduct(&input)
	if err != nil {
		return fmt.Errorf("retrieving Service Catalog Associations for Product/Portfolios: %s", err.Error())
	}
	//TODO - nextPageToken := products.NextPageToken
	portfolioDetails := products.PortfolioDetails
	//TODO - fetch additional pages

	isFound := false
	for _, portfolioDetail := range portfolioDetails {
		if portfolioDetail.Id == portfolioId {
			isFound = true
			d.SetId(*portfolioDetail.Id)//TOFO pordict id + portfolio id
		}
	}
	if !isFound {
		log.Printf("[WARN] Service Catalog Product(%s)/Portfolio(%s Association not found, removing from state",
			productId, portfolioId)
		d.SetId("")
	}
	return nil
}

func updateResource(d *schema.ResourceData, meta interface{}) error {
	const productIdKey = "product_id"
	const portfolioIdKey = "portfolio_id"
	if d.HasChange(productIdKey) || d.HasChange(portfolioIdKey) {
		oldProductId, newProductId := d.GetChange(productIdKey)
		oldPortfolioId, newPortfolioId := d.GetChange(portfolioIdKey)
		d.Set(productIdKey, oldProductId)
		d.Set(portfolioIdKey, oldPortfolioId)
		deleteResource(d, meta)
		d.Set(productIdKey, newProductId)
		d.Set(portfolioIdKey, newPortfolioId)
		createResource(d, meta)
	}
	return readResource(d, meta)
}

func deleteResource(d *schema.ResourceData, meta interface{}) error {
	productId, portfolioId := requiredParameters(d)
	input := servicecatalog.DisassociateProductFromPortfolioInput{
		PortfolioId: aws.String(portfolioId.(string)),
		ProductId: aws.String(productId.(string)),
	}
	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}
	conn := meta.(*AWSClient).scconn
	_, err := conn.DisassociateProductFromPortfolio(&input)
	if err != nil {
		return fmt.Errorf("deleting Service Catalog Product(%s)/Portfolio(%s) Association failed: %s",
			productId, portfolioId, err.Error())
	}
	return nil
}

func requiredParameters(d *schema.ResourceData) (interface{}, interface{}) {
	productId := d.Get("product_id")
	portfolioId := d.Get("portfolio_id")
	return productId, portfolioId
}
