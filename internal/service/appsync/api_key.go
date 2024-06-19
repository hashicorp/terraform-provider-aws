// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appsync_api_key", name="API Key")
func resourceAPIKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAPIKeyCreate,
		ReadWithoutTimeout:   resourceAPIKeyRead,
		UpdateWithoutTimeout: resourceAPIKeyUpdate,
		DeleteWithoutTimeout: resourceAPIKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"api_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			"expires": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Ignore unsetting value.
					if old != "" && new == "" {
						return true
					}
					return false
				},
				ValidateFunc: validation.IsRFC3339Time,
			},
			names.AttrKey: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceAPIKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID := d.Get("api_id").(string)
	input := &appsync.CreateApiKeyInput{
		ApiId:       aws.String(apiID),
		Description: aws.String(d.Get(names.AttrDescription).(string)),
	}

	if v, ok := d.GetOk("expires"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		input.Expires = t.Unix()
	}

	output, err := conn.CreateApiKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Appsync API Key: %s", err)
	}

	d.SetId(apiKeyCreateResourceID(apiID, aws.ToString(output.ApiKey.Id)))

	return append(diags, resourceAPIKeyRead(ctx, d, meta)...)
}

func resourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, keyID, err := apiKeyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	key, err := findAPIKeyByTwoPartKey(ctx, conn, apiID, keyID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppSync API Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appsync API Key (%s): %s", d.Id(), err)
	}

	d.Set("api_id", apiID)
	d.Set("api_key_id", keyID)
	d.Set(names.AttrDescription, key.Description)
	d.Set("expires", time.Unix(key.Expires, 0).UTC().Format(time.RFC3339))
	d.Set(names.AttrKey, key.Id)

	return diags
}

func resourceAPIKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, keyID, err := apiKeyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &appsync.UpdateApiKeyInput{
		ApiId: aws.String(apiID),
		Id:    aws.String(keyID),
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange("expires") {
		t, _ := time.Parse(time.RFC3339, d.Get("expires").(string))
		input.Expires = t.Unix()
	}

	_, err = conn.UpdateApiKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Appsync API Key (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAPIKeyRead(ctx, d, meta)...)
}

func resourceAPIKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, keyID, err := apiKeyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting Appsync API Key: %s", d.Id())
	_, err = conn.DeleteApiKey(ctx, &appsync.DeleteApiKeyInput{
		ApiId: aws.String(apiID),
		Id:    aws.String(keyID),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appsync API Key (%s): %s", d.Id(), err)
	}

	return diags
}

const apiKeyResourceIDSeparator = ":"

func apiKeyCreateResourceID(apiID, keyID string) string {
	parts := []string{apiID, keyID}
	id := strings.Join(parts, apiKeyResourceIDSeparator)

	return id
}

func apiKeyParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, apiKeyResourceIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected API-ID%[2]sAPI-KEY-ID", id, apiKeyResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findAPIKey(ctx context.Context, conn *appsync.Client, input *appsync.ListApiKeysInput, filter tfslices.Predicate[*awstypes.ApiKey]) (*awstypes.ApiKey, error) {
	output, err := findAPIKeys(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAPIKeys(ctx context.Context, conn *appsync.Client, input *appsync.ListApiKeysInput, filter tfslices.Predicate[*awstypes.ApiKey]) ([]awstypes.ApiKey, error) {
	var output []awstypes.ApiKey

	err := listAPIKeysPages(ctx, conn, input, func(page *appsync.ListApiKeysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ApiKeys {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findAPIKeyByTwoPartKey(ctx context.Context, conn *appsync.Client, apiID, keyID string) (*awstypes.ApiKey, error) {
	input := &appsync.ListApiKeysInput{
		ApiId: aws.String(apiID),
	}

	return findAPIKey(ctx, conn, input, func(v *awstypes.ApiKey) bool {
		return aws.ToString(v.Id) == keyID
	})
}
