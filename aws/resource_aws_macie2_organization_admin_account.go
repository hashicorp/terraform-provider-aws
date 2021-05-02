package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsMacie2OrganizationAdminAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMacie2OrganizationAdminAccountCreate,
		ReadWithoutTimeout:   resourceMacie2OrganizationAdminAccountRead,
		DeleteWithoutTimeout: resourceMacie2OrganizationAdminAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"admin_account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMacie2OrganizationAdminAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.EnableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(d.Get("admin_account_id").(string)),
		ClientToken:    aws.String(resource.UniqueId()),
	}

	var err error
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		_, err := conn.EnableOrganizationAdminAccountWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.EnableOrganizationAdminAccountWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Macie OrganizationAdminAccount: %w", err))
	}

	d.SetId(meta.(*AWSClient).accountid)

	return resourceMacie2OrganizationAdminAccountRead(ctx, d, meta)
}

func resourceMacie2OrganizationAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	var err error
	var res *macie2.AdminAccount

	err = conn.ListOrganizationAdminAccountsPages(&macie2.ListOrganizationAdminAccountsInput{}, func(page *macie2.ListOrganizationAdminAccountsOutput, lastPage bool) bool {
		for _, account := range page.AdminAccounts {
			if aws.StringValue(account.AccountId) != d.Get("admin_account_id").(string) {
				res = account
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) {
			log.Printf("[WARN] Macie OrganizationAdminAccount does not exist, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error reading Macie OrganizationAdminAccount (%s): %w", d.Id(), err))
	}

	if res == nil {
		log.Printf("[WARN] Macie OrganizationAdminAccount (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("account_id", res.AccountId)
	d.Set("status", res.Status)

	return nil
}

func resourceMacie2OrganizationAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.DisableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(d.Id()),
	}

	_, err := conn.DisableOrganizationAdminAccountWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Macie OrganizationAdminAccount (%s): %w", d.Id(), err))
	}
	return nil
}
