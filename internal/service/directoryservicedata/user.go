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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_directoryservicedata_user", name="User")
// @IdentityAttribute("directory_id")
// @IdentityAttribute("sam_account_name")
// @ImportIDHandler("userImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(domainTfVar="directoryDomain")
// @Testing(emailAddress="emailAddress")
func newUserResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &userResource{}, nil
}

const (
	ResNameUser             = "User"
	userResourceIDPartCount = 2
)

type userResource struct {
	framework.ResourceWithModel[userResourceModel]
	framework.WithImportByIdentity
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"email_address": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			names.AttrEnabled: schema.BoolAttribute{
				Computed: true,
			},
			"given_name": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"realm": schema.StringAttribute{
				Computed: true,
			},
			"sam_account_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 20),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[\w\-.]+$`), "must contain only word characters, hyphens, and periods"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sid": schema.StringAttribute{
				Computed: true,
			},
			"surname": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"user_principal_name": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DirectoryServiceDataClient(ctx)

	var plan userResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	directoryID, samAccountName := plan.DirectoryID.ValueString(), plan.SAMAccountName.ValueString()

	id := userResourceID(directoryID, samAccountName)

	var input directoryservicedata.CreateUserInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateUser(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	out, err := findUserByTwoPartKey(ctx, conn, directoryID, samAccountName)
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

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DirectoryServiceDataClient(ctx)

	var state userResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findUserByTwoPartKey(ctx, conn, state.DirectoryID.ValueString(), state.SAMAccountName.ValueString())

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, userResourceID(state.DirectoryID.ValueString(), state.SAMAccountName.ValueString()))
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DirectoryServiceDataClient(ctx)

	var plan, state userResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	inputs := map[awstypes.UpdateType]*directoryservicedata.UpdateUserInput{}

	updateInput := func(updateType awstypes.UpdateType) *directoryservicedata.UpdateUserInput {
		if input, ok := inputs[updateType]; ok {
			return input
		}

		input := &directoryservicedata.UpdateUserInput{
			DirectoryId:    plan.DirectoryID.ValueStringPointer(),
			SAMAccountName: plan.SAMAccountName.ValueStringPointer(),
			UpdateType:     updateType,
		}
		inputs[updateType] = input
		return input
	}

	updateType := func(planValue, stateValue types.String) awstypes.UpdateType {
		switch {
		case stateValue.IsNull():
			return awstypes.UpdateTypeAdd
		case planValue.IsNull():
			return awstypes.UpdateTypeRemove
		default:
			return awstypes.UpdateTypeReplace
		}
	}

	if !plan.EmailAddress.Equal(state.EmailAddress) {
		input := updateInput(updateType(plan.EmailAddress, state.EmailAddress))
		input.EmailAddress = plan.EmailAddress.ValueStringPointer()
	}

	if !plan.GivenName.Equal(state.GivenName) {
		input := updateInput(updateType(plan.GivenName, state.GivenName))
		input.GivenName = plan.GivenName.ValueStringPointer()
	}

	if !plan.Surname.Equal(state.Surname) {
		input := updateInput(updateType(plan.Surname, state.Surname))
		input.Surname = plan.Surname.ValueStringPointer()
	}

	// For when more than one attribute needs to be updated
	for _, updateType := range []awstypes.UpdateType{awstypes.UpdateTypeAdd, awstypes.UpdateTypeReplace, awstypes.UpdateTypeRemove} {
		input, ok := inputs[updateType]
		if !ok {
			continue
		}

		_, err := conn.UpdateUser(ctx, input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, userResourceID(plan.DirectoryID.ValueString(), plan.SAMAccountName.ValueString()))
			return
		}
	}

	out, err := findUserByTwoPartKey(ctx, conn, plan.DirectoryID.ValueString(), plan.SAMAccountName.ValueString())

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, userResourceID(plan.DirectoryID.ValueString(), plan.SAMAccountName.ValueString()))
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DirectoryServiceDataClient(ctx)

	var state userResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := directoryservicedata.DeleteUserInput{
		DirectoryId:    state.DirectoryID.ValueStringPointer(),
		SAMAccountName: state.SAMAccountName.ValueStringPointer(),
	}

	_, err := conn.DeleteUser(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, userResourceID(state.DirectoryID.ValueString(), state.SAMAccountName.ValueString()))
		return
	}
}

func findUserByTwoPartKey(ctx context.Context, conn *directoryservicedata.Client, directoryID, samAccountName string) (*directoryservicedata.DescribeUserOutput, error) {
	input := directoryservicedata.DescribeUserInput{
		DirectoryId:    aws.String(directoryID),
		SAMAccountName: aws.String(samAccountName),
	}

	out, err := conn.DescribeUser(ctx, &input)
	// Once the parent Directory Service directory is deleted, DescribeUser
	// can no longer resolve authorization and returns AccessDeniedException
	// instead of ResourceNotFoundException.
	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.AccessDeniedException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return out, nil
}

type userResourceModel struct {
	framework.WithRegionModel
	DirectoryID       types.String `tfsdk:"directory_id"`
	DistinguishedName types.String `tfsdk:"distinguished_name"`
	EmailAddress      types.String `tfsdk:"email_address"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	GivenName         types.String `tfsdk:"given_name"`
	Realm             types.String `tfsdk:"realm"`
	SAMAccountName    types.String `tfsdk:"sam_account_name"`
	SID               types.String `tfsdk:"sid"`
	Surname           types.String `tfsdk:"surname"`
	UserPrincipalName types.String `tfsdk:"user_principal_name"`
}

var (
	_ inttypes.ImportIDParser = userImportID{}
)

type userImportID struct{}

func (userImportID) Parse(id string) (string, map[string]any, error) {
	parts, err := intflex.ExpandResourceId(id, userResourceIDPartCount, false)
	if err != nil {
		return "", nil, smarterr.NewError(err)
	}

	result := map[string]any{
		"directory_id":     parts[0],
		"sam_account_name": parts[1],
	}

	return id, result, nil
}

func userResourceID(directoryID, samAccountName string) string {
	id, _ := intflex.FlattenResourceId(
		[]string{directoryID, samAccountName},
		userResourceIDPartCount,
		false,
	)

	return id
}
