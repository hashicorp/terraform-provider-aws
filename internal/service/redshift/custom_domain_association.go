// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package redshift

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var errCustomDomainAssociationNotUpdated = errors.New("redshift custom domain association not updated yet")

// @FrameworkResource("aws_redshift_custom_domain_association", name="Custom Domain Association")
func newCustomDomainAssociationResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &customDomainAssociationResource{}, nil
}

type customDomainAssociationResource struct {
	framework.ResourceWithModel[customDomainAssociationResourceModel]
	framework.WithImportByID
}

func (r *customDomainAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrClusterIdentifier: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"custom_domain_certificate_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"custom_domain_certificate_expiry_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"custom_domain_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 253),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
		},
	}
}

func (r *customDomainAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data customDomainAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	_, err := conn.CreateCustomDomainAssociation(ctx, &redshift.CreateCustomDomainAssociationInput{
		ClusterIdentifier:          data.ClusterIdentifier.ValueStringPointer(),
		CustomDomainCertificateArn: data.CustomDomainCertificateARN.ValueStringPointer(),
		CustomDomainName:           data.CustomDomainName.ValueStringPointer(),
	})
	if err != nil {
		response.Diagnostics.AddError("creating Redshift Custom Domain Association", err.Error())
		return
	}

	id, err := data.setID()
	if err != nil {
		response.Diagnostics.AddError("creating Redshift Custom Domain Association", err.Error())
		return
	}
	data.ID = types.StringValue(id)

	output, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout, func(ctx context.Context) (*customDomainAssociation, error) {
		return findCustomDomainAssociationByTwoPartKey(ctx, conn, data.ClusterIdentifier.ValueString(), data.CustomDomainName.ValueString())
	})
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Redshift Custom Domain Association (%s) create", data.ID.ValueString()), err.Error())
		return
	}

	if err := data.setFromAPIObject(output); err != nil {
		response.Diagnostics.AddError("creating Redshift Custom Domain Association", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *customDomainAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data customDomainAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	output, err := findCustomDomainAssociationByTwoPartKey(ctx, conn, data.ClusterIdentifier.ValueString(), data.CustomDomainName.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Redshift Custom Domain Association (%s)", data.ID.ValueString()), err.Error())
		return
	}

	if err := data.setFromAPIObject(output); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Redshift Custom Domain Association (%s)", data.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *customDomainAssociationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new customDomainAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	_, err := conn.ModifyCustomDomainAssociation(ctx, &redshift.ModifyCustomDomainAssociationInput{
		ClusterIdentifier:          new.ClusterIdentifier.ValueStringPointer(),
		CustomDomainCertificateArn: new.CustomDomainCertificateARN.ValueStringPointer(),
		CustomDomainName:           new.CustomDomainName.ValueStringPointer(),
	})
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Redshift Custom Domain Association (%s)", old.ID.ValueString()), err.Error())
		return
	}

	output, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func(ctx context.Context) (*customDomainAssociation, error) {
			output, err := findCustomDomainAssociationByTwoPartKey(ctx, conn, new.ClusterIdentifier.ValueString(), new.CustomDomainName.ValueString())
			if err != nil {
				return nil, err
			}

			if aws.ToString(output.CustomDomainCertificateArn) != new.CustomDomainCertificateARN.ValueString() {
				return nil, errCustomDomainAssociationNotUpdated
			}

			return output, nil
		},
		func(err error) (bool, error) {
			if retry.NotFound(err) || errors.Is(err, errCustomDomainAssociationNotUpdated) {
				return true, err
			}

			return false, err
		},
	)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Redshift Custom Domain Association (%s) update", old.ID.ValueString()), err.Error())
		return
	}

	if err := new.setFromAPIObject(output); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Redshift Custom Domain Association (%s)", old.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *customDomainAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data customDomainAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	_, err := conn.DeleteCustomDomainAssociation(ctx, &redshift.DeleteCustomDomainAssociationInput{
		ClusterIdentifier: data.ClusterIdentifier.ValueStringPointer(),
		CustomDomainName:  data.CustomDomainName.ValueStringPointer(),
	})
	if errs.IsA[*awstypes.ClusterNotFoundFault](err) || errs.IsA[*awstypes.CustomDomainAssociationNotFoundFault](err) {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Redshift Custom Domain Association (%s)", data.ID.ValueString()), err.Error())
		return
	}

	_, err = tfresource.RetryUntilNotFound(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
		return findCustomDomainAssociationByTwoPartKey(ctx, conn, data.ClusterIdentifier.ValueString(), data.CustomDomainName.ValueString())
	})
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Redshift Custom Domain Association (%s) delete", data.ID.ValueString()), err.Error())
		return
	}
}

func findCustomDomainAssociationByTwoPartKey(ctx context.Context, conn *redshift.Client, clusterIdentifier, customDomainName string) (*customDomainAssociation, error) {
	input := &redshift.DescribeCustomDomainAssociationsInput{
		CustomDomainName: aws.String(customDomainName),
	}

	for {
		output, err := conn.DescribeCustomDomainAssociations(ctx, input)
		if errs.IsA[*awstypes.CustomDomainAssociationNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}
		if err != nil {
			return nil, err
		}

		for _, association := range output.Associations {
			for _, certificateAssociation := range association.CertificateAssociations {
				if aws.ToString(certificateAssociation.ClusterIdentifier) == clusterIdentifier &&
					aws.ToString(certificateAssociation.CustomDomainName) == customDomainName {
					return &customDomainAssociation{
						ClusterIdentifier:                 certificateAssociation.ClusterIdentifier,
						CustomDomainCertificateArn:        association.CustomDomainCertificateArn,
						CustomDomainCertificateExpiryDate: association.CustomDomainCertificateExpiryDate,
						CustomDomainName:                  certificateAssociation.CustomDomainName,
					}, nil
				}
			}
		}

		if aws.ToString(output.Marker) == "" {
			break
		}
		input.Marker = output.Marker
	}

	return nil, tfresource.NewEmptyResultError()
}

type customDomainAssociation struct {
	ClusterIdentifier                 *string
	CustomDomainCertificateArn        *string
	CustomDomainCertificateExpiryDate *time.Time
	CustomDomainName                  *string
}

type customDomainAssociationResourceModel struct {
	framework.WithRegionModel
	ClusterIdentifier                 types.String      `tfsdk:"cluster_identifier"`
	CustomDomainCertificateARN        fwtypes.ARN       `tfsdk:"custom_domain_certificate_arn"`
	CustomDomainCertificateExpiryTime timetypes.RFC3339 `tfsdk:"custom_domain_certificate_expiry_time"`
	CustomDomainName                  types.String      `tfsdk:"custom_domain_name"`
	ID                                types.String      `tfsdk:"id"`
}

const (
	customDomainAssociationResourceIDPartCount = 2
)

func (data *customDomainAssociationResourceModel) InitFromID() error {
	parts, err := flex.ExpandResourceId(data.ID.ValueString(), customDomainAssociationResourceIDPartCount, false)
	if err != nil {
		return err
	}

	data.ClusterIdentifier = types.StringValue(parts[0])
	data.CustomDomainName = types.StringValue(parts[1])

	return nil
}

func (data *customDomainAssociationResourceModel) setFromAPIObject(apiObject *customDomainAssociation) error {
	data.ClusterIdentifier = types.StringPointerValue(apiObject.ClusterIdentifier)
	data.CustomDomainCertificateARN = fwtypes.ARNValue(aws.ToString(apiObject.CustomDomainCertificateArn))
	data.CustomDomainCertificateExpiryTime = timetypes.NewRFC3339TimePointerValue(apiObject.CustomDomainCertificateExpiryDate)
	data.CustomDomainName = types.StringPointerValue(apiObject.CustomDomainName)

	id, err := data.setID()
	if err != nil {
		return err
	}

	data.ID = types.StringValue(id)

	return nil
}

func (data *customDomainAssociationResourceModel) setID() (string, error) {
	parts := []string{
		data.ClusterIdentifier.ValueString(),
		data.CustomDomainName.ValueString(),
	}

	return flex.FlattenResourceId(parts, customDomainAssociationResourceIDPartCount, false)
}
