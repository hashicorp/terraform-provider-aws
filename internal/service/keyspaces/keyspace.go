// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package keyspaces

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/keyspaces"
	"github.com/aws/aws-sdk-go-v2/service/keyspaces/types"
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

// @SDKResource("aws_keyspaces_keyspace", name="Keyspace")
// @Tags(identifierAttribute="arn")
func resourceKeyspace() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeyspaceCreate,
		ReadWithoutTimeout:   resourceKeyspaceRead,
		UpdateWithoutTimeout: resourceKeyspaceUpdate,
		DeleteWithoutTimeout: resourceKeyspaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexache.MustCompile(`^[0-9A-Za-z][0-9A-Za-z_]{0,47}$`),
					"The name can have up to 48 characters. It must begin with an alpha-numeric character and can only contain alpha-numeric characters and underscores.",
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceKeyspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KeyspacesClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &keyspaces.CreateKeyspaceInput{
		KeyspaceName: aws.String(name),
		Tags:         getTagsIn(ctx),
	}

	_, err := conn.CreateKeyspace(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Keyspaces Keyspace (%s): %s", name, err)
	}

	d.SetId(name)

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return findKeyspaceByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Keyspaces Keyspace (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceKeyspaceRead(ctx, d, meta)...)
}

func resourceKeyspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KeyspacesClient(ctx)

	keyspace, err := findKeyspaceByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Keyspaces Keyspace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Keyspaces Keyspace (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, keyspace.ResourceArn)
	d.Set(names.AttrName, keyspace.KeyspaceName)

	return diags
}

func resourceKeyspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceKeyspaceRead(ctx, d, meta)
}

func resourceKeyspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KeyspacesClient(ctx)

	log.Printf("[DEBUG] Deleting Keyspaces Keyspace: (%s)", d.Id())
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*types.ConflictException](ctx, d.Timeout(schema.TimeoutDelete),
		func() (interface{}, error) {
			return conn.DeleteKeyspace(ctx, &keyspaces.DeleteKeyspaceInput{
				KeyspaceName: aws.String(d.Id()),
			})
		},
		"a table under it is currently being created or deleted")

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Keyspaces Keyspace (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return findKeyspaceByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Keyspaces Keyspace (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findKeyspaceByName(ctx context.Context, conn *keyspaces.Client, name string) (*keyspaces.GetKeyspaceOutput, error) {
	input := keyspaces.GetKeyspaceInput{
		KeyspaceName: aws.String(name),
	}

	output, err := conn.GetKeyspace(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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
