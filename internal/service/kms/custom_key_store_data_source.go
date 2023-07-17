// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_kms_custom_key_store")
func DataSourceCustomKeyStore() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCustomKeyStoreRead,
		Schema: map[string]*schema.Schema{
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
			"cloud_hsm_cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"trust_anchor_certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	DSNameCustomKeyStore = "Custom Key Store"
)

func dataSourceCustomKeyStoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	input := &kms.DescribeCustomKeyStoresInput{}

	var ksID string
	if v, ok := d.GetOk("custom_key_store_id"); ok {
		input.CustomKeyStoreId = aws.String(v.(string))
		ksID = v.(string)
	}
	if v, ok := d.GetOk("custom_key_store_name"); ok {
		input.CustomKeyStoreName = aws.String(v.(string))
		ksID = v.(string)
	}

	keyStore, err := FindCustomKeyStoreByID(ctx, conn, input)

	if err != nil {
		return create.DiagError(names.KMS, create.ErrActionReading, DSNameCustomKeyStore, ksID, err)
	}

	d.SetId(aws.StringValue(keyStore.CustomKeyStoreId))
	d.Set("custom_key_store_name", keyStore.CustomKeyStoreName)
	d.Set("custom_key_store_id", keyStore.CustomKeyStoreId)
	d.Set("cloud_hsm_cluster_id", keyStore.CloudHsmClusterId)
	d.Set("connection_state", keyStore.ConnectionState)
	d.Set("creation_date", keyStore.CreationDate.Format(time.RFC3339))
	d.Set("trust_anchor_certificate", keyStore.TrustAnchorCertificate)

	return nil
}
