package dynamodb

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.
//
// Remember to register this new data source in the provider
// (internal/provider/provider.go) once you finish. Otherwise, Terraform won't
// know about it.

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"log"
	"strings"
)

// TIP: ==== FILE STRUCTURE ====
// All data sources should follow this basic outline. Improve this data source's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main data source function with schema
// 4. Create, read, update, delete functions (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)
func DataSourceTableItem() *schema.Resource {
	return &schema.Resource{
		// TIP: ==== ASSIGN CRUD FUNCTIONS ====
		// Data sources only have a read function.
		ReadWithoutTimeout: dataSourceTableItemRead,

		// TIP: ==== SCHEMA ====
		// In the schema, add each of the arguments and attributes in snake
		// case (e.g., delete_automated_backups).
		// * Alphabetize arguments to make them easier to find.
		// * Do not add a blank line between arguments/attributes.
		//
		// Users can configure argument values while attribute values cannot be
		// configured and are used as output. Arguments have either:
		// Required: true,
		// Optional: true,
		//
		// All attributes will be computed and some arguments. If users will
		// want to read updated information or detect drift for an argument,
		// it should be computed:
		// Computed: true,
		//
		// You will typically find arguments in the input struct
		// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
		// they are only in the input struct (e.g., ModifyDBInstanceInput) for
		// the modify operation.
		//
		// For more about schema options, visit
		// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema#Schema
		Schema: map[string]*schema.Schema{
			"expression_attribute_names": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"item": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateTableItem,
			},
			"projection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"table_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

const (
	DSNameTableItem = "Table Item Data Source"
)

func dataSourceTableItemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TIP: ==== RESOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Get information about a resource from AWS
	// 3. Set the ID
	// 4. Set the arguments and attributes
	// 5. Set the tags
	// 6. Return nil

	// TIP: -- 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).DynamoDBConn

	// TIP: -- 2. Get information about a resource from AWS using an API Get,
	// List, or Describe-type function, or, better yet, using a finder. Data
	// sources mostly have attributes, or, in other words, computed schema
	// elements. However, a data source will have perhaps one or a few arguments
	// that are key to finding the relevant information, such as 'name' below.

	tableName := d.Get("table_name").(string)
	key, err := ExpandTableItemAttributes(d.Get("key").(string))

	id := buildTableItemDataSourceID(tableName, key)

	log.Printf("[DEBUG] DynamoDB item get: %s | %s", tableName, id)

	in := &dynamodb.GetItemInput{
		TableName:      aws.String(tableName),
		ConsistentRead: aws.Bool(true),
		Key:            key,
	}

	if v, ok := d.GetOk("expression_attribute_names"); ok && len(v.(map[string]interface{})) > 0 {
		in.ExpressionAttributeNames = flex.ExpandStringMap(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("projection_expression"); ok {
		in.ProjectionExpression = aws.String(v.(string))
	}

	out, err := conn.GetItem(in)

	fmt.Println(in.String())
	fmt.Println(out.String())

	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionReading, DSNameTableItem, id, err)
	}

	if out.Item == nil {
		return create.DiagError(names.DynamoDB, create.ErrActionReading, DSNameTableItem, id, err)
	}

	// TIP: -- 3. Set the ID
	//
	// If you don't set the ID, the data source will not be stored in state. In
	// fact, that's how a resource can be removed from state - clearing its ID.
	//
	// If this data source is a companion to a resource, often both will use the
	// same ID. Otherwise, the ID will be a unique identifier such as an AWS
	// identifier, ARN, or name.

	d.SetId(id)

	// TIP: -- 4. Set the arguments and attributes
	//
	// For simple data types (i.e., schema.TypeString, schema.TypeBool,
	// schema.TypeInt, and schema.TypeFloat), a simple Set call (e.g.,
	// d.Set("arn", out.Arn) is sufficient. No error or nil checking is
	// necessary.
	//
	// However, there are some situations where more handling is needed.
	// a. Complex data types (e.g., schema.TypeList, schema.TypeSet)
	// b. Where errorneous diffs occur. For example, a schema.TypeString may be
	//    a JSON. AWS may return the JSON in a slightly different order but it
	//    is equivalent to what is already set. In that case, you may check if
	//    it is equivalent before setting the different JSON

	d.Set("projection_expression", in.ProjectionExpression)
	d.Set("expression_attribute_names", aws.StringValueMap(in.ExpressionAttributeNames))
	d.Set("table_name", tableName)

	itemAttrs, err := flattenTableItemAttributes(out.Item)

	if err != nil {
		return create.DiagError(names.DynamoDB, create.ErrActionReading, DSNameTableItem, id, err)
	}
	d.Set("item", itemAttrs)

	// TIP: -- 6. Return nil
	return nil
}

func buildTableItemDataSourceID(tableName string, attrs map[string]*dynamodb.AttributeValue) string {
	id := []string{tableName}

	for key, element := range attrs {
		id = append(id, key, verify.Base64Encode(element.B))
		id = append(id, aws.StringValue(element.S))
		id = append(id, aws.StringValue(element.N))
	}
	return strings.Join(id, "|")
}
