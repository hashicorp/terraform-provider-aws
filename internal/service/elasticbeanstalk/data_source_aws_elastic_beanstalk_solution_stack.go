package elasticbeanstalk

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceSolutionStack() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSolutionStackRead,

		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			// Computed values.
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// dataSourceSolutionStackRead performs the API lookup.
func dataSourceSolutionStackRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn

	nameRegex := d.Get("name_regex")

	var params *elasticbeanstalk.ListAvailableSolutionStacksInput

	log.Printf("[DEBUG] Reading Elastic Beanstalk Solution Stack: %s", params)
	resp, err := conn.ListAvailableSolutionStacks(params)
	if err != nil {
		return err
	}

	var filteredSolutionStacks []*string

	r := regexp.MustCompile(nameRegex.(string))
	for _, solutionStack := range resp.SolutionStacks {
		if r.MatchString(*solutionStack) {
			filteredSolutionStacks = append(filteredSolutionStacks, solutionStack)
		}
	}

	var solutionStack *string
	if len(filteredSolutionStacks) < 1 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	if len(filteredSolutionStacks) == 1 {
		// Query returned single result.
		solutionStack = filteredSolutionStacks[0]
	} else {
		recent := d.Get("most_recent").(bool)
		log.Printf("[DEBUG] aws_elastic_beanstalk_solution_stack - multiple results found and `most_recent` is set to: %t", recent)
		if recent {
			solutionStack = mostRecentSolutionStack(filteredSolutionStacks)
		} else {
			return fmt.Errorf("Your query returned more than one result. Please try a more " +
				"specific search criteria, or set `most_recent` attribute to true.")
		}
	}

	log.Printf("[DEBUG] aws_elastic_beanstalk_solution_stack - Single solution stack found: %s", *solutionStack)
	return solutionStackDescriptionAttributes(d, solutionStack)
}

// Returns the most recent solution stack out of a slice of stacks.
func mostRecentSolutionStack(solutionStacks []*string) *string {
	return solutionStacks[0]
}

// populate the numerous fields that the image description returns.
func solutionStackDescriptionAttributes(d *schema.ResourceData, solutionStack *string) error {
	// Simple attributes first
	d.SetId(aws.StringValue(solutionStack))
	d.Set("name", solutionStack)
	return nil
}
