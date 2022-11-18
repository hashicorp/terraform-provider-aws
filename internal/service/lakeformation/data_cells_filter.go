package lakeformation

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceDataCellsFilter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDataCellsFilterCreate,
		ReadContext:   resourceDataCellsFilterRead,
		//There is no update api. Create and Deleta only
		DeleteContext: resourceDataCellsFilterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// Don't use timeouts. There is no single data cells filter retrieval api.

		Schema: map[string]*schema.Schema{
			"table_catalog_id": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"database_name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validateLFSingleLineString(),
			},
			"table_name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validateLFSingleLineString(),
			},
			"name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validateLFSingleLineString(),
			},
			"column_names": {
				Type:     schema.TypeSet,
				Computed: true,
				ForceNew: true,
				Optional: true,
				Set:      schema.HashString,
				AtLeastOneOf: []string{
					"column_names",
					"column_wildcard",
				},
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.NoZeroValues,
				},
			},
			"column_wildcard": {
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				AtLeastOneOf: []string{
					"column_names",
					"column_wildcard",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"excluded_column_names": {
							Type:     schema.TypeSet,
							ForceNew: true,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.NoZeroValues,
							},
						},
					},
				},
			},
			"row_filter": {
				Type:     schema.TypeList,
				MaxItems: 1,
				ForceNew: true,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter_expression": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validateLFMultiLineURIString(),
							AtLeastOneOf: []string{
								"row_filter.0.filter_expression",
								"row_filter.0.all_rows_wildcard",
							},
						},
						"all_rows_wildcard": {
							Type:     schema.TypeBool,
							Default:  false,
							ForceNew: true,
							Optional: true,
							AtLeastOneOf: []string{
								"row_filter.0.filter_expression",
								"row_filter.0.all_rows_wildcard",
							},
						},
					},
				},
			},
		},
	}
}

const (
	ResDataCellFilters = "Data Cells Filter"
)

func resourceDataCellsFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LakeFormationConn

	in := &lakeformation.CreateDataCellsFilterInput{
		TableData: &lakeformation.DataCellsFilter{
			DatabaseName: aws.String(d.Get("database_name").(string)),
			TableName:    aws.String(d.Get("table_name").(string)),
			Name:         aws.String(d.Get("name").(string)),
			ColumnNames:  flex.ExpandStringSet(d.Get("column_names").(*schema.Set)),
		},
	}

	var catalogID string
	if v, ok := d.GetOk("table_catalog_id"); ok {
		catalogID = v.(string)
	} else {
		catalogID = meta.(*conns.AWSClient).AccountID
	}
	in.TableData.TableCatalogId = aws.String(catalogID)

	if v, ok := d.GetOk("row_filter"); ok && len(v.([]interface{})) > 0 {
		in.TableData.RowFilter = expandRowFilter(v.([]interface{}))

	}

	if v, ok := d.GetOk("column_wildcard"); ok && len(v.([]interface{})) > 0 {
		in.TableData.ColumnWildcard = expandColumnWildcard(v.([]interface{}))
	}

	// no tags on data cells filter

	out, err := conn.CreateDataCellsFilter(in)
	if err != nil {
		return create.DiagError(names.LakeFormation, create.ErrActionCreating, ResDataCellFilters, d.Get("name").(string), err)
	}

	if out == nil {
		return create.DiagError(names.LakeFormation, create.ErrActionCreating, ResDataCellFilters, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(fmt.Sprintf("%s:%s:%s:%s", catalogID, d.Get("database_name").(string), d.Get("table_name").(string), d.Get("name").(string)))

	return resourceDataCellsFilterRead(ctx, d, meta)
}

// there is no api to get a single data cell filter. the only option is to iterate through a list and attempt to match
func resourceDataCellsFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LakeFormationConn

	catalogID, databaseName, tableName, dataCellsFilterName, err := ReadDataCellsFilterID(d.Id())

	if err != nil {
		log.Printf("[WARN] Data Cells Filter %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	inDCFId := fmt.Sprintf("%s:%s:%s:%s", catalogID, databaseName, tableName, dataCellsFilterName)

	//fmt.Printf("[INFO] Find Lake Formation data cells CatalogId:Database:Table:dataCellFilter %s", inDCFId)

	//accept default MaxResults
	//const maxResults int64 = 1000

	in := &lakeformation.ListDataCellsFilterInput{
		Table: &lakeformation.TableResource{
			CatalogId:    aws.String(catalogID),
			DatabaseName: aws.String(databaseName),
			Name:         aws.String(tableName),
			//Table.TableWildcard
		},
		//MaxResults: aws.Int64(maxResults),
	}

	foundDCF := &lakeformation.DataCellsFilter{}
	bMatch := false
	continueIteration := true
	pageNum := 0
	//recNum := 0
	errList := conn.ListDataCellsFilterPagesWithContext(ctx, in, func(page *lakeformation.ListDataCellsFilterOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		pageNum++

		for recNum, dcf := range page.DataCellsFilters {
			//recNum++
			outDCFId := fmt.Sprintf("%s:%s:%s:%s", *dcf.TableCatalogId, *dcf.DatabaseName, *dcf.TableName, *dcf.Name)
			if inDCFId == outDCFId {
				bMatch = true
				foundDCF = dcf
				continueIteration = false
				log.Printf("[INFO] Found DCF in Lake Formation data cells filter output. item: %v on page: %v", recNum, pageNum)
				break
			}
		}

		// return false to stop the function from iterating through pages
		return continueIteration
	})

	if bMatch {
		d.Set("table_catalog_id", foundDCF.TableCatalogId)
		d.Set("database_name", foundDCF.DatabaseName)
		d.Set("table_name", foundDCF.TableName)
		d.Set("name", foundDCF.Name)

		if err := d.Set("column_names", flex.FlattenStringList(foundDCF.ColumnNames)); err != nil {
			return create.DiagError(names.LakeFormation, create.ErrActionSetting, ResDataCellFilters, d.Id(), err)
		}
		if err := d.Set("column_wildcard", flattenColumnWildcard(foundDCF.ColumnWildcard)); err != nil {
			return create.DiagError(names.LakeFormation, create.ErrActionSetting, ResDataCellFilters, d.Id(), err)
		}
		if err := d.Set("row_filter", flattenRowFilter(foundDCF.RowFilter)); err != nil {
			return create.DiagError(names.LakeFormation, create.ErrActionSetting, ResDataCellFilters, d.Id(), err)
		}
	} else {
		log.Printf("[WARN] Data Cells Filter %s not found, removing from state. ErrList: %v", d.Id(), errList)
		d.SetId("")
		return nil
		//return create.DiagError(names.LakeFormation, create.ErrActionSetting, ResDataCellFilters, d.Id(), errList)
	}

	return nil
}

// no api available for data cells filter update
//func resourceDataCellsFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

func resourceDataCellsFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LakeFormationConn

	catalogID, databaseName, tableName, dataCellsFilterName, err := ReadDataCellsFilterID(d.Id())
	if err != nil {
		log.Printf("[WARN] Data Cells Filter %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	in := &lakeformation.DeleteDataCellsFilterInput{
		TableCatalogId: aws.String(catalogID),
		DatabaseName:   aws.String(databaseName),
		TableName:      aws.String(tableName),
		Name:           aws.String(dataCellsFilterName),
	}

	_, err = conn.DeleteDataCellsFilterWithContext(ctx, in)
	if err != nil {

		if tfawserr.ErrCodeEquals(err, "ResourceNotFoundException") {
			return nil
		}

		return create.DiagError(names.LakeFormation, create.ErrActionSetting, ResDataCellFilters, d.Id(), err)
	}

	return nil
}

// no waiting ... no api to lookup a single data cells filter for confirmation

func ReadDataCellsFilterID(id string) (catalogID string, databaseName string, tableName string, dataCellFilter string, err error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 4 {
		return "", "", "", "", fmt.Errorf("unexpected format of ID (%q), expected CATALOG-ID:DATABASE:TABLE:DATA-CELL-FILTER", id)
	}
	return idParts[0], idParts[1], idParts[2], idParts[3], nil
}

func expandRowFilter(l []interface{}) *lakeformation.RowFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	apiObject := &lakeformation.RowFilter{}

	if v, ok := m["filter_expression"]; ok {
		if len(v.(string)) > 0 {
			apiObject.FilterExpression = aws.String(v.(string))
		}
	}

	if v, ok := m["all_rows_wildcard"].(bool); ok && v {
		apiObject.AllRowsWildcard = &lakeformation.AllRowsWildcard{}
	}

	return apiObject
}

// to push true and false pointer values in flatten function
func newTrue() *bool {
	b := true
	return &b
}
func newFalse() *bool {
	b := false
	return &b
}

func flattenRowFilter(apiObject *lakeformation.RowFilter) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if v := apiObject.FilterExpression; v != nil {
		// api returns "TRUE" for filter_expression when AllRowsWildcard = True, so bypass TRUE
		if aws.StringValue(v) != "TRUE" {
			m["filter_expression"] = aws.StringValue(v)
		}
	}

	if v := apiObject.AllRowsWildcard; v != nil {
		m["all_rows_wildcard"] = aws.BoolValue(newTrue())
	} else {
		m["all_rows_wildcard"] = aws.BoolValue(newFalse())
	}

	return []interface{}{m}
}

func expandColumnWildcard(l []interface{}) *lakeformation.ColumnWildcard {
	if len(l) == 0 {
		return nil
	}

	apiObject := &lakeformation.ColumnWildcard{}

	if l[0] == nil {
		// return empty ColumnWildcard
		return apiObject
	} else {
		m := l[0].(map[string]interface{})

		if v, ok := m["excluded_column_names"]; ok {
			if v, ok := v.(*schema.Set); ok && v.Len() > 0 {
				apiObject.ExcludedColumnNames = flex.ExpandStringSet(v)
			}
		}
	}

	return apiObject
}

func flattenColumnWildcard(apiObject *lakeformation.ColumnWildcard) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"excluded_column_names": flex.FlattenStringSet(apiObject.ExcludedColumnNames),
	}

	return []interface{}{m}
}

//URI address single and multi-line string pattern
// AWS surrogate pair and low surrogate UTF-16: [\u0020-\uD7FF\uE000-\uFFFD\uD800\uDC00-\uDBFF\uDFFF\t]
// GOLANG without surrogate pair UTF-8:         [\x{0020}-\x{D800}\x{E000}-\x{FFFD}\x{DBFF}-\x{DC00}\x{DFFF}\t]
func validateLFSingleLineString() schema.SchemaValidateFunc {
	return validation.All(
		validation.StringLenBetween(1, 255),
		validation.StringMatch(regexp.MustCompile(`[\x{0020}-\x{D800}\x{E000}-\x{FFFD}\x{DBFF}-\x{DC00}\x{DFFF}\t]*`), ""),
	)
}

func validateLFMultiLineURIString() schema.SchemaValidateFunc {
	return validation.All(
		validation.StringLenBetween(1, 2048),
		validation.StringMatch(regexp.MustCompile(`[\x{0020}-\x{D800}\x{E000}-\x{FFFD}\x{DBFF}-\x{DC00}\x{DFFF}\r\n\t]*`), ""),
	)
}
