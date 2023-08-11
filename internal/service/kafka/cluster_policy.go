// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_msk_cluster_policy", name="Cluster Policy")
func ResourceClusterPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterPolicyPut,
		ReadWithoutTimeout:   resourceClusterPolicyRead,
		UpdateWithoutTimeout: resourceClusterPolicyPut,
		DeleteWithoutTimeout: resourceClusterPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cluster_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"current_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

const (
	ResNameClusterPolicy = "Cluster Policy"
)

func resourceClusterPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaClient(ctx)
	clusterArn := d.Get("cluster_arn").(string)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return diag.Errorf("policy (%s) is invalid JSON: %s", policy, err)
	}

	in := &kafka.PutClusterPolicyInput{
		ClusterArn: aws.String(d.Get("cluster_arn").(string)),
		Policy:     aws.String(d.Get("policy").(string)),
	}

	_, err = conn.PutClusterPolicy(ctx, in)

	if err != nil {
		return append(diags, create.DiagError(names.Kafka, create.ErrActionCreating, ResNameClusterPolicy, d.Get("policy").(string), err)...)
	}

	d.SetId(clusterArn)

	return append(diags, resourceClusterPolicyRead(ctx, d, meta)...)
}

func resourceClusterPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	clusterArn := d.Id()

	policy, err := findClusterPolicyByID(ctx, conn, clusterArn)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kafka ClusterPolicy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return append(diags, create.DiagError(names.Kafka, create.ErrActionReading, ResNameClusterPolicy, d.Id(), err)...)
	}

	d.Set("cluster_arn", clusterArn)

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.ToString(policy.Policy))

	if err != nil {
		return diag.Errorf("setting policy %s: %s", aws.ToString(policy.Policy), err)
	}

	d.Set("policy", policyToSet)

	return diags
}

func resourceClusterPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	log.Printf("[INFO] Deleting Kafka ClusterPolicy %s", d.Id())

	_, err := conn.DeleteClusterPolicy(ctx, &kafka.DeleteClusterPolicyInput{
		ClusterArn: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return append(diags, create.DiagError(names.Kafka, create.ErrActionDeleting, ResNameClusterPolicy, d.Id(), err)...)
	}

	return diags
}

func findClusterPolicyByID(ctx context.Context, conn *kafka.Client, id string) (*kafka.GetClusterPolicyOutput, error) {
	in := &kafka.GetClusterPolicyInput{
		ClusterArn: aws.String(id),
	}

	out, err := conn.GetClusterPolicy(ctx, in)

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	return out, nil
}
