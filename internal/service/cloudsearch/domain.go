package cloudsearch

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudSearchDomainCreate,
		Read:   resourceCloudSearchDomainRead,
		Update: resourceCloudSearchDomainUpdate,
		Delete: resourceCloudSearchDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			// TODO: Separate access policy resource?
			// TODO: Is it Required?
			"access_policies": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_options": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enforce_https": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"tls_security_policy": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(cloudsearch.TLSSecurityPolicy_Values(), false),
						},
					},
				},
			},
			"multi_az": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z]([a-z0-9-]){2,27}$`), "Search domain names must start with a lowercase letter (a-z) and be at least 3 and no more than 28 lower-case letters, digits or hyphens"),
			},
			"scaling_parameters": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_instance_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(cloudsearch.PartitionInstanceType_Values(), false),
						},
						"desired_partition_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"desired_replication_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},

			"index": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateIndexName,
						},

						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(cloudsearch.IndexFieldType_Values(), false),
						},

						"search": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						"facet": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						"return": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						"sort": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						"highlight": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
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

			"wait_for_endpoints": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool { return true },
			},

			"document_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"search_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// Terraform CRUD Functions
func resourceCloudSearchDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudSearchConn

	name := d.Get("name").(string)
	input := cloudsearch.CreateDomainInput{
		DomainName: aws.String(name),
	}

	log.Printf("[DEBUG] Creating CloudSearch Domain: %s", input)
	_, err := conn.CreateDomain(&input)

	if err != nil {
		return fmt.Errorf("error creating CloudSearch Domain (%s): %w", name, err)
	}

	d.SetId(name)

	log.Printf("[DEBUG] Updating CloudSearch Domain (%s) access policies", name)
	_, err = conn.UpdateServiceAccessPolicies(&cloudsearch.UpdateServiceAccessPoliciesInput{
		DomainName:     aws.String(d.Id()),
		AccessPolicies: aws.String(d.Get("access_policies").(string)),
	})

	if err != nil {
		return fmt.Errorf("error updating CloudSearch Domain (%s) service access policies: %w", name, err)
	}

	if v, ok := d.GetOk("scaling_parameters"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := &cloudsearch.UpdateScalingParametersInput{
			DomainName:        aws.String(d.Id()),
			ScalingParameters: expandScalingParameters(v.([]interface{})[0].(map[string]interface{})),
		}

		log.Printf("[DEBUG] Updating CloudSearch Domain scaling parameters: %s", input)
		_, err := conn.UpdateScalingParameters(input)

		if err != nil {
			return fmt.Errorf("error updating CloudSearch Domain (%s) scaling parameters: %w", name, err)
		}
	}

	if v, ok := d.GetOk("multi_az"); ok {
		input := &cloudsearch.UpdateAvailabilityOptionsInput{
			DomainName: aws.String(d.Id()),
			MultiAZ:    aws.Bool(v.(bool)),
		}

		log.Printf("[DEBUG] Updating CloudSearch Domain availability options: %s", input)
		_, err := conn.UpdateAvailabilityOptions(input)

		if err != nil {
			return fmt.Errorf("error updating CloudSearch Domain (%s) availability options: %w", name, err)
		}
	}

	if v, ok := d.GetOk("scaling_parameters"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := &cloudsearch.UpdateDomainEndpointOptionsInput{
			DomainEndpointOptions: expandDomainEndpointOptions(v.([]interface{})[0].(map[string]interface{})),
			DomainName:            aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating CloudSearch Domain endpoint options: %s", input)
		_, err := conn.UpdateDomainEndpointOptions(input)

		if err != nil {
			return fmt.Errorf("error updating CloudSearch Domain (%s) endpoint options: %w", name, err)
		}
	}

	return resourceCloudSearchDomainUpdate(d, meta)
}

func resourceCloudSearchDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudSearchConn

	domainlist := cloudsearch.DescribeDomainsInput{
		DomainNames: []*string{
			aws.String(d.Get("name").(string)),
		},
	}

	resp, err := conn.DescribeDomains(&domainlist)
	if err != nil {
		return err
	}
	domain := resp.DomainStatusList[0]
	d.Set("id", domain.DomainId)

	if domain.DocService.Endpoint != nil {
		d.Set("document_endpoint", domain.DocService.Endpoint)
	}
	if domain.SearchService.Endpoint != nil {
		d.Set("search_endpoint", domain.SearchService.Endpoint)
	}

	// Read index fields.
	indexResults, err := conn.DescribeIndexFields(&cloudsearch.DescribeIndexFieldsInput{
		DomainName: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return err
	}

	result := make([]map[string]interface{}, 0, len(indexResults.IndexFields))

	for _, raw := range indexResults.IndexFields {
		// Don't read in any fields that are pending deletion.
		if *raw.Status.PendingDeletion {
			continue
		}

		result = append(result, readIndexField(raw.Options))
	}
	d.Set("index", result)

	// Read service access policies.
	policyResult, err := conn.DescribeServiceAccessPolicies(&cloudsearch.DescribeServiceAccessPoliciesInput{
		DomainName: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return err
	}
	d.Set("service_access_policies", policyResult.AccessPolicies.Options)

	// Read availability options (i.e. multi-az).
	availabilityResult, err := conn.DescribeAvailabilityOptions(&cloudsearch.DescribeAvailabilityOptionsInput{
		DomainName: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return err
	}
	d.Set("multi_az", availabilityResult.AvailabilityOptions.Options)

	// Read scaling parameters.
	scalingResult, err := conn.DescribeScalingParameters(&cloudsearch.DescribeScalingParametersInput{
		DomainName: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return err
	}
	d.Set("instance_type", scalingResult.ScalingParameters.Options.DesiredInstanceType)
	d.Set("partition_count", scalingResult.ScalingParameters.Options.DesiredPartitionCount)
	d.Set("replication_count", scalingResult.ScalingParameters.Options.DesiredReplicationCount)

	return err
}

func resourceCloudSearchDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudSearchConn

	_, err := conn.UpdateServiceAccessPolicies(&cloudsearch.UpdateServiceAccessPoliciesInput{
		DomainName:     aws.String(d.Get("name").(string)),
		AccessPolicies: aws.String(d.Get("service_access_policies").(string)),
	})
	if err != nil {
		return err
	}

	_, err = conn.UpdateScalingParameters(&cloudsearch.UpdateScalingParametersInput{
		DomainName: aws.String(d.Get("name").(string)),
		ScalingParameters: &cloudsearch.ScalingParameters{
			DesiredInstanceType:     aws.String(d.Get("instance_type").(string)),
			DesiredReplicationCount: aws.Int64(int64(d.Get("replication_count").(int))),
			DesiredPartitionCount:   aws.Int64(int64(d.Get("partition_count").(int))),
		},
	})
	if err != nil {
		return err
	}

	_, err = conn.UpdateAvailabilityOptions(&cloudsearch.UpdateAvailabilityOptionsInput{
		DomainName: aws.String(d.Get("name").(string)),
		MultiAZ:    aws.Bool(d.Get("multi_az").(bool)),
	})
	if err != nil {
		return err
	}

	updated, err := defineIndexFields(d, conn)
	if err != nil {
		return err
	}

	// When you add fields or modify existing fields, you must explicitly issue a request to re-index your data
	// when you are done making configuration changes.
	// https://docs.aws.amazon.com/cloudsearch/latest/developerguide/configuring-index-fields.html
	if updated {
		_, err := conn.IndexDocuments(&cloudsearch.IndexDocumentsInput{
			DomainName: aws.String(d.Get("name").(string)),
		})
		if err != nil {
			return err
		}
	}

	if d.Get("wait_for_endpoints").(bool) {
		domainsList := cloudsearch.DescribeDomainsInput{
			DomainNames: []*string{
				aws.String(d.Get("name").(string)),
			},
		}

		err = waitForSearchDomainToBeAvailable(d, conn, domainsList)
		if err != nil {
			return fmt.Errorf("%s %q", err, d.Get("name").(string))
		}
	}
	return nil
}

func resourceCloudSearchDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudSearchConn

	log.Printf("[DEBUG] Deleting CloudSearch Domain: %s", d.Id())

	_, err := conn.DeleteDomain(&cloudsearch.DeleteDomainInput{
		DomainName: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error deleting CloudSearch Domain (%s): %w", d.Id(), err)
	}

	return nil
}

func validateIndexName(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if !regexp.MustCompile(`^(\*?[a-z][a-z0-9_]{2,63}|[a-z][a-z0-9_]{2,63}\*?)$`).MatchString(value) {
		es = append(es, fmt.Errorf(
			"%q must begin with a letter and be at least 3 and no more than 64 characters long", k))
	}

	if value == "score" {
		es = append(es, fmt.Errorf("'score' is a reserved field name and cannot be used"))
	}

	return
}

// Waiters
func waitForSearchDomainToBeAvailable(d *schema.ResourceData, conn *cloudsearch.CloudSearch, domainlist cloudsearch.DescribeDomainsInput) error {
	log.Printf("[INFO] cloudsearch (%#v) waiting for domain endpoint. This usually takes 10 minutes.", domainlist.DomainNames[0])
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Waiting"},
		Target:     []string{"OK"},
		Timeout:    30 * time.Minute,
		MinTimeout: 5 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeDomains(&domainlist)
			log.Printf("[DEBUG] Checking %v", domainlist.DomainNames[0])
			if err != nil {
				log.Printf("[ERROR] Could not find domain (%v).  %s", domainlist.DomainNames[0], err)
				return nil, "", err
			}
			// Not good enough to wait for processing, have to check for search endpoint.
			domain := resp.DomainStatusList[0]
			log.Printf("[DEBUG] GLEN: Domain = %s", domain)
			processing := strconv.FormatBool(*domain.Processing)
			log.Printf("[DEBUG] GLEN: Processing = %s", processing)
			if domain.SearchService.Endpoint != nil {
				log.Printf("[DEBUG] GLEN: type: %T", domain.SearchService.Endpoint)
				log.Printf("[DEBUG] GLEN: SearchServiceEndpoint = %s", *domain.SearchService.Endpoint)
			}
			if aws.StringValue(domain.SearchService.Endpoint) == "" {
				return resp, "Waiting", nil
			}
			return resp, "OK", nil

		},
	}

	log.Printf("[DEBUG] Waiting for CloudSearch domain to finish processing: %v", domainlist.DomainNames[0])
	_, err := stateConf.WaitForState()

	// Search service was blank.
	resp, err1 := conn.DescribeDomains(&domainlist)
	if err1 != nil {
		return err1
	}

	domain := resp.DomainStatusList[0]
	d.Set("id", domain.DomainId)
	d.Set("document_endpoint", domain.DocService.Endpoint)
	d.Set("search_endpoint", domain.SearchService.Endpoint)

	if err != nil {
		return fmt.Errorf("Error waiting for CloudSearch domain (%#v) to finish processing: %s", domainlist.DomainNames[0], err)
	}
	return err
}

// Miscellaneous helper functions
func defineIndexFields(d *schema.ResourceData, conn *cloudsearch.CloudSearch) (bool, error) {
	// Early return if we don't have a change.
	if !d.HasChange("index") {
		return false, nil
	}

	o, n := d.GetChange("index")

	old := o.(*schema.Set)
	new := n.(*schema.Set)

	// Returns a set of only old fields, to be deleted.
	toDelete := old.Difference(new)
	for _, index := range toDelete.List() {
		v, _ := index.(map[string]interface{})

		_, err := conn.DeleteIndexField(&cloudsearch.DeleteIndexFieldInput{
			DomainName:     aws.String(d.Get("name").(string)),
			IndexFieldName: aws.String(v["name"].(string)),
		})
		if err != nil {
			return true, err
		}
	}

	// Returns a set of only fields that needs to be added or updated (upserted).
	toUpsert := new.Difference(old)
	for _, index := range toUpsert.List() {
		v, _ := index.(map[string]interface{})

		field, err := generateIndexFieldInput(v)
		if err != nil {
			return true, err
		}

		_, err = conn.DefineIndexField(&cloudsearch.DefineIndexFieldInput{
			DomainName: aws.String(d.Get("name").(string)),
			IndexField: field,
		})
		if err != nil {
			return true, err
		}
	}

	return true, nil
}

/*
extractFromMapToType extracts a specific value from map[string]interface{} into an interface of type
expects: map[string]interface{}, string, interface{}
returns: error
*/
func extractFromMapToType(index map[string]interface{}, property string, t interface{}) error {
	v, ok := index[property]
	if !ok {
		return fmt.Errorf("%s is not a valid property of an index", property)
	}

	if "default_value" == property {
		switch t.(type) {
		case *int:
			d, err := strconv.Atoi(v.(string))
			if err != nil {
				return parseError(v.(string), "int")
			}

			reflect.ValueOf(t).Elem().Set(reflect.ValueOf(d))
		case *float64:
			f, err := strconv.ParseFloat(v.(string), 64)
			if err != nil {
				return parseError(v.(string), "double")
			}

			reflect.ValueOf(t).Elem().Set(reflect.ValueOf(f))
		default:
			if v.(string) != "" {
				reflect.ValueOf(t).Elem().Set(reflect.ValueOf(v))
			}
		}
		return nil
	}

	reflect.ValueOf(t).Elem().Set(reflect.ValueOf(v))
	return nil
}

var parseError = func(d string, t string) error {
	return fmt.Errorf("can't convert default_value '%s' of type '%s' to int", d, t)
}

func generateIndexFieldInput(index map[string]interface{}) (*cloudsearch.IndexField, error) {
	input := &cloudsearch.IndexField{
		IndexFieldName: aws.String(index["name"].(string)),
		IndexFieldType: aws.String(index["type"].(string)),
	}

	// TODO: clean this up, this very likely could be written in a much cleaner way than this.
	var facet bool
	var returnV bool
	var search bool
	var sort bool
	var highlight bool
	var analysisScheme string

	err := extractFromMapToType(index, "facet", &facet)
	if err != nil {
		return nil, err
	}

	err = extractFromMapToType(index, "return", &returnV)
	if err != nil {
		return nil, err
	}

	err = extractFromMapToType(index, "search", &search)
	if err != nil {
		return nil, err
	}

	err = extractFromMapToType(index, "sort", &sort)
	if err != nil {
		return nil, err
	}

	err = extractFromMapToType(index, "highlight", &highlight)
	if err != nil {
		return nil, err
	}

	err = extractFromMapToType(index, "analysis_scheme", &analysisScheme)
	if err != nil {
		return nil, err
	}

	switch index["type"] {
	case "date":
		input.DateOptions = &cloudsearch.DateOptions{
			FacetEnabled:  aws.Bool(facet),
			ReturnEnabled: aws.Bool(returnV),
			SearchEnabled: aws.Bool(search),
			SortEnabled:   aws.Bool(sort),
		}

		if index["default_value"].(string) != "" {
			input.DateOptions.DefaultValue = aws.String(index["default_value"].(string))
		}
	case "date-array":
		input.DateArrayOptions = &cloudsearch.DateArrayOptions{
			FacetEnabled:  aws.Bool(facet),
			ReturnEnabled: aws.Bool(returnV),
			SearchEnabled: aws.Bool(search),
		}

		if index["default_value"].(string) != "" {
			input.DateArrayOptions.DefaultValue = aws.String(index["default_value"].(string))
		}
	case "double":
		input.DoubleOptions = &cloudsearch.DoubleOptions{
			FacetEnabled:  aws.Bool(facet),
			ReturnEnabled: aws.Bool(returnV),
			SearchEnabled: aws.Bool(search),
			SortEnabled:   aws.Bool(sort),
		}

		if index["default_value"].(string) != "" {
			var defaultValue float64
			err = extractFromMapToType(index, "default_value", &defaultValue)
			if err != nil {
				return nil, err
			}
			input.DoubleOptions.DefaultValue = aws.Float64(defaultValue)
		}
	case "double-array":
		input.DoubleArrayOptions = &cloudsearch.DoubleArrayOptions{
			FacetEnabled:  aws.Bool(facet),
			ReturnEnabled: aws.Bool(returnV),
			SearchEnabled: aws.Bool(search),
		}

		if index["default_value"].(string) != "" {
			var defaultValue float64
			err = extractFromMapToType(index, "default_value", &defaultValue)
			if err != nil {
				return nil, err
			}
			input.DoubleArrayOptions.DefaultValue = aws.Float64(defaultValue)
		}
	case "int":
		input.IntOptions = &cloudsearch.IntOptions{
			FacetEnabled:  aws.Bool(facet),
			ReturnEnabled: aws.Bool(returnV),
			SearchEnabled: aws.Bool(search),
			SortEnabled:   aws.Bool(sort),
		}

		if index["default_value"].(string) != "" {
			var defaultValue int
			err = extractFromMapToType(index, "default_value", &defaultValue)
			if err != nil {
				return nil, err
			}
			input.IntOptions.DefaultValue = aws.Int64(int64(defaultValue))
		}
	case "int-array":
		input.IntArrayOptions = &cloudsearch.IntArrayOptions{
			FacetEnabled:  aws.Bool(facet),
			ReturnEnabled: aws.Bool(returnV),
			SearchEnabled: aws.Bool(search),
		}

		if index["default_value"].(string) != "" {
			var defaultValue int
			err = extractFromMapToType(index, "default_value", &defaultValue)
			if err != nil {
				return nil, err
			}
			input.IntArrayOptions.DefaultValue = aws.Int64(int64(defaultValue))
		}
	case "latlon":
		input.LatLonOptions = &cloudsearch.LatLonOptions{
			FacetEnabled:  aws.Bool(facet),
			ReturnEnabled: aws.Bool(returnV),
			SearchEnabled: aws.Bool(search),
			SortEnabled:   aws.Bool(sort),
		}

		if index["default_value"].(string) != "" {
			input.LatLonOptions.DefaultValue = aws.String(index["default_value"].(string))
		}
	case "literal":
		input.LiteralOptions = &cloudsearch.LiteralOptions{
			FacetEnabled:  aws.Bool(facet),
			ReturnEnabled: aws.Bool(returnV),
			SearchEnabled: aws.Bool(search),
			SortEnabled:   aws.Bool(sort),
		}

		if index["default_value"].(string) != "" {
			input.LiteralOptions.DefaultValue = aws.String(index["default_value"].(string))
		}
	case "literal-array":
		input.LiteralArrayOptions = &cloudsearch.LiteralArrayOptions{
			FacetEnabled:  aws.Bool(facet),
			ReturnEnabled: aws.Bool(returnV),
			SearchEnabled: aws.Bool(search),
		}

		if index["default_value"].(string) != "" {
			input.LiteralArrayOptions.DefaultValue = aws.String(index["default_value"].(string))
		}
	case "text":
		input.TextOptions = &cloudsearch.TextOptions{
			SortEnabled:      aws.Bool(sort),
			ReturnEnabled:    aws.Bool(returnV),
			HighlightEnabled: aws.Bool(highlight),
		}

		if analysisScheme != "" {
			input.TextOptions.AnalysisScheme = aws.String(analysisScheme)
		}

		if index["default_value"].(string) != "" {
			input.TextOptions.DefaultValue = aws.String(index["default_value"].(string))
		}
	case "text-array":
		input.TextArrayOptions = &cloudsearch.TextArrayOptions{
			ReturnEnabled:    aws.Bool(returnV),
			HighlightEnabled: aws.Bool(highlight),
		}

		if analysisScheme != "" {
			input.TextArrayOptions.AnalysisScheme = aws.String(analysisScheme)
		}

		if index["default_value"].(string) != "" {
			input.TextArrayOptions.DefaultValue = aws.String(index["default_value"].(string))
		}
	default:
		return input, fmt.Errorf("invalid index field type %s", index["type"])
	}

	return input, nil
}

func readIndexField(raw *cloudsearch.IndexField) map[string]interface{} {
	index := map[string]interface{}{
		"name": raw.IndexFieldName,
		"type": raw.IndexFieldType,
	}

	switch *raw.IndexFieldType {
	case "date":
		index["default_value"] = raw.DateOptions.DefaultValue
		index["facet"] = raw.DateOptions.FacetEnabled
		index["return"] = raw.DateOptions.ReturnEnabled
		index["search"] = raw.DateOptions.SearchEnabled
		index["sort"] = raw.DateOptions.SortEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
	case "date-array":
		index["default_value"] = raw.DateArrayOptions.DefaultValue
		index["facet"] = raw.DateArrayOptions.FacetEnabled
		index["return"] = raw.DateArrayOptions.ReturnEnabled
		index["search"] = raw.DateArrayOptions.SearchEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
		index["sort"] = false
	case "double":
		index["default_value"] = raw.DoubleOptions.DefaultValue
		index["facet"] = raw.DoubleOptions.FacetEnabled
		index["return"] = raw.DoubleOptions.ReturnEnabled
		index["search"] = raw.DoubleOptions.SearchEnabled
		index["sort"] = raw.DoubleOptions.SortEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
	case "double-array":
		index["default_value"] = raw.DoubleArrayOptions.DefaultValue
		index["facet"] = raw.DoubleArrayOptions.FacetEnabled
		index["return"] = raw.DoubleArrayOptions.ReturnEnabled
		index["search"] = raw.DoubleArrayOptions.SearchEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
		index["sort"] = false
	case "int":
		index["default_value"] = raw.IntOptions.DefaultValue
		index["facet"] = raw.IntOptions.FacetEnabled
		index["return"] = raw.IntOptions.ReturnEnabled
		index["search"] = raw.IntOptions.SearchEnabled
		index["sort"] = raw.IntOptions.SortEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
	case "int-array":
		index["default_value"] = raw.IntArrayOptions.DefaultValue
		index["facet"] = raw.IntArrayOptions.FacetEnabled
		index["return"] = raw.IntArrayOptions.ReturnEnabled
		index["search"] = raw.IntArrayOptions.SearchEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
		index["sort"] = false
	case "latlon":
		index["default_value"] = raw.LatLonOptions.DefaultValue
		index["facet"] = raw.LatLonOptions.FacetEnabled
		index["return"] = raw.LatLonOptions.ReturnEnabled
		index["search"] = raw.LatLonOptions.SearchEnabled
		index["sort"] = raw.LatLonOptions.SortEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
	case "literal":
		index["default_value"] = raw.LiteralOptions.DefaultValue
		index["facet"] = raw.LiteralOptions.FacetEnabled
		index["return"] = raw.LiteralOptions.ReturnEnabled
		index["search"] = raw.LiteralOptions.SearchEnabled
		index["sort"] = raw.LiteralOptions.SortEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
	case "literal-array":
		index["default_value"] = raw.LiteralArrayOptions.DefaultValue
		index["facet"] = raw.LiteralArrayOptions.FacetEnabled
		index["return"] = raw.LiteralArrayOptions.ReturnEnabled
		index["search"] = raw.LiteralArrayOptions.SearchEnabled

		// Options that aren't valid for this type.
		index["highlight"] = false
		index["sort"] = false
	case "text":
		index["default_value"] = raw.TextOptions.DefaultValue
		index["analysis_scheme"] = raw.TextOptions.AnalysisScheme
		index["highlight"] = raw.TextOptions.HighlightEnabled
		index["return"] = raw.TextOptions.ReturnEnabled
		index["sort"] = raw.TextOptions.SortEnabled

		// Options that aren't valid for this type.
		index["facet"] = false
		index["search"] = false
	case "text-array":
		index["default_value"] = raw.TextArrayOptions.DefaultValue
		index["analysis_scheme"] = raw.TextArrayOptions.AnalysisScheme
		index["highlight"] = raw.TextArrayOptions.HighlightEnabled
		index["return"] = raw.TextArrayOptions.ReturnEnabled

		// Options that aren't valid for this type.
		index["facet"] = false
		index["search"] = false
		index["sort"] = false
	}

	return index
}

func FindDomainByName(conn *cloudsearch.CloudSearch, name string) (*cloudsearch.DomainStatus, error) {
	input := &cloudsearch.DescribeDomainsInput{
		DomainNames: aws.StringSlice([]string{name}),
	}

	output, err := conn.DescribeDomains(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DomainStatusList) == 0 || output.DomainStatusList[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.DomainStatusList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	domainStatus := output.DomainStatusList[0]

	if aws.BoolValue(domainStatus.Deleted) {
		return nil, &resource.NotFoundError{
			Message:     "deleted",
			LastRequest: input,
		}
	}

	return domainStatus, nil
}

func expandDomainEndpointOptions(tfMap map[string]interface{}) *cloudsearch.DomainEndpointOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudsearch.DomainEndpointOptions{}

	if v, ok := tfMap["enforce_https"].(bool); ok {
		apiObject.EnforceHTTPS = aws.Bool(v)
	}

	if v, ok := tfMap["tls_security_policy"].(string); ok && v != "" {
		apiObject.TLSSecurityPolicy = aws.String(v)
	}

	return apiObject
}

func flattenDomainEndpointOptions(apiObject *cloudsearch.DomainEndpointOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EnforceHTTPS; v != nil {
		tfMap["enforce_https"] = aws.BoolValue(v)
	}

	if v := apiObject.TLSSecurityPolicy; v != nil {
		tfMap["tls_security_policy"] = aws.StringValue(v)
	}

	return tfMap
}

func expandScalingParameters(tfMap map[string]interface{}) *cloudsearch.ScalingParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudsearch.ScalingParameters{}

	if v, ok := tfMap["desired_instance_type"].(string); ok && v != "" {
		apiObject.DesiredInstanceType = aws.String(v)
	}

	if v, ok := tfMap["desired_partition_count"].(int); ok && v != 0 {
		apiObject.DesiredPartitionCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["desired_replication_count"].(int); ok && v != 0 {
		apiObject.DesiredReplicationCount = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenScalingParameters(apiObject *cloudsearch.ScalingParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DesiredInstanceType; v != nil {
		tfMap["desired_instance_type"] = aws.StringValue(v)
	}

	if v := apiObject.DesiredPartitionCount; v != nil {
		tfMap["desired_partition_count"] = aws.Int64Value(v)
	}

	if v := apiObject.DesiredReplicationCount; v != nil {
		tfMap["desired_replication_count"] = aws.Int64Value(v)
	}

	return tfMap
}
