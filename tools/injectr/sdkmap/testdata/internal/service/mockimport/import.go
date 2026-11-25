// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockimport

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKResource("aws_mockimport_passthrough", name="Passthrough")
func resourcePassthrough() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFooCreate,
		ReadWithoutTimeout:   resourceFooRead,
		DeleteWithoutTimeout: resourceFooDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: fooSchema,
	}
}

// @SDKResource("aws_mockimport_function", name="Function")
func resourceFunction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFooCreate,
		ReadWithoutTimeout:   resourceFooRead,
		DeleteWithoutTimeout: resourceFooDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceFunctionImport,
		},

		SchemaFunc: fooSchema,
	}
}

// @SDKResource("aws_mockimport_function_literal", name="Function Literal")
func resourceFunctionLiteral() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFooCreate,
		ReadWithoutTimeout:   resourceFooRead,
		DeleteWithoutTimeout: resourceFooDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				conn := meta.(*conns.AWSClient).MockImportClient(ctx)

				foo, err := findFooByName(ctx, conn, d.Id())

				if err != nil {
					return nil, err
				}

				d.SetId(*foo.FooId)

				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaFunc: fooSchema,
	}
}

// @SDKResource("aws_mockimport_none", name="None")
func resourceNone() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFooCreate,
		ReadWithoutTimeout:   resourceFooRead,
		DeleteWithoutTimeout: resourceFooDelete,

		SchemaFunc: fooSchema,
	}
}

func resourceFunctionImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	parentID, fooID, found := strings.Cut(d.Id(), "/")
	if !found || parentID == "" || fooID == "" {
		return nil, fmt.Errorf("unexpected format for ID (%[1]s), expected PARENT_ID/THING_ID", d.Id())
	}

	d.SetId(fooID)
	d.Set("parent_id", parentID)

	return []*schema.ResourceData{d}, nil
}
