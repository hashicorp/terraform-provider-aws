// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datasync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_datasync_location_hdfs", name="Location HDFS")
// @Tags(identifierAttribute="id")
func resourceLocationHDFS() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLocationHDFSCreate,
		ReadWithoutTimeout:   resourceLocationHDFSRead,
		UpdateWithoutTimeout: resourceLocationHDFSUpdate,
		DeleteWithoutTimeout: resourceLocationHDFSDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"agent_arns": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.HdfsAuthenticationType](),
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
			"kerberos_keytab": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"kerberos_keytab_base64"},
			},
			"kerberos_keytab_base64": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"kerberos_keytab"},
				ValidateFunc:  verify.ValidBase64String,
			},
			"kerberos_krb5_conf": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"kerberos_krb5_conf_base64"},
			},
			"kerberos_krb5_conf_base64": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"kerberos_krb5_conf"},
				ValidateFunc:  verify.ValidBase64String,
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
						names.AttrPort: {
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
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_transfer_protection": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.HdfsDataTransferProtection](),
						},
						"rpc_protection": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.HdfsRpcProtection](),
						},
					},
				},
			},
			"replication_factor": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3,
				ValidateFunc: validation.IntBetween(1, 512),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrURI: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLocationHDFSCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	input := &datasync.CreateLocationHdfsInput{
		AgentArns:          flex.ExpandStringValueSet(d.Get("agent_arns").(*schema.Set)),
		AuthenticationType: awstypes.HdfsAuthenticationType(d.Get("authentication_type").(string)),
		NameNodes:          expandHDFSNameNodes(d.Get("name_node").(*schema.Set)),
		Subdirectory:       aws.String(d.Get("subdirectory").(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("block_size"); ok {
		input.BlockSize = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("kerberos_keytab"); ok {
		input.KerberosKeytab = []byte(v.(string))
	} else if v, ok := d.GetOk("kerberos_keytab_base64"); ok {
		v := v.(string)
		b, err := itypes.Base64Decode(v)
		if err != nil {
			b = []byte(v)
		}
		input.KerberosKeytab = b
	}

	if v, ok := d.GetOk("kerberos_krb5_conf"); ok {
		input.KerberosKrb5Conf = []byte(v.(string))
	} else if v, ok := d.GetOk("kerberos_krb5_conf_base64"); ok {
		v := v.(string)
		b, err := itypes.Base64Decode(v)
		if err != nil {
			b = []byte(v)
		}
		input.KerberosKrb5Conf = b
	}

	if v, ok := d.GetOk("kerberos_principal"); ok {
		input.KerberosPrincipal = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_provider_uri"); ok {
		input.KmsKeyProviderUri = aws.String(v.(string))
	}

	if v, ok := d.GetOk("qop_configuration"); ok && len(v.([]interface{})) > 0 {
		input.QopConfiguration = expandHDFSQOPConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("replication_factor"); ok {
		input.ReplicationFactor = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("simple_user"); ok {
		input.SimpleUser = aws.String(v.(string))
	}

	output, err := conn.CreateLocationHdfs(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Location HDFS: %s", err)
	}

	d.SetId(aws.ToString(output.LocationArn))

	return append(diags, resourceLocationHDFSRead(ctx, d, meta)...)
}

func resourceLocationHDFSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	output, err := findLocationHDFSByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Location HDFS (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location HDFS (%s): %s", d.Id(), err)
	}

	uri := aws.ToString(output.LocationUri)
	subdirectory, err := subdirectoryFromLocationURI(uri)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("agent_arns", output.AgentArns)
	d.Set(names.AttrARN, output.LocationArn)
	d.Set("authentication_type", output.AuthenticationType)
	d.Set("block_size", output.BlockSize)
	d.Set("kerberos_principal", output.KerberosPrincipal)
	d.Set("kms_key_provider_uri", output.KmsKeyProviderUri)
	if err := d.Set("name_node", flattenHDFSNameNodes(output.NameNodes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting name_node: %s", err)
	}
	if err := d.Set("qop_configuration", flattenHDFSQOPConfiguration(output.QopConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting qop_configuration: %s", err)
	}
	d.Set("replication_factor", output.ReplicationFactor)
	d.Set("simple_user", output.SimpleUser)
	d.Set("subdirectory", subdirectory)
	d.Set(names.AttrURI, uri)

	return diags
}

func resourceLocationHDFSUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &datasync.UpdateLocationHdfsInput{
			LocationArn: aws.String(d.Id()),
		}

		if d.HasChange("agent_arns") {
			input.AgentArns = flex.ExpandStringValueSet(d.Get("agent_arns").(*schema.Set))
		}

		if d.HasChange("authentication_type") {
			input.AuthenticationType = awstypes.HdfsAuthenticationType(d.Get("authentication_type").(string))
		}

		if d.HasChange("block_size") {
			input.BlockSize = aws.Int32(int32(d.Get("block_size").(int)))
		}

		if d.HasChanges("kerberos_keytab", "kerberos_keytab_base64") {
			if v, ok := d.GetOk("kerberos_keytab"); ok {
				input.KerberosKeytab = []byte(v.(string))
			} else if v, ok := d.GetOk("kerberos_keytab_base64"); ok {
				v := v.(string)
				b, err := itypes.Base64Decode(v)
				if err != nil {
					b = []byte(v)
				}
				input.KerberosKeytab = b
			}
		}

		if d.HasChanges("kerberos_krb5_conf", "kerberos_krb5_conf_base64") {
			if v, ok := d.GetOk("kerberos_krb5_conf"); ok {
				input.KerberosKrb5Conf = []byte(v.(string))
			} else if v, ok := d.GetOk("kerberos_krb5_conf_base64"); ok {
				v := v.(string)
				b, err := itypes.Base64Decode(v)
				if err != nil {
					b = []byte(v)
				}
				input.KerberosKrb5Conf = b
			}
		}

		if d.HasChange("kerberos_principal") {
			input.KerberosPrincipal = aws.String(d.Get("kerberos_principal").(string))
		}

		if d.HasChange("kms_key_provider_uri") {
			input.KmsKeyProviderUri = aws.String(d.Get("kms_key_provider_uri").(string))
		}

		if d.HasChange("name_node") {
			input.NameNodes = expandHDFSNameNodes(d.Get("name_node").(*schema.Set))
		}

		if d.HasChange("qop_configuration") {
			input.QopConfiguration = expandHDFSQOPConfiguration(d.Get("qop_configuration").([]interface{}))
		}

		if d.HasChange("replication_factor") {
			input.ReplicationFactor = aws.Int32(int32(d.Get("replication_factor").(int)))
		}

		if d.HasChange("simple_user") {
			input.SimpleUser = aws.String(d.Get("simple_user").(string))
		}

		if d.HasChange("subdirectory") {
			input.Subdirectory = aws.String(d.Get("subdirectory").(string))
		}

		_, err := conn.UpdateLocationHdfs(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Location HDFS (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLocationHDFSRead(ctx, d, meta)...)
}

func resourceLocationHDFSDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	log.Printf("[DEBUG] Deleting DataSync Location HDFS: %s", d.Id())
	_, err := conn.DeleteLocation(ctx, &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Location HDFS (%s): %s", d.Id(), err)
	}

	return diags
}

func findLocationHDFSByARN(ctx context.Context, conn *datasync.Client, arn string) (*datasync.DescribeLocationHdfsOutput, error) {
	input := &datasync.DescribeLocationHdfsInput{
		LocationArn: aws.String(arn),
	}

	output, err := conn.DescribeLocationHdfs(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandHDFSNameNodes(l *schema.Set) []awstypes.HdfsNameNode {
	nameNodes := make([]awstypes.HdfsNameNode, 0)
	for _, m := range l.List() {
		raw := m.(map[string]interface{})
		nameNode := awstypes.HdfsNameNode{
			Hostname: aws.String(raw["hostname"].(string)),
			Port:     aws.Int32(int32(raw[names.AttrPort].(int))),
		}
		nameNodes = append(nameNodes, nameNode)
	}

	return nameNodes
}

func flattenHDFSNameNodes(nodes []awstypes.HdfsNameNode) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(nodes))

	for _, raw := range nodes {
		item := make(map[string]interface{})
		item["hostname"] = aws.ToString(raw.Hostname)
		item[names.AttrPort] = aws.ToInt32(raw.Port)

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func expandHDFSQOPConfiguration(l []interface{}) *awstypes.QopConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	qopConfig := &awstypes.QopConfiguration{
		DataTransferProtection: awstypes.HdfsDataTransferProtection(m["data_transfer_protection"].(string)),
		RpcProtection:          awstypes.HdfsRpcProtection(m["rpc_protection"].(string)),
	}

	return qopConfig
}

func flattenHDFSQOPConfiguration(qopConfig *awstypes.QopConfiguration) []interface{} {
	if qopConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"data_transfer_protection": string(qopConfig.DataTransferProtection),
		"rpc_protection":           string(qopConfig.RpcProtection),
	}

	return []interface{}{m}
}
