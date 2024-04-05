// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_wafregional_web_acl_association", name="Web ACL Association")
func resourceWebACLAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWebACLAssociationCreate,
		ReadWithoutTimeout:   resourceWebACLAssociationRead,
		DeleteWithoutTimeout: resourceWebACLAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"web_acl_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceWebACLAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)

	log.Printf(
		"[INFO] Creating WAF Regional Web ACL association: %s => %s",
		d.Get("web_acl_id").(string),
		d.Get("resource_arn").(string))

	params := &wafregional.AssociateWebACLInput{
		WebACLId:    aws.String(d.Get("web_acl_id").(string)),
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}

	// create association and wait on retryable error
	// no response body
	var err error
	err = retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		_, err = conn.AssociateWebACLWithContext(ctx, params)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFUnavailableEntityException) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.AssociateWebACLWithContext(ctx, params)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional Web ACL association: %s", err)
	}

	// Store association id
	d.SetId(fmt.Sprintf("%s:%s", *params.WebACLId, *params.ResourceArn))

	return diags
}

func resourceWebACLAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)

	resourceArn := WebACLAssociationParseID(d.Id())

	input := &wafregional.GetWebACLForResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	output, err := conn.GetWebACLForResourceWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
		log.Printf("[WARN] WAF Regional Web ACL for resource (%s) not found, removing from state", resourceArn)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting WAF Regional Web ACL for resource (%s): %s", resourceArn, err)
	}

	if !d.IsNewResource() && (output == nil || output.WebACLSummary == nil) {
		log.Printf("[WARN] WAF Regional Web ACL for resource (%s) not found, removing from state", resourceArn)
		d.SetId("")
		return diags
	}

	d.Set("resource_arn", resourceArn)
	d.Set("web_acl_id", output.WebACLSummary.WebACLId)

	return diags
}

func resourceWebACLAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)

	resourceArn := WebACLAssociationParseID(d.Id())

	_, err := conn.DisassociateWebACLWithContext(ctx, &wafregional.DisassociateWebACLInput{
		ResourceArn: aws.String(resourceArn),
	})

	if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Web ACL Association (%s): %s", d.Id(), err)
	}

	return diags
}

func WebACLAssociationParseID(id string) (resourceArn string) {
	parts := strings.SplitN(id, ":", 2)
	resourceArn = parts[1]
	return
}
