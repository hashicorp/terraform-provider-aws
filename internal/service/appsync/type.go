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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appsync_type", name="Type")
func resourceType() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"definition": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrFormat: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TypeDefinitionFormat](),
			},
			names.AttrName: {
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
	input := &appsync.CreateTypeInput{
		ApiId:      aws.String(apiID),
		Definition: aws.String(d.Get("definition").(string)),
		Format:     awstypes.TypeDefinitionFormat(d.Get(names.AttrFormat).(string)),
	}

	output, err := conn.CreateType(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Appsync Type: %s", err)
	}

	d.SetId(typeCreateResourceID(apiID, output.Type.Format, aws.ToString(output.Type.Name)))

	return append(diags, resourceTypeRead(ctx, d, meta)...)
}

func resourceTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, format, name, err := typeParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resp, err := findTypeByThreePartKey(ctx, conn, apiID, format, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppSync Type (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appsync Type %q: %s", d.Id(), err)
	}

	d.Set("api_id", apiID)
	d.Set(names.AttrARN, resp.Arn)
	d.Set(names.AttrName, resp.Name)
	d.Set(names.AttrFormat, resp.Format)
	d.Set("definition", resp.Definition)
	d.Set(names.AttrDescription, resp.Description)

	return diags
}

func resourceTypeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, format, name, err := typeParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &appsync.UpdateTypeInput{
		ApiId:      aws.String(apiID),
		Definition: aws.String(d.Get("definition").(string)),
		Format:     format,
		TypeName:   aws.String(name),
	}

	_, err = conn.UpdateType(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Appsync Type (%s): %s", d.Id(), err)
	}

	return append(diags, resourceTypeRead(ctx, d, meta)...)
}

func resourceTypeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, _, name, err := typeParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting Appsync Type: %s", d.Id())
	_, err = conn.DeleteType(ctx, &appsync.DeleteTypeInput{
		ApiId:    aws.String(apiID),
		TypeName: aws.String(name),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appsync Type (%s): %s", d.Id(), err)
	}

	return diags
}

const typeResourceIDSeparator = ":"

func typeCreateResourceID(apiID string, format awstypes.TypeDefinitionFormat, name string) string {
	parts := []string{apiID, string(format), name} // nosemgrep:ci.typed-enum-conversion
	id := strings.Join(parts, typeResourceIDSeparator)

	return id
}

func typeParseResourceID(id string) (string, awstypes.TypeDefinitionFormat, string, error) {
	parts := strings.Split(id, typeResourceIDSeparator)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected API-ID%[2]sFORMAT%[2]sTYPE-NAME", id, typeResourceIDSeparator)
	}

	return parts[0], awstypes.TypeDefinitionFormat(parts[1]), parts[2], nil
}

func findTypeByThreePartKey(ctx context.Context, conn *appsync.Client, apiID string, format awstypes.TypeDefinitionFormat, name string) (*awstypes.Type, error) {
	input := &appsync.GetTypeInput{
		ApiId:    aws.String(apiID),
		Format:   format,
		TypeName: aws.String(name),
	}

	output, err := conn.GetType(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Type == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Type, nil
}
