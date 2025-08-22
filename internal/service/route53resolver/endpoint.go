// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_resolver_endpoint", name="Endpoint")
// @Tags(identifierAttribute="arn")
func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEndpointCreate,
		ReadWithoutTimeout:   resourceEndpointRead,
		UpdateWithoutTimeout: resourceEndpointUpdate,
		DeleteWithoutTimeout: resourceEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,

		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceResolverEndpointV0().CoreConfigSchema().ImpliedType(),
				Upgrade: upgradeResolverEndpointV0toV1,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"direction": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ResolverEndpointDirection](),
			},
			"host_vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrIPAddress: {
				Type:     schema.TypeList,
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
						"ipv6": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsIPv6Address,
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
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validResolverName,
			},
			"resolver_endpoint_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ResolverEndpointType](),
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
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.Protocol](),
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
	}
}

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	input := &route53resolver.CreateResolverEndpointInput{
		CreatorRequestId: aws.String(id.PrefixedUniqueId("tf-r53-resolver-endpoint-")),
		Direction:        awstypes.ResolverEndpointDirection(d.Get("direction").(string)),
		IpAddresses:      expandEndpointIPAddresses(d.Get(names.AttrIPAddress).([]interface{})),
		SecurityGroupIds: flex.ExpandStringValueSet(d.Get(names.AttrSecurityGroupIDs).(*schema.Set)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("protocols"); ok && v.(*schema.Set).Len() > 0 {
		input.Protocols = flex.ExpandStringyValueSet[awstypes.Protocol](v.(*schema.Set))
	}

	if v, ok := d.GetOk("resolver_endpoint_type"); ok {
		input.ResolverEndpointType = awstypes.ResolverEndpointType(v.(string))
	}

	output, err := conn.CreateResolverEndpoint(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Resolver Endpoint: %s", err)
	}

	d.SetId(aws.ToString(output.ResolverEndpoint.Id))

	if _, err := waitEndpointCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Endpoint (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	output, err := findResolverEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Resolver Endpoint (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set("direction", output.Direction)
	d.Set("host_vpc_id", output.HostVPCId)
	d.Set(names.AttrName, output.Name)
	d.Set("protocols", flex.FlattenStringyValueSet[awstypes.Protocol](output.Protocols))
	d.Set("resolver_endpoint_type", output.ResolverEndpointType)
	d.Set(names.AttrSecurityGroupIDs, output.SecurityGroupIds)

	ipAddresses, err := findResolverEndpointIPAddressesByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Route53 Resolver Endpoint (%s) IP addresses: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrIPAddress, flattenEndpointIPAddresses(ipAddresses)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ip_address: %s", err)
	}

	return diags
}

func resourceEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	if d.HasChanges(names.AttrName, "protocols", "resolver_endpoint_type") {
		input := &route53resolver.UpdateResolverEndpointInput{
			ResolverEndpointId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange("protocols") {
			input.Protocols = flex.ExpandStringyValueSet[awstypes.Protocol](d.Get("protocols").(*schema.Set))
		}

		if d.HasChange("resolver_endpoint_type") {
			input.ResolverEndpointType = awstypes.ResolverEndpointType(d.Get("resolver_endpoint_type").(string))
		}

		_, err := conn.UpdateResolverEndpoint(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Resolver Endpoint (%s): %s", d.Id(), err)
		}

		if _, err := waitEndpointUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Endpoint (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrIPAddress) {
		oraw, nraw := d.GetChange(names.AttrIPAddress)
		oList := toIfaceSlice(oraw)
		nList := toIfaceSlice(nraw)

		adds, dels := diffIPAddressLists(oList, nList)

		for _, v := range adds {
			input := &route53resolver.AssociateResolverEndpointIpAddressInput{
				IpAddress:          expandEndpointIPAddressUpdate(v),
				ResolverEndpointId: aws.String(d.Id()),
			}
			if _, err := conn.AssociateResolverEndpointIpAddress(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "associating Route53 Resolver Endpoint (%s) IP address: %s", d.Id(), err)
			}
			if _, err := waitEndpointUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Endpoint (%s) update: %s", d.Id(), err)
			}
		}

		for _, v := range dels {
			input := &route53resolver.DisassociateResolverEndpointIpAddressInput{
				IpAddress:          expandEndpointIPAddressUpdate(v),
				ResolverEndpointId: aws.String(d.Id()),
			}
			if _, err := conn.DisassociateResolverEndpointIpAddress(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "disassociating Route53 Resolver Endpoint (%s) IP address: %s", d.Id(), err)
			}
			if _, err := waitEndpointUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Route53 Resolver Endpoint (%s) update: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53ResolverClient(ctx)

	log.Printf("[DEBUG] Deleting Route53 Resolver Endpoint: %s", d.Id())
	_, err := conn.DeleteResolverEndpoint(ctx, &route53resolver.DeleteResolverEndpointInput{
		ResolverEndpointId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findResolverEndpointByID(ctx context.Context, conn *route53resolver.Client, id string) (*awstypes.ResolverEndpoint, error) {
	input := &route53resolver.GetResolverEndpointInput{
		ResolverEndpointId: aws.String(id),
	}

	output, err := conn.GetResolverEndpoint(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findResolverEndpointIPAddressesByID(ctx context.Context, conn *route53resolver.Client, id string) ([]awstypes.IpAddressResponse, error) {
	input := &route53resolver.ListResolverEndpointIpAddressesInput{
		ResolverEndpointId: aws.String(id),
	}
	var output []awstypes.IpAddressResponse

	pages := route53resolver.NewListResolverEndpointIpAddressesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.IpAddresses...)
	}

	return output, nil
}

func statusEndpoint(ctx context.Context, conn *route53resolver.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findResolverEndpointByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitEndpointCreated(ctx context.Context, conn *route53resolver.Client, id string, timeout time.Duration) (*awstypes.ResolverEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ResolverEndpointStatusCreating),
		Target:     enum.Slice(awstypes.ResolverEndpointStatusOperational),
		Refresh:    statusEndpoint(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResolverEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitEndpointUpdated(ctx context.Context, conn *route53resolver.Client, id string, timeout time.Duration) (*awstypes.ResolverEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ResolverEndpointStatusUpdating),
		Target:     enum.Slice(awstypes.ResolverEndpointStatusOperational),
		Refresh:    statusEndpoint(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResolverEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitEndpointDeleted(ctx context.Context, conn *route53resolver.Client, id string, timeout time.Duration) (*awstypes.ResolverEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ResolverEndpointStatusDeleting),
		Target:     []string{},
		Refresh:    statusEndpoint(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResolverEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func endpointHashIPAddress(v any) int {
	var buf bytes.Buffer
	m := v.(map[string]any)
	fmt.Fprintf(&buf, "%s-%s-", m[names.AttrSubnetID].(string), m["ip"].(string))
	return create.StringHashcode(buf.String())
}

func expandEndpointIPAddressUpdate(vIpAddress any) *awstypes.IpAddressUpdate {
	ipAddressUpdate := &awstypes.IpAddressUpdate{}

	mIpAddress := vIpAddress.(map[string]any)

	if vSubnetId, ok := mIpAddress[names.AttrSubnetID].(string); ok && vSubnetId != "" {
		ipAddressUpdate.SubnetId = aws.String(vSubnetId)
	}
	if vIp, ok := mIpAddress["ip"].(string); ok && vIp != "" {
		ipAddressUpdate.Ip = aws.String(vIp)
	}
	if vIpv6, ok := mIpAddress["ipv6"].(string); ok && vIpv6 != "" {
		ipAddressUpdate.Ipv6 = aws.String(vIpv6)
	}
	if vIpId, ok := mIpAddress["ip_id"].(string); ok && vIpId != "" {
		ipAddressUpdate.IpId = aws.String(vIpId)
	}

	return ipAddressUpdate
}

func expandEndpointIPAddresses(vIpAddresses []interface{}) []awstypes.IpAddressRequest {
	ipAddressRequests := make([]awstypes.IpAddressRequest, 0, len(vIpAddresses))

	for _, vIpAddress := range vIpAddresses {
		ipAddressRequest := awstypes.IpAddressRequest{}

		mIpAddress := vIpAddress.(map[string]any)

		if vSubnetId, ok := mIpAddress[names.AttrSubnetID].(string); ok && vSubnetId != "" {
			ipAddressRequest.SubnetId = aws.String(vSubnetId)
		}
		if vIp, ok := mIpAddress["ip"].(string); ok && vIp != "" {
			ipAddressRequest.Ip = aws.String(vIp)
		}
		if vIpv6, ok := mIpAddress["ipv6"].(string); ok && vIpv6 != "" {
			ipAddressRequest.Ipv6 = aws.String(vIpv6)
		}

		ipAddressRequests = append(ipAddressRequests, ipAddressRequest)
	}

	return ipAddressRequests
}

func flattenEndpointIPAddresses(ipAddresses []awstypes.IpAddressResponse) []any {
	if ipAddresses == nil {
		return []any{}
	}

	vIpAddresses := []any{}

	for _, ipAddress := range ipAddresses {
		mIpAddress := map[string]any{
			names.AttrSubnetID: aws.ToString(ipAddress.SubnetId),
			"ipv6":             aws.ToString(ipAddress.Ipv6),
			"ip":               aws.ToString(ipAddress.Ip),
			"ip_id":            aws.ToString(ipAddress.IpId),
		}

		vIpAddresses = append(vIpAddresses, mIpAddress)
	}

	return vIpAddresses
}

func resourceResolverEndpointV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			names.AttrIPAddress: {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip":               {Type: schema.TypeString, Optional: true, Computed: true},
						"ipv6":             {Type: schema.TypeString, Optional: true, Computed: true},
						"ip_id":            {Type: schema.TypeString, Computed: true},
						names.AttrSubnetID: {Type: schema.TypeString, Required: true},
					},
				},
			},
		},
	}
}

func upgradeResolverEndpointV0toV1(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	const key = names.AttrIPAddress

	v, ok := rawState[key]
	if !ok || v == nil {
		return rawState, nil
	}

	switch t := v.(type) {
	case *schema.Set:
		list := t.List()
		sort.SliceStable(list, func(i, j int) bool {
			mi := list[i].(map[string]interface{})
			mj := list[j].(map[string]interface{})

			si := fmt.Sprint(mi[names.AttrSubnetID])
			sj := fmt.Sprint(mj[names.AttrSubnetID])
			if si != sj {
				return si < sj
			}
			ipi := fmt.Sprint(mi["ip"])
			ipj := fmt.Sprint(mj["ip"])
			if ipi == "" && ipj != "" {
				return false
			}
			if ipj == "" && ipi != "" {
				return true
			}
			return ipi < ipj
		})
		rawState[key] = list

	case []interface{}:
		list := t
		sort.SliceStable(list, func(i, j int) bool {
			mi := list[i].(map[string]interface{})
			mj := list[j].(map[string]interface{})

			si := fmt.Sprint(mi[names.AttrSubnetID])
			sj := fmt.Sprint(mj[names.AttrSubnetID])
			if si != sj {
				return si < sj
			}
			ipi := fmt.Sprint(mi["ip"])
			ipj := fmt.Sprint(mj["ip"])
			if ipi == "" && ipj != "" {
				return false
			}
			if ipj == "" && ipi != "" {
				return true
			}
			return ipi < ipj
		})
		rawState[key] = list

	default:
		return rawState, nil
	}

	return rawState, nil
}
func toIfaceSlice(v any) []interface{} {
	switch t := v.(type) {
	case []interface{}:
		return t
	case *schema.Set:
		return t.List()
	case nil:
		return nil
	default:
		return nil
	}
}

func diffIPAddressLists(oldList, newList []interface{}) (adds []map[string]any, dels []map[string]any) {
	type entry struct {
		m map[string]any
	}
	keyOf := func(m map[string]any) string {
		subnet := fmt.Sprint(m[names.AttrSubnetID])
		ip, _ := m["ip"].(string)
		if ip == "" {
			ip = "(auto)"
		}
		return subnet + "|" + ip
	}

	oldCount := map[string]int{}
	oldFirst := map[string]map[string]any{}
	for _, v := range oldList {
		m := v.(map[string]any)
		k := keyOf(m)
		oldCount[k]++
		if _, ok := oldFirst[k]; !ok {
			oldFirst[k] = m
		}
	}

	newCount := map[string]int{}
	newFirst := map[string]map[string]any{}
	for _, v := range newList {
		m := v.(map[string]any)
		k := keyOf(m)
		newCount[k]++
		if _, ok := newFirst[k]; !ok {
			newFirst[k] = m
		}
	}

	for k, n := range newCount {
		o := oldCount[k]
		for i := 0; i < n-o; i++ {
			if n > o {
				adds = append(adds, newFirst[k])
			}
		}
	}

	for k, o := range oldCount {
		n := newCount[k]
		for i := 0; i < o-n; i++ {
			if o > n {
				dels = append(dels, oldFirst[k])
			}
		}
	}

	return
}
