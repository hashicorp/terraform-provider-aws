// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package exampleimport

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKResource("aws_exampleimport_passthrough", name="Passthrough")
func resourcePassthrough() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		DeleteWithoutTimeout: resourceThingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: thingSchema,
	}
}

// @SDKResource("aws_exampleimport_function", name="Function")
func resourceFunction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		DeleteWithoutTimeout: resourceThingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceFunctionImport,
		},

		SchemaFunc: thingSchema,
	}
}

// @SDKResource("aws_exampleimport_function_literal", name="Function Literal")
func resourceFunctionLiteral() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		DeleteWithoutTimeout: resourceThingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				conn := meta.(*conns.AWSClient).ExampleImportClient(ctx)

				thing, err := findThingByName(ctx, conn, d.Id())

				if err != nil {
					return nil, err
				}

				d.SetId(*thing.ThingId)

				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaFunc: thingSchema,
	}
}

// @SDKResource("aws_exampleimport_none", name="None")
func resourceNone() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		DeleteWithoutTimeout: resourceThingDelete,

		SchemaFunc: thingSchema,
	}
}

func resourceFunctionImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	parentID, thingID, found := strings.Cut(d.Id(), "/")
	if !found || parentID == "" || thingID == "" {
		return nil, fmt.Errorf("unexpected format for ID (%[1]s), expected PARENT_ID/THING_ID", d.Id())
	}

	d.SetId(thingID)
	d.Set("parent_id", parentID)

	return []*schema.ResourceData{d}, nil
}
