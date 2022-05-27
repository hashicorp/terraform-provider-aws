package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTrafficMirrorTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceTrafficMirrorTargetCreate,
		Read:   resourceTrafficMirrorTargetRead,
		Update: resourceTrafficMirrorTargetUpdate,
		Delete: resourceTrafficMirrorTargetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"network_interface_id",
					"network_load_balancer_arn",
				},
			},
			"network_load_balancer_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"network_interface_id",
					"network_load_balancer_arn",
				},
				ValidateFunc: verify.ValidARN,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceTrafficMirrorTargetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateTrafficMirrorTargetInput{}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_interface_id"); ok {
		input.NetworkInterfaceId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_load_balancer_arn"); ok {
		input.NetworkLoadBalancerArn = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.TagSpecifications = tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTrafficMirrorTarget)
	}

	out, err := conn.CreateTrafficMirrorTarget(input)
	if err != nil {
		return fmt.Errorf("Error creating traffic mirror target %v", err)
	}

	d.SetId(aws.StringValue(out.TrafficMirrorTarget.TrafficMirrorTargetId))

	return resourceTrafficMirrorTargetRead(d, meta)
}

func resourceTrafficMirrorTargetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Traffic Mirror Target (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceTrafficMirrorTargetRead(d, meta)
}

func resourceTrafficMirrorTargetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	targetId := d.Id()
	input := &ec2.DescribeTrafficMirrorTargetsInput{
		TrafficMirrorTargetIds: []*string{&targetId},
	}

	out, err := conn.DescribeTrafficMirrorTargets(input)
	if tfawserr.ErrCodeEquals(err, "InvalidTrafficMirrorTargetId.NotFound") {
		log.Printf("[WARN] EC2 Traffic Mirror Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing EC2 Traffic Mirror Target (%s): %w", targetId, err)
	}

	if nil == out || 0 == len(out.TrafficMirrorTargets) {
		log.Printf("[WARN] EC2 Traffic Mirror Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	target := out.TrafficMirrorTargets[0]
	d.Set("description", target.Description)
	d.Set("network_interface_id", target.NetworkInterfaceId)
	d.Set("network_load_balancer_arn", target.NetworkLoadBalancerArn)

	tags := KeyValueTags(target.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("owner_id", target.OwnerId)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.StringValue(target.OwnerId),
		Resource:  fmt.Sprintf("traffic-mirror-target/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceTrafficMirrorTargetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	targetId := d.Id()
	input := &ec2.DeleteTrafficMirrorTargetInput{
		TrafficMirrorTargetId: &targetId,
	}

	_, err := conn.DeleteTrafficMirrorTarget(input)
	if nil != err {
		return fmt.Errorf("error deleting EC2 Traffic Mirror Target (%s): %w", targetId, err)
	}

	return nil
}
