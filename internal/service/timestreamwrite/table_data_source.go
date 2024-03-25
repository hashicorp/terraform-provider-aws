// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
// @SDKDataSource("aws_timestreamwrite_table", name="Table")
func DataSourceTable() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout:   dataSourceTableRead,
		
		Schema: map[string]*schema.Schema{
			"arn": { // TIP: Many, but not all, data sources have an `arn` attribute.
				Type:     schema.TypeString,
				Computed: true,
			},
			"replace_with_arguments": { // TIP: Add all your arguments and attributes.
				Type:     schema.TypeString,
				Optional: true,
			},
			"complex_argument": { // TIP: See setting, getting, flattening, expanding examples below for this complex argument.
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sub_field_one": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
						"sub_field_two": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

const (
	DSNameTable = "Table Data Source"
)

func dataSourceTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// TIP: ==== RESOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Get information about a resource from AWS
	// 3. Set the ID
	// 4. Set the arguments and attributes
	// 5. Set the tags
	// 6. Return diags

	// TIP: -- 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).TimestreamWriteClient(ctx)
	
	// TIP: -- 2. Get information about a resource from AWS using an API Get,
	// List, or Describe-type function, or, better yet, using a finder. Data
	// sources mostly have attributes, or, in other words, computed schema
	// elements. However, a data source will have perhaps one or a few arguments
	// that are key to finding the relevant information, such as 'name' below.
	name := d.Get("name").(string)

	out, err := findTableByName(ctx, conn, name)
	if err != nil {
		return create.AppendDiagError(diags, names.TimestreamWrite, create.ErrActionReading, DSNameTable, name, err)
	}
	
	// TIP: -- 3. Set the ID
	//
	// If you don't set the ID, the data source will not be stored in state. In
	// fact, that's how a resource can be removed from state - clearing its ID.
	// 
	// If this data source is a companion to a resource, often both will use the
	// same ID. Otherwise, the ID will be a unique identifier such as an AWS
	// identifier, ARN, or name.	
	d.SetId(out.ID)
	
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
	d.Set("arn", out.ARN)
	d.Set("name", out.Name)
	
	// TIP: Setting a complex type.
	// For more information, see:
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
	if err := d.Set("complex_argument", flattenComplexArguments(out.ComplexArguments)); err != nil {
		return create.AppendDiagError(diags, names.TimestreamWrite, create.ErrActionSetting, DSNameTable, d.Id(), err)
	}
	
	// TIP: Setting a JSON string to avoid errorneous diffs.
	p, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.ToString(out.Policy))
	if err != nil {
		return create.AppendDiagError(diags, names.TimestreamWrite, create.ErrActionSetting, DSNameTable, d.Id(), err)
	}

	p, err = structure.NormalizeJsonString(p)
	if err != nil {
		return create.AppendDiagError(diags, names.TimestreamWrite, create.ErrActionReading, DSNameTable, d.Id(), err)
	}

	d.Set("policy", p)
	
	// TIP: -- 5. Set the tags
	//
	// TIP: Not all data sources support tags and tags don't always make sense. If
	// your data source doesn't need tags, you can remove the tags lines here and
	// below. Many data sources do include tags so this a reminder to include them
	// where possible.
	
	// TIP: -- 6. Return diags
	return diags
}
