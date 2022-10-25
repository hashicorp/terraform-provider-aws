package route53resolver

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

	input := &route53resolver.UpdateResolverDnssecConfigInput{
		ResourceId: aws.String(d.Get("resource_id").(string)),
		Validation: aws.String(route53resolver.ValidationEnable),
	}

	output, err := conn.UpdateResolverDnssecConfig(input)

	if err != nil {
		return fmt.Errorf("error creating Route 53 Resolver DNSSEC config: %w", err)
	}

	d.SetId(aws.StringValue(output.ResolverDNSSECConfig.Id))

	_, err = waitDNSSECConfigCreated(conn, d.Id())
	if err != nil {
		return err
	}

	return resourceDNSSECConfigRead(d, meta)
}

func resourceDNSSECConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	dnssecConfig, err := FindResolverDNSSECConfigByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver DNSSEC Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Route53 Resolver DNSSEC Config (%s): %w", d.Id(), err)
	}

	ownerID := aws.StringValue(dnssecConfig.OwnerId)
	resourceID := aws.StringValue(dnssecConfig.ResourceId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53resolver",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("resolver-dnssec-config/%s", resourceID),
	}.String()
	d.Set("arn", arn)
	d.Set("owner_id", ownerID)
	d.Set("resource_id", resourceID)
	d.Set("validation_status", dnssecConfig.ValidationStatus)

	return nil
}

func resourceDNSSECConfigDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	err := updateResolverDNSSECConfigValidation(conn, d.Get("resource_id").(string), route53resolver.ValidationDisable)

	if err != nil {
		return fmt.Errorf("deleting Route 53 Resolver DNSSEC Config (%s): %w", d.Id(), err)
	}

	if _, err = waitDNSSECConfigDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for Route 53 Resolver DNSSEC Config (%s) delete: %w", d.Id(), err)
	}

	// // (2) Update Route 53 ResolverDnssecConfig again, effectively deleting the resource
	// _, err = updateResolverDNSSECConfigValidation(conn, aws.StringValue(config.ResourceId), route53resolver.ValidationDisable)

	// if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
	// 	return nil
	// }

	// if err != nil {
	// 	return fmt.Errorf("error deleting Route 53 Resolver DNSSEC config (%s): %w", d.Id(), err)
	// }

	return nil
}

func updateResolverDNSSECConfigValidation(conn *route53resolver.Route53Resolver, resourceID, validation string) error {
	_, err := conn.UpdateResolverDnssecConfig(&route53resolver.UpdateResolverDnssecConfigInput{
		ResourceId: aws.String(resourceID),
		Validation: aws.String(validation),
	})

	return err
}

func FindResolverDNSSECConfigByID(conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverDnssecConfig, error) {
	input := &route53resolver.ListResolverDnssecConfigsInput{}
	var output *route53resolver.ResolverDnssecConfig

	// GetResolverDnssecConfig does not support query by ID.
	err := conn.ListResolverDnssecConfigsPages(input, func(page *route53resolver.ListResolverDnssecConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResolverDnssecConfigs {
			if aws.StringValue(v.Id) == id {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{LastRequest: input}
	}

	if validationStatus := aws.StringValue(output.ValidationStatus); validationStatus == route53resolver.ResolverDNSSECValidationStatusDisabled {
		return nil, &resource.NotFoundError{
			Message:     validationStatus,
			LastRequest: input,
		}
	}

	return output, nil
}

func statusDNSSECConfig(conn *route53resolver.Route53Resolver, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindResolverDNSSECConfigByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ValidationStatus), nil
	}
}

const (
	dnssecConfigCreatedTimeout = 10 * time.Minute
	dnssecConfigDeletedTimeout = 10 * time.Minute
)

func waitDNSSECConfigCreated(conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverDnssecConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverDNSSECValidationStatusEnabling},
		Target:  []string{route53resolver.ResolverDNSSECValidationStatusEnabled},
		Refresh: statusDNSSECConfig(conn, id),
		Timeout: dnssecConfigCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53resolver.ResolverDnssecConfig); ok {
		return output, err
	}

	return nil, err
}

func waitDNSSECConfigDeleted(conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverDnssecConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverDNSSECValidationStatusDisabling},
		Target:  []string{},
		Refresh: statusDNSSECConfig(conn, id),
		Timeout: dnssecConfigDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53resolver.ResolverDnssecConfig); ok {
		return output, err
	}

	return nil, err
}
