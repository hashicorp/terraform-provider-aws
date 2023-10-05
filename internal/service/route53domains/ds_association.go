// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53domains

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53domains/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Delegation Signer Association")
func newResourceDelegationSignerAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDelegationSignerAssocation{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameDelegationSignerAssociation = "DelegationSignerAssociation"
)

type resourceDelegationSignerAssocation struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDelegationSignerAssocation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_route53domains_ds_association"
}

func (r *resourceDelegationSignerAssocation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				Required: true,
			},
			"signing_algorithm_type": schema.Int64Attribute{
				Required: true,
			},
			"flag": schema.Int64Attribute{
				Required: true,
			},
			"public_key": schema.StringAttribute{
				Required: true,
			},
			"dnssec_key_id": framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceDelegationSignerAssocation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().Route53DomainsClient(ctx)

	var plan resourceDelegationSignerAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	flag := int32(plan.Flag.ValueInt64())
	publicKey := plan.PublicKey.ValueString()

	in := &route53domains.AssociateDelegationSignerToDomainInput{
		DomainName: plan.DomainName.ValueStringPointer(),
		SigningAttributes: &awstypes.DnssecSigningAttributes{
			Algorithm: aws.Int32(int32(plan.SigningAlgorithmType.ValueInt64())),
			Flags:     aws.Int32(flag),
			PublicKey: aws.String(publicKey),
		},
	}

	out, err := conn.AssociateDelegationSignerToDomain(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Domains, create.ErrActionCreating, ResNameDelegationSignerAssociation, plan.DomainName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.OperationId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Domains, create.ErrActionCreating, ResNameDelegationSignerAssociation, plan.DomainName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)

	if _, err := waitOperationSucceeded(ctx, conn, *out.OperationId, createTimeout); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Domains, create.ErrActionWaitingForCreation, ResNameDelegationSignerAssociation, plan.DomainName.String(), err),
			err.Error(),
		)
		return
	}

	domainDetail, err := findDomainDetailByName(ctx, conn, plan.DomainName.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Domains, create.ErrActionCheckingExistence, ResNameDelegationSignerAssociation, plan.DomainName.String(), err),
			err.Error(),
		)
		return
	}

	dnssecKey := getDnssecKeyWithFlagAndPublicKey(domainDetail.DnssecKeys, flag, publicKey)
	if dnssecKey == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Domains, create.ErrActionCheckingExistence, ResNameDelegationSignerAssociation, plan.DomainName.String(), nil),
			errors.New("DNSSEC key not found").Error(),
		)
		return
	}

	plan.DNSSECKeyID = flex.StringToFramework(ctx, dnssecKey.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDelegationSignerAssocation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().Route53DomainsClient(ctx)

	var state resourceDelegationSignerAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainDetail, err := findDomainDetailByName(ctx, conn, state.DomainName.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Domains, create.ErrActionCheckingExistence, ResNameDelegationSignerAssociation, state.DomainName.String(), err),
			err.Error(),
		)
		return
	}

	dnssecKey := GetDnssecKeyWithId(domainDetail.DnssecKeys, state.DNSSECKeyID.ValueString())
	if dnssecKey == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.SigningAlgorithmType = flex.Int64ToFramework(ctx, aws.Int64(int64(aws.ToInt32(dnssecKey.Algorithm))))
	state.Flag = flex.Int64ToFramework(ctx, aws.Int64(int64(aws.ToInt32(dnssecKey.Flags))))
	state.PublicKey = flex.StringToFramework(ctx, dnssecKey.PublicKey)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// There is no update API, so this method is a no-op
func (r *resourceDelegationSignerAssocation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceDelegationSignerAssocation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().Route53DomainsClient(ctx)

	var state resourceDelegationSignerAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainDetail, err := findDomainDetailByName(ctx, conn, state.DomainName.ValueString())

	if tfresource.NotFound(err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Domains, create.ErrActionCheckingExistence, ResNameDelegationSignerAssociation, state.DomainName.String(), err),
			err.Error(),
		)
		return
	}

	dnssecKey := GetDnssecKeyWithId(domainDetail.DnssecKeys, state.DNSSECKeyID.ValueString())
	if dnssecKey == nil {
		return
	}

	in := &route53domains.DisassociateDelegationSignerFromDomainInput{
		DomainName: aws.String(state.DomainName.ValueString()),
		Id:         aws.String(state.DNSSECKeyID.ValueString()),
	}

	out, err := conn.DisassociateDelegationSignerFromDomain(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Domains, create.ErrActionDeleting, ResNameDelegationSignerAssociation, state.DomainName.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitOperationSucceeded(ctx, conn, *out.OperationId, deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Route53Domains, create.ErrActionWaitingForDeletion, ResNameDelegationSignerAssociation, state.DomainName.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceDelegationSignerAssocation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "DomainName:DNSSECKeyID"`, req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dnssec_key_id"), parts[1])...)
}

func getDnssecKeyWithFlagAndPublicKey(dnssecKeys []awstypes.DnssecKey, flag int32, publicKey string) *awstypes.DnssecKey {
	for _, dnssecKey := range dnssecKeys {
		if *dnssecKey.Flags == flag && *dnssecKey.PublicKey == publicKey {
			return &dnssecKey
		}
	}
	return nil
}

func GetDnssecKeyWithId(dnssecKeys []awstypes.DnssecKey, dnssec_key_id string) *awstypes.DnssecKey {
	for _, dnssecKey := range dnssecKeys {
		if *dnssecKey.Id == dnssec_key_id {
			return &dnssecKey
		}
	}
	return nil
}

type resourceDelegationSignerAssociationData struct {
	DomainName           types.String   `tfsdk:"domain_name"`
	SigningAlgorithmType types.Int64    `tfsdk:"signing_algorithm_type"`
	Flag                 types.Int64    `tfsdk:"flag"`
	PublicKey            types.String   `tfsdk:"public_key"`
	DNSSECKeyID          types.String   `tfsdk:"dnssec_key_id"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
}
