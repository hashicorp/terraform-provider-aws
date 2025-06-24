// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspacesweb

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspacesweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_workspacesweb_trust_store", name="Trust Store")
// @Tags(identifierAttribute="trust_store_arn")
// @Testing(tagsTest=true)
// @Testing(generator=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workspacesweb/types;types.TrustStore")
// @Testing(importStateIdAttribute="trust_store_arn")
// @Testing(importIgnore="certificate_list")
func newTrustStoreResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &trustStoreResource{}, nil
}

type trustStoreResource struct {
	framework.ResourceWithConfigure
}

func (r *trustStoreResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"associated_portal_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"certificate_list": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"trust_store_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *trustStoreResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data trustStoreResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := &workspacesweb.CreateTrustStoreInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		Tags:        getTagsIn(ctx),
	}

	// Convert string certificates to byte slices
	for _, cert := range data.CertificateList.Elements() {
		// Remove the leading and trailing double quotes. Also remove the literals "\n" to replace with line feeds
		formatted_cert := strings.ReplaceAll(strings.Trim(cert.String(), "\""), `\n`, "\n")
		input.CertificateList = append(input.CertificateList, []byte(formatted_cert))
	}

	output, err := conn.CreateTrustStore(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating WorkSpacesWeb Trust Store", err.Error())
		return
	}

	data.TrustStoreARN = fwflex.StringToFramework(ctx, output.TrustStoreArn)

	// Get the trust store details to populate other fields
	trustStore, err := findTrustStoreByARN(ctx, conn, data.TrustStoreARN.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb Trust Store (%s)", data.TrustStoreARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, trustStore, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *trustStoreResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data trustStoreResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	output, err := findTrustStoreByARN(ctx, conn, data.TrustStoreARN.ValueString())
	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb Trust Store (%s)", data.TrustStoreARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *trustStoreResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old trustStoreResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	if !new.CertificateList.Equal(old.CertificateList) {
		input := &workspacesweb.UpdateTrustStoreInput{
			TrustStoreArn: aws.String(new.TrustStoreARN.ValueString()),
			ClientToken:   aws.String(sdkid.UniqueId()),
		}

		// Handle certificate additions and deletions
		oldCerts := make(map[string]bool)
		for _, cert := range old.CertificateList.Elements() {
			oldCerts[cert.String()] = true
		}

		newCerts := make(map[string]bool)
		for _, cert := range new.CertificateList.Elements() {
			newCerts[cert.String()] = true
		}

		// Find certificates to add
		for _, cert := range new.CertificateList.Elements() {
			if !oldCerts[cert.String()] {
				formatted_cert := strings.ReplaceAll(strings.Trim(cert.String(), "\""), `\n`, "\n")
				input.CertificatesToAdd = append(input.CertificatesToAdd, []byte(formatted_cert))
			}
		}

		// Find certificates to delete
		for _, cert := range old.CertificateList.Elements() {
			if !newCerts[cert.String()] {
				formatted_cert := strings.ReplaceAll(strings.Trim(cert.String(), "\""), `\n`, "\n")
				input.CertificatesToDelete = append(input.CertificatesToDelete, formatted_cert)
			}
		}

		_, err := conn.UpdateTrustStore(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating WorkSpacesWeb Trust Store (%s)", new.TrustStoreARN.ValueString()), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *trustStoreResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data trustStoreResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.DeleteTrustStoreInput{
		TrustStoreArn: aws.String(data.TrustStoreARN.ValueString()),
	}
	_, err := conn.DeleteTrustStore(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting WorkSpacesWeb Trust Store (%s)", data.TrustStoreARN.ValueString()), err.Error())
		return
	}
}

func (r *trustStoreResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("trust_store_arn"), request, response)
}

func findTrustStoreByARN(ctx context.Context, conn *workspacesweb.Client, arn string) (*awstypes.TrustStore, error) {
	input := workspacesweb.GetTrustStoreInput{
		TrustStoreArn: &arn,
	}
	output, err := conn.GetTrustStore(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TrustStore == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TrustStore, nil
}

type trustStoreResourceModel struct {
	AssociatedPortalARNs fwtypes.ListOfString `tfsdk:"associated_portal_arns"`
	CertificateList      fwtypes.ListOfString `tfsdk:"certificate_list"`
	Tags                 tftags.Map           `tfsdk:"tags"`
	TagsAll              tftags.Map           `tfsdk:"tags_all"`
	TrustStoreARN        types.String         `tfsdk:"trust_store_arn"`
}
