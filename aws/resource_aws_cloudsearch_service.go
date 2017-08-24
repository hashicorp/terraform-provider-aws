package aws

import (
	"fmt"
	"log"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
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
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
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
	err = resourceAwsCloudSearchDomainUpdate(d, meta)

	return err
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
			DesiredPartitionCount:   aws.Int64(int64(d.Get("partition_count").(int))),
			DesiredReplicationCount: aws.Int64(int64(d.Get("replication_count").(int))),
		},
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
				log.Printf("MICHAS_DEBUG: delete %s", k)
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

func genIndexFieldInput(index map[string]interface{}) (*cloudsearch.IndexField, error) {
	input := &cloudsearch.IndexField{
		IndexFieldName: aws.String(index["name"].(string)),
		IndexFieldType: aws.String(index["type"].(string)),
	}

	switch index["type"] {
	case "int":
		{
			input.IntOptions = &cloudsearch.IntOptions{
				FacetEnabled:  aws.Bool(index["facet"].(bool)),
				ReturnEnabled: aws.Bool(index["return"].(bool)),
				SearchEnabled: aws.Bool(index["search"].(bool)),
				SortEnabled:   aws.Bool(index["sort"].(bool)),
			}

			v, ok := index["default_value"]
			if ok && v.(string) != "" {
				d, err := strconv.Atoi(v.(string))
				if err != nil {
					return input, parseError(v.(string), index["type"].(string))
				}

				input.IntOptions.DefaultValue = aws.Int64(int64(d))
			}
		}
	case "int-array":
		{
			input.IntArrayOptions = &cloudsearch.IntArrayOptions{
				FacetEnabled:  aws.Bool(index["facet"].(bool)),
				ReturnEnabled: aws.Bool(index["return"].(bool)),
				SearchEnabled: aws.Bool(index["search"].(bool)),
			}

			v, ok := index["default_value"]
			if ok && v.(string) != "" {
				d, err := strconv.Atoi(v.(string))
				if err != nil {
					return input, parseError(v.(string), index["type"].(string))
				}

				input.IntArrayOptions.DefaultValue = aws.Int64(int64(d))
			}
		}
	case "double":
		{
			input.DoubleOptions = &cloudsearch.DoubleOptions{
				FacetEnabled:  aws.Bool(index["facet"].(bool)),
				ReturnEnabled: aws.Bool(index["return"].(bool)),
				SearchEnabled: aws.Bool(index["search"].(bool)),
				SortEnabled:   aws.Bool(index["sort"].(bool)),
			}

			v, ok := index["default_value"]
			if ok && v.(string) != "" {
				f, err := strconv.ParseFloat(v.(string), 64)
				if err != nil {
					return input, parseError(v.(string), index["type"].(string))
				}

				input.DoubleOptions.DefaultValue = aws.Float64(f)
			}
		}
	case "double-array":
		{
			input.DoubleArrayOptions = &cloudsearch.DoubleArrayOptions{
				FacetEnabled:  aws.Bool(index["facet"].(bool)),
				ReturnEnabled: aws.Bool(index["return"].(bool)),
				SearchEnabled: aws.Bool(index["search"].(bool)),
			}

			v, ok := index["default_value"]
			if ok {
				f, err := strconv.ParseFloat(v.(string), 64)
				if err != nil {
					return input, parseError(v.(string), index["type"].(string))
				}

				input.DoubleOptions.DefaultValue = aws.Float64(f)
			}
		}
	case "literal":
		{
			input.LiteralOptions = &cloudsearch.LiteralOptions{
				DefaultValue:  aws.String(index["default_value"].(string)),
				FacetEnabled:  aws.Bool(index["facet"].(bool)),
				ReturnEnabled: aws.Bool(index["return"].(bool)),
				SearchEnabled: aws.Bool(index["search"].(bool)),
				SortEnabled:   aws.Bool(index["sort"].(bool)),
			}
		}
	case "literal-array":
		{
			input.LiteralArrayOptions = &cloudsearch.LiteralArrayOptions{
				DefaultValue:  aws.String(index["default_value"].(string)),
				FacetEnabled:  aws.Bool(index["facet"].(bool)),
				ReturnEnabled: aws.Bool(index["return"].(bool)),
				SearchEnabled: aws.Bool(index["search"].(bool)),
			}
		}
	case "text":
		{
			input.TextOptions = &cloudsearch.TextOptions{
				DefaultValue:     aws.String(index["default_value"].(string)),
				SortEnabled:      aws.Bool(index["sort"].(bool)),
				ReturnEnabled:    aws.Bool(index["return"].(bool)),
				HighlightEnabled: aws.Bool(index["highlight"].(bool)),
				AnalysisScheme:   aws.String(index["analysis_scheme"].(string)),
			}
		}
	case "text-array":
		{
			input.TextOptions = &cloudsearch.TextOptions{
				DefaultValue:     aws.String(index["default_value"].(string)),
				ReturnEnabled:    aws.Bool(index["return"].(bool)),
				HighlightEnabled: aws.Bool(index["highlight"].(bool)),
				AnalysisScheme:   aws.String(index["analysis_scheme"].(string)),
			}
		}
	case "date":
		{
			input.DateOptions = &cloudsearch.DateOptions{
				DefaultValue:  aws.String(index["default_value"].(string)),
				FacetEnabled:  aws.Bool(index["facet"].(bool)),
				ReturnEnabled: aws.Bool(index["return"].(bool)),
				SearchEnabled: aws.Bool(index["search"].(bool)),
				SortEnabled:   aws.Bool(index["sort"].(bool)),
			}
		}
	case "date-aray":
		{
			input.DateArrayOptions = &cloudsearch.DateArrayOptions{
				DefaultValue:  aws.String(index["default_value"].(string)),
				FacetEnabled:  aws.Bool(index["facet"].(bool)),
				ReturnEnabled: aws.Bool(index["return"].(bool)),
				SearchEnabled: aws.Bool(index["search"].(bool)),
			}
		}
	case "latlon":
		{
			input.LatLonOptions = &cloudsearch.LatLonOptions{
				DefaultValue:  aws.String(index["default_value"].(string)),
				FacetEnabled:  aws.Bool(index["facet"].(bool)),
				ReturnEnabled: aws.Bool(index["return"].(bool)),
				SearchEnabled: aws.Bool(index["search"].(bool)),
				SortEnabled:   aws.Bool(index["sort"].(bool)),
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
