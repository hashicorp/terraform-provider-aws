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
			"thing_id": {
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

		thingID := d.Get("thing_id").(string)
		input := examplefuncrefs.PutThingConfigurationInput{
			Configuration:  aws.String(d.Get("configuration").(string)),
			ExistCondition: condition,
			ThingId:        aws.String(thingID),
		}

		_, err := conn.PutThingConfiguration(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting Thing (%s) configuration: %s", thingID, err)
		}

		d.SetId(thingID)

		return append(diags, resourceCallExpressionRead(ctx, d, meta)...)
	}
}

func resourceCallExpressionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleFuncRefsClient(ctx)

	configuration, err := findThingConfigurationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Thing (%s) configuration not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Thing (%s) configuration: %s", d.Id(), err)
	}

	d.Set("configuration", configuration)
	d.Set("thing_id", d.Id())

	return diags
}

func resourceCallExpressionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleFuncRefsClient(ctx)

	log.Printf("[DEBUG] Deleting Thing (%s) configuration", d.Id())
	input := examplefuncrefs.DeleteThingConfigurationInput{
		ThingId: aws.String(d.Id()),
	}
	_, err := conn.DeleteThingConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Thing (%s) configuration: %s", d.Id(), err)
	}

	return diags
}

func findThingConfigurationByID(ctx context.Context, conn *examplefuncrefs.Client, thingID string) (string, error) {
	input := examplefuncrefs.GetThingConfigurationInput{
		ThingId: aws.String(thingID),
	}

	output, err := conn.GetThingConfiguration(ctx, &input)

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
