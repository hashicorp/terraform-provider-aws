package datasync

import (
	"fmt"
	"log"
	"strings"

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

func ResourceLocationNFS() *schema.Resource {
	return &schema.Resource{
		Create: resourceLocationNFSCreate,
		Read:   resourceLocationNFSRead,
		Update: resourceLocationNFSUpdate,
		Delete: resourceLocationNFSDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"on_prem_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"agent_arns": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
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
							Default:      datasync.NfsVersionAutomatic,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(datasync.NfsVersion_Values(), false),
						},
					},
				},
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
				// Ignore missing trailing slash
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "/" {
						return false
					}
					if strings.TrimSuffix(old, "/") == strings.TrimSuffix(new, "/") {
						return true
					}
					return false
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLocationNFSCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &datasync.CreateLocationNfsInput{
		OnPremConfig:   expandOnPremConfig(d.Get("on_prem_config").([]interface{})),
		ServerHostname: aws.String(d.Get("server_hostname").(string)),
		Subdirectory:   aws.String(d.Get("subdirectory").(string)),
		Tags:           Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("mount_options"); ok {
		input.MountOptions = expandNFSMountOptions(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating DataSync Location NFS: %s", input)
	output, err := conn.CreateLocationNfs(input)
	if err != nil {
		return fmt.Errorf("error creating DataSync Location NFS: %w", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return resourceLocationNFSRead(d, meta)
}

func resourceLocationNFSRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &datasync.DescribeLocationNfsInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DataSync Location NFS: %s", input)
	output, err := conn.DescribeLocationNfs(input)

	if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
		log.Printf("[WARN] DataSync Location NFS %q not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Location NFS (%s): %w", d.Id(), err)
	}

	subdirectory, err := SubdirectoryFromLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return err
	}

	d.Set("arn", output.LocationArn)

	if err := d.Set("on_prem_config", flattenOnPremConfig(output.OnPremConfig)); err != nil {
		return fmt.Errorf("error setting on_prem_config: %w", err)
	}

	if err := d.Set("mount_options", flattenNFSMountOptions(output.MountOptions)); err != nil {
		return fmt.Errorf("error setting mount_options: %w", err)
	}

	d.Set("subdirectory", subdirectory)
	d.Set("uri", output.LocationUri)

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for DataSync Location NFS (%s): %w", d.Id(), err)
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

func resourceLocationNFSUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	if d.HasChangesExcept("tags_all", "tags") {
		input := &datasync.UpdateLocationNfsInput{
			LocationArn:  aws.String(d.Id()),
			OnPremConfig: expandOnPremConfig(d.Get("on_prem_config").([]interface{})),
			Subdirectory: aws.String(d.Get("subdirectory").(string)),
		}

		if v, ok := d.GetOk("mount_options"); ok {
			input.MountOptions = expandNFSMountOptions(v.([]interface{}))
		}

		_, err := conn.UpdateLocationNfs(input)
		if err != nil {
			return fmt.Errorf("error updating DataSync Location NFS (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating DataSync Location NFS (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceLocationNFSRead(d, meta)
}

func resourceLocationNFSDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location NFS: %s", input)
	_, err := conn.DeleteLocation(input)

	if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Location NFS (%s): %w", d.Id(), err)
	}

	return nil
}

func expandNFSMountOptions(l []interface{}) *datasync.NfsMountOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	nfsMountOptions := &datasync.NfsMountOptions{
		Version: aws.String(m["version"].(string)),
	}

	return nfsMountOptions
}

func flattenNFSMountOptions(mountOptions *datasync.NfsMountOptions) []interface{} {
	if mountOptions == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"version": aws.StringValue(mountOptions.Version),
	}

	return []interface{}{m}
}

func flattenOnPremConfig(onPremConfig *datasync.OnPremConfig) []interface{} {
	if onPremConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"agent_arns": flex.FlattenStringSet(onPremConfig.AgentArns),
	}

	return []interface{}{m}
}

func expandOnPremConfig(l []interface{}) *datasync.OnPremConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	onPremConfig := &datasync.OnPremConfig{
		AgentArns: flex.ExpandStringSet(m["agent_arns"].(*schema.Set)),
	}

	return onPremConfig
}
