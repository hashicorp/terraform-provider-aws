// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKResource("aws_cognito_user_pool_ui_customization", name="User Pool UI Customization")
func resourceUserPoolUICustomization() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserPoolUICustomizationPut,
		ReadWithoutTimeout:   resourceUserPoolUICustomizationRead,
		UpdateWithoutTimeout: resourceUserPoolUICustomizationPut,
		DeleteWithoutTimeout: resourceUserPoolUICustomizationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ALL",
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"css": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"css", "image_file"},
			},
			"css_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_file": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"image_file", "css"},
			},
			"image_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

const (
	userPoolUICustomizationResourceIDPartCount = 2
)

func resourceUserPoolUICustomizationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn(ctx)

	userPoolID, clientID := d.Get("user_pool_id").(string), d.Get("client_id").(string)
	id := errs.Must(flex.FlattenResourceId([]string{userPoolID, clientID}, userPoolUICustomizationResourceIDPartCount, false))
	input := &cognitoidentityprovider.SetUICustomizationInput{
		ClientId:   aws.String(clientID),
		UserPoolId: aws.String(userPoolID),
	}

	if v, ok := d.GetOk("css"); ok {
		input.CSS = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_file"); ok {
		v, err := itypes.Base64Decode(v.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		input.ImageFile = v
	}

	_, err := conn.SetUICustomizationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Cognito User Pool UI Customization (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceUserPoolUICustomizationRead(ctx, d, meta)...)
}

func resourceUserPoolUICustomizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), userPoolUICustomizationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	userPoolID, clientID := parts[0], parts[1]

	uiCustomization, err := findUserPoolUICustomizationByTwoPartKey(ctx, conn, userPoolID, clientID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cognito User Pool UI Customization %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cognito User Pool UI Customization (%s): %s", d.Id(), err)
	}

	d.Set("client_id", uiCustomization.ClientId)
	d.Set("creation_date", aws.TimeValue(uiCustomization.CreationDate).Format(time.RFC3339))
	d.Set("css", uiCustomization.CSS)
	d.Set("css_version", uiCustomization.CSSVersion)
	d.Set("image_url", uiCustomization.ImageUrl)
	d.Set("last_modified_date", aws.TimeValue(uiCustomization.LastModifiedDate).Format(time.RFC3339))
	d.Set("user_pool_id", uiCustomization.UserPoolId)

	return diags
}

func resourceUserPoolUICustomizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), userPoolUICustomizationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	userPoolID, clientID := parts[0], parts[1]

	log.Printf("[DEBUG] Deleting Cognito User Pool UI Customization: %s", d.Id())
	_, err = conn.SetUICustomizationWithContext(ctx, &cognitoidentityprovider.SetUICustomizationInput{
		ClientId:   aws.String(clientID),
		UserPoolId: aws.String(userPoolID),
	})

	if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cognito User Pool UI Customization (%s): %s", d.Id(), err)
	}

	return diags
}

func findUserPoolUICustomizationByTwoPartKey(ctx context.Context, conn *cognitoidentityprovider.CognitoIdentityProvider, userPoolID, clientID string) (*cognitoidentityprovider.UICustomizationType, error) {
	input := &cognitoidentityprovider.GetUICustomizationInput{
		ClientId:   aws.String(clientID),
		UserPoolId: aws.String(userPoolID),
	}

	output, err := conn.GetUICustomizationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	// The GetUICustomization API operation will return an empty struct
	// if nothing is present rather than nil or an error, so we equate that with nil.
	if output == nil || output.UICustomization == nil || itypes.IsZero(output.UICustomization) {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.UICustomization, nil
}
