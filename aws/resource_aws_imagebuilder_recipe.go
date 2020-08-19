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
)

func resourceAwsImageBuilderRecipe() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsImageBuilderRecipeCreate,
		Read:   resourceAwsImageBuilderRecipeRead,
		Update: resourceAwsImageBuilderRecipeUpdate,
		Delete: resourceAwsImageBuilderRecipeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"block_device_mappings": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"ebs": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delete_on_termination": {
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
										Default:  true,
									},
									"encrypted": {
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
										Default:  false,
									},
									"iops": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(100, 10000),
									},
									"kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validateArn,
									},
									"snapshot_id": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"volume_size": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 16000),
									},
									"volume_type": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.StringInSlice([]string{
											imagebuilder.EbsVolumeTypeStandard,
											imagebuilder.EbsVolumeTypeIo1,
											imagebuilder.EbsVolumeTypeGp2,
											imagebuilder.EbsVolumeTypeSc1,
											imagebuilder.EbsVolumeTypeSt1,
										}, true),
									},
								},
							},
						},
						"no_device": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"virtual_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
					},
				},
			},
			"components": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"datecreated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 126),
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_image": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 126),
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"semantic_version": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsImageBuilderRecipeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.CreateImageRecipeInput{
		ClientToken:         aws.String(resource.UniqueId()),
		Components:          expandImageBuilderComponentConfig(d.Get("components").([]interface{})),
		Name:                aws.String(d.Get("name").(string)),
		ParentImage:         aws.String(d.Get("parent_image").(string)),
		SemanticVersion:     aws.String(d.Get("semantic_version").(string)),
		BlockDeviceMappings: expandImageBuilderBlockDevices(d),
	}

	if v, ok := d.GetOk("description"); ok {
		input.SetDescription(v.(string))
	}
	if v, ok := d.GetOk("tags"); ok {
		input.SetTags(keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().ImagebuilderTags())
	}

	log.Printf("[DEBUG] Creating Recipe: %#v", input)

	resp, err := conn.CreateImageRecipe(input)
	if err != nil {
		return fmt.Errorf("error creating Recipe: %s", err)
	}

	d.SetId(aws.StringValue(resp.ImageRecipeArn))

	return resourceAwsImageBuilderRecipeRead(d, meta)
}

func resourceAwsImageBuilderRecipeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.GetImageRecipe(&imagebuilder.GetImageRecipeInput{
		ImageRecipeArn: aws.String(d.Id()),
	})

	if isAWSErr(err, imagebuilder.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Recipe (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Recipe (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.ImageRecipe.Arn)
	d.Set("block_device_mappings", resp.ImageRecipe.BlockDeviceMappings)
	d.Set("components", resp.ImageRecipe.Components)
	d.Set("description", resp.ImageRecipe.Description)
	d.Set("name", resp.ImageRecipe.Name)
	d.Set("owner", resp.ImageRecipe.Owner)
	d.Set("parent_image", resp.ImageRecipe.ParentImage)
	d.Set("platform", resp.ImageRecipe.Platform)
	d.Set("semantic_version", resp.ImageRecipe.Version)

	tags, err := keyvaluetags.ImagebuilderListTags(conn, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("error listing tags for Recipe (%s): %s", d.Id(), err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsImageBuilderRecipeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	// tags are the only thing we can update!
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.ImagebuilderUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags for Recipe (%s): %s", d.Id(), err)
		}
	}

	return resourceAwsImageBuilderRecipeRead(d, meta)
}

func resourceAwsImageBuilderRecipeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	_, err := conn.DeleteImageRecipe(&imagebuilder.DeleteImageRecipeInput{
		ImageRecipeArn: aws.String(d.Id()),
	})

	if isAWSErr(err, imagebuilder.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Recipe (%s): %s", d.Id(), err)
	}

	return nil
}

func expandImageBuilderComponentConfig(comps []interface{}) []*imagebuilder.ComponentConfiguration {
	var configs []*imagebuilder.ComponentConfiguration

	for _, line := range comps {
		var arn imagebuilder.ComponentConfiguration
		arn.ComponentArn = aws.String(line.(string))
		configs = append(configs, &arn)
	}

	return configs
}

func expandImageBuilderBlockDevices(d *schema.ResourceData) []*imagebuilder.InstanceBlockDeviceMapping {
	var bdmres []*imagebuilder.InstanceBlockDeviceMapping

	v, ok := d.GetOk("block_device_mappings")
	if !ok {
		return bdmres
	}

	bdmlist := v.([]interface{})

	for _, bdm := range bdmlist {
		if bdm == nil {
			continue
		}
		blockDeviceMapping := readIBBlockDeviceMappingFromConfig(bdm.(map[string]interface{}))
		bdmres = append(bdmres, blockDeviceMapping)
	}

	return bdmres
}

func readIBBlockDeviceMappingFromConfig(bdm map[string]interface{}) *imagebuilder.InstanceBlockDeviceMapping {
	blockDeviceMapping := &imagebuilder.InstanceBlockDeviceMapping{}

	if v := bdm["device_name"].(string); v != "" {
		blockDeviceMapping.DeviceName = aws.String(v)
	}

	if v := bdm["no_device"].(string); v != "" {
		blockDeviceMapping.NoDevice = aws.String(v)
	}

	if v := bdm["virtual_name"].(string); v != "" {
		blockDeviceMapping.VirtualName = aws.String(v)
	}

	if v := bdm["ebs"]; len(v.([]interface{})) > 0 {
		ebs := v.([]interface{})
		if len(ebs) > 0 && ebs[0] != nil {
			ebsData := ebs[0].(map[string]interface{})
			imagebuilderEbsBlockDeviceRequest := readIBEbsBlockDeviceFromConfig(ebsData)
			blockDeviceMapping.Ebs = imagebuilderEbsBlockDeviceRequest
		}
	}

	return blockDeviceMapping
}

func readIBEbsBlockDeviceFromConfig(ebs map[string]interface{}) *imagebuilder.EbsInstanceBlockDeviceSpecification {
	ebsDevice := &imagebuilder.EbsInstanceBlockDeviceSpecification{}

	if v, ok := ebs["delete_on_termination"]; ok {
		ebsDevice.DeleteOnTermination = aws.Bool(v.(bool))
	}

	if v, ok := ebs["encrypted"]; ok {
		ebsDevice.Encrypted = aws.Bool(v.(bool))
	}

	if v := ebs["iops"].(int); v > 0 {
		ebsDevice.Iops = aws.Int64(int64(v))
	}

	if v := ebs["kms_key_id"].(string); v != "" {
		ebsDevice.KmsKeyId = aws.String(v)
	}

	if v := ebs["snapshot_id"].(string); v != "" {
		ebsDevice.SnapshotId = aws.String(v)
	}

	if v := ebs["volume_size"]; v != nil {
		ebsDevice.VolumeSize = aws.Int64(int64(v.(int)))
	}

	if v := ebs["volume_type"].(string); v != "" {
		ebsDevice.VolumeType = aws.String(v)
	}

	return ebsDevice
}
