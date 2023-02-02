package comprehend

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/aws/ratelimit"
	"github.com/aws/aws-sdk-go-v2/service/comprehend"
	"github.com/aws/aws-sdk-go-v2/service/comprehend/types"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	entityRecognizerTagKey = "tf-aws_comprehend_entity_recognizer"
)

func ResourceEntityRecognizer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEntityRecognizerCreate,
		ReadWithoutTimeout:   resourceEntityRecognizerRead,
		UpdateWithoutTimeout: resourceEntityRecognizerUpdate,
		DeleteWithoutTimeout: resourceEntityRecognizerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_access_role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"input_data_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"annotations": {
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							ExactlyOneOf: []string{"input_data_config.0.annotations", "input_data_config.0.entity_list"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
									"test_s3_uri": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"augmented_manifests": {
							Type:         schema.TypeSet,
							Optional:     true,
							ExactlyOneOf: []string{"input_data_config.0.augmented_manifests", "input_data_config.0.documents"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"annotation_data_s3_uri": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"attribute_names": {
										Type:     schema.TypeList,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"document_type": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.AugmentedManifestsDocumentTypeFormat](),
										Default:          types.AugmentedManifestsDocumentTypeFormatPlainTextDocument,
									},
									"s3_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
									"source_documents_s3_uri": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"split": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.Split](),
										Default:          types.SplitTrain,
									},
								},
							},
						},
						"data_format": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.EntityRecognizerDataFormat](),
							Default:          types.EntityRecognizerDataFormatComprehendCsv,
						},
						"documents": {
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							ExactlyOneOf: []string{"input_data_config.0.documents", "input_data_config.0.augmented_manifests"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"input_format": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.InputFormat](),
										Default:          types.InputFormatOneDocPerLine,
									},
									"s3_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
									"test_s3_uri": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"entity_list": {
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							ExactlyOneOf: []string{"input_data_config.0.entity_list", "input_data_config.0.annotations"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"entity_types": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 25,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringIsNotEmpty,
											validation.StringDoesNotContainAny("\n\r\t,"),
										),
									},
								},
							},
						},
					},
				},
			},
			"language_code": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.SyntaxLanguageCode](),
			},
			"model_kms_key_id": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: tfkms.DiffSuppressKey,
				ValidateFunc:     tfkms.ValidateKey,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validModelName,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"version_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validModelVersionName,
				ConflictsWith: []string{"version_name_prefix"},
			},
			"version_name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validModelVersionNamePrefix,
				ConflictsWith: []string{"version_name"},
			},
			"volume_kms_key_id": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: tfkms.DiffSuppressKey,
				ValidateFunc:     tfkms.ValidateKey,
			},
			"vpc_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
				tfMap := getEntityRecognizerInputDataConfig(diff)
				if tfMap == nil {
					return nil
				}

				if format := types.EntityRecognizerDataFormat(tfMap["data_format"].(string)); format == types.EntityRecognizerDataFormatComprehendCsv {
					if tfMap["documents"] == nil {
						return fmt.Errorf("documents must be set when data_format is %s", format)
					}
				} else {
					if tfMap["augmented_manifests"] == nil {
						return fmt.Errorf("augmented_manifests must be set when data_format is %s", format)
					}
				}

				return nil
			},
			func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
				tfMap := getEntityRecognizerInputDataConfig(diff)
				if tfMap == nil {
					return nil
				}

				documents := expandDocuments(tfMap["documents"].([]interface{}))
				if documents == nil {
					return nil
				}

				annotations := expandAnnotations(tfMap["annotations"].([]interface{}))
				if annotations == nil {
					return nil
				}

				if documents.TestS3Uri != nil {
					if annotations.TestS3Uri == nil {
						return errors.New("input_data_config.annotations.test_s3_uri must be set when input_data_config.documents.test_s3_uri is set")
					}
				} else {
					if annotations.TestS3Uri != nil {
						return errors.New("input_data_config.documents.test_s3_uri must be set when input_data_config.annotations.test_s3_uri is set")
					}
				}

				return nil
			},
		),
	}
}

func resourceEntityRecognizerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	awsClient := meta.(*conns.AWSClient)
	conn := awsClient.ComprehendClient()

	var versionName *string
	raw := d.GetRawConfig().GetAttr("version_name")
	if raw.IsNull() {
		versionName = aws.String(create.Name("", d.Get("version_name_prefix").(string)))
	} else if v := raw.AsString(); v != "" {
		versionName = aws.String(v)
	}

	diags := entityRecognizerPublishVersion(ctx, conn, d, versionName, create.ErrActionCreating, d.Timeout(schema.TimeoutCreate), awsClient)
	if diags.HasError() {
		return diags
	}

	return append(diags, resourceEntityRecognizerRead(ctx, d, meta)...)
}

func resourceEntityRecognizerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ComprehendClient()

	out, err := FindEntityRecognizerByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Comprehend Entity Recognizer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Comprehend Entity Recognizer (%s): %s", d.Id(), err)
	}

	d.Set("arn", out.EntityRecognizerArn)
	d.Set("data_access_role_arn", out.DataAccessRoleArn)
	d.Set("language_code", out.LanguageCode)
	d.Set("model_kms_key_id", out.ModelKmsKeyId)
	d.Set("version_name", out.VersionName)
	d.Set("version_name_prefix", create.NamePrefixFromName(aws.ToString(out.VersionName)))
	d.Set("volume_kms_key_id", out.VolumeKmsKeyId)

	// DescribeEntityRecognizer() doesn't return the model name
	name, err := EntityRecognizerParseARN(aws.ToString(out.EntityRecognizerArn))
	if err != nil {
		return diag.Errorf("reading Comprehend Entity Recognizer (%s): %s", d.Id(), err)
	}
	d.Set("name", name)

	if err := d.Set("input_data_config", flattenEntityRecognizerInputDataConfig(out.InputDataConfig)); err != nil {
		return diag.Errorf("setting input_data_config: %s", err)
	}

	if err := d.Set("vpc_config", flattenVPCConfig(out.VpcConfig)); err != nil {
		return diag.Errorf("setting vpc_config: %s", err)
	}

	tags, err := ListTags(ctx, conn, d.Id())
	if err != nil {
		return diag.Errorf("listing tags for Comprehend Entity Recognizer (%s): %s", d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceEntityRecognizerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	awsClient := meta.(*conns.AWSClient)
	conn := awsClient.ComprehendClient()

	var diags diag.Diagnostics

	if d.HasChangesExcept("tags", "tags_all") {
		var versionName *string
		if d.HasChange("version_name") {
			versionName = aws.String(d.Get("version_name").(string))
		} else if v := d.Get("version_name_prefix").(string); v != "" {
			versionName = aws.String(create.Name("", d.Get("version_name_prefix").(string)))
		}

		diags := entityRecognizerPublishVersion(ctx, conn, d, versionName, create.ErrActionUpdating, d.Timeout(schema.TimeoutUpdate), awsClient)
		if diags.HasError() {
			return diags
		}
	} else if d.HasChange("tags_all") {
		// For a tags-only change. If tag changes are combined with version publishing, the tags are set
		// by the CreateEntityRecognizer call
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags for Comprehend Entity Recognizer (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceEntityRecognizerRead(ctx, d, meta)...)
}

func resourceEntityRecognizerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ComprehendClient()

	log.Printf("[INFO] Stopping Comprehend Entity Recognizer (%s)", d.Id())

	_, err := conn.StopTrainingEntityRecognizer(ctx, &comprehend.StopTrainingEntityRecognizerInput{
		EntityRecognizerArn: aws.String(d.Id()),
	})
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return diag.Errorf("stopping Comprehend Entity Recognizer (%s): %s", d.Id(), err)
	}

	if _, err := waitEntityRecognizerStopped(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return diag.Errorf("waiting for Comprehend Entity Recognizer (%s) to be stopped: %s", d.Id(), err)
	}

	name, err := EntityRecognizerParseARN(d.Id())
	if err != nil {
		return diag.Errorf("deleting Comprehend Entity Recognizer (%s): %s", d.Id(), err)
	}

	log.Printf("[INFO] Deleting Comprehend Entity Recognizer (%s)", name)

	versions, err := ListEntityRecognizerVersionsByName(ctx, conn, name)
	if err != nil {
		return diag.Errorf("deleting Comprehend Entity Recognizer (%s): %s", name, err)
	}

	var g multierror.Group
	for _, v := range versions {
		v := v
		g.Go(func() error {
			_, err = conn.DeleteEntityRecognizer(ctx, &comprehend.DeleteEntityRecognizerInput{
				EntityRecognizerArn: v.EntityRecognizerArn,
			})
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if !errors.As(err, &nfe) {
					return fmt.Errorf("deleting version (%s): %w", aws.ToString(v.VersionName), err)
				}
			}

			if _, err := waitEntityRecognizerDeleted(ctx, conn, aws.ToString(v.EntityRecognizerArn), d.Timeout(schema.TimeoutDelete)); err != nil {
				return fmt.Errorf("waiting for version (%s) to be deleted: %s", aws.ToString(v.VersionName), err)
			}

			ec2Conn := meta.(*conns.AWSClient).EC2Conn()
			networkInterfaces, err := tfec2.FindNetworkInterfaces(ctx, ec2Conn, &ec2.DescribeNetworkInterfacesInput{
				Filters: []*ec2.Filter{
					tfec2.NewFilter(fmt.Sprintf("tag:%s", entityRecognizerTagKey), []string{aws.ToString(v.EntityRecognizerArn)}),
				},
			})
			if err != nil {
				return fmt.Errorf("finding ENIs for version (%s): %w", aws.ToString(v.VersionName), err)
			}

			for _, v := range networkInterfaces {
				v := v
				g.Go(func() error {
					networkInterfaceID := aws.ToString(v.NetworkInterfaceId)

					if v.Attachment != nil {
						err = tfec2.DetachNetworkInterface(ctx, ec2Conn, networkInterfaceID, aws.ToString(v.Attachment.AttachmentId), d.Timeout(schema.TimeoutDelete))

						if err != nil {
							return fmt.Errorf("detaching ENI (%s): %w", networkInterfaceID, err)
						}
					}

					err = tfec2.DeleteNetworkInterface(ctx, ec2Conn, networkInterfaceID)
					if err != nil {
						return fmt.Errorf("deleting ENI (%s): %w", networkInterfaceID, err)
					}

					return nil
				})
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return diag.Errorf("deleting Comprehend Entity Recognizer (%s): %s", name, err)
	}

	return nil
}

func entityRecognizerPublishVersion(ctx context.Context, conn *comprehend.Client, d *schema.ResourceData, versionName *string, action string, timeout time.Duration, awsClient *conns.AWSClient) diag.Diagnostics {
	in := &comprehend.CreateEntityRecognizerInput{
		DataAccessRoleArn:  aws.String(d.Get("data_access_role_arn").(string)),
		InputDataConfig:    expandEntityRecognizerInputDataConfig(getEntityRecognizerInputDataConfig(d)),
		LanguageCode:       types.LanguageCode(d.Get("language_code").(string)),
		RecognizerName:     aws.String(d.Get("name").(string)),
		VersionName:        versionName,
		VpcConfig:          expandVPCConfig(d.Get("vpc_config").([]interface{})),
		ClientRequestToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.Get("model_kms_key_id").(string); ok && v != "" {
		in.ModelKmsKeyId = aws.String(v)
	}

	if v, ok := d.Get("volume_kms_key_id").(string); ok && v != "" {
		in.VolumeKmsKeyId = aws.String(v)
	}

	defaultTagsConfig := awsClient.DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	// Because the IAM credentials aren't evaluated until training time, we need to ensure we wait for the IAM propagation delay
	time.Sleep(iamPropagationTimeout)

	if in.VpcConfig != nil {
		modelVPCENILock.Lock()
		defer modelVPCENILock.Unlock()
	}

	var out *comprehend.CreateEntityRecognizerOutput
	err := tfresource.Retry(ctx, timeout, func() *resource.RetryError {
		var err error
		out, err = conn.CreateEntityRecognizer(ctx, in)

		if err != nil {
			var tmre *types.TooManyRequestsException
			var qee ratelimit.QuotaExceededError // This is not a typo: the ratelimit.QuotaExceededError is returned as a struct, not a pointer
			if errors.As(err, &tmre) {
				return resource.RetryableError(err)
			} else if errors.As(err, &qee) {
				// Unable to get a rate limit token
				return resource.RetryableError(err)
			} else {
				return resource.NonRetryableError(err)
			}
		}

		return nil
	}, tfresource.WithPollInterval(entityRegcognizerPollInterval))
	if tfresource.TimedOut(err) {
		out, err = conn.CreateEntityRecognizer(ctx, in)
	}
	if err != nil {
		return diag.Errorf("%s Amazon Comprehend Entity Recognizer (%s): %s", action, d.Get("name").(string), err)
	}

	if out == nil || out.EntityRecognizerArn == nil {
		return diag.Errorf("%s Amazon Comprehend Entity Recognizer (%s): empty output", action, d.Get("name").(string))
	}

	d.SetId(aws.ToString(out.EntityRecognizerArn))

	var g multierror.Group
	waitCtx, cancel := context.WithCancel(ctx)

	g.Go(func() error {
		_, err := waitEntityRecognizerCreated(waitCtx, conn, d.Id(), timeout)
		cancel()
		return err
	})

	var diags diag.Diagnostics
	var tobe string
	if action == create.ErrActionCreating {
		tobe = "to be created"
	} else if action == create.ErrActionUpdating {
		tobe = "to be updated"
	} else {
		tobe = "to complete action"
	}

	if in.VpcConfig != nil {
		g.Go(func() error {
			ec2Conn := awsClient.EC2Conn()
			enis, err := findNetworkInterfaces(waitCtx, ec2Conn, in.VpcConfig.SecurityGroupIds, in.VpcConfig.Subnets)
			if err != nil {
				diags = sdkdiag.AppendWarningf(diags, "waiting for Amazon Comprehend Entity Recognizer (%s) %s: %s", d.Id(), tobe, err)
				return nil
			}
			initialENIIds := make(map[string]bool, len(enis))
			for _, v := range enis {
				initialENIIds[aws.ToString(v.NetworkInterfaceId)] = true
			}

			newENI, err := waitNetworkInterfaceCreated(waitCtx, ec2Conn, initialENIIds, in.VpcConfig.SecurityGroupIds, in.VpcConfig.Subnets, d.Timeout(schema.TimeoutCreate))
			if errors.Is(err, context.Canceled) {
				diags = sdkdiag.AppendWarningf(diags, "waiting for Amazon Comprehend Entity Recognizer (%s) %s: %s", d.Id(), tobe, "ENI not found")
				return nil
			}
			if err != nil {
				diags = sdkdiag.AppendWarningf(diags, "waiting for Amazon Comprehend Entity Recognizer (%s) %s: %s", d.Id(), tobe, err)
				return nil
			}

			modelVPCENILock.Unlock()

			_, err = ec2Conn.CreateTagsWithContext(waitCtx, &ec2.CreateTagsInput{
				Resources: []*string{newENI.NetworkInterfaceId},
				Tags: []*ec2.Tag{
					{
						Key:   aws.String(entityRecognizerTagKey),
						Value: aws.String(d.Id()),
					},
				},
			})
			if err != nil {
				diags = sdkdiag.AppendWarningf(diags, "waiting for Amazon Comprehend Entity Recognizer (%s) %s: %s", d.Id(), tobe, err)
				return nil
			}

			return nil
		})
	}

	err = g.Wait().ErrorOrNil()
	if err != nil {
		diags = sdkdiag.AppendErrorf(diags, "waiting for Amazon Comprehend Entity Recognizer (%s) %s: %s", d.Id(), tobe, err)
	}

	return diags
}

func FindEntityRecognizerByID(ctx context.Context, conn *comprehend.Client, id string) (*types.EntityRecognizerProperties, error) {
	in := &comprehend.DescribeEntityRecognizerInput{
		EntityRecognizerArn: aws.String(id),
	}

	out, err := conn.DescribeEntityRecognizer(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.EntityRecognizerProperties == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.EntityRecognizerProperties, nil
}

func ListEntityRecognizerVersionsByName(ctx context.Context, conn *comprehend.Client, name string) ([]types.EntityRecognizerProperties, error) {
	results := []types.EntityRecognizerProperties{}

	input := &comprehend.ListEntityRecognizersInput{
		Filter: &types.EntityRecognizerFilter{
			RecognizerName: aws.String(name),
		},
	}
	paginator := comprehend.NewListEntityRecognizersPaginator(conn, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return []types.EntityRecognizerProperties{}, err
		}
		results = append(results, output.EntityRecognizerPropertiesList...)
	}

	return results, nil
}

func waitEntityRecognizerCreated(ctx context.Context, conn *comprehend.Client, id string, timeout time.Duration) (*types.EntityRecognizerProperties, error) {
	stateConf := &resource.StateChangeConf{
		Pending:      enum.Slice(types.ModelStatusSubmitted, types.ModelStatusTraining),
		Target:       enum.Slice(types.ModelStatusTrained),
		Refresh:      statusEntityRecognizer(ctx, conn, id),
		Delay:        entityRegcognizerCreatedDelay,
		PollInterval: entityRegcognizerPollInterval,
		Timeout:      timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*types.EntityRecognizerProperties); ok {
		if output.Status == types.ModelStatusInError {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Message)))
		}
		return output, err
	}

	return nil, err
}

func waitEntityRecognizerStopped(ctx context.Context, conn *comprehend.Client, id string, timeout time.Duration) (*types.EntityRecognizerProperties, error) {
	stateConf := &resource.StateChangeConf{
		Pending:      enum.Slice(types.ModelStatusSubmitted, types.ModelStatusTraining, types.ModelStatusStopRequested),
		Target:       enum.Slice(types.ModelStatusTrained, types.ModelStatusStopped, types.ModelStatusInError, types.ModelStatusDeleting),
		Refresh:      statusEntityRecognizer(ctx, conn, id),
		Delay:        entityRegcognizerStoppedDelay,
		PollInterval: entityRegcognizerPollInterval,
		Timeout:      timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.EntityRecognizerProperties); ok {
		return out, err
	}

	return nil, err
}

func waitEntityRecognizerDeleted(ctx context.Context, conn *comprehend.Client, id string, timeout time.Duration) (*types.EntityRecognizerProperties, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        enum.Slice(types.ModelStatusSubmitted, types.ModelStatusTraining, types.ModelStatusDeleting, types.ModelStatusInError, types.ModelStatusStopRequested),
		Target:         []string{},
		Refresh:        statusEntityRecognizer(ctx, conn, id),
		Delay:          entityRegcognizerDeletedDelay,
		PollInterval:   entityRegcognizerPollInterval,
		NotFoundChecks: 3,
		Timeout:        timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.EntityRecognizerProperties); ok {
		return out, err
	}

	return nil, err
}

func statusEntityRecognizer(ctx context.Context, conn *comprehend.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindEntityRecognizerByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func flattenEntityRecognizerInputDataConfig(apiObject *types.EntityRecognizerInputDataConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"entity_types":        flattenEntityTypes(apiObject.EntityTypes),
		"annotations":         flattenAnnotations(apiObject.Annotations),
		"augmented_manifests": flattenAugmentedManifests(apiObject.AugmentedManifests),
		"data_format":         apiObject.DataFormat,
		"documents":           flattenDocuments(apiObject.Documents),
		"entity_list":         flattenEntityList(apiObject.EntityList),
	}

	return []interface{}{m}
}

func flattenEntityTypes(apiObjects []types.EntityTypesListItem) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		l = append(l, flattenEntityTypesListItem(&apiObject))
	}

	return l
}

func flattenEntityTypesListItem(apiObject *types.EntityTypesListItem) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"type": aws.ToString(apiObject.Type),
	}

	return m
}

func flattenAnnotations(apiObject *types.EntityRecognizerAnnotations) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"s3_uri": aws.ToString(apiObject.S3Uri),
	}

	if v := apiObject.TestS3Uri; v != nil {
		m["test_s3_uri"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func flattenDocuments(apiObject *types.EntityRecognizerDocuments) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"s3_uri":       aws.ToString(apiObject.S3Uri),
		"input_format": apiObject.InputFormat,
	}

	if v := apiObject.TestS3Uri; v != nil {
		m["test_s3_uri"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func flattenEntityList(apiObject *types.EntityRecognizerEntityList) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"s3_uri": aws.ToString(apiObject.S3Uri),
	}

	return []interface{}{m}
}

func getEntityRecognizerInputDataConfig(d resourceGetter) map[string]any {
	v := d.Get("input_data_config").([]any)
	if len(v) == 0 {
		return nil
	}

	return v[0].(map[string]any)
}

func expandEntityRecognizerInputDataConfig(tfMap map[string]any) *types.EntityRecognizerInputDataConfig {
	if len(tfMap) == 0 {
		return nil
	}

	a := &types.EntityRecognizerInputDataConfig{
		EntityTypes:        expandEntityTypes(tfMap["entity_types"].(*schema.Set)),
		Annotations:        expandAnnotations(tfMap["annotations"].([]interface{})),
		AugmentedManifests: expandAugmentedManifests(tfMap["augmented_manifests"].(*schema.Set)),
		DataFormat:         types.EntityRecognizerDataFormat(tfMap["data_format"].(string)),
		Documents:          expandDocuments(tfMap["documents"].([]interface{})),
		EntityList:         expandEntityList(tfMap["entity_list"].([]interface{})),
	}

	return a
}

func expandEntityTypes(tfSet *schema.Set) []types.EntityTypesListItem {
	if tfSet.Len() == 0 {
		return nil
	}

	var s []types.EntityTypesListItem

	for _, r := range tfSet.List() {
		m, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		a := expandEntityTypesListItem(m)
		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func expandEntityTypesListItem(tfMap map[string]interface{}) *types.EntityTypesListItem {
	if tfMap == nil {
		return nil
	}

	a := &types.EntityTypesListItem{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		a.Type = aws.String(v)
	}

	return a
}

func expandAnnotations(tfList []interface{}) *types.EntityRecognizerAnnotations {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.EntityRecognizerAnnotations{
		S3Uri: aws.String(tfMap["s3_uri"].(string)),
	}

	if v, ok := tfMap["test_s3_uri"].(string); ok && v != "" {
		a.TestS3Uri = aws.String(v)
	}

	return a
}

func expandDocuments(tfList []interface{}) *types.EntityRecognizerDocuments {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.EntityRecognizerDocuments{
		S3Uri:       aws.String(tfMap["s3_uri"].(string)),
		InputFormat: types.InputFormat(tfMap["input_format"].(string)),
	}

	if v, ok := tfMap["test_s3_uri"].(string); ok && v != "" {
		a.TestS3Uri = aws.String(v)
	}

	return a
}

func expandEntityList(tfList []interface{}) *types.EntityRecognizerEntityList {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.EntityRecognizerEntityList{
		S3Uri: aws.String(tfMap["s3_uri"].(string)),
	}

	return a
}

func EntityRecognizerParseARN(arnString string) (string, error) {
	arn, err := arn.Parse(arnString)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`^entity-recognizer/([[:alnum:]-]+)`)
	matches := re.FindStringSubmatch(arn.Resource)
	if len(matches) != 2 {
		return "", fmt.Errorf("unable to parse %q", arnString)
	}
	name := matches[1]

	return name, nil
}
