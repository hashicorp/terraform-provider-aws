package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	route53ResolverDnssecConfigStatusNotFound = "NOT_FOUND"
)

func resourceAwsRoute53ResolverDnssecConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53ResolverDnssecConfigCreate,
		Read:   resourceAwsRoute53ResolverDnssecConfigRead,
		Delete: resourceAwsRoute53ResolverDnssecConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"validation_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsRoute53ResolverDnssecConfigCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	req := &route53resolver.UpdateResolverDnssecConfigInput{
		ResourceId: aws.String(d.Get("resource_id").(string)),
		Validation: aws.String(route53resolver.ValidationEnable),
	}

	log.Printf("[DEBUG] Creating Route53 Resolver DNSSEC config: %#v", req)
	resp, err := conn.UpdateResolverDnssecConfig(req)
	if err != nil {
		return fmt.Errorf("error creating Route53 Resolver DNSSEC config: %w", err)
	}

	d.SetId(aws.StringValue(resp.ResolverDNSSECConfig.ResourceId))

	err = route53ResolverDnssecConfigWait(conn, d.Id(), d.Timeout(schema.TimeoutCreate),
		[]string{route53resolver.ResolverDNSSECValidationStatusEnabling},
		[]string{route53resolver.ResolverDNSSECValidationStatusEnabled})
	if err != nil {
		return err
	}

	return resourceAwsRoute53ResolverDnssecConfigRead(d, meta)
}

func resourceAwsRoute53ResolverDnssecConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn
	ec2Conn := meta.(*AWSClient).ec2conn

	vpc, err := vpcDescribe(ec2Conn, d.Id())
	if err != nil {
		return fmt.Errorf("error getting VPC associated with Route53 Resolver DNSSEC config (%s): %w", d.Id(), err)
	}

	// GetResolverDnssecConfig returns AccessDeniedException if sending a request with non-existing VPC id
	if vpc == nil {
		log.Printf("[WARN] VPC associated with Resolver DNSSEC config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	raw, state, err := route53ResolverDnssecConfigRefresh(conn, d.Id())()
	if err != nil {
		return fmt.Errorf("error getting Route53 Resolver DNSSEC config (%s): %w", d.Id(), err)
	}

	if state == route53ResolverDnssecConfigStatusNotFound || state == route53resolver.ResolverDNSSECValidationStatusDisabled {
		log.Printf("[WARN] Route53 Resolver DNSSEC config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	out := raw.(*route53resolver.ResolverDnssecConfig)
	d.Set("id", out.Id)
	d.Set("owner_id", out.OwnerId)
	d.Set("resource_id", out.ResourceId)
	d.Set("validation_status", out.ValidationStatus)

	configArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "route53resolver",
		Region:    meta.(*AWSClient).region,
		AccountID: aws.StringValue(out.OwnerId),
		Resource:  fmt.Sprintf("resolver-dnssec-config/%s", aws.StringValue(out.ResourceId)),
	}.String()
	d.Set("arn", configArn)

	return nil
}

func resourceAwsRoute53ResolverDnssecConfigDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	log.Printf("[DEBUG] Deleting Route53 Resolver DNSSEC config: %s", d.Id())
	_, err := conn.UpdateResolverDnssecConfig(&route53resolver.UpdateResolverDnssecConfigInput{
		ResourceId: aws.String(d.Id()),
		Validation: aws.String(route53resolver.ValidationDisable),
	})
	if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting Route53 Resolver DNSSEC config (%s): %w", d.Id(), err)
	}

	err = route53ResolverDnssecConfigWait(conn, d.Id(), d.Timeout(schema.TimeoutDelete),
		[]string{route53resolver.ResolverDNSSECValidationStatusDisabling},
		[]string{route53resolver.ResolverDNSSECValidationStatusDisabled})
	if err != nil {
		return err
	}

	return nil
}

func route53ResolverDnssecConfigWait(conn *route53resolver.Route53Resolver, id string, timeout time.Duration, pending, target []string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    route53ResolverDnssecConfigRefresh(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for Route53 Resolver DNSSEC config (%s) to reach target state: %w", id, err)
	}

	return nil
}

func route53ResolverDnssecConfigRefresh(conn *route53resolver.Route53Resolver, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.GetResolverDnssecConfig(&route53resolver.GetResolverDnssecConfigInput{
			ResourceId: aws.String(id),
		})

		if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			return &route53resolver.ResolverDnssecConfig{}, route53ResolverDnssecConfigStatusNotFound, nil
		}

		if err != nil {
			return nil, "", err
		}

		return resp.ResolverDNSSECConfig, aws.StringValue(resp.ResolverDNSSECConfig.ValidationStatus), nil
	}
}
