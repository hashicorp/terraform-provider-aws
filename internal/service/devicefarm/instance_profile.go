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

func ResourceInstanceProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstanceProfileCreate,
		Read:   resourceInstanceProfileRead,
		Update: resourceInstanceProfileUpdate,
		Delete: resourceInstanceProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
			"exclude_app_packages_from_cleanup": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"package_cleanup": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"reboot_after_use": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceInstanceProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &devicefarm.CreateInstanceProfileInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("exclude_app_packages_from_cleanup"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludeAppPackagesFromCleanup = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("package_cleanup"); ok {
		input.PackageCleanup = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("reboot_after_use"); ok {
		input.RebootAfterUse = aws.Bool(v.(bool))
	}

	out, err := conn.CreateInstanceProfile(input)
	if err != nil {
		return fmt.Errorf("Error creating DeviceFarm Instance Profile: %w", err)
	}

	arn := aws.StringValue(out.InstanceProfile.Arn)
	log.Printf("[DEBUG] Successsfully Created DeviceFarm Instance Profile: %s", arn)
	d.SetId(arn)

	if len(tags) > 0 {
		if err := UpdateTags(conn, arn, nil, tags); err != nil {
			return fmt.Errorf("error updating DeviceFarm Instance Profile (%s) tags: %w", arn, err)
		}
	}

	return resourceInstanceProfileRead(d, meta)
}

func resourceInstanceProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instaceProf, err := FindInstanceProfileByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm Instance Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DeviceFarm Instance Profile (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(instaceProf.Arn)
	d.Set("arn", arn)
	d.Set("name", instaceProf.Name)
	d.Set("description", instaceProf.Description)
	d.Set("exclude_app_packages_from_cleanup", flex.FlattenStringSet(instaceProf.ExcludeAppPackagesFromCleanup))
	d.Set("package_cleanup", instaceProf.PackageCleanup)
	d.Set("reboot_after_use", instaceProf.RebootAfterUse)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for DeviceFarm Instance Profile (%s): %w", arn, err)
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

func resourceInstanceProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &devicefarm.UpdateInstanceProfileInput{
			Arn: aws.String(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("exclude_app_packages_from_cleanup") {
			input.ExcludeAppPackagesFromCleanup = flex.ExpandStringSet(d.Get("exclude_app_packages_from_cleanup").(*schema.Set))
		}

		if d.HasChange("package_cleanup") {
			input.PackageCleanup = aws.Bool(d.Get("package_cleanup").(bool))
		}

		if d.HasChange("reboot_after_use") {
			input.RebootAfterUse = aws.Bool(d.Get("reboot_after_use").(bool))
		}

		log.Printf("[DEBUG] Updating DeviceFarm Instance Profile: %s", d.Id())
		_, err := conn.UpdateInstanceProfile(input)
		if err != nil {
			return fmt.Errorf("Error Updating DeviceFarm Instance Profile: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating DeviceFarm Instance Profile (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceInstanceProfileRead(d, meta)
}

func resourceInstanceProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn

	input := &devicefarm.DeleteInstanceProfileInput{
		Arn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DeviceFarm Instance Profile: %s", d.Id())
	_, err := conn.DeleteInstanceProfile(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting DeviceFarm Instance Profile: %w", err)
	}

	return nil
}
