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

		// Keep in sync with aws_default_vpc_dhcp_options' schema.
		// See notes in vpc_default_vpc_dhcp_options.go.
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				AtLeastOneOf: []string{"domain_name", "domain_name_servers", "netbios_name_servers", "netbios_node_type", "ntp_servers"},
			},
			"domain_name_servers": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				AtLeastOneOf: []string{"domain_name", "domain_name_servers", "netbios_name_servers", "netbios_node_type", "ntp_servers"},
			},
			"netbios_name_servers": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				AtLeastOneOf: []string{"domain_name", "domain_name_servers", "netbios_name_servers", "netbios_node_type", "ntp_servers"},
			},
			"netbios_node_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				AtLeastOneOf: []string{"domain_name", "domain_name_servers", "netbios_name_servers", "netbios_node_type", "ntp_servers"},
			},
			"ntp_servers": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				AtLeastOneOf: []string{"domain_name", "domain_name_servers", "netbios_name_servers", "netbios_node_type", "ntp_servers"},
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

var (
	optionsMap = newDHCPOptionsMap(map[string]string{
		"domain_name":          "domain-name",
		"domain_name_servers":  "domain-name-servers",
		"netbios_name_servers": "netbios-name-servers",
		"netbios_node_type":    "netbios-node-type",
		"ntp_servers":          "ntp-servers",
	})
)

func resourceVPCDHCPOptionsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	dhcpConfigurations, err := optionsMap.resourceDataToDHCPConfigurations(d)

	if err != nil {
		return err
	}

	input := &ec2.CreateDhcpOptionsInput{
		DhcpConfigurations: dhcpConfigurations,
		TagSpecifications:  tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeDhcpOptions),
	}

	output, err := conn.CreateDhcpOptions(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 DHCP Options Set: %w", err)
	}

	d.SetId(aws.StringValue(output.DhcpOptions.DhcpOptionsId))

	return resourceVPCDHCPOptionsRead(d, meta)
}

func resourceVPCDHCPOptionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(propagationTimeout, func() (interface{}, error) {
		return FindDHCPOptionsByID(conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 DHCP Options Set %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 DHCP Options Set (%s): %w", d.Id(), err)
	}

	opts := outputRaw.(*ec2.DhcpOptions)

	ownerID := aws.StringValue(opts.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("dhcp-options/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("owner_id", ownerID)

	err = optionsMap.dhcpConfigurationsToResourceData(opts.DhcpConfigurations, d)

	if err != nil {
		return err
	}

	tags := KeyValueTags(opts.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceVPCDHCPOptionsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 DHCP Options Set (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceVPCDHCPOptionsRead(d, meta)
}

func resourceVPCDHCPOptionsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcs, err := FindVPCs(conn, &ec2.DescribeVpcsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"dhcp-options-id": d.Id(),
		}),
	})

	if err != nil {
		return fmt.Errorf("error reading EC2 DHCP Options Set (%s) associated VPCs: %w", d.Id(), err)
	}

	for _, v := range vpcs {
		vpcID := aws.StringValue(v.VpcId)

		log.Printf("[INFO] Disassociating EC2 DHCP Options Set (%s) from VPC (%s)", d.Id(), vpcID)
		_, err := conn.AssociateDhcpOptions(&ec2.AssociateDhcpOptionsInput{
			DhcpOptionsId: aws.String(DefaultDHCPOptionsID),
			VpcId:         aws.String(vpcID),
		})

		if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error disassociating EC2 DHCP Options Set (%s) from VPC (%s): %w", d.Id(), vpcID, err)
		}
	}

	input := &ec2.DeleteDhcpOptionsInput{
		DhcpOptionsId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting EC2 DHCP Options Set: %s", d.Id())
	_, err = tfresource.RetryWhenAWSErrCodeEquals(dhcpOptionSetDeletedTimeout, func() (interface{}, error) {
		return conn.DeleteDhcpOptions(input)
	}, errCodeDependencyViolation)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidDHCPOptionIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 DHCP Options Set (%s): %w", d.Id(), err)
	}

	return nil
}

// dhcpOptionsMap represents a mapping of Terraform resource attribute name to AWS API DHCP Option name.
type dhcpOptionsMap struct {
	tfToApi map[string]string
	apiToTf map[string]string
}

func newDHCPOptionsMap(tfToApi map[string]string) *dhcpOptionsMap {
	apiToTf := make(map[string]string)

	for k, v := range tfToApi {
		apiToTf[v] = k
	}

	return &dhcpOptionsMap{
		tfToApi: tfToApi,
		apiToTf: apiToTf,
	}
}

// dhcpConfigurationsToResourceData sets Terraform ResourceData from a list of AWS API DHCP configurations.
func (m *dhcpOptionsMap) dhcpConfigurationsToResourceData(dhcpConfigurations []*ec2.DhcpConfiguration, d *schema.ResourceData) error {
	for v := range m.tfToApi {
		d.Set(v, nil)
	}

	for _, dhcpConfiguration := range dhcpConfigurations {
		apiName := aws.StringValue(dhcpConfiguration.Key)
		if tfName, ok := m.apiToTf[apiName]; ok {
			switch v := d.Get(tfName).(type) {
			case string:
				d.Set(tfName, dhcpConfiguration.Values[0].Value)
			case []interface{}:
				var values []*string
				for _, v := range dhcpConfiguration.Values {
					values = append(values, v.Value)
				}
				d.Set(tfName, aws.StringValueSlice(values))
			default:
				return fmt.Errorf("Attribute (%s) is of unsupported type: %T", tfName, v)
			}
		} else {
			return fmt.Errorf("Unsupported DHCP option: %s", apiName)
		}
	}

	return nil
}

// resourceDataToNewDhcpConfigurations returns a list of AWS API DHCP configurations from Terraform ResourceData.
func (m *dhcpOptionsMap) resourceDataToDHCPConfigurations(d *schema.ResourceData) ([]*ec2.NewDhcpConfiguration, error) {
	var output []*ec2.NewDhcpConfiguration

	for tfName, apiName := range m.tfToApi {
		switch v := d.Get(tfName).(type) {
		case string:
			if v != "" {
				output = append(output, &ec2.NewDhcpConfiguration{
					Key:    aws.String(apiName),
					Values: aws.StringSlice([]string{v}),
				})
			}
		case []interface{}:
			var values []string
			for _, v := range v {
				v := v.(string)
				if v != "" {
					values = append(values, v)
				}
			}
			if len(values) > 0 {
				output = append(output, &ec2.NewDhcpConfiguration{
					Key:    aws.String(apiName),
					Values: aws.StringSlice(values),
				})
			}
		default:
			return nil, fmt.Errorf("Attribute (%s) is of unsupported type: %T", tfName, v)
		}
	}

	return output, nil
}
