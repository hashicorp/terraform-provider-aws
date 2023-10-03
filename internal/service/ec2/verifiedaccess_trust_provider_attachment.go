// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_verifiedaccess_trust_provider_attachment", name="Verified Access Trust Provider Attachment")
func ResourceTrustProviderAttachment() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: resourceTrustProviderAttachmentRead,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"trust_provider_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceTrustProviderAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	instanceId, trustProviderId, err := VerifiedAccessTrustProviderAttachmentParseId(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Verified Access Trust Provider Attachment (%s): %s", d.Id(), err)
	}

	_, err = FindVerifiedAccessTrustProviderAttachmentByID(ctx, conn, instanceId, trustProviderId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Verified Access Trust Provider Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Verified Access Trust Provider Attachment (%s): %s", d.Id(), err)
	}

	d.Set("instance_id", instanceId)
	d.Set("trust_provider_id", trustProviderId)

	return diags
}

func VerifiedAccessTrustProviderAttachmentParseId(id string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.New("unexpected format of ID, expected instanceId:trustProviderId")
	}

	return parts[0], parts[1], nil
}
