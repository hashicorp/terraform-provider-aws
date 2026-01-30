// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package dsql

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dsql/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_dsql_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/dsql;dsql.GetClusterOutput")
// @Testing(importStateIdAttribute="identifier")
// @Testing(generator=false)
func newClusterResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &clusterResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type clusterResource struct {
	framework.ResourceWithModel[clusterResourceModel]
	framework.WithTimeouts
}

func (r *clusterResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"deletion_protection_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"encryption_details": framework.ResourceComputedListOfObjectsAttribute[encryptionDetailsModel](ctx),
			names.AttrForceDestroy: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			names.AttrIdentifier: framework.IDAttribute(),
			"kms_encryption_key": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.Any(
						fwvalidators.ARN(),
						stringvalidator.OneOf("AWS_OWNED_KMS_KEY"),
					),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"vpc_endpoint_service_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"multi_region_properties": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[multiRegionPropertiesModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"clusters": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.UseStateForUnknown(),
							},
						},
						"witness_region": schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
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

func (r *clusterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data clusterResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSQLClient(ctx)

	var input dsql.CreateClusterInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateCluster(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Aurora DSQL Cluster", err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if v := output.EncryptionDetails; v != nil {
		switch typ := v.EncryptionType; typ {
		case awstypes.EncryptionTypeAwsOwnedKmsKey:
			data.KMSEncryptionKey = fwflex.StringValueToFramework(ctx, typ)
		}
	}

	id := fwflex.StringValueFromFramework(ctx, data.Identifier)
	if _, err := waitClusterCreated(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Aurora DSQL Cluster (%s) create", id), err.Error())

		return
	}

	vpcEndpointServiceName, err := findVPCEndpointServiceNameByID(ctx, conn, id)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Aurora DSQL Cluster (%s) VPC endpoint service name", id), err.Error())

		return
	}

	data.VPCEndpointServiceName = fwflex.StringToFramework(ctx, vpcEndpointServiceName)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *clusterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data clusterResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSQLClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.Identifier)
	output, err := findClusterByID(ctx, conn, id)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Aurora DSQL Cluster (%s)", id), err.Error())

		return
	}

	output.MultiRegionProperties = normalizeMultiRegionProperties(output)

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if v := output.EncryptionDetails; v != nil {
		switch typ := v.EncryptionType; typ {
		case awstypes.EncryptionTypeAwsOwnedKmsKey:
			data.KMSEncryptionKey = fwflex.StringValueToFramework(ctx, typ)
		case awstypes.EncryptionTypeCustomerManagedKmsKey:
			data.KMSEncryptionKey = fwflex.StringToFramework(ctx, v.KmsKeyArn)
		}
	}

	vpcEndpointServiceName, err := findVPCEndpointServiceNameByID(ctx, conn, id)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Aurora DSQL Cluster (%s) VPC endpoint service name", id), err.Error())

		return
	}

	data.VPCEndpointServiceName = fwflex.StringToFramework(ctx, vpcEndpointServiceName)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *clusterResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old clusterResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSQLClient(ctx)

	if !new.DeletionProtectionEnabled.Equal(old.DeletionProtectionEnabled) ||
		!new.KMSEncryptionKey.Equal(old.KMSEncryptionKey) ||
		!new.MultiRegionProperties.Equal(old.MultiRegionProperties) {
		id := fwflex.StringValueFromFramework(ctx, new.Identifier)
		var input dsql.UpdateClusterInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(sdkid.UniqueId())

		_, err := conn.UpdateCluster(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Aurora DSQL Cluster (%s)", id), err.Error())

			return
		}

		if _, err := waitClusterUpdated(ctx, conn, id, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Aurora DSQL Cluster (%s) update", id), err.Error())

			return
		}

		if new.KMSEncryptionKey.Equal(old.KMSEncryptionKey) {
			new.EncryptionDetails = old.EncryptionDetails
		} else {
			output, err := waitClusterEncryptionEnabled(ctx, conn, id, r.UpdateTimeout(ctx, new.Timeouts))

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("waiting for Aurora DSQL Cluster (%s) encryption enable", id), err.Error())

				return
			}

			response.Diagnostics.Append(fwflex.Flatten(ctx, output, &new.EncryptionDetails)...)
			if response.Diagnostics.HasError() {
				return
			}
		}
	} else {
		new.EncryptionDetails = old.EncryptionDetails
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *clusterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data clusterResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSQLClient(ctx)

	if data.ForceDestroy.ValueBool() {
		input := dsql.UpdateClusterInput{
			Identifier:                data.Identifier.ValueStringPointer(),
			DeletionProtectionEnabled: aws.Bool(false),
			ClientToken:               aws.String(sdkid.UniqueId()),
		}
		// Changing DeletionProtectionEnabled is instantaneous, no need to wait.
		if _, err := conn.UpdateCluster(ctx, &input); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("disabling deletion protection for Aurora DSQL Cluster (%s)", data.Identifier.ValueString()), err.Error())
			return
		}
	}

	id := fwflex.StringValueFromFramework(ctx, data.Identifier)
	tflog.Debug(ctx, "deleting Aurora DSQL Cluster", map[string]any{
		names.AttrIdentifier: id,
	})

	input := dsql.DeleteClusterInput{
		Identifier: aws.String(id),
	}
	_, err := conn.DeleteCluster(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Aurora DSQL Cluster (%s)", id), err.Error())

		return
	}

	if _, err := waitClusterDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Aurora DSQL Cluster (%s) delete", id), err.Error())

		return
	}
}

func (r *clusterResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrIdentifier), request, response)

	// Set force_destroy to false on import to prevent accidental deletion
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrForceDestroy), types.BoolValue(false))...)
}

func findClusterByID(ctx context.Context, conn *dsql.Client, id string) (*dsql.GetClusterOutput, error) {
	input := dsql.GetClusterInput{
		Identifier: aws.String(id),
	}
	output, err := conn.GetCluster(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findVPCEndpointServiceNameByID(ctx context.Context, conn *dsql.Client, id string) (*string, error) {
	input := dsql.GetVpcEndpointServiceNameInput{
		Identifier: aws.String(id),
	}
	output, err := conn.GetVpcEndpointServiceName(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ServiceName == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.ServiceName, nil
}

func statusCluster(conn *dsql.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findClusterByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitClusterCreated(ctx context.Context, conn *dsql.Client, id string, timeout time.Duration) (*dsql.GetClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ClusterStatusCreating),
		Target:                    enum.Slice(awstypes.ClusterStatusActive, awstypes.ClusterStatusPendingSetup),
		Refresh:                   statusCluster(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dsql.GetClusterOutput); ok {
		return output, err
	}

	return nil, err
}

func waitClusterUpdated(ctx context.Context, conn *dsql.Client, id string, timeout time.Duration) (*dsql.GetClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ClusterStatusUpdating),
		Target:  enum.Slice(awstypes.ClusterStatusActive),
		Refresh: statusCluster(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dsql.GetClusterOutput); ok {
		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *dsql.Client, id string, timeout time.Duration) (*dsql.GetClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(awstypes.ClusterStatusDeleting, awstypes.ClusterStatusPendingDelete),
		Target:       []string{},
		Refresh:      statusCluster(conn, id),
		Timeout:      timeout,
		Delay:        1 * time.Minute,
		PollInterval: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dsql.GetClusterOutput); ok {
		return output, err
	}

	return nil, err
}

func statusClusterEncryption(conn *dsql.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findClusterByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.EncryptionDetails == nil {
			return nil, "", nil
		}

		return output.EncryptionDetails, string(output.EncryptionDetails.EncryptionStatus), nil
	}
}

func waitClusterEncryptionEnabled(ctx context.Context, conn *dsql.Client, id string, timeout time.Duration) (*awstypes.EncryptionDetails, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EncryptionStatusEnabling, awstypes.EncryptionStatusUpdating),
		Target:  enum.Slice(awstypes.EncryptionStatusEnabled),
		Refresh: statusClusterEncryption(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.EncryptionDetails); ok {
		return output, err
	}

	return nil, err
}

func normalizeMultiRegionProperties(output *dsql.GetClusterOutput) *awstypes.MultiRegionProperties {
	if output == nil || output.MultiRegionProperties == nil {
		return nil
	}

	// Take a deep copy.
	apiObject := awstypes.MultiRegionProperties{
		Clusters:      slices.Clone(output.MultiRegionProperties.Clusters),
		WitnessRegion: aws.String(strings.Clone(aws.ToString(output.MultiRegionProperties.WitnessRegion))),
	}

	if sourceClusterARN := output.Arn; sourceClusterARN != nil {
		// Remove the current cluster from the list of clusters in the multi-region properties
		// This is needed because one of the ARNs of the clusters in the multi-region properties is
		// the same as the ARN of this specific cluster, and we need to remove it from the
		// list of clusters to avoid a conflict when updating the resource.
		apiObject.Clusters = slices.DeleteFunc(apiObject.Clusters, func(s string) bool {
			return strings.EqualFold(s, aws.ToString(sourceClusterARN))
		})
	}

	return &apiObject
}

type clusterResourceModel struct {
	framework.WithRegionModel
	ARN                       types.String                                                `tfsdk:"arn"`
	DeletionProtectionEnabled types.Bool                                                  `tfsdk:"deletion_protection_enabled"`
	EncryptionDetails         fwtypes.ListNestedObjectValueOf[encryptionDetailsModel]     `tfsdk:"encryption_details"`
	ForceDestroy              types.Bool                                                  `tfsdk:"force_destroy"`
	Identifier                types.String                                                `tfsdk:"identifier"`
	KMSEncryptionKey          types.String                                                `tfsdk:"kms_encryption_key"`
	MultiRegionProperties     fwtypes.ListNestedObjectValueOf[multiRegionPropertiesModel] `tfsdk:"multi_region_properties"`
	Tags                      tftags.Map                                                  `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                  `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                                              `tfsdk:"timeouts"`
	VPCEndpointServiceName    types.String                                                `tfsdk:"vpc_endpoint_service_name"`
}

type encryptionDetailsModel struct {
	EncryptionStatus fwtypes.StringEnum[awstypes.EncryptionStatus] `tfsdk:"encryption_status"`
	EncryptionType   fwtypes.StringEnum[awstypes.EncryptionType]   `tfsdk:"encryption_type"`
}

type multiRegionPropertiesModel struct {
	Clusters      fwtypes.SetOfString `tfsdk:"clusters"`
	WitnessRegion types.String        `tfsdk:"witness_region"`
}
