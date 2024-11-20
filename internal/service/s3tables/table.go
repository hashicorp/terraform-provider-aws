// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3tables_table", name="Table")
func newResourceTable(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceTable{}, nil
}

const (
	ResNameTable = "Table"
)

type resourceTable struct {
	framework.ResourceWithConfigure
}

func (r *resourceTable) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_s3tables_table"
}

func (r *resourceTable) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrFormat: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.OpenTableFormat](),
				Required:   true,
				// TODO: Only one format is currently supported. When a new value is added, we can determine if `format` can be changed in-place or must recreate the resource
			},
			"metadata_location": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"modified_by": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-z_]+$`), "must contain only lowercase letters, numbers, or underscores"),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-z]`), "must start with a letter or number"),
					stringvalidator.RegexMatches(regexache.MustCompile(`[0-9a-z]$`), "must end with a letter or number"),
				},
			},
			names.AttrNamespace: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-z_]+$`), "must contain only lowercase letters, numbers, or underscores"),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-z]`), "must start with a letter or number"),
					stringvalidator.RegexMatches(regexache.MustCompile(`[0-9a-z]$`), "must end with a letter or number"),
				},
			},
			names.AttrOwnerAccountID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"table_bucket_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TableType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version_token": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"warehouse_location": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceTable) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var plan resourceTableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input s3tables.CreateTableInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Namespace = plan.Namespace.ValueStringPointer()

	_, err := conn.CreateTable(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, ResNameTable, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	table, err := findTable(ctx, conn, plan.TableBucketARN.ValueString(), plan.Namespace.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, ResNameTable, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, table, &plan, flex.WithFieldNamePrefix("Table"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Namespace = types.StringValue(table.Namespace[0])

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTable) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var state resourceTableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTable(ctx, conn, state.TableBucketARN.ValueString(), state.Namespace.ValueString(), state.Name.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionSetting, ResNameTable, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("Table"))...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Namespace = types.StringValue(out.Namespace[0])

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTable) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var plan, state resourceTableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) || !plan.Namespace.Equal(state.Namespace) {
		input := s3tables.RenameTableInput{
			TableBucketARN: state.TableBucketARN.ValueStringPointer(),
			Namespace:      state.Namespace.ValueStringPointer(),
			Name:           state.Name.ValueStringPointer(),
		}

		if !plan.Name.Equal(state.Name) {
			input.NewName = plan.Name.ValueStringPointer()
		}

		if !plan.Namespace.Equal(state.Namespace) {
			input.NewNamespaceName = plan.Namespace.ValueStringPointer()
		}

		_, err := conn.RenameTable(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.S3Tables, create.ErrActionUpdating, ResNameTable, state.Name.String(), err),
				err.Error(),
			)
		}

		table, err := findTable(ctx, conn, plan.TableBucketARN.ValueString(), plan.Namespace.ValueString(), plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, ResNameTable, plan.Name.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, table, &plan, flex.WithFieldNamePrefix("Table"))...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Namespace = types.StringValue(table.Namespace[0])
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceTable) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var state resourceTableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := s3tables.DeleteTableInput{
		Name:           state.Name.ValueStringPointer(),
		Namespace:      state.Namespace.ValueStringPointer(),
		TableBucketARN: state.TableBucketARN.ValueStringPointer(),
	}

	_, err := conn.DeleteTable(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionDeleting, ResNameTable, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceTable) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	identifier, err := parseTableIdentifier(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import IDs for S3 Tables Tables must use the format <table bucket ARN>"+namespaceIDSeparator+"<namespace>"+namespaceIDSeparator+"<table name>.\n"+
				fmt.Sprintf("Had %q", req.ID),
		)
		return
	}

	var state resourceTableModel
	identifier.Populate(&state)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func findTable(ctx context.Context, conn *s3tables.Client, bucketARN, namespace, name string) (*s3tables.GetTableOutput, error) {
	in := &s3tables.GetTableInput{
		Name:           aws.String(name),
		Namespace:      aws.String(namespace),
		TableBucketARN: aws.String(bucketARN),
	}

	out, err := conn.GetTable(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceTableModel struct {
	ARN               types.String                                 `tfsdk:"arn"`
	CreatedAt         timetypes.RFC3339                            `tfsdk:"created_at"`
	CreatedBy         types.String                                 `tfsdk:"created_by"`
	Format            fwtypes.StringEnum[awstypes.OpenTableFormat] `tfsdk:"format"`
	MetadataLocation  types.String                                 `tfsdk:"metadata_location"`
	ModifiedAt        timetypes.RFC3339                            `tfsdk:"modified_at"`
	ModifiedBy        types.String                                 `tfsdk:"modified_by"`
	Name              types.String                                 `tfsdk:"name"`
	Namespace         types.String                                 `tfsdk:"namespace" autoflex:"-"`
	OwnerAccountID    types.String                                 `tfsdk:"owner_account_id"`
	TableBucketARN    fwtypes.ARN                                  `tfsdk:"table_bucket_arn"`
	Type              fwtypes.StringEnum[awstypes.TableType]       `tfsdk:"type"`
	VersionToken      types.String                                 `tfsdk:"version_token"`
	WarehouseLocation types.String                                 `tfsdk:"warehouse_location"`
}

func tableIDFromTableARN(s string) (string, error) {
	arn, err := arn.Parse(s)
	if err != nil {
		return "", err
	}

	return tableIDFromTableARNResource(arn.Resource), nil
}

func tableIDFromTableARNResource(s string) string {
	parts := strings.SplitN(s, "/", 4)
	return parts[3]
}

type tableIdentifier struct {
	TableBucketARN string
	Namespace      string
	Name           string
}

const (
	tableIDSeparator = ";"
	tableIDParts     = 3
)

func parseTableIdentifier(s string) (tableIdentifier, error) {
	parts := strings.Split(s, tableIDSeparator)
	if len(parts) != tableIDParts {
		return tableIdentifier{}, errors.New("not enough parts")
	}
	for i := range tableIDParts {
		if parts[i] == "" {
			return tableIdentifier{}, errors.New("empty part")
		}
	}

	return tableIdentifier{
		TableBucketARN: parts[0],
		Namespace:      parts[1],
		Name:           parts[2],
	}, nil
}

func (id tableIdentifier) String() string {
	return id.TableBucketARN + tableIDSeparator +
		id.Namespace + tableIDSeparator +
		id.Name
}

func (id tableIdentifier) Populate(m *resourceTableModel) {
	m.TableBucketARN = fwtypes.ARNValue(id.TableBucketARN)
	m.Namespace = types.StringValue(id.Namespace)
	m.Name = types.StringValue(id.Name)
}
