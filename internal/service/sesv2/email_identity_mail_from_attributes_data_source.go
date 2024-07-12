// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_sesv2_email_identity_mail_from_attributes")
func DataSourceEmailIdentityMailFromAttributes() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEmailIdentityMailFromAttributesRead,

		Schema: map[string]*schema.Schema{
			"behavior_on_mx_failure": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email_identity": {
				Type:     schema.TypeString,
				Required: true,
			},
			"mail_from_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	DSNameEmailIdentityMailFromAttributes = "Email Identity Mail From Attributes Data Source"
)

func dataSourceEmailIdentityMailFromAttributesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	name := d.Get("email_identity").(string)

	out, err := FindEmailIdentityByID(ctx, conn, name)

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, ResNameEmailIdentityMailFromAttributes, name, err)
	}

	d.SetId(name)
	d.Set("email_identity", name)

	if out.MailFromAttributes != nil {
		d.Set("behavior_on_mx_failure", out.MailFromAttributes.BehaviorOnMxFailure)
		d.Set("mail_from_domain", out.MailFromAttributes.MailFromDomain)
	} else {
		d.Set("behavior_on_mx_failure", nil)
		d.Set("mail_from_domain", nil)
	}

	return diags
}
