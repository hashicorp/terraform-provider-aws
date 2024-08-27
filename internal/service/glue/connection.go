// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_connection", name="Connection")
// @Tags(identifierAttribute="arn")
func ResourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionCreate,
		ReadWithoutTimeout:   resourceConnectionRead,
		UpdateWithoutTimeout: resourceConnectionUpdate,
		DeleteWithoutTimeout: resourceConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"connection_properties": {
				Type:             schema.TypeMap,
				Optional:         true,
				Sensitive:        true,
				ValidateDiagFunc: verify.MapKeysAre(validation.ToDiagFunc(validation.StringInSlice(connectionPropertyKey_Values(), false))),
				Elem:             &schema.Schema{Type: schema.TypeString},
			},
			"connection_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      awstypes.ConnectionTypeJdbc,
				ValidateFunc: validation.StringInSlice(connectionType_Values(), false),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"match_criteria": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"physical_connection_requirements": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAvailabilityZone: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"security_group_id_list": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetID: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	var catalogID string
	if v, ok := d.GetOkExists(names.AttrCatalogID); ok {
		catalogID = v.(string)
	} else {
		catalogID = meta.(*conns.AWSClient).AccountID
	}
	name := d.Get(names.AttrName).(string)

	input := &glue.CreateConnectionInput{
		CatalogId:       aws.String(catalogID),
		ConnectionInput: expandConnectionInput(d),
		Tags:            getTagsIn(ctx),
	}

	log.Printf("[DEBUG] Creating Glue Connection: %+v", input)
	_, err := conn.CreateConnection(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Connection (%s): %s", name, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", catalogID, name))

	return append(diags, resourceConnectionRead(ctx, d, meta)...)
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, connectionName, err := DecodeConnectionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Connection (%s): %s", d.Id(), err)
	}

	connection, err := FindConnectionByName(ctx, conn, connectionName, catalogID)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Glue Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Connection (%s): %s", d.Id(), err)
	}

	connectionArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("connection/%s", connectionName),
	}.String()
	d.Set(names.AttrARN, connectionArn)

	d.Set(names.AttrCatalogID, catalogID)
	if err := d.Set("connection_properties", connection.ConnectionProperties); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting connection_properties: %s", err)
	}
	d.Set("connection_type", connection.ConnectionType)
	d.Set(names.AttrDescription, connection.Description)
	if err := d.Set("match_criteria", flex.FlattenStringValueList(connection.MatchCriteria)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting match_criteria: %s", err)
	}
	d.Set(names.AttrName, connection.Name)
	if err := d.Set("physical_connection_requirements", flattenPhysicalConnectionRequirements(connection.PhysicalConnectionRequirements)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting physical_connection_requirements: %s", err)
	}

	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		catalogID, connectionName, err := DecodeConnectionID(d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Connection (%s): %s", d.Id(), err)
		}

		input := &glue.UpdateConnectionInput{
			CatalogId:       aws.String(catalogID),
			ConnectionInput: expandConnectionInput(d),
			Name:            aws.String(connectionName),
		}

		log.Printf("[DEBUG] Updating Glue Connection: %+v", input)
		_, err = conn.UpdateConnection(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Connection (%s): %s", d.Id(), err)
		}
	}

	return diags
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, connectionName, err := DecodeConnectionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Connection (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Glue Connection: %s", d.Id())
	err = DeleteConnection(ctx, conn, catalogID, connectionName)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Connection (%s): %s", d.Id(), err)
	}

	return diags
}

func DecodeConnectionID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format CATALOG-ID:NAME, provided: %s", id)
	}
	return idParts[0], idParts[1], nil
}

func DeleteConnection(ctx context.Context, conn *glue.Client, catalogID, connectionName string) error {
	input := &glue.DeleteConnectionInput{
		CatalogId:      aws.String(catalogID),
		ConnectionName: aws.String(connectionName),
	}

	_, err := conn.DeleteConnection(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return nil
		}
		return err
	}

	return nil
}

func expandConnectionInput(d *schema.ResourceData) *awstypes.ConnectionInput {
	connectionProperties := make(map[string]string)
	if val, ok := d.GetOkExists("connection_properties"); ok {
		for k, v := range val.(map[string]interface{}) {
			connectionProperties[k] = v.(string)
		}
	}

	connectionInput := &awstypes.ConnectionInput{
		ConnectionProperties: connectionProperties,
		ConnectionType:       awstypes.ConnectionType(d.Get("connection_type").(string)),
		Name:                 aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		connectionInput.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("match_criteria"); ok {
		connectionInput.MatchCriteria = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk("physical_connection_requirements"); ok && v.([]interface{})[0] != nil {
		physicalConnectionRequirementsMap := v.([]interface{})[0].(map[string]interface{})
		connectionInput.PhysicalConnectionRequirements = expandPhysicalConnectionRequirements(physicalConnectionRequirementsMap)
	}

	return connectionInput
}

func expandPhysicalConnectionRequirements(m map[string]interface{}) *awstypes.PhysicalConnectionRequirements {
	physicalConnectionRequirements := &awstypes.PhysicalConnectionRequirements{}

	if v, ok := m[names.AttrAvailabilityZone]; ok {
		physicalConnectionRequirements.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := m["security_group_id_list"]; ok {
		physicalConnectionRequirements.SecurityGroupIdList = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := m[names.AttrSubnetID]; ok {
		physicalConnectionRequirements.SubnetId = aws.String(v.(string))
	}

	return physicalConnectionRequirements
}

func flattenPhysicalConnectionRequirements(physicalConnectionRequirements *awstypes.PhysicalConnectionRequirements) []map[string]interface{} {
	if physicalConnectionRequirements == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		names.AttrAvailabilityZone: aws.ToString(physicalConnectionRequirements.AvailabilityZone),
		"security_group_id_list":   flex.FlattenStringValueSet(physicalConnectionRequirements.SecurityGroupIdList),
		names.AttrSubnetID:         aws.ToString(physicalConnectionRequirements.SubnetId),
	}

	return []map[string]interface{}{m}
}

func connectionPropertyKey_Values() []string {
	return tfslices.AppendUnique(enum.Values[awstypes.ConnectionPropertyKey](), "SparkProperties")
}

func connectionType_Values() []string {
	return tfslices.AppendUnique(enum.Values[awstypes.ConnectionType](), "AZURECOSMOS", "AZURESQL", "BIGQUERY", "OPENSEARCH", "SNOWFLAKE")
}
