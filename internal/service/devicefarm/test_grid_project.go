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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTestGridProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceTestGridProjectCreate,
		Read:   resourceTestGridProjectRead,
		Update: resourceTestGridProjectUpdate,
		Delete: resourceTestGridProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Required: true,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Required: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTestGridProjectCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &devicefarm.CreateTestGridProjectInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_config"); ok {
		input.VpcConfig = expandTestGridProjectVPCConfig(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating DeviceFarm Test Grid Project: %s", name)
	out, err := conn.CreateTestGridProject(input)
	if err != nil {
		return fmt.Errorf("Error creating DeviceFarm Test Grid Project: %w", err)
	}

	arn := aws.StringValue(out.TestGridProject.Arn)
	log.Printf("[DEBUG] Successsfully Created DeviceFarm Test Grid Project: %s", arn)
	d.SetId(arn)

	if len(tags) > 0 {
		if err := UpdateTags(conn, arn, nil, tags); err != nil {
			return fmt.Errorf("error updating DeviceFarm Test Grid Project (%s) tags: %w", arn, err)
		}
	}

	return resourceTestGridProjectRead(d, meta)
}

func resourceTestGridProjectRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	project, err := FindTestGridProjectByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm Test Grid Project (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DeviceFarm Test Grid Project (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(project.Arn)
	d.Set("name", project.Name)
	d.Set("arn", arn)
	d.Set("description", project.Description)
	if err := d.Set("vpc_config", flattenTestGridProjectVPCConfig(project.VpcConfig)); err != nil {
		return fmt.Errorf("error setting vpc_config: %w", err)
	}

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for DeviceFarm Test Grid Project (%s): %w", arn, err)
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

func resourceTestGridProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &devicefarm.UpdateTestGridProjectInput{
			ProjectArn: aws.String(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		log.Printf("[DEBUG] Updating DeviceFarm Test Grid Project: %s", d.Id())
		_, err := conn.UpdateTestGridProject(input)
		if err != nil {
			return fmt.Errorf("Error Updating DeviceFarm Test Grid Project: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating DeviceFarm Test Grid Project (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceTestGridProjectRead(d, meta)
}

func resourceTestGridProjectDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn

	input := &devicefarm.DeleteTestGridProjectInput{
		ProjectArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DeviceFarm Test Grid Project: %s", d.Id())
	_, err := conn.DeleteTestGridProject(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting DeviceFarm Test Grid Project: %w", err)
	}

	return nil
}

func expandTestGridProjectVPCConfig(l []interface{}) *devicefarm.TestGridVpcConfig {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &devicefarm.TestGridVpcConfig{
		VpcId:            aws.String(m["vpc_id"].(string)),
		SubnetIds:        flex.ExpandStringSet(m["subnet_ids"].(*schema.Set)),
		SecurityGroupIds: flex.ExpandStringSet(m["security_group_ids"].(*schema.Set)),
	}

	return config
}

func flattenTestGridProjectVPCConfig(conf *devicefarm.TestGridVpcConfig) []interface{} {
	if conf == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"vpc_id":             aws.StringValue(conf.VpcId),
		"subnet_ids":         flex.FlattenStringSet(conf.SubnetIds),
		"security_group_ids": flex.FlattenStringSet(conf.SecurityGroupIds),
	}

	return []interface{}{m}
}
