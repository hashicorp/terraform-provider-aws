// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_application_status_check_association", name="Application Status Check Association")
// @IdentityAttribute("application_status_check_id")
// @IdentityAttribute("instance_id", optional="true")
// @IdentityAttribute("target_tag_key", optional="true", testNotNull="true")
// @IdentityAttribute("target_tag_value", optional="true", testNotNull="true")
// @ImportIDHandler("applicationStatusCheckAssociationImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc=testAccApplicationStatusCheckAssociationImportStateIDFunc)
// @Testing(importStateIdAttribute="application_status_check_id")
// @Testing(preCheck="testAccPreCheckApplicationStatusCheck")
func newApplicationStatusCheckAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &applicationStatusCheckAssociationResource{}, nil
}

type applicationStatusCheckAssociationResource struct {
	framework.ResourceWithModel[applicationStatusCheckAssociationResourceModel]
	framework.WithNoUpdate
	framework.WithImportByIdentity
}

func (r *applicationStatusCheckAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_status_check_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrInstanceID: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"target_tag_key": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.AlsoRequires(path.MatchRoot("target_tag_value")),
				},
			},
			"target_tag_value": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("target_tag_key")),
				},
			},
		},
	}
}

func (r *applicationStatusCheckAssociationResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot(names.AttrInstanceID),
			path.MatchRoot("target_tag_key"),
		),
	}
}

func (r *applicationStatusCheckAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan applicationStatusCheckAssociationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	instanceIDs, targetTags, err := expandApplicationStatusCheckAssociationTarget(plan)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ApplicationStatusCheckID.String())
		return
	}

	input := ec2.AssociateApplicationStatusCheckInput{
		ApplicationStatusCheckId: plan.ApplicationStatusCheckID.ValueStringPointer(),
		InstanceIds:              instanceIDs,
		TargetTagAssociations:    targetTags,
	}

	output, err := conn.AssociateApplicationStatusCheck(ctx, &input)
	if err == nil {
		if output == nil {
			err = errors.New("empty output")
		} else {
			err = applicationStatusCheckAssociationResultsError(output.SuccessfulResults, output.UnsuccessfulResults)
		}
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ApplicationStatusCheckID.String())
		return
	}

	// Persist the association identity before waiting so a successful API call is
	// recoverable if the subsequent consistency check fails.
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	association, err := waitApplicationStatusCheckAssociationCreated(ctx, conn, plan)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ApplicationStatusCheckID.String())
		return
	}

	flattenApplicationStatusCheckAssociation(association, &plan)
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *applicationStatusCheckAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state applicationStatusCheckAssociationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findApplicationStatusCheckAssociationByKey(
		ctx,
		conn,
		state.ApplicationStatusCheckID.ValueString(),
		state.InstanceID.ValueString(),
		state.TargetTagKey.ValueString(),
		state.TargetTagValue.ValueString(),
	)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ApplicationStatusCheckID.String())
		return
	}

	flattenApplicationStatusCheckAssociation(output, &state)
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *applicationStatusCheckAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state applicationStatusCheckAssociationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	instanceIDs, targetTags, err := expandApplicationStatusCheckAssociationTarget(state)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ApplicationStatusCheckID.String())
		return
	}

	input := ec2.DisassociateApplicationStatusCheckInput{
		ApplicationStatusCheckId: state.ApplicationStatusCheckID.ValueStringPointer(),
		InstanceIds:              instanceIDs,
		TargetTagAssociations:    targetTags,
	}

	output, err := conn.DisassociateApplicationStatusCheck(ctx, &input)
	if err == nil {
		if output == nil {
			err = errors.New("empty output")
		} else {
			err = applicationStatusCheckAssociationResultsError(output.SuccessfulResults, output.UnsuccessfulResults)
		}
	}
	if err != nil {
		_, findErr := findApplicationStatusCheckAssociationByKey(
			ctx,
			conn,
			state.ApplicationStatusCheckID.ValueString(),
			state.InstanceID.ValueString(),
			state.TargetTagKey.ValueString(),
			state.TargetTagValue.ValueString(),
		)
		if retry.NotFound(findErr) {
			return
		}
		if findErr != nil {
			err = errors.Join(err, findErr)
		}

		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ApplicationStatusCheckID.String())
		return
	}

	if err := waitApplicationStatusCheckAssociationDeleted(ctx, conn, state); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ApplicationStatusCheckID.String())
	}
}

type applicationStatusCheckAssociationResourceModel struct {
	framework.WithRegionModel
	ApplicationStatusCheckID types.String `tfsdk:"application_status_check_id"`
	InstanceID               types.String `tfsdk:"instance_id"`
	TargetTagKey             types.String `tfsdk:"target_tag_key"`
	TargetTagValue           types.String `tfsdk:"target_tag_value"`
}

func expandApplicationStatusCheckAssociationTarget(data applicationStatusCheckAssociationResourceModel) ([]string, []awstypes.CustomTagKeyValueRequestPair, error) {
	instanceID := data.InstanceID.ValueString()
	tagKey := data.TargetTagKey.ValueString()
	tagValue := data.TargetTagValue.ValueString()

	switch {
	case instanceID != "" && tagKey == "" && data.TargetTagValue.IsNull():
		return []string{instanceID}, nil, nil
	case instanceID == "" && tagKey != "" && !data.TargetTagValue.IsNull():
		return nil, []awstypes.CustomTagKeyValueRequestPair{{
			Key:   aws.String(tagKey),
			Value: aws.String(tagValue),
		}}, nil
	default:
		return nil, nil, errors.New("exactly one application status check association target must be specified: instance_id or target_tag_key with target_tag_value")
	}
}

func flattenApplicationStatusCheckAssociation(apiObject *awstypes.ApplicationStatusCheckAssociationObject, data *applicationStatusCheckAssociationResourceModel) {
	data.ApplicationStatusCheckID = types.StringPointerValue(apiObject.ApplicationStatusCheckId)

	switch apiObject.AssociationType {
	case awstypes.AssociationTypeEnumInstanceId:
		data.InstanceID = types.StringPointerValue(apiObject.Value)
		data.TargetTagKey = types.StringNull()
		data.TargetTagValue = types.StringNull()
	case awstypes.AssociationTypeEnumTag:
		data.InstanceID = types.StringNull()
		data.TargetTagKey = types.StringPointerValue(apiObject.Key)
		data.TargetTagValue = types.StringPointerValue(apiObject.Value)
	}
}

func findApplicationStatusCheckAssociationByKey(ctx context.Context, conn *ec2.Client, applicationStatusCheckID, instanceID, tagKey, tagValue string) (*awstypes.ApplicationStatusCheckAssociationObject, error) {
	if instanceID == "" && tagKey == "" {
		return nil, smarterr.NewError(errors.New("application status check association target is empty"))
	}
	if instanceID != "" && tagKey != "" {
		return nil, smarterr.NewError(errors.New("application status check association has multiple targets"))
	}
	input := &ec2.DescribeApplicationStatusCheckAssociationsInput{
		ApplicationStatusCheckIds: []string{applicationStatusCheckID},
	}
	var output []awstypes.ApplicationStatusCheckAssociationObject

	err := describeApplicationStatusCheckAssociationsPages(ctx, conn, input, func(page *ec2.DescribeApplicationStatusCheckAssociationsOutput, _ bool) bool {
		for _, association := range page.Associations {
			if applicationStatusCheckAssociationMatches(&association, applicationStatusCheckID, instanceID, tagKey, tagValue) {
				output = append(output, association)
			}
		}
		return true
	})
	if err != nil {
		_, findErr := findApplicationStatusCheckByID(ctx, conn, applicationStatusCheckID)
		if retry.NotFound(findErr) {
			return nil, smarterr.NewError(findErr)
		}
		if findErr != nil {
			return nil, smarterr.NewError(errors.Join(err, findErr))
		}
		return nil, smarterr.NewError(err)
	}

	result, err := tfresource.AssertSingleValueResult(output)
	return result, smarterr.NewError(err)
}

func applicationStatusCheckAssociationMatches(apiObject *awstypes.ApplicationStatusCheckAssociationObject, applicationStatusCheckID, instanceID, tagKey, tagValue string) bool {
	if apiObject == nil || aws.ToString(apiObject.ApplicationStatusCheckId) != applicationStatusCheckID {
		return false
	}

	switch apiObject.AssociationType {
	case awstypes.AssociationTypeEnumInstanceId:
		return instanceID != "" && tagKey == "" && aws.ToString(apiObject.Value) == instanceID
	case awstypes.AssociationTypeEnumTag:
		return instanceID == "" && aws.ToString(apiObject.Key) == tagKey && aws.ToString(apiObject.Value) == tagValue
	default:
		return false
	}
}

func waitApplicationStatusCheckAssociationCreated(ctx context.Context, conn *ec2.Client, data applicationStatusCheckAssociationResourceModel) (*awstypes.ApplicationStatusCheckAssociationObject, error) {
	output, err := tfresource.RetryWhenNotFound(ctx, ec2PropagationTimeout, func(ctx context.Context) (*awstypes.ApplicationStatusCheckAssociationObject, error) {
		return findApplicationStatusCheckAssociationByKey(
			ctx,
			conn,
			data.ApplicationStatusCheckID.ValueString(),
			data.InstanceID.ValueString(),
			data.TargetTagKey.ValueString(),
			data.TargetTagValue.ValueString(),
		)
	})

	return output, smarterr.NewError(err)
}

func waitApplicationStatusCheckAssociationDeleted(ctx context.Context, conn *ec2.Client, data applicationStatusCheckAssociationResourceModel) error {
	_, err := tfresource.RetryUntilNotFound(ctx, ec2PropagationTimeout, func(ctx context.Context) (any, error) {
		return findApplicationStatusCheckAssociationByKey(
			ctx,
			conn,
			data.ApplicationStatusCheckID.ValueString(),
			data.InstanceID.ValueString(),
			data.TargetTagKey.ValueString(),
			data.TargetTagValue.ValueString(),
		)
	})

	return smarterr.NewError(err)
}

func applicationStatusCheckAssociationResultsError(successful []awstypes.SuccessfulAssociationResponseObject, unsuccessful []awstypes.UnsuccessfulAssociationResponseObject) error {
	var errs []error

	for _, result := range unsuccessful {
		errs = append(errs, fmt.Errorf("%s (%s): %s", aws.ToString(result.AssociationValue), aws.ToString(result.AssociationType), aws.ToString(result.Reason)))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	if len(successful) != 1 {
		return fmt.Errorf("expected one successful result, got %d", len(successful))
	}

	return nil
}

var _ inttypes.ImportIDParser = applicationStatusCheckAssociationImportID{}

type applicationStatusCheckAssociationImportID struct{}

func (applicationStatusCheckAssociationImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.Split(id, ",")
	if len(parts) < 2 {
		return "", nil, applicationStatusCheckAssociationImportIDError(id)
	}

	applicationStatusCheckID, err := url.QueryUnescape(parts[0])
	if err != nil || applicationStatusCheckID == "" {
		return "", nil, applicationStatusCheckAssociationImportIDError(id)
	}

	switch parts[1] {
	case string(awstypes.AssociationTypeEnumInstanceId):
		if len(parts) != 3 {
			return "", nil, applicationStatusCheckAssociationImportIDError(id)
		}
		instanceID, err := url.QueryUnescape(parts[2])
		if err != nil || instanceID == "" {
			return "", nil, applicationStatusCheckAssociationImportIDError(id)
		}

		result := map[string]any{
			"application_status_check_id": applicationStatusCheckID,
			names.AttrInstanceID:          instanceID,
		}
		return applicationStatusCheckAssociationImportIDString(applicationStatusCheckID, instanceID, "", ""), result, nil

	case string(awstypes.AssociationTypeEnumTag):
		if len(parts) != 4 {
			return "", nil, applicationStatusCheckAssociationImportIDError(id)
		}
		tagKey, keyErr := url.QueryUnescape(parts[2])
		tagValue, valueErr := url.QueryUnescape(parts[3])
		if keyErr != nil || valueErr != nil || tagKey == "" {
			return "", nil, applicationStatusCheckAssociationImportIDError(id)
		}

		result := map[string]any{
			"application_status_check_id": applicationStatusCheckID,
			"target_tag_key":              tagKey,
			"target_tag_value":            tagValue,
		}
		return applicationStatusCheckAssociationImportIDString(applicationStatusCheckID, "", tagKey, tagValue), result, nil

	default:
		return "", nil, applicationStatusCheckAssociationImportIDError(id)
	}
}

func applicationStatusCheckAssociationImportIDString(applicationStatusCheckID, instanceID, tagKey, tagValue string) string {
	if instanceID != "" {
		return strings.Join([]string{ // nosemgrep:ci.typed-enum-conversion
			url.QueryEscape(applicationStatusCheckID),
			string(awstypes.AssociationTypeEnumInstanceId),
			url.QueryEscape(instanceID),
		}, ",")
	}

	return strings.Join([]string{ // nosemgrep:ci.typed-enum-conversion
		url.QueryEscape(applicationStatusCheckID),
		string(awstypes.AssociationTypeEnumTag),
		url.QueryEscape(tagKey),
		url.QueryEscape(tagValue),
	}, ",")
}

func applicationStatusCheckAssociationImportIDError(id string) error {
	return fmt.Errorf("id %q should be in format <application-status-check-id>,instance-id,<instance-id> or <application-status-check-id>,tag,<percent-encoded-tag-key>,<percent-encoded-tag-value>", id)
}
