// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplestructfield

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplestructfield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplestructfield/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_examplestructfield_rule", name="Rule")
func resourceRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRuleCreate,
		ReadWithoutTimeout:   resourceRuleRead,
		UpdateWithoutTimeout: resourceRuleUpdate,
		DeleteWithoutTimeout: resourceRuleDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"predicate": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceRuleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleStructFieldClient(ctx)

	name := d.Get("name").(string)
	output, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (any, error) {
		input := examplestructfield.CreateRuleInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateRule(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Rule (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.(*examplestructfield.CreateRuleOutput).Rule.RuleId))

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleStructFieldClient(ctx)

	rule, err := findRuleByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Rule (%s): %s", d.Id(), err)
	}

	d.Set("name", rule.Name)
	d.Set("predicate", rule.Predicate)

	return diags
}

func resourceRuleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleStructFieldClient(ctx)

	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (any, error) {
		input := examplestructfield.UpdateRuleInput{
			ChangeToken: token,
			Predicate:   aws.String(d.Get("predicate").(string)),
			RuleId:      aws.String(d.Id()),
		}

		return conn.UpdateRule(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Rule (%s): %s", d.Id(), err)
	}

	return append(diags, resourceRuleRead(ctx, d, meta)...)
}

func resourceRuleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleStructFieldClient(ctx)

	log.Printf("[INFO] Deleting Rule: %s", d.Id())
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (any, error) {
		input := examplestructfield.DeleteRuleInput{
			ChangeToken: token,
			RuleId:      aws.String(d.Id()),
		}

		return conn.DeleteRule(ctx, &input)
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func findRuleByID(ctx context.Context, conn *examplestructfield.Client, id string) (*awstypes.Rule, error) {
	input := examplestructfield.GetRuleInput{
		RuleId: aws.String(id),
	}

	output, err := conn.GetRule(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Rule == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Rule, nil
}
