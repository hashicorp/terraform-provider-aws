// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockinterface

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockinterface"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockinterface/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_mockinterface_foo", name="Foo")
func resourceFoo() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFooCreate,
		ReadWithoutTimeout:   resourceFooRead,
		DeleteWithoutTimeout: resourceFooDelete,

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

func resourceFooCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockInterfaceClient(ctx)

	name := d.Get("name").(string)
	var creator fooCreator
	switch fooType := awstypes.FooType(d.Get("type").(string)); fooType {
	case awstypes.FooTypeBasic:
		creator = basicFooCreator{}

	case awstypes.FooTypeAdvanced:
		creator = advancedFooCreator{}
	}

	if err := creator.Create(ctx, conn, name, d); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating %s Foo (%s): %s", creator.TypeName(), name, err)
	}

	return append(diags, resourceFooRead(ctx, d, meta)...)
}

func resourceFooRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockInterfaceClient(ctx)

	foo, err := findFooByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Foo (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Foo (%s): %s", d.Id(), err)
	}

	d.Set("name", foo.Name)
	d.Set("type", foo.Type)

	return diags
}

func resourceFooDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockInterfaceClient(ctx)

	log.Printf("[DEBUG] Deleting Foo: %s", d.Id())
	input := mockinterface.DeleteFooInput{
		FooId: aws.String(d.Id()),
	}
	_, err := conn.DeleteFoo(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Foo (%s): %s", d.Id(), err)
	}

	return diags
}

type fooCreator interface {
	TypeName() string
	Create(ctx context.Context, conn *mockinterface.Client, name string, d *schema.ResourceData) error
}

type basicFooCreator struct{}

func (c basicFooCreator) TypeName() string {
	return "Basic"
}

func (c basicFooCreator) Create(ctx context.Context, conn *mockinterface.Client, name string, d *schema.ResourceData) error {
	input := mockinterface.CreateBasicFooInput{
		Name: aws.String(name),
	}

	output, err := conn.CreateBasicFoo(ctx, &input)

	if err != nil {
		return err
	}

	d.SetId(aws.ToString(output.FooId))

	return nil
}

type advancedFooCreator struct{}

func (c advancedFooCreator) TypeName() string {
	return "Advanced"
}

func (c advancedFooCreator) Create(ctx context.Context, conn *mockinterface.Client, name string, d *schema.ResourceData) error {
	input := mockinterface.CreateAdvancedFooInput{
		Name: aws.String(name),
	}

	output, err := conn.CreateAdvancedFoo(ctx, &input)

	if err != nil {
		return err
	}

	d.SetId(aws.ToString(output.FooId))

	return nil
}

// fooDescriber is implemented by *mockinterface.Client, not by any type in this package.
type fooDescriber interface {
	DescribeFoo(ctx context.Context, params *mockinterface.DescribeFooInput, optFns ...func(*mockinterface.Options)) (*mockinterface.DescribeFooOutput, error)
}

func findFooByID(ctx context.Context, conn fooDescriber, id string) (*awstypes.Foo, error) {
	input := mockinterface.DescribeFooInput{
		FooId: aws.String(id),
	}

	output, err := conn.DescribeFoo(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Foo == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Foo, nil
}
