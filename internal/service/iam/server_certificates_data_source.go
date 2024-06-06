// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Server Certificates")
func newDataSourceServerCertificates(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceServerCertificates{}, nil
}

const (
	DSNameServerCertificates = "Server Certificates Data Source"
)

type dataSourceServerCertificates struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceServerCertificates) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "aws_iam_server_certificates"
}

func (d *dataSourceServerCertificates) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrName: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrNamePrefix: schema.StringAttribute{
				Optional: true,
			},
			"path_prefix": schema.StringAttribute{
				Optional: true,
			},
			"latest": schema.BoolAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"server_certificates": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[certificate](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrID: schema.StringAttribute{
							Computed: true,
						},
						names.AttrARN: schema.StringAttribute{
							Computed: true,
						},
						names.AttrPath: schema.StringAttribute{
							Computed: true,
						},
						names.AttrName: schema.StringAttribute{
							Computed: true,
						},
						"expiration_date": schema.StringAttribute{
							Computed: true,
						},
						"upload_date": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceServerCertificates) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().IAMClient(ctx)

	var data serverCertificatesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(d.Meta().AccountID)
	var matcher func(cert awstypes.ServerCertificateMetadata) bool
	if !data.Name.IsNull() {
		matcher = func(cert awstypes.ServerCertificateMetadata) bool {
			return aws.ToString(cert.ServerCertificateName) == data.Name.ValueString()
		}
	} else if !data.NamePrefix.IsNull() {
		matcher = func(cert awstypes.ServerCertificateMetadata) bool {
			return strings.HasPrefix(aws.ToString(cert.ServerCertificateName), data.NamePrefix.ValueString())
		}
	} else {
		matcher = func(_ awstypes.ServerCertificateMetadata) bool {
			return true
		}
	}
	input := &iam.ListServerCertificatesInput{}
	if !data.PathPrefix.IsNull() {
		input.PathPrefix = aws.String(data.PathPrefix.ValueString())
	}
	paginator := iam.NewListServerCertificatesPaginator(conn, input)

	var out iam.ListServerCertificatesOutput
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"reading IAM Server Certificate: listing certificates",
				err.Error(),
			)
			return
		}
		for _, cert := range page.ServerCertificateMetadataList {
			if matcher(cert) {
				out.ServerCertificateMetadataList = append(out.ServerCertificateMetadataList, cert)
			}
		}
	}
	if data.Latest.ValueBool() {
		sort.Sort(CertificateByExpiration(out.ServerCertificateMetadataList))
	}

	res := flex.Flatten(ctx, out, &data)
	resp.Diagnostics.Append(res...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type serverCertificatesDataSourceModel struct {
	ID                            types.String                                 `tfsdk:"id"`
	Name                          types.String                                 `tfsdk:"name"`
	NamePrefix                    types.String                                 `tfsdk:"name_prefix"`
	PathPrefix                    types.String                                 `tfsdk:"path_prefix"`
	Latest                        types.Bool                                   `tfsdk:"latest"`
	ServerCertificateMetadataList fwtypes.ListNestedObjectValueOf[certificate] `tfsdk:"server_certificates"`
}

type certificate struct {
	ServerCertificateID   types.String      `tfsdk:"id"`
	ARN                   types.String      `tfsdk:"arn"`
	Path                  types.String      `tfsdk:"path"`
	ServerCertificateName types.String      `tfsdk:"name"`
	Expiration            timetypes.RFC3339 `tfsdk:"expiration_date"`
	UploadDate            timetypes.RFC3339 `tfsdk:"upload_date"`
}
