// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_glue_classifier")
func ResourceClassifier() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClassifierCreate,
		ReadWithoutTimeout:   resourceClassifierRead,
		UpdateWithoutTimeout: resourceClassifierUpdate,
		DeleteWithoutTimeout: resourceClassifierDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				// ForceNew when changing classifier type
				// InvalidInputException: UpdateClassifierRequest can't change the type of the classifier
				if diff.HasChange("csv_classifier") && diff.HasChange("grok_classifier") {
					diff.ForceNew("csv_classifier")
					diff.ForceNew("grok_classifier")
				}
				if diff.HasChange("csv_classifier") && diff.HasChange("json_classifier") {
					diff.ForceNew("csv_classifier")
					diff.ForceNew("json_classifier")
				}
				if diff.HasChange("csv_classifier") && diff.HasChange("xml_classifier") {
					diff.ForceNew("csv_classifier")
					diff.ForceNew("xml_classifier")
				}
				if diff.HasChange("grok_classifier") && diff.HasChange("json_classifier") {
					diff.ForceNew("grok_classifier")
					diff.ForceNew("json_classifier")
				}
				if diff.HasChange("grok_classifier") && diff.HasChange("xml_classifier") {
					diff.ForceNew("grok_classifier")
					diff.ForceNew("xml_classifier")
				}
				if diff.HasChange("json_classifier") && diff.HasChange("xml_classifier") {
					diff.ForceNew("json_classifier")
					diff.ForceNew("xml_classifier")
				}
				return nil
			},
		),

		Schema: map[string]*schema.Schema{
			"csv_classifier": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"grok_classifier", "json_classifier", "xml_classifier"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_single_column": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"contains_header": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(glue.CsvHeaderOption_Values(), false),
						},
						"custom_datatype_configured": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"custom_datatypes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"BINARY",
									"BOOLEAN",
									"DATE",
									"DECIMAL",
									"DOUBLE",
									"FLOAT",
									"INT",
									"LONG",
									"SHORT",
									"STRING",
									"TIMESTAMP",
								}, false),
							},
						},
						"delimiter": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"disable_value_trimming": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"header": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"quote_symbol": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"grok_classifier": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"csv_classifier", "json_classifier", "xml_classifier"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"classification": {
							Type:     schema.TypeString,
							Required: true,
						},
						"custom_patterns": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 16000),
						},
						"grok_pattern": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
					},
				},
			},
			"json_classifier": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"csv_classifier", "grok_classifier", "xml_classifier"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"json_path": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"xml_classifier": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"csv_classifier", "grok_classifier", "json_classifier"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"classification": {
							Type:     schema.TypeString,
							Required: true,
						},
						"row_tag": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceClassifierCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)
	name := d.Get("name").(string)

	input := &glue.CreateClassifierInput{}

	if v, ok := d.GetOk("csv_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.CsvClassifier = expandCSVClassifierCreate(name, m)
	}

	if v, ok := d.GetOk("grok_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.GrokClassifier = expandGrokClassifierCreate(name, m)
	}

	if v, ok := d.GetOk("json_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.JsonClassifier = expandJSONClassifierCreate(name, m)
	}

	if v, ok := d.GetOk("xml_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.XMLClassifier = expandXmlClassifierCreate(name, m)
	}

	log.Printf("[DEBUG] Creating Glue Classifier: %s", input)
	_, err := conn.CreateClassifierWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Classifier (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceClassifierRead(ctx, d, meta)...)
}

func resourceClassifierRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	classifier, err := FindClassifierByName(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Glue Classifier (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Classifier (%s): %s", d.Id(), err)
	}

	if err := d.Set("csv_classifier", flattenCSVClassifier(classifier.CsvClassifier)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting csv_classifier: %s", err)
	}

	if err := d.Set("grok_classifier", flattenGrokClassifier(classifier.GrokClassifier)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting grok_classifier: %s", err)
	}

	if err := d.Set("json_classifier", flattenJSONClassifier(classifier.JsonClassifier)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting json_classifier: %s", err)
	}

	d.Set("name", d.Id())

	if err := d.Set("xml_classifier", flattenXmlClassifier(classifier.XMLClassifier)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting xml_classifier: %s", err)
	}

	return diags
}

func resourceClassifierUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	input := &glue.UpdateClassifierInput{}

	if v, ok := d.GetOk("csv_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.CsvClassifier = expandCSVClassifierUpdate(d.Id(), m)
	}

	if v, ok := d.GetOk("grok_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.GrokClassifier = expandGrokClassifierUpdate(d.Id(), m)
	}

	if v, ok := d.GetOk("json_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.JsonClassifier = expandJSONClassifierUpdate(d.Id(), m)
	}

	if v, ok := d.GetOk("xml_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.XMLClassifier = expandXmlClassifierUpdate(d.Id(), m)
	}

	log.Printf("[DEBUG] Updating Glue Classifier: %s", input)
	_, err := conn.UpdateClassifierWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Glue Classifier (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceClassifierDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	log.Printf("[DEBUG] Deleting Glue Classifier: %s", d.Id())
	err := DeleteClassifier(ctx, conn, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Classifier (%s): %s", d.Id(), err)
	}

	return diags
}

func DeleteClassifier(ctx context.Context, conn *glue.Glue, name string) error {
	input := &glue.DeleteClassifierInput{
		Name: aws.String(name),
	}

	_, err := conn.DeleteClassifierWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return nil
		}
		return err
	}

	return nil
}

func expandCSVClassifierCreate(name string, m map[string]interface{}) *glue.CreateCsvClassifierRequest {
	csvClassifier := &glue.CreateCsvClassifierRequest{
		AllowSingleColumn:    aws.Bool(m["allow_single_column"].(bool)),
		ContainsHeader:       aws.String(m["contains_header"].(string)),
		Delimiter:            aws.String(m["delimiter"].(string)),
		DisableValueTrimming: aws.Bool(m["disable_value_trimming"].(bool)),
		Name:                 aws.String(name),
	}

	if v, ok := m["quote_symbol"].(string); ok && v != "" {
		csvClassifier.QuoteSymbol = aws.String(v)
	}

	if v, ok := m["header"].([]interface{}); ok {
		csvClassifier.Header = flex.ExpandStringList(v)
	}

	if v, ok := m["custom_datatypes"].([]interface{}); ok && len(v) > 0 {
		if confV, confOk := m["custom_datatype_configured"].(bool); confOk {
			csvClassifier.CustomDatatypeConfigured = aws.Bool(confV)
		}
		csvClassifier.CustomDatatypes = flex.ExpandStringList(v)
	}

	return csvClassifier
}

func expandCSVClassifierUpdate(name string, m map[string]interface{}) *glue.UpdateCsvClassifierRequest {
	csvClassifier := &glue.UpdateCsvClassifierRequest{
		AllowSingleColumn:    aws.Bool(m["allow_single_column"].(bool)),
		ContainsHeader:       aws.String(m["contains_header"].(string)),
		Delimiter:            aws.String(m["delimiter"].(string)),
		DisableValueTrimming: aws.Bool(m["disable_value_trimming"].(bool)),
		Name:                 aws.String(name),
	}

	if v, ok := m["quote_symbol"].(string); ok && v != "" {
		csvClassifier.QuoteSymbol = aws.String(v)
	}

	if v, ok := m["header"].([]interface{}); ok {
		csvClassifier.Header = flex.ExpandStringList(v)
	}

	if v, ok := m["custom_datatypes"].([]interface{}); ok && len(v) > 0 {
		if confV, confOk := m["custom_datatype_configured"].(bool); confOk {
			csvClassifier.CustomDatatypeConfigured = aws.Bool(confV)
		}
		csvClassifier.CustomDatatypes = flex.ExpandStringList(v)
	}

	return csvClassifier
}

func expandGrokClassifierCreate(name string, m map[string]interface{}) *glue.CreateGrokClassifierRequest {
	grokClassifier := &glue.CreateGrokClassifierRequest{
		Classification: aws.String(m["classification"].(string)),
		GrokPattern:    aws.String(m["grok_pattern"].(string)),
		Name:           aws.String(name),
	}

	if v, ok := m["custom_patterns"]; ok && v.(string) != "" {
		grokClassifier.CustomPatterns = aws.String(v.(string))
	}

	return grokClassifier
}

func expandGrokClassifierUpdate(name string, m map[string]interface{}) *glue.UpdateGrokClassifierRequest {
	grokClassifier := &glue.UpdateGrokClassifierRequest{
		Classification: aws.String(m["classification"].(string)),
		GrokPattern:    aws.String(m["grok_pattern"].(string)),
		Name:           aws.String(name),
	}

	if v, ok := m["custom_patterns"]; ok && v.(string) != "" {
		grokClassifier.CustomPatterns = aws.String(v.(string))
	}

	return grokClassifier
}

func expandJSONClassifierCreate(name string, m map[string]interface{}) *glue.CreateJsonClassifierRequest {
	jsonClassifier := &glue.CreateJsonClassifierRequest{
		JsonPath: aws.String(m["json_path"].(string)),
		Name:     aws.String(name),
	}

	return jsonClassifier
}

func expandJSONClassifierUpdate(name string, m map[string]interface{}) *glue.UpdateJsonClassifierRequest {
	jsonClassifier := &glue.UpdateJsonClassifierRequest{
		JsonPath: aws.String(m["json_path"].(string)),
		Name:     aws.String(name),
	}

	return jsonClassifier
}

func expandXmlClassifierCreate(name string, m map[string]interface{}) *glue.CreateXMLClassifierRequest {
	xmlClassifier := &glue.CreateXMLClassifierRequest{
		Classification: aws.String(m["classification"].(string)),
		Name:           aws.String(name),
		RowTag:         aws.String(m["row_tag"].(string)),
	}

	return xmlClassifier
}

func expandXmlClassifierUpdate(name string, m map[string]interface{}) *glue.UpdateXMLClassifierRequest {
	xmlClassifier := &glue.UpdateXMLClassifierRequest{
		Classification: aws.String(m["classification"].(string)),
		Name:           aws.String(name),
		RowTag:         aws.String(m["row_tag"].(string)),
	}

	if v, ok := m["row_tag"]; ok && v.(string) != "" {
		xmlClassifier.RowTag = aws.String(v.(string))
	}

	return xmlClassifier
}

func flattenCSVClassifier(csvClassifier *glue.CsvClassifier) []map[string]interface{} {
	if csvClassifier == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"allow_single_column":        aws.BoolValue(csvClassifier.AllowSingleColumn),
		"contains_header":            aws.StringValue(csvClassifier.ContainsHeader),
		"delimiter":                  aws.StringValue(csvClassifier.Delimiter),
		"disable_value_trimming":     aws.BoolValue(csvClassifier.DisableValueTrimming),
		"header":                     aws.StringValueSlice(csvClassifier.Header),
		"quote_symbol":               aws.StringValue(csvClassifier.QuoteSymbol),
		"custom_datatype_configured": aws.BoolValue(csvClassifier.CustomDatatypeConfigured),
		"custom_datatypes":           aws.StringValueSlice(csvClassifier.CustomDatatypes),
	}

	return []map[string]interface{}{m}
}

func flattenGrokClassifier(grokClassifier *glue.GrokClassifier) []map[string]interface{} {
	if grokClassifier == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"classification":  aws.StringValue(grokClassifier.Classification),
		"custom_patterns": aws.StringValue(grokClassifier.CustomPatterns),
		"grok_pattern":    aws.StringValue(grokClassifier.GrokPattern),
	}

	return []map[string]interface{}{m}
}

func flattenJSONClassifier(jsonClassifier *glue.JsonClassifier) []map[string]interface{} {
	if jsonClassifier == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"json_path": aws.StringValue(jsonClassifier.JsonPath),
	}

	return []map[string]interface{}{m}
}

func flattenXmlClassifier(xmlClassifier *glue.XMLClassifier) []map[string]interface{} {
	if xmlClassifier == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"classification": aws.StringValue(xmlClassifier.Classification),
		"row_tag":        aws.StringValue(xmlClassifier.RowTag),
	}

	return []map[string]interface{}{m}
}
