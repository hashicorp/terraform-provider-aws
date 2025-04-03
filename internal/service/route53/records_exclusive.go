// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_route53_records_exclusive", name="Records Exclusive")
func newResourceRecordsExclusive(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceRecordsExclusive{}

	r.SetDefaultCreateTimeout(45 * time.Minute)
	r.SetDefaultUpdateTimeout(45 * time.Minute)

	return r, nil
}

const (
	ResNameRecordsExclusive = "Records Exclusive"
)

type resourceRecordsExclusive struct {
	framework.ResourceWithConfigure
	framework.WithNoOpDelete
	framework.WithTimeouts
}

func (r *resourceRecordsExclusive) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"zone_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"resource_record_set": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[resourceRecordSetModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"failover": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ResourceRecordSetFailover](),
							Optional:   true,
						},
						"health_check_id": schema.StringAttribute{
							Optional: true,
						},
						"multi_value_answer": schema.BoolAttribute{
							Optional: true,
						},
						names.AttrName: schema.StringAttribute{
							CustomType: fwtypes.DNSNameStringType,
							Required:   true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(1024),
							},
						},
						names.AttrRegion: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ResourceRecordSetRegion](),
							Optional:   true,
						},
						"set_identifier": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 128),
							},
						},
						"traffic_policy_instance_id": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 36),
							},
						},
						"ttl": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.Between(0, 2147483647),
							},
						},
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.RRType](),
							Optional:   true,
						},
						names.AttrWeight: schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.Between(0, 255),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"alias_target": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[aliasTargetModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrDNSName: schema.StringAttribute{
										CustomType: fwtypes.DNSNameStringType,
										Required:   true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(1024),
										},
									},
									"evaluate_target_health": schema.BoolAttribute{
										Required: true,
									},
									names.AttrHostedZoneID: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(32),
										},
									},
								},
							},
						},
						"cidr_routing_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cidrRoutingConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"collection_id": schema.StringAttribute{
										Required: true,
									},
									"location_name": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 16),
										},
									},
								},
							},
						},
						"geolocation": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[geoLocationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"continent_code": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(2, 2),
										},
									},
									"country_code": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 2),
										},
									},
									"subdivision_code": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 3),
										},
									},
								},
							},
						},
						"geoproximity_location": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[geoProximityLocationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"aws_region": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 64),
										},
									},
									"bias": schema.Int64Attribute{
										Optional: true,
										Validators: []validator.Int64{
											int64validator.Between(-99, 99),
										},
									},
									"local_zone_group": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 64),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"coordinates": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[coordinatesModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"latitude": schema.StringAttribute{
													Required: true,
												},
												"longitude": schema.StringAttribute{
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"resource_records": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[resourceRecordModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrValue: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(4000),
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *resourceRecordsExclusive) syncRecordSets(ctx context.Context, plan resourceRecordsExclusiveModel, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := r.Meta().Route53Client(ctx)

	have, err := findResourceRecordSetsForHostedZone(ctx, conn, plan.ZoneID.ValueString())
	if err != nil {
		diags.AddError(
			create.ProblemStandardMessage(names.Route53, "Syncronizing", ResNameRecordsExclusive, plan.ZoneID.String(), err),
			err.Error(),
		)
		return diags
	}

	var want []awstypes.ResourceRecordSet
	diags.Append(flex.Expand(ctx, plan.ResourceRecordSet, &want)...)
	if diags.HasError() {
		return diags
	}

	// Amazon Route 53 can update an existing resource record set only when all
	// of the following values match: Name, Type and SetIdentifier.
	// Ref: http://docs.aws.amazon.com/Route53/latest/APIReference/API_ChangeResourceRecordSets.html.
	add, remove, modify, _ := intflex.DiffSlicesWithModify(have, want, resourceRecordSetEqual, resourceRecordSetIdentifiersEqual)

	var changes []awstypes.Change
	for _, r := range remove {
		changes = append(changes, awstypes.Change{
			Action:            awstypes.ChangeActionDelete,
			ResourceRecordSet: &r,
		})
	}
	for _, r := range add {
		changes = append(changes, awstypes.Change{
			Action:            awstypes.ChangeActionCreate,
			ResourceRecordSet: &r,
		})
	}
	for _, r := range modify {
		changes = append(changes, awstypes.Change{
			Action:            awstypes.ChangeActionUpsert,
			ResourceRecordSet: &r,
		})
	}

	if len(changes) > 0 {
		input := route53.ChangeResourceRecordSetsInput{
			HostedZoneId: plan.ZoneID.ValueStringPointer(),
			ChangeBatch: &awstypes.ChangeBatch{
				Changes: changes,
			},
		}
		out, err := conn.ChangeResourceRecordSets(ctx, &input)
		if err != nil {
			diags.AddError(
				create.ProblemStandardMessage(names.Route53, create.ErrActionCreating, ResNameRecordsExclusive, plan.ZoneID.String(), err),
				err.Error(),
			)
			return diags
		}
		if out == nil || out.ChangeInfo == nil || out.ChangeInfo.Id == nil {
			diags.AddError(
				create.ProblemStandardMessage(names.Route53, create.ErrActionCreating, ResNameRecordsExclusive, plan.ZoneID.String(), nil),
				errors.New("empty output").Error(),
			)
			return diags
		}

		if _, err := waitChangeInsync(ctx, conn, aws.ToString(out.ChangeInfo.Id), timeout); err != nil {
			diags.AddError(
				create.ProblemStandardMessage(names.Route53, create.ErrActionWaitingForCreation, ResNameRecordsExclusive, plan.ZoneID.String(), nil),
				err.Error(),
			)
			return diags
		}
	}

	return diags
}

func (r *resourceRecordsExclusive) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceRecordsExclusiveModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.syncRecordSets(ctx, plan, r.CreateTimeout(ctx, plan.Timeouts))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceRecordsExclusive) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceRecordsExclusiveModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53Client(ctx)
	output, err := findResourceRecordSetsForHostedZone(ctx, conn, state.ZoneID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53, create.ErrActionReading, ResNameRecordsExclusive, state.ZoneID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(
		ctx,
		struct {
			ResourceRecordSets []awstypes.ResourceRecordSet
		}{
			ResourceRecordSets: output,
		},
		&state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceRecordsExclusive) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceRecordsExclusiveModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.syncRecordSets(ctx, plan, r.UpdateTimeout(ctx, plan.Timeouts))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceRecordsExclusive) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("zone_id"), req, resp)
}

func findResourceRecordSetsForHostedZone(ctx context.Context, conn *route53.Client, zoneID string) ([]awstypes.ResourceRecordSet, error) {
	hostedZone, err := findHostedZoneByID(ctx, conn, zoneID)
	if err != nil {
		return nil, err
	}
	hostedZoneName := aws.ToString(hostedZone.HostedZone.Name)

	input := route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	}
	morePages := tfslices.PredicateTrue[*route53.ListResourceRecordSetsOutput]() // exhaust all pages
	filter := func(v *awstypes.ResourceRecordSet) bool {
		// Zone NS & SOA records are created by default and not managed by this resource
		if normalizeDomainName(v.Name) == normalizeDomainName(hostedZoneName) && (v.Type == awstypes.RRTypeNs || v.Type == awstypes.RRTypeSoa) {
			return false
		}
		return true
	}

	return findResourceRecordSets(ctx, conn, &input, morePages, filter)
}

type resourceRecordsExclusiveModel struct {
	ResourceRecordSet fwtypes.SetNestedObjectValueOf[resourceRecordSetModel] `tfsdk:"resource_record_set"`
	Timeouts          timeouts.Value                                         `tfsdk:"timeouts"`
	ZoneID            types.String                                           `tfsdk:"zone_id"`
}

type resourceRecordSetModel struct {
	AliasTarget             fwtypes.ListNestedObjectValueOf[aliasTargetModel]          `tfsdk:"alias_target"`
	CIDRRoutingConfig       fwtypes.ListNestedObjectValueOf[cidrRoutingConfigModel]    `tfsdk:"cidr_routing_config"`
	Failover                fwtypes.StringEnum[awstypes.ResourceRecordSetFailover]     `tfsdk:"failover"`
	GeoLocation             fwtypes.ListNestedObjectValueOf[geoLocationModel]          `tfsdk:"geolocation"`
	GeoProximityLocation    fwtypes.ListNestedObjectValueOf[geoProximityLocationModel] `tfsdk:"geoproximity_location"`
	HealthCheckID           types.String                                               `tfsdk:"health_check_id"`
	MultiValueAnswer        types.Bool                                                 `tfsdk:"multi_value_answer"`
	Name                    fwtypes.DNSNameString                                      `tfsdk:"name"`
	Region                  fwtypes.StringEnum[awstypes.ResourceRecordSetRegion]       `tfsdk:"region"`
	ResourceRecords         fwtypes.ListNestedObjectValueOf[resourceRecordModel]       `tfsdk:"resource_records"`
	SetIdentifier           types.String                                               `tfsdk:"set_identifier"`
	TrafficPolicyInstanceID types.String                                               `tfsdk:"traffic_policy_instance_id"`
	TTL                     types.Int64                                                `tfsdk:"ttl"`
	Type                    fwtypes.StringEnum[awstypes.RRType]                        `tfsdk:"type"`
	Weight                  types.Int64                                                `tfsdk:"weight"`
}

type aliasTargetModel struct {
	DNSName              fwtypes.DNSNameString `tfsdk:"dns_name"`
	EvaluateTargetHealth types.Bool            `tfsdk:"evaluate_target_health"`
	HostedZoneID         types.String          `tfsdk:"hosted_zone_id"`
}

type cidrRoutingConfigModel struct {
	CollectionID types.String `tfsdk:"collection_id"`
	LocationName types.String `tfsdk:"location_name"`
}

type geoLocationModel struct {
	ContinentCode   types.String `tfsdk:"continent_code"`
	CountryCode     types.String `tfsdk:"country_code"`
	SubdivisionCode types.String `tfsdk:"subdivision_code"`
}

type geoProximityLocationModel struct {
	AWSRegion      types.String                                      `tfsdk:"aws_region"`
	Bias           types.Int64                                       `tfsdk:"bias"`
	Coordinates    fwtypes.ListNestedObjectValueOf[coordinatesModel] `tfsdk:"coordinates"`
	LocalZoneGroup types.String                                      `tfsdk:"local_zone_group"`
}

type coordinatesModel struct {
	Latitude  types.String `tfsdk:"latitude"`
	Longitude types.String `tfsdk:"longitude"`
}

type resourceRecordModel struct {
	Value types.String `tfsdk:"value"`
}
