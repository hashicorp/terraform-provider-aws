package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfservicecatalog "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func dataSourceAwsServiceCatalogConstraint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsServiceCatalogConstraintRead,

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "en",
				ValidateFunc: validation.StringInSlice(tfservicecatalog.AcceptLanguage_Values(), false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parameters": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsServiceCatalogConstraintRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	output, err := waiter.ConstraintReady(conn, d.Get("accept_language").(string), d.Get("id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Catalog Constraint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Constraint (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting Service Catalog Constraint (%s): empty response", d.Id())
	}

	acceptLanguage := d.Get("accept_language").(string)

	if acceptLanguage == "" {
		acceptLanguage = "en"
	}

	d.Set("accept_language", acceptLanguage)

	d.Set("parameters", output.ConstraintParameters)
	d.Set("status", output.Status)

	detail := output.ConstraintDetail

	d.Set("description", detail.Description)
	d.Set("owner", detail.Owner)
	d.Set("portfolio_id", detail.PortfolioId)
	d.Set("product_id", detail.ProductId)
	d.Set("type", detail.Type)

	d.SetId(aws.StringValue(detail.ConstraintId))

	return nil
}
