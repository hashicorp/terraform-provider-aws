// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
			"dns_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"private_dns_preference": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.PrivateDnsPreference](),
						},
						"private_dns_specified_domains": {
							Type:     schema.TypeSet,
							MaxItems: 10,
							Optional: true,
							Computed: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
						},
					},
				},
			},
			"private_dns_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
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

	if v, ok := d.GetOk("private_dns_enabled"); ok {
		input.PrivateDnsEnabled = aws.Bool(v.(bool))
	}

	if aws.ToBool(input.PrivateDnsEnabled) {
		if v, ok := d.GetOk("dns_options"); ok && len(v.([]any)) > 0 {
			input.DnsOptions = expandDNSOptions(v.([]any)[0].(map[string]any))
		}
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

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] VPCLattice Service Network VPC Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPCLattice Service Network VPC Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set("created_by", output.CreatedBy)
	d.Set("private_dns_enabled", output.PrivateDnsEnabled)
	if aws.ToBool(output.PrivateDnsEnabled) {
		if err := d.Set("dns_options", flattenDNSOptions(output.DnsOptions)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting dns_options: %s", err)
		}
	}
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

		log.Printf("[INFO] Updating VPCLattice Service Network VPC Association: %s", d.Id())

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
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusServiceNetworkVPCAssociation(conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findServiceNetworkVPCAssociationByID(ctx, conn, id)

		if retry.NotFound(err) {
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
		Refresh:                   statusServiceNetworkVPCAssociation(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*vpclattice.GetServiceNetworkVpcAssociationOutput); ok {
		if output.Status == types.ServiceNetworkVpcAssociationStatusCreateFailed {
			retry.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.FailureCode), aws.ToString(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitServiceNetworkVPCAssociationDeleted(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetServiceNetworkVpcAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ServiceNetworkVpcAssociationStatusDeleteInProgress, types.ServiceNetworkVpcAssociationStatusActive),
		Target:  []string{},
		Refresh: statusServiceNetworkVPCAssociation(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*vpclattice.GetServiceNetworkVpcAssociationOutput); ok {
		if output.Status == types.ServiceNetworkVpcAssociationStatusDeleteFailed {
			retry.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.FailureCode), aws.ToString(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func expandDNSOptions(m map[string]any) *types.DnsOptions {
	if len(m) == 0 {
		return nil
	}

	dnsOptions := &types.DnsOptions{}

	if v, ok := m["private_dns_preference"].(string); ok && v != "" {
		dnsOptions.PrivateDnsPreference = types.PrivateDnsPreference(v)
	}

	if dnsOptions.PrivateDnsPreference == types.PrivateDnsPreferenceVerifiedDomainsAndSpecifiedDomains || dnsOptions.PrivateDnsPreference == types.PrivateDnsPreferenceSpecifiedDomainsOnly {
		if v, ok := m["private_dns_specified_domains"]; ok && v != nil && len(v.(*schema.Set).List()) > 0 {
			dnsOptions.PrivateDnsSpecifiedDomains = flex.ExpandStringValueSet(v.(*schema.Set))
		}
	}
	return dnsOptions
}

func flattenDNSOptions(dnsOptions *types.DnsOptions) []map[string]any {
	if dnsOptions == nil {
		return nil
	}

	m := map[string]any{}

	if dnsOptions.PrivateDnsPreference != "" {
		m["private_dns_preference"] = string(dnsOptions.PrivateDnsPreference)
	}

	if len(dnsOptions.PrivateDnsSpecifiedDomains) > 0 {
		m["private_dns_specified_domains"] = flex.FlattenStringValueList(dnsOptions.PrivateDnsSpecifiedDomains)
	}

	return []map[string]any{m}
}
