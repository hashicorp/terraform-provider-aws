package iot

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDomainConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainConfigurationCreate,
		ReadWithoutTimeout:   resourceDomainConfigurationRead,
		UpdateWithoutTimeout: resourceDomainConfigurationUpdate,
		DeleteWithoutTimeout: resourceDomainConfigurationDelete,

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
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
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"server_certificate_arns": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"service_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      iot.ServiceTypeData,
				ValidateFunc: validation.StringInSlice(iot.ServiceType_Values(), false),
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      iot.DomainConfigurationStatusEnabled,
				ValidateFunc: validation.StringInSlice(iot.DomainConfigurationStatus_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"validation_certificate_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceDomainConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &iot.CreateDomainConfigurationInput{
		DomainConfigurationName: aws.String(name),
	}

	if v, ok := d.GetOk("authorizer_config"); ok {
		input.AuthorizerConfig = expandAuthorizerConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("domain_name"); ok {
		input.DomainName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("server_certificate_arns"); ok {
		input.ServerCertificateArns = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("service_type"); ok {
		input.ServiceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("validation_certificate_arn"); ok {
		input.ValidationCertificateArn = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating IoT Domain Configuration: %s", input)
	output, err := conn.CreateDomainConfigurationWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating IoT Domain Configuration (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.DomainConfigurationName))

	return resourceDomainConfigurationRead(ctx, d, meta)
}

func resourceDomainConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig

	out, err := conn.DescribeDomainConfigurationWithContext(ctx, &iot.DescribeDomainConfigurationInput{
		DomainConfigurationName: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("reading IoT Domain Configuration (%s): %s", d.Id(), err)
	}

	d.Set("arn", out.DomainConfigurationArn)

	tags, _ := ListTags(conn, d.Get("arn").(string))

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceDomainConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn

	if d.HasChange("authorizer_config") {
		input := iot.UpdateDomainConfigurationInput{
			DomainConfigurationName: aws.String(d.Id()),
		}

		if d.HasChange("authorizer_config") {
			input.AuthorizerConfig = expandAuthorizerConfig(d.Get("authorizer_config").([]interface{}))
		}

		log.Printf("[DEBUG] Updating IoT Domain Configuration: %s", d.Id())
		_, err := conn.UpdateDomainConfigurationWithContext(ctx, &input)

		if err != nil {
			return diag.Errorf("updating IoT Domain Configuration (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating IoT Domain Configuration (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceDomainConfigurationRead(ctx, d, meta)
}

func resourceDomainConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn

	if d.Get("status").(string) == iot.DomainConfigurationStatusEnabled {
		log.Printf("[DEBUG] Disabling IoT Domain Configuration: %s", d.Id())
		_, err := conn.UpdateDomainConfigurationWithContext(ctx, &iot.UpdateDomainConfigurationInput{
			DomainConfigurationName:   aws.String(d.Id()),
			DomainConfigurationStatus: aws.String(iot.DomainConfigurationStatusDisabled),
		})

		if err != nil {
			return diag.Errorf("disabling IoT Domain Configuration (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting IoT Domain Configuration: %s", d.Id())
	_, err := conn.DeleteDomainConfigurationWithContext(ctx, &iot.DeleteDomainConfigurationInput{
		DomainConfigurationName: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("deleting IoT Domain Configuration (%s): %s", d.Id(), err)
	}

	return nil
}

func expandAuthorizerConfig(l []interface{}) *iot.AuthorizerConfig {
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
