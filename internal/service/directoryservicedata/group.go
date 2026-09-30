// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directoryservicedata

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservicedata"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservicedata/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @FrameworkResource("aws_directoryservicedata_group", name="Group")
// @IdentityAttribute("directory_id")
// @IdentityAttribute("sam_account_name")
// @ImportIDHandler("groupImportID")
// @Testing(hasNoPreExistingResource=true)

func newGroupResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &groupResource{}, nil
}

const (
	groupResourceIDPartCount = 2
)

type groupResource struct {
	framework.ResourceWithModel[groupResourceModel]
	framework.WithImportByIdentity
}

func (r *groupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"directory_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^d-[0-9a-f]{10}$`), "must be a valid Directory Service directory ID"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"distinguished_name": schema.StringAttribute{
				Computed: true,
			},
			"group_scope": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("DomainLocal", "Global", "Universal", "BuiltinLocal"),
				},
			},
			"group_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("Distribution", "Security"),
				},
			},
			"sam_account_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[^:;|=+"*?<>/\\,\[\]@]+$`), "must not contain any of the following characters: : ; | = + \" * ? < > / \\ , [ ] @"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sid": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DirectoryServiceDataClient(ctx)

	var plan groupResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	directoryID, samAccountName := plan.DirectoryID.ValueString(), plan.SAMAccountName.ValueString()

	id := groupResourceID(directoryID, samAccountName)

	var input directoryservicedata.CreateGroupInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))

	_, err := conn.CreateGroup(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	out, err := findGroupByTwoPartKey(ctx, conn, directoryID, samAccountName)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DirectoryServiceDataClient(ctx)

	var state groupResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findGroupByTwoPartKey(ctx, conn, state.DirectoryID.ValueString(), state.SAMAccountName.ValueString())

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, groupResourceID(state.DirectoryID.ValueString(), state.SAMAccountName.ValueString()))
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DirectoryServiceDataClient(ctx)

	var plan, state groupResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	id := groupResourceID(plan.DirectoryID.ValueString(), plan.SAMAccountName.ValueString())

	scopeChanged := !plan.GroupScope.Equal(state.GroupScope)
	typeChanged := !plan.GroupType.Equal(state.GroupType)

	if scopeChanged || typeChanged {
		input := directoryservicedata.UpdateGroupInput{
			DirectoryId:    plan.DirectoryID.ValueStringPointer(),
			SAMAccountName: plan.SAMAccountName.ValueStringPointer(),
			UpdateType:     awstypes.UpdateTypeReplace,
		}

		if scopeChanged {
			if plan.GroupScope.IsNull() || plan.GroupScope.IsUnknown() {
				smerr.AddError(ctx, &resp.Diagnostics, smarterr.Errorf("cannot update group scope to a null or unknown value"), smerr.ID, id)
				return
			}
			input.GroupScope = awstypes.GroupScope(plan.GroupScope.ValueString())
		}
		if typeChanged {
			if plan.GroupType.IsNull() || plan.GroupType.IsUnknown() {
				smerr.AddError(ctx, &resp.Diagnostics, smarterr.Errorf("cannot update group type to a null or unknown value"), smerr.ID, id)
				return
			}
			input.GroupType = awstypes.GroupType(plan.GroupType.ValueString())
		}

		_, err := conn.UpdateGroup(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
			return
		}
	}

	out, err := findGroupByTwoPartKey(ctx, conn, plan.DirectoryID.ValueString(), plan.SAMAccountName.ValueString())

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DirectoryServiceDataClient(ctx)

	var state groupResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := directoryservicedata.DeleteGroupInput{
		DirectoryId:    state.DirectoryID.ValueStringPointer(),
		SAMAccountName: state.SAMAccountName.ValueStringPointer(),
	}

	_, err := conn.DeleteGroup(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, groupResourceID(state.DirectoryID.ValueString(), state.SAMAccountName.ValueString()))
		return
	}
}

func findGroupByTwoPartKey(ctx context.Context, conn *directoryservicedata.Client, directoryID, samAccountName string) (*directoryservicedata.DescribeGroupOutput, error) {
	input := directoryservicedata.DescribeGroupInput{
		DirectoryId:    aws.String(directoryID),
		SAMAccountName: aws.String(samAccountName),
	}

	out, err := conn.DescribeGroup(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return out, nil
}

type groupResourceModel struct {
	framework.WithRegionModel
	DirectoryID       types.String `tfsdk:"directory_id"`
	DistinguishedName types.String `tfsdk:"distinguished_name"`
	GroupScope        types.String `tfsdk:"group_scope"`
	GroupType         types.String `tfsdk:"group_type"`
	// OtherAttributes   types.Map    `tfsdk:"other_attributes"`
	SAMAccountName types.String `tfsdk:"sam_account_name"`
	SID            types.String `tfsdk:"sid"`
}

var (
	_ inttypes.ImportIDParser = groupImportID{}
)

type groupImportID struct{}

func (groupImportID) Parse(id string) (string, map[string]any, error) {
	parts, err := intflex.ExpandResourceId(id, groupResourceIDPartCount, false)
	if err != nil {
		return "", nil, smarterr.NewError(err)
	}

	result := map[string]any{
		"directory_id":     parts[0],
		"sam_account_name": parts[1],
	}

	return id, result, nil
}

func groupResourceID(directoryID, samAccountName string) string {
	id, _ := intflex.FlattenResourceId(
		[]string{directoryID, samAccountName},
		groupResourceIDPartCount,
		false,
	)

	return id
}
