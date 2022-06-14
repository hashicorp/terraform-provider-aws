package gamelift

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBuild() *schema.Resource {
	return &schema.Resource{
		Create: resourceBuildCreate,
		Read:   resourceBuildRead,
		Update: resourceBuildUpdate,
		Delete: resourceBuildDelete,
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
						"object_version": {
							Type:     schema.TypeString,
							Optional: true,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBuildCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := gamelift.CreateBuildInput{
		Name:            aws.String(d.Get("name").(string)),
		OperatingSystem: aws.String(d.Get("operating_system").(string)),
		StorageLocation: expandStorageLocation(d.Get("storage_location").([]interface{})),
		Tags:            Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}

	log.Printf("[INFO] Creating GameLift Build: %s", input)
	var out *gamelift.CreateBuildOutput
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error
		out, err = conn.CreateBuild(&input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, "Provided build is not accessible.") ||
				tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, "GameLift cannot assume the role") {
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
		return fmt.Errorf("Error creating GameLift build client: %w", err)
	}

	d.SetId(aws.StringValue(out.Build.BuildId))

	if _, err := waitBuildReady(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for GameLift Build (%s) to ready: %w", d.Id(), err)
	}

	return resourceBuildRead(d, meta)
}

func resourceBuildRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[INFO] Reading GameLift Build: %s", d.Id())
	build, err := FindBuildByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Build (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading GameLift Build (%s): %w", d.Id(), err)
	}

	d.Set("name", build.Name)
	d.Set("operating_system", build.OperatingSystem)
	d.Set("version", build.Version)

	arn := aws.StringValue(build.BuildArn)
	d.Set("arn", arn)
	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Game Lift Build (%s): %w", arn, err)
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

func resourceBuildUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn

	if d.HasChangesExcept("tags", "tags_all") {
		log.Printf("[INFO] Updating GameLift Build: %s", d.Id())
		input := gamelift.UpdateBuildInput{
			BuildId: aws.String(d.Id()),
			Name:    aws.String(d.Get("name").(string)),
		}
		if v, ok := d.GetOk("version"); ok {
			input.Version = aws.String(v.(string))
		}

		_, err := conn.UpdateBuild(&input)
		if err != nil {
			return fmt.Errorf("Error updating GameLift build client: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		arn := d.Get("arn").(string)
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Game Lift Build (%s) tags: %w", arn, err)
		}
	}

	return resourceBuildRead(d, meta)
}

func resourceBuildDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn

	log.Printf("[INFO] Deleting GameLift Build: %s", d.Id())
	_, err := conn.DeleteBuild(&gamelift.DeleteBuildInput{
		BuildId: aws.String(d.Id()),
	})
	return err
}

func expandStorageLocation(cfg []interface{}) *gamelift.S3Location {
	loc := cfg[0].(map[string]interface{})

	location := &gamelift.S3Location{
		Bucket:  aws.String(loc["bucket"].(string)),
		Key:     aws.String(loc["key"].(string)),
		RoleArn: aws.String(loc["role_arn"].(string)),
	}

	if v, ok := loc["object_version"].(string); ok && v != "" {
		location.ObjectVersion = aws.String(v)
	}

	return location
}
