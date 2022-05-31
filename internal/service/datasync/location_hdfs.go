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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLocationHDFS() *schema.Resource {
	return &schema.Resource{
		Create: resourceLocationHDFSCreate,
		Read:   resourceLocationHDFSRead,
		Update: resourceLocationHDFSUpdate,
		Delete: resourceLocationHDFSDelete,
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
			"kerberos_keytab": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kerberos_krb5_conf": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kerberos_principal": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"kms_key_provider_uri": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"block_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  128 * 1024 * 1024, // 128 MiB
				ValidateFunc: validation.All(
					validation.IntDivisibleBy(512),
					validation.IntBetween(1048576, 1073741824),
				),
			},
			"replication_factor": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3,
				ValidateFunc: validation.IntBetween(1, 512),
			},
			"name_node": {
				Type:     schema.TypeSet,
				Required: true,
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
			"qop_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_transfer_protection": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(datasync.HdfsDataTransferProtection_Values(), false),
						},
						"rpc_protection": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(datasync.HdfsRpcProtection_Values(), false),
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

func resourceLocationHDFSCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &datasync.CreateLocationHdfsInput{
		AgentArns:          flex.ExpandStringSet(d.Get("agent_arns").(*schema.Set)),
		NameNodes:          expandHDFSNameNodes(d.Get("name_node").(*schema.Set)),
		AuthenticationType: aws.String(d.Get("authentication_type").(string)),
		Subdirectory:       aws.String(d.Get("subdirectory").(string)),
		Tags:               Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("simple_user"); ok {
		input.SimpleUser = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kerberos_krb5_conf"); ok {
		input.KerberosKrb5Conf = []byte(v.(string))
	}

	if v, ok := d.GetOk("kerberos_keytab"); ok {
		input.KerberosKeytab = []byte(v.(string))
	}

	if v, ok := d.GetOk("kerberos_principal"); ok {
		input.KerberosPrincipal = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_provider_uri"); ok {
		input.KmsKeyProviderUri = aws.String(v.(string))
	}

	if v, ok := d.GetOk("block_size"); ok {
		input.BlockSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("replication_factor"); ok {
		input.ReplicationFactor = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("qop_configuration"); ok && len(v.([]interface{})) > 0 {
		input.QopConfiguration = expandHDFSQOPConfiguration(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating DataSync Location HDFS: %s", input)
	output, err := conn.CreateLocationHdfs(input)
	if err != nil {
		return fmt.Errorf("error creating DataSync Location HDFS: %w", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return resourceLocationHDFSRead(d, meta)
}

func resourceLocationHDFSRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindLocationHDFSByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Location HDFS (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Location HDFS (%s): %w", d.Id(), err)
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
	d.Set("block_size", output.BlockSize)
	d.Set("replication_factor", output.ReplicationFactor)
	d.Set("kerberos_principal", output.KerberosPrincipal)
	d.Set("kms_key_provider_uri", output.KmsKeyProviderUri)
	d.Set("subdirectory", subdirectory)

	if err := d.Set("name_node", flattenHDFSNameNodes(output.NameNodes)); err != nil {
		return fmt.Errorf("error setting name_node: %w", err)
	}

	if err := d.Set("qop_configuration", flattenHDFSQOPConfiguration(output.QopConfiguration)); err != nil {
		return fmt.Errorf("error setting qop_configuration: %w", err)
	}

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for DataSync Location HDFS (%s): %w", d.Id(), err)
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

func resourceLocationHDFSUpdate(d *schema.ResourceData, meta interface{}) error {
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

		if d.HasChange("kerberos_keytab") {
			input.KerberosKeytab = []byte(d.Get("kerberos_keytab").(string))
		}

		if d.HasChange("kerberos_krb5_conf") {
			input.KerberosKrb5Conf = []byte(d.Get("kerberos_krb5_conf").(string))
		}

		if d.HasChange("kerberos_principal") {
			input.KerberosPrincipal = aws.String(d.Get("kerberos_principal").(string))
		}

		if d.HasChange("kms_key_provider_uri") {
			input.KmsKeyProviderUri = aws.String(d.Get("kms_key_provider_uri").(string))
		}

		if d.HasChange("block_size") {
			input.BlockSize = aws.Int64(int64(d.Get("block_size").(int)))
		}

		if d.HasChange("replication_factor") {
			input.ReplicationFactor = aws.Int64(int64(d.Get("replication_factor").(int)))
		}

		if d.HasChange("agent_arns") {
			input.AgentArns = flex.ExpandStringSet(d.Get("agent_arns").(*schema.Set))
		}

		if d.HasChange("name_noode") {
			input.NameNodes = expandHDFSNameNodes(d.Get("name_node").(*schema.Set))
		}

		if d.HasChange("qop_configuration") {
			input.QopConfiguration = expandHDFSQOPConfiguration(d.Get("qop_configuration").([]interface{}))
		}

		_, err := conn.UpdateLocationHdfs(input)
		if err != nil {
			return fmt.Errorf("error updating DataSync Location HDFS (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating DataSync HDFS location (%s) tags: %w", d.Id(), err)
		}
	}
	return resourceLocationHDFSRead(d, meta)
}

func resourceLocationHDFSDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location HDFS: %s", input)
	_, err := conn.DeleteLocation(input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Location HDFS (%s): %w", d.Id(), err)
	}

	return nil
}

func expandHDFSNameNodes(l *schema.Set) []*datasync.HdfsNameNode {
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

func flattenHDFSNameNodes(nodes []*datasync.HdfsNameNode) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(nodes))

	for _, raw := range nodes {
		item := make(map[string]interface{})
		item["hostname"] = aws.StringValue(raw.Hostname)
		item["port"] = aws.Int64Value(raw.Port)

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func expandHDFSQOPConfiguration(l []interface{}) *datasync.QopConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	qopConfig := &datasync.QopConfiguration{
		DataTransferProtection: aws.String(m["data_transfer_protection"].(string)),
		RpcProtection:          aws.String(m["rpc_protection"].(string)),
	}

	return qopConfig
}

func flattenHDFSQOPConfiguration(qopConfig *datasync.QopConfiguration) []interface{} {
	if qopConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"data_transfer_protection": aws.StringValue(qopConfig.DataTransferProtection),
		"rpc_protection":           aws.StringValue(qopConfig.RpcProtection),
	}

	return []interface{}{m}
}
