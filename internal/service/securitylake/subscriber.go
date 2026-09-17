// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package securitylake

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_securitylake_subscriber", name="Subscriber")
// @Tags(identifierAttribute="arn")
func newSubscriberResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &subscriberResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type subscriberResource struct {
	framework.ResourceWithModel[subscriberResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *subscriberResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AccessType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN:        framework.ARNAttributeComputedOnly(),
			names.AttrID:         framework.IDAttribute(),
			"resource_share_arn": framework.ARNAttributeComputedOnly(),
			"resource_share_name": schema.StringAttribute{
				Computed: true,
			},
			names.AttrRoleARN: framework.ARNAttributeComputedOnly(),
			"s3_bucket_arn":   framework.ARNAttributeComputedOnly(),
			"subscriber_description": schema.StringAttribute{
				Optional: true,
			},
			"subscriber_endpoint": schema.StringAttribute{
				Computed: true,
			},
			"subscriber_name": schema.StringAttribute{
				Optional: true,
			},
			"subscriber_status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrSource: schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[subscriberLogSourceResourceModel](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"aws_log_source_resource": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[subscriberAWSLogSourceResourceModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"source_name": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.AwsLogSourceName](),
										Required:   true,
									},
									"source_version": schema.StringAttribute{
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"custom_log_source_resource": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrAttributes: framework.ResourceComputedListOfObjectsAttribute[subscriberCustomLogSourceAttributesModel](ctx, listplanmodifier.UseStateForUnknown()),
									"provider":           framework.ResourceComputedListOfObjectsAttribute[subscriberCustomLogSourceProviderModel](ctx, listplanmodifier.UseStateForUnknown()),
									"source_name": schema.StringAttribute{
										Required: true,
									},
									"source_version": schema.StringAttribute{
										Optional: true,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"subscriber_identity": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[subscriberAWSIdentityModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrExternalID: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(2, 1224),
							},
						},
						names.AttrPrincipal: schema.StringAttribute{
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

func (r *subscriberResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data subscriberResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	var input securitylake.CreateSubscriberInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	if !data.AccessType.IsUnknown() && !data.AccessType.IsNull() {
		input.AccessTypes = []awstypes.AccessType{awstypes.AccessType(data.AccessType.ValueString())}
	}
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateSubscriber(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Security Lake Subscriber", err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, output.Subscriber.SubscriberId)
	data.SubscriberARN = fwflex.StringToFramework(ctx, output.Subscriber.SubscriberArn)

	subscriber, err := waitSubscriberCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Security Lake Subscriber (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns after creation is complete.
	response.Diagnostics.Append(fwflex.Flatten(ctx, subscriber, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields
	data.AccessType = fwtypes.StringEnumValue(subscriber.AccessTypes[0])

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *subscriberResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)
	var data subscriberResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findSubscriberByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Lake Subscriber (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields
	data.AccessType = fwtypes.StringEnumValue(output.AccessTypes[0])

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *subscriberResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new subscriberResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	if !new.AccessType.Equal(old.AccessType) ||
		!new.SubscriberDescription.Equal(old.SubscriberDescription) ||
		!new.SubscriberName.Equal(old.SubscriberName) ||
		!new.SubscriberIdentity.Equal(old.SubscriberIdentity) ||
		!new.Sources.Equal(old.Sources) {
		var input securitylake.UpdateSubscriberInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.SubscriberId = fwflex.StringFromFramework(ctx, new.ID)

		_, err := conn.UpdateSubscriber(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Security Lake Subscriber (%s)", new.ID.ValueString()), err.Error())

			return
		}

		subscriber, err := waitSubscriberUpdated(ctx, conn, new.ID.ValueString(), r.CreateTimeout(ctx, new.Timeouts))
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Security Lake Subscriber (%s) update", new.ID.ValueString()), err.Error())

			return
		}

		// Set values for unknowns after update is complete.
		response.Diagnostics.Append(fwflex.Flatten(ctx, subscriber, &new)...)
		if response.Diagnostics.HasError() {
			return
		}
	} else {
		new.ResourceShareName = old.ResourceShareName
		new.SubscriberEndpoint = old.SubscriberEndpoint
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *subscriberResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data subscriberResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	input := securitylake.DeleteSubscriberInput{
		SubscriberId: data.ID.ValueStringPointer(),
	}
	_, err := conn.DeleteSubscriber(ctx, &input)

	// No Subscriber:
	// "An error occurred (AccessDeniedException) when calling the DeleteSubscriber operation: User: ... is not authorized to perform: securitylake:GetSubscriber", or
	// "UnauthorizedException: Unauthorized"
	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "is not authorized to perform") ||
		tfawserr.ErrMessageContains(err, errCodeUnauthorizedException, "Unauthorized") {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Security Lake Subscriber (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err = waitSubscriberDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Security Lake Subscriber (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func findSubscriberByID(ctx context.Context, conn *securitylake.Client, id string) (*awstypes.SubscriberResource, error) {
	input := &securitylake.GetSubscriberInput{
		SubscriberId: aws.String(id),
	}

	output, err := conn.GetSubscriber(ctx, input)

	// No Subscriber:
	// "An error occurred (AccessDeniedException) when calling the DeleteSubscriber operation: User: ... is not authorized to perform: securitylake:GetSubscriber", or
	// "UnauthorizedException: Unauthorized"
	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "is not authorized to perform") ||
		tfawserr.ErrMessageContains(err, errCodeUnauthorizedException, "Unauthorized") {
		return nil, tfresource.NewEmptyResultError()
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Subscriber, nil
}

func statusSubscriber(conn *securitylake.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSubscriberByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.SubscriberStatus), nil
	}
}

func waitSubscriberCreated(ctx context.Context, conn *securitylake.Client, id string, timeout time.Duration) (*awstypes.SubscriberResource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SubscriberStatusPending),
		Target:  enum.Slice(awstypes.SubscriberStatusActive, awstypes.SubscriberStatusReady),
		Refresh: statusSubscriber(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SubscriberResource); ok {
		return output, err
	}

	return nil, err
}

func waitSubscriberUpdated(ctx context.Context, conn *securitylake.Client, id string, timeout time.Duration) (*awstypes.SubscriberResource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SubscriberStatusPending),
		Target:  enum.Slice(awstypes.SubscriberStatusActive, awstypes.SubscriberStatusReady),
		Refresh: statusSubscriber(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SubscriberResource); ok {
		return output, err
	}

	return nil, err
}

func waitSubscriberDeleted(ctx context.Context, conn *securitylake.Client, id string, timeout time.Duration) (*awstypes.SubscriberResource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SubscriberStatusActive, awstypes.SubscriberStatusReady, awstypes.SubscriberStatusDeactivated),
		Target:  []string{},
		Refresh: statusSubscriber(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SubscriberResource); ok {
		return output, err
	}

	return nil, err
}

type subscriberResourceModel struct {
	framework.WithRegionModel
	AccessType            fwtypes.StringEnum[awstypes.AccessType]                          `tfsdk:"access_type" autoflex:"-"`
	ID                    types.String                                                     `tfsdk:"id"`
	ResourceShareARN      types.String                                                     `tfsdk:"resource_share_arn" autoflex:",legacy"`
	ResourceShareName     types.String                                                     `tfsdk:"resource_share_name"`
	RoleARN               types.String                                                     `tfsdk:"role_arn"`
	S3BucketARN           types.String                                                     `tfsdk:"s3_bucket_arn"`
	Sources               fwtypes.SetNestedObjectValueOf[subscriberLogSourceResourceModel] `tfsdk:"source"`
	SubscriberARN         types.String                                                     `tfsdk:"arn"`
	SubscriberDescription types.String                                                     `tfsdk:"subscriber_description"`
	SubscriberEndpoint    types.String                                                     `tfsdk:"subscriber_endpoint"`
	SubscriberIdentity    fwtypes.ListNestedObjectValueOf[subscriberAWSIdentityModel]      `tfsdk:"subscriber_identity"`
	SubscriberName        types.String                                                     `tfsdk:"subscriber_name"`
	SubscriberStatus      types.String                                                     `tfsdk:"subscriber_status"`
	Tags                  tftags.Map                                                       `tfsdk:"tags"`
	TagsAll               tftags.Map                                                       `tfsdk:"tags_all"`
	Timeouts              timeouts.Value                                                   `tfsdk:"timeouts"`
}

type subscriberLogSourceResourceModel struct {
	AWSLogSource    fwtypes.ListNestedObjectValueOf[subscriberAWSLogSourceResourceModel]    `tfsdk:"aws_log_source_resource"`
	CustomLogSource fwtypes.ListNestedObjectValueOf[subscriberCustomLogSourceResourceModel] `tfsdk:"custom_log_source_resource"`
}

var (
	_ fwflex.Expander  = subscriberLogSourceResourceModel{}
	_ fwflex.Flattener = &subscriberLogSourceResourceModel{}
)

func (m subscriberLogSourceResourceModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var result any
	var diags diag.Diagnostics

	switch {
	case !m.AWSLogSource.IsNull():
		awsLogSource, d := m.AWSLogSource.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.LogSourceResourceMemberAwsLogSource
		diags.Append(fwflex.Expand(ctx, awsLogSource, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	case !m.CustomLogSource.IsNull():
		customLogSource, d := m.CustomLogSource.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.LogSourceResourceMemberCustomLogSource
		diags.Append(fwflex.Expand(ctx, customLogSource, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	}

	return result, diags
}

func (m *subscriberLogSourceResourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch v := v.(type) {
	case awstypes.LogSourceResourceMemberAwsLogSource:
		var awsLogSource subscriberAWSLogSourceResourceModel
		diags.Append(fwflex.Flatten(ctx, v.Value, &awsLogSource)...)
		if diags.HasError() {
			return diags
		}

		m.AWSLogSource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &awsLogSource)
	case awstypes.LogSourceResourceMemberCustomLogSource:
		var customLogSource subscriberCustomLogSourceResourceModel
		diags.Append(fwflex.Flatten(ctx, v.Value, &customLogSource)...)
		if diags.HasError() {
			return diags
		}

		m.CustomLogSource = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &customLogSource)
	}

	return diags
}

type subscriberAWSLogSourceResourceModel struct {
	SourceName    fwtypes.StringEnum[awstypes.AwsLogSourceName] `tfsdk:"source_name"`
	SourceVersion types.String                                  `tfsdk:"source_version"`
}

type subscriberCustomLogSourceResourceModel struct {
	Attributes    fwtypes.ListNestedObjectValueOf[subscriberCustomLogSourceAttributesModel] `tfsdk:"attributes"`
	Provider      fwtypes.ListNestedObjectValueOf[subscriberCustomLogSourceProviderModel]   `tfsdk:"provider"`
	SourceName    types.String                                                              `tfsdk:"source_name"`
	SourceVersion types.String                                                              `tfsdk:"source_version"`
}

type subscriberCustomLogSourceAttributesModel struct {
	CrawlerARN  fwtypes.ARN `tfsdk:"crawler_arn"`
	DatabaseARN fwtypes.ARN `tfsdk:"database_arn"`
	TableARN    fwtypes.ARN `tfsdk:"table_arn"`
}

type subscriberCustomLogSourceProviderModel struct {
	Location types.String `tfsdk:"location"`
	RoleARN  fwtypes.ARN  `tfsdk:"role_arn"`
}

type subscriberAWSIdentityModel struct {
	ExternalID types.String `tfsdk:"external_id"`
	Principal  types.String `tfsdk:"principal"`
}
