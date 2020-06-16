package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
    "strings"
	"time"
)

func resourceAwsServiceCatalogPortfolioPrincipalAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogPortfolioPrincipalAssociationCreate,
		Read:   resourceAwsServiceCatalogPortfolioPrincipalAssociationRead,
		Update: resourceAwsServiceCatalogPortfolioPrincipalAssociationUpdate,
		Delete: resourceAwsServiceCatalogPortfolioPrincipalAssociationDelete,
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
			"principal_arn": {
				Type: schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsServiceCatalogPortfolioPrincipalAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	_, portfolioId, principalArn := resourceAwsServiceCatalogPortfolioPrincipalAssociationRequiredParameters(d)
	input := servicecatalog.AssociatePrincipalWithPortfolioInput{
		PortfolioId: aws.String(portfolioId),
		PrincipalARN: aws.String(principalArn),
		PrincipalType: aws.String("IAM"),
	}
	conn := meta.(*AWSClient).scconn
	_, err := conn.AssociatePrincipalWithPortfolio(&input)
	if err != nil {
		return fmt.Errorf("creating Service Catalog Principal(%s)/Portfolio(%s) Association failed: %s",
			principalArn, portfolioId, err.Error())
	}
	
    stateConf := resource.StateChangeConf{
        Pending:      []string{servicecatalog.StatusCreating},
        Target:       []string{servicecatalog.StatusAvailable},
        Timeout:      1 * time.Minute,
        PollInterval: 3 * time.Second,
        Refresh: func() (interface{}, string, error) {
            err := resourceAwsServiceCatalogPortfolioPrincipalAssociationRead(d, meta)
            if err != nil {
                return 42, "", err
            }
            if d.Id() != "" {
                return 42, servicecatalog.StatusAvailable, err
            }
            return 0, servicecatalog.StatusCreating, err
        },
    }
    _, err = stateConf.WaitForState()
    return err
}

func resourceAwsServiceCatalogPortfolioPrincipalAssociationRead(d *schema.ResourceData, meta interface{}) error {
	id, portfolioId, principalArn := resourceAwsServiceCatalogPortfolioPrincipalAssociationRequiredParameters(d)
	input := servicecatalog.ListPrincipalsForPortfolioInput{
		PortfolioId: aws.String(portfolioId),
	}
	conn := meta.(*AWSClient).scconn
	var principals []*servicecatalog.Principal
	var pageToken = ""
	for {
		pageOfDetails, nextPageToken, err := resourceAwsServiceCatalogPortfolioPrincipalAssociationListPrincipalsForPortfolioPage(conn, input, &pageToken)
		if err != nil {
			return err
		}
		principals = append(pageOfDetails)
		if nextPageToken == nil {
			break
		}
		pageToken = *nextPageToken
	}
	isFound := false
	for _, principal := range principals {
		if *principal.PrincipalARN == principalArn {
			isFound = true
			d.SetId(id)
			break
		}
	}
	if !isFound {
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

func resourceAwsServiceCatalogPortfolioPrincipalAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	const principalArnKey = "principal_arn"
	const portfolioIdKey = "portfolio_id"
	if d.HasChange(principalArnKey) || d.HasChange(portfolioIdKey) {
		oldPrincipalArn, newPrincipalArn := d.GetChange(principalArnKey)
		oldPortfolioId, newPortfolioId := d.GetChange(portfolioIdKey)
		d.Set(principalArnKey, oldPrincipalArn)
		d.Set(portfolioIdKey, oldPortfolioId)
		resourceAwsServiceCatalogPortfolioPrincipalAssociationDelete(d, meta)
		d.Set(principalArnKey, newPrincipalArn)
		d.Set(portfolioIdKey, newPortfolioId)
		resourceAwsServiceCatalogPortfolioPrincipalAssociationCreate(d, meta)
	}
	return resourceAwsServiceCatalogPortfolioPrincipalAssociationRead(d, meta)
}

func resourceAwsServiceCatalogPortfolioPrincipalAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	_, portfolioId, principalArn := resourceAwsServiceCatalogPortfolioPrincipalAssociationRequiredParameters(d)
	input := servicecatalog.DisassociatePrincipalFromPortfolioInput{
		PortfolioId: aws.String(portfolioId),
		PrincipalARN: aws.String(principalArn),
	}
	conn := meta.(*AWSClient).scconn
	_, err := conn.DisassociatePrincipalFromPortfolio(&input)
	if err != nil {
		return fmt.Errorf("deleting Service Catalog Principal(%s)/Portfolio(%s) Association failed: %s",
			principalArn, portfolioId, err.Error())
	}
	return nil
}

func resourceAwsServiceCatalogPortfolioPrincipalAssociationRequiredParameters(d *schema.ResourceData) (string, string, string) {
    if principalArn, ok := d.GetOk("principal_arn"); ok {
	    portfolioId := d.Get("portfolio_id").(string)
	    id := portfolioId + "--" + principalArn.(string);
	    return id, portfolioId, principalArn.(string)
    }
    return parseServiceCatalogPortfolioPrincipalAssociationResourceId(d.Id())
}

func parseServiceCatalogPortfolioPrincipalAssociationResourceId(id string) (string, string, string) {
    s := strings.SplitN(id, "--", 2)
    portfolioId := s[0]
    principalArn := s[1]
    return id, portfolioId, principalArn
}
