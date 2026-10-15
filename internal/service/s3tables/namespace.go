// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3tables_namespace", name="Namespace")
func newNamespaceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &namespaceResource{}, nil
}

type namespaceResource struct {
	framework.ResourceWithModel[namespaceResourceModel]
	framework.WithNoUpdate
}

func (r *namespaceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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

func (r *namespaceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data namespaceResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	namespace, tableBucketARN := fwflex.StringValueFromFramework(ctx, data.Namespace), fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	input := s3tables.CreateNamespaceInput{
		Namespace:      []string{namespace},
		TableBucketARN: aws.String(tableBucketARN),
	}

	_, err := conn.CreateNamespace(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Tables Namespace (%s)", namespace), err.Error())

		return
	}

	output, err := findNamespaceByTwoPartKey(ctx, conn, tableBucketARN, namespace)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Namespace (%s)", namespace), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *namespaceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data namespaceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	namespace, tableBucketARN := fwflex.StringValueFromFramework(ctx, data.Namespace), fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	output, err := findNamespaceByTwoPartKey(ctx, conn, tableBucketARN, namespace)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Tables Namespace (%s)", namespace), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *namespaceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data namespaceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3TablesClient(ctx)

	namespace, tableBucketARN := fwflex.StringValueFromFramework(ctx, data.Namespace), fwflex.StringValueFromFramework(ctx, data.TableBucketARN)
	input := s3tables.DeleteNamespaceInput{
		Namespace:      aws.String(namespace),
		TableBucketARN: aws.String(tableBucketARN),
	}
	_, err := conn.DeleteNamespace(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Tables Namespace (%s)", namespace), err.Error())

		return
	}
}

func (r *namespaceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	identifier, err := parseNamespaceIdentifier(request.ID)
	if err != nil {
		response.Diagnostics.AddError(
			"Invalid Import ID",
			"Import IDs for S3 Tables Namespaces must use the format <table bucket ARN>"+namespaceIDSeparator+"<namespace>.\n"+
				fmt.Sprintf("Had %q", request.ID),
		)
		return
	}

	identifier.PopulateState(ctx, &response.State, &response.Diagnostics)
}

func findNamespaceByTwoPartKey(ctx context.Context, conn *s3tables.Client, tableBucketARN, namespace string) (*s3tables.GetNamespaceOutput, error) {
	input := s3tables.GetNamespaceInput{
		Namespace:      aws.String(namespace),
		TableBucketARN: aws.String(tableBucketARN),
	}

	return findNamespace(ctx, conn, &input)
}

func findNamespace(ctx context.Context, conn *s3tables.Client, input *s3tables.GetNamespaceInput) (*s3tables.GetNamespaceOutput, error) {
	output, err := conn.GetNamespace(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type namespaceResourceModel struct {
	framework.WithRegionModel
	CreatedAt      timetypes.RFC3339 `tfsdk:"created_at"`
	CreatedBy      types.String      `tfsdk:"created_by"`
	Namespace      types.String      `tfsdk:"namespace" autoflex:"-"`
	OwnerAccountID types.String      `tfsdk:"owner_account_id"`
	TableBucketARN fwtypes.ARN       `tfsdk:"table_bucket_arn"`
}

var namespaceNameValidator = []validator.String{
	stringvalidator.LengthBetween(1, 255),
	tfstringvalidator.ContainsOnlyLowerCaseLettersNumbersUnderscores,
	tfstringvalidator.StartsWithLetterOrNumber,
	tfstringvalidator.EndsWithLetterOrNumber,
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
