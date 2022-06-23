package devicefarm

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCEConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCEConfigurationCreate,
		Read:   resourceVPCEConfigurationRead,
		Update: resourceVPCEConfigurationUpdate,
		Delete: resourceVPCEConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpce_configuration_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"service_dns_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"vpce_configuration_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"vpce_service_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCEConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &devicefarm.CreateVPCEConfigurationInput{
		ServiceDnsName:        aws.String(d.Get("service_dns_name").(string)),
		VpceServiceName:       aws.String(d.Get("vpce_service_name").(string)),
		VpceConfigurationName: aws.String(d.Get("vpce_configuration_name").(string)),
	}

	if v, ok := d.GetOk("vpce_configuration_description"); ok {
		input.VpceConfigurationDescription = aws.String(v.(string))
	}

	out, err := conn.CreateVPCEConfiguration(input)
	if err != nil {
		return fmt.Errorf("Error creating DeviceFarm VPCE Configuration: %w", err)
	}

	arn := aws.StringValue(out.VpceConfiguration.Arn)
	log.Printf("[DEBUG] Successsfully Created DeviceFarm VPCE Configuration: %s", arn)
	d.SetId(arn)

	if len(tags) > 0 {
		if err := UpdateTags(conn, arn, nil, tags); err != nil {
			return fmt.Errorf("error updating DeviceFarm VPCE Configuration (%s) tags: %w", arn, err)
		}
	}

	return resourceVPCEConfigurationRead(d, meta)
}

func resourceVPCEConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	vpceConf, err := FindVPCEConfigurationByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm VPCE Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DeviceFarm VPCE Configuration (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(vpceConf.Arn)
	d.Set("arn", arn)
	d.Set("service_dns_name", vpceConf.ServiceDnsName)
	d.Set("vpce_configuration_description", vpceConf.VpceConfigurationDescription)
	d.Set("vpce_configuration_name", vpceConf.VpceConfigurationName)
	d.Set("vpce_service_name", vpceConf.VpceServiceName)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for DeviceFarm VPCE Configuration (%s): %w", arn, err)
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

func resourceVPCEConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &devicefarm.UpdateVPCEConfigurationInput{
			Arn: aws.String(d.Id()),
		}

		if d.HasChange("service_dns_name") {
			input.ServiceDnsName = aws.String(d.Get("service_dns_name").(string))
		}

		if d.HasChange("vpce_configuration_description") {
			input.VpceConfigurationDescription = aws.String(d.Get("vpce_configuration_description").(string))
		}

		if d.HasChange("vpce_configuration_name") {
			input.VpceConfigurationName = aws.String(d.Get("vpce_configuration_name").(string))
		}

		if d.HasChange("vpce_service_name") {
			input.VpceServiceName = aws.String(d.Get("vpce_service_name").(string))
		}

		log.Printf("[DEBUG] Updating DeviceFarm VPCE Configuration: %s", d.Id())
		_, err := conn.UpdateVPCEConfiguration(input)
		if err != nil {
			return fmt.Errorf("Error Updating DeviceFarm VPCE Configuration: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating DeviceFarm VPCE Configuration (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceVPCEConfigurationRead(d, meta)
}

func resourceVPCEConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn

	input := &devicefarm.DeleteVPCEConfigurationInput{
		Arn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DeviceFarm VPCE Configuration: %s", d.Id())
	_, err := conn.DeleteVPCEConfiguration(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting DeviceFarm VPCE Configuration: %w", err)
	}

	return nil
}
