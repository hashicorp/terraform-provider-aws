// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_elb_attachment", name="Attachment")
func resourceAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAttachmentCreate,
		ReadWithoutTimeout:   resourceAttachmentRead,
		DeleteWithoutTimeout: resourceAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"elb": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"instance": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName := d.Get("elb").(string)
	instance := d.Get("instance").(string)
	input := &elasticloadbalancing.RegisterInstancesWithLoadBalancerInput{
		Instances:        expandInstances([]interface{}{instance}),
		LoadBalancerName: aws.String(lbName),
	}

	const (
		timeout = 10 * time.Minute
	)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (interface{}, error) {
		return conn.RegisterInstancesWithLoadBalancer(ctx, input)
	}, errCodeInvalidTarget)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELB Classic Attachment (%s/%s): %s", lbName, instance, err)
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(id.PrefixedUniqueId(fmt.Sprintf("%s-", lbName)))

	return diags
}

func resourceAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName := d.Get("elb").(string)
	instance := d.Get("instance").(string)
	err := findLoadBalancerAttachmentByTwoPartKey(ctx, conn, lbName, instance)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELB Classic Attachment (%s/%s) not found, removing from state", lbName, instance)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Attachment (%s/%s): %s", lbName, instance, err)
	}

	return diags
}

func resourceAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	lbName := d.Get("elb").(string)
	instance := d.Get("instance").(string)
	input := &elasticloadbalancing.DeregisterInstancesFromLoadBalancerInput{
		Instances:        expandInstances([]interface{}{instance}),
		LoadBalancerName: aws.String(lbName),
	}

	log.Printf("[DEBUG] Deleting ELB Classic Attachment: %s", d.Id())
	_, err := conn.DeregisterInstancesFromLoadBalancer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELB Classic Attachment (%s/%s): %s", lbName, instance, err)
	}

	return diags
}

func findLoadBalancerAttachmentByTwoPartKey(ctx context.Context, conn *elasticloadbalancing.Client, lbName, instance string) error {
	lb, err := findLoadBalancerByName(ctx, conn, lbName)

	if err != nil {
		return err
	}

	attached := slices.ContainsFunc(lb.Instances, func(v awstypes.Instance) bool {
		return aws.ToString(v.InstanceId) == instance
	})

	if !attached {
		return &retry.NotFoundError{}
	}

	return nil
}
