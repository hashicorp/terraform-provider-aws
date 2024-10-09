// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_kms_custom_key_store", name="Custom Key Store")
func resourceCustomKeyStore() *schema.Resource {
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
				ForceNew: true,
			},
		},
	}
}

func resourceCustomKeyStoreCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	name := d.Get("custom_key_store_name").(string)
	input := &kms.CreateCustomKeyStoreInput{
		CloudHsmClusterId:      aws.String(d.Get("cloud_hsm_cluster_id").(string)),
		CustomKeyStoreName:     aws.String(name),
		KeyStorePassword:       aws.String(d.Get("key_store_password").(string)),
		TrustAnchorCertificate: aws.String(d.Get("trust_anchor_certificate").(string)),
	}

	output, err := conn.CreateCustomKeyStore(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating KMS Custom Key Store (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.CustomKeyStoreId))

	return append(diags, resourceCustomKeyStoreRead(ctx, d, meta)...)
}

func resourceCustomKeyStoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	output, err := findCustomKeyStoreByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS Custom Key Store (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Custom Key Store (%s): %s", d.Id(), err)
	}

	d.Set("cloud_hsm_cluster_id", output.CloudHsmClusterId)
	d.Set("custom_key_store_name", output.CustomKeyStoreName)
	d.Set("key_store_password", d.Get("key_store_password"))
	d.Set("trust_anchor_certificate", output.TrustAnchorCertificate)

	return diags
}

func resourceCustomKeyStoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	input := &kms.UpdateCustomKeyStoreInput{
		CloudHsmClusterId: aws.String(d.Get("cloud_hsm_cluster_id").(string)),
		CustomKeyStoreId:  aws.String(d.Id()),
	}

	if d.HasChange("custom_key_store_name") {
		input.NewCustomKeyStoreName = aws.String(d.Get("custom_key_store_name").(string))
	}

	if d.HasChange("key_store_password") {
		input.KeyStorePassword = aws.String(d.Get("key_store_password").(string))
	}

	_, err := conn.UpdateCustomKeyStore(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating KMS Custom Key Store (%s): %s", d.Id(), err)
	}

	return append(diags, resourceCustomKeyStoreRead(ctx, d, meta)...)
}

func resourceCustomKeyStoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	log.Printf("[INFO] Deleting KMS Custom Key Store: %s", d.Id())
	_, err := conn.DeleteCustomKeyStore(ctx, &kms.DeleteCustomKeyStoreInput{
		CustomKeyStoreId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	return diags
}

func findCustomKeyStoreByID(ctx context.Context, conn *kms.Client, id string) (*awstypes.CustomKeyStoresListEntry, error) {
	input := &kms.DescribeCustomKeyStoresInput{
		CustomKeyStoreId: aws.String(id),
	}

	return findCustomKeyStore(ctx, conn, input, tfslices.PredicateTrue[*awstypes.CustomKeyStoresListEntry]())
}

func findCustomKeyStore(ctx context.Context, conn *kms.Client, input *kms.DescribeCustomKeyStoresInput, filter tfslices.Predicate[*awstypes.CustomKeyStoresListEntry]) (*awstypes.CustomKeyStoresListEntry, error) {
	output, err := findCustomKeyStores(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCustomKeyStores(ctx context.Context, conn *kms.Client, input *kms.DescribeCustomKeyStoresInput, filter tfslices.Predicate[*awstypes.CustomKeyStoresListEntry]) ([]awstypes.CustomKeyStoresListEntry, error) {
	var output []awstypes.CustomKeyStoresListEntry

	pages := kms.NewDescribeCustomKeyStoresPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.CustomKeyStoreNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return output, err
		}

		for _, v := range page.CustomKeyStores {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
