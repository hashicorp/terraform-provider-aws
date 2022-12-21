package sagemaker

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppCreate,
		Read:   resourceAppRead,
		Update: resourceAppUpdate,
		Delete: resourceAppDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"app_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"app_type": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice(sagemaker.AppType_Values(), false),
			},
			"domain_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"resource_spec": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
						},
						"lifecycle_config_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"sagemaker_image_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidARN,
						},
						"sagemaker_image_version_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"space_name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ExactlyOneOf: []string{"space_name", "user_profile_name"},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"user_profile_name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ExactlyOneOf: []string{"space_name", "user_profile_name"},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAppCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &sagemaker.CreateAppInput{
		AppName:  aws.String(d.Get("app_name").(string)),
		AppType:  aws.String(d.Get("app_type").(string)),
		DomainId: aws.String(d.Get("domain_id").(string)),
	}

	if v, ok := d.GetOk("user_profile_name"); ok {
		input.UserProfileName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("space_name"); ok {
		input.SpaceName = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("resource_spec"); ok {
		input.ResourceSpec = expandDomainDefaultResourceSpec(v.([]interface{}))
	}

	log.Printf("[DEBUG] SageMaker App create config: %#v", *input)
	output, err := conn.CreateApp(input)
	if err != nil {
		return fmt.Errorf("creating SageMaker App: %w", err)
	}

	appArn := aws.StringValue(output.AppArn)
	domainID, userProfileOrSpaceName, appType, appName, err := decodeAppID(appArn)
	if err != nil {
		return err
	}

	d.SetId(appArn)

	if _, err := WaitAppInService(conn, domainID, userProfileOrSpaceName, appType, appName); err != nil {
		return fmt.Errorf("waiting for SageMaker App (%s) to create: %w", d.Id(), err)
	}

	return resourceAppRead(d, meta)
}

func resourceAppRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	domainID, userProfileOrSpaceName, appType, appName, err := decodeAppID(d.Id())
	if err != nil {
		return err
	}

	app, err := FindAppByName(conn, domainID, userProfileOrSpaceName, appType, appName)
	if err != nil {
		if !d.IsNewResource() && tfresource.NotFound(err) {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker App (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("reading SageMaker App (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(app.AppArn)
	d.Set("app_name", app.AppName)
	d.Set("app_type", app.AppType)
	d.Set("arn", arn)
	d.Set("domain_id", app.DomainId)
	d.Set("user_profile_name", app.UserProfileName)
	d.Set("space_name", app.SpaceName)

	if err := d.Set("resource_spec", flattenDomainDefaultResourceSpec(app.ResourceSpec)); err != nil {
		return fmt.Errorf("setting resource_spec for SageMaker App (%s): %w", d.Id(), err)
	}

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("listing tags for SageMaker App (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceAppUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("updating SageMaker App (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAppRead(d, meta)
}

func resourceAppDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	appName := d.Get("app_name").(string)
	appType := d.Get("app_type").(string)
	domainID := d.Get("domain_id").(string)
	userProfileOrSpaceName := ""

	input := &sagemaker.DeleteAppInput{
		AppName:  aws.String(appName),
		AppType:  aws.String(appType),
		DomainId: aws.String(domainID),
	}

	if v, ok := d.GetOk("user_profile_name"); ok {
		input.UserProfileName = aws.String(v.(string))
		userProfileOrSpaceName = v.(string)
	}

	if v, ok := d.GetOk("space_name"); ok {
		input.SpaceName = aws.String(v.(string))
		userProfileOrSpaceName = v.(string)
	}

	if _, err := conn.DeleteApp(input); err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "has already been deleted") ||
			tfawserr.ErrMessageContains(err, "ValidationException", "previously failed and was automatically deleted") {
			return nil
		}

		if !tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return fmt.Errorf("deleting SageMaker App (%s): %w", d.Id(), err)
		}
	}

	if _, err := WaitAppDeleted(conn, domainID, userProfileOrSpaceName, appType, appName); err != nil {
		if !tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return fmt.Errorf("waiting for SageMaker App (%s) to delete: %w", d.Id(), err)
		}
	}

	return nil
}

func decodeAppID(id string) (string, string, string, string, error) {
	appArn, err := arn.Parse(id)
	if err != nil {
		return "", "", "", "", err
	}

	appResourceName := strings.TrimPrefix(appArn.Resource, "app/")
	parts := strings.Split(appResourceName, "/")

	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("unexpected format of ID (%q), expected DOMAIN-ID/USER-PROFILE-NAME OR PROFILE-NAME/APP-TYPE/APP-NAME", appResourceName)
	}

	domainID := parts[0]
	userProfileOrSpaceName := parts[1]
	appType := parts[2]

	if appType == "jupyterserver" {
		appType = sagemaker.AppTypeJupyterServer
	} else if appType == "kernelgateway" {
		appType = sagemaker.AppTypeKernelGateway
	} else if appType == "tensorboard" {
		appType = sagemaker.AppTypeTensorBoard
	} else if appType == "rstudioserverpro" {
		appType = sagemaker.AppTypeRstudioServerPro
	} else if appType == "rsessiongateway" {
		appType = sagemaker.AppTypeRsessionGateway
	}

	appName := parts[3]

	return domainID, userProfileOrSpaceName, appType, appName, nil
}
