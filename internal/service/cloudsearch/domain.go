package cloudsearch

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainCreate,
		Read:   resourceDomainRead,
		Update: resourceDomainUpdate,
		Delete: resourceDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"document_service_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_options": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enforce_https": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"tls_security_policy": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(cloudsearch.TLSSecurityPolicy_Values(), false),
						},
					},
				},
			},
			// The index_field schema is based on the AWS Console screen, not the API model.
			"index_field": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"analysis_scheme": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"default_value": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"facet": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"highlight": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateIndexName,
						},
						"return": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"search": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"sort": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"source_fields": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringDoesNotMatch(regexp.MustCompile(`score`), "Cannot be set to reserved field score"),
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(cloudsearch.IndexFieldType_Values(), false),
						},
					},
				},
			},
			"multi_az": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
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
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_instance_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(cloudsearch.PartitionInstanceType_Values(), false),
						},
						"desired_partition_count": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"desired_replication_count": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"search_service_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDomainCreate(d *schema.ResourceData, meta interface{}) error {
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

	if v, ok := d.GetOk("scaling_parameters"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := &cloudsearch.UpdateScalingParametersInput{
			DomainName:        aws.String(d.Id()),
			ScalingParameters: expandScalingParameters(v.([]interface{})[0].(map[string]interface{})),
		}

		log.Printf("[DEBUG] Updating CloudSearch Domain scaling parameters: %s", input)
		_, err := conn.UpdateScalingParameters(input)

		if err != nil {
			return fmt.Errorf("error updating CloudSearch Domain (%s) scaling parameters: %w", d.Id(), err)
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
			return fmt.Errorf("error updating CloudSearch Domain (%s) availability options: %w", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("endpoint_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := &cloudsearch.UpdateDomainEndpointOptionsInput{
			DomainEndpointOptions: expandDomainEndpointOptions(v.([]interface{})[0].(map[string]interface{})),
			DomainName:            aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating CloudSearch Domain endpoint options: %s", input)
		_, err := conn.UpdateDomainEndpointOptions(input)

		if err != nil {
			return fmt.Errorf("error updating CloudSearch Domain (%s) endpoint options: %w", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("index_field"); ok && v.(*schema.Set).Len() > 0 {
		err := defineIndexFields(conn, d.Id(), v.(*schema.Set).List())

		if err != nil {
			return err
		}

		log.Printf("[DEBUG] Indexing CloudSearch Domain documents: %s", d.Id())
		_, err = conn.IndexDocuments(&cloudsearch.IndexDocumentsInput{
			DomainName: aws.String(d.Id()),
		})

		if err != nil {
			return fmt.Errorf("error indexing CloudSearch Domain (%s) documents: %w", d.Id(), err)
		}
	}

	// TODO: Status.RequiresIndexDocuments = true?

	_, err = waitDomainActive(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for CloudSearch Domain (%s) create: %w", d.Id(), err)
	}

	return resourceDomainRead(d, meta)
}

func resourceDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudSearchConn

	domainStatus, err := FindDomainStatusByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudSearch Domain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudSearch Domain (%s): %w", d.Id(), err)
	}

	d.Set("arn", domainStatus.ARN)
	d.Set("domain_id", domainStatus.DomainId)
	d.Set("name", domainStatus.DomainName)

	if domainStatus.DocService != nil {
		d.Set("document_service_endpoint", domainStatus.DocService.Endpoint)
	} else {
		d.Set("document_service_endpoint", nil)
	}
	if domainStatus.SearchService != nil {
		d.Set("search_service_endpoint", domainStatus.SearchService.Endpoint)
	} else {
		d.Set("search_service_endpoint", nil)
	}

	availabilityOptionStatus, err := findAvailabilityOptionsStatusByName(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading CloudSearch Domain (%s) availability options: %w", d.Id(), err)
	}

	d.Set("multi_az", availabilityOptionStatus.Options)

	endpointOptions, err := findDomainEndpointOptionsByName(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading CloudSearch Domain (%s) endpoint options: %w", d.Id(), err)
	}

	if err := d.Set("endpoint_options", []interface{}{flattenDomainEndpointOptions(endpointOptions)}); err != nil {
		return fmt.Errorf("error setting endpoint_options: %w", err)
	}

	scalingParameters, err := findScalingParametersByName(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading CloudSearch Domain (%s) scaling parameters: %w", d.Id(), err)
	}

	if err := d.Set("scaling_parameters", []interface{}{flattenScalingParameters(scalingParameters)}); err != nil {
		return fmt.Errorf("error setting scaling_parameters: %w", err)
	}

	indexResults, err := conn.DescribeIndexFields(&cloudsearch.DescribeIndexFieldsInput{
		DomainName: aws.String(d.Get("name").(string)),
	})

	if err != nil {
		return fmt.Errorf("error reading CloudSearch Domain (%s) index fields: %w", d.Id(), err)
	}

	if tfList, err := flattenIndexFieldStatuses(indexResults.IndexFields); err != nil {
		return err
	} else if err := d.Set("index_field", tfList); err != nil {
		return fmt.Errorf("error setting index_field: %w", err)
	}

	return nil
}

func resourceDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudSearchConn
	requiresIndexDocuments := false

	if d.HasChange("scaling_parameters") {
		input := &cloudsearch.UpdateScalingParametersInput{
			DomainName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("scaling_parameters"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ScalingParameters = expandScalingParameters(v.([]interface{})[0].(map[string]interface{}))
		} else {
			input.ScalingParameters = &cloudsearch.ScalingParameters{}
		}

		log.Printf("[DEBUG] Updating CloudSearch Domain scaling parameters: %s", input)
		output, err := conn.UpdateScalingParameters(input)

		if err != nil {
			return fmt.Errorf("error updating CloudSearch Domain (%s) scaling parameters: %w", d.Id(), err)
		}

		if output != nil && output.ScalingParameters != nil && output.ScalingParameters.Status != nil && aws.StringValue(output.ScalingParameters.Status.State) == cloudsearch.OptionStateRequiresIndexDocuments {
			requiresIndexDocuments = true
		}
	}

	if d.HasChange("multi_az") {
		input := &cloudsearch.UpdateAvailabilityOptionsInput{
			DomainName: aws.String(d.Id()),
			MultiAZ:    aws.Bool(d.Get("multi_az").(bool)),
		}

		log.Printf("[DEBUG] Updating CloudSearch Domain availability options: %s", input)
		output, err := conn.UpdateAvailabilityOptions(input)

		if err != nil {
			return fmt.Errorf("error updating CloudSearch Domain (%s) availability options: %w", d.Id(), err)
		}

		if output != nil && output.AvailabilityOptions != nil && output.AvailabilityOptions.Status != nil && aws.StringValue(output.AvailabilityOptions.Status.State) == cloudsearch.OptionStateRequiresIndexDocuments {
			requiresIndexDocuments = true
		}
	}

	if d.HasChange("endpoint_options") {
		input := &cloudsearch.UpdateDomainEndpointOptionsInput{
			DomainName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("endpoint_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.DomainEndpointOptions = expandDomainEndpointOptions(v.([]interface{})[0].(map[string]interface{}))
		} else {
			input.DomainEndpointOptions = &cloudsearch.DomainEndpointOptions{}
		}

		log.Printf("[DEBUG] Updating CloudSearch Domain endpoint options: %s", input)
		output, err := conn.UpdateDomainEndpointOptions(input)

		if err != nil {
			return fmt.Errorf("error updating CloudSearch Domain (%s) endpoint options: %w", d.Id(), err)
		}

		if output != nil && output.DomainEndpointOptions != nil && output.DomainEndpointOptions.Status != nil && aws.StringValue(output.DomainEndpointOptions.Status.State) == cloudsearch.OptionStateRequiresIndexDocuments {
			requiresIndexDocuments = true
		}
	}

	if d.HasChange("index_field") {
		o, n := d.GetChange("index_field")
		old := o.(*schema.Set)
		new := n.(*schema.Set)

		for _, tfMapRaw := range old.Difference(new).List() {
			tfMap, ok := tfMapRaw.(map[string]interface{})

			if !ok {
				continue
			}

			fieldName, _ := tfMap["name"].(string)

			if fieldName == "" {
				continue
			}

			input := &cloudsearch.DeleteIndexFieldInput{
				DomainName:     aws.String(d.Id()),
				IndexFieldName: aws.String(fieldName),
			}

			log.Printf("[DEBUG] Deleting CloudSearch Domain index field: %s", input)
			_, err := conn.DeleteIndexField(input)

			if err != nil {
				return fmt.Errorf("error deleting CloudSearch Domain (%s) index field (%s): %w", d.Id(), fieldName, err)
			}

			requiresIndexDocuments = true
		}

		if v := new.Difference(old); v.Len() > 0 {
			if err := defineIndexFields(conn, d.Id(), v.List()); err != nil {
				return err
			}

			requiresIndexDocuments = true
		}
	}

	if requiresIndexDocuments {
		log.Printf("[DEBUG] Indexing CloudSearch Domain documents: %s", d.Id())
		_, err := conn.IndexDocuments(&cloudsearch.IndexDocumentsInput{
			DomainName: aws.String(d.Id()),
		})

		if err != nil {
			return fmt.Errorf("error indexing CloudSearch Domain (%s) documents: %w", d.Id(), err)
		}
	}

	_, err := waitDomainActive(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return fmt.Errorf("error waiting for CloudSearch Domain (%s) update: %w", d.Id(), err)
	}

	return resourceDomainRead(d, meta)
}

func resourceDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudSearchConn

	log.Printf("[DEBUG] Deleting CloudSearch Domain: %s", d.Id())
	_, err := conn.DeleteDomain(&cloudsearch.DeleteDomainInput{
		DomainName: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error deleting CloudSearch Domain (%s): %w", d.Id(), err)
	}

	_, err = waitDomainDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for CloudSearch Domain (%s) delete: %w", d.Id(), err)
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

func defineIndexFields(conn *cloudsearch.CloudSearch, domainName string, tfList []interface{}) error {
	// Define index fields with source fields after those without.
	for _, defineWhenSourceFieldsConfigured := range []bool{false, true} {
		for _, tfMapRaw := range tfList {
			tfMap, ok := tfMapRaw.(map[string]interface{})

			if !ok {
				continue
			}

			apiObject, sourceFieldsConfigured, err := expandIndexField(tfMap)

			if err != nil {
				return err
			}

			if apiObject == nil {
				continue
			}

			if sourceFieldsConfigured && !defineWhenSourceFieldsConfigured {
				continue
			}

			if !sourceFieldsConfigured && defineWhenSourceFieldsConfigured {
				continue
			}

			input := &cloudsearch.DefineIndexFieldInput{
				DomainName: aws.String(domainName),
				IndexField: apiObject,
			}

			log.Printf("[DEBUG] Defining CloudSearch Domain index field: %s", input)
			_, err = conn.DefineIndexField(input)

			if err != nil {
				return fmt.Errorf("error defining CloudSearch Domain (%s) index field (%s): %w", domainName, aws.StringValue(apiObject.IndexFieldName), err)
			}
		}
	}

	return nil
}

func FindDomainStatusByName(conn *cloudsearch.CloudSearch, name string) (*cloudsearch.DomainStatus, error) {
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

	return output.DomainStatusList[0], nil
}

func findAvailabilityOptionsStatusByName(conn *cloudsearch.CloudSearch, name string) (*cloudsearch.AvailabilityOptionsStatus, error) {
	input := &cloudsearch.DescribeAvailabilityOptionsInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeAvailabilityOptions(input)

	if tfawserr.ErrCodeEquals(err, cloudsearch.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AvailabilityOptions == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AvailabilityOptions, nil
}

func findDomainEndpointOptionsByName(conn *cloudsearch.CloudSearch, name string) (*cloudsearch.DomainEndpointOptions, error) {
	output, err := findDomainEndpointOptionsStatusByName(conn, name)

	if err != nil {
		return nil, err
	}

	if output.Options == nil {
		return nil, tfresource.NewEmptyResultError(name)
	}

	return output.Options, nil
}

func findDomainEndpointOptionsStatusByName(conn *cloudsearch.CloudSearch, name string) (*cloudsearch.DomainEndpointOptionsStatus, error) {
	input := &cloudsearch.DescribeDomainEndpointOptionsInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeDomainEndpointOptions(input)

	if tfawserr.ErrCodeEquals(err, cloudsearch.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DomainEndpointOptions == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DomainEndpointOptions, nil
}

func findScalingParametersByName(conn *cloudsearch.CloudSearch, name string) (*cloudsearch.ScalingParameters, error) {
	output, err := findScalingParametersStatusByName(conn, name)

	if err != nil {
		return nil, err
	}

	if output.Options == nil {
		return nil, tfresource.NewEmptyResultError(name)
	}

	return output.Options, nil
}

func findScalingParametersStatusByName(conn *cloudsearch.CloudSearch, name string) (*cloudsearch.ScalingParametersStatus, error) {
	input := &cloudsearch.DescribeScalingParametersInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeScalingParameters(input)

	if tfawserr.ErrCodeEquals(err, cloudsearch.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ScalingParameters == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ScalingParameters, nil
}

func statusDomainDeleting(conn *cloudsearch.CloudSearch, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDomainStatusByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.Deleted)), nil
	}
}

func statusDomainProcessing(conn *cloudsearch.CloudSearch, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDomainStatusByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.Processing)), nil
	}
}

func waitDomainActive(conn *cloudsearch.CloudSearch, name string, timeout time.Duration) (*cloudsearch.DomainStatus, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{"true"},
		Target:  []string{"false"},
		Refresh: statusDomainProcessing(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*cloudsearch.DomainStatus); ok {
		return output, err
	}

	return nil, err
}

func waitDomainDeleted(conn *cloudsearch.CloudSearch, name string, timeout time.Duration) (*cloudsearch.DomainStatus, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"true"},
		Target:  []string{},
		Refresh: statusDomainDeleting(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*cloudsearch.DomainStatus); ok {
		return output, err
	}

	return nil, err
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

func expandIndexField(tfMap map[string]interface{}) (*cloudsearch.IndexField, bool, error) {
	if tfMap == nil {
		return nil, false, nil
	}

	apiObject := &cloudsearch.IndexField{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.IndexFieldName = aws.String(v)
	}

	fieldType, ok := tfMap["type"].(string)
	if ok && fieldType != "" {
		apiObject.IndexFieldType = aws.String(fieldType)
	}

	analysisScheme, _ := tfMap["analysis_scheme"].(string)
	facetEnabled, _ := tfMap["facet"].(bool)
	highlightEnabled, _ := tfMap["highlight"].(bool)
	returnEnabled, _ := tfMap["return"].(bool)
	searchEnabled, _ := tfMap["search"].(bool)
	sortEnabled, _ := tfMap["sort"].(bool)
	var sourceFieldsConfigured bool

	switch fieldType {
	case cloudsearch.IndexFieldTypeDate:
		options := &cloudsearch.DateOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if v, ok := tfMap["default_value"].(string); ok && v != "" {
			options.DefaultValue = aws.String(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceField = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.DateOptions = options

	case cloudsearch.IndexFieldTypeDateArray:
		options := &cloudsearch.DateArrayOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
		}

		if v, ok := tfMap["default_value"].(string); ok && v != "" {
			options.DefaultValue = aws.String(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceFields = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.DateArrayOptions = options

	case cloudsearch.IndexFieldTypeDouble:
		options := &cloudsearch.DoubleOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if v, ok := tfMap["default_value"].(string); ok && v != "" {
			v, err := strconv.ParseFloat(v, 64)

			if err != nil {
				return nil, false, err
			}

			options.DefaultValue = aws.Float64(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceField = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.DoubleOptions = options

	case cloudsearch.IndexFieldTypeDoubleArray:
		options := &cloudsearch.DoubleArrayOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
		}

		if v, ok := tfMap["default_value"].(string); ok && v != "" {
			v, err := strconv.ParseFloat(v, 64)

			if err != nil {
				return nil, false, err
			}

			options.DefaultValue = aws.Float64(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceFields = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.DoubleArrayOptions = options

	case cloudsearch.IndexFieldTypeInt:
		options := &cloudsearch.IntOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if v, ok := tfMap["default_value"].(string); ok && v != "" {
			v, err := strconv.Atoi(v)

			if err != nil {
				return nil, false, err
			}

			options.DefaultValue = aws.Int64(int64(v))
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceField = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.IntOptions = options

	case cloudsearch.IndexFieldTypeIntArray:
		options := &cloudsearch.IntArrayOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
		}

		if v, ok := tfMap["default_value"].(string); ok && v != "" {
			v, err := strconv.Atoi(v)

			if err != nil {
				return nil, false, err
			}

			options.DefaultValue = aws.Int64(int64(v))
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceFields = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.IntArrayOptions = options

	case cloudsearch.IndexFieldTypeLatlon:
		options := &cloudsearch.LatLonOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if v, ok := tfMap["default_value"].(string); ok && v != "" {
			options.DefaultValue = aws.String(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceField = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.LatLonOptions = options

	case cloudsearch.IndexFieldTypeLiteral:
		options := &cloudsearch.LiteralOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if v, ok := tfMap["default_value"].(string); ok && v != "" {
			options.DefaultValue = aws.String(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceField = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.LiteralOptions = options

	case cloudsearch.IndexFieldTypeLiteralArray:
		options := &cloudsearch.LiteralArrayOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
		}

		if v, ok := tfMap["default_value"].(string); ok && v != "" {
			options.DefaultValue = aws.String(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceFields = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.LiteralArrayOptions = options

	case cloudsearch.IndexFieldTypeText:
		options := &cloudsearch.TextOptions{
			HighlightEnabled: aws.Bool(highlightEnabled),
			ReturnEnabled:    aws.Bool(returnEnabled),
			SortEnabled:      aws.Bool(sortEnabled),
		}

		if analysisScheme != "" {
			options.AnalysisScheme = aws.String(analysisScheme)
		}

		if v, ok := tfMap["default_value"].(string); ok && v != "" {
			options.DefaultValue = aws.String(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceField = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.TextOptions = options

	case cloudsearch.IndexFieldTypeTextArray:
		options := &cloudsearch.TextArrayOptions{
			HighlightEnabled: aws.Bool(highlightEnabled),
			ReturnEnabled:    aws.Bool(returnEnabled),
		}

		if analysisScheme != "" {
			options.AnalysisScheme = aws.String(analysisScheme)
		}

		if v, ok := tfMap["default_value"].(string); ok && v != "" {
			options.DefaultValue = aws.String(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceFields = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.TextArrayOptions = options

	default:
		return nil, false, fmt.Errorf("unsupported index_field type: %s", fieldType)
	}

	return apiObject, sourceFieldsConfigured, nil
}

func flattenIndexFieldStatus(apiObject *cloudsearch.IndexFieldStatus) (map[string]interface{}, error) {
	if apiObject == nil || apiObject.Options == nil || apiObject.Status == nil {
		return nil, nil
	}

	// Don't read in any fields that are pending deletion.
	if aws.BoolValue(apiObject.Status.PendingDeletion) {
		return nil, nil
	}

	field := apiObject.Options
	tfMap := map[string]interface{}{}

	if v := field.IndexFieldName; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	fieldType := field.IndexFieldType
	if fieldType != nil {
		tfMap["type"] = aws.StringValue(fieldType)
	}

	switch fieldType := aws.StringValue(fieldType); fieldType {
	case cloudsearch.IndexFieldTypeDate:
		options := field.DateOptions

		if v := options.DefaultValue; v != nil {
			tfMap["default_value"] = aws.StringValue(v)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.BoolValue(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.BoolValue(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.BoolValue(v)
		}

		if v := options.SortEnabled; v != nil {
			tfMap["sort"] = aws.BoolValue(v)
		}

		if v := options.SourceField; v != nil {
			tfMap["source_fields"] = aws.StringValue(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false

	case cloudsearch.IndexFieldTypeDateArray:
		options := field.DateArrayOptions

		if v := options.DefaultValue; v != nil {
			tfMap["default_value"] = aws.StringValue(v)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.BoolValue(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.BoolValue(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.BoolValue(v)
		}

		if v := options.SourceFields; v != nil {
			tfMap["source_fields"] = aws.StringValue(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false
		tfMap["sort"] = false

	case cloudsearch.IndexFieldTypeDouble:
		options := field.DoubleOptions

		if v := options.DefaultValue; v != nil {
			tfMap["default_value"] = strconv.FormatFloat(aws.Float64Value(v), 'f', -1, 64)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.BoolValue(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.BoolValue(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.BoolValue(v)
		}

		if v := options.SortEnabled; v != nil {
			tfMap["sort"] = aws.BoolValue(v)
		}

		if v := options.SourceField; v != nil {
			tfMap["source_fields"] = aws.StringValue(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false

	case cloudsearch.IndexFieldTypeDoubleArray:
		options := field.DoubleArrayOptions

		if v := options.DefaultValue; v != nil {
			tfMap["default_value"] = strconv.FormatFloat(aws.Float64Value(v), 'f', -1, 64)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.BoolValue(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.BoolValue(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.BoolValue(v)
		}

		if v := options.SourceFields; v != nil {
			tfMap["source_fields"] = aws.StringValue(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false
		tfMap["sort"] = false

	case cloudsearch.IndexFieldTypeInt:
		options := field.IntOptions

		if v := options.DefaultValue; v != nil {
			tfMap["default_value"] = strconv.FormatInt(aws.Int64Value(v), 10)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.BoolValue(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.BoolValue(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.BoolValue(v)
		}

		if v := options.SortEnabled; v != nil {
			tfMap["sort"] = aws.BoolValue(v)
		}

		if v := options.SourceField; v != nil {
			tfMap["source_fields"] = aws.StringValue(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false

	case cloudsearch.IndexFieldTypeIntArray:
		options := field.IntArrayOptions

		if v := options.DefaultValue; v != nil {
			tfMap["default_value"] = strconv.FormatInt(aws.Int64Value(v), 10)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.BoolValue(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.BoolValue(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.BoolValue(v)
		}

		if v := options.SourceFields; v != nil {
			tfMap["source_fields"] = aws.StringValue(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false
		tfMap["sort"] = false

	case cloudsearch.IndexFieldTypeLatlon:
		options := field.LatLonOptions

		if v := options.DefaultValue; v != nil {
			tfMap["default_value"] = aws.StringValue(v)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.BoolValue(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.BoolValue(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.BoolValue(v)
		}

		if v := options.SortEnabled; v != nil {
			tfMap["sort"] = aws.BoolValue(v)
		}

		if v := options.SourceField; v != nil {
			tfMap["source_fields"] = aws.StringValue(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false

	case cloudsearch.IndexFieldTypeLiteral:
		options := field.LiteralOptions

		if v := options.DefaultValue; v != nil {
			tfMap["default_value"] = aws.StringValue(v)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.BoolValue(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.BoolValue(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.BoolValue(v)
		}

		if v := options.SortEnabled; v != nil {
			tfMap["sort"] = aws.BoolValue(v)
		}

		if v := options.SourceField; v != nil {
			tfMap["source_fields"] = aws.StringValue(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false

	case cloudsearch.IndexFieldTypeLiteralArray:
		options := field.LiteralArrayOptions

		if v := options.DefaultValue; v != nil {
			tfMap["default_value"] = aws.StringValue(v)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.BoolValue(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.BoolValue(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.BoolValue(v)
		}

		if v := options.SourceFields; v != nil {
			tfMap["source_fields"] = aws.StringValue(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false
		tfMap["sort"] = false

	case cloudsearch.IndexFieldTypeText:
		options := field.TextOptions

		if v := options.AnalysisScheme; v != nil {
			tfMap["analysis_scheme"] = aws.StringValue(v)
		}

		if v := options.DefaultValue; v != nil {
			tfMap["default_value"] = aws.StringValue(v)
		}

		if v := options.HighlightEnabled; v != nil {
			tfMap["highlight"] = aws.BoolValue(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.BoolValue(v)
		}

		if v := options.SortEnabled; v != nil {
			tfMap["sort"] = aws.BoolValue(v)
		}

		if v := options.SourceField; v != nil {
			tfMap["source_fields"] = aws.StringValue(v)
		}

		// Defaults not returned via the API.
		tfMap["facet"] = false
		tfMap["search"] = true

	case cloudsearch.IndexFieldTypeTextArray:
		options := field.TextArrayOptions

		if v := options.AnalysisScheme; v != nil {
			tfMap["analysis_scheme"] = aws.StringValue(v)
		}

		if v := options.DefaultValue; v != nil {
			tfMap["default_value"] = aws.StringValue(v)
		}

		if v := options.HighlightEnabled; v != nil {
			tfMap["highlight"] = aws.BoolValue(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.BoolValue(v)
		}

		if v := options.SourceFields; v != nil {
			tfMap["source_fields"] = aws.StringValue(v)
		}

		// Defaults not returned via the API.
		tfMap["facet"] = false
		tfMap["search"] = true
		tfMap["sort"] = false

	default:
		return nil, fmt.Errorf("unsupported index_field type: %s", fieldType)
	}

	return tfMap, nil
}

func flattenIndexFieldStatuses(apiObjects []*cloudsearch.IndexFieldStatus) ([]interface{}, error) {
	if len(apiObjects) == 0 {
		return nil, nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfMap, err := flattenIndexFieldStatus(apiObject)

		if err != nil {
			return nil, err
		}

		tfList = append(tfList, tfMap)
	}

	return tfList, nil
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
