// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_elastic_beanstalk_solution_stack")
func DataSourceSolutionStack() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSolutionStackRead,

		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			names.AttrMostRecent: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			// Computed values.
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// dataSourceSolutionStackRead performs the API lookup.
func dataSourceSolutionStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	nameRegex := d.Get("name_regex")

	var params *elasticbeanstalk.ListAvailableSolutionStacksInput

	log.Printf("[DEBUG] Reading Elastic Beanstalk Solution Stack: %s", params)
	resp, err := conn.ListAvailableSolutionStacks(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Solution Stack: %s", err)
	}

	var filteredSolutionStacks []string

	r := regexache.MustCompile(nameRegex.(string))
	for _, solutionStack := range resp.SolutionStacks {
		if r.MatchString(solutionStack) {
			filteredSolutionStacks = append(filteredSolutionStacks, solutionStack)
		}
	}

	var solutionStack string
	if len(filteredSolutionStacks) < 1 {
		return sdkdiag.AppendErrorf(diags, "Your query returned no results. Please change your search criteria and try again.")
	}

	if len(filteredSolutionStacks) == 1 {
		// Query returned single result.
		solutionStack = filteredSolutionStacks[0]
	} else {
		recent := d.Get(names.AttrMostRecent).(bool)
		log.Printf("[DEBUG] aws_elastic_beanstalk_solution_stack - multiple results found and `most_recent` is set to: %t", recent)
		if recent {
			solutionStack = mostRecentSolutionStack(filteredSolutionStacks)
		} else {
			return sdkdiag.AppendErrorf(diags, "Your query returned more than one result. Please try a more "+
				"specific search criteria, or set `most_recent` attribute to true.")
		}
	}

	d.SetId(solutionStack)
	d.Set(names.AttrName, solutionStack)

	return diags
}

// Returns the most recent solution stack out of a slice of stacks.
func mostRecentSolutionStack(solutionStacks []string) string {
	return solutionStacks[0]
}
