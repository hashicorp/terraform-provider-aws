// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oam

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/oam"
	"github.com/aws/aws-sdk-go-v2/service/oam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_oam_sink_policy")
func ResourceSinkPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSinkPolicyPut,
		ReadWithoutTimeout:   resourceSinkPolicyRead,
		UpdateWithoutTimeout: resourceSinkPolicyPut,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
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

const (
	ResNameSinkPolicy = "Sink Policy"
)

func resourceSinkPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	sinkIdentifier := d.Get("sink_identifier").(string)
	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get(names.AttrPolicy).(string), err)
	}

	in := &oam.PutSinkPolicyInput{
		SinkIdentifier: aws.String(sinkIdentifier),
		Policy:         aws.String(policy),
	}

	_, err = conn.PutSinkPolicy(ctx, in)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting ObservabilityAccessManager Sink Policy (%s): %s", sinkIdentifier, err)
	}

	if d.IsNewResource() {
		d.SetId(sinkIdentifier)
	}

	return append(diags, resourceSinkPolicyRead(ctx, d, meta)...)
}

func resourceSinkPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	out, err := findSinkPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ObservabilityAccessManager SinkPolicy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionReading, ResNameSinkPolicy, d.Id(), err)
	}

	d.Set(names.AttrARN, out.SinkArn)
	d.Set("sink_id", out.SinkId)
	d.Set("sink_identifier", d.Id())

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(out.Policy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return nil
}

func findSinkPolicyByID(ctx context.Context, conn *oam.Client, id string) (*oam.GetSinkPolicyOutput, error) {
	in := &oam.GetSinkPolicyInput{
		SinkIdentifier: aws.String(id),
	}
	out, err := conn.GetSinkPolicy(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
