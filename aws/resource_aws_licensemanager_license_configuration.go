package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsLicenseManagerLicenseConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLicenseManagerLicenseConfigurationCreate,
		Read:   resourceAwsLicenseManagerLicenseConfigurationRead,
		Update: resourceAwsLicenseManagerLicenseConfigurationUpdate,
		Delete: resourceAwsLicenseManagerLicenseConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"license_count": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"license_count_hard_limit": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"license_counting_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					licensemanager.LicenseCountingTypeVCpu,
					licensemanager.LicenseCountingTypeInstance,
					licensemanager.LicenseCountingTypeCore,
					licensemanager.LicenseCountingTypeSocket,
				}, false),
			},
			"license_rules": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsLicenseManagerLicenseConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).licensemanagerconn

	opts := &licensemanager.CreateLicenseConfigurationInput{
		LicenseCountingType: aws.String(d.Get("license_counting_type").(string)),
		Name:                aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		opts.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("license_count"); ok {
		opts.LicenseCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("license_count_hard_limit"); ok {
		opts.LicenseCountHardLimit = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("license_rules"); ok {
		opts.LicenseRules = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("tags"); ok && len(v.(map[string]interface{})) > 0 {
		opts.Tags = tagsFromMapLicenseManager(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] License Manager license configuration: %s", opts)

	resp, err := conn.CreateLicenseConfiguration(opts)
	if err != nil {
		return fmt.Errorf("Error creating License Manager license configuration: %s", err)
	}
	d.SetId(*resp.LicenseConfigurationArn)
	return resourceAwsLicenseManagerLicenseConfigurationRead(d, meta)
}

func resourceAwsLicenseManagerLicenseConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).licensemanagerconn

	resp, err := conn.GetLicenseConfiguration(&licensemanager.GetLicenseConfigurationInput{
		LicenseConfigurationArn: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, licensemanager.ErrCodeInvalidParameterValueException, "") {
			log.Printf("[WARN] License Manager license configuration (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading License Manager license configuration: %s", err)
	}

	d.Set("description", resp.Description)
	d.Set("license_count", resp.LicenseCount)
	d.Set("license_count_hard_limit", resp.LicenseCountHardLimit)
	d.Set("license_counting_type", resp.LicenseCountingType)
	d.Set("license_rules", flattenStringList(resp.LicenseRules))
	d.Set("name", resp.Name)

	if err := d.Set("tags", tagsToMapLicenseManager(resp.Tags)); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsLicenseManagerLicenseConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).licensemanagerconn

	d.Partial(true)

	if d.HasChange("tags") {
		if err := setTagsLicenseManager(conn, d); err != nil {
			return err
		} else {
			d.SetPartial("tags")
		}
	}

	d.Partial(false)

	opts := &licensemanager.UpdateLicenseConfigurationInput{
		LicenseConfigurationArn: aws.String(d.Id()),
		Name:                    aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		opts.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("license_count"); ok {
		opts.LicenseCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("license_count_hard_limit"); ok {
		opts.LicenseCountHardLimit = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] License Manager license configuration: %s", opts)

	_, err := conn.UpdateLicenseConfiguration(opts)
	if err != nil {
		return fmt.Errorf("Error updating License Manager license configuration: %s", err)
	}
	return resourceAwsLicenseManagerLicenseConfigurationRead(d, meta)
}

func resourceAwsLicenseManagerLicenseConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).licensemanagerconn

	opts := &licensemanager.DeleteLicenseConfigurationInput{
		LicenseConfigurationArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteLicenseConfiguration(opts)
	if err != nil {
		if isAWSErr(err, licensemanager.ErrCodeInvalidParameterValueException, "") {
			log.Printf("[WARN] License Manager license configuration (%s) not found", d.Id())
			return nil
		}
		return fmt.Errorf("Error deleting License Manager license configuration: %s", err)
	}

	return nil
}
