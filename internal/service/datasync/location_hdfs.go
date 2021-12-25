package datasync

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLocationHdfs() *schema.Resource {
	return &schema.Resource{
		Create: resourceLocationHdfsCreate,
		Read:   resourceLocationHdfsRead,
		Update: resourceLocationHdfsUpdate,
		Delete: resourceLocationHdfsDelete,
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
			"authentication_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(datasync.HdfsAuthenticationType_Values(), false),
			},
			"name_node": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},
					},
				},
			},
			"simple_user": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"subdirectory": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "/",
				ValidateFunc: validation.StringLenBetween(1, 4096),
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

func resourceLocationHdfsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &datasync.CreateLocationHdfsInput{
		AgentArns:          flex.ExpandStringSet(d.Get("agent_arns").(*schema.Set)),
		NameNodes:          expandDataSyncHdfsNameNodes(d.Get("name_node").(*schema.Set)),
		AuthenticationType: aws.String(d.Get("authentication_type").(string)),
		Subdirectory:       aws.String(d.Get("subdirectory").(string)),
		Tags:               Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("simple_user"); ok {
		input.SimpleUser = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating DataSync Location Hdfs: %s", input)
	output, err := conn.CreateLocationHdfs(input)
	if err != nil {
		return fmt.Errorf("error creating DataSync Location Hdfs: %w", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return resourceLocationHdfsRead(d, meta)
}

func resourceLocationHdfsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindLocationHdfsByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Location Hdfs (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Location Hdfs (%s): %w", d.Id(), err)
	}

	subdirectory, err := SubdirectoryFromLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return err
	}

	d.Set("agent_arns", flex.FlattenStringSet(output.AgentArns))
	d.Set("arn", output.LocationArn)
	d.Set("simple_user", output.SimpleUser)
	d.Set("authentication_type", output.AuthenticationType)
	d.Set("uri", output.LocationUri)

	if err := d.Set("name_node", flattenDataSyncHdfsNameNodes(output.NameNodes)); err != nil {
		return fmt.Errorf("error setting name_node: %w", err)
	}

	d.Set("subdirectory", subdirectory)

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for DataSync Location Hdfs (%s): %w", d.Id(), err)
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

func resourceLocationHdfsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	if d.HasChangesExcept("tags_all", "tags") {
		input := &datasync.UpdateLocationHdfsInput{
			LocationArn: aws.String(d.Id()),
		}

		if d.HasChange("authentication_type") {
			input.AuthenticationType = aws.String(d.Get("authentication_type").(string))
		}

		if d.HasChange("subdirectory") {
			input.Subdirectory = aws.String(d.Get("subdirectory").(string))
		}

		if d.HasChange("simple_user") {
			input.SimpleUser = aws.String(d.Get("simple_user").(string))
		}

		if d.HasChange("agent_arns") {
			input.AgentArns = flex.ExpandStringSet(d.Get("agent_arns").(*schema.Set))
		}

		if d.HasChange("name_noode") {
			input.NameNodes = expandDataSyncHdfsNameNodes(d.Get("name_node").(*schema.Set))
		}

		_, err := conn.UpdateLocationHdfs(input)
		if err != nil {
			return fmt.Errorf("error updating DataSync Location Hdfs (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Datasync Hdfs location (%s) tags: %w", d.Id(), err)
		}
	}
	return resourceLocationHdfsRead(d, meta)
}

func resourceLocationHdfsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location Hdfs: %s", input)
	_, err := conn.DeleteLocation(input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Location Hdfs (%s): %w", d.Id(), err)
	}

	return nil
}

func expandDataSyncHdfsNameNodes(l *schema.Set) []*datasync.HdfsNameNode {
	nameNodes := make([]*datasync.HdfsNameNode, 0)
	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		nameNode := &datasync.HdfsNameNode{
			Hostname: aws.String(raw["hostname"].(string)),
			Port:     aws.Int64(int64(raw["port"].(int))),
		}
		nameNodes = append(nameNodes, nameNode)
	}

	return nameNodes
}

func flattenDataSyncHdfsNameNodes(nodes []*datasync.HdfsNameNode) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(nodes))

	for _, raw := range nodes {
		item := make(map[string]interface{})
		item["hostname"] = aws.StringValue(raw.Hostname)
		item["port"] = aws.Int64Value(raw.Port)

		dataResources = append(dataResources, item)
	}

	return dataResources
}
