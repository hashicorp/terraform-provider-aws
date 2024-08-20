// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_delegation_set", name="Reusable Delegation Set")
func resourceDelegationSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDelegationSetCreate,
		ReadWithoutTimeout:   resourceDelegationSetRead,
		DeleteWithoutTimeout: resourceDelegationSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"reference_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
		},
	}
}

func resourceDelegationSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	callerRef := id.UniqueId()
	if v, ok := d.GetOk("reference_name"); ok {
		callerRef = strings.Join([]string{
			v.(string), "-", callerRef,
		}, "")
	}
	input := &route53.CreateReusableDelegationSetInput{
		CallerReference: aws.String(callerRef),
	}

	output, err := conn.CreateReusableDelegationSet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route 53 Reusable Delegation Set: %s", err)
	}

	d.SetId(cleanDelegationSetID(aws.ToString(output.DelegationSet.Id)))

	return append(diags, resourceDelegationSetRead(ctx, d, meta)...)
}

func resourceDelegationSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	set, err := findDelegationSetByID(ctx, conn, cleanDelegationSetID(d.Id()))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Reusable Delegation Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Reusable Delegation Set (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53",
		Resource:  "delegationset/" + d.Id(),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("name_servers", set.NameServers)

	return diags
}

func resourceDelegationSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	log.Printf("[DEBUG] Deleting Route 53 Reusable Delegation Set: %s", d.Id())
	_, err := conn.DeleteReusableDelegationSet(ctx, &route53.DeleteReusableDelegationSetInput{
		Id: aws.String(cleanDelegationSetID(d.Id())),
	})

	if errs.IsA[*awstypes.NoSuchDelegationSet](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Reusable Delegation Set (%s): %s", d.Id(), err)
	}

	return diags
}

func findDelegationSetByID(ctx context.Context, conn *route53.Client, id string) (*awstypes.DelegationSet, error) {
	input := &route53.GetReusableDelegationSetInput{
		Id: aws.String(id),
	}

	output, err := conn.GetReusableDelegationSet(ctx, input)

	if errs.IsA[*awstypes.NoSuchDelegationSet](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DelegationSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DelegationSet, nil
}
