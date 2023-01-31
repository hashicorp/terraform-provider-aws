package account

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/account"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAlternateContact() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAlternateContactCreate,
		ReadWithoutTimeout:   resourceAlternateContactRead,
		UpdateWithoutTimeout: resourceAlternateContactUpdate,
		DeleteWithoutTimeout: resourceAlternateContactDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(alternateContactCreateTimeout),
			Update: schema.DefaultTimeout(alternateContactUpdateTimeout),
			Delete: schema.DefaultTimeout(alternateContactDeleteTimeout),
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
	conn := meta.(*conns.AWSClient).AccountConn()

	contactType := d.Get("alternate_contact_type").(string)
	input := &account.PutAlternateContactInput{
		AlternateContactType: aws.String(contactType),
		EmailAddress:         aws.String(d.Get("email_address").(string)),
		Name:                 aws.String(d.Get("name").(string)),
		PhoneNumber:          aws.String(d.Get("phone_number").(string)),
		Title:                aws.String(d.Get("title").(string)),
	}

	accountID := d.Get("account_id").(string)
	if accountID != "" {
		input.AccountId = aws.String(accountID)
	}
	id := AlternateContactCreateResourceID(accountID, contactType)

	log.Printf("[DEBUG] Creating Account Alternate Contact: %s", input)
	_, err := conn.PutAlternateContactWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating Account Alternate Contact (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitAlternateContactCreated(ctx, conn, accountID, contactType, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("error waiting for Account Alternate Contact (%s) create: %s", d.Id(), err)
	}

	return resourceAlternateContactRead(ctx, d, meta)
}

func resourceAlternateContactRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountConn()

	accountID, contactType, err := AlternateContactParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	output, err := FindAlternateContactByAccountIDAndContactType(ctx, conn, accountID, contactType)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Account Alternate Contact (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Account Alternate Contact (%s): %s", d.Id(), err)
	}

	d.Set("account_id", accountID)
	d.Set("alternate_contact_type", output.AlternateContactType)
	d.Set("email_address", output.EmailAddress)
	d.Set("name", output.Name)
	d.Set("phone_number", output.PhoneNumber)
	d.Set("title", output.Title)

	return nil
}

func resourceAlternateContactUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountConn()

	accountID, contactType, err := AlternateContactParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	email := d.Get("email_address").(string)
	name := d.Get("name").(string)
	phone := d.Get("phone_number").(string)
	title := d.Get("title").(string)

	input := &account.PutAlternateContactInput{
		AlternateContactType: aws.String(contactType),
		EmailAddress:         aws.String(email),
		Name:                 aws.String(name),
		PhoneNumber:          aws.String(phone),
		Title:                aws.String(title),
	}

	if accountID != "" {
		input.AccountId = aws.String(accountID)
	}

	log.Printf("[DEBUG] Updating Account Alternate Contact: %s", input)
	_, err = conn.PutAlternateContactWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error updating Account Alternate Contact (%s): %s", d.Id(), err)
	}

	if err := waitAlternateContactUpdated(ctx, conn, accountID, contactType, email, name, phone, title, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return diag.Errorf("error waiting for Account Alternate Contact (%s) update: %s", d.Id(), err)
	}

	return resourceAlternateContactRead(ctx, d, meta)
}

func resourceAlternateContactDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountConn()

	accountID, contactType, err := AlternateContactParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &account.DeleteAlternateContactInput{
		AlternateContactType: aws.String(contactType),
	}

	if accountID != "" {
		input.AccountId = aws.String(accountID)
	}

	log.Printf("[DEBUG] Deleting Account Alternate Contact: %s", d.Id())
	_, err = conn.DeleteAlternateContactWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, account.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Account Alternate Contact (%s): %s", d.Id(), err)
	}

	if err := waitAlternateContactDeleted(ctx, conn, accountID, contactType, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("error waiting for Account Alternate Contact (%s) delete: %s", d.Id(), err)
	}

	return nil
}

const alternateContactResourceIDSeparator = "/"

func AlternateContactCreateResourceID(accountID, contactType string) string {
	parts := []string{accountID, contactType}
	id := strings.Join(parts, alternateContactResourceIDSeparator)

	return id
}

func AlternateContactParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, alternateContactResourceIDSeparator)

	switch len(parts) {
	case 1:
		return "", parts[0], nil
	case 2:
		return parts[0], parts[1], nil
	default:
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected ContactType or AccountID%[2]sContactType", id, alternateContactResourceIDSeparator)
	}
}
