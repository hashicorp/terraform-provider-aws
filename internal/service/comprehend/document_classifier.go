// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package comprehend

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/aws/ratelimit"
	"github.com/aws/aws-sdk-go-v2/service/comprehend"
	"github.com/aws/aws-sdk-go-v2/service/comprehend/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	documentClassifierTagKey = "tf-aws_comprehend_document_classifier"
)

// @SDKResource("aws_comprehend_document_classifier", name="Document Classifier")
// @Tags(identifierAttribute="id")
func ResourceDocumentClassifier() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDocumentClassifierCreate,
		ReadWithoutTimeout:   resourceDocumentClassifierRead,
		UpdateWithoutTimeout: resourceDocumentClassifierUpdate,
		DeleteWithoutTimeout: resourceDocumentClassifierDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
						"augmented_manifests": {
							Type:         schema.TypeSet,
							Optional:     true,
							ExactlyOneOf: []string{"input_data_config.0.augmented_manifests", "input_data_config.0.s3_uri"},
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
							ValidateDiagFunc: enum.Validate[types.DocumentClassifierDataFormat](),
							Default:          types.DocumentClassifierDataFormatComprehendCsv,
						},
						"label_delimiter": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(documentClassifierLabelSeparators(), false),
						},
						"s3_uri": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"test_s3_uri": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrLanguageCode: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.SyntaxLanguageCode](),
			},
			names.AttrMode: {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.DocumentClassifierMode](),
				Default:          types.DocumentClassifierModeMultiClass,
			},
			"model_kms_key_id": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: tfkms.DiffSuppressKey,
				ValidateFunc:     tfkms.ValidateKey,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validModelName,
			},
			"output_data_config": {
				Type:             schema.TypeList,
				Optional:         true,
				Computed:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKMSKeyID: {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: tfkms.DiffSuppressKeyOrAlias,
							ValidateFunc:     tfkms.ValidateKeyOrAlias,
						},
						"s3_uri": {
							Type:     schema.TypeString,
							Required: true,
							DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
								o := strings.TrimRight(oldValue, "/")
								n := strings.TrimRight(newValue, "/")
								return o == n
							},
						},
						"output_s3_uri": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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
			names.AttrVPCConfig: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnets: {
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
				tfMap := getDocumentClassifierInputDataConfig(diff)
				if tfMap == nil {
					return nil
				}

				if format := types.DocumentClassifierDataFormat(tfMap["data_format"].(string)); format == types.DocumentClassifierDataFormatComprehendCsv {
					if tfMap["s3_uri"] == nil {
						return fmt.Errorf("s3_uri must be set when data_format is %s", format)
					}
				} else {
					if tfMap["augmented_manifests"] == nil {
						return fmt.Errorf("augmented_manifests must be set when data_format is %s", format)
					}
				}

				return nil
			},
			func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
				mode := types.DocumentClassifierMode(diff.Get(names.AttrMode).(string))

				if mode == types.DocumentClassifierModeMultiClass {
					config := diff.GetRawConfig()
					inputDataConfig := config.GetAttr("input_data_config").Index(cty.NumberIntVal(0))
					labelDelimiter := inputDataConfig.GetAttr("label_delimiter")
					if !labelDelimiter.IsNull() {
						return fmt.Errorf("input_data_config.label_delimiter must not be set when mode is %s", types.DocumentClassifierModeMultiClass)
					}
				}

				return nil
			},
		),
	}
}

func resourceDocumentClassifierCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	awsClient := meta.(*conns.AWSClient)
	conn := awsClient.ComprehendClient(ctx)

	var versionName *string
	raw := d.GetRawConfig().GetAttr("version_name")
	if raw.IsNull() {
		versionName = aws.String(create.Name("", d.Get("version_name_prefix").(string)))
	} else if v := raw.AsString(); v != "" {
		versionName = aws.String(v)
	}

	diags := documentClassifierPublishVersion(ctx, conn, d, versionName, create.ErrActionCreating, d.Timeout(schema.TimeoutCreate), awsClient)
	if diags.HasError() {
		return diags
	}

	return append(diags, resourceDocumentClassifierRead(ctx, d, meta)...)
}

func resourceDocumentClassifierRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ComprehendClient(ctx)

	out, err := FindDocumentClassifierByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Comprehend Document Classifier (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Comprehend Document Classifier (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.DocumentClassifierArn)
	d.Set("data_access_role_arn", out.DataAccessRoleArn)
	d.Set(names.AttrLanguageCode, out.LanguageCode)
	d.Set(names.AttrMode, out.Mode)
	d.Set("model_kms_key_id", out.ModelKmsKeyId)
	d.Set("version_name", out.VersionName)
	d.Set("version_name_prefix", create.NamePrefixFromName(aws.ToString(out.VersionName)))
	d.Set("volume_kms_key_id", out.VolumeKmsKeyId)

	// DescribeDocumentClassifier() doesn't return the model name
	name, err := DocumentClassifierParseARN(aws.ToString(out.DocumentClassifierArn))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Comprehend Document Classifier (%s): %s", d.Id(), err)
	}
	d.Set(names.AttrName, name)

	if err := d.Set("input_data_config", flattenDocumentClassifierInputDataConfig(out.InputDataConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting input_data_config: %s", err)
	}

	if err := d.Set("output_data_config", flattenDocumentClassifierOutputDataConfig(out.OutputDataConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting output_data_config: %s", err)
	}

	if err := d.Set(names.AttrVPCConfig, flattenVPCConfig(out.VpcConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_config: %s", err)
	}

	return diags
}

func resourceDocumentClassifierUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	awsClient := meta.(*conns.AWSClient)
	conn := awsClient.ComprehendClient(ctx)

	var diags diag.Diagnostics

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		var versionName *string
		if d.HasChange("version_name") {
			versionName = aws.String(d.Get("version_name").(string))
		} else if v := d.Get("version_name_prefix").(string); v != "" {
			versionName = aws.String(create.Name("", d.Get("version_name_prefix").(string)))
		}

		diags := documentClassifierPublishVersion(ctx, conn, d, versionName, create.ErrActionUpdating, d.Timeout(schema.TimeoutUpdate), awsClient)
		if diags.HasError() {
			return diags
		}
	}

	return append(diags, resourceDocumentClassifierRead(ctx, d, meta)...)
}

func resourceDocumentClassifierDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ComprehendClient(ctx)

	log.Printf("[INFO] Stopping Comprehend Document Classifier (%s)", d.Id())

	_, err := conn.StopTrainingDocumentClassifier(ctx, &comprehend.StopTrainingDocumentClassifierInput{
		DocumentClassifierArn: aws.String(d.Id()),
	})
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "stopping Comprehend Document Classifier (%s): %s", d.Id(), err)
	}

	if _, err := waitDocumentClassifierStopped(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "waiting for Comprehend Document Classifier (%s) to be stopped: %s", d.Id(), err)
	}

	name, err := DocumentClassifierParseARN(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Comprehend Document Classifier (%s): %s", d.Id(), err)
	}

	log.Printf("[INFO] Deleting Comprehend Document Classifier (%s)", name)

	versions, err := ListDocumentClassifierVersionsByName(ctx, conn, name)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Comprehend Document Classifier (%s): %s", name, err)
	}

	var g multierror.Group
	for _, v := range versions {
		v := v
		g.Go(func() error {
			_, err = conn.DeleteDocumentClassifier(ctx, &comprehend.DeleteDocumentClassifierInput{
				DocumentClassifierArn: v.DocumentClassifierArn,
			})
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if !errors.As(err, &nfe) {
					return fmt.Errorf("deleting version (%s): %w", aws.ToString(v.VersionName), err)
				}
			}

			if _, err := waitDocumentClassifierDeleted(ctx, conn, aws.ToString(v.DocumentClassifierArn), d.Timeout(schema.TimeoutDelete)); err != nil {
				return fmt.Errorf("waiting for version (%s) to be deleted: %s", aws.ToString(v.VersionName), err)
			}

			ec2Conn := meta.(*conns.AWSClient).EC2Client(ctx)
			networkInterfaces, err := tfec2.FindNetworkInterfaces(ctx, ec2Conn, &ec2.DescribeNetworkInterfacesInput{
				Filters: []ec2types.Filter{
					tfec2.NewFilter("tag:"+documentClassifierTagKey, []string{aws.ToString(v.DocumentClassifierArn)}),
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
		return sdkdiag.AppendErrorf(diags, "deleting Comprehend Document Classifier (%s): %s", name, err)
	}

	return diags
}

func documentClassifierPublishVersion(ctx context.Context, conn *comprehend.Client, d *schema.ResourceData, versionName *string, action string, timeout time.Duration, awsClient *conns.AWSClient) diag.Diagnostics {
	var diags diag.Diagnostics

	in := &comprehend.CreateDocumentClassifierInput{
		DataAccessRoleArn:      aws.String(d.Get("data_access_role_arn").(string)),
		InputDataConfig:        expandDocumentClassifierInputDataConfig(d),
		LanguageCode:           types.LanguageCode(d.Get(names.AttrLanguageCode).(string)),
		DocumentClassifierName: aws.String(d.Get(names.AttrName).(string)),
		Mode:                   types.DocumentClassifierMode(d.Get(names.AttrMode).(string)),
		OutputDataConfig:       expandDocumentClassifierOutputDataConfig(d.Get("output_data_config").([]interface{})),
		VersionName:            versionName,
		VpcConfig:              expandVPCConfig(d.Get(names.AttrVPCConfig).([]interface{})),
		ClientRequestToken:     aws.String(id.UniqueId()),
		Tags:                   getTagsIn(ctx),
	}

	if v, ok := d.Get("model_kms_key_id").(string); ok && v != "" {
		in.ModelKmsKeyId = aws.String(v)
	}

	if v, ok := d.Get("volume_kms_key_id").(string); ok && v != "" {
		in.VolumeKmsKeyId = aws.String(v)
	}

	// Because the IAM credentials aren't evaluated until training time, we need to ensure we wait for the IAM propagation delay
	time.Sleep(iamPropagationTimeout)

	if in.VpcConfig != nil {
		modelVPCENILock.Lock()
		defer modelVPCENILock.Unlock()
	}

	var out *comprehend.CreateDocumentClassifierOutput
	err := tfresource.Retry(ctx, timeout, func() *retry.RetryError {
		var err error
		out, err = conn.CreateDocumentClassifier(ctx, in)

		if err != nil {
			var tmre *types.TooManyRequestsException
			var qee ratelimit.QuotaExceededError // This is not a typo: the ratelimit.QuotaExceededError is returned as a struct, not a pointer
			if errors.As(err, &tmre) {
				return retry.RetryableError(err)
			} else if errors.As(err, &qee) {
				// Unable to get a rate limit token
				return retry.RetryableError(err)
			} else {
				return retry.NonRetryableError(err)
			}
		}

		return nil
	}, tfresource.WithPollInterval(documentClassifierPollInterval))
	if tfresource.TimedOut(err) {
		out, err = conn.CreateDocumentClassifier(ctx, in)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "%s Amazon Comprehend Document Classifier (%s): %s", action, d.Get(names.AttrName).(string), err)
	}

	if out == nil || out.DocumentClassifierArn == nil {
		return sdkdiag.AppendErrorf(diags, "%s Amazon Comprehend Document Classifier (%s): empty output", action, d.Get(names.AttrName).(string))
	}

	d.SetId(aws.ToString(out.DocumentClassifierArn))

	var g multierror.Group
	waitCtx, cancel := context.WithCancel(ctx)

	g.Go(func() error {
		_, err := waitDocumentClassifierCreated(waitCtx, conn, d.Id(), timeout)
		cancel()
		return err
	})

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
			ec2Conn := awsClient.EC2Client(ctx)
			enis, err := findNetworkInterfaces(waitCtx, ec2Conn, in.VpcConfig.SecurityGroupIds, in.VpcConfig.Subnets)
			if err != nil {
				diags = sdkdiag.AppendWarningf(diags, "waiting for Amazon Comprehend Document Classifier (%s) %s: %s", d.Id(), tobe, err)
				return nil
			}
			initialENIIds := make(map[string]bool, len(enis))
			for _, v := range enis {
				initialENIIds[aws.ToString(v.NetworkInterfaceId)] = true
			}

			newENI, err := waitNetworkInterfaceCreated(waitCtx, ec2Conn, initialENIIds, in.VpcConfig.SecurityGroupIds, in.VpcConfig.Subnets, d.Timeout(schema.TimeoutCreate))
			if errors.Is(err, context.Canceled) {
				diags = sdkdiag.AppendWarningf(diags, "waiting for Amazon Comprehend Document Classifier (%s) %s: %s", d.Id(), tobe, "ENI not found")
				return nil
			}
			if err != nil {
				diags = sdkdiag.AppendWarningf(diags, "waiting for Amazon Comprehend Document Classifier (%s) %s: %s", d.Id(), tobe, err)
				return nil
			}

			modelVPCENILock.Unlock()

			_, err = ec2Conn.CreateTags(waitCtx, &ec2.CreateTagsInput{ // nosemgrep:ci.semgrep.migrate.aws-api-context
				Resources: []string{aws.ToString(newENI.NetworkInterfaceId)},
				Tags: []ec2types.Tag{
					{
						Key:   aws.String(documentClassifierTagKey),
						Value: aws.String(d.Id()),
					},
				},
			})
			if err != nil {
				diags = sdkdiag.AppendWarningf(diags, "waiting for Amazon Comprehend Document Classifier (%s) %s: %s", d.Id(), tobe, err)
				return nil
			}

			return nil
		})
	}

	err = g.Wait().ErrorOrNil()
	if err != nil {
		diags = sdkdiag.AppendErrorf(diags, "waiting for Amazon Comprehend Document Classifier (%s) %s: %s", d.Id(), tobe, err)
	}

	return diags
}

func FindDocumentClassifierByID(ctx context.Context, conn *comprehend.Client, id string) (*types.DocumentClassifierProperties, error) {
	in := &comprehend.DescribeDocumentClassifierInput{
		DocumentClassifierArn: aws.String(id),
	}

	out, err := conn.DescribeDocumentClassifier(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.DocumentClassifierProperties == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.DocumentClassifierProperties, nil
}

func ListDocumentClassifierVersionsByName(ctx context.Context, conn *comprehend.Client, name string) ([]types.DocumentClassifierProperties, error) {
	results := []types.DocumentClassifierProperties{}

	input := &comprehend.ListDocumentClassifiersInput{
		Filter: &types.DocumentClassifierFilter{
			DocumentClassifierName: aws.String(name),
		},
	}
	paginator := comprehend.NewListDocumentClassifiersPaginator(conn, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return []types.DocumentClassifierProperties{}, err
		}
		results = append(results, output.DocumentClassifierPropertiesList...)
	}

	return results, nil
}

func waitDocumentClassifierCreated(ctx context.Context, conn *comprehend.Client, id string, timeout time.Duration) (*types.DocumentClassifierProperties, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(types.ModelStatusSubmitted, types.ModelStatusTraining),
		Target:       enum.Slice(types.ModelStatusTrained),
		Refresh:      statusDocumentClassifier(ctx, conn, id),
		Delay:        documentClassifierCreatedDelay,
		PollInterval: documentClassifierPollInterval,
		Timeout:      timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*types.DocumentClassifierProperties); ok {
		if output.Status == types.ModelStatusInError {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Message)))
		}
		return output, err
	}

	return nil, err
}

func waitDocumentClassifierStopped(ctx context.Context, conn *comprehend.Client, id string, timeout time.Duration) (*types.DocumentClassifierProperties, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(types.ModelStatusSubmitted, types.ModelStatusTraining, types.ModelStatusStopRequested),
		Target:       enum.Slice(types.ModelStatusTrained, types.ModelStatusStopped, types.ModelStatusInError, types.ModelStatusDeleting),
		Refresh:      statusDocumentClassifier(ctx, conn, id),
		Delay:        documentClassifierStoppedDelay,
		PollInterval: documentClassifierPollInterval,
		Timeout:      timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.DocumentClassifierProperties); ok {
		return out, err
	}

	return nil, err
}

func waitDocumentClassifierDeleted(ctx context.Context, conn *comprehend.Client, id string, timeout time.Duration) (*types.DocumentClassifierProperties, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(types.ModelStatusSubmitted, types.ModelStatusTraining, types.ModelStatusDeleting, types.ModelStatusInError, types.ModelStatusStopRequested),
		Target:         []string{},
		Refresh:        statusDocumentClassifier(ctx, conn, id),
		Delay:          documentClassifierDeletedDelay,
		PollInterval:   documentClassifierPollInterval,
		NotFoundChecks: 3,
		Timeout:        timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.DocumentClassifierProperties); ok {
		return out, err
	}

	return nil, err
}

func statusDocumentClassifier(ctx context.Context, conn *comprehend.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindDocumentClassifierByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func flattenDocumentClassifierInputDataConfig(apiObject *types.DocumentClassifierInputDataConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"augmented_manifests": flattenAugmentedManifests(apiObject.AugmentedManifests),
		"data_format":         apiObject.DataFormat,
		"s3_uri":              aws.ToString(apiObject.S3Uri),
	}

	if apiObject.LabelDelimiter != nil {
		m["label_delimiter"] = aws.ToString(apiObject.LabelDelimiter)
	}

	if apiObject.TestS3Uri != nil {
		m["test_s3_uri"] = aws.ToString(apiObject.TestS3Uri)
	}

	return []interface{}{m}
}

func flattenDocumentClassifierOutputDataConfig(apiObject *types.DocumentClassifierOutputDataConfig) []interface{} {
	if apiObject == nil || apiObject.S3Uri == nil {
		return nil
	}

	// On return, `S3Uri` contains the full path of the output documents, not the storage location
	s3Uri := aws.ToString(apiObject.S3Uri)
	m := map[string]interface{}{
		"output_s3_uri": s3Uri,
	}

	re := regexache.MustCompile(`^(s3://[0-9a-z.-]{3,63}(/.+)?/)[0-9A-Za-z-]+/output/output\.tar\.gz`)
	match := re.FindStringSubmatch(s3Uri)
	if match != nil && match[1] != "" {
		m["s3_uri"] = match[1]
	}

	if apiObject.KmsKeyId != nil {
		m[names.AttrKMSKeyID] = aws.ToString(apiObject.KmsKeyId)
	}

	return []interface{}{m}
}

func getDocumentClassifierInputDataConfig(d resourceGetter) map[string]any {
	v := d.Get("input_data_config").([]any)
	if len(v) == 0 {
		return nil
	}

	return v[0].(map[string]any)
}

func expandDocumentClassifierInputDataConfig(d *schema.ResourceData) *types.DocumentClassifierInputDataConfig {
	tfMap := getDocumentClassifierInputDataConfig(d)
	if len(tfMap) == 0 {
		return nil
	}

	a := &types.DocumentClassifierInputDataConfig{
		AugmentedManifests: expandAugmentedManifests(tfMap["augmented_manifests"].(*schema.Set)),
		DataFormat:         types.DocumentClassifierDataFormat(tfMap["data_format"].(string)),
		S3Uri:              aws.String(tfMap["s3_uri"].(string)),
	}

	if v, ok := tfMap["label_delimiter"].(string); ok && v != "" {
		a.LabelDelimiter = aws.String(v)
	}

	if v, ok := tfMap["test_s3_uri"].(string); ok && v != "" {
		a.TestS3Uri = aws.String(v)
	}

	return a
}

func expandDocumentClassifierOutputDataConfig(tfList []interface{}) *types.DocumentClassifierOutputDataConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.DocumentClassifierOutputDataConfig{
		S3Uri: aws.String(tfMap["s3_uri"].(string)),
	}

	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
		a.KmsKeyId = aws.String(v)
	}

	return a
}

func DocumentClassifierParseARN(arnString string) (string, error) {
	arn, err := arn.Parse(arnString)
	if err != nil {
		return "", err
	}
	re := regexache.MustCompile(`^document-classifier/([[:alnum:]-]+)`)
	matches := re.FindStringSubmatch(arn.Resource)
	if len(matches) != 2 {
		return "", fmt.Errorf("unable to parse %q", arnString)
	}
	name := matches[1]

	return name, nil
}

const DocumentClassifierLabelSeparatorDefault = "|"

func documentClassifierLabelSeparators() []string {
	return []string{
		DocumentClassifierLabelSeparatorDefault,
		"~",
		"!",
		"@",
		"#",
		"$",
		"%",
		"^",
		"*",
		"-",
		"_",
		"+",
		"=",
		"\\",
		":",
		";",
		">",
		"?",
		"/",
		" ",
		"\t",
	}
}
