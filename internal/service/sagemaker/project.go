package sagemaker

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,31}$`),
						"Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"project_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"service_catalog_provisioning_details": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"product_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"provisioning_artifact_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"provisioning_parameter": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProjectCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("project_name").(string)
	input := &sagemaker.CreateProjectInput{
		ProjectName:                       aws.String(name),
		ServiceCatalogProvisioningDetails: expandProjectServiceCatalogProvisioningDetails(d.Get("service_catalog_provisioning_details").([]interface{})),
	}

	if v, ok := d.GetOk("project_description"); ok {
		input.ProjectDescription = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	_, err := verify.RetryOnAWSCode("ValidationException", func() (interface{}, error) {
		return conn.CreateProject(input)
	})
	if err != nil {
		return fmt.Errorf("error creating SageMaker project: %w", err)
	}

	d.SetId(name)

	if _, err := WaitProjectCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for SageMaker Project (%s) to be created: %w", d.Id(), err)
	}

	return resourceProjectRead(d, meta)
}

func resourceProjectRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	project, err := FindProjectByName(conn, d.Id())
	if err != nil {
		if !d.IsNewResource() && tfresource.NotFound(err) {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Project (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker Project (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(project.ProjectArn)
	d.Set("project_name", project.ProjectName)
	d.Set("project_id", project.ProjectId)
	d.Set("arn", arn)
	d.Set("project_description", project.ProjectDescription)

	if err := d.Set("service_catalog_provisioning_details", flattenProjectServiceCatalogProvisioningDetails(project.ServiceCatalogProvisioningDetails)); err != nil {
		return fmt.Errorf("error setting service_catalog_provisioning_details: %w", err)
	}

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker Project (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	if d.HasChangesExcept("tags_all", "tags") {
		input := &sagemaker.UpdateProjectInput{
			ProjectName: aws.String(d.Id()),
		}

		if d.HasChange("project_description") {
			input.ProjectDescription = aws.String(d.Get("project_description").(string))
		}

		if d.HasChange("service_catalog_provisioning_details") {
			input.ServiceCatalogProvisioningUpdateDetails = expandProjectServiceCatalogProvisioningDetailsUpdate(d.Get("service_catalog_provisioning_details").([]interface{}))
		}

		_, err := conn.UpdateProject(input)

		if err != nil {
			return fmt.Errorf("error updating SageMaker Project (%s): %w", d.Id(), err)
		}

		if _, err := WaitProjectUpdated(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for SageMaker Project (%s) to be updated: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker Project (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceProjectRead(d, meta)
}

func resourceProjectDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	input := &sagemaker.DeleteProjectInput{
		ProjectName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteProject(input); err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") ||
			tfawserr.ErrMessageContains(err, "ValidationException", "Cannot delete Project in DeleteCompleted status") {
			return nil
		}
		return fmt.Errorf("error deleting SageMaker Project (%s): %w", d.Id(), err)
	}

	if _, err := WaitProjectDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for SageMaker Project (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func expandProjectServiceCatalogProvisioningDetails(l []interface{}) *sagemaker.ServiceCatalogProvisioningDetails {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	scpd := &sagemaker.ServiceCatalogProvisioningDetails{
		ProductId: aws.String(m["product_id"].(string)),
	}

	if v, ok := m["path_id"].(string); ok && v != "" {
		scpd.PathId = aws.String(v)
	}

	if v, ok := m["provisioning_artifact_id"].(string); ok && v != "" {
		scpd.ProvisioningArtifactId = aws.String(v)
	}

	if v, ok := m["provisioning_parameter"].([]interface{}); ok && len(v) > 0 {
		scpd.ProvisioningParameters = expandProjectProvisioningParameters(v)
	}

	return scpd
}

func expandProjectServiceCatalogProvisioningDetailsUpdate(l []interface{}) *sagemaker.ServiceCatalogProvisioningUpdateDetails {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	scpd := &sagemaker.ServiceCatalogProvisioningUpdateDetails{}

	if v, ok := m["provisioning_artifact_id"].(string); ok && v != "" {
		scpd.ProvisioningArtifactId = aws.String(v)
	}

	if v, ok := m["provisioning_parameter"].([]interface{}); ok && len(v) > 0 {
		scpd.ProvisioningParameters = expandProjectProvisioningParameters(v)
	}

	return scpd
}

func expandProjectProvisioningParameters(l []interface{}) []*sagemaker.ProvisioningParameter {
	if len(l) == 0 {
		return nil
	}

	params := make([]*sagemaker.ProvisioningParameter, 0, len(l))

	for _, lRaw := range l {
		data := lRaw.(map[string]interface{})

		scpd := &sagemaker.ProvisioningParameter{
			Key: aws.String(data["key"].(string)),
		}

		if v, ok := data["value"].(string); ok && v != "" {
			scpd.Value = aws.String(v)
		}

		params = append(params, scpd)
	}

	return params
}

func flattenProjectServiceCatalogProvisioningDetails(scpd *sagemaker.ServiceCatalogProvisioningDetails) []map[string]interface{} {
	if scpd == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"product_id": aws.StringValue(scpd.ProductId),
	}

	if scpd.PathId != nil {
		m["path_id"] = aws.StringValue(scpd.PathId)
	}

	if scpd.ProvisioningArtifactId != nil {
		m["provisioning_artifact_id"] = aws.StringValue(scpd.ProvisioningArtifactId)
	}

	if scpd.ProvisioningParameters != nil {
		m["provisioning_parameter"] = flattenProjectProvisioningParameters(scpd.ProvisioningParameters)
	}

	return []map[string]interface{}{m}
}

func flattenProjectProvisioningParameters(scpd []*sagemaker.ProvisioningParameter) []map[string]interface{} {
	params := make([]map[string]interface{}, 0, len(scpd))

	for _, lRaw := range scpd {
		param := make(map[string]interface{})
		param["key"] = aws.StringValue(lRaw.Key)

		if lRaw.Value != nil {
			param["value"] = lRaw.Value
		}

		params = append(params, param)
	}

	return params
}
