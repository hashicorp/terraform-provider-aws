package glue

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceMLTransform() *schema.Resource {
	return &schema.Resource{
		Create: resourceMLTransformCreate,
		Read:   resourceMLTransformRead,
		Update: resourceMLTransformUpdate,
		Delete: resourceMLTransformDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"input_record_tables": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"table_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"catalog_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"connection_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"parameters": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"find_matches_parameters": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"accuracy_cost_trade_off": {
										Type:         schema.TypeFloat,
										Optional:     true,
										ValidateFunc: validation.FloatAtMost(1.0),
									},
									"enforce_provided_labels": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"precision_recall_trade_off": {
										Type:         schema.TypeFloat,
										Optional:     true,
										ValidateFunc: validation.FloatAtMost(1.0),
									},
									"primary_key_column_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"transform_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(glue.TransformType_Values(), false),
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"glue_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"max_capacity": {
				Type:          schema.TypeFloat,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"number_of_workers", "worker_type"},
				ValidateFunc:  validation.FloatBetween(2, 100),
			},
			"max_retries": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2880,
			},
			"worker_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"max_capacity"},
				ValidateFunc:  validation.StringInSlice(glue.WorkerType_Values(), false),
				RequiredWith:  []string{"number_of_workers"},
			},
			"number_of_workers": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"max_capacity"},
				ValidateFunc:  validation.IntAtLeast(1),
				RequiredWith:  []string{"worker_type"},
			},
			"label_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"schema": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"data_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceMLTransformCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &glue.CreateMLTransformInput{
		Name:              aws.String(d.Get("name").(string)),
		Role:              aws.String(d.Get("role_arn").(string)),
		Tags:              Tags(tags.IgnoreAWS()),
		Timeout:           aws.Int64(int64(d.Get("timeout").(int))),
		InputRecordTables: expandMLTransformInputRecordTables(d.Get("input_record_tables").([]interface{})),
		Parameters:        expandMLTransformParameters(d.Get("parameters").([]interface{})),
	}

	if v, ok := d.GetOk("max_capacity"); ok {
		input.MaxCapacity = aws.Float64(v.(float64))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("glue_version"); ok {
		input.GlueVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_retries"); ok {
		input.MaxRetries = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("worker_type"); ok {
		input.WorkerType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("number_of_workers"); ok {
		input.NumberOfWorkers = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating Glue ML Transform: %s", input)
	output, err := conn.CreateMLTransform(input)
	if err != nil {
		return fmt.Errorf("error creating Glue ML Transform: %w", err)
	}

	d.SetId(aws.StringValue(output.TransformId))

	return resourceMLTransformRead(d, meta)
}

func resourceMLTransformRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &glue.GetMLTransformInput{
		TransformId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Glue ML Transform: %s", input)
	output, err := conn.GetMLTransform(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			log.Printf("[WARN] Glue ML Transform (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Glue ML Transform (%s): %w", d.Id(), err)
	}

	if output == nil {
		log.Printf("[WARN] Glue ML Transform (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] setting Glue ML Transform: %#v", output)

	mlTransformArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("mlTransform/%s", d.Id()),
	}.String()
	d.Set("arn", mlTransformArn)

	d.Set("description", output.Description)
	d.Set("glue_version", output.GlueVersion)
	d.Set("max_capacity", output.MaxCapacity)
	d.Set("max_retries", output.MaxRetries)
	d.Set("name", output.Name)
	d.Set("role_arn", output.Role)
	d.Set("timeout", output.Timeout)
	d.Set("worker_type", output.WorkerType)
	d.Set("number_of_workers", output.NumberOfWorkers)
	d.Set("label_count", output.LabelCount)

	if err := d.Set("input_record_tables", flattenMLTransformInputRecordTables(output.InputRecordTables)); err != nil {
		return fmt.Errorf("error setting input_record_tables: %w", err)
	}

	if err := d.Set("parameters", flattenMLTransformParameters(output.Parameters)); err != nil {
		return fmt.Errorf("error setting parameters: %w", err)
	}

	if err := d.Set("schema", flattenMLTransformSchemaColumns(output.Schema)); err != nil {
		return fmt.Errorf("error setting schema: %w", err)
	}

	tags, err := ListTags(conn, mlTransformArn)

	if err != nil {
		return fmt.Errorf("error listing tags for Glue ML Transform (%s): %w", mlTransformArn, err)
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

func resourceMLTransformUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	if d.HasChanges("description", "glue_version", "max_capacity", "max_retries", "number_of_workers",
		"role_arn", "timeout", "worker_type", "parameters") {

		input := &glue.UpdateMLTransformInput{
			TransformId: aws.String(d.Id()),
			Role:        aws.String(d.Get("role_arn").(string)),
			Timeout:     aws.Int64(int64(d.Get("timeout").(int))),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("worker_type"); ok {
			input.WorkerType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("max_retries"); ok {
			input.MaxRetries = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("number_of_workers"); ok {
			input.NumberOfWorkers = aws.Int64(int64(v.(int)))
		} else {
			if v, ok := d.GetOk("max_capacity"); ok {
				input.MaxCapacity = aws.Float64(v.(float64))
			}
		}

		if v, ok := d.GetOk("glue_version"); ok {
			input.GlueVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("parameters"); ok {
			input.Parameters = expandMLTransformParameters(v.([]interface{}))
		}

		log.Printf("[DEBUG] Updating Glue ML Transform: %s", input)
		_, err := conn.UpdateMLTransform(input)
		if err != nil {
			return fmt.Errorf("error updating Glue ML Transform (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceMLTransformRead(d, meta)
}

func resourceMLTransformDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	log.Printf("[DEBUG] Deleting Glue ML Trasform: %s", d.Id())

	input := &glue.DeleteMLTransformInput{
		TransformId: aws.String(d.Id()),
	}

	_, err := conn.DeleteMLTransform(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting Glue ML Transform (%s): %w", d.Id(), err)
	}

	if _, err := waitMLTransformDeleted(conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return nil
		}
		return fmt.Errorf("error waiting for Glue ML Transform (%s) to be Deleted: %w", d.Id(), err)
	}

	return nil
}

func expandMLTransformInputRecordTables(l []interface{}) []*glue.Table {
	var tables []*glue.Table

	for _, mRaw := range l {
		m := mRaw.(map[string]interface{})

		table := &glue.Table{}

		if v, ok := m["table_name"].(string); ok {
			table.TableName = aws.String(v)
		}

		if v, ok := m["database_name"].(string); ok {
			table.DatabaseName = aws.String(v)
		}

		if v, ok := m["connection_name"].(string); ok && v != "" {
			table.ConnectionName = aws.String(v)
		}

		if v, ok := m["catalog_id"].(string); ok && v != "" {
			table.CatalogId = aws.String(v)
		}

		tables = append(tables, table)
	}

	return tables
}

func flattenMLTransformInputRecordTables(tables []*glue.Table) []interface{} {
	l := []interface{}{}

	for _, table := range tables {
		m := map[string]interface{}{
			"table_name":    aws.StringValue(table.TableName),
			"database_name": aws.StringValue(table.DatabaseName),
		}

		if table.ConnectionName != nil {
			m["connection_name"] = aws.StringValue(table.ConnectionName)
		}

		if table.CatalogId != nil {
			m["catalog_id"] = aws.StringValue(table.CatalogId)
		}

		l = append(l, m)
	}

	return l
}

func expandMLTransformParameters(l []interface{}) *glue.TransformParameters {
	m := l[0].(map[string]interface{})

	param := &glue.TransformParameters{
		TransformType: aws.String(m["transform_type"].(string)),
	}

	if v, ok := m["find_matches_parameters"]; ok && len(v.([]interface{})) > 0 {
		param.FindMatchesParameters = expandMLTransformFindMatchesParameters(v.([]interface{}))
	}

	return param
}

func flattenMLTransformParameters(parameters *glue.TransformParameters) []map[string]interface{} {
	if parameters == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"transform_type": aws.StringValue(parameters.TransformType),
	}

	if parameters.FindMatchesParameters != nil {
		m["find_matches_parameters"] = flattenMLTransformFindMatchesParameters(parameters.FindMatchesParameters)
	}

	return []map[string]interface{}{m}
}

func expandMLTransformFindMatchesParameters(l []interface{}) *glue.FindMatchesParameters {
	m := l[0].(map[string]interface{})

	param := &glue.FindMatchesParameters{}

	if v, ok := m["accuracy_cost_trade_off"]; ok {
		param.AccuracyCostTradeoff = aws.Float64(v.(float64))
	}

	if v, ok := m["precision_recall_trade_off"]; ok {
		param.PrecisionRecallTradeoff = aws.Float64(v.(float64))
	}

	if v, ok := m["enforce_provided_labels"]; ok {
		param.EnforceProvidedLabels = aws.Bool(v.(bool))
	}

	if v, ok := m["primary_key_column_name"]; ok && v != "" {
		param.PrimaryKeyColumnName = aws.String(v.(string))
	}

	return param
}

func flattenMLTransformFindMatchesParameters(parameters *glue.FindMatchesParameters) []map[string]interface{} {
	if parameters == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if parameters.PrimaryKeyColumnName != nil {
		m["primary_key_column_name"] = aws.StringValue(parameters.PrimaryKeyColumnName)
	}

	if parameters.EnforceProvidedLabels != nil {
		m["enforce_provided_labels"] = aws.BoolValue(parameters.EnforceProvidedLabels)
	}

	if parameters.AccuracyCostTradeoff != nil {
		m["accuracy_cost_trade_off"] = aws.Float64Value(parameters.AccuracyCostTradeoff)
	}

	if parameters.PrimaryKeyColumnName != nil {
		m["precision_recall_trade_off"] = aws.Float64Value(parameters.PrecisionRecallTradeoff)
	}

	return []map[string]interface{}{m}
}

func flattenMLTransformSchemaColumns(schemaCols []*glue.SchemaColumn) []interface{} {
	l := []interface{}{}

	for _, schemaCol := range schemaCols {
		m := map[string]interface{}{
			"name":      aws.StringValue(schemaCol.Name),
			"data_type": aws.StringValue(schemaCol.DataType),
		}

		l = append(l, m)
	}

	return l
}
