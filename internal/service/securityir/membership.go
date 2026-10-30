// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityir

import (
	"context"
	"iter"
	"slices"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityir"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityir/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_securityir_membership", name="Membership")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/securityir;securityir.GetMembershipOutput")
// @Testing(existsTakesT=true, destroyTakesT=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(preCheck="testAccPreCheck")
func newMembershipResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &membershipResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

// UpdateMembership accepts at most this many OUs in each of the add and remove lists.
const organizationalUnitsUpdateBatchSize = 5

const (
	membershipAccountsConfigurationStatusPending = "Pending"
	membershipAccountsConfigurationStatusUpdated = "Updated"
)

type membershipResource struct {
	framework.ResourceWithModel[membershipResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *membershipResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAccountID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"cover_entire_organization": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"customer_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CustomerType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"membership_activation_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"membership_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 50),
				},
			},
			"membership_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.MembershipStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"number_of_accounts_covered": schema.Int64Attribute{
				Computed: true,
			},
			"organizational_units": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexache.MustCompile(`^ou-[0-9a-z]{4,32}-[a-z0-9]{8,32}$`), "must be a valid organizational unit ID"),
					),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"incident_response_team": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[incidentResponderModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeBetween(2, 10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"communication_preferences": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringEnumType[awstypes.CommunicationType](),
							ElementType: fwtypes.StringEnumType[awstypes.CommunicationType](),
							Optional:    true,
						},
						names.AttrEmail: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(6, 254),
							},
						},
						"job_title": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 50),
							},
						},
						names.AttrName: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(3, 50),
							},
						},
					},
				},
			},
			"opt_in_feature": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[optInFeatureModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtMost(2),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"feature_name": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.OptInFeatureName](),
							Required:   true,
						},
						"is_enabled": schema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *membershipResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data membershipResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityIRClient(ctx)

	name := data.MembershipName.ValueString()
	var input securityir.CreateMembershipInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateMembership(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	id := aws.ToString(output.MembershipId)
	data.ID = types.StringValue(id)

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	membership, err := waitMembershipCreated(ctx, conn, id, createTimeout)
	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	// CreateMembership does not accept organizational units; they can only be added by UpdateMembership.
	if ous := fwflex.ExpandFrameworkStringValueSet(ctx, data.OrganizationalUnits); len(ous) > 0 {
		if err := updateMembershipOrganizationalUnits(ctx, conn, id, nil, ous); err != nil {
			response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID)
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}

		membership, err = waitMembershipAccountsConfigurationUpdated(ctx, conn, id, data.CoverEntireOrganization.ValueBool(), ous, createTimeout)
		if err != nil {
			response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID)
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, membership, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *membershipResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var data membershipResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	features, d := data.OptInFeatures.ToSlice(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	for _, v := range features {
		if v.FeatureName.ValueEnum() == awstypes.OptInFeatureNameTriage && !v.IsEnabled.IsUnknown() && !v.IsEnabled.ValueBool() {
			response.Diagnostics.AddAttributeError(
				path.Root("opt_in_feature"),
				"Invalid Attribute Value",
				"The Triage opt-in feature cannot be disabled; is_enabled must be true.",
			)
		}
	}
}

func (r *membershipResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data membershipResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityIRClient(ctx)

	output, err := findMembershipByID(ctx, conn, data.ID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *membershipResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state membershipResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityIRClient(ctx)
	id := plan.ID.ValueString()

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input securityir.UpdateMembershipInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Membership")))
		if response.Diagnostics.HasError() {
			return
		}

		if !plan.CoverEntireOrganization.Equal(state.CoverEntireOrganization) {
			input.MembershipAccountsConfigurationsUpdate = &awstypes.MembershipAccountsConfigurationsUpdate{
				CoverEntireOrganization: plan.CoverEntireOrganization.ValueBoolPointer(),
			}
		}

		if _, err := conn.UpdateMembership(ctx, &input); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}

		newOUs := fwflex.ExpandFrameworkStringValueSet(ctx, plan.OrganizationalUnits)
		ousChanged := !plan.OrganizationalUnits.Equal(state.OrganizationalUnits)
		if ousChanged {
			oldOUs := fwflex.ExpandFrameworkStringValueSet(ctx, state.OrganizationalUnits)
			if err := updateMembershipOrganizationalUnits(ctx, conn, id, oldOUs, newOUs); err != nil {
				smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
				return
			}
		}

		// Account coverage changes are applied asynchronously.
		if ousChanged || !plan.CoverEntireOrganization.Equal(state.CoverEntireOrganization) {
			if _, err := waitMembershipAccountsConfigurationUpdated(ctx, conn, id, plan.CoverEntireOrganization.ValueBool(), newOUs, r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
				smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
				return
			}
		}
	}

	// UpdateMembership returns no membership details, and number_of_accounts_covered is unknown
	// in the plan after any change, so always refresh.
	membership, err := findMembershipByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, membership, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *membershipResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data membershipResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityIRClient(ctx)
	id := data.ID.ValueString()

	input := securityir.CancelMembershipInput{
		MembershipId: aws.String(id),
	}
	_, err := conn.CancelMembership(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.SecurityIncidentResponseNotActiveException](err) || errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Membership is not active") {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	if _, err := waitMembershipCancelled(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}
}

func (r *membershipResource) flatten(ctx context.Context, membership *securityir.GetMembershipOutput, data *membershipResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	configuredOptInFeatures := data.OptInFeatures

	diags.Append(fwflex.Flatten(ctx, membership, data, fwflex.WithFieldNamePrefix("Membership"), fwflex.WithIgnoredFieldNamesAppend("Region"))...)
	if diags.HasError() {
		return diags
	}

	// Some opt-in features (for example, Triage) are enabled by default and cannot be disabled,
	// so only the features declared in configuration are tracked.
	optInFeatures, d := filterOptInFeatures(ctx, data.OptInFeatures, configuredOptInFeatures)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	data.OptInFeatures = optInFeatures

	// Account coverage is returned in a nested structure that has no counterpart in the schema.
	var coverEntireOrganization bool
	var ous []string
	if v := membership.MembershipAccountsConfigurations; v != nil {
		coverEntireOrganization = aws.ToBool(v.CoverEntireOrganization)
		ous = v.OrganizationalUnits
	}
	data.CoverEntireOrganization = types.BoolValue(coverEntireOrganization)
	data.OrganizationalUnits = fwflex.FlattenFrameworkStringValueSetOfString(ctx, ous)

	return diags
}

// filterOptInFeatures returns the features in actual whose names appear in configured.
func filterOptInFeatures(ctx context.Context, actual, configured fwtypes.SetNestedObjectValueOf[optInFeatureModel]) (fwtypes.SetNestedObjectValueOf[optInFeatureModel], diag.Diagnostics) {
	if configured.IsNull() || configured.IsUnknown() {
		return fwtypes.NewSetNestedObjectValueOfNull[optInFeatureModel](ctx), nil
	}

	var diags diag.Diagnostics
	wanted, d := configured.ToSlice(ctx)
	diags.Append(d...)
	features, d := actual.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return fwtypes.NewSetNestedObjectValueOfNull[optInFeatureModel](ctx), diags
	}

	features = slices.DeleteFunc(features, func(f *optInFeatureModel) bool {
		return !slices.ContainsFunc(wanted, func(w *optInFeatureModel) bool {
			return w.FeatureName.Equal(f.FeatureName)
		})
	})

	return fwtypes.NewSetNestedObjectValueOfSlice(ctx, features, nil)
}

func statusMembershipAccountsConfiguration(conn *securityir.Client, id string, coverEntireOrganization bool, ous []string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findMembershipByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		var gotCover bool
		var gotOUs []string
		if v := output.MembershipAccountsConfigurations; v != nil {
			gotCover = aws.ToBool(v.CoverEntireOrganization)
			gotOUs = v.OrganizationalUnits
		}

		if gotCover != coverEntireOrganization || len(gotOUs) != len(ous) || slices.ContainsFunc(ous, func(v string) bool { return !slices.Contains(gotOUs, v) }) {
			return output, membershipAccountsConfigurationStatusPending, nil
		}

		return output, membershipAccountsConfigurationStatusUpdated, nil
	}
}

func waitMembershipAccountsConfigurationUpdated(ctx context.Context, conn *securityir.Client, id string, coverEntireOrganization bool, ous []string, timeout time.Duration) (*securityir.GetMembershipOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{membershipAccountsConfigurationStatusPending},
		Target:                    []string{membershipAccountsConfigurationStatusUpdated},
		Refresh:                   statusMembershipAccountsConfiguration(conn, id, coverEntireOrganization, ous),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*securityir.GetMembershipOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// updateMembershipOrganizationalUnits changes the OUs covered by a membership from oldOUs to newOUs,
// splitting the change into batches that UpdateMembership accepts.
func updateMembershipOrganizationalUnits(ctx context.Context, conn *securityir.Client, id string, oldOUs, newOUs []string) error {
	var add, remove []string
	for _, v := range newOUs {
		if !slices.Contains(oldOUs, v) {
			add = append(add, v)
		}
	}
	for _, v := range oldOUs {
		if !slices.Contains(newOUs, v) {
			remove = append(remove, v)
		}
	}

	for len(add) > 0 || len(remove) > 0 {
		update := awstypes.MembershipAccountsConfigurationsUpdate{}
		n := min(len(add), organizationalUnitsUpdateBatchSize)
		update.OrganizationalUnitsToAdd, add = add[:n], add[n:]
		n = min(len(remove), organizationalUnitsUpdateBatchSize)
		update.OrganizationalUnitsToRemove, remove = remove[:n], remove[n:]

		input := securityir.UpdateMembershipInput{
			MembershipAccountsConfigurationsUpdate: &update,
			MembershipId:                           aws.String(id),
		}
		if _, err := conn.UpdateMembership(ctx, &input); err != nil {
			return smarterr.NewError(err)
		}
	}

	return nil
}

func findMembershipByID(ctx context.Context, conn *securityir.Client, id string) (*securityir.GetMembershipOutput, error) {
	input := securityir.GetMembershipInput{
		MembershipId: aws.String(id),
	}

	output, err := conn.GetMembership(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.SecurityIncidentResponseNotActiveException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	// A cancelled or terminated membership cannot be managed and is treated as gone.
	if status := output.MembershipStatus; status != awstypes.MembershipStatusActive {
		return nil, smarterr.NewError(&retry.NotFoundError{
			Message: "status: " + string(status),
		})
	}

	return output, nil
}

func listMemberships(ctx context.Context, conn *securityir.Client, input *securityir.ListMembershipsInput) iter.Seq2[awstypes.ListMembershipItem, error] {
	return func(yield func(awstypes.ListMembershipItem, error) bool) {
		pages := securityir.NewListMembershipsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ListMembershipItem{}, smarterr.NewError(err))
				return
			}

			for _, item := range page.Items {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}

func statusMembership(conn *securityir.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findMembershipByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.MembershipStatus), nil
	}
}

func waitMembershipCreated(ctx context.Context, conn *securityir.Client, id string, timeout time.Duration) (*securityir.GetMembershipOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.MembershipStatusActive),
		Refresh:                   statusMembership(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*securityir.GetMembershipOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitMembershipCancelled(ctx context.Context, conn *securityir.Client, id string, timeout time.Duration) (*securityir.GetMembershipOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.MembershipStatusActive),
		Target:  []string{},
		Refresh: statusMembership(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*securityir.GetMembershipOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type membershipResourceModel struct {
	framework.WithRegionModel
	AccountID                     types.String                                            `tfsdk:"account_id"`
	ARN                           types.String                                            `tfsdk:"arn"`
	CoverEntireOrganization       types.Bool                                              `tfsdk:"cover_entire_organization"`
	CustomerType                  fwtypes.StringEnum[awstypes.CustomerType]               `tfsdk:"customer_type"`
	ID                            types.String                                            `tfsdk:"id"`
	IncidentResponseTeam          fwtypes.ListNestedObjectValueOf[incidentResponderModel] `tfsdk:"incident_response_team"`
	MembershipActivationTimestamp timetypes.RFC3339                                       `tfsdk:"membership_activation_timestamp"`
	MembershipName                types.String                                            `tfsdk:"membership_name"`
	MembershipStatus              fwtypes.StringEnum[awstypes.MembershipStatus]           `tfsdk:"membership_status"`
	NumberOfAccountsCovered       types.Int64                                             `tfsdk:"number_of_accounts_covered"`
	OptInFeatures                 fwtypes.SetNestedObjectValueOf[optInFeatureModel]       `tfsdk:"opt_in_feature"`
	OrganizationalUnits           fwtypes.SetOfString                                     `tfsdk:"organizational_units" autoflex:"-"`
	Tags                          tftags.Map                                              `tfsdk:"tags"`
	TagsAll                       tftags.Map                                              `tfsdk:"tags_all"`
	Timeouts                      timeouts.Value                                          `tfsdk:"timeouts"`
}

type incidentResponderModel struct {
	CommunicationPreferences fwtypes.SetOfStringEnum[awstypes.CommunicationType] `tfsdk:"communication_preferences"`
	Email                    types.String                                        `tfsdk:"email"`
	JobTitle                 types.String                                        `tfsdk:"job_title"`
	Name                     types.String                                        `tfsdk:"name"`
}

type optInFeatureModel struct {
	FeatureName fwtypes.StringEnum[awstypes.OptInFeatureName] `tfsdk:"feature_name"`
	IsEnabled   types.Bool                                    `tfsdk:"is_enabled"`
}
