// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_service_network_vpc_association", name="Service Network VPC Association")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceServiceNetworkVPCAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	input := vpclattice.CreateServiceNetworkVpcAssociationInput{
		ClientToken:              aws.String(sdkid.UniqueId()),
		ServiceNetworkIdentifier: aws.String(d.Get("service_network_identifier").(string)),
		Tags:                     getTagsIn(ctx),
		VpcIdentifier:            aws.String(d.Get("vpc_identifier").(string)),
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok {
		input.SecurityGroupIds = flex.ExpandStringValueList(v.([]any))
	}

	output, err := conn.CreateServiceNetworkVpcAssociation(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating VPCLattice Service Network VPC Association: %s", err)
	}

	d.SetId(aws.ToString(output.Id))

	if _, err := waitServiceNetworkVPCAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPCLattice Service Network VPC Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceServiceNetworkVPCAssociationRead(ctx, d, meta)...)
}

func resourceServiceNetworkVPCAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	output, err := findServiceNetworkVPCAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice Service Network VPC Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPCLattice Service Network VPC Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set("created_by", output.CreatedBy)
	d.Set(names.AttrSecurityGroupIDs, output.SecurityGroupIds)
	d.Set("service_network_identifier", output.ServiceNetworkId)
	d.Set(names.AttrStatus, output.Status)
	d.Set("vpc_identifier", output.VpcId)

	return diags
}

func resourceServiceNetworkVPCAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := vpclattice.UpdateServiceNetworkVpcAssociationInput{
			ServiceNetworkVpcAssociationIdentifier: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrSecurityGroupIDs) {
			input.SecurityGroupIds = flex.ExpandStringValueList(d.Get(names.AttrSecurityGroupIDs).([]any))
		}

		_, err := conn.UpdateServiceNetworkVpcAssociation(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating VPCLattice Service Network VPC Association (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceServiceNetworkVPCAssociationRead(ctx, d, meta)...)
}

func resourceServiceNetworkVPCAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	log.Printf("[INFO] Deleting VPCLattice Service Network VPC Association: %s", d.Id())
	input := vpclattice.DeleteServiceNetworkVpcAssociationInput{
		ServiceNetworkVpcAssociationIdentifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteServiceNetworkVpcAssociation(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting VPCLattice Service Network VPC Association (%s): %s", d.Id(), err)
	}

	if _, err := waitServiceNetworkVPCAssociationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPCLattice Service Network VPC Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findServiceNetworkVPCAssociationByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetServiceNetworkVpcAssociationOutput, error) {
	input := vpclattice.GetServiceNetworkVpcAssociationInput{
		ServiceNetworkVpcAssociationIdentifier: aws.String(id),
	}

	return findServiceNetworkVPCAssociation(ctx, conn, &input)
}

func findServiceNetworkVPCAssociation(ctx context.Context, conn *vpclattice.Client, input *vpclattice.GetServiceNetworkVpcAssociationInput) (*vpclattice.GetServiceNetworkVpcAssociationOutput, error) {
	output, err := conn.GetServiceNetworkVpcAssociation(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

func statusServiceNetworkVPCAssociation(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findServiceNetworkVPCAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitServiceNetworkVPCAssociationCreated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetServiceNetworkVpcAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ServiceNetworkVpcAssociationStatusCreateInProgress),
		Target:                    enum.Slice(types.ServiceNetworkVpcAssociationStatusActive),
		Refresh:                   statusServiceNetworkVPCAssociation(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*vpclattice.GetServiceNetworkVpcAssociationOutput); ok {
		if output.Status == types.ServiceNetworkVpcAssociationStatusCreateFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.FailureCode), aws.ToString(output.FailureMessage)))
		}

		return output, err
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

	if output, ok := outputRaw.(*vpclattice.GetServiceNetworkVpcAssociationOutput); ok {
		if output.Status == types.ServiceNetworkVpcAssociationStatusDeleteFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.FailureCode), aws.ToString(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}
