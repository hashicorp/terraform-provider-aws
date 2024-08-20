// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemas

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/schemas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/schemas/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_schemas_registry", name="Registry")
// @Tags(identifierAttribute="arn")
func resourceRegistry() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegistryCreate,
		ReadWithoutTimeout:   resourceRegistryRead,
		UpdateWithoutTimeout: resourceRegistryUpdate,
		DeleteWithoutTimeout: resourceRegistryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), ""),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRegistryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &schemas.CreateRegistryInput{
		RegistryName: aws.String(name),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.CreateRegistry(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Schemas Registry (%s): %s", name, err)
	}

	d.SetId(aws.ToString(input.RegistryName))

	return append(diags, resourceRegistryRead(ctx, d, meta)...)
}

func resourceRegistryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	output, err := findRegistryByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Schemas Registry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Schemas Registry (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.RegistryArn)
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrName, output.RegistryName)

	return diags
}

func resourceRegistryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	if d.HasChanges(names.AttrDescription) {
		input := &schemas.UpdateRegistryInput{
			Description:  aws.String(d.Get(names.AttrDescription).(string)),
			RegistryName: aws.String(d.Id()),
		}

		_, err := conn.UpdateRegistry(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EventBridge Schemas Registry (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRegistryRead(ctx, d, meta)...)
}

func resourceRegistryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	log.Printf("[INFO] Deleting EventBridge Schemas Registry (%s)", d.Id())
	_, err := conn.DeleteRegistry(ctx, &schemas.DeleteRegistryInput{
		RegistryName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Schemas Registry (%s): %s", d.Id(), err)
	}

	return diags
}

func findRegistryByName(ctx context.Context, conn *schemas.Client, name string) (*schemas.DescribeRegistryOutput, error) {
	input := &schemas.DescribeRegistryInput{
		RegistryName: aws.String(name),
	}

	output, err := conn.DescribeRegistry(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
