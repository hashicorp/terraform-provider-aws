package account

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/account"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAlternateContact() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAlternateContactCreate,
		ReadContext:   resourceAlternateContactRead,
		UpdateContext: resourceAlternateContactUpdate,
		DeleteContext: resourceAlternateContactDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"alternate_contact_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(account.AlternateContactType_Values(), false),
			},
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"email_address": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[\w+=,.-]+@[\w.-]+\.[\w]+`), "must be a valid email address"),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"phone_number": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\s0-9()+-]+$`), "must be a valid phone number"),
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
		},
	}
}

func resourceAlternateContactCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountConn

	contactType := d.Get("alternate_contact_type").(string)
	input := &account.PutAlternateContactInput{
		AlternateContactType: aws.String(contactType),
		EmailAddress:         aws.String(d.Get("email_address").(string)),
		Name:                 aws.String(d.Get("name").(string)),
		PhoneNumber:          aws.String(d.Get("phone_number").(string)),
		Title:                aws.String(d.Get("title").(string)),
	}

	id := contactType
	if v, ok := d.GetOk("account_id"); ok {
		account_id := v.(string)
		input.AccountId = aws.String(account_id)
		id = account_id + "/" + contactType
	}

	log.Printf("[DEBUG] Creating Account Alternate Contact: %s", input)
	_, err := conn.PutAlternateContactWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating Account Alternate Contact (%s): %s", id, err)
	}

	d.SetId(id)

	return resourceAlternateContactRead(ctx, d, meta)
}

func resourceAlternateContactRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountConn

	accountId, contacType, diagErr := DecodeAlternateContactId(d.Id())
	if diagErr != nil {
		return diagErr
	}

	output, err := FindAlternateContactByContactType(ctx, conn, accountId, contacType)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Account Alternate Contact (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Account Alternate Contact (%s): %s", d.Id(), err)
	}

	d.Set("account_id", accountId)
	d.Set("alternate_contact_type", output.AlternateContactType)
	d.Set("email_address", output.EmailAddress)
	d.Set("name", output.Name)
	d.Set("phone_number", output.PhoneNumber)
	d.Set("title", output.Title)

	return nil
}

func resourceAlternateContactUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountConn

	input := &account.PutAlternateContactInput{
		AlternateContactType: aws.String(d.Get("alternate_contact_type").(string)),
		EmailAddress:         aws.String(d.Get("email_address").(string)),
		Name:                 aws.String(d.Get("name").(string)),
		PhoneNumber:          aws.String(d.Get("phone_number").(string)),
		Title:                aws.String(d.Get("title").(string)),
	}

	if v, ok := d.GetOk("account_id"); ok {
		input.AccountId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Updating Account Alternate Contact: %s", input)
	_, err := conn.PutAlternateContactWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error updating Account Alternate Contact (%s): %s", d.Id(), err)
	}

	return resourceAlternateContactRead(ctx, d, meta)
}

func resourceAlternateContactDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountConn

	log.Printf("[DEBUG] Deleting Account Alternate Contact: %s", d.Id())

	accountId, contacType, diagErr := DecodeAlternateContactId(d.Id())
	if diagErr != nil {
		return diagErr
	}

	input := &account.DeleteAlternateContactInput{
		AlternateContactType: aws.String(contacType),
	}

	if accountId != "" {
		input.AccountId = aws.String(accountId)
	}

	_, err := conn.DeleteAlternateContactWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, account.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Account Alternate Contact (%s): %s", d.Id(), err)
	}

	return nil
}

func FindAlternateContactByContactType(ctx context.Context, conn *account.Account, accountId string, contactType string) (*account.AlternateContact, error) {
	input := &account.GetAlternateContactInput{
		AlternateContactType: aws.String(contactType),
	}

	if accountId != "" {
		input.AccountId = aws.String(accountId)
	}

	output, err := conn.GetAlternateContactWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, account.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AlternateContact == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AlternateContact, nil
}

func DecodeAlternateContactId(id string) (string, string, diag.Diagnostics) {
	parts := strings.Split(id, "/")

	switch len(parts) {
	case 1:
		return "", parts[0], nil
	case 2:
		return parts[0], parts[1], nil
	default:
		return "", "", diag.Errorf("Expected ID in the form of AccountId/ContactType or ContactType, given: %q", id)
	}
}
