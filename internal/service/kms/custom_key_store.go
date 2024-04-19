// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kms_custom_key_store")
func ResourceCustomKeyStore() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomKeyStoreCreate,
		ReadWithoutTimeout:   resourceCustomKeyStoreRead,
		UpdateWithoutTimeout: resourceCustomKeyStoreUpdate,
		DeleteWithoutTimeout: resourceCustomKeyStoreDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cloud_hsm_cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"custom_key_store_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_store_password": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(7, 32)),
			},
			"trust_anchor_certificate": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

const (
	ResNameCustomKeyStore = "Custom Key Store"
)

func resourceCustomKeyStoreCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	in := &kms.CreateCustomKeyStoreInput{
		CloudHsmClusterId:      aws.String(d.Get("cloud_hsm_cluster_id").(string)),
		CustomKeyStoreName:     aws.String(d.Get("custom_key_store_name").(string)),
		KeyStorePassword:       aws.String(d.Get("key_store_password").(string)),
		TrustAnchorCertificate: aws.String(d.Get("trust_anchor_certificate").(string)),
	}

	out, err := conn.CreateCustomKeyStoreWithContext(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.KMS, create.ErrActionCreating, ResNameCustomKeyStore, d.Get("custom_key_store_name").(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.KMS, create.ErrActionCreating, ResNameCustomKeyStore, d.Get("custom_key_store_name").(string), errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.CustomKeyStoreId))

	return append(diags, resourceCustomKeyStoreRead(ctx, d, meta)...)
}

func resourceCustomKeyStoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	in := &kms.DescribeCustomKeyStoresInput{
		CustomKeyStoreId: aws.String(d.Id()),
	}
	out, err := FindCustomKeyStoreByID(ctx, conn, in)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS CustomKeyStore (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.KMS, create.ErrActionReading, ResNameCustomKeyStore, d.Id(), err)
	}

	d.Set("cloud_hsm_cluster_id", out.CloudHsmClusterId)
	d.Set("custom_key_store_name", out.CustomKeyStoreName)
	d.Set("trust_anchor_certificate", out.TrustAnchorCertificate)

	return diags
}

func resourceCustomKeyStoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	update := false

	in := &kms.UpdateCustomKeyStoreInput{
		CustomKeyStoreId:  aws.String(d.Id()),
		CloudHsmClusterId: aws.String(d.Get("cloud_hsm_cluster_id").(string)),
	}

	if d.HasChange("key_store_password") {
		in.KeyStorePassword = aws.String(d.Get("key_store_password").(string))
		update = true
	}

	if d.HasChange("custom_key_store_name") {
		in.NewCustomKeyStoreName = aws.String(d.Get("custom_key_store_name").(string))
		update = true
	}

	if !update {
		return diags
	}

	_, err := conn.UpdateCustomKeyStoreWithContext(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.KMS, create.ErrActionUpdating, ResNameCustomKeyStore, d.Id(), err)
	}

	return append(diags, resourceCustomKeyStoreRead(ctx, d, meta)...)
}

func resourceCustomKeyStoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	log.Printf("[INFO] Deleting KMS CustomKeyStore %s", d.Id())

	_, err := conn.DeleteCustomKeyStoreWithContext(ctx, &kms.DeleteCustomKeyStoreInput{
		CustomKeyStoreId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
		return diags
	}

	return diags
}
