// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

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

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/shield/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @SDKDataSource("aws_shield_protection", name="Protection")
func DataSourceProtection() *schema.Resource {
	return &schema.Resource{
		// TIP: ==== ASSIGN CRUD FUNCTIONS ====
		// Data sources only have a read function.
		ReadWithoutTimeout: dataSourceProtectionRead,

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
			"arn": { // TIP: Many, but not all, data sources have an `arn` attribute.
				Type:     schema.TypeString,
				Computed: true,
			},
			"protection_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"resource_arn", "protection_id"},
				Default:      nil,
			},
			"resource_arn": { // TIP: Add all your arguments and attributes.
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
				ExactlyOneOf: []string{"resource_arn", "protection_id"},
				Default:      nil,
			},
		},
	}
}

const (
	DSNameProtection = "Protection Data Source"
)

func dataSourceProtectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldClient(ctx)
	input := &shield.DescribeProtectionInput{}

	if v, ok := d.GetOk("resource_arn"); ok && len(v.(string)) > 0 {
		input.ResourceArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("protection_id"); ok && len(v.(string)) > 0 {
		input.ProtectionId = aws.String(v.(string))
	}

	out, err := findProtection(ctx, conn, input)
	//fmt.Printf("Finding %s%s\n", aws.ToString(input.ResourceArn), aws.ToString(input.ProtectionId))
	if err != nil {
		return create.AppendDiagError(diags, names.Shield, create.ErrActionReading, DSNameProtection, fmt.Sprintf("%s%s", aws.ToString(input.ResourceArn), aws.ToString(input.ProtectionId)), err)
	}

	// TIP: -- 3. Set the ID
	//
	// If you don't set the ID, the data source will not be stored in state. In
	// fact, that's how a resource can be removed from state - clearing its ID.
	//
	// If this data source is a companion to a resource, often both will use the
	// same ID. Otherwise, the ID will be a unique identifier such as an AWS
	// identifier, ARN, or name.
	d.SetId(*out.Id)

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
	//    it is equivalent before setting the different JSON.
	d.Set("resource_arn", *out.ResourceArn)
	d.Set("protection_id", *out.Id)
	d.Set("arn", *out.ProtectionArn)

	// TIP: -- 6. Return diags
	return diags
}
