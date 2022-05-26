package route53resolver

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceDNSSECConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceDNSSECConfigCreate,
		Read:   resourceDNSSECConfigRead,
		Delete: resourceDNSSECConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
	}
}

func resourceDNSSECConfigCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	req := &route53resolver.UpdateResolverDnssecConfigInput{
		ResourceId: aws.String(d.Get("resource_id").(string)),
		Validation: aws.String(route53resolver.ValidationEnable),
	}

	log.Printf("[DEBUG] Creating Route 53 Resolver DNSSEC config: %#v", req)
	resp, err := conn.UpdateResolverDnssecConfig(req)
	if err != nil {
		return fmt.Errorf("error creating Route 53 Resolver DNSSEC config: %w", err)
	}

	d.SetId(aws.StringValue(resp.ResolverDNSSECConfig.Id))

	_, err = WaitDNSSECConfigCreated(conn, d.Id())
	if err != nil {
		return err
	}

	return resourceDNSSECConfigRead(d, meta)
}

func resourceDNSSECConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	config, err := FindResolverDNSSECConfigByID(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error getting Route 53 Resolver DNSSEC config (%s): %w", d.Id(), err)
	}

	if config == nil || aws.StringValue(config.ValidationStatus) == route53resolver.ResolverDNSSECValidationStatusDisabled {
		log.Printf("[WARN] Route 53 Resolver DNSSEC config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("owner_id", config.OwnerId)
	d.Set("resource_id", config.ResourceId)
	d.Set("validation_status", config.ValidationStatus)

	configArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53resolver",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.StringValue(config.OwnerId),
		Resource:  fmt.Sprintf("resolver-dnssec-config/%s", aws.StringValue(config.ResourceId)),
	}.String()
	d.Set("arn", configArn)

	return nil
}

func resourceDNSSECConfigDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	// To delete a Route 53 ResolverDnssecConfig, it must be:
	// (1) updated to a "DISABLED" state
	// (2) updated again to be permanently removed
	//
	// To determine how many Updates are required,
	// we first find the config by ID and proceed as follows:

	config, err := FindResolverDNSSECConfigByID(conn, d.Id())

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route 53 Resolver DNSSEC config (%s): %w", d.Id(), err)
	}

	if config == nil {
		return nil
	}

	// (1) Update Route 53 ResolverDnssecConfig to "DISABLED" state, if necessary
	if aws.StringValue(config.ValidationStatus) == route53resolver.ResolverDNSSECValidationStatusEnabled {
		config, err = updateResolverDNSSECConfigValidation(conn, aws.StringValue(config.ResourceId), route53resolver.ValidationDisable)
		if err != nil {
			return fmt.Errorf("error deleting Route 53 Resolver DNSSEC config (%s): %w", d.Id(), err)
		}
		if config == nil {
			return nil
		}
	}

	// (1.a) Wait for Route 53 ResolverDnssecConfig to reach "DISABLED" state, if necessary
	if aws.StringValue(config.ValidationStatus) != route53resolver.ResolverDNSSECValidationStatusDisabled {
		if _, err = WaitDNSSECConfigDisabled(conn, d.Id()); err != nil {
			if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
				return nil
			}

			return fmt.Errorf("error waiting for Route 53 Resolver DNSSEC config (%s) to be disabled: %w", d.Id(), err)
		}
	}

	// (2) Update Route 53 ResolverDnssecConfig again, effectively deleting the resource
	_, err = updateResolverDNSSECConfigValidation(conn, aws.StringValue(config.ResourceId), route53resolver.ValidationDisable)

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route 53 Resolver DNSSEC config (%s): %w", d.Id(), err)
	}

	return nil
}

func updateResolverDNSSECConfigValidation(conn *route53resolver.Route53Resolver, resourceId, validation string) (*route53resolver.ResolverDnssecConfig, error) {
	output, err := conn.UpdateResolverDnssecConfig(&route53resolver.UpdateResolverDnssecConfigInput{
		ResourceId: aws.String(resourceId),
		Validation: aws.String(validation),
	})

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.ResolverDNSSECConfig, nil
}
