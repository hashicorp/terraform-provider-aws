package quicksight

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2resource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
func newResourceFolder(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceFolder{}, nil
}

const (
	ResNameFolder = "Folder"
)

type resourceFolder struct {
	framework.ResourceWithConfigure
}

func (r *resourceFolder) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_quicksight_folder"
}

func (r *resourceFolder) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed: true,
			},
			"aws_account_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"folder_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"folder_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
			},
			"parent_folder_arn": schema.StringAttribute{
				Optional: true,
			},
			"tags":     tftags.TagsAttribute(),
			"tags_all": tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"permissions": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 2048),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"principal": schema.StringAttribute{
							Required: true,
						},
						"actions": schema.SetAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (r *resourceFolder) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightConn()

	var plan resourceFolderData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID)
	}
	plan.ID = types.StringValue(createFolderID(plan.AWSAccountID.ValueString(), plan.FolderID.ValueString()))

	in := quicksight.CreateFolderInput{
		AwsAccountId: aws.String(plan.AWSAccountID.ValueString()),
		FolderId:     aws.String(plan.FolderID.ValueString()),
		Name:         aws.String(plan.Name.ValueString()),
	}

	if !plan.FolderType.IsNull() && !plan.FolderType.IsUnknown() {
		in.FolderType = aws.String(plan.FolderType.ValueString())
	}
	if !plan.Permissions.IsNull() {
		var permissions []permissionsData
		resp.Diagnostics.Append(plan.Permissions.ElementsAs(ctx, &permissions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		permissionsObj, d := expandPermissions(ctx, permissions)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.Permissions = permissionsObj
	}

	out, err := conn.CreateFolderWithContext(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameFolder, plan.FolderID.String(), nil),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameFolder, plan.FolderID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state := plan
	resp.Diagnostics.Append(state.refresh(ctx, r.Meta())...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceFolder) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightConn()

	var state resourceFolderData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for existence before passing off to refresh
	_, err := FindFolderByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, ResNameFolder, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(state.refresh(ctx, r.Meta())...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceFolder) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().QuickSightConn()

	var plan, state resourceFolderData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.FolderType.Equal(state.FolderType) ||
		!plan.Name.Equal(state.Name) ||
		!plan.ParentFolderARN.Equal(state.ParentFolderARN) ||
		!plan.Permissions.Equal(state.Permissions) {

		in := quicksight.UpdateFolderInput{
			AwsAccountId: aws.String(plan.AWSAccountID.ValueString()),
			FolderId:     aws.String(plan.FolderID.ValueString()),
			Name:         aws.String(plan.Name.ValueString()),
		}

		// TODO: optional attrs

		out, err := conn.UpdateFolderWithContext(ctx, &in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, ResNameFolder, plan.ID.String(), nil),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, ResNameFolder, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(state.refresh(ctx, r.Meta())...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceFolder) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightConn()

	var state resourceFolderData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteFolderWithContext(ctx, &quicksight.DeleteFolderInput{
		AwsAccountId: aws.String(state.AWSAccountID.ValueString()),
		FolderId:     aws.String(state.FolderID.ValueString()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, ResNameFolder, state.ID.String(), nil),
			err.Error(),
		)
	}
}

func (r *resourceFolder) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceFolder) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func FindFolderByID(ctx context.Context, conn *quicksight.QuickSight, id string) (*quicksight.Folder, error) {
	awsAccountID, folderID, err := ParseFolderID(id)
	if err != nil {
		return nil, err
	}

	in := &quicksight.DescribeFolderInput{
		AwsAccountId: aws.String(awsAccountID),
		FolderId:     aws.String(folderID),
	}

	out, err := conn.DescribeFolderWithContext(ctx, in)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return nil, &sdkv2resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Folder == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Folder, nil
}

func FindFolderPermissionsByID(ctx context.Context, conn *quicksight.QuickSight, id string) ([]*quicksight.ResourcePermission, error) {
	awsAccountID, folderID, err := ParseFolderID(id)
	if err != nil {
		return nil, err
	}

	in := &quicksight.DescribeFolderPermissionsInput{
		AwsAccountId: aws.String(awsAccountID),
		FolderId:     aws.String(folderID),
	}

	out, err := conn.DescribeFolderPermissionsWithContext(ctx, in)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return nil, &sdkv2resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Permissions == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Permissions, nil
}

func ParseFolderID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID,FOLDER_ID", id)
	}
	return parts[0], parts[1], nil
}

func createFolderID(awsAccountID, folderID string) string {
	return fmt.Sprintf("%s,%s", awsAccountID, folderID)
}

var (
	permissionsAttrTypes = map[string]attr.Type{
		"actions":   types.SetType{ElemType: types.StringType},
		"principal": types.StringType,
	}
)

type resourceFolderData struct {
	ARN             types.String `tfsdk:"arn"`
	AWSAccountID    types.String `tfsdk:"aws_account_id"`
	FolderID        types.String `tfsdk:"folder_id"`
	FolderType      types.String `tfsdk:"folder_type"`
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	ParentFolderARN types.String `tfsdk:"parent_folder_arn"`
	Permissions     types.List   `tfsdk:"permissions"`
	Tags            types.Map    `tfsdk:"tags"`
	TagsAll         types.Map    `tfsdk:"tags_all"`
}

type permissionsData struct {
	Principal types.String `tfsdk:"principal"`
	Actions   types.Set    `tfsdk:"actions"`
}

// refresh handles reading attributes for an existing folder, as the Create and Update
// response objects contain only a limited set of attributes.
func (rd *resourceFolderData) refresh(ctx context.Context, meta *conns.AWSClient) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.QuickSightConn()

	out, err := FindFolderByID(ctx, conn, rd.ID.ValueString())
	if err != nil {
		diags.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameFolder, rd.ID.String(), nil),
			err.Error(),
		)
		return diags
	}
	if out == nil {
		return diags
	}

	// TODO: parse ID and set AWSAccountID for import
	rd.ARN = flex.StringToFramework(ctx, out.Arn)
	rd.FolderID = flex.StringToFramework(ctx, out.FolderId)
	rd.FolderType = flex.StringToFramework(ctx, out.FolderType)
	rd.Name = flex.StringToFramework(ctx, out.Name)

	permOut, err := FindFolderPermissionsByID(ctx, conn, rd.ID.ValueString())
	if err != nil {
		diags.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameFolder, rd.ID.String(), nil),
			err.Error(),
		)
		return diags
	}

	permissions, d := flattenFolderPermissions(ctx, permOut)
	diags.Append(d...)
	rd.Permissions = permissions

	// TODO: switch to transparent tagging
	tagsOut, err := ListTags(ctx, meta.QuickSightConn(), aws.StringValue(out.Arn))
	if err != nil {
		diags.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameFolder, rd.ID.String(), nil),
			err.Error(),
		)
		return diags
	}

	defaultTagsConfig := meta.DefaultTagsConfig
	ignoreTagsConfig := meta.IgnoreTagsConfig
	tags := tagsOut.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	// AWS APIs often return empty lists of tags when none have been configured.
	if tags := tags.RemoveDefaultConfig(defaultTagsConfig).Map(); len(tags) == 0 {
		rd.Tags = tftags.Null
	} else {
		rd.Tags = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags)
	}
	rd.TagsAll = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags.Map())

	return diags
}

func expandPermissions(ctx context.Context, tfSet []permissionsData) ([]*quicksight.ResourcePermission, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(tfSet) == 0 {
		return nil, diags
	}

	var apiObject []*quicksight.ResourcePermission
	for _, item := range tfSet {
		obj := &quicksight.ResourcePermission{
			Principal: aws.String(item.Principal.ValueString()),
			Actions:   flex.ExpandFrameworkStringSet(ctx, item.Actions),
		}
		apiObject = append(apiObject, obj)
	}

	return apiObject, diags
}

func flattenFolderPermissions(ctx context.Context, apiObject []*quicksight.ResourcePermission) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: permissionsAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, v := range apiObject {
		obj := map[string]attr.Value{
			"principal": flex.StringToFramework(ctx, v.Principal),
			"actions":   flex.FlattenFrameworkStringSet(ctx, v.Actions),
		}
		objVal, d := types.ObjectValue(permissionsAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}
	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}
