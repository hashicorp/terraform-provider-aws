// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// @FrameworkDataSource("aws_route53_records", name="Records")
func newRecordsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &recordsDataSource{}, nil
}

type recordsDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *recordsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name_regex": schema.StringAttribute{
				CustomType: fwtypes.RegexpType,
				Optional:   true,
			},
			"resource_record_sets": framework.DataSourceComputedListOfObjectAttribute[resourceRecordSetModelReadonly](ctx),
			"zone_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *recordsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data recordsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().Route53Client(ctx)

	hostedZoneID := fwflex.StringValueFromFramework(ctx, data.ZoneID)
	input := route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID),
	}
	filter := tfslices.PredicateTrue[*awstypes.ResourceRecordSet]()
	if !data.NameRegex.IsNull() {
		filter = func(v *awstypes.ResourceRecordSet) bool {
			return data.NameRegex.ValueRegexp().MatchString(aws.ToString(v.Name))
		}
	}

	output, err := findResourceRecordSets(ctx, conn, &input, tfslices.PredicateTrue[*route53.ListResourceRecordSetsOutput](), filter)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("listing Route 53 Records (%s)", hostedZoneID), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(
		ctx,
		struct {
			ResourceRecordSets []awstypes.ResourceRecordSet
		}{
			ResourceRecordSets: output,
		},
		&data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type recordsDataSourceModel struct {
	NameRegex          fwtypes.Regexp                                                  `tfsdk:"name_regex"`
	ResourceRecordSets fwtypes.ListNestedObjectValueOf[resourceRecordSetModelReadonly] `tfsdk:"resource_record_sets"`
	ZoneID             types.String                                                    `tfsdk:"zone_id"`
}

type resourceRecordSetModelReadonly struct {
	AliasTarget             fwtypes.ObjectValueOf[aliasTargetModel]                  `tfsdk:"alias_target"`
	CIDRRoutingConfig       fwtypes.ObjectValueOf[cidrRoutingConfigModel]            `tfsdk:"cidr_routing_config"`
	Failover                fwtypes.StringEnum[awstypes.ResourceRecordSetFailover]   `tfsdk:"failover"`
	GeoLocation             fwtypes.ObjectValueOf[geoLocationModel]                  `tfsdk:"geolocation"`
	GeoProximityLocation    fwtypes.ObjectValueOf[geoProximityLocationModelReadonly] `tfsdk:"geoproximity_location"`
	HealthCheckID           types.String                                             `tfsdk:"health_check_id"`
	MultiValueAnswer        types.Bool                                               `tfsdk:"multi_value_answer"`
	Name                    types.String                                             `tfsdk:"name"`
	Region                  fwtypes.StringEnum[awstypes.ResourceRecordSetRegion]     `tfsdk:"region"`
	ResourceRecords         fwtypes.ListNestedObjectValueOf[resourceRecordModel]     `tfsdk:"resource_records"`
	SetIdentifier           types.String                                             `tfsdk:"set_identifier"`
	TrafficPolicyInstanceID types.String                                             `tfsdk:"traffic_policy_instance_id"`
	TTL                     types.Int64                                              `tfsdk:"ttl"`
	Type                    fwtypes.StringEnum[awstypes.RRType]                      `tfsdk:"type"`
	Weight                  types.Int64                                              `tfsdk:"weight"`
}

type geoProximityLocationModelReadonly struct {
	AWSRegion      types.String                            `tfsdk:"aws_region"`
	Bias           types.Int64                             `tfsdk:"bias"`
	Coordinates    fwtypes.ObjectValueOf[coordinatesModel] `tfsdk:"coordinates"`
	LocalZoneGroup types.String                            `tfsdk:"local_zone_group"`
}
