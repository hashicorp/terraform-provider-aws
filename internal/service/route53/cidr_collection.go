// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource
func newResourceCIDRCollection(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCIDRCollection{}

	return r, nil
}

type resourceCIDRCollection struct {
	framework.ResourceWithConfigure
}

func (r *resourceCIDRCollection) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_route53_cidr_collection"
}

func (r *resourceCIDRCollection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"id":  framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(64),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`), `can include letters, digits, underscore (_) and the dash (-) character`),
				},
			},
			"version": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (r *resourceCIDRCollection) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceCIDRCollectionData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53Conn(ctx)

	name := data.Name.ValueString()
	input := &route53.CreateCidrCollectionInput{
		CallerReference: aws.String(id.UniqueId()),
		Name:            aws.String(name),
	}

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (interface{}, error) {
		return conn.CreateCidrCollectionWithContext(ctx, input)
	}, route53.ErrCodeConcurrentModification)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Route 53 CIDR Collection (%s)", name), err.Error())

		return
	}

	output := outputRaw.(*route53.CreateCidrCollectionOutput)
	data.ARN = flex.StringToFramework(ctx, output.Collection.Arn)
	data.ID = flex.StringToFramework(ctx, output.Collection.Id)
	data.Version = flex.Int64ToFramework(ctx, output.Collection.Version)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceCIDRCollection) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceCIDRCollectionData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53Conn(ctx)

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

	data.ARN = flex.StringToFramework(ctx, output.Arn)
	data.Name = flex.StringToFramework(ctx, output.Name)
	data.Version = flex.Int64ToFramework(ctx, output.Version)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceCIDRCollection) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Noop.
}

func (r *resourceCIDRCollection) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceCIDRCollectionData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53Conn(ctx)

	tflog.Debug(ctx, "deleting Route 53 CIDR Collection", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	_, err := conn.DeleteCidrCollectionWithContext(ctx, &route53.DeleteCidrCollectionInput{
		Id: flex.StringFromFramework(ctx, data.ID),
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Route 53 CIDR Collection (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *resourceCIDRCollection) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

type resourceCIDRCollectionData struct {
	ARN     types.String `tfsdk:"arn"`
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Version types.Int64  `tfsdk:"version"`
}

func findCIDRCollectionByID(ctx context.Context, conn *route53.Route53, id string) (*route53.CollectionSummary, error) {
	input := &route53.ListCidrCollectionsInput{}
	var output *route53.CollectionSummary

	err := conn.ListCidrCollectionsPagesWithContext(ctx, input, func(page *route53.ListCidrCollectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CidrCollections {
			if v == nil {
				continue
			}

			if aws.StringValue(v.Id) == id {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}
