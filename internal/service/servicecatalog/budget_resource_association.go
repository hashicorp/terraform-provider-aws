package servicecatalog

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
)

func ResourceBudgetResourceAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceBudgetResourceAssociationCreate,
		Read:   resourceBudgetResourceAssociationRead,
		Delete: resourceBudgetResourceAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"budget_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceBudgetResourceAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceCatalogConn

	input := &servicecatalog.AssociateBudgetWithResourceInput{
		BudgetName: aws.String(d.Get("budget_name").(string)),
		ResourceId: aws.String(d.Get("resource_id").(string)),
	}

	var output *servicecatalog.AssociateBudgetWithResourceOutput
	err := resource.Retry(tfiam.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.AssociateBudgetWithResource(input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.AssociateBudgetWithResource(input)
	}

	if err != nil {
		return fmt.Errorf("error associating Service Catalog Budget with Resource: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Service Catalog Budget Resource Association: empty response")
	}

	d.SetId(BudgetResourceAssociationID(d.Get("budget_name").(string), d.Get("resource_id").(string)))

	return resourceBudgetResourceAssociationRead(d, meta)
}

func resourceBudgetResourceAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceCatalogConn

	budgetName, resourceID, err := BudgetResourceAssociationParseID(d.Id())

	if err != nil {
		return fmt.Errorf("could not parse ID (%s): %w", d.Id(), err)
	}

	output, err := WaitBudgetResourceAssociationReady(conn, budgetName, resourceID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Catalog Budget Resource Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Budget Resource Association (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting Service Catalog Budget Resource Association (%s): empty response", d.Id())
	}

	d.Set("resource_id", resourceID)
	d.Set("budget_name", output.BudgetName)

	return nil
}

func resourceBudgetResourceAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceCatalogConn

	budgetName, resourceID, err := BudgetResourceAssociationParseID(d.Id())

	if err != nil {
		return fmt.Errorf("could not parse ID (%s): %w", d.Id(), err)
	}

	input := &servicecatalog.DisassociateBudgetFromResourceInput{
		ResourceId: aws.String(resourceID),
		BudgetName: aws.String(budgetName),
	}

	_, err = conn.DisassociateBudgetFromResource(input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating Service Catalog Budget from Resource (%s): %w", d.Id(), err)
	}

	err = WaitBudgetResourceAssociationDeleted(conn, budgetName, resourceID)

	if err != nil && !tfresource.NotFound(err) {
		return fmt.Errorf("error waiting for Service Catalog Budget Resource Disassociation (%s): %w", d.Id(), err)
	}

	return nil
}
