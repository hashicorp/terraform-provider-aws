package iot

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDomainNameConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainNameConfigurationCreate,
		Read:   resourceDomainNameConfigurationRead,
		Update: resourceDomainNameConfigurationUpdate,
		Delete: resourceDomainNameConfigurationDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"server_certificate_arns": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
				Optional: true,
				ForceNew: true,
			},
			"service_type": {
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"DATA",
					"CREDENTIAL_PROVIDER",
					"JOBS",
				}, false),
				Optional: true,
				ForceNew: true,
				Default:  "DATA",
			},
			"validation_certificate_arn": {
				Type:         schema.TypeString,
				ValidateFunc: verify.ValidARN,
				Optional:     true,
				ForceNew:     true,
			},
			"authorizer_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_authorizer_override": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"default_authorizer_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ENABLED",
				ValidateFunc: validation.StringInSlice([]string{
					"ENABLED",
					"DISABLED",
				}, false),
			},
		},
	}
}

func resourceDomainNameConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	apiObject := &iot.CreateDomainConfigurationInput{
		DomainConfigurationName: aws.String(name),
	}

	if v, ok := d.GetOk("domain_name"); ok {
		apiObject.DomainName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("server_certificate_arns"); ok {
		apiObject.ServerCertificateArns = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("service_type"); ok {
		apiObject.ServiceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("validation_certificate_arn"); ok {
		apiObject.ValidationCertificateArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("authorizer_config"); ok {
		apiObject.AuthorizerConfig = expandIotDomainNameConfigurationAuthorizerConfig(v.([]interface{}))
	}

	if len(tags) > 0 {
		apiObject.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating IoT Domain Configuration: %s", name)
	output, err := conn.CreateDomainConfiguration(apiObject)

	if err != nil {
		return fmt.Errorf("error creating IoT Domain Configuration (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.DomainConfigurationName))

	return resourceDomainNameConfigurationRead(d, meta)
}

func expandIotDomainNameConfigurationAuthorizerConfig(l []interface{}) *iot.AuthorizerConfig {
	if len(l) < 1 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	iotAuthorizerConfig := &iot.AuthorizerConfig{
		AllowAuthorizerOverride: aws.Bool(m["allow_authorizer_override"].(bool)),
		DefaultAuthorizerName:   aws.String(m["allow_authorizer_override"].(string)),
	}

	return iotAuthorizerConfig
}

func resourceDomainNameConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig

	out, err := conn.DescribeDomainConfiguration(&iot.DescribeDomainConfigurationInput{
		DomainConfigurationName: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error reading domain details: %v", err)
	}

	d.Set("arn", out.DomainConfigurationArn)

	tags, _ := ListTags(conn, d.Get("arn").(string))

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceDomainNameConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	input := iot.UpdateDomainConfigurationInput{
		DomainConfigurationName: aws.String(d.Id()),
	}

	if d.HasChange("authorizer_config") {
		input.AuthorizerConfig = expandIotDomainNameConfigurationAuthorizerConfig(d.Get("authorizer_config").([]interface{}))
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating IoT Domain Name Configuration (%s) tags: %w", d.Id(), err)
		}
	}

	log.Printf("[INFO] Updating IoT Domain Configuration: %s", d.Id())
	_, err := conn.UpdateDomainConfiguration(&input)

	if err != nil {
		return fmt.Errorf("error updating IoT Domain Configuration (%s): %w", d.Id(), err)
	}

	return resourceCertificateRead(d, meta)
}

func resourceDomainNameConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	if d.Get("status").(string) == "ENABLED" {
		log.Printf("[INFO] Disabling IoT Domain Configuration: %s", d.Id())
		_, err := conn.UpdateDomainConfiguration(&iot.UpdateDomainConfigurationInput{
			DomainConfigurationName:   aws.String(d.Id()),
			DomainConfigurationStatus: aws.String("DISABLED"),
		})

		if err != nil {
			return fmt.Errorf("error disabling IoT Domain Configuration (%s): %s", d.Id(), err)
		}
	}

	_, err := conn.DeleteDomainConfiguration(&iot.DeleteDomainConfigurationInput{
		DomainConfigurationName: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error deleting certificate: %v", err)
	}

	return nil
}
