package aws

import (
	"fmt"
	"strings"
	"time"

	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsServiceCatalogPrincipalAssociationPortfolio() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogPrincipalAssociationPortfolioCreate,
		Read:   resourceAwsServiceCatalogPrincipalAssociationPortfolioRead,
		Delete: resourceAwsServiceCatalogPrincipalAssociationPortfolioDelete,

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
			"principal_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsServiceCatalogPrincipalAssociationPortfolioCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.AssociatePrincipalWithPortfolioInput{PrincipalType: aws.String("IAM")}
	if v, ok := d.GetOk("portfolio_id"); ok {
		input.PortfolioId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("principal_arn"); ok {
		input.PrincipalARN = aws.String(v.(string))
	}

	id := principalAssociationPortfolio(input.PortfolioId, input.PrincipalType, input.PrincipalARN)

	_, err := conn.AssociatePrincipalWithPortfolio(&input)
	if err != nil {
		return fmt.Errorf("Adding ServiceCatalog principal_association_portfolio '%s' failed: %s", id, err.Error())
	}
	d.SetId(id)

	return resourceAwsServiceCatalogPrincipalAssociationPortfolioRead(d, meta)
}

func resourceAwsServiceCatalogPrincipalAssociationPortfolioRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()
	x := strings.Split(id, "_")
	d.Set("portfolio_id", x[0])
	d.Set("principal_type", x[1])
	arn, err := base64.URLEncoding.DecodeString(x[2])
	if err != nil {
		return fmt.Errorf("Reading ServiceCatalog principal_association_portfolio '%s' failed: %s", id, err.Error())
	}
	d.Set("principal_arn", string(arn))
	return nil
}

func resourceAwsServiceCatalogPrincipalAssociationPortfolioDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	id := d.Id()
	x := strings.Split(id, "_")
	portfolio_id := x[0]
	arn, err := base64.URLEncoding.DecodeString(x[2])
	if err != nil {
		return fmt.Errorf("Reading ServiceCatalog principal_association_portfolio '%s' failed: %s", id, err.Error())
	}
	principal_arn := string(arn)

	input := servicecatalog.DisassociatePrincipalFromPortfolioInput{
		PortfolioId:  aws.String(portfolio_id),
		PrincipalARN: aws.String(principal_arn),
	}

	_, err = conn.DisassociatePrincipalFromPortfolio(&input)
	if err != nil {
		return fmt.Errorf("Deleting ServiceCatalog principal_association_portfolio '%s' failed: %s", id, err.Error())
	}
	return nil
}

func principalAssociationPortfolio(portfolioId, principalType, principalARN *string) string {
	encodedARN := base64.URLEncoding.EncodeToString([]byte(*principalARN))
	return *portfolioId + "_" + *principalType + "_" + encodedARN
}
