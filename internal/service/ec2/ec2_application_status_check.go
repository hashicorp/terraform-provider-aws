// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tflistplanmodifier "github.com/hashicorp/terraform-provider-aws/internal/framework/planmodifiers/listplanmodifier"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	applicationStatusCheckDefaultThreshold                        = 2
	applicationStatusCheckDefaultInitializationGracePeriodSeconds = 300
	applicationStatusCheckDefaultInterval                         = 60
	applicationStatusCheckDefaultTimeout                          = 6
)

// @FrameworkResource("aws_ec2_application_status_check", name="Application Status Check")
// @Tags(identifierAttribute="id")
// @IdentityAttribute("id")
// @Testing(generator=false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckApplicationStatusCheck")
func newApplicationStatusCheckResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &applicationStatusCheckResource{}, nil
}

type applicationStatusCheckResource struct {
	framework.ResourceWithModel[applicationStatusCheckResourceModel]
	framework.WithImportByIdentity
}

func (r *applicationStatusCheckResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	aggregationType := fwtypes.StringEnumType[awstypes.AggregationStatusEnum]()

	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"aggregation": schema.StringAttribute{
				CustomType: aggregationType,
				Optional:   true,
				Computed:   true,
				Default:    aggregationType.AttributeDefault(awstypes.AggregationStatusEnumIncluded),
			},
			"device_index": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"failure_threshold": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(applicationStatusCheckDefaultThreshold),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"initialization_grace_period_seconds": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(applicationStatusCheckDefaultInitializationGracePeriodSeconds),
				Validators: []validator.Int64{
					int64validator.Between(1, 600),
				},
			},
			names.AttrInterval: schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(applicationStatusCheckDefaultInterval),
				Validators: []validator.Int64{
					int64validator.OneOf(60),
				},
			},
			"ip_scope": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.IpScopeEnum](),
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString(string(awstypes.IpScopeEnumPrivate)),
			},
			"ip_version": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.IpVersionEnum](),
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString(string(awstypes.IpVersionEnumIpv4)),
			},
			names.AttrPath: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("/"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile("^/"), "must start with a forward slash (/)."),
				},
			},
			names.AttrPort: schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			names.AttrProtocol: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.NetworkProtocolEnum](),
				Required:   true,
			},
			"status_code_matcher": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("200"),
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"success_threshold": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(applicationStatusCheckDefaultThreshold),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrTimeout: schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(applicationStatusCheckDefaultTimeout),
				Validators: []validator.Int64{
					int64validator.Between(1, 30),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"health_check_path": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[applicationStatusCheckHealthCheckPathModel](ctx),
				PlanModifiers: []planmodifier.List{
					tflistplanmodifier.RequiresReplaceIfEmptied,
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						names.AttrDestination: schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[applicationStatusCheckHealthCheckPathEndpointModel](ctx),
							Validators: []validator.Set{
								setvalidator.IsRequired(),
								setvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: applicationStatusCheckHealthCheckPathEndpointSchema(),
							},
						},
						names.AttrSource: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[applicationStatusCheckHealthCheckPathEndpointModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: applicationStatusCheckHealthCheckPathEndpointSchema(),
							},
						},
					},
				},
			},
		},
	}
}

func applicationStatusCheckHealthCheckPathEndpointSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"security_group_id": schema.StringAttribute{
			Required: true,
		},
		names.AttrSubnetID: schema.StringAttribute{
			Required: true,
		},
	}
}

func (r *applicationStatusCheckResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data applicationStatusCheckResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	var input ec2.CreateApplicationStatusCheckInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.TagSpecifications = getTagSpecificationsIn(ctx, awstypes.ResourceTypeApplicationStatusCheck)

	output, err := conn.CreateApplicationStatusCheck(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}
	if output == nil || output.ApplicationStatusCheck == nil || output.ApplicationStatusCheck.ApplicationStatusCheckId == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"))
		return
	}

	data.ID = fwflex.StringToFramework(ctx, output.ApplicationStatusCheck.ApplicationStatusCheckId)
	// Persist the ID so the resource is tainted if the subsequent wait fails.
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID))
	if response.Diagnostics.HasError() {
		return
	}

	applicationStatusCheck, err := waitApplicationStatusCheckCreated(ctx, conn, data.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, applicationStatusCheck, &data))
	if response.Diagnostics.HasError() {
		return
	}
	setTagsOut(ctx, applicationStatusCheck.Tags)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *applicationStatusCheckResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data applicationStatusCheckResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)
	id := data.ID.ValueString()

	applicationStatusCheck, err := findApplicationStatusCheckByID(ctx, conn, id)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, applicationStatusCheck, &data))
	if response.Diagnostics.HasError() {
		return
	}
	setTagsOut(ctx, applicationStatusCheck.Tags)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *applicationStatusCheckResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state applicationStatusCheckResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)
	id := plan.ID.ValueString()

	diff, diags := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input ec2.ModifyApplicationStatusCheckInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input, diff.IgnoredFieldNamesOpts()...))
		if response.Diagnostics.HasError() {
			return
		}
		input.ApplicationStatusCheckId = aws.String(id)

		_, err := conn.ModifyApplicationStatusCheck(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *applicationStatusCheckResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data applicationStatusCheckResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)
	id := data.ID.ValueString()

	_, err := conn.DeleteApplicationStatusCheck(ctx, &ec2.DeleteApplicationStatusCheckInput{
		ApplicationStatusCheckId: aws.String(id),
	})
	if err != nil {
		_, findErr := findApplicationStatusCheckByID(ctx, conn, id)
		if retry.NotFound(findErr) {
			return
		}
		if findErr != nil {
			err = errors.Join(err, findErr)
		}
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	if err := waitApplicationStatusCheckDeleted(ctx, conn, id); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.String())
	}
}

func (r *applicationStatusCheckResource) flatten(ctx context.Context, apiObject *awstypes.ApplicationStatusCheckResponseObject, data *applicationStatusCheckResourceModel) diag.Diagnostics { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	return fwflex.Flatten(ctx, apiObject, data, fwflex.WithFieldNamePrefix("ApplicationStatusCheck"))
}

func findApplicationStatusCheckByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ApplicationStatusCheckResponseObject, error) {
	output, err := findApplicationStatusChecks(ctx, conn, &ec2.DescribeApplicationStatusChecksInput{
		ApplicationStatusCheckIds: []string{id},
	})
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	applicationStatusCheck, err := tfresource.AssertSingleValueResult(tfslices.Filter(output, func(v awstypes.ApplicationStatusCheckResponseObject) bool {
		return aws.ToString(v.ApplicationStatusCheckId) == id
	}))
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if applicationStatusCheck.DeletionTime != nil {
		return nil, smarterr.NewError(&retry.NotFoundError{})
	}

	return applicationStatusCheck, nil
}

func findApplicationStatusChecks(ctx context.Context, conn *ec2.Client, input *ec2.DescribeApplicationStatusChecksInput) ([]awstypes.ApplicationStatusCheckResponseObject, error) {
	var output []awstypes.ApplicationStatusCheckResponseObject

	err := describeApplicationStatusChecksPages(ctx, conn, input, func(page *ec2.DescribeApplicationStatusChecksOutput, _ bool) bool {
		output = append(output, page.ApplicationStatusChecks...)
		return true
	})
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return output, nil
}

func waitApplicationStatusCheckCreated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ApplicationStatusCheckResponseObject, error) {
	output, err := tfresource.RetryWhenNotFound(ctx, ec2PropagationTimeout, func(ctx context.Context) (*awstypes.ApplicationStatusCheckResponseObject, error) {
		return findApplicationStatusCheckByID(ctx, conn, id)
	})
	return output, smarterr.NewError(err)
}

func waitApplicationStatusCheckDeleted(ctx context.Context, conn *ec2.Client, id string) error {
	_, err := tfresource.RetryUntilNotFound(ctx, ec2PropagationTimeout, func(ctx context.Context) (any, error) {
		return findApplicationStatusCheckByID(ctx, conn, id)
	})
	return smarterr.NewError(err)
}

type applicationStatusCheckResourceModel struct {
	framework.WithRegionModel
	Aggregation                      fwtypes.StringEnum[awstypes.AggregationStatusEnum]                          `tfsdk:"aggregation"`
	DeviceIndex                      types.Int64                                                                 `tfsdk:"device_index"`
	FailureThreshold                 types.Int64                                                                 `tfsdk:"failure_threshold"`
	HealthCheckPaths                 fwtypes.ListNestedObjectValueOf[applicationStatusCheckHealthCheckPathModel] `tfsdk:"health_check_path"`
	ID                               types.String                                                                `tfsdk:"id"`
	InitializationGracePeriodSeconds types.Int64                                                                 `tfsdk:"initialization_grace_period_seconds"`
	Interval                         types.Int64                                                                 `tfsdk:"interval"`
	IPScope                          fwtypes.StringEnum[awstypes.IpScopeEnum]                                    `tfsdk:"ip_scope"`
	IPVersion                        fwtypes.StringEnum[awstypes.IpVersionEnum]                                  `tfsdk:"ip_version"`
	Path                             types.String                                                                `tfsdk:"path"`
	Port                             types.Int64                                                                 `tfsdk:"port"`
	Protocol                         fwtypes.StringEnum[awstypes.NetworkProtocolEnum]                            `tfsdk:"protocol"`
	StatusCodeMatcher                types.String                                                                `tfsdk:"status_code_matcher"`
	SuccessThreshold                 types.Int64                                                                 `tfsdk:"success_threshold"`
	Tags                             tftags.Map                                                                  `tfsdk:"tags"`
	TagsAll                          tftags.Map                                                                  `tfsdk:"tags_all"`
	Timeout                          types.Int64                                                                 `tfsdk:"timeout"`
}

type applicationStatusCheckHealthCheckPathModel struct {
	Destinations fwtypes.SetNestedObjectValueOf[applicationStatusCheckHealthCheckPathEndpointModel]  `tfsdk:"destination"`
	Source       fwtypes.ListNestedObjectValueOf[applicationStatusCheckHealthCheckPathEndpointModel] `tfsdk:"source"`
}

type applicationStatusCheckHealthCheckPathEndpointModel struct {
	SecurityGroupID types.String `tfsdk:"security_group_id"`
	SubnetID        types.String `tfsdk:"subnet_id"`
}
