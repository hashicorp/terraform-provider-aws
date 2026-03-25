// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package oam

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/oam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/oam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_oam_sink_policy", name="Sink Policy")
func resourceSinkPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSinkPolicyPut,
		ReadWithoutTimeout:   resourceSinkPolicyRead,
		UpdateWithoutTimeout: resourceSinkPolicyPut,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			// These aren't used but are retained for backwards compatibility.
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPolicy: sdkv2.JSONDocumentSchemaRequired(),
			"sink_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sink_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSinkPolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	sinkIdentifier := d.Get("sink_identifier").(string)
	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	in := oam.PutSinkPolicyInput{
		Policy:         aws.String(policy),
		SinkIdentifier: aws.String(sinkIdentifier),
	}

	_, err = conn.PutSinkPolicy(ctx, &in)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting ObservabilityAccessManager Sink Policy (%s): %s", sinkIdentifier, err)
	}

	if d.IsNewResource() {
		d.SetId(sinkIdentifier)
	}

	return append(diags, resourceSinkPolicyRead(ctx, d, meta)...)
}

func resourceSinkPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	out, err := findSinkPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] ObservabilityAccessManager Sink Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ObservabilityAccessManager Sink Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.SinkArn)
	d.Set("sink_id", out.SinkId)
	d.Set("sink_identifier", d.Id())

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(out.Policy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return diags
}

func findSinkPolicyByID(ctx context.Context, conn *oam.Client, id string) (*oam.GetSinkPolicyOutput, error) {
	in := oam.GetSinkPolicyInput{
		SinkIdentifier: aws.String(id),
	}

	return findSinkPolicy(ctx, conn, &in)
}

func findSinkPolicy(ctx context.Context, conn *oam.Client, input *oam.GetSinkPolicyInput) (*oam.GetSinkPolicyOutput, error) {
	output, err := conn.GetSinkPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}
