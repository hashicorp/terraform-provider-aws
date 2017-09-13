package aws

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsServiceCatalogProductAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogProductAssociationCreate,
		Read:   resourceAwsServiceCatalogProductAssociationRead,
		Delete: resourceAwsServiceCatalogProductAssociationDelete,

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
			"source_portfolio_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsServiceCatalogProductAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.AssociateProductWithPortfolioInput{}
	if v, ok := d.GetOk("portfolio_id"); ok {
		input.PortfolioId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("product_id"); ok {
		input.ProductId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_portfolio_id"); ok {
		input.SourcePortfolioId = aws.String(v.(string))
	}

	id := productAssociationId(input.PortfolioId, input.ProductId, input.SourcePortfolioId)

	_, err := conn.AssociateProductWithPortfolio(&input)
	if err != nil {
		return fmt.Errorf("Adding ServiceCatalog product_association '%s' failed: %s", id, err.Error())
	}
	d.SetId(id)

	return resourceAwsServiceCatalogProductAssociationRead(d, meta)
}

func resourceAwsServiceCatalogProductAssociationRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()
	x := strings.Split(id, "_")
	d.Set("portfolio_id", x[0])
	d.Set("product_id", x[1])
	d.Set("source_portfolio_id", x[2])
	return nil
}

func resourceAwsServiceCatalogProductAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	id := d.Id()
	x := strings.Split(id, "_")
	portfolio_id := x[0]
	product_id := x[1]

	input := servicecatalog.DisassociateProductFromPortfolioInput{
		PortfolioId: aws.String(portfolio_id),
		ProductId:   aws.String(product_id),
	}

	_, err := conn.DisassociateProductFromPortfolio(&input)
	if err != nil {
		return fmt.Errorf("Deleting ServiceCatalog product_association '%s' failed: %s", id, err.Error())
	}
	return nil
}

func productAssociationId(portfolioId, productId, sourcePortfolioId *string) string {
	if sourcePortfolioId == nil {
		return *portfolioId + "_" + *productId + "_"
	}

	return *portfolioId + "_" + *productId + "_" + *sourcePortfolioId
}
