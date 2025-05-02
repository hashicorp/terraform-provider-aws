// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package simpledb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_simpledb_domain", name="Domain")
func newDomainResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &domainResource{}

	return r, nil
}

type domainResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithImportByID
}

func (r *domainResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		DeprecationMessage: `The aws_simpledb_domain resource has been deprecated and will be removed in a future version. Use Amazon DynamoDB instead.`,
	}
}

func (r *domainResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data domainResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := simpleDBConn(ctx, r.Meta())

	name := data.Name.ValueString()
	input := &simpledb.CreateDomainInput{
		DomainName: aws.String(name),
	}

	_, err := conn.CreateDomainWithContext(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SimpleDB Domain (%s)", name), err.Error())

		return
	}

	data.ID = types.StringValue(name)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *domainResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data domainResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := simpleDBConn(ctx, r.Meta())

	_, err := findDomainByName(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SimpleDB Domain (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.Name = data.ID

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *domainResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data domainResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := simpleDBConn(ctx, r.Meta())

	tflog.Debug(ctx, "deleting SimpleDB Domain", map[string]any{
		names.AttrID: data.ID.ValueString(),
	})

	_, err := conn.DeleteDomainWithContext(ctx, &simpledb.DeleteDomainInput{
		DomainName: flex.StringFromFramework(ctx, data.ID),
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting SimpleDB Domain (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

type domainResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func findDomainByName(ctx context.Context, conn *simpledb.SimpleDB, name string) (*simpledb.DomainMetadataOutput, error) {
	input := &simpledb.DomainMetadataInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DomainMetadataWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, simpledb.ErrCodeNoSuchDomain) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
