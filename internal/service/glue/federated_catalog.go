// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_glue_federated_catalog", name="Federated Catalog")
func newResourceFederatedCatalog(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceFederatedCatalog{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameFederatedCatalog = "Federated Catalog"
	s3TablesCatalogName     = "s3tablescatalog"
)

type resourceFederatedCatalog struct {
	framework.ResourceWithModel[resourceFederatedCatalogModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceFederatedCatalog) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCatalogID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"federated_catalog": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[federatedCatalogModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"connection_name": schema.StringAttribute{
							Optional: true,
						},
						names.AttrIdentifier: schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"catalog_properties": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[catalogPropertiesModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"data_lake_access_properties": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dataLakeAccessPropertiesModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"catalog_type": schema.StringAttribute{
										Optional: true,
									},
									"data_lake_access": schema.BoolAttribute{
										Optional: true,
									},
									"data_transfer_role": schema.StringAttribute{
										Optional: true,
									},
									names.AttrKMSKey: schema.StringAttribute{
										Optional: true,
									},
								},
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

func (r *resourceFederatedCatalog) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().GlueClient(ctx)
	var plan resourceFederatedCatalogModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input glue.CreateCatalogInput

	input.Name = plan.Name.ValueStringPointer()

	if plan.CatalogId.IsNull() || plan.CatalogId.ValueString() == "" {
		plan.CatalogId = types.StringValue(r.Meta().AccountID(ctx))
	}

	input.CatalogInput = &awstypes.CatalogInput{}

	if !plan.Description.IsNull() {
		input.CatalogInput.Description = plan.Description.ValueStringPointer()
	}

	if plan.FederatedCatalog.IsNull() && plan.CatalogProperties.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Configuration",
			"At least one of 'federated_catalog' or 'catalog_properties' must be specified for a federated catalog.",
		)
		return
	}

	if !plan.FederatedCatalog.IsNull() {
		var fedCatalog []federatedCatalogModel
		resp.Diagnostics.Append(plan.FederatedCatalog.ElementsAs(ctx, &fedCatalog, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(fedCatalog) > 0 {
			fc := &awstypes.FederatedCatalog{}

			if !fedCatalog[0].ConnectionName.IsNull() {
				fc.ConnectionName = fedCatalog[0].ConnectionName.ValueStringPointer()
			}
			if !fedCatalog[0].Identifier.IsNull() {
				fc.Identifier = fedCatalog[0].Identifier.ValueStringPointer()
			}

			input.CatalogInput.FederatedCatalog = fc
		}
	}

	if !plan.CatalogProperties.IsNull() {
		var catalogProps []catalogPropertiesModel
		resp.Diagnostics.Append(plan.CatalogProperties.ElementsAs(ctx, &catalogProps, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(catalogProps) > 0 {
			cp := &awstypes.CatalogProperties{}

			if !catalogProps[0].DataLakeAccessProperties.IsNull() {
				var dataLakeProps []dataLakeAccessPropertiesModel
				resp.Diagnostics.Append(catalogProps[0].DataLakeAccessProperties.ElementsAs(ctx, &dataLakeProps, false)...)
				if resp.Diagnostics.HasError() {
					return
				}

				if len(dataLakeProps) > 0 {
					dlap := &awstypes.DataLakeAccessProperties{}

					if !dataLakeProps[0].CatalogType.IsNull() {
						dlap.CatalogType = dataLakeProps[0].CatalogType.ValueStringPointer()
					}
					if !dataLakeProps[0].DataLakeAccess.IsNull() {
						dlap.DataLakeAccess = dataLakeProps[0].DataLakeAccess.ValueBool()
					}
					if !dataLakeProps[0].DataTransferRole.IsNull() {
						dlap.DataTransferRole = dataLakeProps[0].DataTransferRole.ValueStringPointer()
					}
					if !dataLakeProps[0].KmsKey.IsNull() {
						dlap.KmsKey = dataLakeProps[0].KmsKey.ValueStringPointer()
					}

					cp.DataLakeAccessProperties = dlap
				}
			}

			input.CatalogInput.CatalogProperties = cp
		}
	}

	input.CatalogInput.CreateDatabaseDefaultPermissions = []awstypes.PrincipalPermissions{}
	input.CatalogInput.CreateTableDefaultPermissions = []awstypes.PrincipalPermissions{}

	_, err := conn.CreateCatalog(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
		return
	}

	catalogId := plan.CatalogId.ValueString()
	catalogName := plan.Name.ValueString()
	if catalogName == s3TablesCatalogName {
		catalogId = fmt.Sprintf("%s:%s", catalogId, catalogName)
	}
	id := fmt.Sprintf("%s,%s", catalogId, catalogName)
	plan.ID = types.StringValue(id)
	catalog, err := findFederatedCatalogByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	if catalog.ResourceArn != nil {
		plan.ARN = types.StringValue(aws.ToString(catalog.ResourceArn))
	} else {
		partition := r.Meta().Partition(ctx)
		region := r.Meta().Region(ctx)
		accountID := r.Meta().AccountID(ctx)
		catalogName := plan.Name.ValueString()
		if catalogName == s3TablesCatalogName {
			plan.ARN = types.StringValue(fmt.Sprintf("arn:%s:glue:%s:%s:catalog/%s", partition, region, accountID, catalogName))
		} else {
			plan.ARN = types.StringValue(fmt.Sprintf("arn:%s:glue:%s:%s:catalog", partition, region, accountID))
		}
	}

	if catalog.CatalogId != nil {
		plan.CatalogId = types.StringValue(aws.ToString(catalog.CatalogId))
	}
	if catalog.FederatedCatalog != nil {
		fedCatalogModel := federatedCatalogModel{
			ConnectionName: types.StringPointerValue(catalog.FederatedCatalog.ConnectionName),
			Identifier:     types.StringPointerValue(catalog.FederatedCatalog.Identifier),
		}

		fedCatalogList, diags := fwtypes.NewListNestedObjectValueOfPtr(ctx, &fedCatalogModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.FederatedCatalog = fedCatalogList
	}

	if catalog.CatalogProperties != nil && catalog.CatalogProperties.DataLakeAccessProperties != nil {
		dlap := catalog.CatalogProperties.DataLakeAccessProperties
		if !plan.CatalogProperties.IsNull() || dlap.CatalogType != nil || dlap.DataTransferRole != nil || dlap.KmsKey != nil {
			dataLakePropsModel := dataLakeAccessPropertiesModel{
				CatalogType:      types.StringPointerValue(catalog.CatalogProperties.DataLakeAccessProperties.CatalogType),
				DataLakeAccess:   types.BoolValue(catalog.CatalogProperties.DataLakeAccessProperties.DataLakeAccess),
				DataTransferRole: types.StringPointerValue(catalog.CatalogProperties.DataLakeAccessProperties.DataTransferRole),
				KmsKey:           types.StringPointerValue(catalog.CatalogProperties.DataLakeAccessProperties.KmsKey),
			}

			dataLakeList, diags := fwtypes.NewListNestedObjectValueOfPtr(ctx, &dataLakePropsModel)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			catalogPropsModel := catalogPropertiesModel{
				DataLakeAccessProperties: dataLakeList,
			}

			catalogPropsList, diags := fwtypes.NewListNestedObjectValueOfPtr(ctx, &catalogPropsModel)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			plan.CatalogProperties = catalogPropsList
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceFederatedCatalog) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().GlueClient(ctx)
	var state resourceFederatedCatalogModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findFederatedCatalogByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	if out.ResourceArn != nil {
		state.ARN = types.StringValue(aws.ToString(out.ResourceArn))
	} else {
		partition := r.Meta().Partition(ctx)
		region := r.Meta().Region(ctx)
		accountID := r.Meta().AccountID(ctx)
		catalogName := state.Name.ValueString()
		if catalogName == s3TablesCatalogName {
			state.ARN = types.StringValue(fmt.Sprintf("arn:%s:glue:%s:%s:catalog/%s", partition, region, accountID, catalogName))
		} else {
			state.ARN = types.StringValue(fmt.Sprintf("arn:%s:glue:%s:%s:catalog", partition, region, accountID))
		}
	}

	if out.CatalogId != nil {
		state.CatalogId = types.StringValue(aws.ToString(out.CatalogId))
	}

	if out.Name != nil {
		state.Name = types.StringValue(aws.ToString(out.Name))
	}

	if out.Description != nil {
		state.Description = types.StringValue(aws.ToString(out.Description))
	}
	if out.FederatedCatalog != nil {
		fedCatalogModel := federatedCatalogModel{
			ConnectionName: types.StringPointerValue(out.FederatedCatalog.ConnectionName),
			Identifier:     types.StringPointerValue(out.FederatedCatalog.Identifier),
		}

		fedCatalogList, diags := fwtypes.NewListNestedObjectValueOfPtr(ctx, &fedCatalogModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.FederatedCatalog = fedCatalogList
	}

	if out.CatalogProperties != nil && out.CatalogProperties.DataLakeAccessProperties != nil {
		dlap := out.CatalogProperties.DataLakeAccessProperties
		if !state.CatalogProperties.IsNull() || dlap.CatalogType != nil || dlap.DataTransferRole != nil || dlap.KmsKey != nil {
			dataLakePropsModel := dataLakeAccessPropertiesModel{
				CatalogType:      types.StringPointerValue(out.CatalogProperties.DataLakeAccessProperties.CatalogType),
				DataLakeAccess:   types.BoolValue(out.CatalogProperties.DataLakeAccessProperties.DataLakeAccess),
				DataTransferRole: types.StringPointerValue(out.CatalogProperties.DataLakeAccessProperties.DataTransferRole),
				KmsKey:           types.StringPointerValue(out.CatalogProperties.DataLakeAccessProperties.KmsKey),
			}

			dataLakeList, diags := fwtypes.NewListNestedObjectValueOfPtr(ctx, &dataLakePropsModel)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			catalogPropsModel := catalogPropertiesModel{
				DataLakeAccessProperties: dataLakeList,
			}

			catalogPropsList, diags := fwtypes.NewListNestedObjectValueOfPtr(ctx, &catalogPropsModel)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			state.CatalogProperties = catalogPropsList
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceFederatedCatalog) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceFederatedCatalogModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		resp.Diagnostics.AddError(
			"Update Not Supported",
			"AWS Glue federated catalogs do not support updates. All attributes require replacement.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceFederatedCatalog) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().GlueClient(ctx)
	var state resourceFederatedCatalogModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogId, name, err := readCatalogResourceID(state.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	input := glue.DeleteCatalogInput{
		CatalogId: aws.String(resolveCatalogID(catalogId, name)),
	}

	_, err = conn.DeleteCatalog(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}
}

func findFederatedCatalogByID(ctx context.Context, conn *glue.Client, id string) (*awstypes.Catalog, error) {
	catalogId, name, err := readCatalogResourceID(id)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	input := glue.GetCatalogInput{
		CatalogId: aws.String(catalogId),
	}

	out, err := conn.GetCatalog(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Catalog == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	actualName := aws.ToString(out.Catalog.Name)
	if actualName != name {
		return nil, smarterr.NewError(&retry.NotFoundError{
			Message:     fmt.Sprintf("catalog name mismatch: expected %s, got %s", name, actualName),
			LastRequest: &input,
		})
	}

	return out.Catalog, nil
}

func readCatalogResourceID(id string) (catalogId, name string, err error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected catalog_id,name", id)
	}
	return parts[0], parts[1], nil
}

func resolveCatalogID(catalogId, name string) string {
	if name == s3TablesCatalogName {
		return name
	}
	return catalogId
}

type resourceFederatedCatalogModel struct {
	framework.WithRegionModel
	ARN               types.String                                            `tfsdk:"arn"`
	CatalogId         types.String                                            `tfsdk:"catalog_id"`
	CatalogProperties fwtypes.ListNestedObjectValueOf[catalogPropertiesModel] `tfsdk:"catalog_properties"`
	Description       types.String                                            `tfsdk:"description"`
	FederatedCatalog  fwtypes.ListNestedObjectValueOf[federatedCatalogModel]  `tfsdk:"federated_catalog"`
	ID                types.String                                            `tfsdk:"id"`
	Name              types.String                                            `tfsdk:"name"`
	Timeouts          timeouts.Value                                          `tfsdk:"timeouts"`
}

type federatedCatalogModel struct {
	ConnectionName types.String `tfsdk:"connection_name"`
	Identifier     types.String `tfsdk:"identifier"`
}

type catalogPropertiesModel struct {
	DataLakeAccessProperties fwtypes.ListNestedObjectValueOf[dataLakeAccessPropertiesModel] `tfsdk:"data_lake_access_properties"`
}

type dataLakeAccessPropertiesModel struct {
	CatalogType      types.String `tfsdk:"catalog_type"`
	DataLakeAccess   types.Bool   `tfsdk:"data_lake_access"`
	DataTransferRole types.String `tfsdk:"data_transfer_role"`
	KmsKey           types.String `tfsdk:"kms_key"`
}
