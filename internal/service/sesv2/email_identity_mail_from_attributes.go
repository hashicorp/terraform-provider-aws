// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var errMailFromRequired = errors.New("mail from domain is required if behavior on MX failure is REJECT_MESSAGE")

// @SDKResource("aws_sesv2_email_identity_mail_from_attributes", name="Email Identity Mail From Attributes")
func resourceEmailIdentityMailFromAttributes() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEmailIdentityMailFromAttributesCreate,
		ReadWithoutTimeout:   resourceEmailIdentityMailFromAttributesRead,
		UpdateWithoutTimeout: resourceEmailIdentityMailFromAttributesUpdate,
		DeleteWithoutTimeout: resourceEmailIdentityMailFromAttributesDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"behavior_on_mx_failure": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          string(types.BehaviorOnMxFailureUseDefaultValue),
				ValidateDiagFunc: enum.Validate[types.BehaviorOnMxFailure](),
			},
			"email_identity": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mail_from_domain": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

const (
	resNameEmailIdentityMailFromAttributes = "Email Identity Mail From Attributes"
)

func resourceEmailIdentityMailFromAttributesCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	in := &sesv2.PutEmailIdentityMailFromAttributesInput{
		EmailIdentity: aws.String(d.Get("email_identity").(string)),
	}

	if v, ok := d.GetOk("mail_from_domain"); ok {
		in.MailFromDomain = aws.String(v.(string))
	}

	if v, ok := d.GetOk("behavior_on_mx_failure"); ok {
		in.BehaviorOnMxFailure = types.BehaviorOnMxFailure(v.(string))
	}

	if in.BehaviorOnMxFailure == types.BehaviorOnMxFailureRejectMessage && (in.MailFromDomain == nil || aws.ToString(in.MailFromDomain) == "") {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, resNameEmailIdentityMailFromAttributes, d.Get("email_identity").(string), errMailFromRequired)
	}

	out, err := conn.PutEmailIdentityMailFromAttributes(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, resNameEmailIdentityMailFromAttributes, d.Get("email_identity").(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionCreating, resNameEmailIdentityMailFromAttributes, d.Get("email_identity").(string), errors.New("empty output"))
	}

	d.SetId(d.Get("email_identity").(string))

	return append(diags, resourceEmailIdentityMailFromAttributesRead(ctx, d, meta)...)
}

func resourceEmailIdentityMailFromAttributesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	out, err := findEmailIdentityByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 EmailIdentityMailFromAttributes (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, resNameEmailIdentityMailFromAttributes, d.Id(), err)
	}

	d.Set("email_identity", d.Id())

	if out.MailFromAttributes != nil {
		d.Set("behavior_on_mx_failure", out.MailFromAttributes.BehaviorOnMxFailure)
		d.Set("mail_from_domain", out.MailFromAttributes.MailFromDomain)
	} else {
		d.Set("behavior_on_mx_failure", nil)
		d.Set("mail_from_domain", nil)
	}

	return diags
}

func resourceEmailIdentityMailFromAttributesUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	update := false

	in := &sesv2.PutEmailIdentityMailFromAttributesInput{
		EmailIdentity: aws.String(d.Id()),
	}

	if d.HasChanges("behavior_on_mx_failure", "mail_from_domain") {
		in.BehaviorOnMxFailure = types.BehaviorOnMxFailure((d.Get("behavior_on_mx_failure").(string)))
		in.MailFromDomain = aws.String(d.Get("mail_from_domain").(string))

		if in.BehaviorOnMxFailure == types.BehaviorOnMxFailureRejectMessage && (in.MailFromDomain == nil || aws.ToString(in.MailFromDomain) == "") {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, resNameEmailIdentityMailFromAttributes, d.Get("email_identity").(string), errMailFromRequired)
		}

		update = true
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating SESV2 EmailIdentityMailFromAttributes (%s): %#v", d.Id(), in)
	_, err := conn.PutEmailIdentityMailFromAttributes(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionUpdating, resNameEmailIdentityMailFromAttributes, d.Id(), err)
	}

	return append(diags, resourceEmailIdentityMailFromAttributesRead(ctx, d, meta)...)
}

func resourceEmailIdentityMailFromAttributesDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	log.Printf("[INFO] Deleting SESV2 EmailIdentityMailFromAttributes: %s", d.Id())
	_, err := conn.PutEmailIdentityMailFromAttributes(ctx, &sesv2.PutEmailIdentityMailFromAttributesInput{
		EmailIdentity: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionDeleting, resNameEmailIdentityMailFromAttributes, d.Id(), err)
	}

	return diags
}
