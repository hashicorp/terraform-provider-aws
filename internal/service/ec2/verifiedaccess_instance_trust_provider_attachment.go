// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_verifiedaccess_instance_trust_provider_attachment", name="Verified Access Instance Trust Provider Attachment")
func resourceVerifiedAccessInstanceTrustProviderAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVerifiedAccessInstanceTrustProviderAttachmentCreate,
		ReadWithoutTimeout:   resourceVerifiedAccessInstanceTrustProviderAttachmentRead,
		DeleteWithoutTimeout: resourceVerifiedAccessInstanceTrustProviderAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"verifiedaccess_instance_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"verifiedaccess_trust_provider_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceVerifiedAccessInstanceTrustProviderAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vaiID := d.Get("verifiedaccess_instance_id").(string)
	vatpID := d.Get("verifiedaccess_trust_provider_id").(string)
	resourceID := verifiedAccessInstanceTrustProviderAttachmentCreateResourceID(vaiID, vatpID)
	input := &ec2.AttachVerifiedAccessTrustProviderInput{
		ClientToken:                   aws.String(id.UniqueId()),
		VerifiedAccessInstanceId:      aws.String(vaiID),
		VerifiedAccessTrustProviderId: aws.String(vatpID),
	}

	_, err := conn.AttachVerifiedAccessTrustProvider(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Verified Access Instance Trust Provider Attachment (%s): %s", resourceID, err)
	}

	d.SetId(resourceID)

	return append(diags, resourceVerifiedAccessInstanceTrustProviderAttachmentRead(ctx, d, meta)...)
}

func resourceVerifiedAccessInstanceTrustProviderAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vaiID, vatpID, err := verifiedAccessInstanceTrustProviderAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = findVerifiedAccessInstanceTrustProviderAttachmentExists(ctx, conn, vaiID, vatpID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Verified Access Instance Trust Provider Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Verified Access Instance Trust Provider Attachment (%s): %s", d.Id(), err)
	}

	d.Set("verifiedaccess_instance_id", vaiID)
	d.Set("verifiedaccess_trust_provider_id", vatpID)

	return diags
}

func resourceVerifiedAccessInstanceTrustProviderAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vaiID, vatpID, err := verifiedAccessInstanceTrustProviderAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting Verified Access Instance Trust Provider Attachment: %s", d.Id())
	_, err = conn.DetachVerifiedAccessTrustProvider(ctx, &ec2.DetachVerifiedAccessTrustProviderInput{
		ClientToken:                   aws.String(id.UniqueId()),
		VerifiedAccessInstanceId:      aws.String(vaiID),
		VerifiedAccessTrustProviderId: aws.String(vatpID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVerifiedAccessTrustProviderIdNotFound) ||
		tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "is not attached to instance") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Verified Access Instance Trust Provider Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

const verifiedAccessInstanceTrustProviderAttachmentResourceIDSeparator = "/"

func verifiedAccessInstanceTrustProviderAttachmentCreateResourceID(vaiID, vatpID string) string {
	parts := []string{vaiID, vatpID}
	id := strings.Join(parts, verifiedAccessInstanceTrustProviderAttachmentResourceIDSeparator)

	return id
}

func verifiedAccessInstanceTrustProviderAttachmentParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, verifiedAccessInstanceTrustProviderAttachmentResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected VerifiedAccessInstanceID%[2]sVerifiedAccessTrustProviderID", id, verifiedAccessInstanceTrustProviderAttachmentResourceIDSeparator)
}
