// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

// @FrameworkResource("aws_s3tables_namespace", name="Namespace")
func newResourceNamespace(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceNamespace{}, nil
}

const (
	resNameNamespace = "Namespace"
)

type resourceNamespace struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
}

func (r *resourceNamespace) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
			names.AttrNamespace: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: namespaceNameValidator,
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
		},
	}
}

func (r *resourceNamespace) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var plan resourceNamespaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input s3tables.CreateNamespaceInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Namespace = []string{plan.Namespace.ValueString()}

	out, err := conn.CreateNamespace(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, resNameNamespace, plan.Namespace.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, resNameNamespace, plan.Namespace.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	namespace, err := findNamespace(ctx, conn, plan.TableBucketARN.ValueString(), out.Namespace[0])
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionCreating, resNameNamespace, plan.Namespace.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, namespace, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Namespace = types.StringValue(out.Namespace[0])

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceNamespace) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var state resourceNamespaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findNamespace(ctx, conn, state.TableBucketARN.ValueString(), state.Namespace.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionReading, resNameNamespace, state.Namespace.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceNamespace) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().S3TablesClient(ctx)

	var state resourceNamespaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := s3tables.DeleteNamespaceInput{
		Namespace:      state.Namespace.ValueStringPointer(),
		TableBucketARN: state.TableBucketARN.ValueStringPointer(),
	}

	_, err := conn.DeleteNamespace(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Tables, create.ErrActionDeleting, resNameNamespace, state.Namespace.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceNamespace) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	identifier, err := parseNamespaceIdentifier(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import IDs for S3 Tables Namespaces must use the format <table bucket ARN>"+namespaceIDSeparator+"<namespace>.\n"+
				fmt.Sprintf("Had %q", req.ID),
		)
		return
	}

	identifier.PopulateState(ctx, &resp.State, &resp.Diagnostics)
}

func findNamespace(ctx context.Context, conn *s3tables.Client, bucketARN, name string) (*s3tables.GetNamespaceOutput, error) {
	in := s3tables.GetNamespaceInput{
		Namespace:      aws.String(name),
		TableBucketARN: aws.String(bucketARN),
	}

	out, err := conn.GetNamespace(ctx, &in)
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

type resourceNamespaceModel struct {
	CreatedAt      timetypes.RFC3339 `tfsdk:"created_at"`
	CreatedBy      types.String      `tfsdk:"created_by"`
	Namespace      types.String      `tfsdk:"namespace" autoflex:"-"`
	OwnerAccountID types.String      `tfsdk:"owner_account_id"`
	TableBucketARN fwtypes.ARN       `tfsdk:"table_bucket_arn"`
}

var namespaceNameValidator = []validator.String{
	stringvalidator.LengthBetween(1, 255),
	stringMustContainLowerCaseLettersNumbersUnderscores,
	stringMustStartWithLetterOrNumber,
	stringMustEndWithLetterOrNumber,
}

type namespaceIdentifier struct {
	TableBucketARN string
	Namespace      string
}

const (
	namespaceIDSeparator = ";"
	namespaceIDParts     = 2
)

func parseNamespaceIdentifier(s string) (namespaceIdentifier, error) {
	parts := strings.Split(s, namespaceIDSeparator)
	if len(parts) != namespaceIDParts {
		return namespaceIdentifier{}, errors.New("not enough parts")
	}
	for i := range namespaceIDParts {
		if parts[i] == "" {
			return namespaceIdentifier{}, errors.New("empty part")
		}
	}

	return namespaceIdentifier{
		TableBucketARN: parts[0],
		Namespace:      parts[1],
	}, nil
}

func (id namespaceIdentifier) String() string {
	return id.TableBucketARN + tableIDSeparator +
		id.Namespace
}

func (id namespaceIdentifier) PopulateState(ctx context.Context, s *tfsdk.State, diags *diag.Diagnostics) {
	diags.Append(s.SetAttribute(ctx, path.Root("table_bucket_arn"), id.TableBucketARN)...)
	diags.Append(s.SetAttribute(ctx, path.Root(names.AttrNamespace), id.Namespace)...)
}
