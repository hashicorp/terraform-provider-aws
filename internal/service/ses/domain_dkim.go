// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_domain_dkim", name="Domain DKIM")
func resourceDomainDKIM() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainDKIMCreate,
		ReadWithoutTimeout:   resourceDomainDKIMRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"dkim_tokens": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrDomain: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDomainDKIMCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	domainName := d.Get(names.AttrDomain).(string)
	input := &ses.VerifyDomainDkimInput{
		Domain: aws.String(domainName),
	}

	_, err := conn.VerifyDomainDkim(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "requesting SES Domain DKIM (%s) verification: %s", domainName, err)
	}

	d.SetId(domainName)

	return append(diags, resourceDomainDKIMRead(ctx, d, meta)...)
}

func resourceDomainDKIMRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	verificationAttrs, err := findIdentityDKIMAttributesByIdentity(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SES Domain DKIM (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Domain DKIM (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDomain, d.Id())
	d.Set("dkim_tokens", verificationAttrs.DkimTokens)

	return diags
}

func findIdentityDKIMAttributesByIdentity(ctx context.Context, conn *ses.Client, identity string) (*awstypes.IdentityDkimAttributes, error) {
	input := &ses.GetIdentityDkimAttributesInput{
		Identities: []string{identity},
	}
	output, err := findIdentityDKIMAttributes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if v, ok := output[identity]; ok {
		return &v, nil
	}

	return nil, &retry.NotFoundError{}
}

func findIdentityDKIMAttributes(ctx context.Context, conn *ses.Client, input *ses.GetIdentityDkimAttributesInput) (map[string]awstypes.IdentityDkimAttributes, error) {
	output, err := conn.GetIdentityDkimAttributes(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.DkimAttributes == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DkimAttributes, nil
}
