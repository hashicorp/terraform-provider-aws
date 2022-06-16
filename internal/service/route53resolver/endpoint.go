package route53resolver

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	EndpointStatusDeleted = "DELETED"
)

const (
	endpointCreatedDefaultTimeout = 10 * time.Minute
	endpointUpdatedDefaultTimeout = 10 * time.Minute
	endpointDeletedDefaultTimeout = 10 * time.Minute
)

func ResourceEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceEndpointCreate,
		Read:   resourceEndpointRead,
		Update: resourceEndpointUpdate,
		Delete: resourceEndpointDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"direction": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					route53resolver.ResolverEndpointDirectionInbound,
					route53resolver.ResolverEndpointDirectionOutbound,
				}, false),
			},

			"ip_address": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 2,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Required: true,
						},
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
					},
				},
				Set: endpointHashIPAddress,
			},

			"security_group_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 64,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validResolverName,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"host_vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(endpointCreatedDefaultTimeout),
			Update: schema.DefaultTimeout(endpointUpdatedDefaultTimeout),
			Delete: schema.DefaultTimeout(endpointDeletedDefaultTimeout),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := &route53resolver.CreateResolverEndpointInput{
		CreatorRequestId: aws.String(resource.PrefixedUniqueId("tf-r53-resolver-endpoint-")),
		Direction:        aws.String(d.Get("direction").(string)),
		IpAddresses:      expandEndpointIPAddresses(d.Get("ip_address").(*schema.Set)),
		SecurityGroupIds: flex.ExpandStringSet(d.Get("security_group_ids").(*schema.Set)),
	}
	if v, ok := d.GetOk("name"); ok {
		req.Name = aws.String(v.(string))
	}
	if v, ok := d.GetOk("tags"); ok && len(v.(map[string]interface{})) > 0 {
		req.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Route53 Resolver endpoint: %#v", req)
	resp, err := conn.CreateResolverEndpoint(req)
	if err != nil {
		return fmt.Errorf("error creating Route53 Resolver endpoint: %s", err)
	}

	d.SetId(aws.StringValue(resp.ResolverEndpoint.Id))

	err = EndpointWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutCreate),
		[]string{route53resolver.ResolverEndpointStatusCreating},
		[]string{route53resolver.ResolverEndpointStatusOperational})
	if err != nil {
		return err
	}

	return resourceEndpointRead(d, meta)
}

func resourceEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	epRaw, state, err := endpointRefresh(conn, d.Id())()
	if err != nil {
		return fmt.Errorf("error getting Route53 Resolver endpoint (%s): %s", d.Id(), err)
	}
	if state == EndpointStatusDeleted {
		log.Printf("[WARN] Route53 Resolver endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	ep := epRaw.(*route53resolver.ResolverEndpoint)
	d.Set("arn", ep.Arn)
	d.Set("direction", ep.Direction)
	d.Set("host_vpc_id", ep.HostVPCId)
	d.Set("name", ep.Name)
	if err := d.Set("security_group_ids", flex.FlattenStringSet(ep.SecurityGroupIds)); err != nil {
		return err
	}

	ipAddresses := []interface{}{}
	req := &route53resolver.ListResolverEndpointIpAddressesInput{
		ResolverEndpointId: aws.String(d.Id()),
	}
	for {
		resp, err := conn.ListResolverEndpointIpAddresses(req)
		if err != nil {
			return fmt.Errorf("error getting Route53 Resolver endpoint (%s) IP addresses: %s", d.Id(), err)
		}

		ipAddresses = append(ipAddresses, flattenEndpointIPAddresses(resp.IpAddresses)...)

		if resp.NextToken == nil {
			break
		}
		req.NextToken = resp.NextToken
	}
	if err := d.Set("ip_address", schema.NewSet(endpointHashIPAddress, ipAddresses)); err != nil {
		return err
	}

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for Route53 Resolver endpoint (%s): %s", d.Get("arn").(string), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	if d.HasChange("name") {
		req := &route53resolver.UpdateResolverEndpointInput{
			ResolverEndpointId: aws.String(d.Id()),
			Name:               aws.String(d.Get("name").(string)),
		}

		log.Printf("[DEBUG] Updating Route53 Resolver endpoint: %#v", req)
		_, err := conn.UpdateResolverEndpoint(req)
		if err != nil {
			return fmt.Errorf("error updating Route53 Resolver endpoint (%s): %s", d.Id(), err)
		}

		err = EndpointWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutUpdate),
			[]string{route53resolver.ResolverEndpointStatusUpdating},
			[]string{route53resolver.ResolverEndpointStatusOperational})
		if err != nil {
			return err
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
			_, err := conn.AssociateResolverEndpointIpAddress(&route53resolver.AssociateResolverEndpointIpAddressInput{
				ResolverEndpointId: aws.String(d.Id()),
				IpAddress:          expandEndpointIPAddressUpdate(v),
			})
			if err != nil {
				return fmt.Errorf("error associating Route53 Resolver endpoint (%s) IP address: %s", d.Id(), err)
			}

			err = EndpointWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutUpdate),
				[]string{route53resolver.ResolverEndpointStatusUpdating},
				[]string{route53resolver.ResolverEndpointStatusOperational})
			if err != nil {
				return err
			}
		}

		for _, v := range del {
			_, err := conn.DisassociateResolverEndpointIpAddress(&route53resolver.DisassociateResolverEndpointIpAddressInput{
				ResolverEndpointId: aws.String(d.Id()),
				IpAddress:          expandEndpointIPAddressUpdate(v),
			})
			if err != nil {
				return fmt.Errorf("error disassociating Route53 Resolver endpoint (%s) IP address: %s", d.Id(), err)
			}

			err = EndpointWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutUpdate),
				[]string{route53resolver.ResolverEndpointStatusUpdating},
				[]string{route53resolver.ResolverEndpointStatusOperational})
			if err != nil {
				return err
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Route53 Resolver endpoint (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceEndpointRead(d, meta)
}

func resourceEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	log.Printf("[DEBUG] Deleting Route53 Resolver endpoint: %s", d.Id())
	_, err := conn.DeleteResolverEndpoint(&route53resolver.DeleteResolverEndpointInput{
		ResolverEndpointId: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting Route53 Resolver endpoint (%s): %s", d.Id(), err)
	}

	err = EndpointWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutDelete),
		[]string{route53resolver.ResolverEndpointStatusDeleting},
		[]string{EndpointStatusDeleted})
	if err != nil {
		return err
	}

	return nil
}

func endpointRefresh(conn *route53resolver.Route53Resolver, epId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.GetResolverEndpoint(&route53resolver.GetResolverEndpointInput{
			ResolverEndpointId: aws.String(epId),
		})
		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			return &route53resolver.ResolverEndpoint{}, EndpointStatusDeleted, nil
		}
		if err != nil {
			return nil, "", err
		}

		if statusMessage := aws.StringValue(resp.ResolverEndpoint.StatusMessage); statusMessage != "" {
			log.Printf("[INFO] Route 53 Resolver endpoint (%s) status message: %s", epId, statusMessage)
		}

		return resp.ResolverEndpoint, aws.StringValue(resp.ResolverEndpoint.Status), nil
	}
}

func EndpointWaitUntilTargetState(conn *route53resolver.Route53Resolver, epId string, timeout time.Duration, pending, target []string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    endpointRefresh(conn, epId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for Route53 Resolver endpoint (%s) to reach target state: %s", epId, err)
	}

	return nil
}

func endpointHashIPAddress(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["subnet_id"].(string)))
	return create.StringHashcode(buf.String())
}
