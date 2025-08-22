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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_domain_mail_from", name="MAIL FROM Domain")
func resourceDomainMailFrom() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainMailFromSet,
		ReadWithoutTimeout:   resourceDomainMailFromRead,
		UpdateWithoutTimeout: resourceDomainMailFromSet,
		DeleteWithoutTimeout: resourceDomainMailFromDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"behavior_on_mx_failure": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.BehaviorOnMXFailureUseDefaultValue,
				ValidateDiagFunc: enum.Validate[awstypes.BehaviorOnMXFailure](),
			},
			names.AttrDomain: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mail_from_domain": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceDomainMailFromSet(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	domainName := d.Get(names.AttrDomain).(string)
	input := &ses.SetIdentityMailFromDomainInput{
		BehaviorOnMXFailure: awstypes.BehaviorOnMXFailure(d.Get("behavior_on_mx_failure").(string)),
		Identity:            aws.String(domainName),
		MailFromDomain:      aws.String(d.Get("mail_from_domain").(string)),
	}

	_, err := conn.SetIdentityMailFromDomain(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting SES MAIL FROM Domain (%s): %s", domainName, err)
	}

	if d.IsNewResource() {
		d.SetId(domainName)
	}

	return append(diags, resourceDomainMailFromRead(ctx, d, meta)...)
}

func resourceDomainMailFromRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	attributes, err := findIdentityMailFromDomainAttributesByIdentity(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SES MAIL FROM Domain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES MAIL FROM Domain (%s): %s", d.Id(), err)
	}

	d.Set("behavior_on_mx_failure", attributes.BehaviorOnMXFailure)
	d.Set(names.AttrDomain, d.Id())
	d.Set("mail_from_domain", attributes.MailFromDomain)

	return diags
}

func resourceDomainMailFromDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	log.Printf("[DEBUG] Deleting SES MAIL FROM Domain: %s", d.Id())
	_, err := conn.SetIdentityMailFromDomain(ctx, &ses.SetIdentityMailFromDomainInput{
		Identity: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES MAIL FROM Domain (%s): %s", d.Id(), err)
	}

	return diags
}

func findIdentityMailFromDomainAttributesByIdentity(ctx context.Context, conn *ses.Client, identity string) (*awstypes.IdentityMailFromDomainAttributes, error) {
	input := &ses.GetIdentityMailFromDomainAttributesInput{
		Identities: []string{identity},
	}
	output, err := findIdentityMailFromDomainAttributes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if v, ok := output[identity]; ok {
		return &v, nil
	}

	return nil, &retry.NotFoundError{}
}

func findIdentityMailFromDomainAttributes(ctx context.Context, conn *ses.Client, input *ses.GetIdentityMailFromDomainAttributesInput) (map[string]awstypes.IdentityMailFromDomainAttributes, error) {
	output, err := conn.GetIdentityMailFromDomainAttributes(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.MailFromDomainAttributes == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.MailFromDomainAttributes, nil
}
