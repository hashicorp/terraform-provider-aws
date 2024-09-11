// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_service_network_service_association", name="Service Network Service Association")
// @Tags(identifierAttribute="arn")
func resourceServiceNetworkServiceAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceNetworkServiceAssociationCreate,
		ReadWithoutTimeout:   resourceServiceNetworkServiceAssociationRead,
		UpdateWithoutTimeout: resourceServiceNetworkServiceAssociationUpdate,
		DeleteWithoutTimeout: resourceServiceNetworkServiceAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_entry": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDomainName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrHostedZoneID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"service_identifier": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressEquivalentIDOrARN,
			},
			"service_network_identifier": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressEquivalentIDOrARN,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameServiceNetworkAssociation = "ServiceNetworkAssociation"
)

func resourceServiceNetworkServiceAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	in := &vpclattice.CreateServiceNetworkServiceAssociationInput{
		ClientToken:              aws.String(id.UniqueId()),
		ServiceIdentifier:        aws.String(d.Get("service_identifier").(string)),
		ServiceNetworkIdentifier: aws.String(d.Get("service_network_identifier").(string)),
		Tags:                     getTagsIn(ctx),
	}

	out, err := conn.CreateServiceNetworkServiceAssociation(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionCreating, ResNameServiceNetworkAssociation, "", err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionCreating, ResNameServiceNetworkAssociation, "", errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Id))

	if _, err := waitServiceNetworkServiceAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionWaitingForCreation, ResNameServiceNetworkAssociation, d.Id(), err)
	}

	return append(diags, resourceServiceNetworkServiceAssociationRead(ctx, d, meta)...)
}

func resourceServiceNetworkServiceAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	out, err := findServiceNetworkServiceAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice Service Network Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, ResNameServiceNetworkAssociation, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set("created_by", out.CreatedBy)
	d.Set("custom_domain_name", out.CustomDomainName)
	if out.DnsEntry != nil {
		if err := d.Set("dns_entry", []interface{}{flattenDNSEntry(out.DnsEntry)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting dns_entry: %s", err)
		}
	} else {
		d.Set("dns_entry", nil)
	}
	d.Set("service_identifier", out.ServiceId)
	d.Set("service_network_identifier", out.ServiceNetworkId)
	d.Set(names.AttrStatus, out.Status)

	return diags
}

func resourceServiceNetworkServiceAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceServiceNetworkServiceAssociationRead(ctx, d, meta)
}

func resourceServiceNetworkServiceAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	log.Printf("[INFO] Deleting VPCLattice Service Network Association %s", d.Id())

	_, err := conn.DeleteServiceNetworkServiceAssociation(ctx, &vpclattice.DeleteServiceNetworkServiceAssociationInput{
		ServiceNetworkServiceAssociationIdentifier: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionDeleting, ResNameServiceNetworkAssociation, d.Id(), err)
	}

	if _, err := waitServiceNetworkServiceAssociationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionWaitingForDeletion, ResNameServiceNetworkAssociation, d.Id(), err)
	}

	return diags
}

func findServiceNetworkServiceAssociationByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetServiceNetworkServiceAssociationOutput, error) {
	in := &vpclattice.GetServiceNetworkServiceAssociationInput{
		ServiceNetworkServiceAssociationIdentifier: aws.String(id),
	}
	out, err := conn.GetServiceNetworkServiceAssociation(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func waitServiceNetworkServiceAssociationCreated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetServiceNetworkServiceAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ServiceNetworkVpcAssociationStatusCreateInProgress),
		Target:                    enum.Slice(types.ServiceNetworkVpcAssociationStatusActive),
		Refresh:                   statusServiceNetworkServiceAssociation(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetServiceNetworkServiceAssociationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitServiceNetworkServiceAssociationDeleted(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetServiceNetworkServiceAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ServiceNetworkVpcAssociationStatusDeleteInProgress, types.ServiceNetworkVpcAssociationStatusActive),
		Target:  []string{},
		Refresh: statusServiceNetworkServiceAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetServiceNetworkServiceAssociationOutput); ok {
		return out, err
	}

	return nil, err
}

func statusServiceNetworkServiceAssociation(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findServiceNetworkServiceAssociationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}
