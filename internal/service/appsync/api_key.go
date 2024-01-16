// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_appsync_api_key")
func ResourceAPIKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAPIKeyCreate,
		ReadWithoutTimeout:   resourceAPIKeyRead,
		UpdateWithoutTimeout: resourceAPIKeyUpdate,
		DeleteWithoutTimeout: resourceAPIKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"expires": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Ignore unsetting value
					if old != "" && new == "" {
						return true
					}
					return false
				},
				ValidateFunc: validation.IsRFC3339Time,
			},
			"key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceAPIKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	apiID := d.Get("api_id").(string)

	params := &appsync.CreateApiKeyInput{
		ApiId:       aws.String(apiID),
		Description: aws.String(d.Get("description").(string)),
	}
	if v, ok := d.GetOk("expires"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		params.Expires = aws.Int64(t.Unix())
	}
	resp, err := conn.CreateApiKeyWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Appsync API Key: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", apiID, aws.StringValue(resp.ApiKey.Id)))
	return append(diags, resourceAPIKeyRead(ctx, d, meta)...)
}

func resourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	apiID, keyID, err := DecodeAPIKeyID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appsync API Key (%s): %s", d.Id(), err)
	}

	key, err := GetAPIKey(ctx, apiID, keyID, conn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appsync API Key (%s): %s", d.Id(), err)
	}
	if key == nil && !d.IsNewResource() {
		log.Printf("[WARN] AppSync API Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("api_id", apiID)
	d.Set("key", key.Id)
	d.Set("description", key.Description)
	d.Set("expires", time.Unix(aws.Int64Value(key.Expires), 0).UTC().Format(time.RFC3339))
	return diags
}

func resourceAPIKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	apiID, keyID, err := DecodeAPIKeyID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Appsync API Key (%s): %s", d.Id(), err)
	}

	params := &appsync.UpdateApiKeyInput{
		ApiId: aws.String(apiID),
		Id:    aws.String(keyID),
	}
	if d.HasChange("description") {
		params.Description = aws.String(d.Get("description").(string))
	}
	if d.HasChange("expires") {
		t, _ := time.Parse(time.RFC3339, d.Get("expires").(string))
		params.Expires = aws.Int64(t.Unix())
	}

	_, err = conn.UpdateApiKeyWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Appsync API Key (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAPIKeyRead(ctx, d, meta)...)
}

func resourceAPIKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	apiID, keyID, err := DecodeAPIKeyID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appsync API Key (%s): %s", d.Id(), err)
	}

	input := &appsync.DeleteApiKeyInput{
		ApiId: aws.String(apiID),
		Id:    aws.String(keyID),
	}
	_, err = conn.DeleteApiKeyWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Appsync API Key (%s): %s", d.Id(), err)
	}

	return diags
}

func DecodeAPIKeyID(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected API-ID:API-KEY-ID", id)
	}
	return parts[0], parts[1], nil
}

func GetAPIKey(ctx context.Context, apiID, keyID string, conn *appsync.AppSync) (*appsync.ApiKey, error) {
	input := &appsync.ListApiKeysInput{
		ApiId: aws.String(apiID),
	}
	for {
		resp, err := conn.ListApiKeysWithContext(ctx, input)
		if err != nil {
			return nil, err
		}
		for _, apiKey := range resp.ApiKeys {
			if aws.StringValue(apiKey.Id) == keyID {
				return apiKey, nil
			}
		}
		if resp.NextToken == nil {
			break
		}
		input.NextToken = resp.NextToken
	}
	return nil, nil
}
