package securityhub

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountCreate,
		ReadWithoutTimeout:   resourceAccountRead,
		DeleteWithoutTimeout: resourceAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{},
	}
}

func resourceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn()
	log.Print("[DEBUG] Enabling Security Hub for account")

	_, err := conn.EnableSecurityHubWithContext(ctx, &securityhub.EnableSecurityHubInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling Security Hub for account: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return append(diags, resourceAccountRead(ctx, d, meta)...)
}

func resourceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn()

	log.Printf("[DEBUG] Checking if Security Hub is enabled")
	_, err := conn.GetEnabledStandardsWithContext(ctx, &securityhub.GetEnabledStandardsInput{})

	if err != nil {
		// Can only read enabled standards if Security Hub is enabled
		if tfawserr.ErrMessageContains(err, "InvalidAccessException", "not subscribed to AWS Security Hub") {
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "checking if Security Hub is enabled: %s", err)
	}

	return diags
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn()
	log.Print("[DEBUG] Disabling Security Hub for account")

	err := resource.RetryContext(ctx, adminAccountNotFoundTimeout, func() *resource.RetryError {
		_, err := conn.DisableSecurityHubWithContext(ctx, &securityhub.DisableSecurityHubInput{})

		if tfawserr.ErrMessageContains(err, securityhub.ErrCodeInvalidInputException, "Cannot disable Security Hub on the Security Hub administrator") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DisableSecurityHubWithContext(ctx, &securityhub.DisableSecurityHubInput{})
	}

	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Security Hub for account: %s", err)
	}

	return diags
}
