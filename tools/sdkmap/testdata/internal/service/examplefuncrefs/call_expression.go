// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplefuncrefs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplefuncrefs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplefuncrefs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_examplefuncrefs_call_expression", name="Call Expression")
func resourceCallExpression() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCallExpressionPut(awstypes.ExistConditionMustNotExist),
		ReadWithoutTimeout:   resourceCallExpressionRead,
		UpdateWithoutTimeout: resourceCallExpressionPut(awstypes.ExistConditionMustExist),
		DeleteWithoutTimeout: resourceCallExpressionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"configuration": {
				Type:     schema.TypeString,
				Required: true,
			},
			"foo_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceCallExpressionPut(condition awstypes.ExistCondition) func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		var diags diag.Diagnostics
		conn := meta.(*conns.AWSClient).ExampleFuncRefsClient(ctx)

		fooID := d.Get("foo_id").(string)
		input := examplefuncrefs.PutFooConfigurationInput{
			Configuration:  aws.String(d.Get("configuration").(string)),
			ExistCondition: condition,
			FooId:          aws.String(fooID),
		}

		_, err := conn.PutFooConfiguration(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting Foo (%s) configuration: %s", fooID, err)
		}

		d.SetId(fooID)

		return append(diags, resourceCallExpressionRead(ctx, d, meta)...)
	}
}

func resourceCallExpressionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleFuncRefsClient(ctx)

	configuration, err := findFooConfigurationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Foo (%s) configuration not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Foo (%s) configuration: %s", d.Id(), err)
	}

	d.Set("configuration", configuration)
	d.Set("foo_id", d.Id())

	return diags
}

func resourceCallExpressionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleFuncRefsClient(ctx)

	log.Printf("[DEBUG] Deleting Foo (%s) configuration", d.Id())
	input := examplefuncrefs.DeleteFooConfigurationInput{
		FooId: aws.String(d.Id()),
	}
	_, err := conn.DeleteFooConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Foo (%s) configuration: %s", d.Id(), err)
	}

	return diags
}

func findFooConfigurationByID(ctx context.Context, conn *examplefuncrefs.Client, fooID string) (string, error) {
	input := examplefuncrefs.GetFooConfigurationInput{
		FooId: aws.String(fooID),
	}

	output, err := conn.GetFooConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return "", &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return "", err
	}

	if output == nil || output.Configuration == nil {
		return "", tfresource.NewEmptyResultError()
	}

	return aws.ToString(output.Configuration), nil
}
