package ec2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCDHCPOptions() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCDHCPOptionsCreate,
		Read:   resourceVPCDHCPOptionsRead,
		Update: resourceVPCDHCPOptionsUpdate,
		Delete: resourceVPCDHCPOptionsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"domain_name_servers": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"ntp_servers": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"netbios_node_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"netbios_name_servers": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"tags": tftags.TagsSchema(),

			"tags_all": tftags.TagsSchemaComputed(),

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCDHCPOptionsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	setDHCPOption := func(key string) *ec2.NewDhcpConfiguration {
		log.Printf("[DEBUG] Setting DHCP option %s...", key)
		tfKey := strings.Replace(key, "-", "_", -1)

		value, ok := d.GetOk(tfKey)
		if !ok {
			return nil
		}

		if v, ok := value.(string); ok {
			return &ec2.NewDhcpConfiguration{
				Key: aws.String(key),
				Values: []*string{
					aws.String(v),
				},
			}
		}

		if v, ok := value.([]interface{}); ok {
			var s []*string
			for _, attr := range v {
				s = append(s, aws.String(attr.(string)))
			}

			return &ec2.NewDhcpConfiguration{
				Key:    aws.String(key),
				Values: s,
			}
		}

		return nil
	}

	createOpts := &ec2.CreateDhcpOptionsInput{
		DhcpConfigurations: []*ec2.NewDhcpConfiguration{
			setDHCPOption("domain-name"),
			setDHCPOption("domain-name-servers"),
			setDHCPOption("ntp-servers"),
			setDHCPOption("netbios-node-type"),
			setDHCPOption("netbios-name-servers"),
		},
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeDhcpOptions),
	}

	resp, err := conn.CreateDhcpOptions(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating DHCP Options Set: %s", err)
	}

	dos := resp.DhcpOptions
	d.SetId(aws.StringValue(dos.DhcpOptionsId))
	log.Printf("[INFO] DHCP Options Set ID: %s", d.Id())

	// Wait for the DHCP Options to become available
	log.Printf("[DEBUG] Waiting for DHCP Options (%s) to become available", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"created"},
		Refresh: resourceDHCPOptionsStateRefreshFunc(conn, d.Id()),
		Timeout: 5 * time.Minute,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for DHCP Options (%s) to become available: %s",
			d.Id(), err)
	}

	return resourceVPCDHCPOptionsRead(d, meta)
}

func resourceVPCDHCPOptionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeDhcpOptionsInput{
		DhcpOptionsIds: []*string{
			aws.String(d.Id()),
		},
	}

	resp, err := conn.DescribeDhcpOptions(req)
	if err != nil {
		if isNoSuchDhcpOptionIDErr(err) {
			log.Printf("[WARN] DHCP Options (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving DHCP Options: %s", err.Error())
	}

	if len(resp.DhcpOptions) == 0 {
		return nil
	}

	opts := resp.DhcpOptions[0]

	tags := tftags.Ec2KeyValueTags(opts.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("owner_id", opts.OwnerId)

	for _, cfg := range opts.DhcpConfigurations {
		tfKey := strings.Replace(*cfg.Key, "-", "_", -1)

		if _, ok := d.Get(tfKey).(string); ok {
			d.Set(tfKey, cfg.Values[0].Value)
		} else {
			values := make([]string, 0, len(cfg.Values))
			for _, v := range cfg.Values {
				values = append(values, *v.Value)
			}

			d.Set(tfKey, values)
		}
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.StringValue(opts.OwnerId),
		Resource:  fmt.Sprintf("dhcp-options/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceVPCDHCPOptionsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceVPCDHCPOptionsRead(d, meta)
}

func resourceVPCDHCPOptionsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	err := resource.Retry(3*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteDhcpOptions(&ec2.DeleteDhcpOptionsInput{
			DhcpOptionsId: aws.String(d.Id()),
		})

		if err == nil {
			return nil
		}

		ec2err, ok := err.(awserr.Error)
		if !ok {
			return resource.RetryableError(err)
		}

		switch ec2err.Code() {
		case "InvalidDhcpOptionsID.NotFound", "InvalidDhcpOptionID.NotFound":
			return nil
		case "DependencyViolation":
			// If it is a dependency violation, we want to disassociate
			// all VPCs using the given DHCP Options ID, and retry deleting.
			vpcs, err2 := FindVPCsByDHCPOptionsID(conn, d.Id())
			if err2 != nil {
				log.Printf("[ERROR] %s", err2)
				return resource.RetryableError(err2)
			}

			for _, vpc := range vpcs {
				log.Printf("[INFO] Disassociating DHCP Options Set %s from VPC %s...", d.Id(), *vpc.VpcId)
				if _, err := conn.AssociateDhcpOptions(&ec2.AssociateDhcpOptionsInput{
					DhcpOptionsId: aws.String("default"),
					VpcId:         vpc.VpcId,
				}); err != nil {
					return resource.RetryableError(err)
				}
			}
			return resource.RetryableError(err)
		default:
			return resource.NonRetryableError(err)
		}
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteDhcpOptions(&ec2.DeleteDhcpOptionsInput{
			DhcpOptionsId: aws.String(d.Id()),
		})
	}
	return err
}

func FindVPCsByDHCPOptionsID(conn *ec2.EC2, id string) ([]*ec2.Vpc, error) {
	req := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("dhcp-options-id"),
				Values: []*string{
					aws.String(id),
				},
			},
		},
	}

	resp, err := conn.DescribeVpcs(req)
	if err != nil {
		if tfawserr.ErrMessageContains(err, "InvalidVpcID.NotFound", "") {
			return nil, nil
		}
		return nil, err
	}

	return resp.Vpcs, nil
}

func resourceDHCPOptionsStateRefreshFunc(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		DescribeDhcpOpts := &ec2.DescribeDhcpOptionsInput{
			DhcpOptionsIds: []*string{
				aws.String(id),
			},
		}

		resp, err := conn.DescribeDhcpOptions(DescribeDhcpOpts)
		if err != nil {
			if isNoSuchDhcpOptionIDErr(err) {
				resp = nil
			} else {
				log.Printf("Error on DHCPOptionsStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		dos := resp.DhcpOptions[0]
		return dos, "created", nil
	}
}

func isNoSuchDhcpOptionIDErr(err error) bool {
	return tfawserr.ErrMessageContains(err, "InvalidDhcpOptionID.NotFound", "") || tfawserr.ErrMessageContains(err, "InvalidDhcpOptionsID.NotFound", "")
}
