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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_service_network_vpc_association", name="Service Network VPC Association")
// @Tags(identifierAttribute="arn")
func resourceServiceNetworkVPCAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceNetworkVPCAssociationCreate,
		ReadWithoutTimeout:   resourceServiceNetworkVPCAssociationRead,
		UpdateWithoutTimeout: resourceServiceNetworkVPCAssociationUpdate,
		DeleteWithoutTimeout: resourceServiceNetworkVPCAssociationDelete,

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
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeList,
				MaxItems: 5,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"vpc_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameServiceNetworkVPCAssociation = "ServiceNetworkVPCAssociation"
)

func resourceServiceNetworkVPCAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	in := &vpclattice.CreateServiceNetworkVpcAssociationInput{
		ClientToken:              aws.String(id.UniqueId()),
		ServiceNetworkIdentifier: aws.String(d.Get("service_network_identifier").(string)),
		VpcIdentifier:            aws.String(d.Get("vpc_identifier").(string)),
		Tags:                     getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok {
		in.SecurityGroupIds = flex.ExpandStringValueList(v.([]interface{}))
	}

	out, err := conn.CreateServiceNetworkVpcAssociation(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionCreating, ResNameServiceNetworkVPCAssociation, "", err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionCreating, ResNameServiceNetworkVPCAssociation, "", errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Id))

	if _, err := waitServiceNetworkVPCAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionWaitingForCreation, ResNameServiceNetworkVPCAssociation, d.Id(), err)
	}

	return append(diags, resourceServiceNetworkVPCAssociationRead(ctx, d, meta)...)
}

func resourceServiceNetworkVPCAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	out, err := findServiceNetworkVPCAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice Service Network VPC Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, ResNameServiceNetworkVPCAssociation, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set("created_by", out.CreatedBy)
	d.Set("vpc_identifier", out.VpcId)
	d.Set("service_network_identifier", out.ServiceNetworkId)
	d.Set(names.AttrSecurityGroupIDs, out.SecurityGroupIds)
	d.Set(names.AttrStatus, out.Status)

	return diags
}

func resourceServiceNetworkVPCAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)
	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		in := &vpclattice.UpdateServiceNetworkVpcAssociationInput{
			ServiceNetworkVpcAssociationIdentifier: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrSecurityGroupIDs) {
			in.SecurityGroupIds = flex.ExpandStringValueList(d.Get(names.AttrSecurityGroupIDs).([]interface{}))
		}

		log.Printf("[DEBUG] Updating VPCLattice ServiceNetwork VPC Association (%s): %#v", d.Id(), in)
		_, err := conn.UpdateServiceNetworkVpcAssociation(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionUpdating, ResNameServiceNetworkVPCAssociation, d.Id(), err)
		}
	}

	return append(diags, resourceServiceNetworkVPCAssociationRead(ctx, d, meta)...)
}

func resourceServiceNetworkVPCAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	log.Printf("[INFO] Deleting VPCLattice Service Network VPC Association %s", d.Id())

	_, err := conn.DeleteServiceNetworkVpcAssociation(ctx, &vpclattice.DeleteServiceNetworkVpcAssociationInput{
		ServiceNetworkVpcAssociationIdentifier: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionDeleting, ResNameServiceNetworkVPCAssociation, d.Id(), err)
	}

	if _, err := waitServiceNetworkVPCAssociationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionWaitingForDeletion, ResNameServiceNetworkVPCAssociation, d.Id(), err)
	}

	return diags
}

func findServiceNetworkVPCAssociationByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetServiceNetworkVpcAssociationOutput, error) {
	in := &vpclattice.GetServiceNetworkVpcAssociationInput{
		ServiceNetworkVpcAssociationIdentifier: aws.String(id),
	}
	out, err := conn.GetServiceNetworkVpcAssociation(ctx, in)

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

func waitServiceNetworkVPCAssociationCreated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetServiceNetworkVpcAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ServiceNetworkVpcAssociationStatusCreateInProgress),
		Target:                    enum.Slice(types.ServiceNetworkVpcAssociationStatusActive),
		Refresh:                   statusServiceNetworkVPCAssociation(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetServiceNetworkVpcAssociationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitServiceNetworkVPCAssociationDeleted(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetServiceNetworkVpcAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ServiceNetworkVpcAssociationStatusDeleteInProgress, types.ServiceNetworkVpcAssociationStatusActive),
		Target:  []string{},
		Refresh: statusServiceNetworkVPCAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetServiceNetworkVpcAssociationOutput); ok {
		return out, err
	}

	return nil, err
}

func statusServiceNetworkVPCAssociation(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findServiceNetworkVPCAssociationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}
