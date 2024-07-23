// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_endpoint_policy", "VPC Endpoint Policy")
func resourceVPCEndpointPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointPolicyPut,
		UpdateWithoutTimeout: resourceVPCEndpointPolicyPut,
		ReadWithoutTimeout:   resourceVPCEndpointPolicyRead,
		DeleteWithoutTimeout: resourceVPCEndpointPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Optional:              true,
				Computed:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			names.AttrVPCEndpointID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceVPCEndpointPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID := d.Get(names.AttrVPCEndpointID).(string)
	req := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId: aws.String(endpointID),
	}

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy contains an invalid JSON: %s", err)
	}

	if policy == "" {
		req.ResetPolicy = aws.Bool(true)
	} else {
		req.PolicyDocument = aws.String(policy)
	}

	log.Printf("[DEBUG] Updating VPC Endpoint Policy: %#v", req)
	if _, err := conn.ModifyVpcEndpoint(ctx, req); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating VPC Endpoint Policy: %s", err)
	}
	d.SetId(endpointID)

	_, err = waitVPCEndpointAvailable(ctx, conn, endpointID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Endpoint (%s) to policy to set: %s", endpointID, err)
	}

	return append(diags, resourceVPCEndpointPolicyRead(ctx, d, meta)...)
}

func resourceVPCEndpointPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpce, err := findVPCEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPC Endpoint Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrVPCEndpointID, d.Id())

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get(names.AttrPolicy).(string), aws.ToString(vpce.PolicyDocument))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policyToSet, err)
	}

	d.Set(names.AttrPolicy, policyToSet)
	return diags
}

func resourceVPCEndpointPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	req := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId: aws.String(d.Id()),
		ResetPolicy:   aws.Bool(true),
	}

	log.Printf("[DEBUG] Resetting VPC Endpoint Policy: %#v", req)
	if _, err := conn.ModifyVpcEndpoint(ctx, req); err != nil {
		if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointIdNotFound) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "Resetting VPC Endpoint Policy: %s", err)
	}

	_, err := waitVPCEndpointAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Endpoint (%s) to be reset: %s", d.Id(), err)
	}

	return diags
}
