package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"log"
	"regexp"
)

func resourceAwsImageBuilderDistributionConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsImageBuilderDistributionConfigurationCreate,
		Read:   resourceAwsImageBuilderDistributionConfigurationRead,
		Update: resourceAwsImageBuilderDistributionConfigurationUpdate,
		Delete: resourceAwsImageBuilderDistributionConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"distributions": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ami_distribution_configuration": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ami_tags": tagsSchema(),
									"description": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 1024),
									},
									"kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									"launch_permission": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"user_groups": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringLenBetween(1, 1024),
													},
												},
												"user_ids": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringLenBetween(1, 1024),
													},
												},
											},
										},
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 127),
											validation.StringMatch(regexp.MustCompile(`^[-_A-Za-z0-9{][-_A-Za-z0-9\s:{}]+[-_A-Za-z0-9}]$`), "must contain only alphanumeric characters, periods, underscores, and hyphens"),
										),
									},
								},
							},
						},
						"license_configuration_arn": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"region": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 126),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsImageBuilderDistributionConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.CreateDistributionConfigurationInput{
		ClientToken:   aws.String(resource.UniqueId()),
		Name:          aws.String(d.Get("name").(string)),
		Distributions: expandAwsImageBuilderDistributions(d),
	}

	if v, ok := d.GetOk("description"); ok {
		input.SetDescription(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.SetTags(keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().ImagebuilderTags())
	}

	log.Printf("[DEBUG] Creating DistributionConfiguration: %#v", input)

	resp, err := conn.CreateDistributionConfiguration(input)
	if err != nil {
		return fmt.Errorf("error creating DistributionConfiguration: %s", err)
	}

	d.SetId(aws.StringValue(resp.DistributionConfigurationArn))

	return resourceAwsImageBuilderDistributionConfigurationRead(d, meta)
}

func resourceAwsImageBuilderDistributionConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	resp, err := conn.GetDistributionConfiguration(&imagebuilder.GetDistributionConfigurationInput{
		DistributionConfigurationArn: aws.String(d.Id()),
	})

	if isAWSErr(err, imagebuilder.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] DistributionConfiguration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DistributionConfiguration (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.DistributionConfiguration.Arn)
	d.Set("date_created", resp.DistributionConfiguration.DateCreated)
	d.Set("date_updated", resp.DistributionConfiguration.DateUpdated)
	d.Set("description", resp.DistributionConfiguration.Description)
	d.Set("distributions", flattenAwsImageBuilderDistributions(resp.DistributionConfiguration.Distributions))
	d.Set("name", resp.DistributionConfiguration.Name)

	tags, err := keyvaluetags.ImagebuilderListTags(conn, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("error listing tags for DistributionConfiguration (%s): %s", d.Id(), err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(meta.(*AWSClient).IgnoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsImageBuilderDistributionConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	upd := imagebuilder.UpdateDistributionConfigurationInput{
		DistributionConfigurationArn: aws.String(d.Id()),
		Distributions:                expandAwsImageBuilderDistributions(d),
	}

	if description, ok := d.GetOk("description"); ok {
		upd.Description = aws.String(description.(string))
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.ImagebuilderUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags for DistributionConfiguration (%s): %s", d.Id(), err)
		}
	}

	_, err := conn.UpdateDistributionConfiguration(&upd)
	if err != nil {
		return fmt.Errorf("error updating Distribution Configuration (%s): %s", d.Id(), err)
	}

	return resourceAwsImageBuilderDistributionConfigurationRead(d, meta)
}

func resourceAwsImageBuilderDistributionConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	_, err := conn.DeleteDistributionConfiguration(&imagebuilder.DeleteDistributionConfigurationInput{
		DistributionConfigurationArn: aws.String(d.Id()),
	})

	if isAWSErr(err, imagebuilder.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Distribution Config (%s): %s", d.Id(), err)
	}

	return nil
}

func expandAwsImageBuilderDistributions(d *schema.ResourceData) []*imagebuilder.Distribution {
	configs := d.Get("distributions").(*schema.Set).List()

	if len(configs) == 0 {
		return nil
	}

	distriblist := make([]*imagebuilder.Distribution, 0)

	for _, config := range configs {
		distrib := expandAwsImageBuilderDistribution(config.(map[string]interface{}))
		distriblist = append(distriblist, &distrib)
	}

	return distriblist
}

func flattenAwsImageBuilderDistributions(distribs []*imagebuilder.Distribution) []interface{} {
	if distribs == nil {
		return nil
	}

	distriblist := []interface{}{}

	for _, dist := range distribs {
		distrib := flattenAwsImageBuilderDistribution(dist)
		distriblist = append(distriblist, distrib)
	}

	return distriblist
}

func expandAwsImageBuilderDistribution(data map[string]interface{}) imagebuilder.Distribution {
	distrib := imagebuilder.Distribution{
		Region:                   aws.String(data["region"].(string)),
		LicenseConfigurationArns: expandStringList(data["license_configuration_arn"].(*schema.Set).List()),
	}

	if v, ok := data["ami_distribution_configuration"]; ok {
		if len(v.([]interface{})) > 0 {
			adc := v.([]interface{})[0].(map[string]interface{})

			amidistconfig := imagebuilder.AmiDistributionConfiguration{
				Description: aws.String((adc["description"]).(string)),
				Name:        aws.String(adc["name"].(string)),
			}

			if len(adc["kms_key_id"].(string)) != 0 {
				amidistconfig.KmsKeyId = aws.String(adc["kms_key_id"].(string))
			}

			if len(adc["launch_permission"].([]interface{})) > 0 {

				lp := adc["launch_permission"].([]interface{})[0].(map[string]interface{})

				launchperm := imagebuilder.LaunchPermissionConfiguration{
					UserGroups: aws.StringSlice(sIfTosString(lp["user_groups"].(*schema.Set).List())),
					UserIds:    aws.StringSlice(sIfTosString(lp["user_ids"].(*schema.Set).List())),
				}
				amidistconfig.LaunchPermission = &launchperm
			}

			tags := adc["ami_tags"].(map[string]interface{})
			if len(tags) > 0 {
				amitags := make(map[string]*string, len(tags))
				for k, v := range tags {
					amitags[k] = aws.String(v.(string))
				}
				amidistconfig.AmiTags = amitags
			}
			distrib.AmiDistributionConfiguration = &amidistconfig
		}
	}

	return distrib
}

func flattenAwsImageBuilderDistribution(distribution *imagebuilder.Distribution) map[string]interface{} {
	distrib := map[string]interface{}{}

	distrib["license_configuration_arn"] = schema.NewSet(schema.HashString, flattenStringList(distribution.LicenseConfigurationArns))
	distrib["region"] = *distribution.Region

	distribconf := make(map[string]interface{})
	if distribution.AmiDistributionConfiguration != nil {
		distribconf["ami_tags"] = keyvaluetags.ImagebuilderKeyValueTags(distribution.AmiDistributionConfiguration.AmiTags)
		distribconf["description"] = aws.String(*distribution.AmiDistributionConfiguration.Description)
		distribconf["kms_key_id"] = distribution.AmiDistributionConfiguration.KmsKeyId
		distribconf["name"] = distribution.AmiDistributionConfiguration.Name

		if distribution.AmiDistributionConfiguration.LaunchPermission != nil {
			distribconf["launch_permission"] = []interface{}{
				map[string]interface{}{
					"user_ids":    distribution.AmiDistributionConfiguration.LaunchPermission.UserIds,
					"user_groups": distribution.AmiDistributionConfiguration.LaunchPermission.UserGroups,
				},
			}
		}
	}

	distrib["ami_distribution_configuration"] = []interface{}{distribconf}

	return distrib
}

func sIfTosString(in []interface{}) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = fmt.Sprint(v)
	}

	return out
}
