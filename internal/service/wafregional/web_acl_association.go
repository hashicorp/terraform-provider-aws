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
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrResourceARN: {
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

	webACLID := d.Get("web_acl_id").(string)
	resourceARN := d.Get(names.AttrResourceARN).(string)
	id := webACLAssociationCreateResourceID(webACLID, resourceARN)
	input := &wafregional.AssociateWebACLInput{
		ResourceArn: aws.String(resourceARN),
		WebACLId:    aws.String(webACLID),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.AssociateWebACLWithContext(ctx, input)
	}, wafregional.ErrCodeWAFUnavailableEntityException)

	if err != nil {
		return diag.Errorf("creating WAF Regional WebACL Association (%s): %s", id, err)
	}

	d.SetId(id)

	return diags
}

func resourceWebACLAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)

	_, resourceARN, err := webACLAssociationParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	webACL, err := findWebACLByResourceARN(ctx, conn, resourceARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Regional WebACL Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading WAF Regional WebACL Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrResourceARN, resourceARN)
	d.Set("web_acl_id", webACL.WebACLId)

	return diags
}

func resourceWebACLAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)

	_, resourceARN, err := webACLAssociationParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = conn.DisassociateWebACLWithContext(ctx, &wafregional.DisassociateWebACLInput{
		ResourceArn: aws.String(resourceARN),
	})

	if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Web ACL Association (%s): %s", d.Id(), err)
	}

	return diags
}

func findWebACLByResourceARN(ctx context.Context, conn *wafregional.WAFRegional, arn string) (*waf.WebACLSummary, error) {
	input := &wafregional.GetWebACLForResourceInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetWebACLForResourceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.WebACLSummary == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.WebACLSummary, nil
}

const webACLAssociationIDSeparator = ":"

func webACLAssociationCreateResourceID(webACLID, resourceARN string) string {
	parts := []string{webACLID, resourceARN}
	id := strings.Join(parts, webACLAssociationIDSeparator)

	return id
}

func webACLAssociationParseResourceID(id string) (string, string, error) { //nolint:unparam
	parts := strings.SplitN(id, webACLAssociationIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected WEB-ACL-ID[2]sRESOURCE-ARN", id, webACLAssociationIDSeparator)
	}

	return parts[0], parts[1], nil
}
