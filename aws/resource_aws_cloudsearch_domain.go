package aws

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCloudSearchDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudSearchDomainCreate,
		Read:   resourceAwsCloudSearchDomainRead,
		Update: resourceAwsCloudSearchDomainUpdate,
		Delete: resourceAwsCloudSearchDomainDelete,

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateDomainNameRegex,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"replication_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"partition_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"indexes": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"search": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"facet": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"return": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"sort": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"highlight": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"analysis_scheme": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"default_value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"access_policy": {
				Type:             schema.TypeString,
				ValidateFunc:     validateIAMPolicyJson,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
		},
	}
}

func resourceAwsCloudSearchDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudsearchconn

	input := cloudsearch.CreateDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	output, err := conn.CreateDomain(&input)
	if err != nil {
		return fmt.Errorf("%s %q", err, d.Get("domain_name").(string))
	}

	d.SetId(*output.DomainStatus.ARN)
	return resourceAwsCloudSearchDomainUpdate(d, meta)
}

func resourceAwsCloudSearchDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudsearchconn

	input := cloudsearch.DescribeIndexFieldsInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	_, err := conn.DescribeIndexFields(&input)

	return err
}

func resourceAwsCloudSearchDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudsearchconn

	err := updateScalingParameters(d, meta, conn)
	if err != nil {
		return err
	}

	updated, err := defineIndexFields(d, meta, conn)
	if err != nil {
		return err
	}

	err = updateAccessPolicy(d, meta, conn)
	if err != nil {
		return err
	}

	if updated {
		_, err := conn.IndexDocuments(&cloudsearch.IndexDocumentsInput{
			DomainName: aws.String(d.Get("domain_name").(string)),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func updateScalingParameters(d *schema.ResourceData, meta interface{}, conn *cloudsearch.CloudSearch) error {
	input := cloudsearch.UpdateScalingParametersInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
		ScalingParameters: &cloudsearch.ScalingParameters{
			DesiredInstanceType:     aws.String(d.Get("instance_type").(string)),
			DesiredReplicationCount: aws.Int64(int64(d.Get("replication_count").(int))),
		},
	}

	if d.Get("instance_type").(string) == "search.m3.2xlarge" {
		input.ScalingParameters.DesiredPartitionCount = aws.Int64(int64(d.Get("partition_count").(int)))
	}

	_, err := conn.UpdateScalingParameters(&input)

	return err
}

func defineIndexFields(d *schema.ResourceData, meta interface{}, conn *cloudsearch.CloudSearch) (bool, error) {
	if d.HasChange("indexes") {
		old := make(map[string]interface{})
		new := make(map[string]interface{})

		o, n := d.GetChange("indexes")

		for _, ot := range o.([]interface{}) {
			os := ot.(map[string]interface{})
			old[os["name"].(string)] = os
		}

		for _, nt := range n.([]interface{}) {
			ns := nt.(map[string]interface{})
			new[ns["name"].(string)] = ns
		}

		// Handle Removal
		for k := range old {
			if _, ok := new[k]; !ok {
				deleteIndexField(d.Get("domain_name").(string), k, conn)
			}
		}

		for _, v := range new {
			// Handle replaces & additions
			err := defineIndexField(d.Get("domain_name").(string), v.(map[string]interface{}), conn)
			if err != nil {
				return true, err
			}
		}
		return true, nil
	}
	return false, nil
}

func defineIndexField(domainName string, index map[string]interface{}, conn *cloudsearch.CloudSearch) error {
	i, err := genIndexFieldInput(index)
	if err != nil {
		return err
	}

	input := cloudsearch.DefineIndexFieldInput{
		DomainName: aws.String(domainName),
		IndexField: i,
	}

	_, err = conn.DefineIndexField(&input)
	return err
}

func deleteIndexField(domainName string, indexName string, conn *cloudsearch.CloudSearch) error {
	input := cloudsearch.DeleteIndexFieldInput{
		DomainName:     aws.String(domainName),
		IndexFieldName: aws.String(indexName),
	}

	_, err := conn.DeleteIndexField(&input)
	return err
}

var parseError = func(d string, t string) error {
	return fmt.Errorf("can't convert default_value '%s' of type '%s' to int", d, t)
}

/*
extractFromMapToType extracts a specific value from map[string]interface{} into an interface of type
expects: map[string]interface{}, string, interface{}
returns: error
*/
func extractFromMapToType(index map[string]interface{}, prop string, t interface{}) error {
	v, ok := index[prop]
	if !ok {
		return fmt.Errorf("%s is not a valid propery of an index", prop)
	}

	if "default_value" == prop {
		switch t.(type) {
		case *int:
			{
				d, err := strconv.Atoi(v.(string))
				if err != nil {
					return parseError(v.(string), "int")
				}

				reflect.ValueOf(t).Elem().Set(reflect.ValueOf(d))
			}
		case *float64:
			{
				f, err := strconv.ParseFloat(v.(string), 64)
				if err != nil {
					return parseError(v.(string), "double")
				}

				reflect.ValueOf(t).Elem().Set(reflect.ValueOf(f))
			}
		default:
			{
				if v.(string) != "" {
					reflect.ValueOf(t).Elem().Set(reflect.ValueOf(v))
				}
			}
		}
		return nil
	}

	reflect.ValueOf(t).Elem().Set(reflect.ValueOf(v))
	return nil
}

func genIndexFieldInput(index map[string]interface{}) (*cloudsearch.IndexField, error) {
	input := &cloudsearch.IndexField{
		IndexFieldName: aws.String(index["name"].(string)),
		IndexFieldType: aws.String(index["type"].(string)),
	}

	var facet bool
	var returnV bool
	var search bool
	var sort bool
	var highlight bool
	var analysisScheme string

	extractFromMapToType(index, "facet", &facet)
	extractFromMapToType(index, "return", &returnV)
	extractFromMapToType(index, "search", &search)
	extractFromMapToType(index, "sort", &sort)
	extractFromMapToType(index, "highlight", &highlight)
	extractFromMapToType(index, "analysis_scheme", &analysisScheme)

	switch index["type"] {
	case "int":
		{
			input.IntOptions = &cloudsearch.IntOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
				SortEnabled:   aws.Bool(sort),
			}

			if index["default_value"].(string) != "" {
				var defaultValue int
				extractFromMapToType(index, "default_value", &defaultValue)
				input.IntOptions.DefaultValue = aws.Int64(int64(defaultValue))
			}
		}
	case "int-array":
		{
			input.IntArrayOptions = &cloudsearch.IntArrayOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
			}

			if index["default_value"].(string) != "" {
				var defaultValue int
				extractFromMapToType(index, "default_value", &defaultValue)
				input.IntArrayOptions.DefaultValue = aws.Int64(int64(defaultValue))
			}
		}
	case "double":
		{
			input.DoubleOptions = &cloudsearch.DoubleOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
				SortEnabled:   aws.Bool(sort),
			}

			if index["default_value"].(string) != "" {
				var defaultValue float64
				extractFromMapToType(index, "default_value", &defaultValue)
				input.DoubleOptions.DefaultValue = aws.Float64(float64(defaultValue))
			}
		}
	case "double-array":
		{
			input.DoubleArrayOptions = &cloudsearch.DoubleArrayOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
			}

			if index["default_value"].(string) != "" {
				var defaultValue float64
				extractFromMapToType(index, "default_value", &defaultValue)
				input.DoubleArrayOptions.DefaultValue = aws.Float64(float64(defaultValue))
			}
		}
	case "literal":
		{
			input.LiteralOptions = &cloudsearch.LiteralOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
				SortEnabled:   aws.Bool(sort),
			}

			if index["default_value"].(string) != "" {
				input.LiteralOptions.DefaultValue = aws.String(index["default_value"].(string))
			}
		}
	case "literal-array":
		{
			input.LiteralArrayOptions = &cloudsearch.LiteralArrayOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
			}

			if index["default_value"].(string) != "" {
				input.LiteralArrayOptions.DefaultValue = aws.String(index["default_value"].(string))
			}
		}
	case "text":
		{
			input.TextOptions = &cloudsearch.TextOptions{
				SortEnabled:      aws.Bool(sort),
				ReturnEnabled:    aws.Bool(returnV),
				HighlightEnabled: aws.Bool(highlight),
				AnalysisScheme:   aws.String(analysisScheme),
			}

			if index["default_value"].(string) != "" {
				input.TextOptions.DefaultValue = aws.String(index["default_value"].(string))
			}
		}
	case "text-array":
		{
			input.TextArrayOptions = &cloudsearch.TextArrayOptions{
				ReturnEnabled:    aws.Bool(returnV),
				HighlightEnabled: aws.Bool(highlight),
				AnalysisScheme:   aws.String(analysisScheme),
			}

			if index["default_value"].(string) != "" {
				input.TextArrayOptions.DefaultValue = aws.String(index["default_value"].(string))
			}
		}
	case "date":
		{
			input.DateOptions = &cloudsearch.DateOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
				SortEnabled:   aws.Bool(sort),
			}

			if index["default_value"].(string) != "" {
				input.DateOptions.DefaultValue = aws.String(index["default_value"].(string))
			}
		}
	case "date-aray":
		{
			input.DateArrayOptions = &cloudsearch.DateArrayOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
			}

			if index["default_value"].(string) != "" {
				input.DateArrayOptions.DefaultValue = aws.String(index["default_value"].(string))
			}
		}
	case "latlon":
		{
			input.LatLonOptions = &cloudsearch.LatLonOptions{
				FacetEnabled:  aws.Bool(facet),
				ReturnEnabled: aws.Bool(returnV),
				SearchEnabled: aws.Bool(search),
				SortEnabled:   aws.Bool(sort),
			}

			if index["default_value"].(string) != "" {
				input.LatLonOptions.DefaultValue = aws.String(index["default_value"].(string))
			}
		}
	default:
		return input, fmt.Errorf("invalid index field type %s", index["type"])
	}

	return input, nil
}

func updateAccessPolicy(d *schema.ResourceData, meta interface{}, conn *cloudsearch.CloudSearch) error {
	input := cloudsearch.UpdateServiceAccessPoliciesInput{
		DomainName:     aws.String(d.Get("domain_name").(string)),
		AccessPolicies: aws.String(d.Get("access_policy").(string)),
	}

	_, err := conn.UpdateServiceAccessPolicies(&input)
	return err
}

func resourceAwsCloudSearchDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudsearchconn

	dm := d.Get("domain_name").(string)
	input := cloudsearch.DeleteDomainInput{
		DomainName: aws.String(dm),
	}

	_, err := conn.DeleteDomain(&input)

	return err
}

func validateDomainNameRegex(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[a-z]([a-z0-9-]){2,27}$`).MatchString(value) {
		es = append(es, fmt.Errorf(
			"%q must begin with a lower-case letter, contain only [a-z0-9-] and be at least 3 and at most 28 characters", k))
	}
	return
}
