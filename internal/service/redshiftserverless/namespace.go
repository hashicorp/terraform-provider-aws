// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"errors"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshiftserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_redshiftserverless_namespace", name="Namespace")
// @Tags(identifierAttribute="arn")
func newResourceNamespace(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceNamespace{}

	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameNamespace     = "Namespace"
	DefaultAdminUsername = "admin"
)

type resourceNamespace struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceNamespace) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"admin_password_secret_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"admin_password_secret_kms_key_id": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			"admin_user_password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("admin_user_password_wo")),
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("admin_user_password_wo")),
				},
			},
			"admin_user_password_wo": schema.StringAttribute{
				Optional:  true,
				WriteOnly: true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("admin_user_password_wo_version")),
					stringvalidator.ConflictsWith(path.MatchRoot("admin_user_password")),
				},
			},
			"admin_user_password_wo_version": schema.Int32Attribute{
				Optional: true,
				Validators: []validator.Int32{
					int32validator.AlsoRequires(path.MatchRoot("admin_user_password_wo")),
				},
			},
			"admin_username": schema.StringAttribute{
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"db_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				// Default and Computed are required to prevent the drift.
				// CreateNamespace API call does not set the default database name,
				// but once workgroup is associated, the namespace is updated behind the scenes with "dev" as the database name.
				Default: stringdefault.StaticString("dev"),
			},
			"default_iam_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			names.AttrID: framework.IDAttribute(),
			"iam_roles": schema.SetAttribute{
				CustomType:  fwtypes.SetOfARNType,
				ElementType: fwtypes.ARNType,
				Optional:    true,
			}, // This cannot be of ARNType since the default value is "AWS_OWNED_KMS_KEY" string
			names.AttrKMSKeyID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"log_exports": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
			},
			"manage_admin_password": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"namespace_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
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

func (r resourceNamespace) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config resourceNamespaceData
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.AdminUsername.IsNull() && config.AdminUsername.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("admin_username"),
			"Invalid Configuration",
			"Attribute admin_user_name can't be blank. Provide a value or remove it from configuration.")
	}

	if config.ManageAdminPassword.IsNull() || !config.ManageAdminPassword.ValueBool() {
		if !config.AdminUsername.IsNull() && config.AdminUserPassword.IsNull() && config.AdminUserPasswordWO.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("admin_user_password"),
				"Invalid Configuration",
				"You must specify admin_user_name if you provide an admin user password. The password can be set using either admin_user_password or admin_user_password_wo. You may specify one of these, both, or neither—but not just admin_user_password_wo alone.")
		}
	} else {
		if !config.AdminUserPassword.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("admin_user_password"),
				"Invalid Configuration",
				"The admin_user_password cannot be provided if manage_admin_password is true.")
		}
		if !config.AdminUserPasswordWO.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("admin_user_password_wo"),
				"Invalid Configuration",
				"The admin_user_password_wo cannot be provided if manage_admin_password is true.")
		}
	}

	if !config.DefaultIAMRoleARN.IsNull() && config.DefaultIAMRoleARN.ValueString() != "" {
		if slices.Index(config.IAMRoles.Elements(), attr.Value(config.DefaultIAMRoleARN)) == -1 {
			resp.Diagnostics.AddAttributeError(path.Root("default_iam_role_arn"), "Invalid Configuration", "The default_iam_role_arn must be in the list of iam_roles.")
		}
	}

	if !config.KMSKeyID.IsNull() && !config.KMSKeyID.IsUnknown() && !arn.IsARN(config.KMSKeyID.ValueString()) {
		resp.Diagnostics.AddAttributeError(path.Root(names.AttrKMSKeyID), "Invalid Configuration", "The provided value cannot be parsed as an ARN.\n\n"+
			"Value: "+config.KMSKeyID.ValueString())
	}
}

func (r *resourceNamespace) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RedshiftServerlessClient(ctx)

	var plan, config resourceNamespaceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &redshiftserverless.CreateNamespaceInput{
		NamespaceName: fwflex.StringFromFramework(ctx, plan.NamespaceName),
		Tags:          getTagsIn(ctx),
	}

	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.AdminUserPasswordWO.IsNull() {
		in.AdminUserPassword = fwflex.StringFromFramework(ctx, config.AdminUserPasswordWO)
	}

	out, err := conn.CreateNamespace(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(create.ProblemStandardMessage(names.RedshiftServerless, create.ErrActionCreating, ResNameNamespace, plan.NamespaceName.String(), err), err.Error())
		return
	}
	if out == nil || out.Namespace == nil {
		resp.Diagnostics.AddError(create.ProblemStandardMessage(names.RedshiftServerless, create.ErrActionCreating, ResNameNamespace, plan.NamespaceName.String(), nil), errors.New("empty output").Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out.Namespace, &plan, fwflex.WithIgnoredFieldNamesAppend("IamRoles"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = fwflex.StringToFramework(ctx, out.Namespace.NamespaceName)
	plan.ARN = fwflex.StringToFrameworkARN(ctx, out.Namespace.NamespaceArn)

	if len(out.Namespace.IamRoles) > 0 {
		plan.IAMRoles = fwtypes.NewSetValueOfMust[fwtypes.ARN](ctx, flattenNamespaceIAMRoles(ctx, out.Namespace.IamRoles))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceNamespace) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RedshiftServerlessClient(ctx)

	var state resourceNamespaceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findNamespaceByName(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(create.ProblemStandardMessage(names.RedshiftServerless, create.ErrActionSetting, ResNameNamespace, state.ID.String(), err), err.Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state, fwflex.WithIgnoredFieldNamesAppend("IamRoles"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ID = fwflex.StringToFramework(ctx, out.NamespaceName)
	state.ARN = fwflex.StringToFrameworkARN(ctx, out.NamespaceArn)
	// GetNamespace API call does not currently return value for ManageAdminPassword
	// The value is still required to detect drift, and can be derived from the presence of AdminPasswordSecretArn
	state.ManageAdminPassword = types.BoolValue(out.AdminPasswordSecretArn != nil)

	if len(out.IamRoles) > 0 {
		state.IAMRoles = fwtypes.NewSetValueOfMust[fwtypes.ARN](ctx, flattenNamespaceIAMRoles(ctx, out.IamRoles))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceNamespace) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().RedshiftServerlessClient(ctx)

	var plan, state, config resourceNamespaceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	update := func(in *redshiftserverless.UpdateNamespaceInput) (*awstypes.Namespace, error) {
		out, err := conn.UpdateNamespace(ctx, in)
		if err != nil {
			return nil, err
		}
		if out == nil || out.Namespace == nil {
			return nil, errors.New("empty output")
		}
		_, err = waitNamespaceUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
		if err != nil {
			return nil, err
		}
		return out.Namespace, nil
	}

	reportUpdateError := func(err error) {
		resp.Diagnostics.AddError(create.ProblemStandardMessage(names.RedshiftServerless, create.ErrActionUpdating, ResNameNamespace, plan.ID.String(), err), err.Error())
	}

	if !plan.ManageAdminPassword.Equal(state.ManageAdminPassword) ||
		!plan.AdminUsername.Equal(state.AdminUsername) ||
		!plan.AdminUserPassword.Equal(state.AdminUserPassword) ||
		!plan.AdminUserPasswordWOVersion.Equal(state.AdminUserPasswordWOVersion) ||
		!plan.AdminPasswordSecretKMSKeyID.Equal(state.AdminPasswordSecretKMSKeyID) {
		in := &redshiftserverless.UpdateNamespaceInput{
			NamespaceName:       fwflex.StringFromFramework(ctx, plan.NamespaceName),
			ManageAdminPassword: fwflex.BoolFromFramework(ctx, plan.ManageAdminPassword),
		}
		if plan.ManageAdminPassword.ValueBool() {
			if !plan.AdminUsername.Equal(state.AdminUsername) {
				in.AdminUsername = fwflex.StringFromFramework(ctx, plan.AdminUsername)
			}
			if !plan.AdminPasswordSecretKMSKeyID.Equal(state.AdminPasswordSecretKMSKeyID) {
				in.AdminPasswordSecretKmsKeyId = fwflex.StringFromFramework(ctx, plan.AdminPasswordSecretKMSKeyID)
			}
		} else {
			if !plan.AdminUsername.Equal(state.AdminUsername) || !plan.AdminUserPassword.Equal(state.AdminUserPassword) {
				in.AdminUsername = fwflex.StringFromFramework(ctx, plan.AdminUsername)
				in.AdminUserPassword = fwflex.StringFromFramework(ctx, plan.AdminUserPassword)
			}
			if !plan.AdminUsername.Equal(state.AdminUsername) || !plan.AdminUserPasswordWOVersion.Equal(state.AdminUserPasswordWOVersion) {
				in.AdminUsername = fwflex.StringFromFramework(ctx, plan.AdminUsername)
				in.AdminUserPassword = fwflex.StringFromFramework(ctx, config.AdminUserPasswordWO)
			}
		}
		out, err := update(in)
		if err != nil {
			reportUpdateError(err)
			return
		}
		plan.AdminPasswordSecretArn = fwflex.StringToFrameworkARN(ctx, out.AdminPasswordSecretArn)
	}

	if !plan.DefaultIAMRoleARN.Equal(state.DefaultIAMRoleARN) ||
		!plan.IAMRoles.Equal(state.IAMRoles) {
		in := &redshiftserverless.UpdateNamespaceInput{
			NamespaceName:     fwflex.StringFromFramework(ctx, plan.NamespaceName),
			DefaultIamRoleArn: fwflex.StringFromFramework(ctx, plan.DefaultIAMRoleARN),
			IamRoles:          fwflex.ExpandFrameworkStringValueSet(ctx, plan.IAMRoles),
		}
		// The API requires an empty string rather than nil to clear the default IAM role
		// and all framework conversions return nil for empty strings or ARNs
		if in.DefaultIamRoleArn == nil {
			in.DefaultIamRoleArn = aws.String("")
		}
		if _, err := update(in); err != nil {
			reportUpdateError(err)
			return
		}
	}

	if !(plan.KMSKeyID.IsUnknown() || plan.KMSKeyID.Equal(state.KMSKeyID)) {
		in := &redshiftserverless.UpdateNamespaceInput{
			NamespaceName: fwflex.StringFromFramework(ctx, plan.NamespaceName),
			KmsKeyId:      fwflex.StringFromFramework(ctx, plan.KMSKeyID),
		}
		out, err := update(in)
		if err != nil {
			reportUpdateError(err)
			return
		}
		plan.KMSKeyID = fwflex.StringToFramework(ctx, out.KmsKeyId)
	} else {
		plan.KMSKeyID = state.KMSKeyID
	}

	if !plan.LogExports.Equal(state.LogExports) {
		in := &redshiftserverless.UpdateNamespaceInput{
			NamespaceName: fwflex.StringFromFramework(ctx, plan.NamespaceName),
			LogExports:    fwflex.ExpandFrameworkStringyValueSet[awstypes.LogExport](ctx, plan.LogExports),
		}
		if _, err := update(in); err != nil {
			reportUpdateError(err)
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceNamespace) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RedshiftServerlessClient(ctx)

	var state resourceNamespaceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Deleting Redshift Serverless Namespace: %s", state.ID.ValueString())
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ConflictException](ctx, r.DeleteTimeout(ctx, state.Timeouts), func() (any, error) {
		return conn.DeleteNamespace(ctx, &redshiftserverless.DeleteNamespaceInput{
			NamespaceName: fwflex.StringFromFramework(ctx, state.ID),
		})
	}, // "ConflictException: There is an operation running on the namespace. Try deleting the namespace again later."
		"operation running")

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(create.ProblemStandardMessage(names.RedshiftServerless, create.ErrActionDeleting, ResNameNamespace, state.ID.String(), err), err.Error())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitNamespaceDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(create.ProblemStandardMessage(names.RedshiftServerless, create.ErrActionWaitingForDeletion, ResNameNamespace, state.ID.String(), err), err.Error())
		return
	}
}

func (r *resourceNamespace) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func (r *resourceNamespace) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// if not deleting
	if !req.Plan.Raw.IsNull() {
		var plan, config resourceNamespaceData

		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// if creating
		if req.State.Raw.IsNull() {
			// When creating, if admin_username is missing from config but password is specified, enforce the default value "admin"
			if config.AdminUsername.IsNull() && !(config.AdminUserPassword.IsNull() && config.AdminUserPasswordWO.IsNull()) {
				resp.Plan.SetAttribute(ctx, path.Root("admin_username"), types.StringValue(DefaultAdminUsername))
			}
		} else {
			var state resourceNamespaceData

			resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
			if resp.Diagnostics.HasError() {
				return
			}

			// When updating, if admin_username is missing from config, always enforce the default value "admin"
			if config.AdminUsername.IsNull() {
				resp.Plan.SetAttribute(ctx, path.Root("admin_username"), types.StringValue(DefaultAdminUsername))
			}

			// This covers two scenarios for admin_password_secret_arn attribute
			// in situations when none of password related attributes are changed:
			// 1. Suppressing (known after apply) during terraform plan
			// 2. Preventing "All values must be known after apply" error during apply
			if plan.AdminPasswordSecretArn.IsUnknown() {
				if !plan.ManageAdminPassword.ValueBool() {
					resp.Plan.SetAttribute(ctx, path.Root("admin_password_secret_arn"), types.StringNull())
				} else if !state.AdminPasswordSecretArn.IsNull() {
					// use state as unknown
					resp.Plan.SetAttribute(ctx, path.Root("admin_password_secret_arn"), state.AdminPasswordSecretArn)
				}
			}

			// If admin_username, in managed password scenario, is changing, this can be due to drift.
			// To reconcile drifted admin_password_secret_arn value with the new one about to be created
			// it is required to set the admin_password_secret_arn to unknown.
			if !state.AdminUsername.Equal(config.AdminUsername) &&
				!(state.AdminUsername.ValueString() == DefaultAdminUsername && config.AdminUsername.IsNull()) &&
				config.ManageAdminPassword.ValueBool() {
				resp.Plan.SetAttribute(ctx, path.Root("admin_password_secret_arn"), fwtypes.ARNUnknown())
			}

			// Once set, IAMRoles cannot be removed from config. Replacing the resource
			if (config.IAMRoles.IsNull() || len(config.IAMRoles.Elements()) == 0) &&
				!state.IAMRoles.IsNull() && len(state.IAMRoles.Elements()) > 0 {
				resp.RequiresReplace = []path.Path{path.Root("iam_roles")}
			}

			if config.KMSKeyID.IsNull() {
				if state.KMSKeyID.ValueString() == "AWS_OWNED_KMS_KEY" {
					// Suppress "known after apply"
					resp.Plan.SetAttribute(ctx, path.Root(names.AttrKMSKeyID), state.KMSKeyID)
				} else {
					// Once set to a non default value, KMSKeyID cannot be removed from config.
					// Replacing the resource.
					resp.RequiresReplace = []path.Path{path.Root(names.AttrKMSKeyID)}
					resp.Plan.SetAttribute(ctx, path.Root(names.AttrKMSKeyID), types.StringNull())
				}
			}
		}
	}
}

func waitNamespaceUpdated(ctx context.Context, conn *redshiftserverless.Client, id string, timeout time.Duration) (*awstypes.Namespace, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.NamespaceStatusModifying),
		Target:                    enum.Slice(awstypes.NamespaceStatusAvailable),
		Refresh:                   statusNamespace(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Namespace); ok {
		return out, err
	}

	return nil, err
}

func waitNamespaceDeleted(ctx context.Context, conn *redshiftserverless.Client, id string, timeout time.Duration) (*awstypes.Namespace, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NamespaceStatusDeleting),
		Target:  []string{},
		Refresh: statusNamespace(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Namespace); ok {
		return out, err
	}

	return nil, err
}

func statusNamespace(ctx context.Context, conn *redshiftserverless.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findNamespaceByName(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findNamespaceByName(ctx context.Context, conn *redshiftserverless.Client, id string) (*awstypes.Namespace, error) {
	in := &redshiftserverless.GetNamespaceInput{
		NamespaceName: aws.String(id),
	}

	output, err := conn.GetNamespace(ctx, in)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}
		return nil, err
	}

	if output == nil || output.Namespace == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return output.Namespace, nil
}

func flattenNamespaceIAMRoles(ctx context.Context, input []string) []attr.Value {
	var result []attr.Value

	// Input elements have the following format:
	// "IamRole(applyStatus=in-sync, iamRoleArn=arn:aws:iam::123456789012:role/service-role/test)"
	for _, roleString := range input {
		// Extract the part after "iamRoleArn="
		parts := strings.Split(roleString, "iamRoleArn=")
		if len(parts) != 2 {
			continue // Skip malformed strings
		}

		// Remove the closing parenthesis and any whitespace
		iamRoleArn := strings.TrimSpace(strings.TrimRight(parts[1], ")"))

		result = append(result, fwflex.StringToFrameworkARN(ctx, &iamRoleArn))
	}
	return result
}

type resourceNamespaceData struct {
	AdminPasswordSecretArn      fwtypes.ARN                     `tfsdk:"admin_password_secret_arn"`
	AdminPasswordSecretKMSKeyID fwtypes.ARN                     `tfsdk:"admin_password_secret_kms_key_id"`
	AdminUserPassword           types.String                    `tfsdk:"admin_user_password"`
	AdminUserPasswordWO         types.String                    `tfsdk:"admin_user_password_wo"`
	AdminUserPasswordWOVersion  types.Int32                     `tfsdk:"admin_user_password_wo_version"`
	AdminUsername               types.String                    `tfsdk:"admin_username"`
	ARN                         fwtypes.ARN                     `tfsdk:"arn"`
	DBName                      types.String                    `tfsdk:"db_name"`
	DefaultIAMRoleARN           fwtypes.ARN                     `tfsdk:"default_iam_role_arn"`
	IAMRoles                    fwtypes.SetValueOf[fwtypes.ARN] `tfsdk:"iam_roles"`
	ID                          types.String                    `tfsdk:"id"`
	KMSKeyID                    types.String                    `tfsdk:"kms_key_id"`
	LogExports                  types.Set                       `tfsdk:"log_exports"`
	ManageAdminPassword         types.Bool                      `tfsdk:"manage_admin_password"`
	NamespaceID                 types.String                    `tfsdk:"namespace_id"`
	NamespaceName               types.String                    `tfsdk:"namespace_name"`
	Tags                        tftags.Map                      `tfsdk:"tags"`
	TagsAll                     tftags.Map                      `tfsdk:"tags_all"`
	Timeouts                    timeouts.Value                  `tfsdk:"timeouts"`
}
