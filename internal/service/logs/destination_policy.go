// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_cloudwatch_log_destination_policy", name="Destination Policy")
func resourceDestinationPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDestinationPolicyPut,
		ReadWithoutTimeout:   resourceDestinationPolicyRead,
		UpdateWithoutTimeout: resourceDestinationPolicyPut,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"access_policy": sdkv2.IAMPolicyDocumentSchemaRequired(),
			"destination_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"force_update": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceDestinationPolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("access_policy").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	name := d.Get("destination_name").(string)
	input := &cloudwatchlogs.PutDestinationPolicyInput{
		AccessPolicy:    aws.String(policy),
		DestinationName: aws.String(name),
	}

	if v, ok := d.GetOk("force_update"); ok {
		input.ForceUpdate = aws.Bool(v.(bool))
	}

	_, err = conn.PutDestinationPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting CloudWatch Logs Destination Policy (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return append(diags, resourceDestinationPolicyRead(ctx, d, meta)...)
}

func resourceDestinationPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	destination, err := findDestinationPolicyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Destination Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Destination Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("access_policy").(string), aws.ToString(destination.AccessPolicy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("access_policy", policyToSet)
	d.Set("destination_name", destination.DestinationName)

	return diags
}

func findDestinationPolicyByName(ctx context.Context, conn *cloudwatchlogs.Client, name string) (*awstypes.Destination, error) {
	output, err := findDestinationByName(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if output.AccessPolicy == nil {
		return nil, tfresource.NewEmptyResultError(name)
	}

	return output, err
}
