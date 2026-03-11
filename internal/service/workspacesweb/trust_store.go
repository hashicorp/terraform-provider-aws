// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package workspacesweb

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspacesweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
	framework.ResourceWithModel[trustStoreResourceModel]
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
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"trust_store_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrCertificate: schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[certificateModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(0),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"body": schema.StringAttribute{
							Required: true,
						},
						names.AttrIssuer: schema.StringAttribute{
							Computed: true,
						},
						"not_valid_after": schema.StringAttribute{
							Computed: true,
						},
						"not_valid_before": schema.StringAttribute{
							Computed: true,
						},
						"subject": schema.StringAttribute{
							Computed: true,
						},
						"thumbprint": schema.StringAttribute{
							Computed: true,
						},
					},
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

	input := workspacesweb.CreateTrustStoreInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		Tags:        getTagsIn(ctx),
	}

	// Convert string certificates to byte slices
	for _, certificate := range data.Certificates.Elements() {
		var cert certificateModel
		response.Diagnostics.Append(tfsdk.ValueAs(ctx, certificate, &cert)...)
		if response.Diagnostics.HasError() {
			return
		}

		formattedCert := strings.ReplaceAll(strings.Trim(cert.Body.ValueString(), "\""), `\n`, "\n")
		input.CertificateList = append(input.CertificateList, []byte(formattedCert))
	}

	output, err := conn.CreateTrustStore(ctx, &input)

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

	// Populate certificate details
	certificates, err := listTrustStoreCertificates(ctx, conn, data.TrustStoreARN.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("listing WorkSpacesWeb Trust Store Certificates (%s)", data.TrustStoreARN.ValueString()), err.Error())
		return
	}

	var diags diag.Diagnostics
	data.Certificates, diags = fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, certificates)
	response.Diagnostics.Append(diags...)
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
	if retry.NotFound(err) {
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

	// Populate certificate details by merging existing state with computed values
	certificates, err := listTrustStoreCertificates(ctx, conn, data.TrustStoreARN.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("listing WorkSpacesWeb Trust Store Certificates (%s)", data.TrustStoreARN.ValueString()), err.Error())
		return
	}

	var diags diag.Diagnostics
	data.Certificates, diags = fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, certificates)
	response.Diagnostics.Append(diags...)
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

	if !new.Certificates.Equal(old.Certificates) {
		input := workspacesweb.UpdateTrustStoreInput{
			ClientToken:   aws.String(sdkid.UniqueId()),
			TrustStoreArn: new.TrustStoreARN.ValueStringPointer(),
		}

		// Handle certificate additions and deletions
		oldCerts := make(map[string]string) // cert content -> thumbprint
		for _, certificate := range old.Certificates.Elements() {
			var cert certificateModel
			response.Diagnostics.Append(tfsdk.ValueAs(ctx, certificate, &cert)...)
			if response.Diagnostics.HasError() {
				return
			}

			oldCerts[base64.StdEncoding.EncodeToString([]byte(cert.Body.ValueString()))] = cert.Thumbprint.ValueString()
		}

		newCertContents := make(map[string]bool)
		for _, certificate := range new.Certificates.Elements() {
			var cert certificateModel
			response.Diagnostics.Append(tfsdk.ValueAs(ctx, certificate, &cert)...)
			if response.Diagnostics.HasError() {
				return
			}

			formattedCert := strings.ReplaceAll(strings.Trim(cert.Body.ValueString(), "\""), `\n`, "\n")
			newCertContents[base64.StdEncoding.EncodeToString([]byte(formattedCert))] = true
		}

		// Find certificates to add
		for _, certificate := range new.Certificates.Elements() {
			var cert certificateModel
			response.Diagnostics.Append(tfsdk.ValueAs(ctx, certificate, &cert)...)
			if response.Diagnostics.HasError() {
				return
			}

			formattedCert := strings.ReplaceAll(strings.Trim(cert.Body.String(), "\""), `\n`, "\n")
			certEncoded := base64.StdEncoding.EncodeToString([]byte(formattedCert))
			if _, exists := oldCerts[certEncoded]; !exists {
				input.CertificatesToAdd = append(input.CertificatesToAdd, []byte(formattedCert))
			}
		}

		// Find certificates to delete (by thumbprint)
		for certEncoded, thumbprint := range oldCerts {
			if !newCertContents[certEncoded] {
				input.CertificatesToDelete = append(input.CertificatesToDelete, thumbprint)
			}
		}

		_, err := conn.UpdateTrustStore(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating WorkSpacesWeb Trust Store (%s)", new.TrustStoreARN.ValueString()), err.Error())
			return
		}
	}

	// Read the updated state to get computed values
	updatedTrustStore, err := findTrustStoreByARN(ctx, conn, new.TrustStoreARN.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb Trust Store (%s) after update", new.TrustStoreARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, updatedTrustStore, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Populate certificate details by merging planned data with computed values
	certificates, err := listTrustStoreCertificates(ctx, conn, new.TrustStoreARN.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("listing WorkSpacesWeb Trust Store Certificates (%s) after update", new.TrustStoreARN.ValueString()), err.Error())
		return
	}

	var diags diag.Diagnostics
	new.Certificates, diags = fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, certificates)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
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
		TrustStoreArn: data.TrustStoreARN.ValueStringPointer(),
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
		TrustStoreArn: aws.String(arn),
	}
	output, err := conn.GetTrustStore(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TrustStore == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.TrustStore, nil
}

func listTrustStoreCertificates(ctx context.Context, conn *workspacesweb.Client, arn string) ([]certificateModel, error) {
	input := workspacesweb.ListTrustStoreCertificatesInput{
		TrustStoreArn: aws.String(arn),
	}

	var certificates []certificateModel
	pages := workspacesweb.NewListTrustStoreCertificatesPaginator(conn, &input)
	for pages.HasMorePages() {
		output, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, certSummary := range output.CertificateList {
			// Get detailed certificate information
			input := workspacesweb.GetTrustStoreCertificateInput{
				Thumbprint:    certSummary.Thumbprint,
				TrustStoreArn: aws.String(arn),
			}

			output, err := conn.GetTrustStoreCertificate(ctx, &input)

			if err != nil {
				return nil, err
			}

			if output.Certificate != nil {
				cert := certificateModel{
					Body:           types.StringValue(string(output.Certificate.Body)),
					Issuer:         types.StringPointerValue(output.Certificate.Issuer),
					NotValidAfter:  types.StringValue(aws.ToTime(output.Certificate.NotValidAfter).Format(time.RFC3339)),
					NotValidBefore: types.StringValue(aws.ToTime(output.Certificate.NotValidBefore).Format(time.RFC3339)),
					Subject:        types.StringPointerValue(output.Certificate.Subject),
					Thumbprint:     types.StringPointerValue(output.Certificate.Thumbprint),
				}
				certificates = append(certificates, cert)
			}
		}
	}

	return certificates, nil
}

type trustStoreResourceModel struct {
	framework.WithRegionModel
	AssociatedPortalARNs fwtypes.ListOfString                             `tfsdk:"associated_portal_arns"`
	Certificates         fwtypes.SetNestedObjectValueOf[certificateModel] `tfsdk:"certificate"`
	Tags                 tftags.Map                                       `tfsdk:"tags"`
	TagsAll              tftags.Map                                       `tfsdk:"tags_all"`
	TrustStoreARN        types.String                                     `tfsdk:"trust_store_arn"`
}

type certificateModel struct {
	Body           types.String `tfsdk:"body"`
	Issuer         types.String `tfsdk:"issuer"`
	NotValidAfter  types.String `tfsdk:"not_valid_after"`
	NotValidBefore types.String `tfsdk:"not_valid_before"`
	Subject        types.String `tfsdk:"subject"`
	Thumbprint     types.String `tfsdk:"thumbprint"`
}
