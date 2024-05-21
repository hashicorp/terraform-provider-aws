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

// @SDKDataSource("aws_iam_server_certificates", name="Server Certificates")
func dataSourceServerCertificates() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServerCertificatesRead,

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
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
			"server_certificates": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrPath: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
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
					},
				},
			},
		},
	}
}

func dataSourceServerCertificatesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	log.Printf("[DEBUG] Reading IAM Server Certificates")
	pages := iam.NewListServerCertificatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading IAM Server Certificates: listing certificates: %s", err)
		}
		for _, cert := range page.ServerCertificateMetadataList {
			if matcher(cert) {
				metadatas = append(metadatas, cert)
			}
		}
	}

	if len(metadatas) == 0 {
		return sdkdiag.AppendErrorf(diags, "Search for AWS IAM server certificates returned no results")
	}

	sort.Sort(CertificateByExpiration(metadatas))

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("server_certificates", flattenServerCerts(metadatas))

	return diags
}

func flattenServerCerts(apiObjects []awstypes.ServerCertificateMetadata) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == (awstypes.ServerCertificateMetadata{}) {
			continue
		}
		tfList = append(tfList, flattenServerCert(apiObject))
	}

	return tfList
}

func flattenServerCert(apiObject awstypes.ServerCertificateMetadata) map[string]interface{} {
	if apiObject == (awstypes.ServerCertificateMetadata{}) {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Arn; v != nil {
		m[names.AttrARN] = aws.ToString(v)
	}
	if v := apiObject.Expiration; v != nil {
		m["expiration_date"] = aws.ToTime(v).Format(time.RFC3339)
	}
	if v := apiObject.UploadDate; v != nil {
		m["upload_date"] = aws.ToTime(v).Format(time.RFC3339)
	}
	if v := apiObject.Path; v != nil {
		m[names.AttrPath] = aws.ToString(v)
	}
	if v := apiObject.ServerCertificateName; v != nil {
		m[names.AttrName] = aws.ToString(v)
	}
	if v := apiObject.ServerCertificateId; v != nil {
		m[names.AttrID] = aws.ToString(v)
	}

	return m
}
