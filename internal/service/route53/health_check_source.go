// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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
	lazystringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator/lazy"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_route53_health_check_source", name="Health Check Source")
// @Tags(identifierAttribute="id", resourceType="healthchecksource")
// @IdentityAttribute("id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/route53/types;awstypes;awstypes.HealthCheckSource")
// @Testing(hasNoPreExistingResource=true)
func newHealthCheckSourceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &healthCheckSourceResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

// There is no API to change a health check source, so every argument forces
// replacement. The embedded no-op Update still applies tag changes in place.
type healthCheckSourceResource struct {
	framework.ResourceWithModel[healthCheckSourceResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *healthCheckSourceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.HealthCheckSourceStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVPCID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"subnet_configuration": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[subnetConfigurationModel](ctx),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					setvalidator.SizeBetween(2, 3),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"availability_zone_id": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								lazystringvalidator.RegexMatches(`^[a-z0-9]+-az[0-9]+$`, "must be an Availability Zone ID, for example use1-az1"),
							},
						},
						names.AttrSecurityGroupIDs: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Required:    true,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
						names.AttrSubnetID: schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *healthCheckSourceResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var data healthCheckSourceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	if data.SubnetConfigurationList.IsNull() || data.SubnetConfigurationList.IsUnknown() {
		return
	}

	subnets, diags := data.SubnetConfigurationList.ToSlice(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	if response.Diagnostics.HasError() {
		return
	}

	seen := make(map[string]bool, len(subnets))
	for _, subnet := range subnets {
		if subnet.AvailabilityZoneID.IsNull() || subnet.AvailabilityZoneID.IsUnknown() {
			continue
		}

		azID := subnet.AvailabilityZoneID.ValueString()
		if seen[azID] {
			response.Diagnostics.AddAttributeError(
				path.Root("subnet_configuration"),
				"Duplicate Availability Zone",
				fmt.Sprintf("Each subnet_configuration must be in a different Availability Zone; %q is used more than once.", azID),
			)
			return
		}
		seen[azID] = true
	}
}

func (r *healthCheckSourceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data healthCheckSourceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53Client(ctx)

	var config awstypes.HealthCheckSourceConfig
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &config))
	if response.Diagnostics.HasError() {
		return
	}

	input := route53.CreateHealthCheckSourceInput{
		CallerReference:         aws.String(create.UniqueId(ctx)),
		HealthCheckSourceConfig: &config,
		Tags:                    getTagsIn(ctx),
	}

	output, err := conn.CreateHealthCheckSource(ctx, &input)

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	id := aws.ToString(output.HealthCheckSourceId)
	data.ID = fwflex.StringValueToFramework(ctx, id)
	data.ARN = fwflex.StringValueToFramework(ctx, r.healthCheckSourceARN(ctx, id))

	// Save the ID so that a source that fails to become ready is tainted, not orphaned.
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID))
	if response.Diagnostics.HasError() {
		return
	}

	source, err := waitHealthCheckSourceCreated(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	data.Status = fwtypes.StringEnumValue(source.HealthCheckSourceStatus)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *healthCheckSourceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data healthCheckSourceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53Client(ctx)

	id := data.ID.ValueString()
	source, err := findHealthCheckSourceByID(ctx, conn, id)

	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err), smerr.ID, id)
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, source.HealthCheckSourceConfig, &data), smerr.ID, id)
	if response.Diagnostics.HasError() {
		return
	}

	data.ARN = fwflex.StringValueToFramework(ctx, r.healthCheckSourceARN(ctx, id))
	data.Status = fwtypes.StringEnumValue(source.HealthCheckSourceStatus)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *healthCheckSourceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data healthCheckSourceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53Client(ctx)

	id := data.ID.ValueString()
	input := route53.DeleteHealthCheckSourceInput{
		HealthCheckSourceId: aws.String(id),
	}

	deleteTimeout := r.DeleteTimeout(ctx, data.Timeouts)

	// Route 53 rejects the delete with InvalidInput while a creation is in
	// progress, so let any creation finish first.
	source, err := findHealthCheckSourceByID(ctx, conn, id)

	if retry.NotFound(err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	switch source.HealthCheckSourceStatus {
	case awstypes.HealthCheckSourceStatusCreatePending, awstypes.HealthCheckSourceStatusCreateInProgress:
		if _, err := waitHealthCheckSourceCreated(ctx, conn, id, deleteTimeout); err != nil && !errs.IsA[*retry.UnexpectedStateError](err) {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}
	}

	// Route 53 also rejects the delete with InvalidInput while a health check
	// references the source, which is briefly the case after that health check
	// is deleted. Retry for a short time only, so that a source still in use
	// fails fast.
	_, err = tfresource.RetryWhenIsOneOf2[any, *awstypes.InvalidInput, *awstypes.PriorRequestNotComplete](ctx, healthCheckSourceInUseTimeout, func(ctx context.Context) (any, error) {
		return conn.DeleteHealthCheckSource(ctx, &input)
	})

	if errs.IsA[*awstypes.NoSuchHealthCheckSource](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	if _, err := waitHealthCheckSourceDeleted(ctx, conn, id, deleteTimeout); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}
}

const healthCheckSourceInUseTimeout = 2 * time.Minute

func (r *healthCheckSourceResource) healthCheckSourceARN(ctx context.Context, id string) string {
	return r.Meta().GlobalARNNoAccount(ctx, "route53", "healthchecksource/"+id)
}

type healthCheckSourceResourceModel struct {
	ARN                     types.String                                             `tfsdk:"arn"`
	ID                      types.String                                             `tfsdk:"id"`
	Status                  fwtypes.StringEnum[awstypes.HealthCheckSourceStatus]     `tfsdk:"status"`
	SubnetConfigurationList fwtypes.SetNestedObjectValueOf[subnetConfigurationModel] `tfsdk:"subnet_configuration"`
	Tags                    tftags.Map                                               `tfsdk:"tags"`
	TagsAll                 tftags.Map                                               `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                           `tfsdk:"timeouts"`
	VPCID                   types.String                                             `tfsdk:"vpc_id"`
}

type subnetConfigurationModel struct {
	AvailabilityZoneID  types.String        `tfsdk:"availability_zone_id"`
	SecurityGroupIDList fwtypes.SetOfString `tfsdk:"security_group_ids"`
	SubnetID            types.String        `tfsdk:"subnet_id"`
}

func findHealthCheckSourceByID(ctx context.Context, conn *route53.Client, id string) (*awstypes.HealthCheckSource, error) {
	input := route53.GetHealthCheckSourceInput{
		HealthCheckSourceId: aws.String(id),
	}

	output, err := conn.GetHealthCheckSource(ctx, &input)

	if errs.IsA[*awstypes.NoSuchHealthCheckSource](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.HealthCheckSource == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	if output.HealthCheckSource.HealthCheckSourceStatus == awstypes.HealthCheckSourceStatusDeleteCompleted {
		return nil, smarterr.NewError(&retry.NotFoundError{
			Message: string(output.HealthCheckSource.HealthCheckSourceStatus),
		})
	}

	return output.HealthCheckSource, nil
}

func findHealthCheckSources(ctx context.Context, conn *route53.Client) ([]awstypes.HealthCheckSource, error) {
	var input route53.ListHealthCheckSourcesInput
	var output []awstypes.HealthCheckSource

	pages := route53.NewListHealthCheckSourcesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, page.HealthCheckSources...)
	}

	return output, nil
}

func statusHealthCheckSource(conn *route53.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findHealthCheckSourceByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.HealthCheckSourceStatus), nil
	}
}

func waitHealthCheckSourceCreated(ctx context.Context, conn *route53.Client, id string, timeout time.Duration) (*awstypes.HealthCheckSource, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.HealthCheckSourceStatusCreatePending, awstypes.HealthCheckSourceStatusCreateInProgress),
		Target:                    enum.Slice(awstypes.HealthCheckSourceStatusCreateCompleted),
		Refresh:                   statusHealthCheckSource(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.HealthCheckSource); ok {
		retry.SetLastError(err, healthCheckSourceStatusError(output))

		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitHealthCheckSourceDeleted(ctx context.Context, conn *route53.Client, id string, timeout time.Duration) (*awstypes.HealthCheckSource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.HealthCheckSourceStatusDeletePending, awstypes.HealthCheckSourceStatusDeleteInProgress),
		Target:  []string{},
		Refresh: statusHealthCheckSource(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.HealthCheckSource); ok {
		retry.SetLastError(err, healthCheckSourceStatusError(output))

		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// healthCheckSourceStatusError returns the reason that Route 53 gives for a
// CREATE_FAILED or DELETE_FAILED status, or nil if there is none.
func healthCheckSourceStatusError(source *awstypes.HealthCheckSource) error {
	if source.StatusReasonCode == "" && source.StatusReason == nil {
		return nil
	}

	return fmt.Errorf("%s: %s", source.StatusReasonCode, aws.ToString(source.StatusReason))
}
