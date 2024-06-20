package glue

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_catalog_table_optimizer", name="Catalog Table Optimizer")
func ResourceCatalogTableOptimizer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCatalogTableOptimizerCreate,
		ReadWithoutTimeout:   resourceCatalogTableOptimizerRead,
		UpdateWithoutTimeout: resourceCatalogTableOptimizerUpdate,
		DeleteWithoutTimeout: resourceCatalogTableOptimizerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTableName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"compaction"}, false),
			},
			"configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceCatalogTableOptimizerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)
	catalogID := createCatalogID(d, meta.(*conns.AWSClient).AccountID)
	dbName := d.Get(names.AttrDatabaseName).(string)
	tableName := d.Get(names.AttrTableName).(string)

	config := d.Get("configuration").([]interface{})[0].(map[string]interface{})

	input := &glue.CreateTableOptimizerInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableName:    aws.String(tableName),
		TableOptimizerConfiguration: &glue.TableOptimizerConfiguration{
			RoleArn: aws.String(config["role_arn"].(string)),
			Enabled: aws.Bool(config["enabled"].(bool)),
		},
		Type: aws.String(d.Get("type").(string)),
	}

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err := conn.CreateTableOptimizerWithContext(ctx, input)
		if err != nil {
			// Retry IAM propagation errors
			if tfawserr.ErrMessageContains(err, "AccessDeniedException", "does not have the correct trust policies and is unable to be assumed by our service") {
				return retry.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "AccessDeniedException", "does not have the proper IAM permissions to call Glue APIs") {
				return retry.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "AccessDeniedException", "is not authorized to perform") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Catalog Table Optimizer: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s:%s", catalogID, dbName, tableName, d.Get("type").(string)))

	return append(diags, resourceCatalogTableOptimizerRead(ctx, d, meta)...)
}

func resourceCatalogTableOptimizerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 4 {
		return diag.Errorf("unexpected format of ID (%q), expected catalog_id:database_name:table_name:type", d.Id())
	}
	catalogID, dbName, tableName, optimizerType := idParts[0], idParts[1], idParts[2], idParts[3]

	input := &glue.GetTableOptimizerInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableName:    aws.String(tableName),
		Type:         aws.String(optimizerType),
	}

	optimizer, err := conn.GetTableOptimizerWithContext(ctx, input)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Glue Catalog Table Optimizer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Catalog Table Optimizer (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrCatalogID, catalogID)
	d.Set(names.AttrDatabaseName, dbName)
	d.Set(names.AttrTableName, tableName)
	d.Set("type", optimizerType)
	if err := d.Set("configuration", flattenConfiguration(optimizer.TableOptimizer)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
	}

	return diags
}

func resourceCatalogTableOptimizerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	catalogID := d.Get(names.AttrCatalogID).(string)
	dbName := d.Get(names.AttrDatabaseName).(string)
	tableName := d.Get(names.AttrTableName).(string)
	optimizerType := d.Get("type").(string)

	if d.HasChanges("configuration") {
		config := d.Get("configuration").([]interface{})[0].(map[string]interface{})

		input := &glue.UpdateTableOptimizerInput{
			CatalogId:    aws.String(catalogID),
			DatabaseName: aws.String(dbName),
			TableName:    aws.String(tableName),
			TableOptimizerConfiguration: &glue.TableOptimizerConfiguration{
				RoleArn: aws.String(config["role_arn"].(string)),
				Enabled: aws.Bool(config["enabled"].(bool)),
			},
			Type: aws.String(optimizerType),
		}

		err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
			_, err := conn.UpdateTableOptimizerWithContext(ctx, input)
			if err != nil {
				// Retry IAM propagation errors
				if tfawserr.ErrMessageContains(err, "AccessDeniedException", "does not have the correct trust policies and is unable to be assumed by our service") {
					return retry.RetryableError(err)
				}
				if tfawserr.ErrMessageContains(err, "AccessDeniedException", "does not have the proper IAM permissions to call Glue APIs") {
					return retry.RetryableError(err)
				}
				if tfawserr.ErrMessageContains(err, "AccessDeniedException", "is not authorized to perform") {
					return retry.RetryableError(err)
				}

				return retry.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Catalog Table Optimizer (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceCatalogTableOptimizerRead(ctx, d, meta)...)
}

func resourceCatalogTableOptimizerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 4 {
		return diag.Errorf("unexpected format of ID (%q), expected catalog_id:database_name:table_name:type", d.Id())
	}
	catalogID, dbName, tableName, optimizerType := idParts[0], idParts[1], idParts[2], idParts[3]

	input := &glue.DeleteTableOptimizerInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableName:    aws.String(tableName),
		Type:         aws.String(optimizerType),
	}

	log.Printf("[DEBUG] Deleting Glue Catalog Table Optimizer: %s", d.Id())
	_, err := conn.DeleteTableOptimizerWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Glue Catalog Table Optimizer (%s): %s", d.Id(), err)
	}

	return diags
}

func flattenConfiguration(optimizer *glue.TableOptimizer) []interface{} {
	if optimizer == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"role_arn": aws.StringValue(optimizer.Configuration.RoleArn),
			"enabled":  aws.BoolValue(optimizer.Configuration.Enabled),
		},
	}
}
