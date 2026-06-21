// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_kms_key_last_usage", name="Key Last Usage")
func newKeyLastUsageDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &keyLastUsageDataSource{}, nil
}

type keyLastUsageDataSource struct {
	framework.DataSourceWithModel[keyLastUsageDataSourceModel]
}

func (d *keyLastUsageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"key_creation_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrKeyID: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.Any(
						stringvalidator.RegexMatches(keyIDRegex, "must be a KMS Key ID"),
						fwvalidators.ARN(),
					),
				},
			},
			"key_last_usage": framework.DataSourceComputedListOfObjectAttribute[dsKeyLastUsageData](ctx),
			"tracking_start_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (d *keyLastUsageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().KMSClient(ctx)

	var data keyLastUsageDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	keyID := data.KeyID.ValueString()
	out, err := findKeyLastUsageByID(ctx, conn, keyID)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, keyID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data), smerr.ID, keyID)
	if resp.Diagnostics.HasError() {
		return
	}

	if id := keyIDFromARN(keyID); id != "" {
		data.KeyID = types.StringValue(id)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type keyLastUsageDataSourceModel struct {
	framework.WithRegionModel
	KeyCreationDate   timetypes.RFC3339                                   `tfsdk:"key_creation_date"`
	KeyID             types.String                                        `tfsdk:"key_id" autoflex:"-"`
	KeyLastUsage      fwtypes.ListNestedObjectValueOf[dsKeyLastUsageData] `tfsdk:"key_last_usage"`
	TrackingStartDate timetypes.RFC3339                                   `tfsdk:"tracking_start_date"`
}

type dsKeyLastUsageData struct {
	CloudTrailEventID types.String                                               `tfsdk:"cloud_trail_event_id"`
	KMSRequestID      types.String                                               `tfsdk:"kms_request_id"`
	Operation         fwtypes.StringEnum[awstypes.KeyLastUsageTrackingOperation] `tfsdk:"operation"`
	Timestamp         timetypes.RFC3339                                          `tfsdk:"timestamp"`
}

func findKeyLastUsageByID(ctx context.Context, conn *kms.Client, id string) (*kms.GetKeyLastUsageOutput, error) {
	input := kms.GetKeyLastUsageInput{
		KeyId: aws.String(id),
	}

	output, err := conn.GetKeyLastUsage(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}
