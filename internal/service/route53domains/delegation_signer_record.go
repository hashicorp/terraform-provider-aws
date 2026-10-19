// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package route53domains

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53domains/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_route53domains_delegation_signer_record", name="Delegation Signer Record")
func newDelegationSignerRecordResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &delegationSignerRecordResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type delegationSignerRecordResource struct {
	framework.ResourceWithModel[delegationSignerRecordResourceModel]
	framework.WithNoUpdate
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *delegationSignerRecordResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"dnssec_key_id": framework.IDAttribute(),
			names.AttrDomainName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"signing_attributes": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[delegationSignerRecordSigningAttributesModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"algorithm": schema.Int64Attribute{
							Required: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
						"flags": schema.Int64Attribute{
							Required: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
						names.AttrPublicKey: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *delegationSignerRecordResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data delegationSignerRecordResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53DomainsClient(ctx)

	var input route53domains.AssociateDelegationSignerToDomainInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.AssociateDelegationSignerToDomain(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Route 53 Domains Delegation Signer Record", err.Error())

		return
	}

	if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError("waiting for Route 53 Domains Delegation Signer Record create", err.Error())

		return
	}

	signingAttributes, diags := data.SigningAttributes.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	domainName := fwflex.StringValueFromFramework(ctx, data.DomainName)
	flags := fwflex.Int32ValueFromFrameworkInt64(ctx, signingAttributes.Flags)
	algorithm := fwflex.Int32ValueFromFrameworkInt64(ctx, signingAttributes.Algorithm)
	publicKey := fwflex.StringValueFromFramework(ctx, signingAttributes.PublicKey)
	dnssecKey, err := findDNSSECKeyByThreePartKey(ctx, conn, domainName, flags, algorithm, publicKey)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Route 53 Domains Domain (%s) DNSSEC key", domainName), err.Error())

		return
	}

	// Store actual AWS ID for Delete operations, but compute digest for state ID stability.
	// This ensures registrar-agnostic state tracking (Gandi vs Amazon Registrar, etc).
	expectedDigest := calculateExpectedDigest(domainName, flags, algorithm, publicKey)

	// Set state ID using digest for stability.
	parts := []string{
		domainName,
		expectedDigest,
	}
	resourceID, err := flex.FlattenResourceId(parts, delegationSignerRecordResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewCreatingResourceIDErrorDiagnostic(err))
		return
	}

	// Store actual AWS ID for API operations (Delete, etc).
	data.DNSSECKeyID = fwflex.StringToFramework(ctx, dnssecKey.Id)
	data.ID = fwflex.StringValueToFramework(ctx, resourceID)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *delegationSignerRecordResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data delegationSignerRecordResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Extract domain and digest from stable resource ID
	idParts, err := flex.ExpandResourceId(data.ID.ValueString(), delegationSignerRecordResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))
		return
	}
	domainName := idParts[0]
	digest := idParts[1]

	conn := r.Meta().Route53DomainsClient(ctx)

	signingAttributes, diags := data.SigningAttributes.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Look up DNSSEC key by digest (deterministic across registrars).
	// Ensures stable state tracking regardless of registry (Gandi vs Amazon Registrar, etc).
	dnssecKey, err := findDNSSECKeyByDigest(ctx, conn, domainName, digest)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Route 53 Domains Domain (%s) DNSSEC key", data.DomainName.ValueString()), err.Error())

		return
	}

	// Update DNSSECKeyID with actual AWS ID for Delete operations.
	data.DNSSECKeyID = fwflex.StringToFramework(ctx, dnssecKey.Id)

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, dnssecKey, signingAttributes)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.SigningAttributes = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, signingAttributes)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *delegationSignerRecordResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data delegationSignerRecordResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53DomainsClient(ctx)

	input := route53domains.DisassociateDelegationSignerFromDomainInput{
		DomainName: fwflex.StringFromFramework(ctx, data.DomainName),
		Id:         fwflex.StringFromFramework(ctx, data.DNSSECKeyID),
	}
	output, err := conn.DisassociateDelegationSignerFromDomain(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Route 53 Domains Delegation Signer Record (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		if !errs.Contains(err, "The DNSSEC you specified is not found on domain") {
			response.Diagnostics.AddError("waiting for Route 53 Domains Delegation Signer Record delete", err.Error())

			return
		}
	}
}

type delegationSignerRecordResourceModel struct {
	DNSSECKeyID       types.String                                                                  `tfsdk:"dnssec_key_id"`
	DomainName        types.String                                                                  `tfsdk:"domain_name"`
	ID                types.String                                                                  `tfsdk:"id"`
	SigningAttributes fwtypes.ListNestedObjectValueOf[delegationSignerRecordSigningAttributesModel] `tfsdk:"signing_attributes"`
	Timeouts          timeouts.Value                                                                `tfsdk:"timeouts"`
}

type delegationSignerRecordSigningAttributesModel struct {
	Algorithm types.Int64  `tfsdk:"algorithm"`
	Flags     types.Int64  `tfsdk:"flags"`
	PublicKey types.String `tfsdk:"public_key"`
}

const (
	delegationSignerRecordResourceIDPartCount int   = 2
	dnssecProtocol                            uint8 = 3 // RFC 4034: DNSSEC protocol number
)

func findDNSSECKeyByTwoPartKey(ctx context.Context, conn *route53domains.Client, domainName, keyID string) (*awstypes.DnssecKey, error) {
	output, err := findDomainDetailByName(ctx, conn, domainName)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output.DnssecKeys, func(v awstypes.DnssecKey) bool {
		return aws.ToString(v.Id) == keyID
	}))
}

func findDNSSECKeyByDigest(ctx context.Context, conn *route53domains.Client, domainName, digest string) (*awstypes.DnssecKey, error) {
	output, err := findDomainDetailByName(ctx, conn, domainName)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output.DnssecKeys, func(v awstypes.DnssecKey) bool {
		return strings.EqualFold(aws.ToString(v.Digest), digest)
	}))
}

func findDNSSECKeyByThreePartKey(ctx context.Context, conn *route53domains.Client, domainName string, flags, algorithm int32, publicKey string) (*awstypes.DnssecKey, error) {
	output, err := findDomainDetailByName(ctx, conn, domainName)

	if err != nil {
		return nil, err
	}

	// For most TLDs, AWS returns full DNSSEC key info (Flags, PublicKey).
	// For .uk/.co.uk domains, the registry only returns Digest (partial data).
	// Try exact match first on Flags + PublicKey (standard TLDs).
	result, err := tfresource.AssertSingleValueResult(tfslices.Filter(output.DnssecKeys, func(v awstypes.DnssecKey) bool {
		return aws.ToInt32(v.Flags) == flags && aws.ToString(v.PublicKey) == publicKey
	}))

	// Fallback: If exact match fails but keys exist, compute RFC 4034 digest locally
	// and match against the AWS-returned Digest field (always present, registrar-agnostic).
	// This handles registry limitations and multi-key scenarios (e.g., during key rollover).
	if err != nil && len(output.DnssecKeys) > 0 {
		expectedDigest := calculateExpectedDigest(domainName, flags, algorithm, publicKey)

		fallbackResult, fallbackErr := tfresource.AssertSingleValueResult(tfslices.Filter(output.DnssecKeys, func(v awstypes.DnssecKey) bool {
			return strings.EqualFold(aws.ToString(v.Digest), expectedDigest)
		}))

		if fallbackErr == nil {
			return fallbackResult, nil
		}
	}

	return result, err
}

// CalculateExpectedDigest generates the RFC 4034 compliant SHA-256 digest
// from Terraform resource configuration. This is deterministic and works
// across all registrars since AWS always returns the Digest field.
func calculateExpectedDigest(domainName string, flags, algorithm int32, publicKey string) string {
	var buf bytes.Buffer

	// 1. Convert domain name to DNS wire-format labels (lowercase per RFC 4034)
	domain := strings.ToLower(strings.TrimSuffix(domainName, "."))
	if domain != "" {
		for label := range strings.SplitSeq(domain, ".") {
			buf.WriteByte(byte(len(label)))
			buf.WriteString(label)
		}
	}
	buf.WriteByte(0) // Root label terminator

	// 2. Append Flags (16-bit, Big Endian)
	flagsBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(flagsBytes, uint16(flags))
	buf.Write(flagsBytes)

	// 3. Append Protocol (8-bit, always 3 for DNSSEC)
	buf.WriteByte(dnssecProtocol)

	// 4. Append Algorithm (8-bit)
	buf.WriteByte(byte(algorithm))

	// 5. Append Public Key (Base64 decoded)
	pubKeyBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err == nil {
		buf.Write(pubKeyBytes)
	}

	// 6. Compute SHA-256 and return as uppercase hex
	hash := sha256.Sum256(buf.Bytes())
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}
