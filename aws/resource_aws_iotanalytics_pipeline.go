package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func generateAddAttributesActivity() *schema.Resource {	
	return schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type: schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type: schema.TypeString,
				Optional: true,
			},
			"attributes": {
				Type: schema.TypeMap,
				Required: true,
				MinItems: 1,
				MaxItems: 50, 
				Elem: schema.TypeString,
			},
		},
	}
}

func generateRemoveAttributesActivity() *schema.Resource {	
	return schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type: schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type: schema.TypeString,
				Optional: true,
			},
			"attribute": {
				Type: schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 50, 
				Elem: schema.TypeString,
			},
		},
	}
}

func generateSelectAttributesActivity() *schema.Resource {
	return schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type: schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type: schema.TypeString,
				Optional: true,
			},
			"attribute": {
				Type: schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 50, 
				Elem: schema.TypeString,
			},
		},
	}
}

func generateChannelActivity() *schema.Resource {	
	return schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type: schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type: schema.TypeString,
				Optional: true,
			},
			"channel_name": {
				Type: schema.TypeString,
				Required: true,
			},
		},
	}
}

func generateDatastoreActivity() *schema.Resource {	
	return schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type: schema.TypeString,
				Required: true,
			},
			"datastore_name": {
				Type: schema.TypeString,
				Required: true,
			},
		},
	}
}

func generateDeviceRegistryEnrichActivity() *schema.Resource {	
	return schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type: schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type: schema.TypeString,
				Optional: true,
			},
			"role_arn": {
				Type: schema.TypeString,
				Required: true,
				ValidateFunc: validateArn,
			},
			"thing_name": {
				Type: schema.TypeString,
				Optional: true,
			},
			"attribute": {
				Type: schema.TypeString,
				Required: true,
			},		
		},
	}
}

func generateDeviceShadowEnrichActivity() *schema.Resource {	
	return schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type: schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type: schema.TypeString,
				Optional: true,
			},
			"role_arn": {
				Type: schema.TypeString,
				Required: true,
				ValidateFunc: validateArn,
			},
			"thing_name": {
				Type: schema.TypeString,
				Optional: true,
			},
			"attribute": {
				Type: schema.TypeString,
				Required: true,
			},		
		},
	}
}

func generateFilterActivity() *schema.Resource {	
	return schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type: schema.TypeString,
				Required: true,
			},
			"filter": {
				Type: schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type: schema.TypeString,
				Optional: true,
			},
		},
	}
}

func generateLambdaActivity() *schema.Resource {	
	return schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type: schema.TypeString,
				Required: true,
			},
			"lambda_name": {
				Type: schema.TypeString,
				Required: true,
			},
			"batch_size": {
				Type: schema.TypeInt,
				Required: true,
			},
			"next_activity": {
				Type: schema.TypeString,
				Optional: true,
			},
		},
	}
}

func generateMathActivity() *schema.Resource {	
	return schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type: schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type: schema.TypeString,
				Optional: true,
			},
			"attribute": {
				Type: schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type: schema.TypeString,
				Optional: true,
			},		
		},
	}
}


func generatePipelineActivitySchema() *schema.Resource {
	return schema.Resource{
		Schema: map[string]*schema.Schema{
			"add_attributes": {
				Type: schema.TypeSet,
				MaxItems: 1,
				Optional: true ,
				Elem: generateAddAttributesActivity(),
			},
			"remove_attributes": {
				Type: schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: generateRemoveAttributesActivity(),
			},
			"select_attributes": {
				Type: schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: generateSelectAttributesActivity(),
			},
			"channel": {
				Type: schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: generateChannelActivity(),
			},
			"datastore": {
				Type: schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: generateDatastoreActivity(),
			},
			"device_registry_enrich": {
				Type: schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: generateDeviceRegistryEnrichActivity(),
			},
			"device_shadow_enrich": {
				Type: schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: generateDeviceShadowEnrichActivity(),
			},
			"filter": {
				Type: schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: generateFilterActivity(),
			},
			"lambda": {
				Type: schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: generateLambdaActivity(),
			},
			"math": {
				Type: schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: generateMathActivity(),
			},
		}
	}
}

func resourceAwsIotAnalyticsPipeline() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotAnalyticsPipelineCreate,
		Read:   resourceAwsIotAnalyticsPipelineRead,
		Update: resourceAwsIotAnalyticsPipelineUpdate,
		Delete: resourceAwsIotAnalyticsPipelineDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type: schema.TypeString,
				Required: true
			},
			"pipeline_activity": {
				Type: schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 25,
				Elem: generatePipelineActivitySchema(),
			},
		},
	}
}

func parseAddAttributesActivity(rawAddAttributesActivity map[string]interface{}) *iotanalytics.AddAttributesActivity {	
	attributes := make(map[string]string)
	rawAttributes := rawAddAttributesActivity["attributes"].(map[string]string)
	for k, v := range rawAttributes {
		attributes[k] = aws.String(v)
	}

	name := rawAddAttributesActivity["name"].(string)
	addAttrActivity := &iotanalytics.AddAttributesActivity{
		Name: aws.String(name),
		Attributes: attributes,
	}

	if v, ok := rawAddAttributesActivity["next_activity"].(string); ok && v != "" {
		addAttrActivity.Next = aws.String(v)
	}

	return addAttrActivity
}

func parseChannelActivity(rawChannelActivity map[string]interface{}) *iotanalytics.ChannelActivity {	
	name := rawChannelActivity["name"].(string)
	channelName := rawChannelActivity["channel_name"].(string)

	channelActivity := &iotanalytics.ChannelActivity{
		Name: aws.String(name),
		ChannelName: aws.String(channelName),
	}

	if v, ok := rawChannelActivity["next_activity"].(string); ok && v != "" {
		channelActivity.Next = aws.String(v)
	}

	return channelActivity
}

func parseDatastoreActivity(rawDatastoreActivity map[string]interface{}) *iotanalytics.DatastoreActivity {	
	name := rawDatastoreActivity["name"].(string)
	datastoreName := rawDatastoreActivity["datastore_name"].(string)

	datastore := &iotanalytics.DatastoreActivity{
		Name: aws.String(name),
		DatastoreName: aws.String(datastoreName),
	}

	return datastore
}

func parseDeviceRegistryEnrichActivity(rawDeviceRegistryEnrichActivity map[string]interface{}) *iotanalytics.DeviceRegistryEnrichActivity {	
	name := rawDeviceRegistryEnrichActivity["name"].(string)
	attribute := rawDeviceRegistryEnrichActivity["attribute"].(string)
	roleArn := rawDeviceRegistryEnrichActivity["role_arn"].(string)
	thingName := rawDeviceRegistryEnrichActivity["thing_name"].(string)


	deviceRegistryEnrich := &iotanalytics.DeviceRegistryEnrichActivity{
		Name: aws.String(name),
		Attribute: aws.String(attribute),
		RoleArn: aws.String(roleArn),
		ThingName:  aws.String(thingName),
	}

	if v, ok := rawDeviceRegistryEnrichActivity["next_activity"].(string); ok && v != "" {
		deviceRegistryEnrich.Next = aws.String(v)
	}

	return deviceRegistryEnrich
}

func parseDeviceShadowEnrichActivity(rawDeviceShadowEnrichActivity map[string]interface{}) *iotanalytics.DeviceShadowEnrichActivity {	
	name := rawDeviceShadowEnrichActivity["name"].(string)
	attribute := rawDeviceShadowEnrichActivity["attribute"].(string)
	roleArn := rawDeviceShadowEnrichActivity["role_arn"].(string)
	thingName := rawDeviceShadowEnrichActivity["thing_name"].(string)


	deviceShadowEnrich := &iotanalytics.DeviceShadowEnrichActivity{
		Name: aws.String(name),
		Attribute: aws.String(attribute),
		RoleArn: aws.String(roleArn),
		ThingName:  aws.String(thingName),
	}

	if v, ok := rawDeviceShadowEnrichActivity["next_activity"].(string); ok && v != "" {
		deviceShadowEnrich.Next = aws.String(v)
	}

	return deviceShadowEnrich
}

func parseFilterActivity(rawFilterActivity map[string]interface{}) *iotanalytics.FilterActivity {
	name := rawFilterActivity["name"].(string)
	filter := rawFilterActivity["filter"].(string)

	filterActivity := &iotanalytics.FilterActivity{
		Name: aws.String(name),
		Filter: aws.String(channelName),
	}

	if v, ok := rawFilterActivity["next_activity"].(string); ok && v != "" {
		filterActivity.Next = aws.String(v)
	}

	return filterActivity
}

func parseLambdaActivity(rawLambdaActivity map[string]interface{}) *iotanalytics.LambdaActivity {	
	name := rawLambdaActivity["name"].(string)
	lambdaName := rawLambdaActivity["lambda_name"].(string)
	batchSize := int64(rawLambdaActivity["batch_size"].(int))

	lambdaActivity := &iotanalytics.LambdaActivity{
		Name: aws.String(name),
		LambdaName: aws.String(channelName),
		BatchSize: aws.Int64(batchSize),
	}

	if v, ok := rawLambdaActivity["next_activity"].(string); ok && v != "" {
		lambdaActivity.Next = aws.String(v)
	}

	return lambdaActivity
}

func parseMathActivity(rawMathActivity map[string]interface{}) *iotanalytics.MathActivity {	
	name := rawMathActivity["name"].(string)
	attribute := rawMathActivity["attribute"].(string)
	math := rawMathActivity["math"].(string)

	mathActivity := &iotanalytics.MathActivity{
		Name: aws.String(name),
		Attribute: aws.String(attribute),
		Math: aws.String(math),
	}

	if v, ok := rawMathActivity["next_activity"].(string); ok && v != "" {
		mathActivity.Next = aws.String(v)
	}

	return mathActivity
}

func parseRemoveAttributesActivity(rawRemoveAttributesActivity map[string]interface{}) *iotanalytics.RemoveAttributesActivity {	
	name := rawRemoveAttributesActivity["name"].(string)
	attributes := rawRemoveAttributesActivity["attribute"].([]string)

	removeAttrActivity := &iotanalytics.RemoveAttributesActivity{
		Name: aws.String(name),
		Attributes: attributes,
	}

	if v, ok := rawRemoveAttributesActivity["next_activity"].(string); ok && v != "" {
		removeAttrActivity.Next = aws.String(v)
	}

	return removeAttrActivity
}

func parseSelectAttributesActivity(rawSelectAttributesActivity map[string]interface{}) *iotanalytics.SelectAttributesActivity {
	name := rawSelectAttributesActivity["name"].(string)
	attributes := rawSelectAttributesActivity["attribute"].([]string)

	selectAttrActivity := &iotanalytics.SelectAttributesActivity{
		Name: aws.String(name),
		Attributes: attributes,
	}

	if v, ok := rawSelectAttributesActivity["next_activity"].(string); ok && v != "" {
		selectAttrActivity.Next = aws.String(v)
	}

	return selectAttrActivity
}

func parsePipelineActivity(rawPipelineActivity map[string]interface{}) *iotanalytics.PipelineActivity {
	pipelineActivity := &iotanalytics.PipelineActivity{}

	if v := rawPipelineActivity["add_attributes"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.AddAttributes = parseAddAttributesActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["select_attributes"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.SelectAttributes = parseSelectAttributesActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["remove_attributes"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.RemoveAttributes = parseRemoveAttributesActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["channel"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.Channel = parseChannelActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["datastore"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.Datastore = parseDatastoreActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["device_registry_enrich"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.DeviceRegistryEnrich = parseDeviceRegistryEnrichActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["device_shadow_enrich"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.DeviceShadowEnrich = parseDeviceShadowEnrichActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["filter"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.Filter = parseFilterActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["lambda"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.Lambda = parseLambdaActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["math"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.Math = parseMathActivity(v[0].(map[string]interface{}))
	}

	return pipelineActivity
}


func resourceAwsIotAnalyticsPipelineCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn
	
	rawPipelineActivites := d.Get("pipeline_activity").(*schema.Set).List()
	pipelineActivities := make([]map[string]interface{}, 0)
	for _, rawPA := range rawPipelineActivites {
		pA := parsePipelineActivity(rawPA.(map[string]interface{}))
		pipelineActivities := append(pipelineActivities, pA)
	}

	pipelineName := d.Get("name").(string)
	params := &iotanalytics.CreatePipelineInput{
		PipelineName: aws.String(pipelineName),
		PipelineActivities: pipelineActivities,
	}

	_, err := conn.CreatePipeline(params)
	
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Create IoT Analytics Pipeline: %s", params)

	d.SetId(pipelineName)

	return resourceAwsIotAnalyticsPipelineRead(d, meta)
}

func flattenAddAttributesActivity(addAttributesActivity *iotanalytics.AddAttributesActivity) map[string]interface{} {
	rawAddAttrActivity := make(map[string]interface{})
	rawAddAttrActivity["name"] = aws.StringValue(addAttrActivity.Name)

	rawAttrs := make(map[string]interface{})
	for k, v := range addAttrActivity.Attributes {
		rawAttrs[k] = aws.StringValue(v)
	}
	rawAddAttrActivity["attributes"] = rawAttrs

	if addAttrActivity.Next != nil && *addAttrActivity.Next != "" {
		rawAddAttrActivity["next"] = *addAttrActivity.Next
	}

	return rawAddAttrActivity
}

func flattenChannelActivity(channelActivity *iotanalytics.ChannelActivity) map[string]interface{} {
}

func flattenDatastoreActivity(datastoreActivity *iotanalytics.DatastoreActivity) map[string]interface{} {
}

func flattenDeviceRegistryEnrichActivity(deviceRegistryEnrichActivity *iotanalytics.DeviceRegistryEnrichActivity) map[string]interface{} {
}

func flattenDeviceShadowEnrichActivity(deviceShadowEnrichActivity *iotanalytics.DeviceShadowEnrichActivity) map[string]interface{} {
}

func flattenFilterActivity(filterActivity *iotanalytics.FilterActivity) map[string]interface{} {
}

func flattenLambdaActivity(lambdaActivity *iotanalytics.LambdaActivity) map[string]interface{} {
}

func flattenMathActivity(mathActivity *iotanalytics.MathActivity) map[string]interface{} {
}

func flattenRemoveAttributesActivity(removeAttributesActivity *iotanalytics.RemoveAttributesActivity) map[string]interface{} {
}

func flattenSelectAttributesActivity(selectAttributesActivity *iotanalytics.SelectAttributesActivity) map[string]interface{} {
}

func flattenPipelineActivity(pipelineActivity *iotanalytics.PipelineActivity) map[string]interface{} {
}

func resourceAwsIotAnalyticsPipelineRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	return nil
}

func resourceAwsIotAnalyticsPipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn
	pipelineName := d.Get("name").(string)
	
	rawPipelineActivites := d.Get("pipeline_activity").(*schema.Set).List()
	pipelineActivities := make([]map[string]interface{}, 0)
	for _, rawPA := range rawPipelineActivites {
		pA := parsePipelineActivity(rawPA.(map[string]interface{}))
		pipelineActivities := append(pipelineActivities, pA)
	}

	params := &iotanalytics.UpdatePipelineInput{
		PipelineName: aws.String(pipelineName),
		PipelineActivities: pipelineActivities,
	}

	_, err := conn.UpdatePipeline(params)
	
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Update IoT Analytics Pipeline: %s", params)

	return resourceAwsIotAnalyticsPipelineRead(d, meta)

}

func resourceAwsIotAnalyticsPipelineDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalyticsconn.DeletePipelineInput{
		RuleName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting IoT Analytics Pipeline: %s", params)
	_, err := conn.DeletePipeline(params)
}
