// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	webACLAssociationResourceIDPartCount = 2
)

// @SDKResource("aws_wafv2_web_acl_association", name="Web ACL Association")
func resourceWebACLAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWebACLAssociationCreate,
		ReadWithoutTimeout:   resourceWebACLAssociationRead,
		DeleteWithoutTimeout: resourceWebACLAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrResourceARN: {
					Type:         schema.TypeString,
					ForceNew:     true,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
				"web_acl_arn": {
					Type:         schema.TypeString,
					ForceNew:     true,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
			}
		},
	}
}

func resourceWebACLAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	webACLARN := d.Get("web_acl_arn").(string)
	resourceARN := d.Get(names.AttrResourceARN).(string)
	id := errs.Must(flex.FlattenResourceId([]string{webACLARN, resourceARN}, webACLAssociationResourceIDPartCount, true))
	input := &wafv2.AssociateWebACLInput{
		ResourceArn: aws.String(resourceARN),
		WebACLArn:   aws.String(webACLARN),
	}

	log.Printf("[INFO] Creating WAFv2 WebACL Association: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*awstypes.WAFUnavailableEntityException](ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.AssociateWebACL(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAFv2 WebACL Association (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceWebACLAssociationRead(ctx, d, meta)...)
}

func resourceWebACLAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), webACLAssociationResourceIDPartCount, true)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resourceARN := parts[1]
	webACL, err := findWebACLByResourceARN(ctx, conn, resourceARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAFv2 WebACL Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACL Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrResourceARN, resourceARN)
	d.Set("web_acl_arn", webACL.ARN)

	return diags
}

func resourceWebACLAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), webACLAssociationResourceIDPartCount, true)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting WAFv2 WebACL Association: %s", d.Id())
	resourceARN := parts[1]
	_, err = conn.DisassociateWebACL(ctx, &wafv2.DisassociateWebACLInput{
		ResourceArn: aws.String(resourceARN),
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAFv2 WebACL Association (%s): %s", d.Id(), err)
	}

	return diags
}

func findWebACLByResourceARN(ctx context.Context, conn *wafv2.Client, arn string) (*awstypes.WebACL, error) {
	input := &wafv2.GetWebACLForResourceInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetWebACLForResource(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.WebACL == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.WebACL, nil
}
