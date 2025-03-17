// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_api_gateway_domain_name_access_association", name="Domain Name Access Association")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigateway;types.DomainNameAccessAssociation")
// @Testing(generator="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.RandomSubdomain()")
// @Testing(tlsKey=true, tlsKeyDomain="rName")
func newDomainNameAccessAssociationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &domainNameAccessAssociationResource{}

	return r, nil
}

type domainNameAccessAssociationResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[domainNameAccessAssociationResourceModel]
	framework.WithImportByID
}

func (r *domainNameAccessAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_association_source": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"access_association_source_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AccessAssociationSourceType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"domain_name_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID:      framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrARN)),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *domainNameAccessAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data domainNameAccessAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayClient(ctx)

	input := apigateway.CreateDomainNameAccessAssociationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateDomainNameAccessAssociation(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating API Gateway Domain Name Access Association", err.Error())

		return
	}

	// Set values for unknowns.
	data.DomainNameAccessAssociationARN = fwflex.StringToFramework(ctx, output.DomainNameAccessAssociationArn)
	data.ID = data.DomainNameAccessAssociationARN

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *domainNameAccessAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data domainNameAccessAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayClient(ctx)

	output, err := findDomainNameAccessAssociationByARN(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading API Gateway Domain Name Access Association (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *domainNameAccessAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data domainNameAccessAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayClient(ctx)

	input := apigateway.DeleteDomainNameAccessAssociationInput{
		DomainNameAccessAssociationArn: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeleteDomainNameAccessAssociation(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting API Gateway Domain Name Access Association (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findDomainNameAccessAssociationByARN(ctx context.Context, conn *apigateway.Client, arn string) (*awstypes.DomainNameAccessAssociation, error) {
	input := apigateway.GetDomainNameAccessAssociationsInput{
		ResourceOwner: awstypes.ResourceOwnerSelf,
	}

	return findDomainNameAccessAssociation(ctx, conn, &input, func(v *awstypes.DomainNameAccessAssociation) bool {
		return aws.ToString(v.DomainNameAccessAssociationArn) == arn
	})
}

func findDomainNameAccessAssociation(ctx context.Context, conn *apigateway.Client, input *apigateway.GetDomainNameAccessAssociationsInput, filter tfslices.Predicate[*awstypes.DomainNameAccessAssociation]) (*awstypes.DomainNameAccessAssociation, error) {
	output, err := findDomainNameAccessAssociations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDomainNameAccessAssociations(ctx context.Context, conn *apigateway.Client, input *apigateway.GetDomainNameAccessAssociationsInput, filter tfslices.Predicate[*awstypes.DomainNameAccessAssociation]) ([]awstypes.DomainNameAccessAssociation, error) {
	var output []awstypes.DomainNameAccessAssociation

	err := getDomainNameAccessAssociationsPages(ctx, conn, input, func(page *apigateway.GetDomainNameAccessAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

type domainNameAccessAssociationResourceModel struct {
	AccessAssociationSource        types.String                                             `tfsdk:"access_association_source"`
	AccessAssociationSourceType    fwtypes.StringEnum[awstypes.AccessAssociationSourceType] `tfsdk:"access_association_source_type"`
	DomainNameAccessAssociationARN types.String                                             `tfsdk:"arn"`
	DomainNameARN                  fwtypes.ARN                                              `tfsdk:"domain_name_arn"`
	ID                             types.String                                             `tfsdk:"id"`
	Tags                           tftags.Map                                               `tfsdk:"tags"`
	TagsAll                        tftags.Map                                               `tfsdk:"tags_all"`
}

func (model *domainNameAccessAssociationResourceModel) InitFromID() error {
	model.DomainNameAccessAssociationARN = model.ID

	return nil
}
