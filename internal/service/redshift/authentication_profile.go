// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_redshift_authentication_profile", name="Authentication Profile")
func resourceAuthenticationProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAuthenticationProfileCreate,
		ReadWithoutTimeout:   resourceAuthenticationProfileRead,
		UpdateWithoutTimeout: resourceAuthenticationProfileUpdate,
		DeleteWithoutTimeout: resourceAuthenticationProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"authentication_profile_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"authentication_profile_content": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceAuthenticationProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	authProfileName := d.Get("authentication_profile_name").(string)

	input := redshift.CreateAuthenticationProfileInput{
		AuthenticationProfileName:    aws.String(authProfileName),
		AuthenticationProfileContent: aws.String(d.Get("authentication_profile_content").(string)),
	}

	out, err := conn.CreateAuthenticationProfileWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Authentication Profile (%s): %s", authProfileName, err)
	}

	d.SetId(aws.StringValue(out.AuthenticationProfileName))

	return append(diags, resourceAuthenticationProfileRead(ctx, d, meta)...)
}

func resourceAuthenticationProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	out, err := findAuthenticationProfileByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Authentication Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Authentication Profile (%s): %s", d.Id(), err)
	}

	d.Set("authentication_profile_content", out.AuthenticationProfileContent)
	d.Set("authentication_profile_name", out.AuthenticationProfileName)

	return diags
}

func resourceAuthenticationProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	input := &redshift.ModifyAuthenticationProfileInput{
		AuthenticationProfileName:    aws.String(d.Id()),
		AuthenticationProfileContent: aws.String(d.Get("authentication_profile_content").(string)),
	}

	_, err := conn.ModifyAuthenticationProfileWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "modifying Redshift Authentication Profile (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAuthenticationProfileRead(ctx, d, meta)...)
}

func resourceAuthenticationProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	deleteInput := redshift.DeleteAuthenticationProfileInput{
		AuthenticationProfileName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Redshift Authentication Profile: %s", d.Id())
	_, err := conn.DeleteAuthenticationProfileWithContext(ctx, &deleteInput)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeAuthenticationProfileNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Authentication Profile (%s): %s", d.Id(), err)
	}

	return diags
}
