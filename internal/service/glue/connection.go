// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package glue

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_connection", name="Connection")
// @Tags(identifierAttribute="arn")
func resourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionCreate,
		ReadWithoutTimeout:   resourceConnectionRead,
		UpdateWithoutTimeout: resourceConnectionUpdate,
		DeleteWithoutTimeout: resourceConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"athena_properties": {
				Type:      schema.TypeMap,
				Optional:  true,
				Sensitive: true,
				Elem:      &schema.Schema{Type: schema.TypeString},
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

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	var catalogID string
	if v, ok := d.GetOk(names.AttrCatalogID); ok {
		catalogID = v.(string)
	} else {
		catalogID = meta.(*conns.AWSClient).AccountID(ctx)
	}
	name := d.Get(names.AttrName).(string)

	input := glue.CreateConnectionInput{
		CatalogId:       aws.String(catalogID),
		ConnectionInput: expandConnectionInput(d),
		Tags:            getTagsIn(ctx),
	}

	_, err := conn.CreateConnection(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Connection (%s): %s", name, err)
	}

	d.SetId(connectionCreateResourceID(catalogID, name))

	return append(diags, resourceConnectionRead(ctx, d, meta)...)
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.GlueClient(ctx)

	catalogID, connectionName, err := connectionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	connection, err := findConnectionByTwoPartKey(ctx, conn, connectionName, catalogID)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Glue Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Connection (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, connectionARN(ctx, c, connectionName))
	d.Set("athena_properties", connection.AthenaProperties)
	d.Set(names.AttrCatalogID, catalogID)
	d.Set("connection_properties", connection.ConnectionProperties)
	d.Set("connection_type", connection.ConnectionType)
	d.Set(names.AttrDescription, connection.Description)
	d.Set("match_criteria", connection.MatchCriteria)
	d.Set(names.AttrName, connection.Name)
	if err := d.Set("physical_connection_requirements", flattenPhysicalConnectionRequirements(connection.PhysicalConnectionRequirements)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting physical_connection_requirements: %s", err)
	}

	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		catalogID, connectionName, err := connectionParseResourceID(d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Connection (%s): %s", d.Id(), err)
		}

		input := glue.UpdateConnectionInput{
			CatalogId:       aws.String(catalogID),
			ConnectionInput: expandConnectionInput(d),
			Name:            aws.String(connectionName),
		}

		_, err = conn.UpdateConnection(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Connection (%s): %s", d.Id(), err)
		}
	}

	return diags
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, connectionName, err := connectionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Glue Connection: %s", d.Id())
	if err := deleteConnection(ctx, conn, catalogID, connectionName); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Connection (%s): %s", d.Id(), err)
	}

	return diags
}

const connectionResourceIDSeparator = ":"

func connectionCreateResourceID(catalogID, connectionName string) string {
	parts := []string{catalogID, connectionName}
	id := strings.Join(parts, connectionResourceIDSeparator)

	return id
}

func connectionParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, connectionResourceIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected CATALOG-ID%[2]sNAME", id, connectionResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func deleteConnection(ctx context.Context, conn *glue.Client, catalogID, connectionName string) error {
	input := glue.DeleteConnectionInput{
		CatalogId:      aws.String(catalogID),
		ConnectionName: aws.String(connectionName),
	}

	_, err := conn.DeleteConnection(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil
	}

	return err
}

func findConnectionByTwoPartKey(ctx context.Context, conn *glue.Client, name, catalogID string) (*awstypes.Connection, error) {
	input := glue.GetConnectionInput{
		CatalogId: aws.String(catalogID),
		Name:      aws.String(name),
	}
	output, err := conn.GetConnection(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Connection == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Connection, nil
}

func connectionARN(ctx context.Context, c *conns.AWSClient, connectionName string) string {
	return c.RegionalARN(ctx, "glue", "connection/"+connectionName)
}

func expandConnectionInput(d *schema.ResourceData) *awstypes.ConnectionInput {
	apiObject := &awstypes.ConnectionInput{
		ConnectionType: awstypes.ConnectionType(d.Get("connection_type").(string)),
		Name:           aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("athena_properties"); ok && len(v.(map[string]any)) > 0 {
		apiObject.AthenaProperties = flex.ExpandStringValueMap(v.(map[string]any))
	}

	if v, ok := d.GetOk("connection_properties"); ok && len(v.(map[string]any)) > 0 {
		apiObject.ConnectionProperties = flex.ExpandStringValueMap(v.(map[string]any))
	} else {
		apiObject.ConnectionProperties = make(map[string]string)
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		apiObject.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("match_criteria"); ok && len(v.([]any)) > 0 {
		apiObject.MatchCriteria = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := d.GetOk("physical_connection_requirements"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		apiObject.PhysicalConnectionRequirements = expandPhysicalConnectionRequirements(v.([]any)[0].(map[string]any))
	}

	return apiObject
}

func expandPhysicalConnectionRequirements(tfMap map[string]any) *awstypes.PhysicalConnectionRequirements {
	apiObject := &awstypes.PhysicalConnectionRequirements{}

	if v, ok := tfMap[names.AttrAvailabilityZone]; ok {
		apiObject.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := tfMap["security_group_id_list"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIdList = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnetID]; ok {
		apiObject.SubnetId = aws.String(v.(string))
	}

	return apiObject
}

func flattenPhysicalConnectionRequirements(apiObject *awstypes.PhysicalConnectionRequirements) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrAvailabilityZone: aws.ToString(apiObject.AvailabilityZone),
		"security_group_id_list":   apiObject.SecurityGroupIdList,
		names.AttrSubnetID:         aws.ToString(apiObject.SubnetId),
	}

	return []any{tfMap}
}

func connectionPropertyKey_Values() []string {
	return tfslices.AppendUnique(enum.Values[awstypes.ConnectionPropertyKey](), "SparkProperties")
}

func connectionType_Values() []string {
	return tfslices.AppendUnique(enum.Values[awstypes.ConnectionType](), "AZURECOSMOS", "AZURESQL", "BIGQUERY", "DYNAMODB", "OPENSEARCH", "SNOWFLAKE")
}
