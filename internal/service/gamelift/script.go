package gamelift

import (
	"fmt"
	"log"
	"os"

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
	"github.com/mitchellh/go-homedir"
)

const scriptMutex = `aws_gamelift_script`

func ResourceScript() *schema.Resource {
	return &schema.Resource{
		Create: resourceScriptCreate,
		Read:   resourceScriptRead,
		Update: resourceScriptUpdate,
		Delete: resourceScriptDelete,
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
			"storage_location": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"zip_file", "storage_location"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"object_version": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
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
			"zip_file": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"zip_file", "storage_location"},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceScriptCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := gamelift.CreateScriptInput{
		Name: aws.String(d.Get("name").(string)),
		Tags: Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("storage_location"); ok && len(v.([]interface{})) > 0 {
		input.StorageLocation = expandStorageLocation(v.([]interface{}))
	}

	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}

	if v, ok := d.GetOk("zip_file"); ok {
		conns.GlobalMutexKV.Lock(scriptMutex)
		defer conns.GlobalMutexKV.Unlock(scriptMutex)

		file, err := loadFileContent(v.(string))
		if err != nil {
			return fmt.Errorf("unable to load %q: %w", v.(string), err)
		}
		input.ZipFile = file
	}

	log.Printf("[INFO] Creating GameLift Script: %s", input)
	var out *gamelift.CreateScriptOutput
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error
		out, err = conn.CreateScript(&input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, "GameLift cannot assume the role") ||
				tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, "Provided resource is not accessible") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.CreateScript(&input)
	}
	if err != nil {
		return fmt.Errorf("Error creating GameLift script client: %w", err)
	}

	d.SetId(aws.StringValue(out.Script.ScriptId))

	return resourceScriptRead(d, meta)
}

func resourceScriptRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[INFO] Reading GameLift Script: %s", d.Id())
	script, err := FindScriptByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Script (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading GameLift Script (%s): %w", d.Id(), err)
	}

	d.Set("name", script.Name)
	d.Set("version", script.Version)

	if err := d.Set("storage_location", flattenStorageLocation(script.StorageLocation)); err != nil {
		return fmt.Errorf("error setting storage_location: %w", err)
	}

	arn := aws.StringValue(script.ScriptArn)
	d.Set("arn", arn)
	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Game Lift Script (%s): %w", arn, err)
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

func resourceScriptUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn

	if d.HasChangesExcept("tags", "tags_all") {
		log.Printf("[INFO] Updating GameLift Script: %s", d.Id())
		input := gamelift.UpdateScriptInput{
			ScriptId: aws.String(d.Id()),
			Name:     aws.String(d.Get("name").(string)),
		}

		if d.HasChange("version") {
			if v, ok := d.GetOk("version"); ok {
				input.Version = aws.String(v.(string))
			}
		}

		if d.HasChange("storage_location") {
			if v, ok := d.GetOk("storage_location"); ok {
				input.StorageLocation = expandStorageLocation(v.([]interface{}))
			}
		}

		if d.HasChange("zip_file") {
			if v, ok := d.GetOk("zip_file"); ok {
				conns.GlobalMutexKV.Lock(scriptMutex)
				defer conns.GlobalMutexKV.Unlock(scriptMutex)

				file, err := loadFileContent(v.(string))
				if err != nil {
					return fmt.Errorf("unable to load %q: %w", v.(string), err)
				}
				input.ZipFile = file
			}
		}

		_, err := conn.UpdateScript(&input)
		if err != nil {
			return fmt.Errorf("Error updating GameLift Script: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		arn := d.Get("arn").(string)
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Game Lift Script (%s) tags: %w", arn, err)
		}
	}

	return resourceScriptRead(d, meta)
}

func resourceScriptDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn

	log.Printf("[INFO] Deleting GameLift Script: %s", d.Id())
	_, err := conn.DeleteScript(&gamelift.DeleteScriptInput{
		ScriptId: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting GameLift script: %w", err)
	}

	return nil
}

func flattenStorageLocation(sl *gamelift.S3Location) []interface{} {
	if sl == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"bucket":         aws.StringValue(sl.Bucket),
		"key":            aws.StringValue(sl.Key),
		"role_arn":       aws.StringValue(sl.RoleArn),
		"object_version": aws.StringValue(sl.ObjectVersion),
	}

	return []interface{}{m}
}

// loadFileContent returns contents of a file in a given path
func loadFileContent(v string) ([]byte, error) {
	filename, err := homedir.Expand(v)
	if err != nil {
		return nil, err
	}
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return fileContent, nil
}
