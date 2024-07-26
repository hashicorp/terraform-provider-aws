// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_kms_custom_key_store", name="Custom Key Store")
func dataSourceCustomKeyStore() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCustomKeyStoreRead,

		Schema: map[string]*schema.Schema{
			"cloud_hsm_cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_key_store_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"custom_key_store_name"},
			},
			"custom_key_store_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"custom_key_store_id"},
			},
			"trust_anchor_certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCustomKeyStoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	input := &kms.DescribeCustomKeyStoresInput{}

	var ksID string
	if v, ok := d.GetOk("custom_key_store_id"); ok {
		input.CustomKeyStoreId = aws.String(v.(string))
		ksID = v.(string)
	} else if v, ok := d.GetOk("custom_key_store_name"); ok {
		input.CustomKeyStoreName = aws.String(v.(string))
		ksID = v.(string)
	}

	keyStore, err := findCustomKeyStore(ctx, conn, input, tfslices.PredicateTrue[*awstypes.CustomKeyStoresListEntry]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Custom Key Store (%s): %s", ksID, err)
	}

	d.SetId(aws.ToString(keyStore.CustomKeyStoreId))
	d.Set("cloud_hsm_cluster_id", keyStore.CloudHsmClusterId)
	d.Set("connection_state", keyStore.ConnectionState)
	d.Set(names.AttrCreationDate, keyStore.CreationDate.Format(time.RFC3339))
	d.Set("custom_key_store_id", keyStore.CustomKeyStoreId)
	d.Set("custom_key_store_name", keyStore.CustomKeyStoreName)
	d.Set("trust_anchor_certificate", keyStore.TrustAnchorCertificate)

	return diags
}
