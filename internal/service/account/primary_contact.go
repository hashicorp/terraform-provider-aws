package account

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/account"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_account_primary_contact")
func ResourcePrimaryContact() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePrimaryContactCreate,
		ReadWithoutTimeout:   resourcePrimaryContactRead,
		UpdateWithoutTimeout: resourcePrimaryContactUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"address_line_1": {
				Type:     schema.TypeString,
				Required: true,
			},
			"address_line_2": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"address_line_3": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"city": {
				Type:     schema.TypeString,
				Required: true,
			},
			"company_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"country_code": {
				Type:     schema.TypeString,
				Required: true,
			},
			"district_or_county": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"full_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"phone_number": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\s0-9()+-]+$`), "must be a valid phone number"),
			},
			"postal_code": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state_or_region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"website_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourcePrimaryContactCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountConn()

	input := putContactInformationFromSchema(d)

	log.Printf("[DEBUG] Creating Account Primary Contact: %s", input)
	_, err := conn.PutContactInformationWithContext(ctx, input)

	id := d.Get("account_id").(string)
	if id == "" {
		id = "default"
	}

	if err != nil {
		return diag.Errorf("error creating Account Primary Contact (%s): %s", id, err)
	}

	d.SetId(id)

	return resourcePrimaryContactRead(ctx, d, meta)
}

func resourcePrimaryContactRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountConn()

	input := &account.GetContactInformationInput{}

	accountID := d.Get("account_id").(string)
	if accountID != "" {
		input.AccountId = aws.String(accountID)
	}

	output, err := conn.GetContactInformationWithContext(ctx, input)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Account Primary Contact (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Account Primary Contact (%s): %s", d.Id(), err)
	}

	d.Set("account_id", accountID)
	d.Set("address_line_1", output.ContactInformation.AddressLine1)
	d.Set("address_line_2", output.ContactInformation.AddressLine2)
	d.Set("address_line_3", output.ContactInformation.AddressLine3)
	d.Set("city", output.ContactInformation.City)
	d.Set("company_name", output.ContactInformation.CompanyName)
	d.Set("country_code", output.ContactInformation.CountryCode)
	d.Set("district_or_county", output.ContactInformation.DistrictOrCounty)
	d.Set("full_name", output.ContactInformation.FullName)
	d.Set("phone_number", output.ContactInformation.PhoneNumber)
	d.Set("postal_code", output.ContactInformation.PostalCode)
	d.Set("state_or_region", output.ContactInformation.StateOrRegion)
	d.Set("website_url", output.ContactInformation.WebsiteUrl)

	return nil
}

func resourcePrimaryContactUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountConn()

	input := putContactInformationFromSchema(d)

	log.Printf("[DEBUG] Updating Account Primary Contact: %s", input)
	_, err := conn.PutContactInformationWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error updating Account Primary Contact (%s): %s", d.Id(), err)
	}

	return resourcePrimaryContactRead(ctx, d, meta)
}

func putContactInformationFromSchema(d *schema.ResourceData) *account.PutContactInformationInput {
	input := &account.PutContactInformationInput{}

	accountID := d.Get("account_id").(string)
	if accountID != "" {
		input.AccountId = aws.String(accountID)
	}

	input.ContactInformation = &account.ContactInformation{
		AddressLine1: aws.String(d.Get("address_line_1").(string)),
		City:         aws.String(d.Get("city").(string)),
		CountryCode:  aws.String(d.Get("country_code").(string)),
		FullName:     aws.String(d.Get("full_name").(string)),
		PhoneNumber:  aws.String(d.Get("phone_number").(string)),
		PostalCode:   aws.String(d.Get("postal_code").(string)),
	}

	addressLine2 := d.Get("address_line_2").(string)
	if addressLine2 != "" {
		input.ContactInformation.AddressLine2 = aws.String(addressLine2)
	}
	addressLine3 := d.Get("address_line_3").(string)
	if addressLine2 != "" {
		input.ContactInformation.AddressLine3 = aws.String(addressLine3)
	}
	companyName := d.Get("company_name").(string)
	if companyName != "" {
		input.ContactInformation.CompanyName = aws.String(companyName)
	}
	districtOrCounty := d.Get("district_or_county").(string)
	if districtOrCounty != "" {
		input.ContactInformation.DistrictOrCounty = aws.String(districtOrCounty)
	}
	stateOrRegion := d.Get("state_or_region").(string)
	if stateOrRegion != "" {
		input.ContactInformation.StateOrRegion = aws.String(stateOrRegion)
	}
	websiteUrl := d.Get("website_url").(string)
	if websiteUrl != "" {
		input.ContactInformation.WebsiteUrl = aws.String(websiteUrl)
	}

	return input
}
