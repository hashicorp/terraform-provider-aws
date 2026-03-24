// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	trainingJobNovaModelARNEnvVar = "SAGEMAKER_TRAINING_JOB_NOVA_MODEL_ARN"
	trainingJobCustomImageEnvVar  = "SAGEMAKER_TRAINING_JOB_CUSTOM_IMAGE"
)

func TestNormalizeAlgoSpecMetricDefinitions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	injectedMetrics := testTrainingJobMetricDefinitionsValue(ctx, []*tfsagemaker.TrainingJobMetricDefinitionModel{
		{
			Name:  types.StringValue("aws:loss"),
			Regex: types.StringValue("aws=(.*)"),
		},
	})

	nullMetrics := fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TrainingJobMetricDefinitionModel](ctx)

	testCases := []struct {
		name   string
		config fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobAlgorithmSpecificationModel]
		remote fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobAlgorithmSpecificationModel]
		want   fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobAlgorithmSpecificationModel]
	}{
		{
			name:   "config omitted suppresses injected metric definitions",
			config: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TrainingJobAlgorithmSpecificationModel](ctx),
			remote: testTrainingJobAlgorithmSpecificationValue(ctx, injectedMetrics),
			want:   testTrainingJobAlgorithmSpecificationValue(ctx, nullMetrics),
		},
		{
			name:   "unknown config value is a no op",
			config: fwtypes.NewListNestedObjectValueOfUnknown[tfsagemaker.TrainingJobAlgorithmSpecificationModel](ctx),
			remote: testTrainingJobAlgorithmSpecificationValue(ctx, injectedMetrics),
			want:   testTrainingJobAlgorithmSpecificationValue(ctx, injectedMetrics),
		},
		{
			name:   "empty target is left unchanged",
			config: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TrainingJobAlgorithmSpecificationModel](ctx),
			remote: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TrainingJobAlgorithmSpecificationModel](ctx),
			want:   fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TrainingJobAlgorithmSpecificationModel](ctx),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.remote
			var diags diag.Diagnostics

			tfsagemaker.NormalizeAlgoSpecMetricDefinitions(ctx, testCase.config, &got, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			if !got.Equal(testCase.want) {
				t.Errorf("got = %#v, want = %#v", got, testCase.want)
			}
		})
	}
}

func TestNormalizeStoppingCondition(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	targetStoppingCondition := testTrainingJobStoppingConditionValue(ctx, 86400)
	explicitStoppingCondition := testTrainingJobStoppingConditionValue(ctx, 3600)
	serverlessJobConfig := testTrainingJobServerlessJobConfigValue(ctx)

	testCases := []struct {
		name                string
		config              fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobStoppingConditionModel]
		serverlessJobConfig fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobServerlessJobConfigModel]
		target              fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobStoppingConditionModel]
		want                fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobStoppingConditionModel]
	}{
		{
			name:                "serverless default is suppressed when config omitted stopping condition",
			config:              fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TrainingJobStoppingConditionModel](ctx),
			serverlessJobConfig: serverlessJobConfig,
			target:              targetStoppingCondition,
			want:                fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TrainingJobStoppingConditionModel](ctx),
		},
		{
			name:                "explicit stopping condition is preserved for serverless jobs",
			config:              explicitStoppingCondition,
			serverlessJobConfig: serverlessJobConfig,
			target:              targetStoppingCondition,
			want:                targetStoppingCondition,
		},
		{
			name:                "unknown config value is a no op",
			config:              fwtypes.NewListNestedObjectValueOfUnknown[tfsagemaker.TrainingJobStoppingConditionModel](ctx),
			serverlessJobConfig: serverlessJobConfig,
			target:              targetStoppingCondition,
			want:                targetStoppingCondition,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.target

			tfsagemaker.NormalizeStoppingCondition(ctx, testCase.config, testCase.serverlessJobConfig, &got)

			if !got.Equal(testCase.want) {
				t.Errorf("got = %#v, want = %#v", got, testCase.want)
			}
		})
	}
}

func TestServerlessJobConfigEqualityFunc(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	baseModelARNV1 := "arn:aws:sagemaker:us-west-2:aws:hub-content/example/Model/test/1.0.0" //lintignore:AWSAT003,AWSAT005
	baseModelARNV2 := "arn:aws:sagemaker:us-west-2:aws:hub-content/example/Model/test/2.0.0" //lintignore:AWSAT003,AWSAT005

	testCases := []struct {
		name     string
		oldValue fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobServerlessJobConfigModel]
		newValue fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobServerlessJobConfigModel]
		want     bool
	}{
		{
			name:     "same values are equal",
			oldValue: testTrainingJobServerlessJobConfigValue(ctx),
			newValue: testTrainingJobServerlessJobConfigValue(ctx),
			want:     true,
		},
		{
			name:     "base model arn version differences are semantically equal",
			oldValue: testTrainingJobServerlessJobConfigValueWithOptions(ctx, baseModelARNV1, types.StringNull()),
			newValue: testTrainingJobServerlessJobConfigValueWithOptions(ctx, baseModelARNV2, types.StringNull()),
			want:     true,
		},
		{
			name:     "different non arn fields are not equal",
			oldValue: testTrainingJobServerlessJobConfigValueWithOptions(ctx, baseModelARNV1, types.StringNull()),
			newValue: testTrainingJobServerlessJobConfigValueWithOptions(ctx, baseModelARNV1, types.StringValue("different-evaluator")),
			want:     false,
		},
		{
			name:     "null values are equal",
			oldValue: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TrainingJobServerlessJobConfigModel](ctx),
			newValue: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TrainingJobServerlessJobConfigModel](ctx),
			want:     true,
		},
		{
			name:     "null and populated values are not equal",
			oldValue: fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TrainingJobServerlessJobConfigModel](ctx),
			newValue: testTrainingJobServerlessJobConfigValue(ctx),
			want:     false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, diags := tfsagemaker.ServerlessJobConfigEqualityFunc(ctx, testCase.oldValue, testCase.newValue)

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags)
			}

			if got != testCase.want {
				t.Errorf("got = %v, want = %v", got, testCase.want)
			}
		})
	}
}

func TestVPCConfigFromState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name         string
		vpcConfig    fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobVPCConfigModel]
		wantSGIDs    []string
		wantSubnets  []string
		wantHasError bool
	}{
		{
			name:         "null config returns empty",
			vpcConfig:    fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TrainingJobVPCConfigModel](ctx),
			wantSGIDs:    nil,
			wantSubnets:  nil,
			wantHasError: false,
		},
		{
			name:         "unknown config returns empty",
			vpcConfig:    fwtypes.NewListNestedObjectValueOfUnknown[tfsagemaker.TrainingJobVPCConfigModel](ctx),
			wantSGIDs:    nil,
			wantSubnets:  nil,
			wantHasError: false,
		},
		{
			name:         "empty known config returns empty",
			vpcConfig:    fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.TrainingJobVPCConfigModel{}),
			wantSGIDs:    nil,
			wantSubnets:  nil,
			wantHasError: false,
		},
		{
			name:         "valid config extracts IDs",
			vpcConfig:    testTrainingJobVPCConfigValue(ctx, []string{"sg-12345", "sg-67890"}, []string{"subnet-1", "subnet-2"}),
			wantSGIDs:    []string{"sg-12345", "sg-67890"},
			wantSubnets:  []string{"subnet-1", "subnet-2"},
			wantHasError: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			gotSGIDs, gotSubnets, diags := tfsagemaker.VPCConfigFromState(ctx, testCase.vpcConfig)

			if diags.HasError() != testCase.wantHasError {
				t.Errorf("got = %v, want %v", diags.HasError(), testCase.wantHasError)
			}

			if !testCase.wantHasError {
				if !slices.Equal(gotSGIDs, testCase.wantSGIDs) {
					t.Errorf("got SG IDs = %v, want %v", gotSGIDs, testCase.wantSGIDs)
				}

				if !slices.Equal(gotSubnets, testCase.wantSubnets) {
					t.Errorf("got subnets = %v, want %v", gotSubnets, testCase.wantSubnets)
				}
			}
		})
	}
}

func TestModelPackageGroupARNFromState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name         string
		config       fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobModelPackageConfigModel]
		want         string
		wantHasError bool
	}{
		{
			name:         "null config returns empty",
			config:       fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TrainingJobModelPackageConfigModel](ctx),
			want:         "",
			wantHasError: false,
		},
		{
			name:         "unknown config returns empty",
			config:       fwtypes.NewListNestedObjectValueOfUnknown[tfsagemaker.TrainingJobModelPackageConfigModel](ctx),
			want:         "",
			wantHasError: false,
		},
		{
			name:         "empty known config returns empty",
			config:       fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.TrainingJobModelPackageConfigModel{}),
			want:         "",
			wantHasError: false,
		},
		{
			name:         "valid config extracts group arn",
			config:       testTrainingJobModelPackageConfigValue(ctx, "arn:aws:sagemaker:us-west-2:123456789012:model-package-group/example"), //lintignore:AWSAT003,AWSAT005
			want:         "arn:aws:sagemaker:us-west-2:123456789012:model-package-group/example",                                              //lintignore:AWSAT003,AWSAT005
			wantHasError: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, diags := tfsagemaker.ModelPackageGroupARNFromState(ctx, testCase.config)

			if diags.HasError() != testCase.wantHasError {
				t.Errorf("got = %v, want %v", diags.HasError(), testCase.wantHasError)
			}

			if !testCase.wantHasError && got != testCase.want {
				t.Errorf("got = %q, want %q", got, testCase.want)
			}
		})
	}
}

func testTrainingJobAlgorithmSpecificationValue(ctx context.Context, metricDefinitions fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobMetricDefinitionModel]) fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobAlgorithmSpecificationModel] {
	return fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.TrainingJobAlgorithmSpecificationModel{
		{
			AlgorithmName:                    types.StringNull(),
			ContainerArguments:               fwtypes.NewListValueOfNull[basetypes.StringValue](ctx),
			ContainerEntrypoint:              fwtypes.NewListValueOfNull[basetypes.StringValue](ctx),
			EnableSageMakerMetricsTimeSeries: types.BoolNull(),
			MetricDefinitions:                metricDefinitions,
			TrainingImage:                    types.StringValue("123456789012.dkr.ecr.us-west-2.amazonaws.com/test:latest"), //lintignore:AWSAT003
			TrainingImageConfig:              fwtypes.NewListNestedObjectValueOfNull[tfsagemaker.TrainingJobTrainingImageConfigModel](ctx),
			TrainingInputMode:                types.StringNull(),
		},
	})
}

func testTrainingJobMetricDefinitionsValue(ctx context.Context, metricDefinitions []*tfsagemaker.TrainingJobMetricDefinitionModel) fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobMetricDefinitionModel] {
	return fwtypes.NewListNestedObjectValueOfSliceMust(ctx, metricDefinitions)
}

func testTrainingJobStoppingConditionValue(ctx context.Context, maxRuntimeInSeconds int64) fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobStoppingConditionModel] {
	return fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.TrainingJobStoppingConditionModel{
		{
			MaxPendingTimeInSeconds: types.Int64Null(),
			MaxRuntimeInSeconds:     types.Int64Value(maxRuntimeInSeconds),
			MaxWaitTimeInSeconds:    types.Int64Null(),
		},
	})
}

func testTrainingJobServerlessJobConfigValue(ctx context.Context) fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobServerlessJobConfigModel] {
	return testTrainingJobServerlessJobConfigValueWithOptions(ctx, "arn:aws:sagemaker:us-west-2:aws:hub-content/example/Model/test/1.0.0", types.StringNull()) //lintignore:AWSAT003,AWSAT005
}

func testTrainingJobServerlessJobConfigValueWithOptions(ctx context.Context, baseModelARN string, evaluatorARN types.String) fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobServerlessJobConfigModel] {
	return fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.TrainingJobServerlessJobConfigModel{
		{
			AcceptEULA:             types.BoolNull(),
			BaseModelARN:           types.StringValue(baseModelARN),
			CustomizationTechnique: fwtypes.StringEnumNull[awstypes.CustomizationTechnique](),
			EvaluationType:         fwtypes.StringEnumNull[awstypes.EvaluationType](),
			EvaluatorARN:           evaluatorARN,
			JobType:                fwtypes.StringEnumNull[awstypes.ServerlessJobType](),
			Peft:                   fwtypes.StringEnumNull[awstypes.Peft](),
		},
	})
}

func testTrainingJobVPCConfigValue(ctx context.Context, securityGroupIDs, subnets []string) fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobVPCConfigModel] {
	sgValues := make([]attr.Value, len(securityGroupIDs))
	for i, sg := range securityGroupIDs {
		sgValues[i] = types.StringValue(sg)
	}
	subnetValues := make([]attr.Value, len(subnets))
	for i, subnet := range subnets {
		subnetValues[i] = types.StringValue(subnet)
	}

	return fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.TrainingJobVPCConfigModel{
		{
			SecurityGroupIDs: fwtypes.NewListValueOfMust[basetypes.StringValue](ctx, sgValues),
			Subnets:          fwtypes.NewListValueOfMust[basetypes.StringValue](ctx, subnetValues),
		},
	})
}

func testTrainingJobModelPackageConfigValue(ctx context.Context, modelPackageGroupARN string) fwtypes.ListNestedObjectValueOf[tfsagemaker.TrainingJobModelPackageConfigModel] {
	return fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfsagemaker.TrainingJobModelPackageConfigModel{
		{
			ModelPackageGroupARN:  fwtypes.ARNValue(modelPackageGroupARN),
			SourceModelPackageARN: fwtypes.ARNNull(),
		},
	})
}

func TestAccSageMakerTrainingJob_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:                 testAccSageMakerTrainingJob_basic,
		acctest.CtDisappears:            testAccSageMakerTrainingJob_disappears,
		"vpc":                           testAccSageMakerTrainingJob_vpc,
		"debugConfig":                   testAccSageMakerTrainingJob_debugConfig,
		"profilerConfig":                testAccSageMakerTrainingJob_profilerConfig,
		"environmentAndHyperParameters": testAccSageMakerTrainingJob_environmentAndHyperParameters,
		"checkpointConfig":              testAccSageMakerTrainingJob_checkpointConfig,
		"tensorBoardOutputConfig":       testAccSageMakerTrainingJob_tensorBoardOutputConfig,
		"inputDataConfig":               testAccSageMakerTrainingJob_inputDataConfig,
		"outputDataConfig":              testAccSageMakerTrainingJob_outputDataConfig,
		"algorithmSpecificationMetrics": testAccSageMakerTrainingJob_algorithmSpecificationMetrics,
		"retryStrategy":                 testAccSageMakerTrainingJob_retryStrategy,
		"serverless":                    testAccSageMakerTrainingJob_serverless,
		"tags":                          testAccSageMakerTrainingJob_tags,
		"infraCheckConfig":              testAccSageMakerTrainingJob_infraCheckConfig,
		"mlflowConfig":                  testAccSageMakerTrainingJob_mlflowConfig,
		"remoteDebugConfig":             testAccSageMakerTrainingJob_remoteDebugConfig,
		"sessionChainingConfig":         testAccSageMakerTrainingJob_sessionChainingConfig,
		"Identity":                      testAccSageMakerTrainingJob_identitySerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccSageMakerTrainingJob_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`training-job/.+`)),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.training_input_mode", "File"),
					resource.TestCheckResourceAttrPair(resourceName, "algorithm_specification.0.training_image", "data.aws_sagemaker_prebuilt_ecr_image.test", "registry_path"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.enable_sagemaker_metrics_time_series", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.0.s3_output_path", fmt.Sprintf("s3://%s/output/", rName)),
					resource.TestCheckResourceAttr(resourceName, "resource_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_config.0.instance_type", "ml.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "resource_config.0.instance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_config.0.volume_size_in_gb", "30"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "3600"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func testAccSageMakerTrainingJob_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsagemaker.ResourceTrainingJob, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSageMakerTrainingJob_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_vpc(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "delete_vpc_enis_on_destroy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "3600"),
				),
			},
			{
				Config: testAccTrainingJobConfig_vpcUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "delete_vpc_enis_on_destroy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "7200"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
				ImportStateVerifyIgnore:              []string{"delete_vpc_enis_on_destroy"},
			},
		},
	})
}

func testAccSageMakerTrainingJob_debugConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_debug(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "debug_hook_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "debug_hook_config.0.s3_output_path", fmt.Sprintf("s3://%s/debug/", rName)),
					resource.TestCheckResourceAttr(resourceName, "debug_rule_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "debug_rule_configurations.0.s3_output_path", fmt.Sprintf("s3://%s/debug-rules/", rName)),
				),
			},
			{
				Config: testAccTrainingJobConfig_debugUpdate(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "debug_hook_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "debug_hook_config.0.s3_output_path", fmt.Sprintf("s3://%s/debug-updated/", rNameUpdated)),
					resource.TestCheckResourceAttr(resourceName, "debug_rule_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "debug_rule_configurations.0.s3_output_path", fmt.Sprintf("s3://%s/debug-rules-updated/", rNameUpdated)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func testAccSageMakerTrainingJob_profilerConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_profiler(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.0.disable_profiler", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.0.profiling_interval_in_milliseconds", "500"),
					resource.TestCheckResourceAttr(resourceName, "profiler_rule_configurations.#", "1"),
				),
			},
			{
				Config: testAccTrainingJobConfig_profilerUpdated(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.0.disable_profiler", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "profiler_config.0.profiling_interval_in_milliseconds", "1000"),
					resource.TestCheckResourceAttr(resourceName, "profiler_rule_configurations.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func testAccSageMakerTrainingJob_environmentAndHyperParameters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_environmentAndHyperParameters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "environment.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment.TEST_ENV", "test_value"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameters.epochs", "10"),
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "3600"),
				),
			},
			{
				Config: testAccTrainingJobConfig_environmentAndHyperParametersUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "environment.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment.TEST_ENV", "updated_value"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "hyper_parameters.epochs", "20"),
					resource.TestCheckResourceAttr(resourceName, "enable_inter_container_traffic_encryption", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "stopping_condition.0.max_runtime_in_seconds", "7200"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func testAccSageMakerTrainingJob_checkpointConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_checkpoint(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "checkpoint_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "checkpoint_config.0.local_path", "/opt/ml/checkpoints"),
					resource.TestCheckResourceAttr(resourceName, "checkpoint_config.0.s3_uri", fmt.Sprintf("s3://%s/checkpoints/", rName)),
				),
			},
			{
				Config: testAccTrainingJobConfig_checkpointUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "checkpoint_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "checkpoint_config.0.local_path", "/opt/ml/checkpoints"),
					resource.TestCheckResourceAttr(resourceName, "checkpoint_config.0.s3_uri", fmt.Sprintf("s3://%s/checkpoints-v2/", rName)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func testAccSageMakerTrainingJob_tensorBoardOutputConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_tensorBoard(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tensor_board_output_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tensor_board_output_config.0.local_path", "/opt/ml/output/tensorboard"),
					resource.TestCheckResourceAttr(resourceName, "tensor_board_output_config.0.s3_output_path", fmt.Sprintf("s3://%s/tensorboard/", rName)),
				),
			},
			{
				Config: testAccTrainingJobConfig_tensorBoardUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tensor_board_output_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tensor_board_output_config.0.local_path", "/opt/ml/output/tensorboard"),
					resource.TestCheckResourceAttr(resourceName, "tensor_board_output_config.0.s3_output_path", fmt.Sprintf("s3://%s/tensorboard-v2/", rName)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func testAccSageMakerTrainingJob_inputDataConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_inputData(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.channel_name", "training"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.input_mode", "File"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.data_source.0.s3_data_source.0.s3_uri", fmt.Sprintf("s3://%s/input/", rName)),
				),
			},
			{
				Config: testAccTrainingJobConfig_inputDataUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.channel_name", "training"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.input_mode", "File"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.data_source.0.s3_data_source.0.s3_uri", fmt.Sprintf("s3://%s/input-v2/", rName)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func testAccSageMakerTrainingJob_outputDataConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_outputData(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.0.compression_type", "GZIP"),
					resource.TestCheckResourceAttrSet(resourceName, "output_data_config.0.kms_key_id"),
				),
			},
			{
				Config: testAccTrainingJobConfig_outputDataUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_data_config.0.compression_type", "NONE"),
					resource.TestCheckResourceAttrSet(resourceName, "output_data_config.0.kms_key_id"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func testAccSageMakerTrainingJob_algorithmSpecificationMetrics(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"
	customImage := acctest.SkipIfEnvVarNotSet(t, trainingJobCustomImageEnvVar)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_algorithmMetrics(rName, customImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.0.name", "train:loss"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.0.regex", "loss: ([0-9\\.]+)"),
				),
			},
			{
				Config: testAccTrainingJobConfig_algorithmMetricsUpdate(rNameUpdated, customImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.0.name", "train:loss"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.0.regex", "loss: ([0-9\\.]+)"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.1.name", "validation:accuracy"),
					resource.TestCheckResourceAttr(resourceName, "algorithm_specification.0.metric_definitions.1.regex", "accuracy: ([0-9\\.]+)"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
				ImportStateVerifyIgnore:              []string{"algorithm_specification.0.metric_definitions"},
			},
		},
	})
}

func testAccSageMakerTrainingJob_retryStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_retryStrategy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.0.maximum_retry_attempts", "3"),
				),
			},
			{
				Config: testAccTrainingJobConfig_retryStrategyUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "retry_strategy.0.maximum_retry_attempts", "5"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func testAccSageMakerTrainingJob_serverless(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_serverless(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "delete_model_packages_on_destroy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.0.job_type", "FineTuning"),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.0.accept_eula", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.0.customization_technique", "SFT"),
					resource.TestCheckResourceAttrSet(resourceName, "serverless_job_config.0.base_model_arn"),
					resource.TestCheckResourceAttr(resourceName, "model_package_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "model_package_config.0.model_package_group_arn"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.channel_name", "train"),
				),
			},
			{
				Config: testAccTrainingJobConfig_serverlessUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "delete_model_packages_on_destroy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.0.job_type", "FineTuning"),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.0.accept_eula", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "serverless_job_config.0.customization_technique", "DPO"),
					resource.TestCheckResourceAttrSet(resourceName, "serverless_job_config.0.base_model_arn"),
					resource.TestCheckResourceAttr(resourceName, "model_package_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "model_package_config.0.model_package_group_arn"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_data_config.0.channel_name", "train"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
				ImportStateVerifyIgnore:              []string{"delete_model_packages_on_destroy", "serverless_job_config.0.base_model_arn"},
			},
		},
	})
}

func testAccSageMakerTrainingJob_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
				ImportStateVerifyIgnore:              []string{"algorithm_specification.0.metric_definitions"},
			},
			{
				Config: testAccTrainingJobConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccTrainingJobConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccSageMakerTrainingJob_infraCheckConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_infraCheck(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "infra_check_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "infra_check_config.0.enable_infra_check", acctest.CtTrue),
				),
			},
			{
				Config: testAccTrainingJobConfig_infraCheckUpdate(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "infra_check_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "infra_check_config.0.enable_infra_check", acctest.CtFalse),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func testAccSageMakerTrainingJob_mlflowConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"
	novaModelARN := acctest.SkipIfEnvVarNotSet(t, trainingJobNovaModelARNEnvVar)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_mlflow(rName, novaModelARN),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "delete_model_packages_on_destroy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mlflow_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mlflow_config.0.mlflow_experiment_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "mlflow_config.0.mlflow_resource_arn"),
					resource.TestCheckResourceAttr(resourceName, "mlflow_config.0.mlflow_run_name", rName),
				),
			},
			{
				Config: testAccTrainingJobConfig_mlflowUpdate(rNameUpdated, novaModelARN),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "delete_model_packages_on_destroy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "mlflow_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mlflow_config.0.mlflow_experiment_name", rNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "mlflow_config.0.mlflow_resource_arn"),
					resource.TestCheckResourceAttr(resourceName, "mlflow_config.0.mlflow_run_name", rNameUpdated),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
				ImportStateVerifyIgnore:              []string{"delete_model_packages_on_destroy"},
			},
		},
	})
}

func testAccSageMakerTrainingJob_remoteDebugConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"
	customImage := acctest.SkipIfEnvVarNotSet(t, trainingJobCustomImageEnvVar)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_remoteDebug(rName, rName, customImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "remote_debug_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "remote_debug_config.0.enable_remote_debug", acctest.CtFalse),
				),
			},
			{
				Config: testAccTrainingJobConfig_remoteDebugUpdate(rName, rNameUpdated, customImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "remote_debug_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "remote_debug_config.0.enable_remote_debug", acctest.CtTrue),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
			},
		},
	})
}

func testAccSageMakerTrainingJob_sessionChainingConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var trainingjob sagemaker.DescribeTrainingJobOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_training_job.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTrainingJobs(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrainingJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTrainingJobConfig_sessionChaining(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "session_chaining_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "session_chaining_config.0.enable_session_tag_chaining", acctest.CtTrue),
				),
			},
			{
				Config: testAccTrainingJobConfig_sessionChainingUpdate(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrainingJobExists(ctx, t, resourceName, &trainingjob),
					resource.TestCheckResourceAttr(resourceName, "training_job_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "session_chaining_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "session_chaining_config.0.enable_session_tag_chaining", acctest.CtFalse),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "training_job_name"),
				ImportStateVerifyIdentifierAttribute: "training_job_name",
				ImportStateVerifyIgnore:              []string{"session_chaining_config"},
			},
		},
	})
}

func testAccCheckTrainingJobDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_training_job" {
				continue
			}

			trainingJobName := rs.Primary.Attributes["training_job_name"]
			if trainingJobName == "" {
				return fmt.Errorf("No SageMaker Training Job name is set")
			}

			_, err := tfsagemaker.FindTrainingJobByName(ctx, conn, trainingJobName)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker Training Job %s still exists", trainingJobName)
		}

		return nil
	}
}

func testAccCheckTrainingJobExists(ctx context.Context, t *testing.T, name string, trainingjob *sagemaker.DescribeTrainingJobOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		trainingJobName := rs.Primary.Attributes["training_job_name"]
		if trainingJobName == "" {
			return fmt.Errorf("No SageMaker Training Job name is set")
		}

		output, err := tfsagemaker.FindTrainingJobByName(ctx, conn, trainingJobName)
		if err != nil {
			return err
		}

		*trainingjob = *output

		return nil
	}
}

func testAccPreCheckTrainingJobs(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

	input := &sagemaker.ListTrainingJobsInput{}

	_, err := conn.ListTrainingJobs(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTrainingJobConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity", "sts:TagSession"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "linear-learner"
  image_tag       = "1"
}
`, rName)
}

func testAccTrainingJobConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode                  = "File"
    training_image                       = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    enable_sagemaker_metrics_time_series = true
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_vpc(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name          = %[1]q
  role_arn                   = aws_iam_role.test.arn
  delete_vpc_enis_on_destroy = true

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test[0].id]
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_vpcUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name          = %[1]q
  role_arn                   = aws_iam_role.test.arn
  delete_vpc_enis_on_destroy = true

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 7200
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test[0].id, aws_subnet.test[1].id]
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_debug(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
  statement {
    actions = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.test.arn
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.s3.json
}

data "aws_sagemaker_prebuilt_ecr_image" "debugger" {
  repository_name = "sagemaker-debugger-rules"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  debug_hook_config {
    local_path     = "/opt/ml/output/tensors"
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/debug/"
  }

  debug_rule_configurations {
    local_path              = "/opt/ml/processing/test1"
    rule_configuration_name = "LossNotDecreasing"
    rule_evaluator_image    = data.aws_sagemaker_prebuilt_ecr_image.debugger.registry_path
    rule_parameters = {
      "rule_to_invoke" = "LossNotDecreasing"
    }
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/debug-rules/"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccTrainingJobConfig_debugUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
  statement {
    actions = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.test.arn
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.s3.json
}

data "aws_sagemaker_prebuilt_ecr_image" "debugger" {
  repository_name = "sagemaker-debugger-rules"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  debug_hook_config {
    local_path     = "/opt/ml/output/tensors"
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/debug-updated/"
  }

  debug_rule_configurations {
    local_path              = "/opt/ml/processing/test1"
    rule_configuration_name = "LossNotDecreasing"
    rule_evaluator_image    = data.aws_sagemaker_prebuilt_ecr_image.debugger.registry_path
    rule_parameters = {
      "rule_to_invoke" = "LossNotDecreasing"
    }
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/debug-rules-updated/"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccTrainingJobConfig_profiler(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
  statement {
    actions = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.test.arn
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.s3.json
}

data "aws_sagemaker_prebuilt_ecr_image" "debugger" {
  repository_name = "sagemaker-debugger-rules"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  profiler_config {
    disable_profiler                   = false
    profiling_interval_in_milliseconds = 500
    profiling_parameters = {
      "profile_cpu" = "true"
    }
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/profiler/"
  }

  profiler_rule_configurations {
    local_path              = "/opt/ml/processing/test"
    rule_configuration_name = "ProfilerReport"
    rule_evaluator_image    = data.aws_sagemaker_prebuilt_ecr_image.debugger.registry_path
    rule_parameters = {
      "rule_to_invoke" = "ProfilerReport"
    }
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/profiler-rules/"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccTrainingJobConfig_profilerUpdated(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
  statement {
    actions = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.test.arn
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.s3.json
}

data "aws_sagemaker_prebuilt_ecr_image" "debugger" {
  repository_name = "sagemaker-debugger-rules"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  profiler_config {
    disable_profiler                   = false
    profiling_interval_in_milliseconds = 1000
    profiling_parameters = {
      "profile_cpu" = "false"
    }
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/profiler/"
  }

  profiler_rule_configurations {
    local_path              = "/opt/ml/processing/test"
    rule_configuration_name = "ProfilerReport"
    rule_evaluator_image    = data.aws_sagemaker_prebuilt_ecr_image.debugger.registry_path
    rule_parameters = {
      "rule_to_invoke" = "ProfilerReport"
    }
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/profiler-rules/"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test]
}
`, rName))
}

func testAccTrainingJobConfig_environmentAndHyperParameters(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "pytorch-training"
  image_tag       = "2.0.0-cpu-py310-ubuntu20.04-sagemaker"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  enable_inter_container_traffic_encryption = true
  enable_managed_spot_training              = true
  enable_network_isolation                  = false

  environment = {
    "TEST_ENV"    = "test_value"
    "ANOTHER_ENV" = "another_value"
  }

  hyper_parameters = {
    "epochs"     = "10"
    "batch_size" = "32"
  }

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds   = 3600
    max_wait_time_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_environmentAndHyperParametersUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "pytorch-training"
  image_tag       = "2.0.0-cpu-py310-ubuntu20.04-sagemaker"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  enable_inter_container_traffic_encryption = false
  enable_managed_spot_training              = true
  enable_network_isolation                  = false

  environment = {
    "TEST_ENV"    = "updated_value"
    "ANOTHER_ENV" = "another_value"
  }

  hyper_parameters = {
    "epochs"     = "20"
    "batch_size" = "32"
  }

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds   = 7200
    max_wait_time_in_seconds = 8000
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccTrainingJobConfig_checkpoint(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  checkpoint_config {
    local_path = "/opt/ml/checkpoints"
    s3_uri     = "s3://${aws_s3_bucket.test.bucket}/checkpoints/"
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_checkpointUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  checkpoint_config {
    local_path = "/opt/ml/checkpoints"
    s3_uri     = "s3://${aws_s3_bucket.test.bucket}/checkpoints-v2/"
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_tensorBoard(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  tensor_board_output_config {
    local_path     = "/opt/ml/output/tensorboard"
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/tensorboard/"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_tensorBoardUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  tensor_board_output_config {
    local_path     = "/opt/ml/output/tensorboard"
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/tensorboard-v2/"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_inputData(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
  statement {
    actions = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.test.arn
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.s3.json
}

resource "aws_s3_object" "input" {
  bucket  = aws_s3_bucket.test.id
  key     = "input/placeholder.csv"
  content = "feature1,label\n1.0,0\n"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  input_data_config {
    channel_name        = "training"
    compression_type    = "None"
    content_type        = "text/csv"
    input_mode          = "File"
    record_wrapper_type = "None"

    data_source {
      s3_data_source {
        s3_data_distribution_type = "FullyReplicated"
        s3_data_type              = "S3Prefix"
        s3_uri                    = "s3://${aws_s3_bucket.test.bucket}/input/"
      }
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test, aws_s3_object.input]
}
`, rName))
}

func testAccTrainingJobConfig_inputDataUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_iam_policy_document" "s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject"
    ]
    resources = [
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
  statement {
    actions = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.test.arn
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.s3.json
}

resource "aws_s3_object" "input_v2" {
  bucket  = aws_s3_bucket.test.id
  key     = "input-v2/placeholder.csv"
  content = "feature1,label\n1.0,0\n"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  input_data_config {
    channel_name        = "training"
    compression_type    = "None"
    content_type        = "text/csv"
    input_mode          = "File"
    record_wrapper_type = "None"

    data_source {
      s3_data_source {
        s3_data_distribution_type = "FullyReplicated"
        s3_data_type              = "S3Prefix"
        s3_uri                    = "s3://${aws_s3_bucket.test.bucket}/input-v2/"
      }
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.test, aws_s3_object.input_v2]
}
`, rName))
}

func testAccTrainingJobConfig_outputData(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "KMS key for SageMaker training job"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    compression_type = "GZIP"
    kms_key_id       = aws_kms_key.test.arn
    s3_output_path   = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_outputDataUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "KMS key for SageMaker training job"
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    compression_type = "NONE"
    kms_key_id       = aws_kms_key.test.arn
    s3_output_path   = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_algorithmMetrics(rName, customImage string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_iam_role_policy" "ecr" {
  name = "%[1]s-ecr"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer",
          "ecr:GetAuthorizationToken",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = %[2]q

    metric_definitions {
      name  = "train:loss"
      regex = "loss: ([0-9\\.]+)"
    }

  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.ecr]
}
`, rName, customImage))
}

func testAccTrainingJobConfig_algorithmMetricsUpdate(rName, customImage string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_iam_role_policy" "ecr" {
  name = "%[1]s-ecr"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer",
          "ecr:GetAuthorizationToken",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = %[2]q

    metric_definitions {
      name  = "train:loss"
      regex = "loss: ([0-9\\.]+)"
    }

    metric_definitions {
      name  = "validation:accuracy"
      regex = "accuracy: ([0-9\\.]+)"
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.ecr]
}
`, rName, customImage))
}

func testAccTrainingJobConfig_retryStrategy(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  retry_strategy {
    maximum_retry_attempts = 3
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_retryStrategyUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  retry_strategy {
    maximum_retry_attempts = 5
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_serverless(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_iam_role_policy" "hub_access" {
  name = "%[1]s-hub"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["sagemaker:DescribeHubContent"]
      Resource = ["*"]
    }]
  })
}

resource "aws_iam_role_policy" "s3" {
  name = "%[1]s-s3"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = ["s3:GetObject", "s3:PutObject", "s3:ListBucket", "s3:DeleteObject"]
      Resource = [
        "arn:${data.aws_partition.current.partition}:s3:::%[1]s",
        "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"
      ]
    }]
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "training" {
  bucket  = aws_s3_bucket.test.id
  key     = "train/placeholder.jsonl"
  content = "{\"prompt\": \"hello\", \"completion\": \"world\"}\n"
}

resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name                = %[1]q
  role_arn                         = aws_iam_role.test.arn
  delete_model_packages_on_destroy = true

  input_data_config {
    channel_name = "train"
    content_type = "application/jsonlines"
    input_mode   = "File"

    data_source {
      s3_data_source {
        s3_data_distribution_type = "FullyReplicated"
        s3_data_type              = "S3Prefix"
        s3_uri                    = "s3://${aws_s3_bucket.test.bucket}/train/"
      }
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  model_package_config {
    model_package_group_arn = aws_sagemaker_model_package_group.test.arn
  }

  serverless_job_config {
    accept_eula             = true
    base_model_arn          = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.name}:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct/2.40.0"
    job_type                = "FineTuning"
    customization_technique = "SFT"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.hub_access, aws_iam_role_policy.s3, aws_s3_object.training]
}
`, rName)
}

func testAccTrainingJobConfig_serverlessUpdate(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_iam_role_policy" "hub_access" {
  name = "%[1]s-hub"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["sagemaker:DescribeHubContent"]
      Resource = ["*"]
    }]
  })
}

resource "aws_iam_role_policy" "s3" {
  name = "%[1]s-s3"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = ["s3:GetObject", "s3:PutObject", "s3:ListBucket", "s3:DeleteObject"]
      Resource = [
        "arn:${data.aws_partition.current.partition}:s3:::%[1]s",
        "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"
      ]
    }]
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "training" {
  bucket  = aws_s3_bucket.test.id
  key     = "train/placeholder.jsonl"
  content = "{\"prompt\": \"hello\", \"completion\": \"world\"}\n"
}

resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name                = %[1]q
  role_arn                         = aws_iam_role.test.arn
  delete_model_packages_on_destroy = true

  input_data_config {
    channel_name = "train"
    content_type = "application/jsonlines"
    input_mode   = "File"

    data_source {
      s3_data_source {
        s3_data_distribution_type = "FullyReplicated"
        s3_data_type              = "S3Prefix"
        s3_uri                    = "s3://${aws_s3_bucket.test.bucket}/train/"
      }
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  model_package_config {
    model_package_group_arn = aws_sagemaker_model_package_group.test.arn
  }

  serverless_job_config {
    accept_eula             = true
    base_model_arn          = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.name}:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct/2.40.0"
    job_type                = "FineTuning"
    customization_technique = "DPO"
    peft                    = "LORA"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.hub_access, aws_iam_role_policy.s3, aws_s3_object.training]
}
`, rName)
}

func testAccTrainingJobConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1))
}

func testAccTrainingJobConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccTrainingJobConfig_infraCheck(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  infra_check_config {
    enable_infra_check = true
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_infraCheckUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  infra_check_config {
    enable_infra_check = false
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_mlflow(rName, novaModelARN string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_iam_role_policy" "s3" {
  name = "%[1]s-s3"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = ["s3:GetObject", "s3:PutObject", "s3:ListBucket", "s3:DeleteObject"]
      Resource = [
        "arn:${data.aws_partition.current.partition}:s3:::%[1]s",
        "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"
      ]
    }]
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "training" {
  bucket  = aws_s3_bucket.test.id
  key     = "train/placeholder.jsonl"
  content = "{\"prompt\": \"hello\", \"completion\": \"world\"}\n"
}

resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_mlflow_tracking_server" "test" {
  tracking_server_name = %[1]q
  artifact_store_uri   = "s3://${aws_s3_bucket.test.bucket}/mlflow/"
  role_arn             = aws_iam_role.test.arn
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name                = %[1]q
  role_arn                         = aws_iam_role.test.arn
  delete_model_packages_on_destroy = true

  input_data_config {
    channel_name = "train"
    content_type = "application/jsonlines"
    input_mode   = "File"

    data_source {
      s3_data_source {
        s3_data_distribution_type = "FullyReplicated"
        s3_data_type              = "S3Prefix"
        s3_uri                    = "s3://${aws_s3_bucket.test.bucket}/train/"
      }
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  serverless_job_config {
    accept_eula             = true
    base_model_arn          = %[2]q
    job_type                = "FineTuning"
    customization_technique = "SFT"
  }

  model_package_config {
    model_package_group_arn = aws_sagemaker_model_package_group.test.arn
  }

  mlflow_config {
    mlflow_experiment_name = %[1]q
    mlflow_resource_arn    = aws_sagemaker_mlflow_tracking_server.test.arn
    mlflow_run_name        = %[1]q
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.s3, aws_s3_object.training]
}
`, rName, novaModelARN)
}

func testAccTrainingJobConfig_mlflowUpdate(rName, novaModelARN string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole", "sts:SetSourceIdentity"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_iam_role_policy" "s3" {
  name = "%[1]s-s3"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = ["s3:GetObject", "s3:PutObject", "s3:ListBucket", "s3:DeleteObject"]
      Resource = [
        "arn:${data.aws_partition.current.partition}:s3:::%[1]s",
        "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"
      ]
    }]
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "training" {
  bucket  = aws_s3_bucket.test.id
  key     = "train/placeholder.jsonl"
  content = "{\"prompt\": \"hello\", \"completion\": \"world\"}\n"
}

resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_mlflow_tracking_server" "test" {
  tracking_server_name = %[1]q
  artifact_store_uri   = "s3://${aws_s3_bucket.test.bucket}/mlflow/"
  role_arn             = aws_iam_role.test.arn
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name                = %[1]q
  role_arn                         = aws_iam_role.test.arn
  delete_model_packages_on_destroy = true

  input_data_config {
    channel_name = "train"
    content_type = "application/jsonlines"
    input_mode   = "File"

    data_source {
      s3_data_source {
        s3_data_distribution_type = "FullyReplicated"
        s3_data_type              = "S3Prefix"
        s3_uri                    = "s3://${aws_s3_bucket.test.bucket}/train/"
      }
    }
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  serverless_job_config {
    accept_eula             = true
    base_model_arn          = %[2]q
    job_type                = "FineTuning"
    customization_technique = "SFT"
  }

  model_package_config {
    model_package_group_arn = aws_sagemaker_model_package_group.test.arn
  }

  mlflow_config {
    mlflow_experiment_name = %[1]q
    mlflow_resource_arn    = aws_sagemaker_mlflow_tracking_server.test.arn
    mlflow_run_name        = %[1]q
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.s3, aws_s3_object.training]
}
`, rName, novaModelARN)
}

func testAccTrainingJobConfig_remoteDebug(rName, jobName, customImage string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_iam_role_policy" "ecr" {
  name = "%[1]s-ecr"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer",
          "ecr:GetAuthorizationToken",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[2]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = %[3]q
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  remote_debug_config {
    enable_remote_debug = false
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.ecr]
}
`, rName, jobName, customImage))
}

func testAccTrainingJobConfig_remoteDebugUpdate(rName, jobName, customImage string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_iam_role_policy" "ecr" {
  name = "%[1]s-ecr"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer",
          "ecr:GetAuthorizationToken",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[2]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = %[3]q
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  remote_debug_config {
    enable_remote_debug = true
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.ecr]
}
`, rName, jobName, customImage))
}

func testAccTrainingJobConfig_sessionChaining(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  session_chaining_config {
    enable_session_tag_chaining = true
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccTrainingJobConfig_sessionChainingUpdate(rName string) string {
	return acctest.ConfigCompose(testAccTrainingJobConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_training_job" "test" {
  training_job_name = %[1]q
  role_arn          = aws_iam_role.test.arn

  algorithm_specification {
    training_input_mode = "File"
    training_image      = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }

  output_data_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/output/"
  }

  resource_config {
    instance_type     = "ml.m5.large"
    instance_count    = 1
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  session_chaining_config {
    enable_session_tag_chaining = false
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}
