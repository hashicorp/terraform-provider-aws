package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsServiceCatalogPortfolioPrincipalAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogPortfolioPrincipalAssociationCreate,
		Read:   resourceAwsServiceCatalogPortfolioPrincipalAssociationRead,
		Delete: resourceAwsServiceCatalogPortfolioPrincipalAssociationDelete,
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
			"principal_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
				ForceNew:     true,
			},
		},
	}
}

func resourceAwsServiceCatalogPortfolioPrincipalAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	_, portfolioId, principalArn, err := resourceAwsServiceCatalogPortfolioPrincipalAssociationRequiredParameters(d)
	if err != nil {
		return err
	}
	input := servicecatalog.AssociatePrincipalWithPortfolioInput{
		PortfolioId:   aws.String(portfolioId),
		PrincipalARN:  aws.String(principalArn),
		PrincipalType: aws.String(servicecatalog.PrincipalTypeIam),
	}
	conn := meta.(*AWSClient).scconn
	_, err = conn.AssociatePrincipalWithPortfolio(&input)
	if err != nil {
		return fmt.Errorf("creating Service Catalog Principal(%s)/Portfolio(%s) Association failed: %s",
			principalArn, portfolioId, err.Error())
	}

	result := resourceAwsServiceCatalogPortfolioPrincipalAssociationRead(d, meta)
	// even after one successful read, the eventual consistency can regress, so delay a bit more before
	// reporting this as created to prevent dependencies (eg products being provisioned) running too early
	time.Sleep(time.Second * 5)
	return result
}

func resourceAwsServiceCatalogPortfolioPrincipalAssociationRead(d *schema.ResourceData, meta interface{}) error {
	id, portfolioId, principalArn, err := resourceAwsServiceCatalogPortfolioPrincipalAssociationRequiredParameters(d)
	if err != nil {
		return err
	}
	input := servicecatalog.ListPrincipalsForPortfolioInput{
		PortfolioId: aws.String(portfolioId),
	}
	conn := meta.(*AWSClient).scconn
	isFound := false

	// listing principals for portfolio is a paginated operation
	// and if a principal has recently been added, it can contain the ID while it is stabilising,
	// so we retry for up to 1 minute if it is stabilising and the ARN we are looking for is not found
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		var pageToken = ""
		for {
			nonArnFound := false
			pageOfDetails, nextPageToken, err := resourceAwsServiceCatalogPortfolioPrincipalAssociationListPrincipalsForPortfolioPage(conn, input, &pageToken)
			if err != nil {
				return resource.NonRetryableError(err)
			}
			for _, principal := range pageOfDetails {
				if aws.StringValue(principal.PrincipalARN) == principalArn {
					isFound = true
					return nil
				}
				if !strings.HasPrefix(aws.StringValue(principal.PrincipalARN), "arn:") {
					nonArnFound = true
				}
			}
			if nextPageToken == nil {
				if nonArnFound {
					log.Printf("[DEBUG] Service Catalog Principal(%s)/Portfolio(%s) Association not found, but principals detected as stabilizing",
						principalArn, portfolioId)
					return resource.RetryableError(fmt.Errorf("Principals stabilizing"))
				} else {
					return nil
				}
			}
			pageToken = aws.StringValue(nextPageToken)
		}
	})
	if err != nil {
		return err
	}
	if isFound {
		d.SetId(id)
	} else {
		log.Printf("[WARN] Service Catalog Principal(%s)/Portfolio(%s) Association not found, removing from state",
			principalArn, portfolioId)
		d.SetId("")
	}
	d.Set("principal_arn", principalArn)
	d.Set("portfolio_id", portfolioId)
	return nil
}

func resourceAwsServiceCatalogPortfolioPrincipalAssociationListPrincipalsForPortfolioPage(conn *servicecatalog.ServiceCatalog, input servicecatalog.ListPrincipalsForPortfolioInput, nextPageToken *string) ([]*servicecatalog.Principal, *string, error) {
	input.PageToken = nextPageToken
	var page, err = conn.ListPrincipalsForPortfolio(&input)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieving Service Catalog Associations for Principal/Portfolios: %s", err.Error())
	}
	principalDetails := page.Principals
	return principalDetails, page.NextPageToken, nil
}

func resourceAwsServiceCatalogPortfolioPrincipalAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	_, portfolioId, principalArn, err := resourceAwsServiceCatalogPortfolioPrincipalAssociationRequiredParameters(d)
	if err != nil {
		return err
	}
	input := servicecatalog.DisassociatePrincipalFromPortfolioInput{
		PortfolioId:  aws.String(portfolioId),
		PrincipalARN: aws.String(principalArn),
	}
	conn := meta.(*AWSClient).scconn
	_, err = conn.DisassociatePrincipalFromPortfolio(&input)
	if err != nil {
		return fmt.Errorf("deleting Service Catalog Principal(%s)/Portfolio(%s) Association failed: %s",
			principalArn, portfolioId, err.Error())
	}
	return nil
}

func resourceAwsServiceCatalogPortfolioPrincipalAssociationRequiredParameters(d *schema.ResourceData) (string, string, string, error) {
	// ":" recommended as separator where multiple fields needed to uniquely identify and import, based on https://www.terraform.io/docs/extend/resources/import.html#importer-state-function
	// (as in this case where AWS doesn't treat this association as a first class resource; it has no AWS identifier)
	// this is not a valid "identifier" character according to https://www.terraform.io/docs/configuration/syntax.html#identifiers
	// but that does not seem to apply to this internal "id"
	principalArn, ok := d.GetOk("principal_arn")
	portfolioId, ok2 := d.GetOk("portfolio_id")
	if ok && ok2 {
		id := portfolioId.(string) + ":" + principalArn.(string)
		return id, portfolioId.(string), principalArn.(string), nil
	} else if ok || ok2 {
		return "", "", "", fmt.Errorf("Invalid state - principal_arn and portfolio_id must both be set or neither set to infer from ID")
	} else if d.Id() != "" {
		return parseServiceCatalogPortfolioPrincipalAssociationResourceId(d.Id())
	} else {
		return "", "", "", fmt.Errorf("Invalid state - principal_arn and portfolio_id must be set, or ID set to import")
	}
}

func parseServiceCatalogPortfolioPrincipalAssociationResourceId(id string) (string, string, string, error) {
	s := strings.SplitN(id, ":", 2)
	if len(s) != 2 {
		return "", "", "", fmt.Errorf("Invalid ID '%s' - should be of format <portfolio_id>:<principal-arn>", id)
	}
	return id, s[0], s[1], nil
}
