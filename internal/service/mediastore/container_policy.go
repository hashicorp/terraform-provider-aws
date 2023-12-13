// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediastore

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_media_store_container_policy")
func ResourceContainerPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContainerPolicyPut,
		ReadWithoutTimeout:   resourceContainerPolicyRead,
		UpdateWithoutTimeout: resourceContainerPolicyPut,
		DeleteWithoutTimeout: resourceContainerPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"container_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     verify.ValidIAMPolicyJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceContainerPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaStoreConn(ctx)

	name := d.Get("container_name").(string)
	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting MediaStore Container Policy (%s): %s", name, err)
	}

	input := &mediastore.PutContainerPolicyInput{
		ContainerName: aws.String(name),
		Policy:        aws.String(policy),
	}

	_, err = conn.PutContainerPolicyWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting MediaStore Container Policy (%s): %s", name, err)
	}

	d.SetId(name)
	return append(diags, resourceContainerPolicyRead(ctx, d, meta)...)
}

func resourceContainerPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaStoreConn(ctx)

	input := &mediastore.GetContainerPolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	resp, err := conn.GetContainerPolicyWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, mediastore.ErrCodeContainerNotFoundException) {
			log.Printf("[WARN] MediaStore Container Policy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		if tfawserr.ErrCodeEquals(err, mediastore.ErrCodePolicyNotFoundException) {
			log.Printf("[WARN] MediaStore Container Policy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading MediaStore Container Policy (%s): %s", d.Id(), err)
	}

	d.Set("container_name", d.Id())

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.StringValue(resp.Policy))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MediaStore Container Policy (%s): %s", d.Id(), err)
	}

	d.Set("policy", policyToSet)

	return diags
}

func resourceContainerPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaStoreConn(ctx)

	input := &mediastore.DeleteContainerPolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	_, err := conn.DeleteContainerPolicyWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, mediastore.ErrCodeContainerNotFoundException) {
			return diags
		}
		if tfawserr.ErrCodeEquals(err, mediastore.ErrCodePolicyNotFoundException) {
			return diags
		}
		// if isAWSErr(err, mediastore.ErrCodeContainerInUseException, "Container must be ACTIVE in order to perform this operation") {
		// 	return nil
		// }
		return sdkdiag.AppendErrorf(diags, "deleting MediaStore Container Policy (%s): %s", d.Id(), err)
	}

	return diags
}
