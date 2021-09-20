package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/securityhub/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/securityhub/waiter"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceOrganizationAdminAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceOrganizationAdminAccountCreate,
		Read:   resourceOrganizationAdminAccountRead,
		Delete: resourceOrganizationAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"admin_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
		},
	}
}

func resourceOrganizationAdminAccountCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SecurityHubConn

	adminAccountID := d.Get("admin_account_id").(string)

	input := &securityhub.EnableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(adminAccountID),
	}

	_, err := conn.EnableOrganizationAdminAccount(input)

	if err != nil {
		return fmt.Errorf("error enabling Security Hub Organization Admin Account (%s): %w", adminAccountID, err)
	}

	d.SetId(adminAccountID)

	if _, err := waiter.waitAdminAccountEnabled(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Security Hub Organization Admin Account (%s) to enable: %w", d.Id(), err)
	}

	return resourceOrganizationAdminAccountRead(d, meta)
}

func resourceOrganizationAdminAccountRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SecurityHubConn

	adminAccount, err := finder.FindAdminAccount(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Security Hub Organization Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Security Hub Organization Admin Account (%s): %w", d.Id(), err)
	}

	if adminAccount == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading Security Hub Organization Admin Account (%s): %w", d.Id(), err)
		}

		log.Printf("[WARN] Security Hub Organization Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("admin_account_id", adminAccount.AccountId)

	return nil
}

func resourceOrganizationAdminAccountDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SecurityHubConn

	input := &securityhub.DisableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(d.Id()),
	}

	_, err := conn.DisableOrganizationAdminAccount(input)

	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disabling Security Hub Organization Admin Account (%s): %w", d.Id(), err)
	}

	if _, err := waiter.waitAdminAccountNotFound(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Security Hub Organization Admin Account (%s) to disable: %w", d.Id(), err)
	}

	return nil
}
