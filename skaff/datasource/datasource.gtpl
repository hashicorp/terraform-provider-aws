// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package {{ .ServicePackage }}

{{- if .IncludeComments }}
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
{{- end }}

import (
{{- if .IncludeComments }}
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/{{ .SDKPackage }}/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
{{- end }}
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }}"
	"github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }}/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
{{- if .IncludeTags }}
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
{{- end }}
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)
{{ if .IncludeComments }}
// TIP: ==== FILE STRUCTURE ====
// All data sources should follow this basic outline. Improve this data source's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main data source function with schema
// 4. Create, read, update, delete functions (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)
{{- end }}

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @SDKDataSource("{{ .ProviderResourceName }}", name="{{ .HumanDataSourceName }}")
func DataSource{{ .DataSource }}() *schema.Resource {
	return &schema.Resource{
		{{- if .IncludeComments }}
		// TIP: ==== ASSIGN CRUD FUNCTIONS ====
		// Data sources only have a read function.
		{{- end }}
		ReadWithoutTimeout:   dataSource{{ .DataSource }}Read,
		{{ if .IncludeComments }}
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
		{{- end }}
		Schema: map[string]*schema.Schema{
			"arn": { {{- if .IncludeComments }} // TIP: Many, but not all, data sources have an `arn` attribute.{{- end }}
				Type:     schema.TypeString,
				Computed: true,
			},
			"replace_with_arguments": { {{- if .IncludeComments }} // TIP: Add all your arguments and attributes.{{- end }}
				Type:     schema.TypeString,
				Optional: true,
			},
			"complex_argument": { {{- if .IncludeComments }} // TIP: See setting, getting, flattening, expanding examples below for this complex argument.{{- end }}
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
			{{- if .IncludeTags }}
			"tags":         tftags.TagsSchemaComputed(), {{- if .IncludeComments }} // TIP: Many, but not all, data sources have `tags` attributes.{{- end }}
			{{- end }}
		},
	}
}

const (
	DSName{{ .DataSource }} = "{{ .HumanDataSourceName }} Data Source"
)

func dataSource{{ .DataSource }}Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	{{- if .IncludeComments }}
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
	{{- end }}
	{{- if .IncludeComments }}

	// TIP: -- 1. Get a client connection to the relevant service
	{{- end }}
	conn := meta.(*conns.AWSClient).{{ .Service }}Client(ctx)
	{{ if .IncludeComments }}
	// TIP: -- 2. Get information about a resource from AWS using an API Get,
	// List, or Describe-type function, or, better yet, using a finder. Data
	// sources mostly have attributes, or, in other words, computed schema
	// elements. However, a data source will have perhaps one or a few arguments
	// that are key to finding the relevant information, such as 'name' below.
	{{- end }}
	name := d.Get("name").(string)

	out, err := find{{ .DataSource }}ByName(ctx, conn, name)
	if err != nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionReading, DSName{{ .DataSource }}, name, err)
	}
	{{ if .IncludeComments }}
	// TIP: -- 3. Set the ID
	//
	// If you don't set the ID, the data source will not be stored in state. In
	// fact, that's how a resource can be removed from state - clearing its ID.
	// 
	// If this data source is a companion to a resource, often both will use the
	// same ID. Otherwise, the ID will be a unique identifier such as an AWS
	// identifier, ARN, or name.
	{{- end }}	
	d.SetId(out.ID)
	{{ if .IncludeComments }}
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
	{{- end }}
	d.Set("arn", out.ARN)
	d.Set("name", out.Name)
	{{ if .IncludeComments }}
	// TIP: Setting a complex type.
	// For more information, see:
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
	{{- end }}
	if err := d.Set("complex_argument", flattenComplexArguments(out.ComplexArguments)); err != nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionSetting, DSName{{ .DataSource }}, d.Id(), err)
	}
	{{ if .IncludeComments }}
	// TIP: Setting a JSON string to avoid errorneous diffs.
	{{- end }}
	p, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.ToString(out.Policy))
	if err != nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionSetting, DSName{{ .DataSource }}, d.Id(), err)
	}

	p, err = structure.NormalizeJsonString(p)
	if err != nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionReading, DSName{{ .DataSource }}, d.Id(), err)
	}

	d.Set("policy", p)
	{{ if .IncludeComments }}
	// TIP: -- 5. Set the tags
	//
	// TIP: Not all data sources support tags and tags don't always make sense. If
	// your data source doesn't need tags, you can remove the tags lines here and
	// below. Many data sources do include tags so this a reminder to include them
	// where possible.
	{{- end }}
	{{- if .IncludeTags }}
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)

	if err := d.Set("tags", KeyValueTags(out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionSetting, DSName{{ .DataSource }}, d.Id(), err)
	}
	{{- end }}
	{{ if .IncludeComments }}
	// TIP: -- 6. Return diags
	{{- end }}
	return diags
}
