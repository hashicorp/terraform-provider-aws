package aws

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/account"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsAccountAlternateContact() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAccountAlternateContactCreate,
		Read:   resourceAwsAccountAlternateContactRead,
		Update: resourceAwsAccountAlternateContactUpdate,
		Delete: resourceAwsAccountAlternateContactDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"alternate_contact_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"email_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"phone_number": {
				Type:     schema.TypeString,
				Required: true,
			},
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsAccountAlternateContactCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).accountconn

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
	_, err := conn.PutAlternateContact(input)

	if err != nil {
		return err
	}

	d.SetId(id)

	return resourceAwsAccountAlternateContactRead(d, meta)
}

func resourceAwsAccountAlternateContactRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).accountconn

	accountID, contactType, err := AlternateContactParseResourceID(d.Id())

	if err != nil {
		return err
	}

	output, err := FindAlternateContactByAccountIDAndContactType(conn, accountID, contactType)

	if err != nil {
		return err
	}

	d.Set("account_id", accountID)
	d.Set("alternate_contact_type", output.AlternateContactType)
	d.Set("email_address", output.EmailAddress)
	d.Set("name", output.Name)
	d.Set("phone_number", output.PhoneNumber)
	d.Set("title", output.Title)

	return nil
}

func resourceAwsAccountAlternateContactUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsAccountAlternateContactDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func FindAlternateContactByAccountIDAndContactType(conn *account.Account, accountID, contactType string) (*account.AlternateContact, error) {
	input := &account.GetAlternateContactInput{
		AlternateContactType: aws.String(contactType),
	}

	if accountID != "" {
		input.AccountId = aws.String(accountID)
	}

	output, err := conn.GetAlternateContact(input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.AlternateContact == nil {
		return nil, errors.New("EmptyResultError")
	}

	return output.AlternateContact, nil
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
