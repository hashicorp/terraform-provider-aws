package aws

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsEc2Host() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2HostCreate,
		Read:   resourceAwsEc2HostRead,
		Update: resourceAwsEc2HostUpdate,
		Delete: resourceAwsEc2HostDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_placement": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.AutoPlacementOn,
				ValidateFunc: validation.StringInSlice(ec2.AutoPlacement_Values(), false),
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"host_recovery": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.HostRecoveryOff,
				ValidateFunc: validation.StringInSlice(ec2.HostRecovery_Values(), false),
			},
			"instance_family": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"instance_family", "instance_type"},
			},
			"instance_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"instance_family", "instance_type"},
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

type awsHostsOpts struct {
	AutoPlacement    *string
	AvailabilityZone *string
	InstanceType     *string
	HostRecovery     *string
}

func buildAwsHostsOpts(d *schema.ResourceData) *awsHostsOpts {

	instanceType := d.Get("instance_type").(string)
	opts := &awsHostsOpts{
		AutoPlacement:    aws.String(d.Get("auto_placement").(string)),
		AvailabilityZone: aws.String(d.Get("availability_zone").(string)),
		InstanceType:     aws.String(instanceType),
		HostRecovery:     aws.String(d.Get("host_recovery").(string)),
	}
	return opts
}

// resourceAwsEc2HostCreate allocates a Dedicated Host to your account.
// https://docs.aws.amazon.com/en_pv/AWSEC2/latest/APIReference/API_AllocateHosts.html
func resourceAwsEc2HostCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	hostOpts := buildAwsHostsOpts(d)

	tagsSpec := ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2.ResourceTypeDedicatedHost)

	// Build the creation struct
	runOpts := &ec2.AllocateHostsInput{
		AvailabilityZone: hostOpts.AvailabilityZone,
		Quantity:         aws.Int64(int64(1)),
		InstanceType:     hostOpts.InstanceType,
		HostRecovery:     hostOpts.HostRecovery,
		AutoPlacement:    hostOpts.AutoPlacement,
	}

	if len(tagsSpec) > 0 {
		runOpts.TagSpecifications = tagsSpec
	}

	var runResp *ec2.AllocateHostsOutput
	err := resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		runResp, err = conn.AllocateHosts(runOpts)
		return resource.RetryableError(err)
	})
	if isResourceTimeoutError(err) {
		runResp, err = conn.AllocateHosts(runOpts)
	}
	if err != nil {
		return fmt.Errorf("Error launching host : %s", err)
	}
	if runResp == nil || len(runResp.HostIds) == 0 {
		return errors.New("Error launching source host: no hosts returned in response")
	}

	log.Printf("[INFO] Host ID: %s", *runResp.HostIds[0])
	d.SetId(*runResp.HostIds[0])

	return resourceAwsEc2HostRead(d, meta)
}

func resourceAwsEc2HostRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	host, err := finder.HostByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Host %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Host (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*AWSClient).region,
		AccountID: aws.StringValue(host.OwnerId),
		Resource:  fmt.Sprintf("dedicated-host/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("auto_placement", host.AutoPlacement)
	d.Set("availibility_zone", host.AvailabilityZone)
	d.Set("host_recovery", host.HostRecovery)
	d.Set("instance_family", host.HostProperties.InstanceFamily)
	d.Set("instance_type", host.HostProperties.InstanceType)
	d.Set("owner_id", host.OwnerId)

	tags := keyvaluetags.Ec2KeyValueTags(host.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

// resourceAwsDedicatedHostUpdate modifies AWS Host AutoPlacement and HostRecovery settings.
// When auto-placement is enabled, any instances that you launch with a tenancy of host but without a specific host ID are placed onto any available
// Dedicated Host in your account that has auto-placement enabled.
// https://docs.aws.amazon.com/en_pv/AWSEC2/latest/APIReference/API_ModifyHosts.html
func resourceAwsEc2HostUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChangesExcept("tags", "tags_all") {
		requestUpdate := false
		req := &ec2.ModifyHostsInput{
			HostIds: []*string{aws.String(d.Id())},
		}
		if d.HasChange("auto_placement") {
			req.AutoPlacement = aws.String(d.Get("auto_placement").(string))
			requestUpdate = true
		}
		// Indicates whether to enable or disable host recovery for the Dedicated Host.
		// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/dedicated-hosts-recovery.html
		if d.HasChange("host_recovery") {
			req.HostRecovery = aws.String(d.Get("host_recovery").(string))
			requestUpdate = true
		}
		if requestUpdate {
			err := resource.Retry(30*time.Second, func() *resource.RetryError {
				_, err := conn.ModifyHosts(req)
				if err != nil {
					return resource.RetryableError(err)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("Error modifying host %s: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Host (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsEc2HostRead(d, meta)
}

func resourceAwsEc2HostDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[INFO] Deleting EC2 Host: %s", d.Id())
	output, err := conn.ReleaseHosts(&ec2.ReleaseHostsInput{
		HostIds: aws.StringSlice([]string{d.Id()}),
	})

	if err == nil && output != nil {
		err = tfec2.UnsuccessfulItemsError(output.Unsuccessful)
	}

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeClientInvalidHostIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error releasing EC2 Host (%s): %w", d.Id(), err)
	}

	if _, err := waiter.HostDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Host (%s) delete: %w", d.Id(), err)
	}

	return nil
}
