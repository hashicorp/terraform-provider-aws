package datasync

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLocationSMB() *schema.Resource {
	return &schema.Resource{
		Create: resourceLocationSMBCreate,
		Read:   resourceLocationSMBRead,
		Update: resourceLocationSMBUpdate,
		Delete: resourceLocationSMBDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"agent_arns": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"domain": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 253),
			},
			"mount_options": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:         schema.TypeString,
							Default:      datasync.SmbVersionAutomatic,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(datasync.SmbVersion_Values(), false),
						},
					},
				},
			},
			"password": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 104),
			},
			"server_hostname": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"subdirectory": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 4096),
				/*// Ignore missing trailing slash
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "/" {
						return false
					}
					if strings.TrimSuffix(old, "/") == strings.TrimSuffix(new, "/") {
						return true
					}
					return false
				},
				*/
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 104),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLocationSMBCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &datasync.CreateLocationSmbInput{
		AgentArns:      flex.ExpandStringSet(d.Get("agent_arns").(*schema.Set)),
		MountOptions:   expandSMBMountOptions(d.Get("mount_options").([]interface{})),
		Password:       aws.String(d.Get("password").(string)),
		ServerHostname: aws.String(d.Get("server_hostname").(string)),
		Subdirectory:   aws.String(d.Get("subdirectory").(string)),
		Tags:           Tags(tags.IgnoreAWS()),
		User:           aws.String(d.Get("user").(string)),
	}

	if v, ok := d.GetOk("domain"); ok {
		input.Domain = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating DataSync Location SMB: %s", input)
	output, err := conn.CreateLocationSmb(input)
	if err != nil {
		return fmt.Errorf("error creating DataSync Location SMB: %w", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return resourceLocationSMBRead(d, meta)
}

func resourceLocationSMBRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &datasync.DescribeLocationSmbInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DataSync Location SMB: %s", input)
	output, err := conn.DescribeLocationSmb(input)

	if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
		log.Printf("[WARN] DataSync Location SMB %q not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Location SMB (%s): %w", d.Id(), err)
	}

	tagsInput := &datasync.ListTagsForResourceInput{
		ResourceArn: output.LocationArn,
	}

	log.Printf("[DEBUG] Reading DataSync Location SMB tags: %s", tagsInput)
	tagsOutput, err := conn.ListTagsForResource(tagsInput)

	if err != nil {
		return fmt.Errorf("error reading DataSync Location SMB (%s) tags: %w", d.Id(), err)
	}

	subdirectory, err := SubdirectoryFromLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return err
	}

	d.Set("agent_arns", flex.FlattenStringSet(output.AgentArns))

	d.Set("arn", output.LocationArn)

	d.Set("domain", output.Domain)

	if err := d.Set("mount_options", flattenSMBMountOptions(output.MountOptions)); err != nil {
		return fmt.Errorf("error setting mount_options: %w", err)
	}

	d.Set("subdirectory", subdirectory)

	tags := KeyValueTags(tagsOutput.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("user", output.User)

	d.Set("uri", output.LocationUri)

	return nil
}

func resourceLocationSMBUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	if d.HasChangesExcept("tags_all", "tags") {
		input := &datasync.UpdateLocationSmbInput{
			LocationArn:  aws.String(d.Id()),
			AgentArns:    flex.ExpandStringSet(d.Get("agent_arns").(*schema.Set)),
			MountOptions: expandSMBMountOptions(d.Get("mount_options").([]interface{})),
			Password:     aws.String(d.Get("password").(string)),
			Subdirectory: aws.String(d.Get("subdirectory").(string)),
			User:         aws.String(d.Get("user").(string)),
		}

		if v, ok := d.GetOk("domain"); ok {
			input.Domain = aws.String(v.(string))
		}

		_, err := conn.UpdateLocationSmb(input)
		if err != nil {
			return fmt.Errorf("error updating DataSync Location SMB (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating DataSync SMB location (%s) tags: %w", d.Id(), err)
		}
	}
	return resourceLocationSMBRead(d, meta)
}

func resourceLocationSMBDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location SMB: %s", input)
	_, err := conn.DeleteLocation(input)

	if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Location SMB (%s): %w", d.Id(), err)
	}

	return nil
}

func flattenSMBMountOptions(mountOptions *datasync.SmbMountOptions) []interface{} {
	if mountOptions == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"version": aws.StringValue(mountOptions.Version),
	}

	return []interface{}{m}
}

func expandSMBMountOptions(l []interface{}) *datasync.SmbMountOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	smbMountOptions := &datasync.SmbMountOptions{
		Version: aws.String(m["version"].(string)),
	}

	return smbMountOptions
}
