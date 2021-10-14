package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBuild() *schema.Resource {
	return &schema.Resource{
		Create: resourceBuildCreate,
		Read:   resourceBuildRead,
		Update: resourceBuildUpdate,
		Delete: resourceBuildDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"operating_system": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(gamelift.OperatingSystem_Values(), false),
			},
			"storage_location": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"version": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchema(),

			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBuildCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	sl := expandGameliftStorageLocation(d.Get("storage_location").([]interface{}))
	input := gamelift.CreateBuildInput{
		Name:            aws.String(d.Get("name").(string)),
		OperatingSystem: aws.String(d.Get("operating_system").(string)),
		StorageLocation: sl,
		Tags:            tags.IgnoreAws().GameliftTags(),
	}
	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}
	log.Printf("[INFO] Creating Gamelift Build: %s", input)
	var out *gamelift.CreateBuildOutput
	err := resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		out, err = conn.CreateBuild(&input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, "Provided build is not accessible.") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.CreateBuild(&input)
	}
	if err != nil {
		return fmt.Errorf("Error creating Gamelift build client: %s", err)
	}

	d.SetId(aws.StringValue(out.Build.BuildId))

	stateConf := resource.StateChangeConf{
		Pending: []string{gamelift.BuildStatusInitialized},
		Target:  []string{gamelift.BuildStatusReady},
		Timeout: 1 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			out, err := conn.DescribeBuild(&gamelift.DescribeBuildInput{
				BuildId: aws.String(d.Id()),
			})
			if err != nil {
				return 42, "", err
			}

			return out, *out.Build.Status, nil
		},
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return resourceBuildRead(d, meta)
}

func resourceBuildRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[INFO] Reading Gamelift Build: %s", d.Id())
	out, err := conn.DescribeBuild(&gamelift.DescribeBuildInput{
		BuildId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, gamelift.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] Gamelift Build (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	b := out.Build

	d.Set("name", b.Name)
	d.Set("operating_system", b.OperatingSystem)
	d.Set("version", b.Version)

	arn := aws.StringValue(b.BuildArn)
	d.Set("arn", arn)
	tags, err := tftags.GameliftListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Game Lift Build (%s): %s", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceBuildUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn

	log.Printf("[INFO] Updating Gamelift Build: %s", d.Id())
	input := gamelift.UpdateBuildInput{
		BuildId: aws.String(d.Id()),
		Name:    aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}

	_, err := conn.UpdateBuild(&input)
	if err != nil {
		return err
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := tftags.GameliftUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Game Lift Build (%s) tags: %s", arn, err)
		}
	}

	return resourceBuildRead(d, meta)
}

func resourceBuildDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn

	log.Printf("[INFO] Deleting Gamelift Build: %s", d.Id())
	_, err := conn.DeleteBuild(&gamelift.DeleteBuildInput{
		BuildId: aws.String(d.Id()),
	})
	return err
}

func expandGameliftStorageLocation(cfg []interface{}) *gamelift.S3Location {
	loc := cfg[0].(map[string]interface{})
	return &gamelift.S3Location{
		Bucket:  aws.String(loc["bucket"].(string)),
		Key:     aws.String(loc["key"].(string)),
		RoleArn: aws.String(loc["role_arn"].(string)),
	}
}
