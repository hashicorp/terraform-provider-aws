// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Security Group Ingress Rule")
// @Tags(identifierAttribute="id")
func newResourceSecurityGroupIngressRule(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceSecurityGroupIngressRule{}
	r.create = r.createSecurityGroupRule
	r.delete = r.deleteSecurityGroupRule
	r.findByID = r.findSecurityGroupRuleByID

	return r, nil
}

type resourceSecurityGroupIngressRule struct {
	resourceSecurityGroupRule
}

func (r *resourceSecurityGroupIngressRule) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_vpc_security_group_ingress_rule"
}

func (r *resourceSecurityGroupIngressRule) createSecurityGroupRule(ctx context.Context, data *resourceSecurityGroupRuleData) (string, error) {
	conn := r.Meta().EC2Conn(ctx)

	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       flex.StringFromFramework(ctx, data.SecurityGroupID),
		IpPermissions: []*ec2.IpPermission{r.expandIPPermission(ctx, data)},
	}

	output, err := conn.AuthorizeSecurityGroupIngressWithContext(ctx, input)

	if err != nil {
		return "", err
	}

	return aws.StringValue(output.SecurityGroupRules[0].SecurityGroupRuleId), nil
}

func (r *resourceSecurityGroupIngressRule) deleteSecurityGroupRule(ctx context.Context, data *resourceSecurityGroupRuleData) error {
	conn := r.Meta().EC2Conn(ctx)

	_, err := conn.RevokeSecurityGroupIngressWithContext(ctx, &ec2.RevokeSecurityGroupIngressInput{
		GroupId:              flex.StringFromFramework(ctx, data.SecurityGroupID),
		SecurityGroupRuleIds: flex.StringSliceFromFramework(ctx, data.ID),
	})

	return err
}

func (r *resourceSecurityGroupIngressRule) findSecurityGroupRuleByID(ctx context.Context, id string) (*ec2.SecurityGroupRule, error) {
	conn := r.Meta().EC2Conn(ctx)

	return FindSecurityGroupIngressRuleByID(ctx, conn, id)
}

// Base structure and methods for VPC security group rules.

type resourceSecurityGroupRule struct {
	framework.ResourceWithConfigure

	create   func(context.Context, *resourceSecurityGroupRuleData) (string, error)
	delete   func(context.Context, *resourceSecurityGroupRuleData) error
	findByID func(context.Context, string) (*ec2.SecurityGroupRule, error)
}

func (r *resourceSecurityGroupRule) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cidr_ipv4": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					fwvalidators.IPv4CIDRNetworkAddress(),
				},
			},
			"cidr_ipv6": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					fwvalidators.IPv6CIDRNetworkAddress(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"from_port": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(-1, 65535),
				},
			},
			"id": framework.IDAttribute(),
			"ip_protocol": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					NormalizeIPProtocol(),
				},
			},
			"prefix_list_id": schema.StringAttribute{
				Optional: true,
			},
			"referenced_security_group_id": schema.StringAttribute{
				Optional: true,
			},
			"security_group_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"security_group_rule_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"to_port": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(-1, 65535),
				},
			},
		},
	}
}

func (r *resourceSecurityGroupRule) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceSecurityGroupRuleData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	securityGroupRuleID, err := r.create(ctx, &data)

	if err != nil {
		response.Diagnostics.AddError("creating VPC Security Group Rule", err.Error())

		return
	}

	data.ID = types.StringValue(securityGroupRuleID)

	conn := r.Meta().EC2Conn(ctx)
	if err := createTags(ctx, conn, data.ID.ValueString(), getTagsIn(ctx)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("setting VPC Security Group Rule (%s) tags", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.ARN = r.arn(ctx, securityGroupRuleID)
	data.SecurityGroupRuleID = types.StringValue(securityGroupRuleID)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceSecurityGroupRule) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceSecurityGroupRuleData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	output, err := r.findByID(ctx, data.ID.ValueString())

	if tfresource.NotFound(err) {
		tflog.Warn(ctx, "VPC Security Group Rule not found, removing from state", map[string]interface{}{
			"id": data.ID.ValueString(),
		})
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPC Security Group Rule (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.ARN = r.arn(ctx, data.ID.ValueString())
	data.CIDRIPv4 = flex.StringToFramework(ctx, output.CidrIpv4)
	data.CIDRIPv6 = flex.StringToFramework(ctx, output.CidrIpv6)
	data.Description = flex.StringToFramework(ctx, output.Description)
	data.IPProtocol = flex.StringToFramework(ctx, output.IpProtocol)
	data.PrefixListID = flex.StringToFramework(ctx, output.PrefixListId)
	data.ReferencedSecurityGroupID = r.flattenReferencedSecurityGroup(ctx, output.ReferencedGroupInfo)
	data.SecurityGroupID = flex.StringToFramework(ctx, output.GroupId)
	data.SecurityGroupRuleID = flex.StringToFramework(ctx, output.SecurityGroupRuleId)

	// If planned from_port or to_port are null and values of -1 are returned, propagate null.
	if v := aws.Int64Value(output.FromPort); v == -1 && data.FromPort.IsNull() {
		data.FromPort = types.Int64Null()
	} else {
		data.FromPort = flex.Int64ToFramework(ctx, output.FromPort)
	}
	if v := aws.Int64Value(output.ToPort); v == -1 && data.ToPort.IsNull() {
		data.ToPort = types.Int64Null()
	} else {
		data.ToPort = flex.Int64ToFramework(ctx, output.ToPort)
	}

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceSecurityGroupRule) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourceSecurityGroupRuleData

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Conn(ctx)

	if !new.CIDRIPv4.Equal(old.CIDRIPv4) ||
		!new.CIDRIPv6.Equal(old.CIDRIPv6) ||
		!new.Description.Equal(old.Description) ||
		!new.FromPort.Equal(old.FromPort) ||
		!new.IPProtocol.Equal(old.IPProtocol) ||
		!new.PrefixListID.Equal(old.PrefixListID) ||
		!new.ReferencedSecurityGroupID.Equal(old.ReferencedSecurityGroupID) ||
		!new.ToPort.Equal(old.ToPort) {
		input := &ec2.ModifySecurityGroupRulesInput{
			GroupId: flex.StringFromFramework(ctx, new.SecurityGroupID),
			SecurityGroupRules: []*ec2.SecurityGroupRuleUpdate{{
				SecurityGroupRule:   r.expandSecurityGroupRuleRequest(ctx, &new),
				SecurityGroupRuleId: flex.StringFromFramework(ctx, new.ID),
			}},
		}

		_, err := conn.ModifySecurityGroupRulesWithContext(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating VPC Security Group Rule (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourceSecurityGroupRule) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceSecurityGroupRuleData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting VPC Security Group Rule", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
	err := r.delete(ctx, &data)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidGroupNotFound, errCodeInvalidSecurityGroupRuleIdNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPC Security Group Rule (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *resourceSecurityGroupRule) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *resourceSecurityGroupRule) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if !request.State.Raw.IsNull() && !request.Plan.Raw.IsNull() {
		var old, new resourceSecurityGroupRuleData

		response.Diagnostics.Append(request.State.Get(ctx, &old)...)

		if response.Diagnostics.HasError() {
			return
		}

		response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

		if response.Diagnostics.HasError() {
			return
		}

		// When you modify a rule, you cannot change the rule's source type.
		if new, old := new.sourceAttributeName(), old.sourceAttributeName(); new != old {
			response.RequiresReplace = []path.Path{path.Root(old), path.Root(new)}
		}
	}

	r.SetTagsAll(ctx, request, response)
}

func (r *resourceSecurityGroupRule) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("cidr_ipv4"),
			path.MatchRoot("cidr_ipv6"),
			path.MatchRoot("prefix_list_id"),
			path.MatchRoot("referenced_security_group_id"),
		),
	}
}

func (r *resourceSecurityGroupRule) arn(_ context.Context, id string) types.String {
	arn := arn.ARN{
		Partition: r.Meta().Partition,
		Service:   ec2.ServiceName,
		Region:    r.Meta().Region,
		AccountID: r.Meta().AccountID,
		Resource:  fmt.Sprintf("security-group-rule/%s", id),
	}.String()
	return types.StringValue(arn)
}

func (r *resourceSecurityGroupRule) expandIPPermission(ctx context.Context, data *resourceSecurityGroupRuleData) *ec2.IpPermission {
	apiObject := &ec2.IpPermission{
		FromPort:   flex.Int64FromFramework(ctx, data.FromPort),
		IpProtocol: flex.StringFromFramework(ctx, data.IPProtocol),
		ToPort:     flex.Int64FromFramework(ctx, data.ToPort),
	}

	if !data.CIDRIPv4.IsNull() {
		apiObject.IpRanges = []*ec2.IpRange{{
			CidrIp:      flex.StringFromFramework(ctx, data.CIDRIPv4),
			Description: flex.StringFromFramework(ctx, data.Description),
		}}
	}

	if !data.CIDRIPv6.IsNull() {
		apiObject.Ipv6Ranges = []*ec2.Ipv6Range{{
			CidrIpv6:    flex.StringFromFramework(ctx, data.CIDRIPv6),
			Description: flex.StringFromFramework(ctx, data.Description),
		}}
	}

	if !data.PrefixListID.IsNull() {
		apiObject.PrefixListIds = []*ec2.PrefixListId{{
			PrefixListId: flex.StringFromFramework(ctx, data.PrefixListID),
			Description:  flex.StringFromFramework(ctx, data.Description),
		}}
	}

	if !data.ReferencedSecurityGroupID.IsNull() {
		apiObject.UserIdGroupPairs = []*ec2.UserIdGroupPair{{
			Description: flex.StringFromFramework(ctx, data.Description),
		}}

		// [UserID/]GroupID.
		if parts := strings.Split(data.ReferencedSecurityGroupID.ValueString(), "/"); len(parts) == 2 {
			apiObject.UserIdGroupPairs[0].GroupId = aws.String(parts[1])
			apiObject.UserIdGroupPairs[0].UserId = aws.String(parts[0])
		} else {
			apiObject.UserIdGroupPairs[0].GroupId = flex.StringFromFramework(ctx, data.ReferencedSecurityGroupID)
		}
	}

	return apiObject
}

func (r *resourceSecurityGroupRule) expandSecurityGroupRuleRequest(ctx context.Context, data *resourceSecurityGroupRuleData) *ec2.SecurityGroupRuleRequest {
	apiObject := &ec2.SecurityGroupRuleRequest{
		CidrIpv4:          flex.StringFromFramework(ctx, data.CIDRIPv4),
		CidrIpv6:          flex.StringFromFramework(ctx, data.CIDRIPv6),
		Description:       flex.StringFromFramework(ctx, data.Description),
		FromPort:          flex.Int64FromFramework(ctx, data.FromPort),
		IpProtocol:        flex.StringFromFramework(ctx, data.IPProtocol),
		PrefixListId:      flex.StringFromFramework(ctx, data.PrefixListID),
		ReferencedGroupId: flex.StringFromFramework(ctx, data.ReferencedSecurityGroupID),
		ToPort:            flex.Int64FromFramework(ctx, data.ToPort),
	}

	return apiObject
}

func (r *resourceSecurityGroupRule) flattenReferencedSecurityGroup(ctx context.Context, apiObject *ec2.ReferencedSecurityGroup) types.String {
	if apiObject == nil {
		return types.StringNull()
	}

	if apiObject.UserId == nil || aws.StringValue(apiObject.UserId) == r.Meta().AccountID {
		return flex.StringToFramework(ctx, apiObject.GroupId)
	}

	// [UserID/]GroupID.
	return types.StringValue(strings.Join([]string{aws.StringValue(apiObject.UserId), aws.StringValue(apiObject.GroupId)}, "/"))
}

type resourceSecurityGroupRuleData struct {
	ARN                       types.String `tfsdk:"arn"`
	CIDRIPv4                  types.String `tfsdk:"cidr_ipv4"`
	CIDRIPv6                  types.String `tfsdk:"cidr_ipv6"`
	Description               types.String `tfsdk:"description"`
	FromPort                  types.Int64  `tfsdk:"from_port"`
	ID                        types.String `tfsdk:"id"`
	IPProtocol                types.String `tfsdk:"ip_protocol"`
	PrefixListID              types.String `tfsdk:"prefix_list_id"`
	ReferencedSecurityGroupID types.String `tfsdk:"referenced_security_group_id"`
	SecurityGroupID           types.String `tfsdk:"security_group_id"`
	SecurityGroupRuleID       types.String `tfsdk:"security_group_rule_id"`
	Tags                      types.Map    `tfsdk:"tags"`
	TagsAll                   types.Map    `tfsdk:"tags_all"`
	ToPort                    types.Int64  `tfsdk:"to_port"`
}

func (d *resourceSecurityGroupRuleData) sourceAttributeName() string {
	switch {
	case !d.CIDRIPv4.IsNull():
		return "cidr_ipv4"
	case !d.CIDRIPv6.IsNull():
		return "cidr_ipv6"
	case !d.PrefixListID.IsNull():
		return "prefix_list_id"
	case !d.ReferencedSecurityGroupID.IsNull():
		return "referenced_security_group_id"
	}

	return ""
}

type normalizeIPProtocol struct{}

func NormalizeIPProtocol() planmodifier.String {
	return normalizeIPProtocol{}
}

func (m normalizeIPProtocol) Description(context.Context) string {
	return "Resolve differences between IP protocol names and numbers"
}

func (m normalizeIPProtocol) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m normalizeIPProtocol) PlanModifyString(ctx context.Context, request planmodifier.StringRequest, response *planmodifier.StringResponse) {
	if request.StateValue.IsNull() {
		response.PlanValue = request.PlanValue
		return
	}

	// If the state value is semantically equivalent to the planned value
	// then return the state value, else return the planned value.
	if ProtocolForValue(request.StateValue.ValueString()) == ProtocolForValue(request.PlanValue.ValueString()) {
		response.PlanValue = request.StateValue
		return
	}
	response.PlanValue = request.PlanValue
}
