// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// this enum is missing from the AWS SDK sesv2/types package.
//
// See the sesv2 documentation for valid values.
// https://docs.aws.amazon.com/ses/latest/APIReference-V2/API_GetAccount.html
type EnforcementStatus string

const (
	EnforcementStatusHealthy   EnforcementStatus = "HEALTHY"
	EnforcementStatusProbation EnforcementStatus = "PROBATION"
	EnforcementStatusShutdown  EnforcementStatus = "SHUTDOWN"
)

func (EnforcementStatus) Values() []EnforcementStatus {
	return []EnforcementStatus{
		EnforcementStatusHealthy,
		EnforcementStatusProbation,
		EnforcementStatusShutdown,
	}
}

// @FrameworkDataSource("aws_sesv2_account", name="Account")
func newDataSourceAccount(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAccount{}, nil
}

const (
	DSNameAccount = "Account Data Source"
)

type dataSourceAccount struct {
	framework.DataSourceWithModel[dataSourceAccountModel]
}

func (d *dataSourceAccount) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"dedicated_ip_auto_warmup_enabled": schema.BoolAttribute{
				Computed: true,
			},
			"details": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[detailsModel](ctx),
				Computed:   true,
			},
			"enforcement_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[EnforcementStatus](),
				Computed:   true,
			},
			"production_access_enabled": schema.BoolAttribute{
				Computed: true,
			},
			"send_quota": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[sendQuotaModel](ctx),
				Computed:   true,
			},
			"sending_enabled": schema.BoolAttribute{
				Computed: true,
			},
			"suppression_attributes": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[suppressionAttributesModel](ctx),
				Computed:   true,
			},
			"vdm_attributes": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[vdmAttributesModel](ctx),
				Computed:   true,
			},
		},
	}
}

func (d *dataSourceAccount) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SESV2Client(ctx)

	var data dataSourceAccountModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAccount(ctx, conn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data), smerr.ID)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID)
}

func findAccount(ctx context.Context, conn *sesv2.Client) (*sesv2.GetAccountOutput, error) {
	input := &sesv2.GetAccountInput{}

	output, err := conn.GetAccount(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type dataSourceAccountModel struct {
	framework.WithRegionModel
	DedicatedIpAutoWarmupEnabled types.Bool                                        `tfsdk:"dedicated_ip_auto_warmup_enabled"`
	Details                      fwtypes.ObjectValueOf[detailsModel]               `tfsdk:"details"`
	EnforcementStatus            fwtypes.StringEnum[EnforcementStatus]             `tfsdk:"enforcement_status"`
	ProductionAccessEnabled      types.Bool                                        `tfsdk:"production_access_enabled"`
	SendQuota                    fwtypes.ObjectValueOf[sendQuotaModel]             `tfsdk:"send_quota"`
	SendingEnabled               types.Bool                                        `tfsdk:"sending_enabled"`
	SuppressionAttributes        fwtypes.ObjectValueOf[suppressionAttributesModel] `tfsdk:"suppression_attributes"`
	VdmAttributes                fwtypes.ObjectValueOf[vdmAttributesModel]         `tfsdk:"vdm_attributes"`
}

type dashboardAttributesModel struct {
	EngagementMetrics fwtypes.StringEnum[awstypes.FeatureStatus] `tfsdk:"engagement_metrics"`
}

type detailsModel struct {
	AdditionalContactEmailAddresses fwtypes.SetOfString                          `tfsdk:"additional_contact_email_addresses"`
	ContactLanguage                 fwtypes.StringEnum[awstypes.ContactLanguage] `tfsdk:"contact_language"`
	MailType                        fwtypes.StringEnum[awstypes.MailType]        `tfsdk:"mail_type"`
	ReviewDetails                   fwtypes.ObjectValueOf[reviewDetailsModel]    `tfsdk:"review_details"`
	UseCaseDescription              types.String                                 `tfsdk:"use_case_description"`
	WebsiteURL                      types.String                                 `tfsdk:"website_url"`
}

type guardianAttributesModel struct {
	OptimizedSharedDelivery fwtypes.StringEnum[awstypes.FeatureStatus] `tfsdk:"optimized_shared_delivery"`
}

type reviewDetailsModel struct {
	CaseId types.String                              `tfsdk:"case_id"`
	Status fwtypes.StringEnum[awstypes.ReviewStatus] `tfsdk:"status"`
}

type sendQuotaModel struct {
	Max24HourSend   types.Float64 `tfsdk:"max_24_hour_send"`
	MaxSendRate     types.Float64 `tfsdk:"max_send_rate"`
	SentLast24Hours types.Float64 `tfsdk:"sent_last_24_hours"`
}

type suppressionAttributesModel struct {
	SuppressedReasons fwtypes.SetOfStringEnum[awstypes.SuppressionListReason] `tfsdk:"suppressed_reasons"`
}

type vdmAttributesModel struct {
	DashboardAttributes fwtypes.ObjectValueOf[dashboardAttributesModel] `tfsdk:"dashboard_attributes"`
	GuardianAttributes  fwtypes.ObjectValueOf[guardianAttributesModel]  `tfsdk:"guardian_attributes"`
	VdmEnabled          fwtypes.StringEnum[awstypes.FeatureStatus]      `tfsdk:"vdm_enabled"`
}
