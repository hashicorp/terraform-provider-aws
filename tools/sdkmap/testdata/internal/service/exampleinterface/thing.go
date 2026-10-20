// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package exampleinterface

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/exampleinterface"
	awstypes "github.com/aws/aws-sdk-go-v2/service/exampleinterface/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_exampleinterface_thing", name="Thing")
func resourceThing() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		DeleteWithoutTimeout: resourceThingDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceThingCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleInterfaceClient(ctx)

	name := d.Get("name").(string)
	var creator thingCreator
	switch thingType := awstypes.ThingType(d.Get("type").(string)); thingType {
	case awstypes.ThingTypeBasic:
		creator = basicThingCreator{}

	case awstypes.ThingTypeAdvanced:
		creator = advancedThingCreator{}
	}

	if err := creator.Create(ctx, conn, name, d); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating %s Thing (%s): %s", creator.TypeName(), name, err)
	}

	return append(diags, resourceThingRead(ctx, d, meta)...)
}

func resourceThingRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleInterfaceClient(ctx)

	thing, err := findThingByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Thing (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Thing (%s): %s", d.Id(), err)
	}

	d.Set("name", thing.Name)
	d.Set("type", thing.Type)

	return diags
}

func resourceThingDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleInterfaceClient(ctx)

	log.Printf("[DEBUG] Deleting Thing: %s", d.Id())
	input := exampleinterface.DeleteThingInput{
		ThingId: aws.String(d.Id()),
	}
	_, err := conn.DeleteThing(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Thing (%s): %s", d.Id(), err)
	}

	return diags
}

type thingCreator interface {
	TypeName() string
	Create(ctx context.Context, conn *exampleinterface.Client, name string, d *schema.ResourceData) error
}

type basicThingCreator struct{}

func (c basicThingCreator) TypeName() string {
	return "Basic"
}

func (c basicThingCreator) Create(ctx context.Context, conn *exampleinterface.Client, name string, d *schema.ResourceData) error {
	input := exampleinterface.CreateBasicThingInput{
		Name: aws.String(name),
	}

	output, err := conn.CreateBasicThing(ctx, &input)

	if err != nil {
		return err
	}

	d.SetId(aws.ToString(output.ThingId))

	return nil
}

type advancedThingCreator struct{}

func (c advancedThingCreator) TypeName() string {
	return "Advanced"
}

func (c advancedThingCreator) Create(ctx context.Context, conn *exampleinterface.Client, name string, d *schema.ResourceData) error {
	input := exampleinterface.CreateAdvancedThingInput{
		Name: aws.String(name),
	}

	output, err := conn.CreateAdvancedThing(ctx, &input)

	if err != nil {
		return err
	}

	d.SetId(aws.ToString(output.ThingId))

	return nil
}

// thingDescriber is implemented by *exampleinterface.Client, not by any type in this package.
type thingDescriber interface {
	DescribeThing(ctx context.Context, params *exampleinterface.DescribeThingInput, optFns ...func(*exampleinterface.Options)) (*exampleinterface.DescribeThingOutput, error)
}

func findThingByID(ctx context.Context, conn thingDescriber, id string) (*awstypes.Thing, error) {
	input := exampleinterface.DescribeThingInput{
		ThingId: aws.String(id),
	}

	output, err := conn.DescribeThing(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Thing == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Thing, nil
}
