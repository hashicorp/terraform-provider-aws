// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_hsm_client_certificate", name="HSM Client Certificate")
// @Tags(identifierAttribute="arn")
func resourceHSMClientCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHSMClientCertificateCreate,
		ReadWithoutTimeout:   resourceHSMClientCertificateRead,
		UpdateWithoutTimeout: resourceHSMClientCertificateUpdate,
		DeleteWithoutTimeout: resourceHSMClientCertificateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hsm_client_certificate_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hsm_client_certificate_public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceHSMClientCertificateCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	certIdentifier := d.Get("hsm_client_certificate_identifier").(string)

	input := redshift.CreateHsmClientCertificateInput{
		HsmClientCertificateIdentifier: aws.String(certIdentifier),
		Tags:                           getTagsIn(ctx),
	}

	out, err := conn.CreateHsmClientCertificate(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift HSM Client Certificate (%s): %s", certIdentifier, err)
	}

	d.SetId(aws.ToString(out.HsmClientCertificate.HsmClientCertificateIdentifier))

	return append(diags, resourceHSMClientCertificateRead(ctx, d, meta)...)
}

func resourceHSMClientCertificateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	out, err := findHSMClientCertificateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift HSM Client Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift HSM Client Certificate (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   names.Redshift,
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("hsmclientcertificate:%s", d.Id()),
	}.String()

	d.Set(names.AttrARN, arn)

	d.Set("hsm_client_certificate_identifier", out.HsmClientCertificateIdentifier)
	d.Set("hsm_client_certificate_public_key", out.HsmClientCertificatePublicKey)

	setTagsOut(ctx, out.Tags)

	return diags
}

func resourceHSMClientCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceHSMClientCertificateRead(ctx, d, meta)...)
}

func resourceHSMClientCertificateDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	deleteInput := redshift.DeleteHsmClientCertificateInput{
		HsmClientCertificateIdentifier: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Redshift HSM Client Certificate: %s", d.Id())
	_, err := conn.DeleteHsmClientCertificate(ctx, &deleteInput)

	if err != nil {
		if errs.IsA[*awstypes.HsmClientCertificateNotFoundFault](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "updating Redshift HSM Client Certificate (%s) tags: %s", d.Get(names.AttrARN).(string), err)
	}

	return diags
}
