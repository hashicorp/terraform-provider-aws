package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	tfservicecatalog "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsServiceCatalogPortfolioShare() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogPortfolioShareCreate,
		Read:   resourceAwsServiceCatalogPortfolioShareRead,
		Update: resourceAwsServiceCatalogPortfolioShareUpdate,
		Delete: resourceAwsServiceCatalogPortfolioShareDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "en",
				ValidateFunc: validation.StringInSlice(tfservicecatalog.AcceptLanguage_Values(), false),
			},
			"accepted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateAwsAccountId,
				ExactlyOneOf: []string{
					"organization_node",
					"account_id",
				},
			},
			"organization_node": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ExactlyOneOf: []string{
					"organization_node",
					"account_id",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(servicecatalog.OrganizationNodeType_Values(), false),
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"share_tag_options": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(servicecatalog.DescribePortfolioShareType_Values(), false),
			},
		},
	}
}

func resourceAwsServiceCatalogPortfolioShareCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := &servicecatalog.CreatePortfolioShareInput{
		PortfolioId: aws.String(d.Get("portfolio_id").(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("account_id"); ok {
		input.AccountId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("organization_node"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.OrganizationNode = expandServiceCatalogOrganizationNode(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("share_tag_options"); ok {
		input.ShareTagOptions = aws.Bool(v.(bool))
	}

	var output *servicecatalog.CreatePortfolioShareOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.CreatePortfolioShare(input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreatePortfolioShare(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Service Catalog Portfolio Share: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Service Catalog Portfolio Share: empty response")
	}

	d.SetId(serviceCatalogPortfolioShareID(
		d.Get("portfolio_id").(string),
		d.Get("account_id").(string),
		input.OrganizationNode,
	))

	// only get a token if organization node, otherwise check without token
	if output.PortfolioShareToken != nil {
		if _, err := waiter.PortfolioShareCreatedWithToken(conn, aws.StringValue(output.PortfolioShareToken)); err != nil {
			return fmt.Errorf("error waiting for Service Catalog Portfolio Share (%s) to be ready: %w", d.Id(), err)
		}
	} else {
		orgNodeValue := ""

		if input.OrganizationNode != nil && input.OrganizationNode.Value != nil {
			orgNodeValue = aws.StringValue(input.OrganizationNode.Value)
		}

		if _, err := waiter.PortfolioShareReady(conn, d.Get("portfolio_id").(string), d.Get("type").(string), d.Get("account_id").(string), orgNodeValue); err != nil {
			return fmt.Errorf("error waiting for Service Catalog Portfolio Share (%s) to be ready: %w", d.Id(), err)
		}
	}

	return resourceAwsServiceCatalogPortfolioShareRead(d, meta)
}

func resourceAwsServiceCatalogPortfolioShareRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	orgNodeValue := ""

	if v, ok := d.GetOk("organization_node"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		orgNode := expandServiceCatalogOrganizationNode(v.([]interface{})[0].(map[string]interface{}))

		if orgNode.Value != nil {
			orgNodeValue = aws.StringValue(orgNode.Value)
		}
	}

	output, err := waiter.PortfolioShareReady(conn, d.Get("portfolio_id").(string), d.Get("type").(string), d.Get("account_id").(string), orgNodeValue)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Service Catalog Portfolio Share (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Portfolio Share (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting Service Catalog Portfolio Share (%s): empty response", d.Id())
	}

	d.Set("accepted", output.Accepted)
	d.Set("share_tag_options", output.ShareTagOptions)
	d.Set("type", output.Type)

	switch aws.StringValue(output.Type) {
	case servicecatalog.DescribePortfolioShareTypeAccount:
		d.Set("account_id", output.PrincipalId)
		d.Set("organization_node", nil)
	default:
		d.Set("account_id", nil)

		orgNode := &servicecatalog.OrganizationNode{
			Type:  output.Type,
			Value: output.PrincipalId,
		}

		if err := d.Set("organization_node", flattenServiceCatalogOrganizationNode(orgNode)); err != nil {
			return fmt.Errorf("error setting organization_node: %w", err)
		}
	}

	return nil
}

func resourceAwsServiceCatalogPortfolioShareUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	if d.HasChanges("accept_language", "account_id", "organization_node", "share_tag_options") {
		input := &servicecatalog.UpdatePortfolioShareInput{
			PortfolioId: aws.String(d.Get("portfolio_id").(string)),
		}

		if v, ok := d.GetOk("accept_language"); ok {
			input.AcceptLanguage = aws.String(v.(string))
		}

		if v, ok := d.GetOk("account_id"); ok {
			input.AccountId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("organization_node"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.OrganizationNode = expandServiceCatalogOrganizationNode(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("share_tag_options"); ok {
			input.ShareTagOptions = aws.Bool(v.(bool))
		}

		err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
			_, err := conn.UpdatePortfolioShare(input)

			if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdatePortfolioShare(input)
		}

		if err != nil {
			return fmt.Errorf("error updating Service Catalog Portfolio Share (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsServiceCatalogPortfolioShareRead(d, meta)
}

func resourceAwsServiceCatalogPortfolioShareDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := &servicecatalog.DeletePortfolioShareInput{
		PortfolioId: aws.String(d.Get("portfolio_id").(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("account_id"); ok {
		input.AccountId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("organization_node"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.OrganizationNode = expandServiceCatalogOrganizationNode(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.DeletePortfolioShare(input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Service Catalog Portfolio Share (%s): %w", d.Id(), err)
	}

	// only get a token if organization node, otherwise check without token
	if output.PortfolioShareToken != nil {
		if _, err := waiter.PortfolioShareDeletedWithToken(conn, aws.StringValue(output.PortfolioShareToken)); err != nil {
			return fmt.Errorf("error waiting for Service Catalog Portfolio Share (%s) to be deleted: %w", d.Id(), err)
		}
	} else {
		orgNodeValue := ""

		if input.OrganizationNode != nil && input.OrganizationNode.Value != nil {
			orgNodeValue = aws.StringValue(input.OrganizationNode.Value)
		}

		if _, err := waiter.PortfolioShareDeleted(conn, d.Get("portfolio_id").(string), d.Get("type").(string), d.Get("account_id").(string), orgNodeValue); err != nil {
			return fmt.Errorf("error waiting for Service Catalog Portfolio Share (%s) to be deleted: %w", d.Id(), err)
		}
	}

	return nil
}

func expandServiceCatalogOrganizationNode(tfMap map[string]interface{}) *servicecatalog.OrganizationNode {
	if tfMap == nil {
		return nil
	}

	apiObject := &servicecatalog.OrganizationNode{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func flattenServiceCatalogOrganizationNode(apiObject *servicecatalog.OrganizationNode) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap["value"] = aws.StringValue(v)
	}

	return tfMap
}

func serviceCatalogPortfolioShareID(portfolioID, accountID string, orgNode *servicecatalog.OrganizationNode) string {
	var pieces []string

	if portfolioID != "" {
		pieces = append(pieces, portfolioID)
	}

	if accountID != "" {
		pieces = append(pieces, accountID)
	}

	if orgNode != nil {
		if orgNode.Type != nil {
			pieces = append(pieces, aws.StringValue(orgNode.Type))
		}

		if orgNode.Value != nil {
			pieces = append(pieces, aws.StringValue(orgNode.Value))
		}
	}

	return strings.Join(pieces, ":")
}
