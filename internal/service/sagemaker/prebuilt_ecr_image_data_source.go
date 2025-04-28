// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// SageMaker Algorithm AutoGluon Training
	repositoryAutoGluonTraining = "autogluon-training"
	// SageMaker Algorithm AutoGluon Inference
	repositoryAutoGluonInference = "autogluon-inference"
	// SageMaker Algorithm BlazingText
	repositoryBlazingText = "blazingtext"
	// SageMaker DLC Chainer
	repositoryChainer = "sagemaker-chainer"
	// SageMaker Algorithm Clarify
	repositoryClarify = "sagemaker-clarify-processing"
	// SageMaker Algorithm DJL DeepSpeed
	repositoryDJLDeepSpeed = "djl-inference"
	// SageMaker Algorithm Data Wrangler
	repositoryDataWrangler = "sagemaker-data-wrangler-container"
	// SageMaker Algorithm Debugger
	repositoryDebugger = "sagemaker-debugger-rules"
	// SageMaker Algorithm DeepAR Forecasting
	repositoryDeepARForecasting = "forecasting-deepar"
	// SageMaker Algorithm Factorization Machines
	repositoryFactorizationMachines = "factorization-machines"
	// SageMaker Algorithm HuggingFace TensorFlow Training
	repositoryHuggingFaceTensorFlowTraining = "huggingface-tensorflow-training"
	// SageMaker Algorithm HuggingFace PyTorch Training
	repositoryHuggingFacePyTorchTraining = "huggingface-pytorch-training"
	// SageMaker Algorithm HuggingFace PyTorch Training NeuronX
	repositoryHuggingFacePyTorchTrainingNeuronX = "huggingface-pytorch-training-neuronx"
	// SageMaker Algorithm HuggingFace PyTorch Training Compiler
	repositoryHuggingFacePyTorchTrainingCompiler = "huggingface-pytorch-trcomp-training"
	// SageMaker Algorithm HuggingFace TensorFlow Training Compiler
	repositoryHuggingFaceTensorFlowTrainingCompiler = "huggingface-tensorflow-trcomp-training"
	// SageMaker Algorithm HuggingFace TensorFlow Inference
	repositoryHuggingFaceTensorFlowInference = "huggingface-tensorflow-inference"
	// SageMaker Algorithm HuggingFace PyTorch Inference
	repositoryHuggingFacePyTorchInference = "huggingface-pytorch-inference"
	// SageMaker Algorithm HuggingFace PyTorch Inference Neuron
	repositoryHuggingFacePyTorchInferenceNeuron = "huggingface-pytorch-inference-neuron"
	// SageMaker Algorithm HuggingFace PyTorch Inference NeuronX
	repositoryHuggingFacePyTorchInferenceNeuronX = "huggingface-pytorch-inference-neuronx"
	// SageMaker LLM HuggingFace Pytorch TGI Inference
	repositoryHuggingFacePyTorchTGIInference = "huggingface-pytorch-tgi-inference"
	// SageMaker Algorithm HuggingFace TEI
	repositoryHuggingFaceTEI = "tei"
	// SageMaker Algorithm HuggingFace TEI CPU
	repositoryHuggingFaceTEICPU = "tei-cpu"
	// SageMaker Algorithm IP Insights
	repositoryIPInsights = "ipinsights"
	// SageMaker Algorithm Image Classification
	repositoryImageClassification = "image-classification"
	// SageMaker DLC Inferentia MXNet
	repositoryInferentiaMXNet = "sagemaker-neo-mxnet"
	// SageMaker DLC Inferentia PyTorch
	repositoryInferentiaPyTorch = "sagemaker-neo-pytorch"
	// SageMaker Algorithm k-means
	repositoryKMeans = "kmeans"
	// SageMaker Algorithm k-nearest-neighbor
	repositoryKNearestNeighbor = "knn"
	// SageMaker Algorithm Latent Dirichlet Allocation
	repositoryLDA = "lda"
	// SageMaker Algorithm Linear Learner
	repositoryLinearLearner = "linear-learner"
	// SageMaker DLC MXNet Training
	repositoryMXNetTraining = "mxnet-training"
	// SageMaker DLC MXNet Inference
	repositoryMXNetInference = "mxnet-inference"
	// SageMaker DLC MXNet Inference EIA
	repositoryMXNetInferenceEIA = "mxnet-inference-eia"
	// SageMaker DLC SageMaker MXNet
	repositorySageMakerMXNet = "sagemaker-mxnet" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker DLC SageMaker MXNet Serving
	repositorySageMakerMXNetServing = "sagemaker-mxnet-serving" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker DLC SageMaker MXNet EIA
	repositorySageMakerMXNetEIA = "sagemaker-mxnet-eia" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker DLC SageMaker MXNet Serving EIA
	repositorySageMakerMXNetServingEIA = "sagemaker-mxnet-serving-eia" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker DLC MXNet Coach
	repositoryMXNetCoach = "sagemaker-rl-mxnet"
	// SageMaker Model Monitor
	repositoryModelMonitor = "sagemaker-model-monitor-analyzer"
	// SageMaker Algorithm Neural Topic Model
	repositoryNeuralTopicModel = "ntm"
	// SageMaker Algorithm Neo Image Classification
	repositoryNeoImageClassification = "image-classification-neo"
	// SageMaker DLC Neo MXNet
	repositoryNeoMXNet = "sagemaker-inference-mxnet"
	// SageMaker DLC Neo PyTorch
	repositoryNeoPyTorch = "sagemaker-inference-pytorch"
	// SageMaker DLC Neo Tensorflow
	repositoryNeoTensorflow = "sagemaker-inference-tensorflow"
	// SageMaker DLC Neo XGBoost
	repositoryNeoXGBoost = "xgboost-neo"
	// SageMaker Algorithm Object Detection
	repositoryObjectDetection = "object-detection"
	// SageMaker Algorithm Object2Vec
	repositoryObject2Vec = "object2vec"
	// SageMaker Algorithm PCA
	repositoryPCA = "pca"
	// SageMaker DLC PyTorch Training
	repositoryPyTorchTraining = "pytorch-training"
	// SageMaker DLC PyTorch Training NeuronX
	repositoryPyTorchTrainingNeuronX = "pytorch-training-neuronx"
	// SageMaker DLC PyTorch Training Compiler
	repositoryPyTorchTrainingCompiler = "pytorch-trcomp-training"
	// SageMaker DLC SageMaker PyTorch Inference
	repositoryPyTorchInference = "pytorch-inference"
	// SageMaker DLC PyTorch Inference EIA
	repositoryPyTorchInferenceEIA = "pytorch-inference-eia"
	// SageMaker DLC PyTorch Inference Graviton
	repositoryPyTorchInferenceGraviton = "pytorch-inference-graviton"
	// SageMaker DLC PyTorch Inference NeuronX
	repositoryPyTorchInferenceNeuronX = "pytorch-inference-neuronx"
	// SageMaker DLC SageMaker PyTorch
	repositorySageMakerPyTorch = "sagemaker-pytorch" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker Algorithm Random Cut Forest
	repositoryRandomCutForest = "randomcutforest"
	// SageMaker DLC RL Ray PyTorch
	repositoryRLRayPyTorch = "sagemaker-rl-ray-container"
	// SageMaker DLC RL Coach Tensorflow
	repositoryRLCoachPyTorch = "sagemaker-rl-coach-container"
	// SageMaker Library scikit-learn
	repositoryScikitLearn = "sagemaker-scikit-learn"
	// SageMaker Algorithm Semantic Segmentation
	repositorySemanticSegmentation = "semantic-segmentation"
	// SageMaker Algorithm Seq2Seq
	repositorySeq2Seq = "seq2seq"
	// SageMaker Algorithm Spark
	repositorySpark = "sagemaker-spark-processing"
	// SageMaker Algorithm Spark ML
	repositorySparkML = "sagemaker-sparkml-serving"
	// SageMaker DLC TensorFlow Training
	repositoryTensorFlowTraining = "tensorflow-training"
	// SageMaker DLC TensorFlow Inference
	repositoryTensorFlowInference = "tensorflow-inference"
	// SageMaker Repo TensorFlow Inference EIA
	repositoryTensorFlowInferenceEIA = "tensorflow-inference-eia"
	// SageMaker DLC TensorFlow Inference Graviton
	repositoryTensorFlowInferenceGraviton = "tensorflow-inference-graviton"
	// SageMaker DLC SageMaker TensorFlow
	repositorySageMakerTensorFlow = "sagemaker-tensorflow" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker DLC SageMaker TensorFlow EIA
	repositorySageMakerTensorFlowEIA = "sagemaker-tensorflow-eia" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker DLC SageMaker TensorFlow Serving
	repositoryTensorFlowServing = "sagemaker-tensorflow-serving"
	// SageMaker DLC SageMaker TensorFlow Serving EIA
	repositoryTensorFlowServingEIA = "sagemaker-tensorflow-serving-eia"
	// SageMaker DLC SageMaker TensorFlow Serving
	repositorySageMakerTensorFlowScriptMode = "sagemaker-tensorflow-scriptmode" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker DLC Tensorflow Coach
	repositoryTensorflowCoach = "sagemaker-rl-tensorflow"
	// SageMaker DLC Tensorflow Inferentia
	repositoryTensorflowInferentia = "sagemaker-neo-tensorflow"
	// SageMaker Algorithm StabilityAI
	repositoryStabilityAI = "stabilityai-pytorch-inference"
	// SageMaker Algorithm VW
	repositoryVW = "sagemaker-rl-vw-container"
	// SageMaker Algorithm XGBoost
	repositoryXGBoost = "sagemaker-xgboost"
	// SageMaker Base Python
	repositorySageMakerBasePython = "sagemaker-base-python" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker Geospatial
	repositorySageMakerGeospatial = "sagemaker-geospatial-v1-0" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker NVIDIA Triton Inference
	repositoryNVIDIATritonInference = "sagemaker-tritonserver"
)

// https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/blazingtext.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/image-classification.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/object-detection.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/semantic-segmentation.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/seq2seq.json

var prebuiltECRImageIDByRegion_blazing = map[string]string{
	endpoints.AfSouth1RegionID:     "455444449433",
	endpoints.ApEast1RegionID:      "286214385809",
	endpoints.ApNortheast1RegionID: "501404015308",
	endpoints.ApNortheast2RegionID: "306986355934",
	endpoints.ApNortheast3RegionID: "867004704886",
	endpoints.ApSouth1RegionID:     "991648021394",
	endpoints.ApSouth2RegionID:     "628508329040",
	endpoints.ApSoutheast1RegionID: "475088953585",
	endpoints.ApSoutheast2RegionID: "544295431143",
	endpoints.ApSoutheast3RegionID: "951798379941",
	endpoints.ApSoutheast4RegionID: "106583098589",
	endpoints.CaCentral1RegionID:   "469771592824",
	endpoints.CaWest1RegionID:      "190319476487",
	endpoints.CnNorth1RegionID:     "390948362332",
	endpoints.CnNorthwest1RegionID: "387376663083",
	endpoints.EuCentral1RegionID:   "813361260812",
	endpoints.EuCentral2RegionID:   "680994064768",
	endpoints.EuNorth1RegionID:     "669576153137",
	endpoints.EuSouth1RegionID:     "257386234256",
	endpoints.EuSouth2RegionID:     "104374241257",
	endpoints.EuWest1RegionID:      "685385470294",
	endpoints.EuWest2RegionID:      "644912444149",
	endpoints.EuWest3RegionID:      "749696950732",
	endpoints.IlCentral1RegionID:   "898809789911",
	endpoints.MeCentral1RegionID:   "272398656194",
	endpoints.MeSouth1RegionID:     "249704162688",
	endpoints.SaEast1RegionID:      "855470959533",
	endpoints.UsEast1RegionID:      "811284229777",
	endpoints.UsEast2RegionID:      "825641698319",
	endpoints.UsGovEast1RegionID:   "237065988967",
	endpoints.UsGovWest1RegionID:   "226302683700",
	endpoints.UsIsoEast1RegionID:   "490574956308",
	endpoints.UsIsobEast1RegionID:  "765400339828",
	endpoints.UsWest1RegionID:      "632365934929",
	endpoints.UsWest2RegionID:      "433757028032",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/clarify.json

var prebuiltECRImageIDByRegion_clarify = map[string]string{
	endpoints.AfSouth1RegionID:     "811711786498",
	endpoints.ApEast1RegionID:      "098760798382",
	endpoints.ApNortheast1RegionID: "377024640650",
	endpoints.ApNortheast2RegionID: "263625296855",
	endpoints.ApNortheast3RegionID: "912233562940",
	endpoints.ApSouth1RegionID:     "452307495513",
	endpoints.ApSoutheast1RegionID: "834264404009",
	endpoints.ApSoutheast2RegionID: "007051062584",
	endpoints.ApSoutheast3RegionID: "705930551576",
	endpoints.CaCentral1RegionID:   "675030665977",
	endpoints.CnNorth1RegionID:     "122526803553",
	endpoints.CnNorthwest1RegionID: "122578899357",
	endpoints.EuCentral1RegionID:   "017069133835",
	endpoints.EuNorth1RegionID:     "763603941244",
	endpoints.EuSouth1RegionID:     "638885417683",
	endpoints.EuWest1RegionID:      "131013547314",
	endpoints.EuWest2RegionID:      "440796970383",
	endpoints.EuWest3RegionID:      "341593696636",
	endpoints.MeSouth1RegionID:     "835444307964",
	endpoints.SaEast1RegionID:      "520018980103",
	endpoints.UsEast1RegionID:      "205585389593",
	endpoints.UsEast2RegionID:      "211330385671",
	endpoints.UsGovWest1RegionID:   "598674086554",
	endpoints.UsWest1RegionID:      "740489534195",
	endpoints.UsWest2RegionID:      "306415355426",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/data-wrangler.json

var prebuiltECRImageIDByRegion_dataWrangler = map[string]string{
	endpoints.AfSouth1RegionID:     "143210264188",
	endpoints.ApEast1RegionID:      "707077482487",
	endpoints.ApNortheast1RegionID: "649008135260",
	endpoints.ApNortheast2RegionID: "131546521161",
	endpoints.ApNortheast3RegionID: "913387583493",
	endpoints.ApSouth1RegionID:     "089933028263",
	endpoints.ApSoutheast1RegionID: "119527597002",
	endpoints.ApSoutheast2RegionID: "422173101802",
	endpoints.CaCentral1RegionID:   "557239378090",
	endpoints.CnNorth1RegionID:     "245909111842",
	endpoints.CnNorthwest1RegionID: "249157047649",
	endpoints.EuCentral1RegionID:   "024640144536",
	endpoints.EuNorth1RegionID:     "054986407534",
	endpoints.EuSouth1RegionID:     "488287956546",
	endpoints.EuWest1RegionID:      "245179582081",
	endpoints.EuWest2RegionID:      "894491911112",
	endpoints.EuWest3RegionID:      "807237891255",
	endpoints.IlCentral1RegionID:   "406833011540",
	endpoints.MeSouth1RegionID:     "376037874950",
	endpoints.SaEast1RegionID:      "424196993095",
	endpoints.UsEast1RegionID:      "663277389841",
	endpoints.UsEast2RegionID:      "415577184552",
	endpoints.UsWest1RegionID:      "926135532090",
	endpoints.UsWest2RegionID:      "174368400705",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/debugger.json

var prebuiltECRImageIDByRegion_debugger = map[string]string{
	endpoints.AfSouth1RegionID:     "314341159256",
	endpoints.ApEast1RegionID:      "199566480951",
	endpoints.ApNortheast1RegionID: "430734990657",
	endpoints.ApNortheast2RegionID: "578805364391",
	endpoints.ApNortheast3RegionID: "479947661362",
	endpoints.ApSouth1RegionID:     "904829902805",
	endpoints.ApSoutheast1RegionID: "972752614525",
	endpoints.ApSoutheast2RegionID: "184798709955",
	endpoints.CaCentral1RegionID:   "519511493484",
	endpoints.CnNorth1RegionID:     "618459771430",
	endpoints.CnNorthwest1RegionID: "658757709296",
	endpoints.EuCentral1RegionID:   "482524230118",
	endpoints.EuNorth1RegionID:     "314864569078",
	endpoints.EuSouth1RegionID:     "563282790590",
	endpoints.EuWest1RegionID:      "929884845733",
	endpoints.EuWest2RegionID:      "250201462417",
	endpoints.EuWest3RegionID:      "447278800020",
	endpoints.MeSouth1RegionID:     "986000313247",
	endpoints.SaEast1RegionID:      "818342061345",
	endpoints.UsEast1RegionID:      "503895931360",
	endpoints.UsEast2RegionID:      "915447279597",
	endpoints.UsGovWest1RegionID:   "515509971035",
	endpoints.UsWest1RegionID:      "685455198987",
	endpoints.UsWest2RegionID:      "895741380848",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/inferentia-mxnet.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/inferentia-pytorch.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/image-classification-neo.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/neo-mxnet.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/neo-pytorch.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/neo-tensorflow.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/xgboost-neo.json

var prebuiltECRImageIDByRegion_inferentiaNeo = map[string]string{
	endpoints.AfSouth1RegionID:     "774647643957",
	endpoints.ApEast1RegionID:      "110948597952",
	endpoints.ApNortheast1RegionID: "941853720454",
	endpoints.ApNortheast2RegionID: "151534178276",
	endpoints.ApNortheast3RegionID: "925152966179",
	endpoints.ApSouth1RegionID:     "763008648453",
	endpoints.ApSoutheast1RegionID: "324986816169",
	endpoints.ApSoutheast2RegionID: "355873309152",
	endpoints.CaCentral1RegionID:   "464438896020",
	endpoints.CnNorth1RegionID:     "472730292857",
	endpoints.CnNorthwest1RegionID: "474822919863",
	endpoints.EuCentral1RegionID:   "746233611703",
	endpoints.EuNorth1RegionID:     "601324751636",
	endpoints.EuSouth1RegionID:     "966458181534",
	endpoints.EuWest1RegionID:      "802834080501",
	endpoints.EuWest2RegionID:      "205493899709",
	endpoints.EuWest3RegionID:      "254080097072",
	endpoints.IlCentral1RegionID:   "275950707576",
	endpoints.MeSouth1RegionID:     "836785723513",
	endpoints.SaEast1RegionID:      "756306329178",
	endpoints.UsEast1RegionID:      "785573368785",
	endpoints.UsEast2RegionID:      "007439368137",
	endpoints.UsGovWest1RegionID:   "263933020539",
	endpoints.UsIsoEast1RegionID:   "167761179201",
	endpoints.UsIsobEast1RegionID:  "406031935815",
	endpoints.UsWest1RegionID:      "710691900526",
	endpoints.UsWest2RegionID:      "301217895009",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/chainer.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/pytorch.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/coach-mxnet.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/mxnet.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/coach-tensorflow.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/tensorflow.json

var prebuiltECRImageIDByRegion_SageMakerCustom = map[string]string{ // nosemgrep:ci.sagemaker-in-var-name
	endpoints.AfSouth1RegionID:     "313743910680",
	endpoints.ApEast1RegionID:      "057415533634",
	endpoints.ApNortheast1RegionID: "520713654638",
	endpoints.ApNortheast2RegionID: "520713654638",
	endpoints.ApSouth1RegionID:     "520713654638",
	endpoints.ApSoutheast1RegionID: "520713654638",
	endpoints.ApSoutheast2RegionID: "520713654638",
	endpoints.CaCentral1RegionID:   "520713654638",
	endpoints.CnNorth1RegionID:     "422961961927",
	endpoints.CnNorthwest1RegionID: "423003514399",
	endpoints.EuCentral1RegionID:   "520713654638",
	endpoints.EuNorth1RegionID:     "520713654638",
	endpoints.EuSouth1RegionID:     "048378556238",
	endpoints.EuWest1RegionID:      "520713654638",
	endpoints.EuWest2RegionID:      "520713654638",
	endpoints.EuWest3RegionID:      "520713654638",
	endpoints.MeSouth1RegionID:     "724002660598",
	endpoints.SaEast1RegionID:      "520713654638",
	endpoints.UsEast1RegionID:      "520713654638",
	endpoints.UsEast2RegionID:      "520713654638",
	endpoints.UsGovWest1RegionID:   "246785580436",
	endpoints.UsIsoEast1RegionID:   "744548109606",
	endpoints.UsIsobEast1RegionID:  "453391408702",
	endpoints.UsWest1RegionID:      "520713654638",
	endpoints.UsWest2RegionID:      "520713654638",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/ray-pytorch.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/coach-tensorflow.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/vw.json

var prebuiltECRImageIDByRegion_SageMakerRL = map[string]string{ // nosemgrep:ci.sagemaker-in-var-name
	endpoints.ApNortheast1RegionID: "462105765813",
	endpoints.ApNortheast2RegionID: "462105765813",
	endpoints.ApSouth1RegionID:     "462105765813",
	endpoints.ApSoutheast1RegionID: "462105765813",
	endpoints.ApSoutheast2RegionID: "462105765813",
	endpoints.CaCentral1RegionID:   "462105765813",
	endpoints.EuCentral1RegionID:   "462105765813",
	endpoints.EuWest1RegionID:      "462105765813",
	endpoints.EuWest2RegionID:      "462105765813",
	endpoints.UsEast1RegionID:      "462105765813",
	endpoints.UsEast2RegionID:      "462105765813",
	endpoints.UsWest1RegionID:      "462105765813",
	endpoints.UsWest2RegionID:      "462105765813",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/spark.json

var prebuiltECRImageIDByRegion_spark = map[string]string{
	endpoints.AfSouth1RegionID:     "309385258863",
	endpoints.ApEast1RegionID:      "732049463269",
	endpoints.ApNortheast1RegionID: "411782140378",
	endpoints.ApNortheast2RegionID: "860869212795",
	endpoints.ApNortheast3RegionID: "102471314380",
	endpoints.ApSouth1RegionID:     "105495057255",
	endpoints.ApSouth2RegionID:     "873151114052",
	endpoints.ApSoutheast1RegionID: "759080221371",
	endpoints.ApSoutheast2RegionID: "440695851116",
	endpoints.ApSoutheast3RegionID: "800295151634",
	endpoints.ApSoutheast4RegionID: "819679513684",
	endpoints.CaCentral1RegionID:   "446299261295",
	endpoints.CaWest1RegionID:      "000907499111",
	endpoints.CnNorth1RegionID:     "671472414489",
	endpoints.CnNorthwest1RegionID: "844356804704",
	endpoints.EuCentral1RegionID:   "906073651304",
	endpoints.EuCentral2RegionID:   "142351485170",
	endpoints.EuNorth1RegionID:     "330188676905",
	endpoints.EuSouth1RegionID:     "753923664805",
	endpoints.EuSouth2RegionID:     "833944533722",
	endpoints.EuWest1RegionID:      "571004829621",
	endpoints.EuWest2RegionID:      "836651553127",
	endpoints.EuWest3RegionID:      "136845547031",
	endpoints.IlCentral1RegionID:   "408426139102",
	endpoints.MeCentral1RegionID:   "395420993607",
	endpoints.MeSouth1RegionID:     "750251592176",
	endpoints.SaEast1RegionID:      "737130764395",
	endpoints.UsEast1RegionID:      "173754725891",
	endpoints.UsEast2RegionID:      "314815235551",
	endpoints.UsGovEast1RegionID:   "260923028637",
	endpoints.UsGovWest1RegionID:   "271483468897",
	endpoints.UsWest1RegionID:      "667973535471",
	endpoints.UsWest2RegionID:      "153931337802",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/sagemaker-base-python.json

var prebuiltECRImageIDByRegion_SageMakerBasePython = map[string]string{ // nosemgrep:ci.sagemaker-in-var-name
	endpoints.AfSouth1RegionID:     "559312083959",
	endpoints.ApEast1RegionID:      "493642496378",
	endpoints.ApNortheast1RegionID: "102112518831",
	endpoints.ApNortheast2RegionID: "806072073708",
	endpoints.ApNortheast3RegionID: "792733760839",
	endpoints.ApSouth1RegionID:     "394103062818",
	endpoints.ApSoutheast1RegionID: "492261229750",
	endpoints.ApSoutheast2RegionID: "452832661640",
	endpoints.ApSoutheast3RegionID: "276181064229",
	endpoints.CaCentral1RegionID:   "310906938811",
	endpoints.CnNorth1RegionID:     "390048526115",
	endpoints.CnNorthwest1RegionID: "390780980154",
	endpoints.EuCentral1RegionID:   "936697816551",
	endpoints.EuNorth1RegionID:     "243637512696",
	endpoints.EuSouth1RegionID:     "592751261982",
	endpoints.EuSouth2RegionID:     "127363102723",
	endpoints.EuWest1RegionID:      "470317259841",
	endpoints.EuWest2RegionID:      "712779665605",
	endpoints.EuWest3RegionID:      "615547856133",
	endpoints.IlCentral1RegionID:   "380164790875",
	endpoints.MeCentral1RegionID:   "103105715889",
	endpoints.MeSouth1RegionID:     "117516905037",
	endpoints.SaEast1RegionID:      "782484402741",
	endpoints.UsEast1RegionID:      "081325390199",
	endpoints.UsEast2RegionID:      "429704687514",
	endpoints.UsGovEast1RegionID:   "107072934176",
	endpoints.UsGovWest1RegionID:   "107173498710",
	endpoints.UsWest1RegionID:      "742091327244",
	endpoints.UsWest2RegionID:      "236514542706",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/sagemaker-geospatial.json

var prebuiltECRImageIDByRegion_SageMakerGeospatial = map[string]string{ // nosemgrep:ci.sagemaker-in-var-name
	endpoints.UsWest2RegionID: "081189585635",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/forecasting-deepar.json

var prebuiltECRImageIDByRegion_deepAR = map[string]string{
	endpoints.AfSouth1RegionID:     "455444449433",
	endpoints.ApEast1RegionID:      "286214385809",
	endpoints.ApNortheast1RegionID: "633353088612",
	endpoints.ApNortheast2RegionID: "204372634319",
	endpoints.ApNortheast3RegionID: "867004704886",
	endpoints.ApSouth1RegionID:     "991648021394",
	endpoints.ApSouth2RegionID:     "628508329040",
	endpoints.ApSoutheast1RegionID: "475088953585",
	endpoints.ApSoutheast2RegionID: "514117268639",
	endpoints.ApSoutheast3RegionID: "951798379941",
	endpoints.ApSoutheast4RegionID: "106583098589",
	endpoints.CaCentral1RegionID:   "469771592824",
	endpoints.CaWest1RegionID:      "190319476487",
	endpoints.CnNorth1RegionID:     "390948362332",
	endpoints.CnNorthwest1RegionID: "387376663083",
	endpoints.EuCentral1RegionID:   "495149712605",
	endpoints.EuCentral2RegionID:   "680994064768",
	endpoints.EuNorth1RegionID:     "669576153137",
	endpoints.EuSouth1RegionID:     "257386234256",
	endpoints.EuSouth2RegionID:     "104374241257",
	endpoints.EuWest1RegionID:      "224300973850",
	endpoints.EuWest2RegionID:      "644912444149",
	endpoints.EuWest3RegionID:      "749696950732",
	endpoints.IlCentral1RegionID:   "898809789911",
	endpoints.MeCentral1RegionID:   "272398656194",
	endpoints.MeSouth1RegionID:     "249704162688",
	endpoints.SaEast1RegionID:      "855470959533",
	endpoints.UsEast1RegionID:      "522234722520",
	endpoints.UsEast2RegionID:      "566113047672",
	endpoints.UsGovEast1RegionID:   "237065988967",
	endpoints.UsGovWest1RegionID:   "226302683700",
	endpoints.UsIsoEast1RegionID:   "490574956308",
	endpoints.UsIsobEast1RegionID:  "765400339828",
	endpoints.UsWest1RegionID:      "632365934929",
	endpoints.UsWest2RegionID:      "156387875391",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/factorization-machines.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/ipinsights.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/linear-learner.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/ntm.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/object2vec.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/pca.json

var prebuiltECRImageIDByRegion_factorMachines = map[string]string{
	endpoints.AfSouth1RegionID:     "455444449433",
	endpoints.ApEast1RegionID:      "286214385809",
	endpoints.ApNortheast1RegionID: "351501993468",
	endpoints.ApNortheast2RegionID: "835164637446",
	endpoints.ApNortheast3RegionID: "867004704886",
	endpoints.ApSouth1RegionID:     "991648021394",
	endpoints.ApSouth2RegionID:     "628508329040",
	endpoints.ApSoutheast1RegionID: "475088953585",
	endpoints.ApSoutheast2RegionID: "712309505854",
	endpoints.ApSoutheast3RegionID: "951798379941",
	endpoints.ApSoutheast4RegionID: "106583098589",
	endpoints.CaCentral1RegionID:   "469771592824",
	endpoints.CaWest1RegionID:      "190319476487",
	endpoints.CnNorth1RegionID:     "390948362332",
	endpoints.CnNorthwest1RegionID: "387376663083",
	endpoints.EuCentral1RegionID:   "664544806723",
	endpoints.EuCentral2RegionID:   "680994064768",
	endpoints.EuNorth1RegionID:     "669576153137",
	endpoints.EuSouth1RegionID:     "257386234256",
	endpoints.EuSouth2RegionID:     "104374241257",
	endpoints.EuWest1RegionID:      "438346466558",
	endpoints.EuWest2RegionID:      "644912444149",
	endpoints.EuWest3RegionID:      "749696950732",
	endpoints.IlCentral1RegionID:   "898809789911",
	endpoints.MeCentral1RegionID:   "272398656194",
	endpoints.MeSouth1RegionID:     "249704162688",
	endpoints.SaEast1RegionID:      "855470959533",
	endpoints.UsEast1RegionID:      "382416733822",
	endpoints.UsEast2RegionID:      "404615174143",
	endpoints.UsGovEast1RegionID:   "237065988967",
	endpoints.UsGovWest1RegionID:   "226302683700",
	endpoints.UsIsoEast1RegionID:   "490574956308",
	endpoints.UsIsobEast1RegionID:  "765400339828",
	endpoints.UsWest1RegionID:      "632365934929",
	endpoints.UsWest2RegionID:      "174872318107",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/lda.json

var prebuiltECRImageIDByRegion_lda = map[string]string{
	endpoints.ApNortheast1RegionID: "258307448986",
	endpoints.ApNortheast2RegionID: "293181348795",
	endpoints.ApSouth1RegionID:     "991648021394",
	endpoints.ApSoutheast1RegionID: "475088953585",
	endpoints.ApSoutheast2RegionID: "297031611018",
	endpoints.CaCentral1RegionID:   "469771592824",
	endpoints.EuCentral1RegionID:   "353608530281",
	endpoints.EuWest1RegionID:      "999678624901",
	endpoints.EuWest2RegionID:      "644912444149",
	endpoints.UsEast1RegionID:      "766337827248",
	endpoints.UsEast2RegionID:      "999911452149",
	endpoints.UsGovWest1RegionID:   "226302683700",
	endpoints.UsIsoEast1RegionID:   "490574956308",
	endpoints.UsIsobEast1RegionID:  "765400339828",
	endpoints.UsWest1RegionID:      "632365934929",
	endpoints.UsWest2RegionID:      "266724342769",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/xgboost.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/huggingface-tei.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/huggingface-tei-cpu.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/sparkml-serving.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/sklearn.json

var prebuiltECRImageIDByRegion_XGBoost = map[string]string{
	endpoints.AfSouth1RegionID:     "510948584623",
	endpoints.ApEast1RegionID:      "651117190479",
	endpoints.ApNortheast1RegionID: "354813040037",
	endpoints.ApNortheast2RegionID: "366743142698",
	endpoints.ApNortheast3RegionID: "867004704886",
	endpoints.ApSouth1RegionID:     "720646828776",
	endpoints.ApSouth2RegionID:     "628508329040",
	endpoints.ApSoutheast1RegionID: "121021644041",
	endpoints.ApSoutheast2RegionID: "783357654285",
	endpoints.ApSoutheast3RegionID: "951798379941",
	endpoints.ApSoutheast4RegionID: "106583098589",
	endpoints.CaCentral1RegionID:   "341280168497",
	endpoints.CaWest1RegionID:      "190319476487",
	endpoints.CnNorth1RegionID:     "450853457545",
	endpoints.CnNorthwest1RegionID: "451049120500",
	endpoints.EuCentral1RegionID:   "492215442770",
	endpoints.EuCentral2RegionID:   "680994064768",
	endpoints.EuNorth1RegionID:     "662702820516",
	endpoints.EuSouth1RegionID:     "978288397137",
	endpoints.EuSouth2RegionID:     "104374241257",
	endpoints.EuWest1RegionID:      "141502667606",
	endpoints.EuWest2RegionID:      "764974769150",
	endpoints.EuWest3RegionID:      "659782779980",
	endpoints.IlCentral1RegionID:   "898809789911",
	endpoints.MeCentral1RegionID:   "272398656194",
	endpoints.MeSouth1RegionID:     "801668240914",
	endpoints.SaEast1RegionID:      "737474898029",
	endpoints.UsEast1RegionID:      "683313688378",
	endpoints.UsEast2RegionID:      "257758044811",
	endpoints.UsGovEast1RegionID:   "237065988967",
	endpoints.UsGovWest1RegionID:   "414596584902",
	endpoints.UsIsoEast1RegionID:   "833128469047",
	endpoints.UsIsobEast1RegionID:  "281123927165",
	endpoints.UsWest1RegionID:      "746614075791",
	endpoints.UsWest2RegionID:      "246618743249",
}

// https://github.com/aws/deep-learning-containers/blob/master/available_images.md
// https://github.com/aws/sagemaker-tensorflow-serving-container
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/autogluon.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/djl-deepspeed.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/djl-fastertransformer.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/djl-lmi.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/djl-neuronx.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/djl-tensorrtllm.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/pytorch.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/stabilityai.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/sagemaker-tritonserver.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/tensorflow.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/huggingface-llm.json

var prebuiltECRImageIDByRegion_deepLearning = map[string]string{
	endpoints.AfSouth1RegionID:     "626614931356",
	endpoints.ApEast1RegionID:      "871362719292",
	endpoints.ApNortheast1RegionID: "763104351884",
	endpoints.ApNortheast2RegionID: "763104351884",
	endpoints.ApNortheast3RegionID: "364406365360",
	endpoints.ApSouth1RegionID:     "763104351884",
	endpoints.ApSouth2RegionID:     "772153158452",
	endpoints.ApSoutheast1RegionID: "763104351884",
	endpoints.ApSoutheast2RegionID: "763104351884",
	endpoints.ApSoutheast3RegionID: "907027046896",
	endpoints.ApSoutheast4RegionID: "457447274322",
	endpoints.CaCentral1RegionID:   "763104351884",
	endpoints.CaWest1RegionID:      "204538143572",
	endpoints.CnNorth1RegionID:     "727897471807",
	endpoints.CnNorthwest1RegionID: "727897471807",
	endpoints.EuCentral1RegionID:   "763104351884",
	endpoints.EuCentral2RegionID:   "380420809688",
	endpoints.EuNorth1RegionID:     "763104351884",
	endpoints.EuWest1RegionID:      "763104351884",
	endpoints.EuWest2RegionID:      "763104351884",
	endpoints.EuWest3RegionID:      "763104351884",
	endpoints.EuSouth1RegionID:     "692866216735",
	endpoints.EuSouth2RegionID:     "503227376785",
	endpoints.IlCentral1RegionID:   "780543022126",
	endpoints.MeCentral1RegionID:   "914824155844",
	endpoints.MeSouth1RegionID:     "217643126080",
	endpoints.SaEast1RegionID:      "763104351884",
	endpoints.UsEast1RegionID:      "763104351884",
	endpoints.UsEast2RegionID:      "763104351884",
	endpoints.UsWest1RegionID:      "763104351884",
	endpoints.UsWest2RegionID:      "763104351884",
	endpoints.UsGovEast1RegionID:   "446045086412",
	endpoints.UsGovWest1RegionID:   "442386744353",
	endpoints.UsIsoEast1RegionID:   "886529160074",
	endpoints.UsIsobEast1RegionID:  "094389454867",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/model-monitor.json

var prebuiltECRImageIDByRegion_modelMonitor = map[string]string{
	endpoints.AfSouth1RegionID:     "875698925577",
	endpoints.ApEast1RegionID:      "001633400207",
	endpoints.ApNortheast1RegionID: "574779866223",
	endpoints.ApNortheast2RegionID: "709848358524",
	endpoints.ApNortheast3RegionID: "990339680094",
	endpoints.ApSouth1RegionID:     "126357580389",
	endpoints.ApSoutheast1RegionID: "245545462676",
	endpoints.ApSoutheast2RegionID: "563025443158",
	endpoints.ApSoutheast3RegionID: "669540362728",
	endpoints.CaCentral1RegionID:   "536280801234",
	endpoints.CnNorth1RegionID:     "453000072557",
	endpoints.CnNorthwest1RegionID: "453252182341",
	endpoints.EuCentral1RegionID:   "048819808253",
	endpoints.EuNorth1RegionID:     "895015795356",
	endpoints.EuSouth1RegionID:     "933208885752",
	endpoints.EuSouth2RegionID:     "437450045455",
	endpoints.EuWest1RegionID:      "468650794304",
	endpoints.EuWest2RegionID:      "749857270468",
	endpoints.EuWest3RegionID:      "680080141114",
	endpoints.IlCentral1RegionID:   "843974653677",
	endpoints.MeCentral1RegionID:   "588750061953",
	endpoints.MeSouth1RegionID:     "607024016150",
	endpoints.SaEast1RegionID:      "539772159869",
	endpoints.UsEast1RegionID:      "156813124566",
	endpoints.UsEast2RegionID:      "777275614652",
	endpoints.UsWest1RegionID:      "890145073186",
	endpoints.UsWest2RegionID:      "159807026194",
}

// @SDKDataSource("aws_sagemaker_prebuilt_ecr_image", name="Prebuilt ECR Image")
func dataSourcePrebuiltECRImage() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePrebuiltECRImageRead,
		Schema: map[string]*schema.Schema{
			names.AttrRepositoryName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					repositoryAutoGluonTraining,
					repositoryAutoGluonInference,
					repositoryBlazingText,
					repositoryChainer,
					repositoryClarify,
					repositoryDJLDeepSpeed,
					repositoryDataWrangler,
					repositoryDebugger,
					repositoryDeepARForecasting,
					repositoryFactorizationMachines,
					repositoryHuggingFaceTensorFlowTraining,
					repositoryHuggingFacePyTorchTraining,
					repositoryHuggingFacePyTorchTrainingNeuronX,
					repositoryHuggingFacePyTorchTrainingCompiler,
					repositoryHuggingFaceTensorFlowTrainingCompiler,
					repositoryHuggingFaceTensorFlowInference,
					repositoryHuggingFacePyTorchInference,
					repositoryHuggingFacePyTorchInferenceNeuron,
					repositoryHuggingFacePyTorchInferenceNeuronX,
					repositoryHuggingFacePyTorchTGIInference,
					repositoryHuggingFaceTEI,
					repositoryHuggingFaceTEICPU,
					repositoryIPInsights,
					repositoryImageClassification,
					repositoryInferentiaMXNet,
					repositoryInferentiaPyTorch,
					repositoryKMeans,
					repositoryKNearestNeighbor,
					repositoryLDA,
					repositoryLinearLearner,
					repositoryModelMonitor,
					repositoryMXNetTraining,
					repositoryMXNetInference,
					repositoryMXNetInferenceEIA,
					repositoryMXNetCoach,
					repositoryNeuralTopicModel,
					repositoryNeoImageClassification,
					repositoryNeoMXNet,
					repositoryNeoPyTorch,
					repositoryNeoTensorflow,
					repositoryNeoXGBoost,
					repositoryNVIDIATritonInference,
					repositoryObjectDetection,
					repositoryObject2Vec,
					repositoryPCA,
					repositoryPyTorchTraining,
					repositoryPyTorchTrainingNeuronX,
					repositoryPyTorchTrainingCompiler,
					repositoryPyTorchInference,
					repositoryPyTorchInferenceEIA,
					repositoryPyTorchInferenceGraviton,
					repositoryPyTorchInferenceNeuronX,
					repositoryRandomCutForest,
					repositoryRLRayPyTorch,
					repositoryRLCoachPyTorch,
					repositorySageMakerBasePython,
					repositorySageMakerGeospatial,
					repositorySageMakerMXNet,
					repositorySageMakerMXNetServing,
					repositorySageMakerMXNetEIA,
					repositorySageMakerMXNetServingEIA,
					repositorySageMakerPyTorch,
					repositorySageMakerTensorFlow,
					repositorySageMakerTensorFlowEIA,
					repositoryScikitLearn,
					repositorySemanticSegmentation,
					repositorySeq2Seq,
					repositorySpark,
					repositorySparkML,
					repositoryTensorFlowTraining,
					repositoryTensorFlowInference,
					repositoryTensorFlowInferenceEIA,
					repositoryTensorFlowInferenceGraviton,
					repositoryTensorFlowServing,
					repositoryTensorFlowServingEIA,
					repositorySageMakerTensorFlowScriptMode,
					repositoryTensorflowCoach,
					repositoryTensorflowInferentia,
					repositoryStabilityAI,
					repositoryVW,
					repositoryXGBoost,
				}, false),
			},

			"dns_suffix": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"image_tag": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1",
			},

			names.AttrRegion: {
				Type:     schema.TypeString,
				Optional: true,
			},

			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"registry_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePrebuiltECRImageRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	region := meta.(*conns.AWSClient).Region(ctx)
	if v, ok := d.GetOk(names.AttrRegion); ok {
		region = v.(string)
	}

	suffix := meta.(*conns.AWSClient).DNSSuffix(ctx)
	if v, ok := d.GetOk("dns_suffix"); ok {
		suffix = v.(string)
	}

	repo := d.Get(names.AttrRepositoryName).(string)

	var id string
	switch repo {
	case repositoryBlazingText,
		repositoryImageClassification,
		repositoryObjectDetection,
		repositorySemanticSegmentation,
		repositorySeq2Seq:
		id = prebuiltECRImageIDByRegion_blazing[region]
	case repositoryClarify:
		id = prebuiltECRImageIDByRegion_clarify[region]
	case repositoryDataWrangler:
		id = prebuiltECRImageIDByRegion_dataWrangler[region]
	case repositoryDebugger:
		id = prebuiltECRImageIDByRegion_debugger[region]
	case repositoryDeepARForecasting:
		id = prebuiltECRImageIDByRegion_deepAR[region]
	case repositoryInferentiaMXNet,
		repositoryInferentiaPyTorch,
		repositoryMXNetCoach,
		repositoryNeoImageClassification,
		repositoryNeoMXNet,
		repositoryNeoPyTorch,
		repositoryNeoTensorflow,
		repositoryNeoXGBoost,
		repositoryTensorflowInferentia:
		id = prebuiltECRImageIDByRegion_inferentiaNeo[region]
	case repositoryLDA:
		id = prebuiltECRImageIDByRegion_lda[region]
	case repositoryModelMonitor:
		id = prebuiltECRImageIDByRegion_modelMonitor[region]
	case repositoryXGBoost,
		repositoryScikitLearn,
		repositorySparkML,
		repositoryHuggingFaceTEI,
		repositoryHuggingFaceTEICPU:
		id = prebuiltECRImageIDByRegion_XGBoost[region]
	case repositoryChainer,
		repositorySageMakerMXNet,
		repositorySageMakerMXNetServing,
		repositorySageMakerMXNetEIA,
		repositorySageMakerMXNetServingEIA,
		repositorySageMakerPyTorch,
		repositorySageMakerTensorFlow,
		repositorySageMakerTensorFlowEIA,
		repositorySageMakerTensorFlowScriptMode,
		repositoryTensorflowCoach,
		repositoryTensorFlowServing,
		repositoryTensorFlowServingEIA:
		id = prebuiltECRImageIDByRegion_SageMakerCustom[region]
	case repositoryAutoGluonTraining,
		repositoryAutoGluonInference,
		repositoryDJLDeepSpeed,
		repositoryHuggingFaceTensorFlowTraining,
		repositoryHuggingFacePyTorchTraining,
		repositoryHuggingFacePyTorchTrainingNeuronX,
		repositoryHuggingFacePyTorchTrainingCompiler,
		repositoryHuggingFaceTensorFlowTrainingCompiler,
		repositoryHuggingFaceTensorFlowInference,
		repositoryHuggingFacePyTorchInference,
		repositoryHuggingFacePyTorchInferenceNeuron,
		repositoryHuggingFacePyTorchInferenceNeuronX,
		repositoryHuggingFacePyTorchTGIInference,
		repositoryMXNetTraining,
		repositoryMXNetInference,
		repositoryMXNetInferenceEIA,
		repositoryPyTorchTraining,
		repositoryPyTorchTrainingNeuronX,
		repositoryPyTorchTrainingCompiler,
		repositoryPyTorchInference,
		repositoryPyTorchInferenceEIA,
		repositoryPyTorchInferenceGraviton,
		repositoryPyTorchInferenceNeuronX,
		repositoryStabilityAI,
		repositoryTensorFlowTraining,
		repositoryTensorFlowInference,
		repositoryTensorFlowInferenceEIA,
		repositoryTensorFlowInferenceGraviton,
		repositoryNVIDIATritonInference:
		id = prebuiltECRImageIDByRegion_deepLearning[region]
	case repositoryRLRayPyTorch,
		repositoryRLCoachPyTorch,
		repositoryVW:
		id = prebuiltECRImageIDByRegion_SageMakerRL[region]
	case repositorySageMakerBasePython:
		id = prebuiltECRImageIDByRegion_SageMakerBasePython[region]
	case repositorySageMakerGeospatial:
		id = prebuiltECRImageIDByRegion_SageMakerGeospatial[region]
	case repositorySpark:
		id = prebuiltECRImageIDByRegion_spark[region]
	default:
		id = prebuiltECRImageIDByRegion_factorMachines[region]
	}

	if id == "" {
		return sdkdiag.AppendErrorf(diags, "no registry ID available for region (%s) and repository (%s)", region, repo)
	}

	d.SetId(id)
	d.Set("registry_id", id)
	d.Set("registry_path", prebuiltECRImageCreatePath(id, region, suffix, repo, d.Get("image_tag").(string)))
	return diags
}

func prebuiltECRImageCreatePath(id, region, suffix, repo, imageTag string) string {
	return fmt.Sprintf("%s.dkr.ecr.%s.%s/%s:%s", id, region, suffix, repo, imageTag)
}
