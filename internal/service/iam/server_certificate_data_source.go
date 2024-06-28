// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_server_certificate", name="Server Certificate")
func dataSourceServerCertificate() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServerCertificateRead,

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validation.StringLenBetween(0, 128),
			},

			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validation.StringLenBetween(0, 128-id.UniqueIDSuffixLength),
			},

			"path_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"latest": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrPath: {
				Type:     schema.TypeString,
				Computed: true,
			},

			"expiration_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"upload_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"certificate_body": {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrCertificateChain: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

type CertificateByExpiration []awstypes.ServerCertificateMetadata

func (m CertificateByExpiration) Len() int {
	return len(m)
}

func (m CertificateByExpiration) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m CertificateByExpiration) Less(i, j int) bool {
	return m[i].Expiration.After(*m[j].Expiration)
}

func dataSourceServerCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	var matcher = func(cert awstypes.ServerCertificateMetadata) bool {
		return strings.HasPrefix(aws.ToString(cert.ServerCertificateName), d.Get(names.AttrNamePrefix).(string))
	}
	if v, ok := d.GetOk(names.AttrName); ok {
		matcher = func(cert awstypes.ServerCertificateMetadata) bool {
			return aws.ToString(cert.ServerCertificateName) == v.(string)
		}
	}

	var metadatas []awstypes.ServerCertificateMetadata
	input := &iam.ListServerCertificatesInput{}
	if v, ok := d.GetOk("path_prefix"); ok {
		input.PathPrefix = aws.String(v.(string))
	}
	log.Printf("[DEBUG] Reading IAM Server Certificate")
	pages := iam.NewListServerCertificatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading IAM Server Certificate: listing certificates: %s", err)
		}
		for _, cert := range page.ServerCertificateMetadataList {
			if matcher(cert) {
				metadatas = append(metadatas, cert)
			}
		}
	}

	if len(metadatas) == 0 {
		return sdkdiag.AppendErrorf(diags, "Search for AWS IAM server certificate returned no results")
	}
	if len(metadatas) > 1 {
		if !d.Get("latest").(bool) {
			return sdkdiag.AppendErrorf(diags, "Search for AWS IAM server certificate returned too many results")
		}

		sort.Sort(CertificateByExpiration(metadatas))
	}

	metadata := metadatas[0]
	d.SetId(aws.ToString(metadata.ServerCertificateId))
	d.Set(names.AttrARN, metadata.Arn)
	d.Set(names.AttrPath, metadata.Path)
	d.Set(names.AttrName, metadata.ServerCertificateName)
	if metadata.Expiration != nil {
		d.Set("expiration_date", metadata.Expiration.Format(time.RFC3339))
	}

	log.Printf("[DEBUG] Get Public Key Certificate for %s", *metadata.ServerCertificateName)
	serverCertificateResp, err := conn.GetServerCertificate(ctx, &iam.GetServerCertificateInput{
		ServerCertificateName: metadata.ServerCertificateName,
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Server Certificate: getting certificate details: %s", err)
	}
	d.Set("upload_date", serverCertificateResp.ServerCertificate.ServerCertificateMetadata.UploadDate.Format(time.RFC3339))
	d.Set("certificate_body", serverCertificateResp.ServerCertificate.CertificateBody)
	d.Set(names.AttrCertificateChain, serverCertificateResp.ServerCertificate.CertificateChain)

	return diags
}
