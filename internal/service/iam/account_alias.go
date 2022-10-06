package iam

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceAccountAlias() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccountAliasCreate,
		Read:   resourceAccountAliasRead,
		Delete: resourceAccountAliasDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_alias": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validAccountAlias,
			},
		},
	}
}

func resourceAccountAliasCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	account_alias := d.Get("account_alias").(string)

	params := &iam.CreateAccountAliasInput{
		AccountAlias: aws.String(account_alias),
	}

	_, err := conn.CreateAccountAlias(params)

	if err != nil {
		return fmt.Errorf("Error creating account alias with name '%s': %w", account_alias, err)
	}

	d.SetId(account_alias)

	return nil
}

func resourceAccountAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	params := &iam.ListAccountAliasesInput{}

	resp, err := conn.ListAccountAliases(params)

	if err != nil {
		return fmt.Errorf("Error listing account aliases: %w", err)
	}

	if !d.IsNewResource() && (resp == nil || len(resp.AccountAliases) == 0) {
		d.SetId("")
		return nil
	}

	account_alias := aws.StringValue(resp.AccountAliases[0])

	d.SetId(account_alias)
	d.Set("account_alias", account_alias)

	return nil
}

func resourceAccountAliasDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	account_alias := d.Get("account_alias").(string)

	params := &iam.DeleteAccountAliasInput{
		AccountAlias: aws.String(account_alias),
	}

	_, err := conn.DeleteAccountAlias(params)

	if err != nil {
		return fmt.Errorf("Error deleting account alias with name '%s': %s", account_alias, err)
	}

	return nil
}
