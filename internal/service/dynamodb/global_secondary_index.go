// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/dlclark/regexp2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	resNameGlobalSecondaryIndex = "global_secondary_index"
)

// @SDKResource("aws_dynamodb_global_secondary_index", name="Global Secondary Index")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/dynamodb/types;types.TableDescription")
func resourceGlobalSecondaryIndex() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceGlobalSecondaryIndexCreate,
		ReadWithoutTimeout:   resourceGlobalSecondaryIndexRead,
		UpdateWithoutTimeout: resourceGlobalSecondaryIndexUpdate,
		DeleteWithoutTimeout: resourceGlobalSecondaryIndexDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(createTableTimeout),
			Delete: schema.DefaultTimeout(deleteTableTimeout),
			Update: schema.DefaultTimeout(updateTableTimeoutTotal),
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			names.AttrIndexARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"table": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hash_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hash_key_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"range_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"range_key_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"non_key_attributes": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"on_demand_throughput": onDemandThroughputSchema(),
			"projection_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ProjectionType](),
			},
			"read_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"write_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},

		CustomizeDiff: customdiff.All(
			func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				return validateNewGSIAttributes(ctx, diff, meta)
			},
		),
	}
}

func resourceGlobalSecondaryIndexCreate(ctx context.Context, res *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*conns.AWSClient)
	conn := c.DynamoDBClient(ctx)
	diags := diag.Diagnostics{}
	tableName := res.Get("table").(string)

	if err := waitAllGSIActive(ctx, conn, tableName, res.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForCreation, resNameGlobalSecondaryIndex, res.Id(), err)
	}

	table, err := findTableByName(ctx, conn, tableName)
	if err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForCreation, resNameGlobalSecondaryIndex, res.Id(), err)
	}

	knownAttributes := map[string]awstypes.ScalarAttributeType{}
	ut := &dynamodb.UpdateTableInput{
		TableName:            aws.String(tableName),
		AttributeDefinitions: []awstypes.AttributeDefinition{},
	}

	for _, ad := range table.AttributeDefinitions {
		ut.AttributeDefinitions = append(ut.AttributeDefinitions, ad)
		knownAttributes[aws.ToString(ad.AttributeName)] = ad.AttributeType
	}

	hashKey := res.Get("hash_key").(string)
	rangeKey, hasRangeKey := res.GetOk("range_key")
	keySchema := []awstypes.KeySchemaElement{}

	if hashKeyType, found := knownAttributes[hashKey]; found {
		if err = res.Set("hash_key_type", string(hashKeyType)); err != nil {
			diags = append(diags, errs.NewErrorDiagnostic(`Unable to set field "hash_key_type"`, fmt.Sprintf("Error: %v", err)))
		}
	} else {
		hashKeyType, hasHashKeyType := res.GetOk("hash_key_type")
		if hasHashKeyType {
			ut.AttributeDefinitions = append(ut.AttributeDefinitions, awstypes.AttributeDefinition{
				AttributeName: aws.String(hashKey),
				AttributeType: awstypes.ScalarAttributeType(hashKeyType.(string)),
			})
			knownAttributes[hashKey] = awstypes.ScalarAttributeType(hashKeyType.(string))
		} else {
			diags = append(diags, errs.NewErrorDiagnostic(
				`"hash_key_type" must be set in this configuration`,
				`When using "hash_key" that is not defined in the attributes list of the table, you must specify the "hash_key_type"`,
			))
		}
	}

	keySchema = append(keySchema, awstypes.KeySchemaElement{
		AttributeName: aws.String(hashKey),
		KeyType:       awstypes.KeyTypeHash,
	})

	if hasRangeKey {
		if rangeKeyType, found := knownAttributes[rangeKey.(string)]; found {
			if err = res.Set("range_key_type", string(rangeKeyType)); err != nil {
				diags = append(diags, errs.NewErrorDiagnostic(`Unable to set field "range_key_type"`, fmt.Sprintf("Error: %v", err)))
			}
		} else {
			rangeKeyType, hasRangeKeyType := res.GetOk("range_key_type")
			if hasRangeKeyType {
				ut.AttributeDefinitions = append(ut.AttributeDefinitions, awstypes.AttributeDefinition{
					AttributeName: aws.String(rangeKey.(string)),
					AttributeType: awstypes.ScalarAttributeType(rangeKeyType.(string)),
				})
				knownAttributes[rangeKey.(string)] = awstypes.ScalarAttributeType(rangeKeyType.(string))
			} else {
				diags = append(diags, errs.NewErrorDiagnostic(
					`"range_key_type" must be set in this configuration`,
					`When using "range_key" that is not defined in the attributes list of the table, you must specify the "range_key_type"`,
				))
			}
		}
		keySchema = append(keySchema, awstypes.KeySchemaElement{
			AttributeName: aws.String(rangeKey.(string)),
			KeyType:       awstypes.KeyTypeRange,
		})
	}

	projection := &awstypes.Projection{
		ProjectionType: awstypes.ProjectionType(res.Get("projection_type").(string)),
	}

	if v, ok := res.Get("non_key_attributes").([]any); ok && len(v) > 0 {
		projection.NonKeyAttributes = flex.ExpandStringValueList(v)
	}

	if v, ok := res.Get("non_key_attributes").(*schema.Set); ok && v.Len() > 0 {
		projection.NonKeyAttributes = flex.ExpandStringValueSet(v)
	}

	action := &awstypes.CreateGlobalSecondaryIndexAction{
		IndexName:  aws.String(res.Get(names.AttrName).(string)),
		KeySchema:  keySchema,
		Projection: projection,
	}

	billingMode := awstypes.BillingModeProvisioned
	if table.BillingModeSummary != nil {
		billingMode = table.BillingModeSummary.BillingMode
	}

	if billingMode == awstypes.BillingModeProvisioned {
		rc := res.Get("read_capacity").(int)
		wc := res.Get("write_capacity").(int)

		if rc == 0 && table.ProvisionedThroughput != nil {
			rc = int(aws.ToInt64(table.ProvisionedThroughput.ReadCapacityUnits))
		}
		if wc == 0 && table.ProvisionedThroughput != nil {
			wc = int(aws.ToInt64(table.ProvisionedThroughput.WriteCapacityUnits))
		}

		provisionedThroughputData := map[string]any{
			"read_capacity":  rc,
			"write_capacity": wc,
		}
		action.ProvisionedThroughput = expandProvisionedThroughput(provisionedThroughputData, billingMode)
	} else if v, ok := res.Get("on_demand_throughput").([]any); ok && len(v) > 0 && v[0] != nil {
		action.OnDemandThroughput = expandOnDemandThroughput(v[0].(map[string]any))
	}

	ut.GlobalSecondaryIndexUpdates = []awstypes.GlobalSecondaryIndexUpdate{
		{
			Create: action,
		},
	}

	if utRes, err := conn.UpdateTable(ctx, ut); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionCreating, resNameGlobalSecondaryIndex, res.Id(), fmt.Errorf("creating GSI (%s): %w", res.Get(names.AttrName).(string), err))
	} else {
		for _, gsi := range utRes.TableDescription.GlobalSecondaryIndexes {
			if aws.ToString(gsi.IndexName) == res.Get(names.AttrName).(string) {
				if err := res.Set(names.AttrIndexARN, aws.ToString(gsi.IndexArn)); err != nil {
					diags = append(diags, errs.NewErrorDiagnostic(fmt.Sprintf(`Unable to set field "%s"`, names.AttrIndexARN), fmt.Sprintf("Error: %v", err)))
				}

				res.SetId(aws.ToString(gsi.IndexArn))
			}
		}
	}

	if _, err = waitTableActive(ctx, conn, tableName, res.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForCreation, resNameGlobalSecondaryIndex, res.Id(), err)
	}

	if _, err := waitGSIActive(ctx, conn, tableName, res.Get(names.AttrName).(string), res.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForCreation, resNameGlobalSecondaryIndex, res.Id(), fmt.Errorf("GSI (%s): %w", res.Get(names.AttrName).(string), err))
	}

	return diags
}

func resourceGlobalSecondaryIndexRead(ctx context.Context, res *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*conns.AWSClient)
	conn := c.DynamoDBClient(ctx)
	diags := diag.Diagnostics{}

	re := regexp2.MustCompile(":table/([^\\/]+)/index/(.+)", regexp2.IgnoreCase)
	m, err := re.FindStringMatch(res.Id())
	if err != nil {
		diags = create.AppendDiagError(
			diags,
			names.DynamoDB,
			create.ErrActionReading,
			"global_secondary_index",
			res.Id(),
			err,
		)

		return diags
	}

	grps := m.Groups()
	if len(grps) != 3 {
		diags = create.AppendDiagError(
			diags,
			names.DynamoDB,
			create.ErrActionReading,
			"global_secondary_index",
			res.Id(),
			errors.New("unable to determine the index name"),
		)

		return diags
	}

	table, err := findTableByName(ctx, conn, grps[1].String())

	// table not found, everything is fresh
	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		create.LogNotFoundRemoveState(names.DynamoDB, create.ErrActionReading, resNameGlobalSecondaryIndex, res.Id())
		res.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(
			diags,
			names.DynamoDB,
			create.ErrActionReading,
			"global_secondary_index",
			res.Id(),
			err,
		)
	}

	found := false
	for _, g := range table.GlobalSecondaryIndexes {
		if aws.ToString(g.IndexArn) == res.Id() {
			found = true

			if err := res.Set(names.AttrName, aws.ToString(g.IndexName)); err != nil {
				diags = append(diags, errs.NewErrorDiagnostic(fmt.Sprintf(`Unable to set field "%s"`, names.AttrName), fmt.Sprintf("Error: %v", err)))
			}
			if err := res.Set(names.AttrIndexARN, aws.ToString(g.IndexArn)); err != nil {
				diags = append(diags, errs.NewErrorDiagnostic(fmt.Sprintf(`Unable to set field "%s"`, names.AttrIndexARN), fmt.Sprintf("Error: %v", err)))
			}
			if err := res.Set("table", aws.ToString(table.TableName)); err != nil {
				diags = append(diags, errs.NewErrorDiagnostic(`Unable to set field "table"`, fmt.Sprintf("Error: %v", err)))
			}

			if g.ProvisionedThroughput != nil {
				if err = res.Set("write_capacity", aws.ToInt64(g.ProvisionedThroughput.WriteCapacityUnits)); err != nil {
					diags = append(diags, errs.NewErrorDiagnostic(`Unable to set field "write_capacity"`, fmt.Sprintf("Error: %v", err)))
				}
				if err = res.Set("read_capacity", aws.ToInt64(g.ProvisionedThroughput.ReadCapacityUnits)); err != nil {
					diags = append(diags, errs.NewErrorDiagnostic(`Unable to set field "read_capacity"`, fmt.Sprintf("Error: %v", err)))
				}
			}

			for _, attribute := range g.KeySchema {
				if attribute.KeyType == awstypes.KeyTypeHash {
					if err = res.Set("hash_key", aws.ToString(attribute.AttributeName)); err != nil {
						diags = append(diags, errs.NewErrorDiagnostic(`Unable to set field "hash_key"`, fmt.Sprintf("Error: %v", err)))
					}

					for _, ad := range table.AttributeDefinitions {
						if aws.ToString(ad.AttributeName) == aws.ToString(attribute.AttributeName) {
							if err = res.Set("hash_key_type", string(ad.AttributeType)); err != nil {
								diags = append(diags, errs.NewErrorDiagnostic(`Unable to set field "hash_key_type"`, fmt.Sprintf("Error: %v", err)))
							}
						}
					}
				}

				if attribute.KeyType == awstypes.KeyTypeRange {
					if err = res.Set("range_key", aws.ToString(attribute.AttributeName)); err != nil {
						diags = append(diags, errs.NewErrorDiagnostic(`Unable to set field "range_key"`, fmt.Sprintf("Error: %v", err)))
					}

					for _, ad := range table.AttributeDefinitions {
						if aws.ToString(ad.AttributeName) == aws.ToString(attribute.AttributeName) {
							if err = res.Set("range_key_type", string(ad.AttributeType)); err != nil {
								diags = append(diags, errs.NewErrorDiagnostic(`Unable to set field "range_key_type"`, fmt.Sprintf("Error: %v", err)))
							}
						}
					}
				}
			}

			if g.Projection != nil {
				if err = res.Set("projection_type", g.Projection.ProjectionType); err != nil {
					diags = append(diags, errs.NewErrorDiagnostic(`Unable to set field "projection_type"`, fmt.Sprintf("Error: %v", err)))
				}
				if err = res.Set("non_key_attributes", g.Projection.NonKeyAttributes); err != nil {
					diags = append(diags, errs.NewErrorDiagnostic(`Unable to set field "non_key_attributes"`, fmt.Sprintf("Error: %v", err)))
				}
			}

			if g.OnDemandThroughput != nil {
				if err = res.Set("on_demand_throughput", flattenOnDemandThroughput(g.OnDemandThroughput)); err != nil {
					diags = append(diags, errs.NewErrorDiagnostic(`Unable to set field "on_demand_throughput"`, fmt.Sprintf("Error: %v", err)))
				}
			}
		}
	}

	// external removal of index or recreation of table without indexes
	if !found {
		create.LogNotFoundRemoveState(names.DynamoDB, create.ErrActionReading, resNameGlobalSecondaryIndex, res.Id())

		res.SetId("")
	}

	if len(diags) > 0 {
		return diags
	}

	return diags
}

func resourceGlobalSecondaryIndexUpdate(ctx context.Context, res *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*conns.AWSClient)
	conn := c.DynamoDBClient(ctx)
	diags := diag.Diagnostics{}
	tableName := res.Get("table").(string)

	if err := waitAllGSIActive(ctx, conn, tableName, res.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForUpdate, resNameGlobalSecondaryIndex, res.Id(), err)
	}

	table, err := findTableByName(ctx, conn, tableName)

	if err != nil {
		return create.AppendDiagError(
			diags,
			names.DynamoDB,
			create.ErrActionWaitingForUpdate,
			resNameGlobalSecondaryIndex,
			res.Id(),
			err,
		)
	}

	action := &awstypes.UpdateGlobalSecondaryIndexAction{
		IndexName: aws.String(res.Get(names.AttrName).(string)),
	}

	billingMode := awstypes.BillingModeProvisioned
	if table.BillingModeSummary != nil {
		billingMode = table.BillingModeSummary.BillingMode
	}

	hasUpdate := false

	if billingMode == awstypes.BillingModeProvisioned {
		provisionedThroughputData := map[string]any{
			"read_capacity":  res.Get("read_capacity").(int),
			"write_capacity": res.Get("write_capacity").(int),
		}
		newProvisionedThroughput := expandProvisionedThroughput(provisionedThroughputData, billingMode)

		idx := slices.Filter(table.GlobalSecondaryIndexes, func(description awstypes.GlobalSecondaryIndexDescription) bool {
			return aws.ToString(description.IndexArn) == res.Id()
		})

		if len(idx) == 0 {
			return create.AppendDiagError(
				diags,
				names.DynamoDB,
				create.ErrActionWaitingForUpdate,
				resNameGlobalSecondaryIndex,
				res.Id(),
				fmt.Errorf("unable to find index with arn: %s", res.Id()),
			)
		}

		if idx[0].ProvisionedThroughput == nil {
			action.ProvisionedThroughput = newProvisionedThroughput
			hasUpdate = true
		} else if aws.ToInt64(idx[0].ProvisionedThroughput.ReadCapacityUnits) != aws.ToInt64(newProvisionedThroughput.ReadCapacityUnits) || aws.ToInt64(idx[0].ProvisionedThroughput.WriteCapacityUnits) != aws.ToInt64(newProvisionedThroughput.WriteCapacityUnits) {
			action.ProvisionedThroughput = newProvisionedThroughput
			hasUpdate = true
		}
	} else if v, ok := res.Get("on_demand_throughput").([]any); ok && len(v) > 0 && v[0] != nil {
		action.OnDemandThroughput = expandOnDemandThroughput(v[0].(map[string]any))
		hasUpdate = true
	}

	ut := &dynamodb.UpdateTableInput{
		TableName: aws.String(tableName),
		GlobalSecondaryIndexUpdates: []awstypes.GlobalSecondaryIndexUpdate{
			{
				Update: action,
			},
		},
	}

	if hasUpdate {
		if _, err := conn.UpdateTable(ctx, ut); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameGlobalSecondaryIndex, res.Id(), fmt.Errorf("updating GSI (%s): %w", res.Get(names.AttrName).(string), err))
		}

		if _, err := waitTableActive(ctx, conn, tableName, res.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForUpdate, resNameGlobalSecondaryIndex, res.Id(), err)
		}

		if _, err := waitGSIActive(ctx, conn, tableName, res.Get(names.AttrName).(string), res.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForUpdate, resNameGlobalSecondaryIndex, res.Id(), fmt.Errorf("GSI (%s): %w", res.Get(names.AttrName).(string), err))
		}
	}

	return diags
}

func resourceGlobalSecondaryIndexDelete(ctx context.Context, res *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*conns.AWSClient)
	conn := c.DynamoDBClient(ctx)
	diags := diag.Diagnostics{}
	tableName := res.Get("table").(string)

	if err := waitAllGSIActive(ctx, conn, tableName, res.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForCreation, resNameGlobalSecondaryIndex, res.Id(), err)
	}

	table, err := findTableByName(ctx, conn, tableName)

	// already deleted, nothing to do
	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		diags = create.AppendDiagError(
			diags,
			names.DynamoDB,
			create.ErrActionReading,
			"global_secondary_index",
			res.Id(),
			err,
		)

		return diags
	}

	knownAttributes := map[string]int{}
	for _, l := range table.LocalSecondaryIndexes {
		for _, ks := range l.KeySchema {
			knownAttributes[aws.ToString(ks.AttributeName)] = knownAttributes[aws.ToString(ks.AttributeName)] + 1
		}
	}

	for _, g := range table.GlobalSecondaryIndexes {
		if res.Id() != aws.ToString(g.IndexArn) {
			for _, ks := range g.KeySchema {
				knownAttributes[aws.ToString(ks.AttributeName)] = knownAttributes[aws.ToString(ks.AttributeName)] + 1
			}
		}
	}

	ut := &dynamodb.UpdateTableInput{
		TableName:            aws.String(res.Get("table").(string)),
		AttributeDefinitions: []awstypes.AttributeDefinition{},
		GlobalSecondaryIndexUpdates: []awstypes.GlobalSecondaryIndexUpdate{
			{
				Delete: &awstypes.DeleteGlobalSecondaryIndexAction{
					IndexName: aws.String(res.Get(names.AttrName).(string)),
				},
			},
		},
	}

	for _, ad := range table.AttributeDefinitions {
		if knownAttributes[aws.ToString(ad.AttributeName)] > 0 {
			ut.AttributeDefinitions = append(ut.AttributeDefinitions, ad)
		}
	}

	if utRes, err := conn.UpdateTable(ctx, ut); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionUpdating, resNameGlobalSecondaryIndex, res.Id(), fmt.Errorf("creating GSI (%s): %w", res.Get(names.AttrName).(string), err))
	} else {
		for _, gsi := range utRes.TableDescription.GlobalSecondaryIndexes {
			if aws.ToString(gsi.IndexName) == res.Get(names.AttrName).(string) {
				if err := res.Set(names.AttrIndexARN, aws.ToString(gsi.IndexArn)); err != nil {
					diags = append(diags, errs.NewErrorDiagnostic(fmt.Sprintf(`Unable to set field "%s"`, names.AttrIndexARN), fmt.Sprintf("Error: %v", err)))
				}

				res.SetId(aws.ToString(gsi.IndexArn))
			}
		}
	}

	if _, err = waitTableActive(ctx, conn, tableName, res.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForCreation, resNameGlobalSecondaryIndex, res.Id(), err)
	}

	if _, err := waitGSIDeleted(ctx, conn, tableName, res.Get(names.AttrName).(string), res.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.DynamoDB, create.ErrActionWaitingForCreation, resNameGlobalSecondaryIndex, res.Id(), fmt.Errorf("GSI (%s): %w", res.Get(names.AttrName).(string), err))
	}

	if len(diags) > 0 {
		return diags
	}

	return diags
}

func waitAllGSIActive(ctx context.Context, conn *dynamodb.Client, tableName string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TableStatusCreating, awstypes.TableStatusUpdating),
		Target:  enum.Slice(awstypes.TableStatusActive),
		Refresh: statusTable(ctx, conn, tableName),
		Timeout: max(createTableTimeout, timeout),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return err
	}

	if output, ok := outputRaw.(*awstypes.TableDescription); ok {
		for _, g := range output.GlobalSecondaryIndexes {
			if g.IndexStatus != awstypes.IndexStatusActive {
				if _, err := waitGSIActive(ctx, conn, tableName, aws.ToString(g.IndexName), timeout); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func validateNewGSIAttributes(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
	if diff.Id() != "" {
		return nil
	}

	c := meta.(*conns.AWSClient)
	conn := c.DynamoDBClient(ctx)
	tableName := diff.Get("table").(string)

	var errs []error

	table, err := findTableByName(ctx, conn, tableName)

	// table is fresh, nothing to validate
	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return err
	}

	// count attribute usages
	counts := map[string]int{}
	for _, ks := range table.KeySchema {
		counts[aws.ToString(ks.AttributeName)] = counts[aws.ToString(ks.AttributeName)] + 1
	}

	for _, l := range table.LocalSecondaryIndexes {
		for _, ks := range l.KeySchema {
			counts[aws.ToString(ks.AttributeName)] = counts[aws.ToString(ks.AttributeName)] + 1
		}
	}

	for _, g := range table.GlobalSecondaryIndexes {
		if aws.ToString(g.IndexName) != diff.Get(names.AttrName).(string) {
			for _, ks := range g.KeySchema {
				counts[aws.ToString(ks.AttributeName)] = counts[aws.ToString(ks.AttributeName)] + 1
			}
		}
	}

	for _, prop := range []string{"hash_key_type", "range_key_type"} {
		name := diff.Get(prop[:len(prop)-5]).(string)
		if name == "" {
			continue
		}

		existing := ""
		for _, ad := range table.AttributeDefinitions {
			if aws.ToString(ad.AttributeName) == name {
				existing = string(ad.AttributeType)
			}
		}

		if existing == "" {
			continue
		}

		if diff.Get(prop).(string) == "" {
			if err := diff.SetNew(prop, existing); err != nil {
				errs = append(errs, err)
			}
		}

		if existing != diff.Get(prop).(string) && counts[name] > 0 {
			errs = append(errs, fmt.Errorf(
				`creation of index "%s" on table "%s" is attempting to change already existing attribute "%s" from type "%s" to "%s"`,
				diff.Get(names.AttrName).(string),
				tableName,
				name,
				existing,
				diff.Get(prop).(string),
			))
		}
	}

	return errors.Join(errs...)
}
