// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"errors"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_authorizer", name="Authorizer")
// @Tags(identifierAttribute="arn")
func resourceAuthorizer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAuthorizerCreate,
		ReadWithoutTimeout:   resourceAuthorizerRead,
		UpdateWithoutTimeout: resourceAuthorizerUpdate,
		DeleteWithoutTimeout: resourceAuthorizerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			resourceAuthorizerCustomizeDiff,
		),

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexache.MustCompile(`^[\w=,@-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),
				),
			},
			"signing_disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrStatus: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.AuthorizerStatusActive,
				ValidateDiagFunc: enum.Validate[awstypes.AuthorizerStatus](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"token_key_name": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),
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
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &iot.CreateAuthorizerInput{
		AuthorizerFunctionArn: aws.String(d.Get("authorizer_function_arn").(string)),
		AuthorizerName:        aws.String(name),
		EnableCachingForHttp:  aws.Bool(d.Get("enable_caching_for_http").(bool)),
		SigningDisabled:       aws.Bool(d.Get("signing_disabled").(bool)),
		Status:                awstypes.AuthorizerStatus((d.Get(names.AttrStatus).(string))),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk("token_key_name"); ok {
		input.TokenKeyName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("token_signing_public_keys"); ok {
		input.TokenSigningPublicKeys = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	output, err := conn.CreateAuthorizer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Authorizer (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.AuthorizerName))

	return append(diags, resourceAuthorizerRead(ctx, d, meta)...)
}

func resourceAuthorizerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	authorizer, err := findAuthorizerByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Authorizer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Authorizer (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, authorizer.AuthorizerArn)
	d.Set("authorizer_function_arn", authorizer.AuthorizerFunctionArn)
	d.Set("enable_caching_for_http", authorizer.EnableCachingForHttp)
	d.Set(names.AttrName, authorizer.AuthorizerName)
	d.Set("signing_disabled", authorizer.SigningDisabled)
	d.Set(names.AttrStatus, authorizer.Status)
	d.Set("token_key_name", authorizer.TokenKeyName)
	d.Set("token_signing_public_keys", aws.StringMap(authorizer.TokenSigningPublicKeys))

	return diags
}

func resourceAuthorizerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	input := iot.UpdateAuthorizerInput{
		AuthorizerName: aws.String(d.Id()),
	}

	if d.HasChange("authorizer_function_arn") {
		input.AuthorizerFunctionArn = aws.String(d.Get("authorizer_function_arn").(string))
	}

	if d.HasChange("enable_caching_for_http") {
		input.EnableCachingForHttp = aws.Bool(d.Get("enable_caching_for_http").(bool))
	}

	if d.HasChange(names.AttrStatus) {
		input.Status = awstypes.AuthorizerStatus(d.Get(names.AttrStatus).(string))
	}

	if d.HasChange("token_key_name") {
		input.TokenKeyName = aws.String(d.Get("token_key_name").(string))
	}

	if d.HasChange("token_signing_public_keys") {
		input.TokenSigningPublicKeys = flex.ExpandStringValueMap(d.Get("token_signing_public_keys").(map[string]interface{}))
	}

	_, err := conn.UpdateAuthorizer(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IoT Authorizer (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAuthorizerRead(ctx, d, meta)...)
}

func resourceAuthorizerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	// In order to delete an IoT Authorizer, you must set it inactive first.
	if d.Get(names.AttrStatus).(string) == string(awstypes.AuthorizerStatusActive) {
		_, err := conn.UpdateAuthorizer(ctx, &iot.UpdateAuthorizerInput{
			AuthorizerName: aws.String(d.Id()),
			Status:         awstypes.AuthorizerStatusInactive,
		})

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deactivating IoT Authorizer (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[INFO] Deleting IoT Authorizer: %s", d.Id())
	_, err := conn.DeleteAuthorizer(ctx, &iot.DeleteAuthorizerInput{
		AuthorizerName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findAuthorizerByName(ctx context.Context, conn *iot.Client, name string) (*awstypes.AuthorizerDescription, error) {
	input := &iot.DescribeAuthorizerInput{
		AuthorizerName: aws.String(name),
	}

	output, err := conn.DescribeAuthorizer(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AuthorizerDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AuthorizerDescription, nil
}
