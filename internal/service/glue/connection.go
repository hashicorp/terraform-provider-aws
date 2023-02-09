package glue

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"catalog_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"connection_properties": {
				Type:         schema.TypeMap,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: mapKeyInSlice(glue.ConnectionPropertyKey_Values(), false),
				Elem:         &schema.Schema{Type: schema.TypeString},
			},
			"connection_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      glue.ConnectionTypeJdbc,
				ValidateFunc: validation.StringInSlice(glue.ConnectionType_Values(), false),
			},
			"description": {
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
			"name": {
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
						"availability_zone": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"security_group_id_list": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var catalogID string
	if v, ok := d.GetOkExists("catalog_id"); ok {
		catalogID = v.(string)
	} else {
		catalogID = meta.(*conns.AWSClient).AccountID
	}
	name := d.Get("name").(string)

	input := &glue.CreateConnectionInput{
		CatalogId:       aws.String(catalogID),
		ConnectionInput: expandConnectionInput(d),
		Tags:            Tags(tags.IgnoreAWS()),
	}

	log.Printf("[DEBUG] Creating Glue Connection: %s", input)
	_, err := conn.CreateConnectionWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Connection (%s): %s", name, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", catalogID, name))

	return append(diags, resourceConnectionRead(ctx, d, meta)...)
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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
	d.Set("arn", connectionArn)

	d.Set("catalog_id", catalogID)
	if err := d.Set("connection_properties", aws.StringValueMap(connection.ConnectionProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting connection_properties: %s", err)
	}
	d.Set("connection_type", connection.ConnectionType)
	d.Set("description", connection.Description)
	if err := d.Set("match_criteria", flex.FlattenStringList(connection.MatchCriteria)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting match_criteria: %s", err)
	}
	d.Set("name", connection.Name)
	if err := d.Set("physical_connection_requirements", flattenPhysicalConnectionRequirements(connection.PhysicalConnectionRequirements)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting physical_connection_requirements: %s", err)
	}

	tags, err := ListTags(ctx, conn, connectionArn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Glue Connection (%s): %s", connectionArn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()

	if d.HasChangesExcept("tags", "tags_all") {
		catalogID, connectionName, err := DecodeConnectionID(d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Connection (%s): %s", d.Id(), err)
		}

		input := &glue.UpdateConnectionInput{
			CatalogId:       aws.String(catalogID),
			ConnectionInput: expandConnectionInput(d),
			Name:            aws.String(connectionName),
		}

		log.Printf("[DEBUG] Updating Glue Connection: %s", input)
		_, err = conn.UpdateConnectionWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Connection (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return diags
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()

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

func DeleteConnection(ctx context.Context, conn *glue.Glue, catalogID, connectionName string) error {
	input := &glue.DeleteConnectionInput{
		CatalogId:      aws.String(catalogID),
		ConnectionName: aws.String(connectionName),
	}

	_, err := conn.DeleteConnectionWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return nil
		}
		return err
	}

	return nil
}

func expandConnectionInput(d *schema.ResourceData) *glue.ConnectionInput {
	connectionProperties := make(map[string]string)
	if val, ok := d.GetOkExists("connection_properties"); ok {
		for k, v := range val.(map[string]interface{}) {
			connectionProperties[k] = v.(string)
		}
	}

	connectionInput := &glue.ConnectionInput{
		ConnectionProperties: aws.StringMap(connectionProperties),
		ConnectionType:       aws.String(d.Get("connection_type").(string)),
		Name:                 aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		connectionInput.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("match_criteria"); ok {
		connectionInput.MatchCriteria = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("physical_connection_requirements"); ok {
		physicalConnectionRequirementsList := v.([]interface{})
		physicalConnectionRequirementsMap := physicalConnectionRequirementsList[0].(map[string]interface{})
		connectionInput.PhysicalConnectionRequirements = expandPhysicalConnectionRequirements(physicalConnectionRequirementsMap)
	}

	return connectionInput
}

func expandPhysicalConnectionRequirements(m map[string]interface{}) *glue.PhysicalConnectionRequirements {
	physicalConnectionRequirements := &glue.PhysicalConnectionRequirements{}

	if v, ok := m["availability_zone"]; ok {
		physicalConnectionRequirements.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := m["security_group_id_list"]; ok {
		physicalConnectionRequirements.SecurityGroupIdList = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := m["subnet_id"]; ok {
		physicalConnectionRequirements.SubnetId = aws.String(v.(string))
	}

	return physicalConnectionRequirements
}

func flattenPhysicalConnectionRequirements(physicalConnectionRequirements *glue.PhysicalConnectionRequirements) []map[string]interface{} {
	if physicalConnectionRequirements == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"availability_zone":      aws.StringValue(physicalConnectionRequirements.AvailabilityZone),
		"security_group_id_list": flex.FlattenStringSet(physicalConnectionRequirements.SecurityGroupIdList),
		"subnet_id":              aws.StringValue(physicalConnectionRequirements.SubnetId),
	}

	return []map[string]interface{}{m}
}
