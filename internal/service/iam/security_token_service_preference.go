package iam

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_iam_security_token_service_preferences", name="Security Token Service Preferences")
// @Tags
func ResourceSecurityTokenServicePreferences() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityTokenServicePreferencesUpsert,
		ReadWithoutTimeout:   resourceSecurityTokenServicePreferencesRead,
		UpdateWithoutTimeout: resourceSecurityTokenServicePreferencesUpsert,
		DeleteWithoutTimeout: resourceSecurityTokenServicePreferencesDelete,

		Schema: map[string]*schema.Schema{
			"global_endpoint_token_version": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceSecurityTokenServicePreferencesUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	input := &iam.SetSecurityTokenServicePreferencesInput{
		GlobalEndpointTokenVersion: aws.String(d.Get("global_endpoint_token_version").(string)),
	}

	_, err := conn.SetSecurityTokenServicePreferencesWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting IAM Security Token Service Preferences: %s", err)
	}

	d.SetId(fmt.Sprintf("iam-security-token-service-preferences.%s", *input.GlobalEndpointTokenVersion))

	return append(diags, resourceSecurityTokenServicePreferencesRead(ctx, d, meta)...)
}

func resourceSecurityTokenServicePreferencesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	d.Set("id", d.Id())

	return diags
}

func resourceSecurityTokenServicePreferencesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	log.Printf("[INFO] Deleting IAM Instance Profile: %s", d.Id())

	input := &iam.SetSecurityTokenServicePreferencesInput{
		GlobalEndpointTokenVersion: aws.String("v1Token"),
	}

	_, err := conn.SetSecurityTokenServicePreferencesWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting IAM Security Token Service Preferences: %s", err)
	}

	d.SetId("")

	return diags
}
