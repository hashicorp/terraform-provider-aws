// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cognito_user_pool_replica", name="User Pool Replica")
// @Tags(identifierAttribute="user_pool_arn")
// @IdentityAttribute("user_pool_id")
// @IdentityAttribute("region_name")
// @ImportIDHandler(userPoolReplicaImportID)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types;awstypes;awstypes.UserPoolReplicaType")
// @Testing(tagsTest=false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(identityTest=false)
func newUserPoolReplicaResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &userPoolReplicaResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type userPoolReplicaResource struct {
	framework.ResourceWithModel[userPoolReplicaResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *userPoolReplicaResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrUserPoolID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.UpdateReplicaStatusType](),
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString(string(awstypes.UpdateReplicaStatusTypeInactive)),
			},
			"role": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ReplicaRoleType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_pool_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *userPoolReplicaResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data userPoolReplicaResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	userPoolID := data.UserPoolID.ValueString()
	regionName := data.RegionName.ValueString()

	input := cognitoidentityprovider.CreateUserPoolReplicaInput{
		UserPoolId:   aws.String(userPoolID),
		RegionName:   aws.String(regionName),
		UserPoolTags: getTagsIn(ctx),
	}

	_, err := conn.CreateUserPoolReplica(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Cognito User Pool Replica (%s/%s)", userPoolID, regionName), err.Error())
		return
	}

	replica, err := waitUserPoolReplicaCreated(ctx, conn, userPoolID, regionName, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Cognito User Pool Replica (%s/%s) create", userPoolID, regionName), err.Error())
		return
	}

	// Activate if requested (replicas settle to INACTIVE after creation).
	if data.Status.ValueEnum() == awstypes.UpdateReplicaStatusTypeActive {
		replica, err = updateUserPoolReplicaStatus(ctx, conn, userPoolID, regionName, awstypes.UpdateReplicaStatusTypeActive, r.CreateTimeout(ctx, data.Timeouts))
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("activating Cognito User Pool Replica (%s/%s)", userPoolID, regionName), err.Error())
			return
		}
	}

	data.flattenReplica(replica)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *userPoolReplicaResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data userPoolReplicaResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	userPoolID := data.UserPoolID.ValueString()
	regionName := data.RegionName.ValueString()

	replica, err := findUserPoolReplicaByTwoPartKey(ctx, conn, userPoolID, regionName)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito User Pool Replica (%s/%s)", userPoolID, regionName), err.Error())
		return
	}

	data.flattenReplica(replica)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *userPoolReplicaResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state userPoolReplicaResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	userPoolID := plan.UserPoolID.ValueString()
	regionName := plan.RegionName.ValueString()

	if !plan.Status.Equal(state.Status) {
		replica, err := updateUserPoolReplicaStatus(ctx, conn, userPoolID, regionName, plan.Status.ValueEnum(), r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Cognito User Pool Replica (%s/%s)", userPoolID, regionName), err.Error())
			return
		}
		plan.flattenReplica(replica)
	} else {
		plan.Role = state.Role
		plan.UserPoolARN = state.UserPoolARN
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *userPoolReplicaResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data userPoolReplicaResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	userPoolID := data.UserPoolID.ValueString()
	regionName := data.RegionName.ValueString()

	input := cognitoidentityprovider.DeleteUserPoolReplicaInput{
		UserPoolId: aws.String(userPoolID),
		RegionName: aws.String(regionName),
	}

	_, err := conn.DeleteUserPoolReplica(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Cognito User Pool Replica (%s/%s)", userPoolID, regionName), err.Error())
		return
	}

	if _, err := waitUserPoolReplicaDeleted(ctx, conn, userPoolID, regionName, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Cognito User Pool Replica (%s/%s) delete", userPoolID, regionName), err.Error())
		return
	}
}

// updateUserPoolReplicaStatus calls UpdateUserPoolReplica and waits for the target status.
func updateUserPoolReplicaStatus(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, regionName string, status awstypes.UpdateReplicaStatusType, timeout time.Duration) (*awstypes.UserPoolReplicaType, error) {
	input := cognitoidentityprovider.UpdateUserPoolReplicaInput{
		UserPoolId: aws.String(userPoolID),
		RegionName: aws.String(regionName),
		Status:     status,
	}

	if _, err := conn.UpdateUserPoolReplica(ctx, &input); err != nil {
		return nil, err
	}

	return waitUserPoolReplicaStatus(ctx, conn, userPoolID, regionName, status, timeout)
}

func findUserPoolReplicaByTwoPartKey(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, regionName string) (*awstypes.UserPoolReplicaType, error) {
	input := cognitoidentityprovider.ListUserPoolReplicasInput{
		UserPoolId: aws.String(userPoolID),
	}

	for {
		page, err := conn.ListUserPoolReplicas(ctx, &input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{LastError: err}
		}
		if err != nil {
			return nil, err
		}

		for _, replica := range page.UserPoolReplicas {
			if replica.Role == awstypes.ReplicaRoleTypeSecondary && aws.ToString(replica.RegionName) == regionName {
				return &replica, nil
			}
		}

		if page.NextToken == nil {
			break
		}
		input.NextToken = page.NextToken
	}

	return nil, &retry.NotFoundError{}
}

func statusUserPoolReplica(conn *cognitoidentityprovider.Client, userPoolID, regionName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		replica, err := findUserPoolReplicaByTwoPartKey(ctx, conn, userPoolID, regionName)

		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return replica, string(replica.Status), nil
	}
}

func waitUserPoolReplicaCreated(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, regionName string, timeout time.Duration) (*awstypes.UserPoolReplicaType, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(awstypes.ReplicaStatusTypeCreating)},
		Target:  []string{string(awstypes.ReplicaStatusTypeInactive), string(awstypes.ReplicaStatusTypeActive)},
		Refresh: statusUserPoolReplica(conn, userPoolID, regionName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.UserPoolReplicaType); ok {
		return output, err
	}

	return nil, err
}

func waitUserPoolReplicaStatus(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, regionName string, status awstypes.UpdateReplicaStatusType, timeout time.Duration) (*awstypes.UserPoolReplicaType, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			string(awstypes.ReplicaStatusTypeCreating),
			string(awstypes.ReplicaStatusTypeActive),
			string(awstypes.ReplicaStatusTypeInactive),
		},
		Target:  []string{string(status)},
		Refresh: statusUserPoolReplica(conn, userPoolID, regionName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.UserPoolReplicaType); ok {
		return output, err
	}

	return nil, err
}

func waitUserPoolReplicaDeleted(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, regionName string, timeout time.Duration) (*awstypes.UserPoolReplicaType, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			string(awstypes.ReplicaStatusTypeActive),
			string(awstypes.ReplicaStatusTypeInactive),
			string(awstypes.ReplicaStatusTypeDeleting),
		},
		Target:  []string{},
		Refresh: statusUserPoolReplica(conn, userPoolID, regionName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.UserPoolReplicaType); ok {
		return output, err
	}

	return nil, err
}

type userPoolReplicaResourceModel struct {
	framework.WithRegionModel
	RegionName  types.String                                         `tfsdk:"region_name"`
	Role        fwtypes.StringEnum[awstypes.ReplicaRoleType]         `tfsdk:"role"`
	Status      fwtypes.StringEnum[awstypes.UpdateReplicaStatusType] `tfsdk:"status"`
	Tags        tftags.Map                                           `tfsdk:"tags"`
	TagsAll     tftags.Map                                           `tfsdk:"tags_all"`
	Timeouts    timeouts.Value                                       `tfsdk:"timeouts"`
	UserPoolARN fwtypes.ARN                                          `tfsdk:"user_pool_arn"`
	UserPoolID  types.String                                         `tfsdk:"user_pool_id"`
}

// flattenReplica copies API response fields into the model. Status is mapped from
// ReplicaStatusType (CREATING/ACTIVE/INACTIVE/DELETING) to the desired-state enum
// (ACTIVE/INACTIVE); transient states map to INACTIVE.
func (m *userPoolReplicaResourceModel) flattenReplica(replica *awstypes.UserPoolReplicaType) {
	m.Role = fwtypes.StringEnumValue(replica.Role)
	m.UserPoolARN = fwtypes.ARNValue(aws.ToString(replica.UserPoolArn))
	if replica.Status == awstypes.ReplicaStatusTypeActive {
		m.Status = fwtypes.StringEnumValue(awstypes.UpdateReplicaStatusTypeActive)
	} else {
		m.Status = fwtypes.StringEnumValue(awstypes.UpdateReplicaStatusTypeInactive)
	}
}

var _ inttypes.ImportIDParser = userPoolReplicaImportID{}

type userPoolReplicaImportID struct{}

func (userPoolReplicaImportID) Parse(id string) (string, map[string]any, error) {
	const partCount = 2
	parts, err := intflex.ExpandResourceId(id, partCount, true)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		names.AttrUserPoolID: parts[0],
		"region_name":        parts[1],
	}

	return id, result, nil
}
