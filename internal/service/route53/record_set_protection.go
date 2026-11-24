// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"errors"
	"fmt"

	"github.com/YakDriver/smarterr"
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
	"github.com/hashicorp/terraform-provider-aws/internal/dns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_route53_record_set_protection", name="Record Set Protection")
// @IdentityAttribute("zone_id")
// @IdentityAttribute("record_name")
// @IdentityAttribute("record_type")
// @ImportIDHandler("recordSetProtectionImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(subdomainTfVar="zoneName;recordName")
// @Testing(generator=false)
// @Testing(importStateIdAttributes="zone_id;record_name;record_type", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(plannableImportAction="NoOp")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/route53/types;awstypes;awstypes.ResourceRecordSetProtection")
func newRecordSetProtectionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &recordSetProtectionResource{}, nil
}

type recordSetProtectionResource struct {
	framework.ResourceWithModel[recordSetProtectionResourceModel]
	framework.WithImportByIdentity
	// There is no API to update a protection, so every argument forces replacement.
	framework.WithNoUpdate
}

func (r *recordSetProtectionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrComment: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"record_name": schema.StringAttribute{
				CustomType: fwtypes.DNSNameStringType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
				},
			},
			"record_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RRType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *recordSetProtectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data recordSetProtectionResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53Client(ctx)

	zoneID, recordName, recordType := data.key()
	id := recordSetProtectionID(zoneID, recordName, recordType)
	input := route53.CreateResourceRecordSetProtectionsInput{
		HostedZoneId: aws.String(zoneID),
		Protections: []awstypes.ResourceRecordSetProtectionConfig{{
			Comment:    fwflex.StringFromFramework(ctx, data.Comment),
			RecordName: aws.String(recordName),
			RecordType: awstypes.RRType(recordType),
		}},
	}

	_, err := conn.CreateResourceRecordSetProtections(ctx, &input)

	// Adopting an existing protection would let a later destroy remove a
	// protection that someone else created, so surface it instead.
	if errs.IsA[*awstypes.ProtectionAlreadyExists](err) {
		err = fmt.Errorf("%w; to manage it with Terraform, import it with the ID %q", err, id)
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, protectionBatchError(err), smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *recordSetProtectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data recordSetProtectionResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53Client(ctx)

	zoneID, recordName, recordType := data.key()
	id := recordSetProtectionID(zoneID, recordName, recordType)
	output, err := findRecordSetProtectionByThreePartKey(ctx, conn, zoneID, recordName, recordType)

	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err), smerr.ID, id)
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	// record_name keeps the configured spelling; DNSNameStringType treats the
	// API's fully qualified form as equal.
	data.Comment = fwflex.StringToFramework(ctx, output.Comment)
	data.RecordType = fwtypes.StringEnumValue(output.RecordType)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *recordSetProtectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data recordSetProtectionResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53Client(ctx)

	zoneID, recordName, recordType := data.key()
	id := recordSetProtectionID(zoneID, recordName, recordType)
	input := route53.DeleteResourceRecordSetProtectionsInput{
		HostedZoneId: aws.String(zoneID),
		Protections: []awstypes.ResourceRecordSetProtectionKey{{
			RecordName: aws.String(recordName),
			RecordType: awstypes.RRType(recordType),
		}},
	}

	_, err := conn.DeleteResourceRecordSetProtections(ctx, &input)

	if errs.IsA[*awstypes.NoSuchHostedZone](err) || isProtectionNotFound(err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, protectionBatchError(err), smerr.ID, id)
		return
	}
}

type recordSetProtectionResourceModel struct {
	Comment    types.String                        `tfsdk:"comment"`
	RecordName fwtypes.DNSNameString               `tfsdk:"record_name"`
	RecordType fwtypes.StringEnum[awstypes.RRType] `tfsdk:"record_type"`
	ZoneID     types.String                        `tfsdk:"zone_id"`
}

func (m recordSetProtectionResourceModel) key() (string, string, string) {
	return m.ZoneID.ValueString(), m.RecordName.ValueString(), m.RecordType.ValueString()
}

func recordSetProtectionID(zoneID, recordName, recordType string) string {
	return zoneID + intflex.ResourceIdSeparator + recordName + intflex.ResourceIdSeparator + recordType
}

var _ inttypes.ImportIDParser = recordSetProtectionImportID{}

type recordSetProtectionImportID struct{}

func (recordSetProtectionImportID) Parse(id string) (string, map[string]any, error) {
	const partCount = 3
	parts, err := intflex.ExpandResourceId(id, partCount, false)
	if err != nil {
		return "", nil, fmt.Errorf("id %q should be in the format <zone-id>%[2]s<record-name>%[2]s<record-type>", id, intflex.ResourceIdSeparator)
	}

	return id, map[string]any{
		"zone_id":     parts[0],
		"record_name": parts[1],
		"record_type": parts[2],
	}, nil
}

// isProtectionNotFound reports whether err is a batch error in which every
// entry failed only because it has no protection.
func isProtectionNotFound(err error) bool {
	v, ok := errors.AsType[*awstypes.InvalidResourceRecordSetProtectionBatch](err)
	if !ok || len(v.Errors) == 0 {
		return false
	}

	for _, e := range v.Errors {
		if aws.ToString(e.Code) != protectionErrorCodeProtectionNotFound {
			return false
		}
	}

	return true
}

const (
	protectionErrorCodeProtectionNotFound = "ProtectionNotFound"
)

// protectionBatchError adds the per-entry reasons, which the batch error's own
// message omits, to an InvalidResourceRecordSetProtectionBatch error.
func protectionBatchError(err error) error {
	v, ok := errors.AsType[*awstypes.InvalidResourceRecordSetProtectionBatch](err)
	if !ok || len(v.Errors) == 0 {
		return err
	}

	entries := make([]error, 0, len(v.Errors))
	for _, e := range v.Errors {
		entries = append(entries, fmt.Errorf("%s %s: %s: %s", aws.ToString(e.RecordName), e.RecordType, aws.ToString(e.Code), aws.ToString(e.Message)))
	}

	return fmt.Errorf("%w: %w", err, errors.Join(entries...))
}

// findRecordSetProtectionByThreePartKey starts the listing at the record's
// name and type, in the canonical form that Route 53 returns, so the
// protection, if it exists, is the first result.
func findRecordSetProtectionByThreePartKey(ctx context.Context, conn *route53.Client, zoneID, recordName, recordType string) (*awstypes.ResourceRecordSetProtection, error) {
	input := route53.ListResourceRecordSetProtectionsInput{
		HostedZoneId:    aws.String(zoneID),
		MaxResults:      aws.Int32(1),
		StartRecordName: aws.String(fqdn(dns.Normalize(recordName))),
		StartRecordType: awstypes.RRType(recordType),
	}

	output, err := conn.ListResourceRecordSetProtections(ctx, &input)

	if errs.IsA[*awstypes.NoSuchHostedZone](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	for _, v := range output.ResourceRecordSetProtections {
		if dns.Normalize(aws.ToString(v.RecordName)) == dns.Normalize(recordName) && string(v.RecordType) == recordType {
			return &v, nil
		}
	}

	return nil, smarterr.NewError(&retry.NotFoundError{})
}

// findRecordSetProtectionsByZoneID returns every protection in a hosted zone.
func findRecordSetProtectionsByZoneID(ctx context.Context, conn *route53.Client, zoneID string) ([]awstypes.ResourceRecordSetProtection, error) {
	input := route53.ListResourceRecordSetProtectionsInput{
		HostedZoneId: aws.String(zoneID),
	}

	var output []awstypes.ResourceRecordSetProtection
	pages := route53.NewListResourceRecordSetProtectionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchHostedZone](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, page.ResourceRecordSetProtections...)
	}

	return output, nil
}
