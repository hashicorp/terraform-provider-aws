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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
			"arn": {
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
			"ip_address": {
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
						"subnet_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: endpointHashIPAddress,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validResolverName,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 64,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
	conn := meta.(*conns.AWSClient).Route53ResolverConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &route53resolver.CreateResolverEndpointInput{
		CreatorRequestId: aws.String(resource.PrefixedUniqueId("tf-r53-resolver-endpoint-")),
		Direction:        aws.String(d.Get("direction").(string)),
		IpAddresses:      expandEndpointIPAddresses(d.Get("ip_address").(*schema.Set)),
		SecurityGroupIds: flex.ExpandStringSet(d.Get("security_group_ids").(*schema.Set)),
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateResolverEndpointWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Route53 Resolver Endpoint: %s", err)
	}

	d.SetId(aws.StringValue(output.ResolverEndpoint.Id))

	if _, err := waitEndpointCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Route53 Resolver Endpoint (%s) create: %s", d.Id(), err)
	}

	return resourceEndpointRead(ctx, d, meta)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ep, err := FindResolverEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Route53 Resolver Endpoint (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(ep.Arn)
	d.Set("arn", arn)
	d.Set("direction", ep.Direction)
	d.Set("host_vpc_id", ep.HostVPCId)
	d.Set("name", ep.Name)
	d.Set("security_group_ids", aws.StringValueSlice(ep.SecurityGroupIds))

	ipAddresses, err := findResolverEndpointIPAddressesByID(ctx, conn, d.Id())

	if err != nil {
		return diag.Errorf("listing Route53 Resolver Endpoint (%s) IP addresses: %s", d.Id(), err)
	}

	if err := d.Set("ip_address", schema.NewSet(endpointHashIPAddress, flattenEndpointIPAddresses(ipAddresses))); err != nil {
		return diag.Errorf("setting ip_address: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for Route53 Resolver Endpoint (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	if d.HasChange("name") {
		_, err := conn.UpdateResolverEndpointWithContext(ctx, &route53resolver.UpdateResolverEndpointInput{
			Name:               aws.String(d.Get("name").(string)),
			ResolverEndpointId: aws.String(d.Id()),
		})

		if err != nil {
			return diag.Errorf("updating Route53 Resolver Endpoint (%s): %s", d.Id(), err)
		}

		if _, err := waitEndpointUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for Route53 Resolver Endpoint (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("ip_address") {
		oraw, nraw := d.GetChange("ip_address")
		o := oraw.(*schema.Set)
		n := nraw.(*schema.Set)
		del := o.Difference(n).List()
		add := n.Difference(o).List()

		// Add new before deleting old so number of IP addresses doesn't drop below 2.
		for _, v := range add {
			_, err := conn.AssociateResolverEndpointIpAddressWithContext(ctx, &route53resolver.AssociateResolverEndpointIpAddressInput{
				IpAddress:          expandEndpointIPAddressUpdate(v),
				ResolverEndpointId: aws.String(d.Id()),
			})

			if err != nil {
				return diag.Errorf("associating Route53 Resolver Endpoint (%s) IP address: %s", d.Id(), err)
			}

			if _, err := waitEndpointUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return diag.Errorf("waiting for Route53 Resolver Endpoint (%s) update: %s", d.Id(), err)
			}
		}

		for _, v := range del {
			_, err := conn.DisassociateResolverEndpointIpAddressWithContext(ctx, &route53resolver.DisassociateResolverEndpointIpAddressInput{
				IpAddress:          expandEndpointIPAddressUpdate(v),
				ResolverEndpointId: aws.String(d.Id()),
			})

			if err != nil {
				return diag.Errorf("disassociating Route53 Resolver Endpoint (%s) IP address: %s", d.Id(), err)
			}

			if _, err := waitEndpointUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return diag.Errorf("waiting for Route53 Resolver Endpoint (%s) update: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating Route53 Resolver Endpoint (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceEndpointRead(ctx, d, meta)
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	log.Printf("[DEBUG] Deleting Route53 Resolver Endpoint: %s", d.Id())
	_, err := conn.DeleteResolverEndpointWithContext(ctx, &route53resolver.DeleteResolverEndpointInput{
		ResolverEndpointId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Route53 Resolver Endpoint (%s): %s", d.Id(), err)
	}

	if _, err := waitEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Route53 Resolver Endpoint (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindResolverEndpointByID(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverEndpoint, error) {
	input := &route53resolver.GetResolverEndpointInput{
		ResolverEndpointId: aws.String(id),
	}

	output, err := conn.GetResolverEndpointWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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

func statusEndpoint(ctx context.Context, conn *route53resolver.Route53Resolver, id string) resource.StateRefreshFunc {
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
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
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
	buf.WriteString(fmt.Sprintf("%s-%s-", m["subnet_id"].(string), m["ip"].(string)))
	return create.StringHashcode(buf.String())
}

func expandEndpointIPAddressUpdate(vIpAddress interface{}) *route53resolver.IpAddressUpdate {
	ipAddressUpdate := &route53resolver.IpAddressUpdate{}

	mIpAddress := vIpAddress.(map[string]interface{})

	if vSubnetId, ok := mIpAddress["subnet_id"].(string); ok && vSubnetId != "" {
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

		if vSubnetId, ok := mIpAddress["subnet_id"].(string); ok && vSubnetId != "" {
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
			"subnet_id": aws.StringValue(ipAddress.SubnetId),
			"ip":        aws.StringValue(ipAddress.Ip),
			"ip_id":     aws.StringValue(ipAddress.IpId),
		}

		vIpAddresses = append(vIpAddresses, mIpAddress)
	}

	return vIpAddresses
}
