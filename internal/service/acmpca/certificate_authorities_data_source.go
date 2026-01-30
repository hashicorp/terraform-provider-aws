// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	awstypes "github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_acmpca_certificate_authorities", name="Certificate Authorities")
func newDataSourceCertificateAuthorities(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceCertificateAuthorities{}, nil
}

const (
	DSNameCertificateAuthorities = "Certificate Authorities Data Source"
)

type dataSourceCertificateAuthorities struct {
	framework.DataSourceWithModel[dataSourceCertificateAuthoritiesModel]
}

func (d *dataSourceCertificateAuthorities) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrResourceOwner: schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.ResourceOwner](),
				Optional:    true,
				Description: "Filter by resource owner. Valid values are `SELF` and `OTHER_ACCOUNTS`.",
			},
			names.AttrARNs: schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *dataSourceCertificateAuthorities) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ACMPCAClient(ctx)

	var data dataSourceCertificateAuthoritiesModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := findCertificateAuthorities(ctx, conn, data.ResourceOwner.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
		return
	}
	data.ARNs = flex.FlattenFrameworkStringValueListOfString(ctx, out)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID)
}

type dataSourceCertificateAuthoritiesModel struct {
	framework.WithRegionModel
	ResourceOwner fwtypes.StringEnum[awstypes.ResourceOwner] `tfsdk:"resource_owner"`
	ARNs          fwtypes.ListOfString                       `tfsdk:"arns"`
}

func findCertificateAuthorities(ctx context.Context, client *acmpca.Client, resourceOwner string) ([]string, error) {
	var arns []string
	var input acmpca.ListCertificateAuthoritiesInput
	if resourceOwner != "" {
		input.ResourceOwner = awstypes.ResourceOwner(resourceOwner)
	}
	paginator := acmpca.NewListCertificateAuthoritiesPaginator(client, &input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, ca := range page.CertificateAuthorities {
			arns = append(arns, aws.ToString(ca.Arn))
		}
	}
	return arns, nil
}
