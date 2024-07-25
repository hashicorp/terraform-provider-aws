// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudsearch

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudsearch"
	"github.com/aws/aws-sdk-go-v2/service/cloudsearch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudsearch_domain", name="Domain")
func resourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainCreate,
		ReadWithoutTimeout:   resourceDomainRead,
		UpdateWithoutTimeout: resourceDomainUpdate,
		DeleteWithoutTimeout: resourceDomainDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[types.TLSSecurityPolicy](),
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
						names.AttrDefaultValue: {
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
						names.AttrName: {
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
							ValidateFunc: validation.StringDoesNotMatch(regexache.MustCompile(`score`), "Cannot be set to reserved field score"),
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.IndexFieldType](),
						},
					},
				},
			},
			"multi_az": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-z]([0-9a-z-]){2,27}$`), "Search domain names must start with a lowercase letter (a-z) and be at least 3 and no more than 28 lower-case letters, digits or hyphens"),
			},
			"scaling_parameters": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_instance_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[types.PartitionInstanceType](),
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

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudSearchClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &cloudsearch.CreateDomainInput{
		DomainName: aws.String(name),
	}
	_, err := conn.CreateDomain(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudSearch Domain (%s): %s", name, err)
	}

	d.SetId(name)

	if v, ok := d.GetOk("scaling_parameters"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := &cloudsearch.UpdateScalingParametersInput{
			DomainName:        aws.String(d.Id()),
			ScalingParameters: expandScalingParameters(v.([]interface{})[0].(map[string]interface{})),
		}

		_, err := conn.UpdateScalingParameters(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudSearch Domain (%s) scaling parameters: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("multi_az"); ok {
		input := &cloudsearch.UpdateAvailabilityOptionsInput{
			DomainName: aws.String(d.Id()),
			MultiAZ:    aws.Bool(v.(bool)),
		}

		_, err := conn.UpdateAvailabilityOptions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudSearch Domain (%s) availability options: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("endpoint_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := &cloudsearch.UpdateDomainEndpointOptionsInput{
			DomainEndpointOptions: expandDomainEndpointOptions(v.([]interface{})[0].(map[string]interface{})),
			DomainName:            aws.String(d.Id()),
		}

		_, err := conn.UpdateDomainEndpointOptions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudSearch Domain (%s) endpoint options: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("index_field"); ok && v.(*schema.Set).Len() > 0 {
		err := defineIndexFields(ctx, conn, d.Id(), v.(*schema.Set).List())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating CloudSearch Domain (%s): %s", name, err)
		}

		input := &cloudsearch.IndexDocumentsInput{
			DomainName: aws.String(d.Id()),
		}

		_, err = conn.IndexDocuments(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "indexing CloudSearch Domain (%s) documents: %s", d.Id(), err)
		}
	}

	// TODO: Status.RequiresIndexDocuments = true?

	if _, err := waitDomainActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudSearch Domain (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudSearchClient(ctx)

	domain, err := findDomainByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudSearch Domain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudSearch Domain (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, domain.ARN)
	if domain.DocService != nil {
		d.Set("document_service_endpoint", domain.DocService.Endpoint)
	} else {
		d.Set("document_service_endpoint", nil)
	}
	d.Set("domain_id", domain.DomainId)
	d.Set(names.AttrName, domain.DomainName)
	if domain.SearchService != nil {
		d.Set("search_service_endpoint", domain.SearchService.Endpoint)
	} else {
		d.Set("search_service_endpoint", nil)
	}

	availabilityOptionStatus, err := findAvailabilityOptionsStatusByName(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudSearch Domain (%s) availability options: %s", d.Id(), err)
	}

	d.Set("multi_az", availabilityOptionStatus.Options)

	endpointOptions, err := findDomainEndpointOptionsByName(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudSearch Domain (%s) endpoint options: %s", d.Id(), err)
	}

	if err := d.Set("endpoint_options", []interface{}{flattenDomainEndpointOptions(endpointOptions)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint_options: %s", err)
	}

	scalingParameters, err := findScalingParametersByName(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudSearch Domain (%s) scaling parameters: %s", d.Id(), err)
	}

	if err := d.Set("scaling_parameters", []interface{}{flattenScalingParameters(scalingParameters)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting scaling_parameters: %s", err)
	}

	indexResults, err := conn.DescribeIndexFields(ctx, &cloudsearch.DescribeIndexFieldsInput{
		DomainName: aws.String(d.Get(names.AttrName).(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudSearch Domain (%s) index fields: %s", d.Id(), err)
	}

	if tfList, err := flattenIndexFieldStatuses(indexResults.IndexFields); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudSearch Domain (%s): %s", d.Id(), err)
	} else if err := d.Set("index_field", tfList); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting index_field: %s", err)
	}

	return diags
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudSearchClient(ctx)

	requiresIndexDocuments := false
	if d.HasChange("scaling_parameters") {
		input := &cloudsearch.UpdateScalingParametersInput{
			DomainName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("scaling_parameters"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ScalingParameters = expandScalingParameters(v.([]interface{})[0].(map[string]interface{}))
		} else {
			input.ScalingParameters = &types.ScalingParameters{}
		}

		output, err := conn.UpdateScalingParameters(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudSearch Domain (%s) scaling parameters: %s", d.Id(), err)
		}

		if output != nil && output.ScalingParameters != nil && output.ScalingParameters.Status != nil && output.ScalingParameters.Status.State == types.OptionStateRequiresIndexDocuments {
			requiresIndexDocuments = true
		}
	}

	if d.HasChange("multi_az") {
		input := &cloudsearch.UpdateAvailabilityOptionsInput{
			DomainName: aws.String(d.Id()),
			MultiAZ:    aws.Bool(d.Get("multi_az").(bool)),
		}

		output, err := conn.UpdateAvailabilityOptions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudSearch Domain (%s) availability options: %s", d.Id(), err)
		}

		if output != nil && output.AvailabilityOptions != nil && output.AvailabilityOptions.Status != nil && output.AvailabilityOptions.Status.State == types.OptionStateRequiresIndexDocuments {
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
			input.DomainEndpointOptions = &types.DomainEndpointOptions{}
		}

		output, err := conn.UpdateDomainEndpointOptions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudSearch Domain (%s) endpoint options: %s", d.Id(), err)
		}

		if output != nil && output.DomainEndpointOptions != nil && output.DomainEndpointOptions.Status != nil && output.DomainEndpointOptions.Status.State == types.OptionStateRequiresIndexDocuments {
			requiresIndexDocuments = true
		}
	}

	if d.HasChange("index_field") {
		o, n := d.GetChange("index_field")
		os, ns := o.(*schema.Set), n.(*schema.Set)

		for _, tfMapRaw := range os.Difference(ns).List() {
			tfMap, ok := tfMapRaw.(map[string]interface{})
			if !ok {
				continue
			}

			fieldName, _ := tfMap[names.AttrName].(string)

			if fieldName == "" {
				continue
			}

			input := &cloudsearch.DeleteIndexFieldInput{
				DomainName:     aws.String(d.Id()),
				IndexFieldName: aws.String(fieldName),
			}

			_, err := conn.DeleteIndexField(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting CloudSearch Domain (%s) index field (%s): %s", d.Id(), fieldName, err)
			}

			requiresIndexDocuments = true
		}

		if v := ns.Difference(os); v.Len() > 0 {
			if err := defineIndexFields(ctx, conn, d.Id(), v.List()); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating CloudSearch Domain (%s): %s", d.Id(), err)
			}

			requiresIndexDocuments = true
		}
	}

	if requiresIndexDocuments {
		input := &cloudsearch.IndexDocumentsInput{
			DomainName: aws.String(d.Id()),
		}

		_, err := conn.IndexDocuments(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "indexing CloudSearch Domain (%s) documents: %s", d.Id(), err)
		}
	}

	if _, err := waitDomainActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudSearch Domain (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudSearchClient(ctx)

	log.Printf("[DEBUG] Deleting CloudSearch Domain: %s", d.Id())
	_, err := conn.DeleteDomain(ctx, &cloudsearch.DeleteDomainInput{
		DomainName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudSearch Domain (%s): %s", d.Id(), err)
	}

	if _, err := waitDomainDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudSearch Domain (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func validateIndexName(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if !regexache.MustCompile(`^(\*?[a-z][0-9a-z_]{2,63}|[a-z][0-9a-z_]{2,63}\*?)$`).MatchString(value) {
		es = append(es, fmt.Errorf(
			"%q must begin with a letter and be at least 3 and no more than 64 characters long", k))
	}

	if value == "score" {
		es = append(es, fmt.Errorf("'score' is a reserved field name and cannot be used"))
	}

	return
}

func defineIndexFields(ctx context.Context, conn *cloudsearch.Client, domainName string, tfList []interface{}) error {
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

			_, err = conn.DefineIndexField(ctx, input)

			if err != nil {
				return fmt.Errorf("defining CloudSearch Domain (%s) index field (%s): %w", domainName, aws.ToString(apiObject.IndexFieldName), err)
			}
		}
	}

	return nil
}

func findDomainByName(ctx context.Context, conn *cloudsearch.Client, name string) (*types.DomainStatus, error) {
	input := &cloudsearch.DescribeDomainsInput{
		DomainNames: []string{name},
	}

	output, err := conn.DescribeDomains(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfresource.AssertSingleValueResult(output.DomainStatusList)
}

func findAvailabilityOptionsStatusByName(ctx context.Context, conn *cloudsearch.Client, name string) (*types.AvailabilityOptionsStatus, error) {
	input := &cloudsearch.DescribeAvailabilityOptionsInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeAvailabilityOptions(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
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

func findDomainEndpointOptionsByName(ctx context.Context, conn *cloudsearch.Client, name string) (*types.DomainEndpointOptions, error) {
	output, err := findDomainEndpointOptionsStatusByName(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if output.Options == nil {
		return nil, tfresource.NewEmptyResultError(name)
	}

	return output.Options, nil
}

func findDomainEndpointOptionsStatusByName(ctx context.Context, conn *cloudsearch.Client, name string) (*types.DomainEndpointOptionsStatus, error) {
	input := &cloudsearch.DescribeDomainEndpointOptionsInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeDomainEndpointOptions(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
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

func findScalingParametersByName(ctx context.Context, conn *cloudsearch.Client, name string) (*types.ScalingParameters, error) {
	output, err := findScalingParametersStatusByName(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if output.Options == nil {
		return nil, tfresource.NewEmptyResultError(name)
	}

	return output.Options, nil
}

func findScalingParametersStatusByName(ctx context.Context, conn *cloudsearch.Client, name string) (*types.ScalingParametersStatus, error) {
	input := &cloudsearch.DescribeScalingParametersInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeScalingParameters(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
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

func statusDomainDeleting(ctx context.Context, conn *cloudsearch.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDomainByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, flex.BoolToStringValue(output.Deleted), nil
	}
}

func statusDomainProcessing(ctx context.Context, conn *cloudsearch.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDomainByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, flex.BoolToStringValue(output.Processing), nil
	}
}

func waitDomainActive(ctx context.Context, conn *cloudsearch.Client, name string, timeout time.Duration) (*types.DomainStatus, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{"true"},
		Target:  []string{"false"},
		Refresh: statusDomainProcessing(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DomainStatus); ok {
		return output, err
	}

	return nil, err
}

func waitDomainDeleted(ctx context.Context, conn *cloudsearch.Client, name string, timeout time.Duration) (*types.DomainStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"true"},
		Target:  []string{},
		Refresh: statusDomainDeleting(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DomainStatus); ok {
		return output, err
	}

	return nil, err
}

func expandDomainEndpointOptions(tfMap map[string]interface{}) *types.DomainEndpointOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.DomainEndpointOptions{}

	if v, ok := tfMap["enforce_https"].(bool); ok {
		apiObject.EnforceHTTPS = aws.Bool(v)
	}

	if v, ok := tfMap["tls_security_policy"].(string); ok && v != "" {
		apiObject.TLSSecurityPolicy = types.TLSSecurityPolicy(v)
	}

	return apiObject
}

func flattenDomainEndpointOptions(apiObject *types.DomainEndpointOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"tls_security_policy": apiObject.TLSSecurityPolicy,
	}

	if v := apiObject.EnforceHTTPS; v != nil {
		tfMap["enforce_https"] = aws.ToBool(v)
	}

	return tfMap
}

func expandIndexField(tfMap map[string]interface{}) (*types.IndexField, bool, error) {
	if tfMap == nil {
		return nil, false, nil
	}

	apiObject := &types.IndexField{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.IndexFieldName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.IndexFieldType = types.IndexFieldType(v)
	}

	analysisScheme, _ := tfMap["analysis_scheme"].(string)
	facetEnabled, _ := tfMap["facet"].(bool)
	highlightEnabled, _ := tfMap["highlight"].(bool)
	returnEnabled, _ := tfMap["return"].(bool)
	searchEnabled, _ := tfMap["search"].(bool)
	sortEnabled, _ := tfMap["sort"].(bool)
	var sourceFieldsConfigured bool

	switch fieldType := apiObject.IndexFieldType; fieldType {
	case types.IndexFieldTypeDate:
		options := &types.DateOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if v, ok := tfMap[names.AttrDefaultValue].(string); ok && v != "" {
			options.DefaultValue = aws.String(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceField = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.DateOptions = options

	case types.IndexFieldTypeDateArray:
		options := &types.DateArrayOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
		}

		if v, ok := tfMap[names.AttrDefaultValue].(string); ok && v != "" {
			options.DefaultValue = aws.String(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceFields = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.DateArrayOptions = options

	case types.IndexFieldTypeDouble:
		options := &types.DoubleOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if v, ok := tfMap[names.AttrDefaultValue].(string); ok && v != "" {
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

	case types.IndexFieldTypeDoubleArray:
		options := &types.DoubleArrayOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
		}

		if v, ok := tfMap[names.AttrDefaultValue].(string); ok && v != "" {
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

	case types.IndexFieldTypeInt:
		options := &types.IntOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if v, ok := tfMap[names.AttrDefaultValue].(string); ok && v != "" {
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

	case types.IndexFieldTypeIntArray:
		options := &types.IntArrayOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
		}

		if v, ok := tfMap[names.AttrDefaultValue].(string); ok && v != "" {
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

	case types.IndexFieldTypeLatlon:
		options := &types.LatLonOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if v, ok := tfMap[names.AttrDefaultValue].(string); ok && v != "" {
			options.DefaultValue = aws.String(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceField = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.LatLonOptions = options

	case types.IndexFieldTypeLiteral:
		options := &types.LiteralOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if v, ok := tfMap[names.AttrDefaultValue].(string); ok && v != "" {
			options.DefaultValue = aws.String(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceField = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.LiteralOptions = options

	case types.IndexFieldTypeLiteralArray:
		options := &types.LiteralArrayOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
		}

		if v, ok := tfMap[names.AttrDefaultValue].(string); ok && v != "" {
			options.DefaultValue = aws.String(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceFields = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.LiteralArrayOptions = options

	case types.IndexFieldTypeText:
		options := &types.TextOptions{
			HighlightEnabled: aws.Bool(highlightEnabled),
			ReturnEnabled:    aws.Bool(returnEnabled),
			SortEnabled:      aws.Bool(sortEnabled),
		}

		if analysisScheme != "" {
			options.AnalysisScheme = aws.String(analysisScheme)
		}

		if v, ok := tfMap[names.AttrDefaultValue].(string); ok && v != "" {
			options.DefaultValue = aws.String(v)
		}

		if v, ok := tfMap["source_fields"].(string); ok && v != "" {
			options.SourceField = aws.String(v)
			sourceFieldsConfigured = true
		}

		apiObject.TextOptions = options

	case types.IndexFieldTypeTextArray:
		options := &types.TextArrayOptions{
			HighlightEnabled: aws.Bool(highlightEnabled),
			ReturnEnabled:    aws.Bool(returnEnabled),
		}

		if analysisScheme != "" {
			options.AnalysisScheme = aws.String(analysisScheme)
		}

		if v, ok := tfMap[names.AttrDefaultValue].(string); ok && v != "" {
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

func flattenIndexFieldStatus(apiObject types.IndexFieldStatus) (map[string]interface{}, error) {
	if apiObject.Options == nil || apiObject.Status == nil {
		return nil, nil
	}

	// Don't read in any fields that are pending deletion.
	if aws.ToBool(apiObject.Status.PendingDeletion) {
		return nil, nil
	}

	field := apiObject.Options
	tfMap := map[string]interface{}{}

	if v := field.IndexFieldName; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	fieldType := field.IndexFieldType
	tfMap[names.AttrType] = fieldType

	switch fieldType {
	case types.IndexFieldTypeDate:
		options := field.DateOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			tfMap[names.AttrDefaultValue] = aws.ToString(v)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.ToBool(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.ToBool(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.ToBool(v)
		}

		if v := options.SortEnabled; v != nil {
			tfMap["sort"] = aws.ToBool(v)
		}

		if v := options.SourceField; v != nil {
			tfMap["source_fields"] = aws.ToString(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false

	case types.IndexFieldTypeDateArray:
		options := field.DateArrayOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			tfMap[names.AttrDefaultValue] = aws.ToString(v)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.ToBool(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.ToBool(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.ToBool(v)
		}

		if v := options.SourceFields; v != nil {
			tfMap["source_fields"] = aws.ToString(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false
		tfMap["sort"] = false

	case types.IndexFieldTypeDouble:
		options := field.DoubleOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			tfMap[names.AttrDefaultValue] = flex.Float64ToStringValue(v)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.ToBool(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.ToBool(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.ToBool(v)
		}

		if v := options.SortEnabled; v != nil {
			tfMap["sort"] = aws.ToBool(v)
		}

		if v := options.SourceField; v != nil {
			tfMap["source_fields"] = aws.ToString(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false

	case types.IndexFieldTypeDoubleArray:
		options := field.DoubleArrayOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			tfMap[names.AttrDefaultValue] = flex.Float64ToStringValue(v)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.ToBool(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.ToBool(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.ToBool(v)
		}

		if v := options.SourceFields; v != nil {
			tfMap["source_fields"] = aws.ToString(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false
		tfMap["sort"] = false

	case types.IndexFieldTypeInt:
		options := field.IntOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			tfMap[names.AttrDefaultValue] = flex.Int64ToStringValue(v)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.ToBool(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.ToBool(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.ToBool(v)
		}

		if v := options.SortEnabled; v != nil {
			tfMap["sort"] = aws.ToBool(v)
		}

		if v := options.SourceField; v != nil {
			tfMap["source_fields"] = aws.ToString(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false

	case types.IndexFieldTypeIntArray:
		options := field.IntArrayOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			tfMap[names.AttrDefaultValue] = flex.Int64ToStringValue(v)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.ToBool(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.ToBool(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.ToBool(v)
		}

		if v := options.SourceFields; v != nil {
			tfMap["source_fields"] = aws.ToString(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false
		tfMap["sort"] = false

	case types.IndexFieldTypeLatlon:
		options := field.LatLonOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			tfMap[names.AttrDefaultValue] = aws.ToString(v)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.ToBool(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.ToBool(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.ToBool(v)
		}

		if v := options.SortEnabled; v != nil {
			tfMap["sort"] = aws.ToBool(v)
		}

		if v := options.SourceField; v != nil {
			tfMap["source_fields"] = aws.ToString(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false

	case types.IndexFieldTypeLiteral:
		options := field.LiteralOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			tfMap[names.AttrDefaultValue] = aws.ToString(v)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.ToBool(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.ToBool(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.ToBool(v)
		}

		if v := options.SortEnabled; v != nil {
			tfMap["sort"] = aws.ToBool(v)
		}

		if v := options.SourceField; v != nil {
			tfMap["source_fields"] = aws.ToString(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false

	case types.IndexFieldTypeLiteralArray:
		options := field.LiteralArrayOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			tfMap[names.AttrDefaultValue] = aws.ToString(v)
		}

		if v := options.FacetEnabled; v != nil {
			tfMap["facet"] = aws.ToBool(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.ToBool(v)
		}

		if v := options.SearchEnabled; v != nil {
			tfMap["search"] = aws.ToBool(v)
		}

		if v := options.SourceFields; v != nil {
			tfMap["source_fields"] = aws.ToString(v)
		}

		// Defaults not returned via the API.
		tfMap["analysis_scheme"] = ""
		tfMap["highlight"] = false
		tfMap["sort"] = false

	case types.IndexFieldTypeText:
		options := field.TextOptions
		if options == nil {
			break
		}

		if v := options.AnalysisScheme; v != nil {
			tfMap["analysis_scheme"] = aws.ToString(v)
		}

		if v := options.DefaultValue; v != nil {
			tfMap[names.AttrDefaultValue] = aws.ToString(v)
		}

		if v := options.HighlightEnabled; v != nil {
			tfMap["highlight"] = aws.ToBool(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.ToBool(v)
		}

		if v := options.SortEnabled; v != nil {
			tfMap["sort"] = aws.ToBool(v)
		}

		if v := options.SourceField; v != nil {
			tfMap["source_fields"] = aws.ToString(v)
		}

		// Defaults not returned via the API.
		tfMap["facet"] = false
		tfMap["search"] = true

	case types.IndexFieldTypeTextArray:
		options := field.TextArrayOptions
		if options == nil {
			break
		}

		if v := options.AnalysisScheme; v != nil {
			tfMap["analysis_scheme"] = aws.ToString(v)
		}

		if v := options.DefaultValue; v != nil {
			tfMap[names.AttrDefaultValue] = aws.ToString(v)
		}

		if v := options.HighlightEnabled; v != nil {
			tfMap["highlight"] = aws.ToBool(v)
		}

		if v := options.ReturnEnabled; v != nil {
			tfMap["return"] = aws.ToBool(v)
		}

		if v := options.SourceFields; v != nil {
			tfMap["source_fields"] = aws.ToString(v)
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

func flattenIndexFieldStatuses(apiObjects []types.IndexFieldStatus) ([]interface{}, error) {
	if len(apiObjects) == 0 {
		return nil, nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap, err := flattenIndexFieldStatus(apiObject)

		if err != nil {
			return nil, err
		}

		tfList = append(tfList, tfMap)
	}

	return tfList, nil
}

func expandScalingParameters(tfMap map[string]interface{}) *types.ScalingParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ScalingParameters{}

	if v, ok := tfMap["desired_instance_type"].(string); ok && v != "" {
		apiObject.DesiredInstanceType = types.PartitionInstanceType(v)
	}

	if v, ok := tfMap["desired_partition_count"].(int); ok && v != 0 {
		apiObject.DesiredPartitionCount = int32(v)
	}

	if v, ok := tfMap["desired_replication_count"].(int); ok && v != 0 {
		apiObject.DesiredReplicationCount = int32(v)
	}

	return apiObject
}

func flattenScalingParameters(apiObject *types.ScalingParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"desired_instance_type":     apiObject.DesiredInstanceType,
		"desired_partition_count":   apiObject.DesiredPartitionCount,
		"desired_replication_count": apiObject.DesiredReplicationCount,
	}

	return tfMap
}
