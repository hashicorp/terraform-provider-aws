// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_appsync_type")
func ResourceType() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTypeCreate,
		ReadWithoutTimeout:   resourceTypeRead,
		UpdateWithoutTimeout: resourceTypeUpdate,
		DeleteWithoutTimeout: resourceTypeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"definition": {
				Type:     schema.TypeString,
				Required: true,
			},
			"format": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TypeDefinitionFormat](),
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTypeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID := d.Get("api_id").(string)

	params := &appsync.CreateTypeInput{
		ApiId:      aws.String(apiID),
		Definition: aws.String(d.Get("definition").(string)),
		Format:     awstypes.TypeDefinitionFormat(d.Get("format").(string)),
	}

	out, err := conn.CreateType(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Appsync Type: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", apiID, string(out.Type.Format), aws.ToString(out.Type.Name)))

	return append(diags, resourceTypeRead(ctx, d, meta)...)
}

func resourceTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, format, name, err := DecodeTypeID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appsync Type %q: %s", d.Id(), err)
	}

	resp, err := FindTypeByThreePartKey(ctx, conn, apiID, format, name)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppSync Type (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appsync Type %q: %s", d.Id(), err)
	}

	d.Set("api_id", apiID)
	d.Set("arn", resp.Arn)
	d.Set("name", resp.Name)
	d.Set("format", resp.Format)
	d.Set("definition", resp.Definition)
	d.Set("description", resp.Description)

	return diags
}

func resourceTypeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	params := &appsync.UpdateTypeInput{
		ApiId:      aws.String(d.Get("api_id").(string)),
		Format:     awstypes.TypeDefinitionFormat(d.Get("format").(string)),
		TypeName:   aws.String(d.Get("name").(string)),
		Definition: aws.String(d.Get("definition").(string)),
	}

	_, err := conn.UpdateType(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Appsync Type %q: %s", d.Id(), err)
	}

	return append(diags, resourceTypeRead(ctx, d, meta)...)
}

func resourceTypeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	input := &appsync.DeleteTypeInput{
		ApiId:    aws.String(d.Get("api_id").(string)),
		TypeName: aws.String(d.Get("name").(string)),
	}
	_, err := conn.DeleteType(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Appsync Type: %s", err)
	}

	return diags
}

func DecodeTypeID(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("Unexpected format of ID (%q), expected API-ID:FORMAT:TYPE-NAME", id)
	}
	return parts[0], parts[1], parts[2], nil
}
