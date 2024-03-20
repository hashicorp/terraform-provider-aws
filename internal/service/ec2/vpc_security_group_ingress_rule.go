// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Security Group Ingress Rule")
// @Tags(identifierAttribute="id")
func newSecurityGroupIngressRuleResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &securityGroupIngressRuleResource{}
	r.securityGroupRule = r

	return r, nil
}

type securityGroupIngressRuleResource struct {
	securityGroupRuleResource
}

func (*securityGroupIngressRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_vpc_security_group_ingress_rule"
}

func (r *securityGroupIngressRuleResource) create(ctx context.Context, data *securityGroupRuleResourceModel) (string, error) {
	conn := r.Meta().EC2Conn(ctx)

	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       flex.StringFromFramework(ctx, data.SecurityGroupID),
		IpPermissions: []*ec2.IpPermission{data.expandIPPermission(ctx)},
	}

	output, err := conn.AuthorizeSecurityGroupIngressWithContext(ctx, input)

	if err != nil {
		return "", err
	}

	return aws.StringValue(output.SecurityGroupRules[0].SecurityGroupRuleId), nil
}

func (r *securityGroupIngressRuleResource) delete(ctx context.Context, data *securityGroupRuleResourceModel) error {
	conn := r.Meta().EC2Conn(ctx)

	_, err := conn.RevokeSecurityGroupIngressWithContext(ctx, &ec2.RevokeSecurityGroupIngressInput{
		GroupId:              flex.StringFromFramework(ctx, data.SecurityGroupID),
		SecurityGroupRuleIds: flex.StringSliceFromFramework(ctx, data.ID),
	})

	return err
}

func (r *securityGroupIngressRuleResource) findByID(ctx context.Context, id string) (*ec2.SecurityGroupRule, error) {
	conn := r.Meta().EC2Conn(ctx)

	return FindSecurityGroupIngressRuleByID(ctx, conn, id)
}

// Base structure and methods for VPC security group rules.

type securityGroupRule interface {
	create(context.Context, *securityGroupRuleResourceModel) (string, error)
	delete(context.Context, *securityGroupRuleResourceModel) error
	findByID(context.Context, string) (*ec2.SecurityGroupRule, error)
}

type securityGroupRuleResource struct {
	securityGroupRule
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *securityGroupRuleResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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
				CustomType: ipProtocolType{},
				Required:   true,
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

func (r *securityGroupRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data securityGroupRuleResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	securityGroupRuleID, err := r.securityGroupRule.create(ctx, &data)

	if err != nil {
		response.Diagnostics.AddError("creating VPC Security Group Rule", err.Error())

		return
	}

	// Set values for unknowns.
	data.ARN = r.securityGroupRuleARN(ctx, securityGroupRuleID)
	data.SecurityGroupRuleID = types.StringValue(securityGroupRuleID)
	data.setID()

	conn := r.Meta().EC2Conn(ctx)
	if err := createTags(ctx, conn, data.ID.ValueString(), getTagsIn(ctx)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("setting VPC Security Group Rule (%s) tags", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *securityGroupRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data securityGroupRuleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	output, err := r.securityGroupRule.findByID(ctx, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPC Security Group Rule (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.ARN = r.securityGroupRuleARN(ctx, data.ID.ValueString())
	data.CIDRIPv4 = flex.StringToFramework(ctx, output.CidrIpv4)
	data.CIDRIPv6 = flex.StringToFramework(ctx, output.CidrIpv6)
	data.Description = flex.StringToFramework(ctx, output.Description)
	data.IPProtocol = flex.StringToFrameworkValuable[ipProtocol](ctx, output.IpProtocol)
	data.PrefixListID = flex.StringToFramework(ctx, output.PrefixListId)
	data.ReferencedSecurityGroupID = flattenReferencedSecurityGroup(ctx, output.ReferencedGroupInfo, r.Meta().AccountID)
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

func (r *securityGroupRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new securityGroupRuleResourceModel
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
				SecurityGroupRule:   new.expandSecurityGroupRuleRequest(ctx),
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

func (r *securityGroupRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data securityGroupRuleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting VPC Security Group Rule", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
	err := r.securityGroupRule.delete(ctx, &data)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidGroupNotFound, errCodeInvalidSecurityGroupRuleIdNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPC Security Group Rule (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *securityGroupRuleResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if !request.State.Raw.IsNull() && !request.Plan.Raw.IsNull() {
		var old, new securityGroupRuleResourceModel
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

func (r *securityGroupRuleResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("cidr_ipv4"),
			path.MatchRoot("cidr_ipv6"),
			path.MatchRoot("prefix_list_id"),
			path.MatchRoot("referenced_security_group_id"),
		),
	}
}

func (r *securityGroupRuleResource) securityGroupRuleARN(_ context.Context, id string) types.String {
	return types.StringValue(r.RegionalARN(names.EC2, fmt.Sprintf("security-group-rule/%s", id)))
}

func flattenReferencedSecurityGroup(ctx context.Context, apiObject *ec2.ReferencedSecurityGroup, accountID string) types.String {
	if apiObject == nil {
		return types.StringNull()
	}

	if apiObject.UserId == nil || aws.StringValue(apiObject.UserId) == accountID {
		return flex.StringToFramework(ctx, apiObject.GroupId)
	}

	// [UserID/]GroupID.
	return types.StringValue(strings.Join([]string{aws.StringValue(apiObject.UserId), aws.StringValue(apiObject.GroupId)}, "/"))
}

type securityGroupRuleResourceModel struct {
	ARN                       types.String `tfsdk:"arn"`
	CIDRIPv4                  types.String `tfsdk:"cidr_ipv4"`
	CIDRIPv6                  types.String `tfsdk:"cidr_ipv6"`
	Description               types.String `tfsdk:"description"`
	FromPort                  types.Int64  `tfsdk:"from_port"`
	ID                        types.String `tfsdk:"id"`
	IPProtocol                ipProtocol   `tfsdk:"ip_protocol"`
	PrefixListID              types.String `tfsdk:"prefix_list_id"`
	ReferencedSecurityGroupID types.String `tfsdk:"referenced_security_group_id"`
	SecurityGroupID           types.String `tfsdk:"security_group_id"`
	SecurityGroupRuleID       types.String `tfsdk:"security_group_rule_id"`
	Tags                      types.Map    `tfsdk:"tags"`
	TagsAll                   types.Map    `tfsdk:"tags_all"`
	ToPort                    types.Int64  `tfsdk:"to_port"`
}

func (model *securityGroupRuleResourceModel) InitFromID() error {
	model.SecurityGroupRuleID = model.ID

	return nil
}

func (model *securityGroupRuleResourceModel) setID() {
	model.ID = model.SecurityGroupRuleID
}

func (d *securityGroupRuleResourceModel) expandIPPermission(ctx context.Context) *ec2.IpPermission {
	apiObject := &ec2.IpPermission{
		FromPort:   flex.Int64FromFramework(ctx, d.FromPort),
		IpProtocol: flex.StringFromFramework(ctx, d.IPProtocol),
		ToPort:     flex.Int64FromFramework(ctx, d.ToPort),
	}

	if !d.CIDRIPv4.IsNull() {
		apiObject.IpRanges = []*ec2.IpRange{{
			CidrIp:      flex.StringFromFramework(ctx, d.CIDRIPv4),
			Description: flex.StringFromFramework(ctx, d.Description),
		}}
	}

	if !d.CIDRIPv6.IsNull() {
		apiObject.Ipv6Ranges = []*ec2.Ipv6Range{{
			CidrIpv6:    flex.StringFromFramework(ctx, d.CIDRIPv6),
			Description: flex.StringFromFramework(ctx, d.Description),
		}}
	}

	if !d.PrefixListID.IsNull() {
		apiObject.PrefixListIds = []*ec2.PrefixListId{{
			PrefixListId: flex.StringFromFramework(ctx, d.PrefixListID),
			Description:  flex.StringFromFramework(ctx, d.Description),
		}}
	}

	if !d.ReferencedSecurityGroupID.IsNull() {
		apiObject.UserIdGroupPairs = []*ec2.UserIdGroupPair{{
			Description: flex.StringFromFramework(ctx, d.Description),
		}}

		// [UserID/]GroupID.
		if parts := strings.Split(d.ReferencedSecurityGroupID.ValueString(), "/"); len(parts) == 2 {
			apiObject.UserIdGroupPairs[0].GroupId = aws.String(parts[1])
			apiObject.UserIdGroupPairs[0].UserId = aws.String(parts[0])
		} else {
			apiObject.UserIdGroupPairs[0].GroupId = flex.StringFromFramework(ctx, d.ReferencedSecurityGroupID)
		}
	}

	return apiObject
}

func (d *securityGroupRuleResourceModel) expandSecurityGroupRuleRequest(ctx context.Context) *ec2.SecurityGroupRuleRequest {
	apiObject := &ec2.SecurityGroupRuleRequest{
		CidrIpv4:          flex.StringFromFramework(ctx, d.CIDRIPv4),
		CidrIpv6:          flex.StringFromFramework(ctx, d.CIDRIPv6),
		Description:       flex.StringFromFramework(ctx, d.Description),
		FromPort:          flex.Int64FromFramework(ctx, d.FromPort),
		IpProtocol:        flex.StringFromFramework(ctx, d.IPProtocol),
		PrefixListId:      flex.StringFromFramework(ctx, d.PrefixListID),
		ReferencedGroupId: flex.StringFromFramework(ctx, d.ReferencedSecurityGroupID),
		ToPort:            flex.Int64FromFramework(ctx, d.ToPort),
	}

	return apiObject
}

func (d *securityGroupRuleResourceModel) sourceAttributeName() string {
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

var (
	_ basetypes.StringTypable                    = (*ipProtocolType)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*ipProtocol)(nil)
)

type ipProtocolType struct {
	basetypes.StringType
}

func (t ipProtocolType) Equal(o attr.Type) bool {
	other, ok := o.(ipProtocolType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (ipProtocolType) String() string {
	return "IPProtocolType"
}

func (t ipProtocolType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return ipProtocolNull(), diags
	}
	if in.IsUnknown() {
		return ipProtocolUnknown(), diags
	}

	return ipProtocolValue(in.ValueString()), diags
}

func (t ipProtocolType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (ipProtocolType) ValueType(context.Context) attr.Value {
	return ipProtocol{}
}

type ipProtocol struct {
	basetypes.StringValue
}

func (v ipProtocol) Equal(o attr.Value) bool {
	other, ok := o.(ipProtocol)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (ipProtocol) Type(context.Context) attr.Type {
	return ipProtocolType{}
}

func (v ipProtocol) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(ipProtocol)
	if !ok {
		return false, diags
	}

	return protocolForValue(newValue.ValueString()) == protocolForValue(v.ValueString()), diags
}

func ipProtocolNull() ipProtocol {
	return ipProtocol{StringValue: basetypes.NewStringNull()}
}

func ipProtocolUnknown() ipProtocol {
	return ipProtocol{StringValue: basetypes.NewStringUnknown()}
}

func ipProtocolValue(value string) ipProtocol {
	return ipProtocol{StringValue: basetypes.NewStringValue(value)}
}
