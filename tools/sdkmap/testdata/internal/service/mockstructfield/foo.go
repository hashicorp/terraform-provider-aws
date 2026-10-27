// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockstructfield

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockstructfield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockstructfield/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_mockstructfield_foo", name="Foo")
func resourceFoo() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFooCreate,
		ReadWithoutTimeout:   resourceFooRead,
		UpdateWithoutTimeout: resourceFooUpdate,
		DeleteWithoutTimeout: resourceFooDelete,

		Schema: map[string]*schema.Schema{
			"engine_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"parent_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceFooCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.MockStructFieldClient(ctx)

	if v, ok := d.GetOk("parent_id"); ok {
		if _, err := newParentLookup(c.MockCrossServiceClient(ctx)).Find(ctx, v.(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Foo (%s): %s", v, err)
		}
	}

	name := d.Get("name").(string)
	input := mockstructfield.CreateFooInput{
		EngineVersion: aws.String(d.Get("engine_version").(string)),
		Name:          aws.String(name),
	}

	output, err := conn.CreateFoo(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Foo (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Foo.FooId))

	return append(diags, resourceFooRead(ctx, d, meta)...)
}

func resourceFooRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockStructFieldClient(ctx)

	foo, err := findFooByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Foo (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Foo (%s): %s", d.Id(), err)
	}

	d.Set("engine_version", foo.EngineVersion)
	d.Set("name", foo.Name)

	return diags
}

func resourceFooUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockStructFieldClient(ctx)

	if d.HasChange("engine_version") {
		orchestrator := newDeploymentOrchestrator(conn)
		defer orchestrator.CleanUp(ctx)

		input := mockstructfield.CreateDeploymentInput{
			FooId:               aws.String(d.Id()),
			TargetEngineVersion: aws.String(d.Get("engine_version").(string)),
		}

		deployment, err := orchestrator.CreateDeployment(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Foo (%s): %s", d.Id(), err)
		}

		if err := orchestrator.Switchover(ctx, aws.ToString(deployment.DeploymentId)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Foo (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceFooRead(ctx, d, meta)...)
}

func resourceFooDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockStructFieldClient(ctx)

	log.Printf("[DEBUG] Deleting Foo: %s", d.Id())
	input := mockstructfield.DeleteFooInput{
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

func findFooByID(ctx context.Context, conn *mockstructfield.Client, id string) (*awstypes.Foo, error) {
	input := mockstructfield.DescribeFooInput{
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
