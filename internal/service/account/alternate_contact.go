package account

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/account"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(account.AlternateContactType_Values(), false),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
			"email_address": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[\w+=,.-]+@[\w.-]+\.[\w]+`), "must be a valid email address"),
			},
			"phone_number": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\s0-9()+-]+$`), "must be a valid phone number"),
			},
		},
	}
}

func resourceAlternateContactCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountConn

	contactType := d.Get("type").(string)

	input := &account.PutAlternateContactInput{
		AlternateContactType: aws.String(contactType),
		Name:                 aws.String(d.Get("name").(string)),
		EmailAddress:         aws.String(d.Get("email_address").(string)),
		PhoneNumber:          aws.String(d.Get("phone_number").(string)),
		Title:                aws.String(d.Get("title").(string)),
	}

	log.Printf("[DEBUG] Creating %s account alternate contact: %s", contactType, input)
	_, err := conn.PutAlternateContactWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("error creating %s account alternate contact: %s", contactType, err)
	}

	d.SetId(contactType)
	return resourceAlternateContactRead(ctx, d, meta)
}

func resourceAlternateContactRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountConn

	alcon, err := conn.GetAlternateContactWithContext(ctx, &account.GetAlternateContactInput{
		AlternateContactType: aws.String(d.Id()),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, account.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] %s account alternate contact not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading %s account alternate contact: %s", d.Id(), err)
	}

	d.Set("type", d.Id())
	d.Set("name", aws.StringValue(alcon.AlternateContact.Name))
	d.Set("title", aws.StringValue(alcon.AlternateContact.Title))
	d.Set("email_address", aws.StringValue(alcon.AlternateContact.EmailAddress))
	d.Set("phone_number", aws.StringValue(alcon.AlternateContact.PhoneNumber))

	return nil
}

func resourceAlternateContactUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountConn

	contactType := d.Get("type").(string)

	input := &account.PutAlternateContactInput{
		AlternateContactType: aws.String(contactType),
		Name:                 aws.String(d.Get("name").(string)),
		EmailAddress:         aws.String(d.Get("email_address").(string)),
		PhoneNumber:          aws.String(d.Get("phone_number").(string)),
		Title:                aws.String(d.Get("title").(string)),
	}

	log.Printf("[DEBUG] Updating %s account alternate contact: %s", contactType, input)
	_, err := conn.PutAlternateContactWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("error updating %s account alternate contact: %s", contactType, err)
	}

	return resourceAlternateContactRead(ctx, d, meta)
}

func resourceAlternateContactDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountConn

	input := &account.DeleteAlternateContactInput{
		AlternateContactType: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting %s account alternate contact: %s", d.Id(), input)
	_, err := conn.DeleteAlternateContactWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, account.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting %s account alternate contact: %s", d.Id(), err)
	}

	return nil
}
