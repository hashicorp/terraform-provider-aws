// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplefuncrefs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/examplefuncrefs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplefuncrefs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_examplefuncrefs_noop", name="Noop Delete")
func resourceNoopDelete() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNoopDeleteUpsert,
		ReadWithoutTimeout:   resourceNoopDeleteRead,
		UpdateWithoutTimeout: resourceNoopDeleteUpsert,
		DeleteWithoutTimeout: schema.NoopContext,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"default_foo_mode": {
					Type:     schema.TypeString,
					Required: true,
				},
			}
		},
	}
}

func resourceNoopDeleteUpsert(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleFuncRefsClient(ctx)

	input := examplefuncrefs.SetAccountSettingsInput{
		DefaultFooMode: awstypes.FooMode(d.Get("default_foo_mode").(string)),
	}

	_, err := conn.SetAccountSettings(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting account settings: %s", err)
	}

	if d.IsNewResource() {
		d.SetId(meta.(*conns.AWSClient).AccountID(ctx))
	}

	return append(diags, resourceNoopDeleteRead(ctx, d, meta)...)
}

func resourceNoopDeleteRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleFuncRefsClient(ctx)

	output, err := conn.GetAccountSettings(ctx, &examplefuncrefs.GetAccountSettingsInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading account settings: %s", err)
	}

	d.Set("default_foo_mode", output.DefaultFooMode)

	return diags
}
