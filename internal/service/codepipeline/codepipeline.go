// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codepipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	providerGitHub                      = "GitHub"
	gitHubActionConfigurationOAuthToken = "OAuthToken"
)

// @SDKResource("aws_codepipeline", name="Pipeline")
// @Tags(identifierAttribute="arn")
func resourcePipeline() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePipelineCreate,
		ReadWithoutTimeout:   resourcePipelineRead,
		UpdateWithoutTimeout: resourcePipelineUpdate,
		DeleteWithoutTimeout: resourcePipelineDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			conditionsSchema := func() map[string]*schema.Schema {
				return map[string]*schema.Schema{
					"result": {
						Type:             schema.TypeString,
						Optional:         true,
						ValidateDiagFunc: enum.Validate[types.Result](),
					},
					names.AttrRule: {
						Type:     schema.TypeList,
						MinItems: 1,
						MaxItems: 5,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"commands": {
									Type:     schema.TypeList,
									Optional: true,
									MaxItems: 50,
									Elem: &schema.Schema{
										Type: schema.TypeString,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 1000),
										),
									},
								},
								names.AttrConfiguration: {
									Type:     schema.TypeMap,
									Optional: true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 10000),
										),
									},
								},
								"input_artifacts": {
									Type:     schema.TypeList,
									Optional: true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
										ValidateFunc: validation.All(
											validation.StringMatch(regexache.MustCompile(`[a-zA-Z0-9_\-]+`), ""),
											validation.StringLenBetween(1, 100),
										),
									},
								},
								names.AttrName: {
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringMatch(regexache.MustCompile(`[A-Za-z0-9.@\-_]+`), ""),
								},
								names.AttrRegion: {
									Type:     schema.TypeString,
									Optional: true,
								},
								names.AttrRoleARN: {
									Type:         schema.TypeString,
									Optional:     true,
									ValidateFunc: verify.ValidARN,
								},
								"rule_type_id": {
									Type:     schema.TypeList,
									Required: true,
									MaxItems: 1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"category": {
												Type:             schema.TypeString,
												Required:         true,
												ValidateDiagFunc: enum.Validate[types.RuleCategory](),
											},
											names.AttrOwner: {
												Type:             schema.TypeString,
												Optional:         true,
												ValidateDiagFunc: enum.Validate[types.RuleOwner](),
											},
											"provider": {
												Type:     schema.TypeString,
												Required: true,
											},
											names.AttrVersion: {
												Type:     schema.TypeString,
												Optional: true,
												ValidateFunc: validation.All(
													validation.StringLenBetween(1, 9),
													validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_-]+`), ""),
												),
											},
										},
									},
								},
								"timeout_in_minutes": {
									Type:         schema.TypeInt,
									Optional:     true,
									ValidateFunc: validation.IntBetween(5, 86400),
								},
							},
						},
					},
				}
			}

			triggerSchema := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 50,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"git_configuration": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"pull_request": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 3,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"branches": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"excludes": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 8,
																	Elem: &schema.Schema{
																		Type:         schema.TypeString,
																		ValidateFunc: validation.StringLenBetween(1, 255),
																	},
																},
																"includes": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 8,
																	Elem: &schema.Schema{
																		Type:         schema.TypeString,
																		ValidateFunc: validation.StringLenBetween(1, 255),
																	},
																},
															},
														},
													},
													"events": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 3,
														Elem: &schema.Schema{
															Type:             schema.TypeString,
															ValidateDiagFunc: enum.Validate[types.GitPullRequestEventType](),
														},
													},
													"file_paths": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"excludes": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 8,
																	Elem: &schema.Schema{
																		Type:         schema.TypeString,
																		ValidateFunc: validation.StringLenBetween(1, 255),
																	},
																},
																"includes": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 8,
																	Elem: &schema.Schema{
																		Type:         schema.TypeString,
																		ValidateFunc: validation.StringLenBetween(1, 255),
																	},
																},
															},
														},
													},
												},
											},
										},
										"push": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 3,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"branches": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"excludes": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 8,
																	Elem: &schema.Schema{
																		Type:         schema.TypeString,
																		ValidateFunc: validation.StringLenBetween(1, 255),
																	},
																},
																"includes": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 8,
																	Elem: &schema.Schema{
																		Type:         schema.TypeString,
																		ValidateFunc: validation.StringLenBetween(1, 255),
																	},
																},
															},
														},
													},
													"file_paths": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"excludes": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 8,
																	Elem: &schema.Schema{
																		Type:         schema.TypeString,
																		ValidateFunc: validation.StringLenBetween(1, 255),
																	},
																},
																"includes": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 8,
																	Elem: &schema.Schema{
																		Type:         schema.TypeString,
																		ValidateFunc: validation.StringLenBetween(1, 255),
																	},
																},
															},
														},
													},
													names.AttrTags: {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"excludes": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 8,
																	Elem: &schema.Schema{
																		Type:         schema.TypeString,
																		ValidateFunc: validation.StringLenBetween(1, 255),
																	},
																},
																"includes": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 8,
																	Elem: &schema.Schema{
																		Type:         schema.TypeString,
																		ValidateFunc: validation.StringLenBetween(1, 255),
																	},
																},
															},
														},
													},
												},
											},
										},
										"source_action_name": {
											Type:     schema.TypeString,
											Required: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 100),
												validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_.@-]+`), ""),
											),
										},
									},
								},
							},
							"provider_type": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.PipelineTriggerProviderType](),
							},
						},
					},
				}
			}

			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"artifact_store": {
					Type:     schema.TypeSet,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"encryption_key": {
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrID: {
											Type:     schema.TypeString,
											Required: true,
										},
										names.AttrType: {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[types.EncryptionKeyType](),
										},
									},
								},
							},
							names.AttrLocation: {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrRegion: {
								Type:     schema.TypeString,
								Optional: true,
							},
							names.AttrType: {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.ArtifactStoreType](),
							},
						},
					},
				},
				"execution_mode": {
					Type:             schema.TypeString,
					Optional:         true,
					Default:          types.ExecutionModeSuperseded,
					ValidateDiagFunc: enum.Validate[types.ExecutionMode](),
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 100),
						validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_.@-]+`), ""),
					),
				},
				"pipeline_type": {
					Type:             schema.TypeString,
					Optional:         true,
					Default:          types.PipelineTypeV1,
					ValidateDiagFunc: enum.Validate[types.PipelineType](),
				},
				names.AttrRoleARN: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
				names.AttrStage: {
					Type:     schema.TypeList,
					MinItems: 2,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrAction: {
								Type:     schema.TypeList,
								Required: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"category": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[types.ActionCategory](),
										},
										names.AttrConfiguration: {
											Type:     schema.TypeMap,
											Optional: true,
											ValidateDiagFunc: validation.AllDiag(
												validation.MapKeyLenBetween(1, 50),
												validation.MapKeyLenBetween(1, 1000),
											),
											Elem:             &schema.Schema{Type: schema.TypeString},
											DiffSuppressFunc: pipelineSuppressStageActionConfigurationDiff,
										},
										"input_artifacts": {
											Type:     schema.TypeList,
											Optional: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										names.AttrName: {
											Type:     schema.TypeString,
											Required: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 100),
												validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_.@-]+`), ""),
											),
										},
										names.AttrNamespace: {
											Type:     schema.TypeString,
											Optional: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 100),
												validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_@-]+`), ""),
											),
										},
										"output_artifacts": {
											Type:     schema.TypeList,
											Optional: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										names.AttrOwner: {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[types.ActionOwner](),
										},
										"provider": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: pipelineValidateActionProvider,
										},
										names.AttrRegion: {
											Type:     schema.TypeString,
											Optional: true,
											Computed: true,
										},
										names.AttrRoleARN: {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: verify.ValidARN,
										},
										"run_order": {
											Type:         schema.TypeInt,
											Optional:     true,
											Computed:     true,
											ValidateFunc: validation.IntBetween(1, 999),
										},
										"timeout_in_minutes": {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IntBetween(5, 86400),
										},
										names.AttrVersion: {
											Type:     schema.TypeString,
											Required: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 9),
												validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_-]+`), ""),
											),
										},
									},
								},
							},
							"before_entry": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrCondition: {
											Type:     schema.TypeList,
											Required: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: conditionsSchema(),
											},
										},
									},
								},
							},
							names.AttrName: {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 100),
									validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_.@-]+`), ""),
								),
							},
							"on_failure": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrCondition: {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: conditionsSchema(),
											},
										},
										"result": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[types.Result](),
										},
										"retry_configuration": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"retry_mode": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.StageRetryMode](),
													},
												},
											},
										},
									},
								},
							},
							"on_success": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrCondition: {
											Type:     schema.TypeList,
											Required: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: conditionsSchema(),
											},
										},
									},
								},
							},
						},
					},
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"trigger":         triggerSchema(),
				"trigger_all":     sdkv2.DataSourcePropertyFromResourceProperty(triggerSchema()),
				"variable": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDefaultValue: {
								Type:     schema.TypeString,
								Optional: true,
							},
							names.AttrDescription: {
								Type:     schema.TypeString,
								Optional: true,
							},
							names.AttrName: {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
			}
		},
	}
}

func resourcePipelineCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodePipelineClient(ctx)

	pipeline, err := expandPipelineDeclaration(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	name := d.Get(names.AttrName).(string)
	input := &codepipeline.CreatePipelineInput{
		Pipeline: pipeline,
		Tags:     getTagsIn(ctx),
	}

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidStructureException](ctx, propagationTimeout, func() (any, error) {
		return conn.CreatePipeline(ctx, input)
	}, "not authorized")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodePipeline Pipeline (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*codepipeline.CreatePipelineOutput).Pipeline.Name))

	return append(diags, resourcePipelineRead(ctx, d, meta)...)
}

func resourcePipelineRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodePipelineClient(ctx)

	output, err := findPipelineByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodePipeline Pipeline %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodePipeline Pipeline (%s): %s", d.Id(), err)
	}

	metadata := output.Metadata
	pipeline := output.Pipeline
	arn := aws.ToString(metadata.PipelineArn)
	d.Set(names.AttrARN, arn)
	if pipeline.ArtifactStore != nil {
		if err := d.Set("artifact_store", []any{flattenArtifactStore(pipeline.ArtifactStore)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting artifact_store: %s", err)
		}
	} else if pipeline.ArtifactStores != nil {
		if err := d.Set("artifact_store", flattenArtifactStores(pipeline.ArtifactStores)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting artifact_store: %s", err)
		}
	}
	d.Set("execution_mode", pipeline.ExecutionMode)
	d.Set(names.AttrName, pipeline.Name)
	d.Set("pipeline_type", pipeline.PipelineType)
	d.Set(names.AttrRoleARN, pipeline.RoleArn)
	if err := d.Set(names.AttrStage, flattenStageDeclarations(d, pipeline.Stages)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting stage: %s", err)
	}
	d.Set("trigger", d.Get("trigger"))
	if err := d.Set("trigger_all", flattenTriggerDeclarations(pipeline.Triggers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting trigger_all: %s", err)
	}
	if err := d.Set("variable", flattenVariableDeclarations(pipeline.Variables)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting variable: %s", err)
	}

	return diags
}

func resourcePipelineUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodePipelineClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		pipeline, err := expandPipelineDeclaration(d)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &codepipeline.UpdatePipelineInput{
			Pipeline: pipeline,
		}

		_, err = conn.UpdatePipeline(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodePipeline Pipeline (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePipelineRead(ctx, d, meta)...)
}

func resourcePipelineDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodePipelineClient(ctx)

	log.Printf("[INFO] Deleting CodePipeline Pipeline: %s", d.Id())
	input := codepipeline.DeletePipelineInput{
		Name: aws.String(d.Id()),
	}
	_, err := conn.DeletePipeline(ctx, &input)

	if errs.IsA[*types.PipelineNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodePipeline Pipeline (%s): %s", d.Id(), err)
	}

	return diags
}

func findPipelineByName(ctx context.Context, conn *codepipeline.Client, name string) (*codepipeline.GetPipelineOutput, error) {
	input := &codepipeline.GetPipelineInput{
		Name: aws.String(name),
	}

	output, err := conn.GetPipeline(ctx, input)

	if errs.IsA[*types.PipelineNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Metadata == nil || output.Pipeline == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func pipelineValidateActionProvider(i any, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	v, ok := i.(string)
	if !ok {
		return sdkdiag.AppendErrorf(diags, "expected type to be string")
	}

	if v == providerGitHub {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "The CodePipeline GitHub version 1 action provider is no longer recommended.",
				Detail:   "Use a GitHub version 2 action (with a CodeStar Connection `aws_codestarconnections_connection`) as recommended instead. See https://docs.aws.amazon.com/codepipeline/latest/userguide/update-github-action-connections.html",
			},
		}
	}

	return diags
}

func pipelineSuppressStageActionConfigurationDiff(k, old, new string, d *schema.ResourceData) bool {
	parts := strings.Split(k, ".")
	parts = parts[:len(parts)-2]
	providerAddr := strings.Join(append(parts, "provider"), ".")
	provider := d.Get(providerAddr).(string)

	if provider == providerGitHub && strings.HasSuffix(k, gitHubActionConfigurationOAuthToken) {
		hash := hashGitHubToken(new)
		return old == hash
	}

	return false
}

func hashGitHubToken(token string) string {
	const gitHubTokenHashPrefix = "hash-"

	// Without this check, the value was getting encoded twice
	if strings.HasPrefix(token, gitHubTokenHashPrefix) {
		return token
	}

	sum := sha256.Sum256([]byte(token))
	return gitHubTokenHashPrefix + hex.EncodeToString(sum[:])
}

func expandPipelineDeclaration(d *schema.ResourceData) (*types.PipelineDeclaration, error) {
	pipelineType := types.PipelineType(d.Get("pipeline_type").(string))
	apiObject := &types.PipelineDeclaration{
		Name:         aws.String(d.Get(names.AttrName).(string)),
		PipelineType: pipelineType,
		RoleArn:      aws.String(d.Get(names.AttrRoleARN).(string)),
		Stages:       expandStageDeclarations(d.Get(names.AttrStage).([]any)),
	}

	if v, ok := d.GetOk("artifact_store"); ok && v.(*schema.Set).Len() > 0 {
		artifactStores := expandArtifactStores(v.(*schema.Set).List())

		switch n := len(artifactStores); n {
		case 1:
			for region, v := range artifactStores {
				if region != "" {
					return nil, errors.New("region cannot be set for a single-region CodePipeline Pipeline")
				}
				apiObject.ArtifactStore = &v
			}

		default:
			for region := range artifactStores {
				if region == "" {
					return nil, errors.New("region must be set for a cross-region CodePipeline Pipeline")
				}
			}
			if n != v.(*schema.Set).Len() {
				return nil, errors.New("only one Artifact Store can be defined per region for a cross-region CodePipeline Pipeline")
			}
			apiObject.ArtifactStores = artifactStores
		}
	}

	if v, ok := d.GetOk("execution_mode"); ok {
		apiObject.ExecutionMode = types.ExecutionMode(v.(string))
	}

	// explicitly send trigger for all V2 pipelines (even when unset) to ensure
	// removed custom triggers are handled correctly
	if v, ok := d.GetOk("trigger"); (ok && len(v.([]any)) > 0) || pipelineType == types.PipelineTypeV2 {
		apiObject.Triggers = expandTriggerDeclarations(v.([]any))
	}

	if v, ok := d.GetOk("variable"); ok && len(v.([]any)) > 0 {
		apiObject.Variables = expandVariableDeclarations(v.([]any))
	}

	return apiObject, nil
}

func expandArtifactStore(tfMap map[string]any) *types.ArtifactStore {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ArtifactStore{}

	if v, ok := tfMap["encryption_key"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.EncryptionKey = expandEncryptionKey(v[0].(map[string]any))
	}

	if v, ok := tfMap[names.AttrLocation].(string); ok && v != "" {
		apiObject.Location = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = types.ArtifactStoreType(v)
	}

	return apiObject
}

func expandArtifactStores(tfList []any) map[string]types.ArtifactStore {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make(map[string]types.ArtifactStore, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandArtifactStore(tfMap)

		if apiObject == nil {
			continue
		}

		var region string

		if v, ok := tfMap[names.AttrRegion].(string); ok && v != "" {
			region = v
		}

		apiObjects[region] = *apiObject // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-assignment
	}

	return apiObjects
}

func expandEncryptionKey(tfMap map[string]any) *types.EncryptionKey {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EncryptionKey{}

	if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
		apiObject.Id = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = types.EncryptionKeyType(v)
	}

	return apiObject
}

func expandStageDeclaration(tfMap map[string]any) *types.StageDeclaration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.StageDeclaration{}

	if v, ok := tfMap[names.AttrAction].([]any); ok && len(v) > 0 {
		apiObject.Actions = expandActionDeclarations(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["before_entry"].([]any); ok && len(v) > 0 {
		apiObject.BeforeEntry = expandBeforeEntryDeclaration(v[0].(map[string]any))
	}

	if v, ok := tfMap["on_success"].([]any); ok && len(v) > 0 {
		apiObject.OnSuccess = expandOnSuccessDeclaration(v[0].(map[string]any))
	}

	if v, ok := tfMap["on_failure"].([]any); ok && len(v) > 0 {
		apiObject.OnFailure = expandOnFailureDeclaration(v[0].(map[string]any))
	}

	return apiObject
}

func expandStageDeclarations(tfList []any) []types.StageDeclaration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.StageDeclaration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandStageDeclaration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandActionDeclaration(tfMap map[string]any) *types.ActionDeclaration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ActionDeclaration{
		ActionTypeId: &types.ActionTypeId{},
	}

	if v, ok := tfMap["category"].(string); ok && v != "" {
		apiObject.ActionTypeId.Category = types.ActionCategory(v)
	}

	if v, ok := tfMap[names.AttrConfiguration].(map[string]any); ok && len(v) > 0 {
		apiObject.Configuration = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["input_artifacts"].([]any); ok && len(v) > 0 {
		apiObject.InputArtifacts = expandInputArtifacts(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrNamespace].(string); ok && v != "" {
		apiObject.Namespace = aws.String(v)
	}

	if v, ok := tfMap["output_artifacts"].([]any); ok && len(v) > 0 {
		apiObject.OutputArtifacts = expandOutputArtifacts(v)
	}

	if v, ok := tfMap[names.AttrOwner].(string); ok && v != "" {
		apiObject.ActionTypeId.Owner = types.ActionOwner(v)
	}

	if v, ok := tfMap["provider"].(string); ok && v != "" {
		apiObject.ActionTypeId.Provider = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRegion].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["run_order"].(int); ok && v != 0 {
		apiObject.RunOrder = aws.Int32(int32(v))
	}

	if v, ok := tfMap["timeout_in_minutes"].(int); ok && v != 0 {
		apiObject.TimeoutInMinutes = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrVersion].(string); ok && v != "" {
		apiObject.ActionTypeId.Version = aws.String(v)
	}

	return apiObject
}

func expandActionDeclarations(tfList []any) []types.ActionDeclaration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.ActionDeclaration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandActionDeclaration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandInputArtifacts(tfList []any) []types.InputArtifact {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.InputArtifact

	for _, v := range tfList {
		v, ok := v.(string)

		if !ok {
			continue
		}

		apiObject := types.InputArtifact{
			Name: aws.String(v),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandOutputArtifacts(tfList []any) []types.OutputArtifact {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.OutputArtifact

	for _, v := range tfList {
		v, ok := v.(string)

		if !ok {
			continue
		}

		apiObject := types.OutputArtifact{
			Name: aws.String(v),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandVariableDeclaration(tfMap map[string]any) *types.PipelineVariableDeclaration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipelineVariableDeclaration{}

	if v, ok := tfMap[names.AttrDefaultValue].(string); ok && v != "" {
		apiObject.DefaultValue = aws.String(v)
	}

	if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandVariableDeclarations(tfList []any) []types.PipelineVariableDeclaration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.PipelineVariableDeclaration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandVariableDeclaration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandGitBranchFilterCriteria(tfMap map[string]any) *types.GitBranchFilterCriteria {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.GitBranchFilterCriteria{}

	if v, ok := tfMap["excludes"].([]any); ok && len(v) != 0 {
		for _, exclude := range v {
			apiObject.Excludes = append(apiObject.Excludes, exclude.(string))
		}
	}

	if v, ok := tfMap["includes"].([]any); ok && len(v) != 0 {
		for _, include := range v {
			apiObject.Includes = append(apiObject.Includes, include.(string))
		}
	}

	return apiObject
}

func expandGitFilePathFilterCriteria(tfMap map[string]any) *types.GitFilePathFilterCriteria {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.GitFilePathFilterCriteria{}

	if v, ok := tfMap["excludes"].([]any); ok && len(v) != 0 {
		for _, exclude := range v {
			apiObject.Excludes = append(apiObject.Excludes, exclude.(string))
		}
	}

	if v, ok := tfMap["includes"].([]any); ok && len(v) != 0 {
		for _, include := range v {
			apiObject.Includes = append(apiObject.Includes, include.(string))
		}
	}

	return apiObject
}

func expandGitTagFilterCriteria(tfMap map[string]any) *types.GitTagFilterCriteria {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.GitTagFilterCriteria{}

	if v, ok := tfMap["excludes"].([]any); ok && len(v) != 0 {
		for _, exclude := range v {
			apiObject.Excludes = append(apiObject.Excludes, exclude.(string))
		}
	}

	if v, ok := tfMap["includes"].([]any); ok && len(v) != 0 {
		for _, include := range v {
			apiObject.Includes = append(apiObject.Includes, include.(string))
		}
	}

	return apiObject
}

func expandGitPullRequestEventTypes(tfList []any) []types.GitPullRequestEventType {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := []types.GitPullRequestEventType{}

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, types.GitPullRequestEventType(tfMap))
	}

	return apiObjects
}

func expandGitPullRequestFilters(tfList []any) []types.GitPullRequestFilter {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := []types.GitPullRequestFilter{}

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := types.GitPullRequestFilter{}

		if v, ok := tfMap["branches"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.Branches = expandGitBranchFilterCriteria(v[0].(map[string]any))
		}

		if v, ok := tfMap["events"].([]any); ok && len(v) > 0 && v != nil {
			apiObject.Events = expandGitPullRequestEventTypes(v)
		}

		if v, ok := tfMap["file_paths"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.FilePaths = expandGitFilePathFilterCriteria(v[0].(map[string]any))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandGitPushFilters(tfList []any) []types.GitPushFilter {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := []types.GitPushFilter{}

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := types.GitPushFilter{}

		if v, ok := tfMap["branches"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.Branches = expandGitBranchFilterCriteria(v[0].(map[string]any))
		}

		if v, ok := tfMap["file_paths"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.FilePaths = expandGitFilePathFilterCriteria(v[0].(map[string]any))
		}

		if v, ok := tfMap[names.AttrTags].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.Tags = expandGitTagFilterCriteria(v[0].(map[string]any))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandGitConfigurationDeclaration(tfMap map[string]any) *types.GitConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.GitConfiguration{}

	if v, ok := tfMap["pull_request"].([]any); ok && len(v) > 0 && v != nil {
		apiObject.PullRequest = expandGitPullRequestFilters(v)
	}

	if v, ok := tfMap["push"].([]any); ok && len(v) > 0 && v != nil {
		apiObject.Push = expandGitPushFilters(v)
	}

	apiObject.SourceActionName = aws.String(tfMap["source_action_name"].(string))

	return apiObject
}

func expandTriggerDeclaration(tfMap map[string]any) *types.PipelineTriggerDeclaration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PipelineTriggerDeclaration{}

	if v, ok := tfMap["git_configuration"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.GitConfiguration = expandGitConfigurationDeclaration(v[0].(map[string]any))
	}

	apiObject.ProviderType = types.PipelineTriggerProviderType(tfMap["provider_type"].(string))

	return apiObject
}

func expandTriggerDeclarations(tfList []any) []types.PipelineTriggerDeclaration {
	var apiObjects []types.PipelineTriggerDeclaration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandTriggerDeclaration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandConditionRuleTypeId(tfMap map[string]any) *types.RuleTypeId {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.RuleTypeId{}

	if v, ok := tfMap["category"].(string); ok && v != "" {
		apiObject.Category = types.RuleCategory(v)
	}

	if v, ok := tfMap[names.AttrOwner].(string); ok && v != "" {
		apiObject.Owner = types.RuleOwner(v)
	}

	if v, ok := tfMap["provider"].(string); ok && v != "" {
		apiObject.Provider = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVersion].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandConditionRuleInputArtifacts(tfList []any) []types.InputArtifact {
	if len(tfList) == 0 {
		return nil
	}
	var apiObjects []types.InputArtifact

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObject := types.InputArtifact{
			Name: aws.String(tfMap),
		}

		apiObjects = append(apiObjects, apiObject)
	}
	return apiObjects
}

func expandConditionRule(tfMap map[string]any) *types.RuleDeclaration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.RuleDeclaration{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["rule_type_id"].([]any); ok && len(v) > 0 {
		apiObject.RuleTypeId = expandConditionRuleTypeId(v[0].(map[string]any))
	}

	if v, ok := tfMap["commands"].([]any); ok && len(v) > 0 {
		for _, command := range v {
			apiObject.Commands = append(apiObject.Commands, command.(string))
		}
	}

	if v, ok := tfMap[names.AttrConfiguration].(map[string]any); ok && v != nil {
		apiObject.Configuration = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["input_artifacts"].([]any); ok && len(v) > 0 {
		apiObject.InputArtifacts = expandConditionRuleInputArtifacts(v)
	}

	if v, ok := tfMap[names.AttrRegion].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["timeout_in_minutes"].(int32); ok && v != 0 {
		apiObject.TimeoutInMinutes = aws.Int32(v)
	}

	return apiObject
}

func expandConditionRules(tfList []any) []types.RuleDeclaration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.RuleDeclaration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandConditionRule(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandCondition(tfMap map[string]any) *types.Condition {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Condition{}

	if v, ok := tfMap["result"].(string); ok && v != "" {
		apiObject.Result = types.Result(v)
	}

	if v, ok := tfMap[names.AttrRule].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Rules = expandConditionRules(v)
	}

	return apiObject
}

func expandConditions(tfList []any) []types.Condition {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.Condition

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}
		apiObject := expandCondition(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandBeforeEntryDeclaration(tfMap map[string]any) *types.BeforeEntryConditions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.BeforeEntryConditions{}

	if v, ok := tfMap[names.AttrCondition].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Conditions = expandConditions(v)
	}

	return apiObject
}

func expandOnSuccessDeclaration(tfMap map[string]any) *types.SuccessConditions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.SuccessConditions{}

	if v, ok := tfMap[names.AttrCondition].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Conditions = expandConditions(v)
	}

	return apiObject
}

func expandRetryConfiguration(tfMap map[string]any) *types.RetryConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.RetryConfiguration{}

	if v, ok := tfMap["retry_mode"].(string); ok && v != "" {
		apiObject.RetryMode = types.StageRetryMode(v)
	}

	return apiObject
}

func expandOnFailureDeclaration(tfMap map[string]any) *types.FailureConditions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.FailureConditions{}

	if v, ok := tfMap[names.AttrCondition].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Conditions = expandConditions(v)
	}

	if v, ok := tfMap["result"].(string); ok && v != "" {
		apiObject.Result = types.Result(v)
	}

	if v, ok := tfMap["retry_configuration"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.RetryConfiguration = expandRetryConfiguration(v[0].(map[string]any))
	}

	return apiObject
}

func flattenArtifactStore(apiObject *types.ArtifactStore) map[string]any {
	tfMap := map[string]any{
		names.AttrType: apiObject.Type,
	}

	if v := apiObject.EncryptionKey; v != nil {
		tfMap["encryption_key"] = []any{flattenEncryptionKey(v)}
	}

	if v := apiObject.Location; v != nil {
		tfMap[names.AttrLocation] = aws.ToString(v)
	}

	return tfMap
}

func flattenArtifactStores(apiObjects map[string]types.ArtifactStore) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for region, apiObject := range apiObjects {
		tfMap := flattenArtifactStore(&apiObject)
		tfMap[names.AttrRegion] = region

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenEncryptionKey(apiObject *types.EncryptionKey) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrType: apiObject.Type,
	}

	if v := apiObject.Id; v != nil {
		tfMap[names.AttrID] = aws.ToString(v)
	}

	return tfMap
}

func flattenConditionRuleTypeId(apiObject *types.RuleTypeId) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Category; v != "" {
		tfMap["category"] = v
	}

	if v := apiObject.Owner; v != "" {
		tfMap[names.AttrOwner] = v
	}

	if v := apiObject.Provider; v != nil {
		tfMap["provider"] = aws.ToString(v)
	}

	if v := apiObject.Version; v != nil {
		tfMap[names.AttrVersion] = aws.ToString(v)
	}

	return tfMap
}

func flattenConditionRule(apiObjects types.RuleDeclaration) map[string]any {
	tfMap := map[string]any{}

	if v := apiObjects.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObjects.RuleTypeId; v != nil {
		tfMap["rule_type_id"] = []any{flattenConditionRuleTypeId(v)}
	}

	if v := apiObjects.Commands; v != nil {
		var tfList []any
		for _, command := range apiObjects.Commands {
			tfList = append(tfList, command)
		}
		tfMap["commands"] = tfList
	}

	if v := apiObjects.Configuration; v != nil {
		tfMap[names.AttrConfiguration] = v
	}

	if v := apiObjects.InputArtifacts; v != nil {
		tfMap["input_artifacts"] = flattenInputArtifacts(v)
	}

	if v := apiObjects.Region; v != nil {
		tfMap[names.AttrRegion] = aws.ToString(v)
	}

	if v := apiObjects.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	if v := apiObjects.TimeoutInMinutes; v != nil {
		tfMap["timeout_in_minutes"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenConditionRules(apiObjects []types.RuleDeclaration) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenConditionRule(apiObject))
	}

	return tfList
}

func flattenCondition(apiObject types.Condition) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Result; v != "" {
		tfMap["result"] = v
	}

	if v := apiObject.Rules; v != nil {
		tfMap[names.AttrRule] = flattenConditionRules(v)
	}

	return tfMap
}

func flattenConditions(apiObjects []types.Condition) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCondition(apiObject))
	}

	return tfList
}

func flattenBeforeEntryDeclaration(apiObject *types.BeforeEntryConditions) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Conditions; v != nil {
		tfMap[names.AttrCondition] = flattenConditions(v)
	}

	return tfMap
}

func flattenOnSuccessDeclaration(apiObject *types.SuccessConditions) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Conditions; v != nil {
		tfMap[names.AttrCondition] = flattenConditions(v)
	}

	return tfMap
}

func flattenRetryConfiguration(apiObject *types.RetryConfiguration) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.RetryMode; v != "" {
		tfMap["retry_mode"] = v
	}

	return tfMap
}

func flattenOnFailureDeclaration(apiObject *types.FailureConditions) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Conditions; v != nil {
		tfMap[names.AttrCondition] = flattenConditions(v)
	}

	if v := apiObject.Result; v != "" {
		tfMap["result"] = v
	}

	if v := apiObject.RetryConfiguration; v != nil {
		tfMap["retry_configuration"] = []any{flattenRetryConfiguration(v)}
	}

	return tfMap
}

func flattenStageDeclaration(d *schema.ResourceData, i int, apiObject types.StageDeclaration) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Actions; v != nil {
		tfMap[names.AttrAction] = flattenActionDeclarations(d, i, v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.BeforeEntry; v != nil {
		tfMap["before_entry"] = []any{flattenBeforeEntryDeclaration(v)}
	}

	if v := apiObject.OnSuccess; v != nil {
		tfMap["on_success"] = []any{flattenOnSuccessDeclaration(v)}
	}

	if v := apiObject.OnFailure; v != nil {
		tfMap["on_failure"] = []any{flattenOnFailureDeclaration(v)}
	}

	return tfMap
}

func flattenStageDeclarations(d *schema.ResourceData, apiObjects []types.StageDeclaration) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for i, apiObject := range apiObjects {
		tfList = append(tfList, flattenStageDeclaration(d, i, apiObject))
	}

	return tfList
}

func flattenActionDeclaration(d *schema.ResourceData, i, j int, apiObject types.ActionDeclaration) map[string]any {
	var actionProvider string
	tfMap := map[string]any{}

	if apiObject := apiObject.ActionTypeId; apiObject != nil {
		tfMap["category"] = apiObject.Category
		tfMap[names.AttrOwner] = apiObject.Owner

		if v := apiObject.Provider; v != nil {
			actionProvider = aws.ToString(v)
			tfMap["provider"] = actionProvider
		}

		if v := apiObject.Version; v != nil {
			tfMap[names.AttrVersion] = aws.ToString(v)
		}
	}

	if v := apiObject.Configuration; v != nil {
		// The AWS API returns "****" for the OAuthToken value. Copy the value from the configuration.
		if actionProvider == providerGitHub {
			if _, ok := v[gitHubActionConfigurationOAuthToken]; ok {
				key := fmt.Sprintf("stage.%d.action.%d.configuration.OAuthToken", i, j)
				v[gitHubActionConfigurationOAuthToken] = d.Get(key).(string)
			}
		}

		tfMap[names.AttrConfiguration] = v
	}

	if v := apiObject.InputArtifacts; len(v) > 0 {
		tfMap["input_artifacts"] = flattenInputArtifacts(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Namespace; v != nil {
		tfMap[names.AttrNamespace] = aws.ToString(v)
	}

	if v := apiObject.OutputArtifacts; len(v) > 0 {
		tfMap["output_artifacts"] = flattenOutputArtifacts(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap[names.AttrRegion] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	if v := apiObject.RunOrder; v != nil {
		tfMap["run_order"] = aws.ToInt32(v)
	}

	if v := apiObject.TimeoutInMinutes; v != nil {
		tfMap["timeout_in_minutes"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenActionDeclarations(d *schema.ResourceData, i int, apiObjects []types.ActionDeclaration) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for j, apiObject := range apiObjects {
		tfList = append(tfList, flattenActionDeclaration(d, i, j, apiObject))
	}

	return tfList
}

func flattenInputArtifacts(apiObjects []types.InputArtifact) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []*string

	for _, apiObject := range apiObjects {
		tfList = append(tfList, apiObject.Name)
	}

	return aws.ToStringSlice(tfList)
}

func flattenOutputArtifacts(apiObjects []types.OutputArtifact) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []*string

	for _, apiObject := range apiObjects {
		tfList = append(tfList, apiObject.Name)
	}

	return aws.ToStringSlice(tfList)
}

func flattenVariableDeclaration(apiObject types.PipelineVariableDeclaration) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.DefaultValue; v != nil {
		tfMap[names.AttrDefaultValue] = aws.ToString(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap[names.AttrDescription] = aws.ToString(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	return tfMap
}

func flattenVariableDeclarations(apiObjects []types.PipelineVariableDeclaration) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenVariableDeclaration(apiObject))
	}

	return tfList
}

func flattenGitBranchFilterCriteria(apiObject *types.GitBranchFilterCriteria) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Excludes; v != nil {
		var tfList []any
		for _, exclude := range apiObject.Excludes {
			tfList = append(tfList, exclude)
		}
		tfMap["excludes"] = tfList
	}

	if v := apiObject.Includes; v != nil {
		var tfList []any
		for _, include := range apiObject.Includes {
			tfList = append(tfList, include)
		}
		tfMap["includes"] = tfList
	}

	return tfMap
}

func flattenGitFilePathFilterCriteria(apiObject *types.GitFilePathFilterCriteria) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Excludes; v != nil {
		var tfList []any
		for _, exclude := range apiObject.Excludes {
			tfList = append(tfList, exclude)
		}
		tfMap["excludes"] = tfList
	}

	if v := apiObject.Includes; v != nil {
		var tfList []any
		for _, include := range apiObject.Includes {
			tfList = append(tfList, include)
		}
		tfMap["includes"] = tfList
	}

	return tfMap
}

func flattenGitPullRequestEventTypes(apiObjects []types.GitPullRequestEventType) []any {
	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, apiObject)
	}

	return tfList
}

func flattenGitTagFilterCriteria(apiObject *types.GitTagFilterCriteria) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Excludes; v != nil {
		var tfList []any
		for _, exclude := range apiObject.Excludes {
			tfList = append(tfList, exclude)
		}
		tfMap["excludes"] = tfList
	}

	if v := apiObject.Includes; v != nil {
		var tfList []any
		for _, include := range apiObject.Includes {
			tfList = append(tfList, include)
		}
		tfMap["includes"] = tfList
	}

	return tfMap
}

func flattenPullRequestFilter(apiObject types.GitPullRequestFilter) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Branches; v != nil {
		tfMap["branches"] = []any{flattenGitBranchFilterCriteria(apiObject.Branches)}
	}

	if v := apiObject.Events; v != nil {
		tfMap["events"] = flattenGitPullRequestEventTypes(apiObject.Events)
	}

	if v := apiObject.FilePaths; v != nil {
		tfMap["file_paths"] = []any{flattenGitFilePathFilterCriteria(apiObject.FilePaths)}
	}

	return tfMap
}

func flattenPullRequestFilters(apiObjects []types.GitPullRequestFilter) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPullRequestFilter(apiObject))
	}

	return tfList
}

func flattenGitPushFilter(apiObject types.GitPushFilter) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Branches; v != nil {
		tfMap["branches"] = []any{flattenGitBranchFilterCriteria(apiObject.Branches)}
	}

	if v := apiObject.FilePaths; v != nil {
		tfMap["file_paths"] = []any{flattenGitFilePathFilterCriteria(apiObject.FilePaths)}
	}

	if v := apiObject.Tags; v != nil {
		tfMap[names.AttrTags] = []any{flattenGitTagFilterCriteria(apiObject.Tags)}
	}

	return tfMap
}

func flattenGitPushFilters(apiObjects []types.GitPushFilter) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenGitPushFilter(apiObject))
	}

	return tfList
}

func flattenGitConfiguration(apiObject *types.GitConfiguration) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.PullRequest; v != nil {
		tfMap["pull_request"] = flattenPullRequestFilters(apiObject.PullRequest)
	}

	if v := apiObject.Push; v != nil {
		tfMap["push"] = flattenGitPushFilters(apiObject.Push)
	}

	tfMap["source_action_name"] = apiObject.SourceActionName

	return tfMap
}

func flattenTriggerDeclaration(apiObject types.PipelineTriggerDeclaration) map[string]any {
	tfMap := map[string]any{}

	tfMap["git_configuration"] = []any{flattenGitConfiguration(apiObject.GitConfiguration)}
	tfMap["provider_type"] = apiObject.ProviderType

	return tfMap
}

func flattenTriggerDeclarations(apiObjects []types.PipelineTriggerDeclaration) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenTriggerDeclaration(apiObject))
	}

	return tfList
}
