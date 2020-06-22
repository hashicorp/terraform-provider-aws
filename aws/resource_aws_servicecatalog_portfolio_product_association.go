package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsServiceCatalogPortfolioProductAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogPortfolioProductAssociationCreate,
		Read:   resourceAwsServiceCatalogPortfolioProductAssociationRead,
		Delete: resourceAwsServiceCatalogPortfolioProductAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsServiceCatalogPortfolioProductAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	_, portfolioId, productId, err := resourceAwsServiceCatalogPortfolioProductAssociationRequiredParameters(d)
	if err != nil {
		return err
	}
	input := servicecatalog.AssociateProductWithPortfolioInput{
		PortfolioId: aws.String(portfolioId),
		ProductId:   aws.String(productId),
	}
	conn := meta.(*AWSClient).scconn
	_, err = conn.AssociateProductWithPortfolio(&input)
	if err != nil {
		return fmt.Errorf("creating Service Catalog Product(%s)/Portfolio(%s) Association failed: %s",
			productId, portfolioId, err.Error())
	}
	result := resourceAwsServiceCatalogPortfolioProductAssociationRead(d, meta)
	// even after one successful read, the eventual consistency can regress, so delay a bit more before
	// reporting this as created to prevent dependencies (eg products being provisioned) running too early
	time.Sleep(time.Second * 5)
	return result
}

func resourceAwsServiceCatalogPortfolioProductAssociationRead(d *schema.ResourceData, meta interface{}) error {
	id, portfolioId, productId, err := resourceAwsServiceCatalogPortfolioProductAssociationRequiredParameters(d)
	if err != nil {
		return err
	}
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
			if aws.StringValue(portfolioDetail.Id) == portfolioId {
				isFound = true
				d.SetId(id)
				break
			}
		}
		if nextPageToken == nil || isFound {
			break
		}
		pageToken = aws.StringValue(nextPageToken)
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

func resourceAwsServiceCatalogPortfolioProductAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	_, portfolioId, productId, err := resourceAwsServiceCatalogPortfolioProductAssociationRequiredParameters(d)
	if err != nil {
		return err
	}
	input := servicecatalog.DisassociateProductFromPortfolioInput{
		PortfolioId: aws.String(portfolioId),
		ProductId:   aws.String(productId),
	}
	conn := meta.(*AWSClient).scconn
	_, err = conn.DisassociateProductFromPortfolio(&input)
	if err != nil {
		return fmt.Errorf("deleting Service Catalog Product(%s)/Portfolio(%s) Association failed: %s",
			productId, portfolioId, err.Error())
	}
	return nil
}

func resourceAwsServiceCatalogPortfolioProductAssociationRequiredParameters(d *schema.ResourceData) (string, string, string, error) {
	// ":" recommended as separator where multiple fields needed to uniquely identify and import, based on https://www.terraform.io/docs/extend/resources/import.html#importer-state-function
	// (as in this case where AWS doesn't treat this association as a first class resource; it has no AWS identifier)
	// this is not a valid "identifier" character according to https://www.terraform.io/docs/configuration/syntax.html#identifiers
	// but that does not seem to apply to this internal "id"
	productId, ok := d.GetOk("product_id")
	portfolioId, ok2 := d.GetOk("portfolio_id")
	if ok && ok2 {
		id := portfolioId.(string) + ":" + productId.(string)
		return id, portfolioId.(string), productId.(string), nil
	} else if ok || ok2 {
		return "", "", "", fmt.Errorf("Invalid state - product_id and portfolio_id must both be set or neither set to infer from ID")
	} else if d.Id() != "" {
		return parseServiceCatalogPortfolioProductAssociationResourceId(d.Id())
	} else {
		return "", "", "", fmt.Errorf("Invalid state - product_id and portfolio_id must be set, or ID set to import")
	}
}

func parseServiceCatalogPortfolioProductAssociationResourceId(id string) (string, string, string, error) {
	s := strings.SplitN(id, ":", 2)
	if len(s) != 2 {
		return "", "", "", fmt.Errorf("Invalid ID '%s' - should be of format <portfolio_id>:<product-id>", id)
	}
	return id, s[0], s[1], nil
}
