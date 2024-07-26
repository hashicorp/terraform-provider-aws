// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediastore

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediastore"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mediastore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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
			names.AttrPolicy: {
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
	conn := meta.(*conns.AWSClient).MediaStoreClient(ctx)

	name := d.Get("container_name").(string)
	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting MediaStore Container Policy (%s): %s", name, err)
	}

	input := &mediastore.PutContainerPolicyInput{
		ContainerName: aws.String(name),
		Policy:        aws.String(policy),
	}

	_, err = conn.PutContainerPolicy(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting MediaStore Container Policy (%s): %s", name, err)
	}

	d.SetId(name)
	return append(diags, resourceContainerPolicyRead(ctx, d, meta)...)
}

func resourceContainerPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaStoreClient(ctx)

	resp, err := findContainerPolicyByContainerName(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		log.Printf("[WARN] MediaStore Container Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MediaStore Container Policy (%s): %s", d.Id(), err)
	}

	d.Set("container_name", d.Id())

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(resp.Policy))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MediaStore Container Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return diags
}

func resourceContainerPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaStoreClient(ctx)

	input := &mediastore.DeleteContainerPolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	_, err := conn.DeleteContainerPolicy(ctx, input)

	if errs.IsA[*awstypes.ContainerNotFoundException](err) || errs.IsA[*awstypes.PolicyNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MediaStore Container Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findContainerPolicyByContainerName(ctx context.Context, conn *mediastore.Client, id string) (*mediastore.GetContainerPolicyOutput, error) {
	input := &mediastore.GetContainerPolicyInput{
		ContainerName: aws.String(id),
	}

	output, err := conn.GetContainerPolicy(ctx, input)

	if errs.IsA[*awstypes.ContainerNotFoundException](err) || errs.IsA[*awstypes.PolicyNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil || output.Policy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
