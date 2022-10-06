package redshift

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceHSMConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceHSMConfigurationCreate,
		Read:   resourceHSMConfigurationRead,
		Update: resourceHSMConfigurationUpdate,
		Delete: resourceHSMConfigurationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hsm_configuration_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hsm_ip_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hsm_partition_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hsm_partition_password": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"hsm_server_public_certificate": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceHSMConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	hsmConfigurationID := d.Get("hsm_configuration_identifier").(string)
	input := &redshift.CreateHsmConfigurationInput{
		Description:                aws.String(d.Get("description").(string)),
		HsmConfigurationIdentifier: aws.String(hsmConfigurationID),
		HsmIpAddress:               aws.String(d.Get("hsm_ip_address").(string)),
		HsmPartitionName:           aws.String(d.Get("hsm_partition_name").(string)),
		HsmPartitionPassword:       aws.String(d.Get("hsm_partition_password").(string)),
		HsmServerPublicCertificate: aws.String(d.Get("hsm_server_public_certificate").(string)),
	}

	input.Tags = Tags(tags.IgnoreAWS())

	output, err := conn.CreateHsmConfiguration(input)

	if err != nil {
		return fmt.Errorf("creating Redshift HSM Configuration (%s): %s", hsmConfigurationID, err)
	}

	d.SetId(aws.StringValue(output.HsmConfiguration.HsmConfigurationIdentifier))

	return resourceHSMConfigurationRead(d, meta)
}

func resourceHSMConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	hsmConfiguration, err := FindHSMConfigurationByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift HSM Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Redshift HSM Configuration (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   redshift.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("hsmconfiguration:%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("hsm_configuration_identifier", hsmConfiguration.HsmConfigurationIdentifier)
	d.Set("hsm_ip_address", hsmConfiguration.HsmIpAddress)
	d.Set("hsm_partition_name", hsmConfiguration.HsmPartitionName)
	d.Set("description", hsmConfiguration.Description)
	d.Set("hsm_partition_password", d.Get("hsm_partition_password").(string))
	d.Set("hsm_server_public_certificate", d.Get("hsm_server_public_certificate").(string))

	tags := KeyValueTags(hsmConfiguration.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceHSMConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("updating Redshift HSM Configuration (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceHSMConfigurationRead(d, meta)
}

func resourceHSMConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	log.Printf("[DEBUG] Deleting Redshift HSM Configuration: %s", d.Id())
	_, err := conn.DeleteHsmConfiguration(&redshift.DeleteHsmConfigurationInput{
		HsmConfigurationIdentifier: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeHsmConfigurationNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Redshift HSM Configuration (%s): %w", d.Id(), err)
	}

	return nil
}
