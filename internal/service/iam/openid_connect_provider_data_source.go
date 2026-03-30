// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_openid_connect_provider", name="OIDC Provider")
// @Tags
// @Testing(tagsIdentifierAttribute="arn", tagsResourceType="OIDCProvider")
func dataSourceOpenIDConnectProvider() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOpenIDConnectProviderRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
				ExactlyOneOf: []string{names.AttrARN, names.AttrURL},
			},
			"client_id_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"thumbprint_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrURL: {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validOpenIDURL,
				DiffSuppressFunc: suppressOpenIDURL,
				ExactlyOneOf:     []string{names.AttrARN, names.AttrURL},
			},
		},
	}
}

func dataSourceOpenIDConnectProviderRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	var input iam.GetOpenIDConnectProviderInput

	if v, ok := d.GetOk(names.AttrARN); ok {
		input.OpenIDConnectProviderArn = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrURL); ok {
		url := v.(string)

		oidcpEntry, err := findOpenIDConnectProviderByURL(ctx, conn, url)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading IAM OIDC Provider (%s): %s", url, err)
		}

		input.OpenIDConnectProviderArn = oidcpEntry.Arn
	}

	output, err := findOpenIDConnectProvider(ctx, conn, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM OIDC Provider: %s", err)
	}

	arn := aws.ToString(input.OpenIDConnectProviderArn)
	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	d.Set("client_id_list", output.ClientIDList)
	d.Set("thumbprint_list", output.ThumbprintList)
	d.Set(names.AttrURL, output.Url)

	setTagsOut(ctx, output.Tags)

	return diags
}

func findOpenIDConnectProviderByURL(ctx context.Context, conn *iam.Client, url string) (*awstypes.OpenIDConnectProviderListEntry, error) {
	var input iam.ListOpenIDConnectProvidersInput

	output, err := conn.ListOpenIDConnectProviders(ctx, &input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.OpenIDConnectProviderList {
		if p := &v; inttypes.IsZero(p) {
			continue
		}

		arnUrl, err := urlFromOpenIDConnectProviderARN(aws.ToString(v.Arn))
		if err != nil {
			return nil, err
		}

		if arnUrl == strings.TrimPrefix(url, "https://") {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func urlFromOpenIDConnectProviderARN(arn string) (string, error) {
	parts := strings.SplitN(arn, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("reading OpenID Connect Provider expected the arn to be like: arn:PARTITION:iam::ACCOUNT:oidc-provider/URL but got: %s", arn)
	}
	return parts[1], nil
}
