package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTrafficMirrorFilter() *schema.Resource {
	return &schema.Resource{
		Create: resourceTrafficMirrorFilterCreate,
		Read:   resourceTrafficMirrorFilterRead,
		Update: resourceTrafficMirrorFilterUpdate,
		Delete: resourceTrafficMirrorFilterDelete,
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
			"network_services": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"amazon-dns",
					}, false),
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceTrafficMirrorFilterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateTrafficMirrorFilterInput{}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.TagSpecifications = tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTrafficMirrorFilter)
	}

	out, err := conn.CreateTrafficMirrorFilter(input)

	if err != nil {
		return fmt.Errorf("creating EC2 Traffic Mirror Filter: %w", err)
	}

	d.SetId(aws.StringValue(out.TrafficMirrorFilter.TrafficMirrorFilterId))

	if v, ok := d.GetOk("network_services"); ok && v.(*schema.Set).Len() > 0 {
		input := &ec2.ModifyTrafficMirrorFilterNetworkServicesInput{
			AddNetworkServices:    flex.ExpandStringSet(v.(*schema.Set)),
			TrafficMirrorFilterId: aws.String(d.Id()),
		}

		_, err := conn.ModifyTrafficMirrorFilterNetworkServices(input)

		if err != nil {
			return fmt.Errorf("updating EC2 Traffic Mirror Filter (%s) network services: %w", d.Id(), err)
		}
	}

	return resourceTrafficMirrorFilterRead(d, meta)
}

func resourceTrafficMirrorFilterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	trafficMirrorFilter, err := FindTrafficMirrorFilterByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Traffic Mirror Filter %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Traffic Mirror Filter (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("traffic-mirror-filter/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("description", trafficMirrorFilter.Description)
	d.Set("network_services", aws.StringValueSlice(trafficMirrorFilter.NetworkServices))

	tags := KeyValueTags(trafficMirrorFilter.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceTrafficMirrorFilterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("network_services") {
		input := &ec2.ModifyTrafficMirrorFilterNetworkServicesInput{
			TrafficMirrorFilterId: aws.String(d.Id()),
		}

		o, n := d.GetChange("network_services")
		add := n.(*schema.Set).Difference(o.(*schema.Set))
		if add.Len() > 0 {
			input.AddNetworkServices = flex.ExpandStringSet(add)
		}
		del := o.(*schema.Set).Difference(n.(*schema.Set))
		if del.Len() > 0 {
			input.RemoveNetworkServices = flex.ExpandStringSet(del)
		}

		_, err := conn.ModifyTrafficMirrorFilterNetworkServices(input)

		if err != nil {
			return fmt.Errorf("updating EC2 Traffic Mirror Filter (%s) network services: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating EC2 Traffic Mirror Filter (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceTrafficMirrorFilterRead(d, meta)
}

func resourceTrafficMirrorFilterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Traffic Mirror Filter: %s", d.Id())
	_, err := conn.DeleteTrafficMirrorFilter(&ec2.DeleteTrafficMirrorFilterInput{
		TrafficMirrorFilterId: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("deleting EC2 Traffic Mirror Filter (%s): %w", d.Id(), err)
	}

	return nil
}
