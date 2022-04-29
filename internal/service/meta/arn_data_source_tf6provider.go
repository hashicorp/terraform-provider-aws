package meta

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var DataSourceARNType tfsdk.DataSourceType = &arnDataSourceType{}

type arnDataSourceType struct{}

func (t *arnDataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	// TODO Generate "id" attribute.
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"account": {
				Type:     types.StringType,
				Computed: true,
			},
			"arn": {
				Type:     types.StringType,
				Required: true,
				// TODO: Validate or custom ARN type.
			},
			"partition": {
				Type:     types.StringType,
				Computed: true,
			},
			"region": {
				Type:     types.StringType,
				Computed: true,
			},
			"resource": {
				Type:     types.StringType,
				Computed: true,
			},
			"service": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}

	return schema, nil
}

func (t *arnDataSourceType) NewDataSource(ctx context.Context, provider tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return &arnDataSource{}, nil
}

type arnDataSource struct{}

func (t *arnDataSource) Read(ctx context.Context, request tfsdk.ReadDataSourceRequest, response *tfsdk.ReadDataSourceResponse) {
	var data arnData

	if diags := request.Config.Get(ctx, &data); diags.HasError() {
		response.Diagnostics.Append(diags...)

		return
	}

	if !request.Config.Raw.IsFullyKnown() {
		response.Diagnostics.AddError("Unknown Value", "An attribute value is not yet known")
	}

	arn, err := arn.Parse(*data.ARN)

	if err != nil {
		response.Diagnostics.AddError("ARN parse", err.Error())

		return
	}

	data.Account = &arn.AccountID
	data.Partition = &arn.Partition
	data.Region = &arn.Region
	data.Resource = &arn.Resource
	data.Service = &arn.Service

	if diags := response.State.Set(ctx, data); diags.HasError() {
		response.Diagnostics.Append(diags...)

		return
	}
}

type arnData struct {
	Account   *string `tfsdk:"account"`
	ARN       *string `tfsdk:"arn"`
	Partition *string `tfsdk:"partition"`
	Region    *string `tfsdk:"region"`
	Resource  *string `tfsdk:"resource"`
	Service   *string `tfsdk:"service"`
}
