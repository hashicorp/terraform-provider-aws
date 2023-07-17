// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"errors"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_iot_authorizer")
func ResourceAuthorizer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAuthorizerCreate,
		ReadWithoutTimeout:   resourceAuthorizerRead,
		UpdateWithoutTimeout: resourceAuthorizerUpdate,
		DeleteWithoutTimeout: resourceAuthorizerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: resourceAuthorizerCustomizeDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorizer_function_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"enable_caching_for_http": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[\w=,@-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),
				),
			},
			"signing_disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      iot.AuthorizerStatusActive,
				ValidateFunc: validation.StringInSlice(iot.AuthorizerStatus_Values(), false),
			},
			"token_key_name": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9_-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),
				),
			},
			"token_signing_public_keys": {
				Type:      schema.TypeMap,
				Optional:  true,
				Elem:      &schema.Schema{Type: schema.TypeString},
				Sensitive: true,
			},
		},
	}
}

func resourceAuthorizerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	name := d.Get("name").(string)
	input := &iot.CreateAuthorizerInput{
		AuthorizerFunctionArn: aws.String(d.Get("authorizer_function_arn").(string)),
		AuthorizerName:        aws.String(name),
		EnableCachingForHttp:  aws.Bool(d.Get("enable_caching_for_http").(bool)),
		SigningDisabled:       aws.Bool(d.Get("signing_disabled").(bool)),
		Status:                aws.String(d.Get("status").(string)),
	}

	if v, ok := d.GetOk("token_key_name"); ok {
		input.TokenKeyName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("token_signing_public_keys"); ok {
		input.TokenSigningPublicKeys = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	log.Printf("[INFO] Creating IoT Authorizer: %s", input)
	output, err := conn.CreateAuthorizerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Authorizer (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.AuthorizerName))

	return append(diags, resourceAuthorizerRead(ctx, d, meta)...)
}

func resourceAuthorizerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	authorizer, err := FindAuthorizerByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Authorizer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Authorizer (%s): %s", d.Id(), err)
	}

	d.Set("arn", authorizer.AuthorizerArn)
	d.Set("authorizer_function_arn", authorizer.AuthorizerFunctionArn)
	d.Set("enable_caching_for_http", authorizer.EnableCachingForHttp)
	d.Set("name", authorizer.AuthorizerName)
	d.Set("signing_disabled", authorizer.SigningDisabled)
	d.Set("status", authorizer.Status)
	d.Set("token_key_name", authorizer.TokenKeyName)
	d.Set("token_signing_public_keys", aws.StringValueMap(authorizer.TokenSigningPublicKeys))

	return diags
}

func resourceAuthorizerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	input := iot.UpdateAuthorizerInput{
		AuthorizerName: aws.String(d.Id()),
	}

	if d.HasChange("authorizer_function_arn") {
		input.AuthorizerFunctionArn = aws.String(d.Get("authorizer_function_arn").(string))
	}

	if d.HasChange("enable_caching_for_http") {
		input.EnableCachingForHttp = aws.Bool(d.Get("enable_caching_for_http").(bool))
	}

	if d.HasChange("status") {
		input.Status = aws.String(d.Get("status").(string))
	}

	if d.HasChange("token_key_name") {
		input.TokenKeyName = aws.String(d.Get("token_key_name").(string))
	}

	if d.HasChange("token_signing_public_keys") {
		input.TokenSigningPublicKeys = flex.ExpandStringMap(d.Get("token_signing_public_keys").(map[string]interface{}))
	}

	log.Printf("[INFO] Updating IoT Authorizer: %s", input)
	_, err := conn.UpdateAuthorizerWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IoT Authorizer (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAuthorizerRead(ctx, d, meta)...)
}

func resourceAuthorizerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	// In order to delete an IoT Authorizer, you must set it inactive first.
	if d.Get("status").(string) == iot.AuthorizerStatusActive {
		log.Printf("[INFO] Deactivating IoT Authorizer: %s", d.Id())
		_, err := conn.UpdateAuthorizerWithContext(ctx, &iot.UpdateAuthorizerInput{
			AuthorizerName: aws.String(d.Id()),
			Status:         aws.String(iot.AuthorizerStatusInactive),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deactivating IoT Authorizer (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[INFO] Deleting IoT Authorizer: %s", d.Id())
	_, err := conn.DeleteAuthorizerWithContext(ctx, &iot.DeleteAuthorizerInput{
		AuthorizerName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IOT Authorizer (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceAuthorizerCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if !diff.Get("signing_disabled").(bool) {
		if _, ok := diff.GetOk("token_key_name"); !ok {
			return errors.New(`"token_key_name" is required when signing is enabled`)
		}
		if _, ok := diff.GetOk("token_signing_public_keys"); !ok {
			return errors.New(`"token_signing_public_keys" is required when signing is enabled`)
		}
	}

	return nil
}
