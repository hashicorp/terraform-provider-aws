// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_endpoint", name="Endpoint")
// @Tags(identifierAttribute="arn")
func ResourceEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEndpointCreate,
		ReadWithoutTimeout:   resourceEndpointRead,
		UpdateWithoutTimeout: resourceEndpointUpdate,
		DeleteWithoutTimeout: resourceEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"direction": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(route53resolver.ResolverEndpointDirection_Values(), false),
			},
			"host_vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrIPAddress: {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 2,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"ip_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrSubnetID: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: endpointHashIPAddress,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validResolverName,
			},
			"resolver_endpoint_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(route53resolver.ResolverEndpointType_Values(), false),
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 64,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"protocols": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MinItems: 1,
				MaxItems: 2,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(route53resolver.Protocol_Values(), false),
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	input := &route53resolver.CreateResolverEndpointInput{
		CreatorRequestId: aws.String(id.PrefixedUniqueId("tf-r53-resolver-endpoint-")),
		Direction:        aws.String(d.Get("direction").(string)),
		IpAddresses:      expandEndpointIPAddresses(d.Get(names.AttrIPAddress).(*schema.Set)),
		SecurityGroupIds: flex.ExpandStringSet(d.Get(names.AttrSecurityGroupIDs).(*schema.Set)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("protocols"); ok && v.(*schema.Set).Len() > 0 {
		input.Protocols = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("resolver_endpoint_type"); ok {
		input.ResolverEndpointType = aws.String(v.(string))
	}

	output, err := conn.CreateResolverEndpointWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Endpoint: %s", err)
	}

	d.SetId(aws.StringValue(output.ResolverEndpoint.Id))

	if _, err := waitEndpointCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Endpoint (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	ep, err := FindResolverEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Endpoint (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(ep.Arn)
	d.Set(names.AttrARN, arn)
	d.Set("direction", ep.Direction)
	d.Set("host_vpc_id", ep.HostVPCId)
	d.Set(names.AttrName, ep.Name)
	d.Set("protocols", aws.StringValueSlice(ep.Protocols))
	d.Set("resolver_endpoint_type", ep.ResolverEndpointType)
	d.Set(names.AttrSecurityGroupIDs, aws.StringValueSlice(ep.SecurityGroupIds))

	ipAddresses, err := findResolverEndpointIPAddressesByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Route53 Resolver Endpoint (%s) IP addresses: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrIPAddress, schema.NewSet(endpointHashIPAddress, flattenEndpointIPAddresses(ipAddresses))); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ip_address: %s", err)
	}

	return diags
}

func resourceEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	if d.HasChanges(names.AttrName, "protocols", "resolver_endpoint_type") {
		input := &route53resolver.UpdateResolverEndpointInput{
			ResolverEndpointId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange("protocols") {
			input.Protocols = flex.ExpandStringSet(d.Get("protocols").(*schema.Set))
		}

		if d.HasChange("resolver_endpoint_type") {
			input.ResolverEndpointType = aws.String(d.Get("resolver_endpoint_type").(string))
		}

		_, err := conn.UpdateResolverEndpointWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Resolver Endpoint (%s): %s", d.Id(), err)
		}

		if _, err := waitEndpointUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Endpoint (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrIPAddress) {
		oraw, nraw := d.GetChange(names.AttrIPAddress)
		o := oraw.(*schema.Set)
		n := nraw.(*schema.Set)
		del := o.Difference(n).List()
		add := n.Difference(o).List()

		// Add new before deleting old so number of IP addresses doesn't drop below 2.
		for _, v := range add {
			input := &route53resolver.AssociateResolverEndpointIpAddressInput{
				IpAddress:          expandEndpointIPAddressUpdate(v),
				ResolverEndpointId: aws.String(d.Id()),
			}

			_, err := conn.AssociateResolverEndpointIpAddressWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "associating Route53 Resolver Endpoint (%s) IP address: %s", d.Id(), err)
			}

			if _, err := waitEndpointUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Endpoint (%s) update: %s", d.Id(), err)
			}
		}

		for _, v := range del {
			input := &route53resolver.DisassociateResolverEndpointIpAddressInput{
				IpAddress:          expandEndpointIPAddressUpdate(v),
				ResolverEndpointId: aws.String(d.Id()),
			}

			_, err := conn.DisassociateResolverEndpointIpAddressWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disassociating Route53 Resolver Endpoint (%s) IP address: %s", d.Id(), err)
			}

			if _, err := waitEndpointUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Endpoint (%s) update: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverConn(ctx)

	log.Printf("[DEBUG] Deleting Route53 Resolver Endpoint: %s", d.Id())
	_, err := conn.DeleteResolverEndpointWithContext(ctx, &route53resolver.DeleteResolverEndpointInput{
		ResolverEndpointId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Resolver Endpoint (%s): %s", d.Id(), err)
	}

	if _, err := waitEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Endpoint (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindResolverEndpointByID(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverEndpoint, error) {
	input := &route53resolver.GetResolverEndpointInput{
		ResolverEndpointId: aws.String(id),
	}

	output, err := conn.GetResolverEndpointWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResolverEndpoint == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ResolverEndpoint, nil
}

func findResolverEndpointIPAddressesByID(ctx context.Context, conn *route53resolver.Route53Resolver, id string) ([]*route53resolver.IpAddressResponse, error) {
	input := &route53resolver.ListResolverEndpointIpAddressesInput{
		ResolverEndpointId: aws.String(id),
	}
	var output []*route53resolver.IpAddressResponse

	err := conn.ListResolverEndpointIpAddressesPagesWithContext(ctx, input, func(page *route53resolver.ListResolverEndpointIpAddressesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.IpAddresses...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusEndpoint(ctx context.Context, conn *route53resolver.Route53Resolver, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindResolverEndpointByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitEndpointCreated(ctx context.Context, conn *route53resolver.Route53Resolver, id string, timeout time.Duration) (*route53resolver.ResolverEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{route53resolver.ResolverEndpointStatusCreating},
		Target:     []string{route53resolver.ResolverEndpointStatusOperational},
		Refresh:    statusEndpoint(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53resolver.ResolverEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitEndpointUpdated(ctx context.Context, conn *route53resolver.Route53Resolver, id string, timeout time.Duration) (*route53resolver.ResolverEndpoint, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    []string{route53resolver.ResolverEndpointStatusUpdating},
		Target:     []string{route53resolver.ResolverEndpointStatusOperational},
		Refresh:    statusEndpoint(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53resolver.ResolverEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitEndpointDeleted(ctx context.Context, conn *route53resolver.Route53Resolver, id string, timeout time.Duration) (*route53resolver.ResolverEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{route53resolver.ResolverEndpointStatusDeleting},
		Target:     []string{},
		Refresh:    statusEndpoint(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53resolver.ResolverEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func endpointHashIPAddress(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-%s-", m[names.AttrSubnetID].(string), m["ip"].(string)))
	return create.StringHashcode(buf.String())
}

func expandEndpointIPAddressUpdate(vIpAddress interface{}) *route53resolver.IpAddressUpdate {
	ipAddressUpdate := &route53resolver.IpAddressUpdate{}

	mIpAddress := vIpAddress.(map[string]interface{})

	if vSubnetId, ok := mIpAddress[names.AttrSubnetID].(string); ok && vSubnetId != "" {
		ipAddressUpdate.SubnetId = aws.String(vSubnetId)
	}
	if vIp, ok := mIpAddress["ip"].(string); ok && vIp != "" {
		ipAddressUpdate.Ip = aws.String(vIp)
	}
	if vIpId, ok := mIpAddress["ip_id"].(string); ok && vIpId != "" {
		ipAddressUpdate.IpId = aws.String(vIpId)
	}

	return ipAddressUpdate
}

func expandEndpointIPAddresses(vIpAddresses *schema.Set) []*route53resolver.IpAddressRequest {
	ipAddressRequests := []*route53resolver.IpAddressRequest{}

	for _, vIpAddress := range vIpAddresses.List() {
		ipAddressRequest := &route53resolver.IpAddressRequest{}

		mIpAddress := vIpAddress.(map[string]interface{})

		if vSubnetId, ok := mIpAddress[names.AttrSubnetID].(string); ok && vSubnetId != "" {
			ipAddressRequest.SubnetId = aws.String(vSubnetId)
		}
		if vIp, ok := mIpAddress["ip"].(string); ok && vIp != "" {
			ipAddressRequest.Ip = aws.String(vIp)
		}

		ipAddressRequests = append(ipAddressRequests, ipAddressRequest)
	}

	return ipAddressRequests
}

func flattenEndpointIPAddresses(ipAddresses []*route53resolver.IpAddressResponse) []interface{} {
	if ipAddresses == nil {
		return []interface{}{}
	}

	vIpAddresses := []interface{}{}

	for _, ipAddress := range ipAddresses {
		mIpAddress := map[string]interface{}{
			names.AttrSubnetID: aws.StringValue(ipAddress.SubnetId),
			"ip":               aws.StringValue(ipAddress.Ip),
			"ip_id":            aws.StringValue(ipAddress.IpId),
		}

		vIpAddresses = append(vIpAddresses, mIpAddress)
	}

	return vIpAddresses
}
