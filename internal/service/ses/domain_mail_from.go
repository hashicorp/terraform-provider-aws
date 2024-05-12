// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_domain_mail_from")
func ResourceDomainMailFrom() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainMailFromSet,
		ReadWithoutTimeout:   resourceDomainMailFromRead,
		UpdateWithoutTimeout: resourceDomainMailFromSet,
		DeleteWithoutTimeout: resourceDomainMailFromDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrDomain: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mail_from_domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"behavior_on_mx_failure": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ses.BehaviorOnMXFailureUseDefaultValue,
			},
		},
	}
}

func resourceDomainMailFromSet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	behaviorOnMxFailure := d.Get("behavior_on_mx_failure").(string)
	domainName := d.Get(names.AttrDomain).(string)
	mailFromDomain := d.Get("mail_from_domain").(string)

	input := &ses.SetIdentityMailFromDomainInput{
		BehaviorOnMXFailure: aws.String(behaviorOnMxFailure),
		Identity:            aws.String(domainName),
		MailFromDomain:      aws.String(mailFromDomain),
	}

	_, err := conn.SetIdentityMailFromDomainWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting MAIL FROM domain: %s", err)
	}

	d.SetId(domainName)

	return append(diags, resourceDomainMailFromRead(ctx, d, meta)...)
}

func resourceDomainMailFromRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	domainName := d.Id()

	readOpts := &ses.GetIdentityMailFromDomainAttributesInput{
		Identities: []*string{
			aws.String(domainName),
		},
	}

	out, err := conn.GetIdentityMailFromDomainAttributesWithContext(ctx, readOpts)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "fetching SES MAIL FROM domain attributes for %s: %s", domainName, err)
	}

	if out == nil {
		return sdkdiag.AppendErrorf(diags, "fetching SES MAIL FROM domain attributes for %s: empty response", domainName)
	}

	attributes, ok := out.MailFromDomainAttributes[domainName]

	if !ok {
		log.Printf("[WARN] SES Domain Identity (%s) not found, removing from state", domainName)
		d.SetId("")
		return diags
	}

	d.Set("behavior_on_mx_failure", attributes.BehaviorOnMXFailure)
	d.Set(names.AttrDomain, domainName)
	d.Set("mail_from_domain", attributes.MailFromDomain)

	return diags
}

func resourceDomainMailFromDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESConn(ctx)

	domainName := d.Id()

	deleteOpts := &ses.SetIdentityMailFromDomainInput{
		Identity:       aws.String(domainName),
		MailFromDomain: nil,
	}

	_, err := conn.SetIdentityMailFromDomainWithContext(ctx, deleteOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES domain identity: %s", err)
	}

	return diags
}
