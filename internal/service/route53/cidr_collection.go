// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
func newCIDRCollectionResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &cidrCollectionResource{}

	return r, nil
}

type cidrCollectionResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithImportByID
}

func (*cidrCollectionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_route53_cidr_collection"
}

func (r *cidrCollectionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(64),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), `can include letters, digits, underscore (_) and the dash (-) character`),
				},
			},
			names.AttrVersion: schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (r *cidrCollectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data cidrCollectionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53Client(ctx)

	name := data.Name.ValueString()
	input := &route53.CreateCidrCollectionInput{
		CallerReference: aws.String(id.UniqueId()),
		Name:            aws.String(name),
	}

	const (
		timeout = 2 * time.Minute
	)
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.ConcurrentModification](ctx, timeout, func() (interface{}, error) {
		return conn.CreateCidrCollection(ctx, input)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Route 53 CIDR Collection (%s)", name), err.Error())

		return
	}

	output := outputRaw.(*route53.CreateCidrCollectionOutput)
	data.ARN = fwflex.StringToFramework(ctx, output.Collection.Arn)
	data.ID = fwflex.StringToFramework(ctx, output.Collection.Id)
	data.Version = fwflex.Int64ToFramework(ctx, output.Collection.Version)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *cidrCollectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data cidrCollectionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53Client(ctx)

	output, err := findCIDRCollectionByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Route 53 CIDR Collection (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.ARN = fwflex.StringToFramework(ctx, output.Arn)
	data.Name = fwflex.StringToFramework(ctx, output.Name)
	data.Version = fwflex.Int64ToFramework(ctx, output.Version)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *cidrCollectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data cidrCollectionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53Client(ctx)

	tflog.Debug(ctx, "deleting Route 53 CIDR Collection", map[string]interface{}{
		names.AttrID: data.ID.ValueString(),
	})

	_, err := conn.DeleteCidrCollection(ctx, &route53.DeleteCidrCollectionInput{
		Id: fwflex.StringFromFramework(ctx, data.ID),
	})

	if errs.IsA[*awstypes.NoSuchCidrCollectionException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Route 53 CIDR Collection (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

type cidrCollectionResourceModel struct {
	ARN     types.String `tfsdk:"arn"`
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`
}

func findCIDRCollectionByID(ctx context.Context, conn *route53.Client, id string) (*awstypes.CollectionSummary, error) {
	input := &route53.ListCidrCollectionsInput{}

	return findCIDRCollection(ctx, conn, input, func(v *awstypes.CollectionSummary) bool {
		return aws.ToString(v.Id) == id
	})
}

func findCIDRCollection(ctx context.Context, conn *route53.Client, input *route53.ListCidrCollectionsInput, filter tfslices.Predicate[*awstypes.CollectionSummary]) (*awstypes.CollectionSummary, error) {
	output, err := findCIDRCollections(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCIDRCollections(ctx context.Context, conn *route53.Client, input *route53.ListCidrCollectionsInput, filter tfslices.Predicate[*awstypes.CollectionSummary]) ([]awstypes.CollectionSummary, error) {
	var output []awstypes.CollectionSummary

	pages := route53.NewListCidrCollectionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.CidrCollections {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
