// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_msk_cluster_policy", name="Cluster Policy")
func resourceClusterPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterPolicyPut,
		ReadWithoutTimeout:   resourceClusterPolicyRead,
		UpdateWithoutTimeout: resourceClusterPolicyPut,
		DeleteWithoutTimeout: resourceClusterPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"cluster_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"current_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceClusterPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	clusterARN := d.Get("cluster_arn").(string)
	input := &kafka.PutClusterPolicyInput{
		ClusterArn: aws.String(clusterARN),
		Policy:     aws.String(policy),
	}

	if !d.IsNewResource() {
		input.CurrentVersion = aws.String(d.Get("current_version").(string))
	}

	_, err = conn.PutClusterPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting MSK Cluster Policy (%s): %s", clusterARN, err)
	}

	if d.IsNewResource() {
		d.SetId(clusterARN)
	}

	return append(diags, resourceClusterPolicyRead(ctx, d, meta)...)
}

func resourceClusterPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	output, err := findClusterPolicyByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MSK Cluster Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Cluster Policy (%s): %s", d.Id(), err)
	}

	d.Set("cluster_arn", d.Id())
	d.Set("current_version", output.CurrentVersion)
	if output.Policy != nil {
		policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(output.Policy))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set(names.AttrPolicy, policyToSet)
	} else {
		d.Set(names.AttrPolicy, nil)
	}

	return diags
}

func resourceClusterPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	log.Printf("[INFO] Deleting MSK Cluster Policy: %s", d.Id())
	_, err := conn.DeleteClusterPolicy(ctx, &kafka.DeleteClusterPolicyInput{
		ClusterArn: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MSK Cluster Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findClusterPolicyByARN(ctx context.Context, conn *kafka.Client, id string) (*kafka.GetClusterPolicyOutput, error) {
	in := &kafka.GetClusterPolicyInput{
		ClusterArn: aws.String(id),
	}

	out, err := conn.GetClusterPolicy(ctx, in)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
