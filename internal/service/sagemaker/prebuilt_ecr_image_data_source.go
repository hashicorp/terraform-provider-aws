// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// SageMaker Algorithm AutoGluon Training
	RepositoryAutoGluonTraining = "autogluon-training"
	// SageMaker Algorithm AutoGluon Inference
	RepositoryAutoGluonInference = "autogluon-inference"
	// SageMaker Algorithm BlazingText
	RepositoryBlazingText = "blazingtext"
	// SageMaker DLC Chainer
	RepositoryChainer = "sagemaker-chainer"
	// SageMaker Algorithm Clarify
	RepositoryClarify = "sagemaker-clarify-processing"
	// SageMaker Algorithm DJL DeepSpeed
	RepositoryDJLDeepSpeed = "djl-inference"
	// SageMaker Algorithm Data Wrangler
	RepositoryDataWrangler = "sagemaker-data-wrangler-container"
	// SageMaker Algorithm Debugger
	RepositoryDebugger = "sagemaker-debugger-rules"
	// SageMaker Algorithm DeepAR Forecasting
	RepositoryDeepARForecasting = "forecasting-deepar"
	// SageMaker Algorithm Factorization Machines
	RepositoryFactorizationMachines = "factorization-machines"
	// SageMaker Algorithm HuggingFace TensorFlow Training
	RepositoryHuggingFaceTensorFlowTraining = "huggingface-tensorflow-training"
	// SageMaker Algorithm HuggingFace PyTorch Training
	RepositoryHuggingFacePyTorchTraining = "huggingface-pytorch-training"
	// SageMaker Algorithm HuggingFace PyTorch Training NeuronX
	RepositoryHuggingFacePyTorchTrainingNeuronX = "huggingface-pytorch-training-neuronx"
	// SageMaker Algorithm HuggingFace PyTorch Training Compiler
	RepositoryHuggingFacePyTorchTrainingCompiler = "huggingface-pytorch-trcomp-training"
	// SageMaker Algorithm HuggingFace TensorFlow Training Compiler
	RepositoryHuggingFaceTensorFlowTrainingCompiler = "huggingface-tensorflow-trcomp-training"
	// SageMaker Algorithm HuggingFace TensorFlow Inference
	RepositoryHuggingFaceTensorFlowInference = "huggingface-tensorflow-inference"
	// SageMaker Algorithm HuggingFace PyTorch Inference
	RepositoryHuggingFacePyTorchInference = "huggingface-pytorch-inference"
	// SageMaker Algorithm HuggingFace PyTorch Inference Neuron
	RepositoryHuggingFacePyTorchInferenceNeuron = "huggingface-pytorch-inference-neuron"
	// SageMaker Algorithm HuggingFace PyTorch Inference NeuronX
	RepositoryHuggingFacePyTorchInferenceNeuronX = "huggingface-pytorch-inference-neuronx"
	// SageMaker LLM HuggingFace Pytorch TGI Inference
	RepositoryHuggingFacePyTorchTGIInference = "huggingface-pytorch-tgi-inference"
	// SageMaker Algorithm HuggingFace TEI
	RepositoryHuggingFaceTEI = "tei"
	// SageMaker Algorithm HuggingFace TEI CPU
	RepositoryHuggingFaceTEICPU = "tei-cpu"
	// SageMaker Algorithm IP Insights
	RepositoryIPInsights = "ipinsights"
	// SageMaker Algorithm Image Classification
	RepositoryImageClassification = "image-classification"
	// SageMaker DLC Inferentia MXNet
	RepositoryInferentiaMXNet = "sagemaker-neo-mxnet"
	// SageMaker DLC Inferentia PyTorch
	RepositoryInferentiaPyTorch = "sagemaker-neo-pytorch"
	// SageMaker Algorithm k-means
	RepositoryKMeans = "kmeans"
	// SageMaker Algorithm k-nearest-neighbor
	RepositoryKNearestNeighbor = "knn"
	// SageMaker Algorithm Latent Dirichlet Allocation
	RepositoryLDA = "lda"
	// SageMaker Algorithm Linear Learner
	RepositoryLinearLearner = "linear-learner"
	// SageMaker DLC MXNet Training
	RepositoryMXNetTraining = "mxnet-training"
	// SageMaker DLC MXNet Inference
	RepositoryMXNetInference = "mxnet-inference"
	// SageMaker DLC MXNet Inference EIA
	RepositoryMXNetInferenceEIA = "mxnet-inference-eia"
	// SageMaker DLC SageMaker MXNet
	RepositorySageMakerMXNet = "sagemaker-mxnet" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker DLC SageMaker MXNet Serving
	RepositorySageMakerMXNetServing = "sagemaker-mxnet-serving" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker DLC SageMaker MXNet EIA
	RepositorySageMakerMXNetEIA = "sagemaker-mxnet-eia" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker DLC SageMaker MXNet Serving EIA
	RepositorySageMakerMXNetServingEIA = "sagemaker-mxnet-serving-eia" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker DLC MXNet Coach
	RepositoryMXNetCoach = "sagemaker-rl-mxnet"
	// SageMaker Model Monitor
	RepositoryModelMonitor = "sagemaker-model-monitor-analyzer"
	// SageMaker Algorithm Neural Topic Model
	RepositoryNeuralTopicModel = "ntm"
	// SageMaker Algorithm Neo Image Classification
	RepositoryNeoImageClassification = "image-classification-neo"
	// SageMaker DLC Neo MXNet
	RepositoryNeoMXNet = "sagemaker-inference-mxnet"
	// SageMaker DLC Neo PyTorch
	RepositoryNeoPyTorch = "sagemaker-inference-pytorch"
	// SageMaker DLC Neo Tensorflow
	RepositoryNeoTensorflow = "sagemaker-inference-tensorflow"
	// SageMaker DLC Neo XGBoost
	RepositoryNeoXGBoost = "xgboost-neo"
	// SageMaker Algorithm Object Detection
	RepositoryObjectDetection = "object-detection"
	// SageMaker Algorithm Object2Vec
	RepositoryObject2Vec = "object2vec"
	// SageMaker Algorithm PCA
	RepositoryPCA = "pca"
	// SageMaker DLC PyTorch Training
	RepositoryPyTorchTraining = "pytorch-training"
	// SageMaker DLC PyTorch Training NeuronX
	RepositoryPyTorchTrainingNeuronX = "pytorch-training-neuronx"
	// SageMaker DLC PyTorch Training Compiler
	RepositoryPyTorchTrainingCompiler = "pytorch-trcomp-training"
	// SageMaker DLC SageMaker PyTorch Inference
	RepositoryPyTorchInference = "pytorch-inference"
	// SageMaker DLC PyTorch Inference EIA
	RepositoryPyTorchInferenceEIA = "pytorch-inference-eia"
	// SageMaker DLC PyTorch Inference Graviton
	RepositoryPyTorchInferenceGraviton = "pytorch-inference-graviton"
	// SageMaker DLC PyTorch Inference NeuronX
	RepositoryPyTorchInferenceNeuronX = "pytorch-inference-neuronx"
	// SageMaker DLC SageMaker PyTorch
	RepositorySageMakerPyTorch = "sagemaker-pytorch" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker Algorithm Random Cut Forest
	RepositoryRandomCutForest = "randomcutforest"
	// SageMaker DLC RL Ray PyTorch
	RepositoryRLRayPyTorch = "sagemaker-rl-ray-container"
	// SageMaker DLC RL Coach Tensorflow
	RepositoryRLCoachPyTorch = "sagemaker-rl-coach-container"
	// SageMaker Library scikit-learn
	RepositoryScikitLearn = "sagemaker-scikit-learn"
	// SageMaker Algorithm Semantic Segmentation
	RepositorySemanticSegmentation = "semantic-segmentation"
	// SageMaker Algorithm Seq2Seq
	RepositorySeq2Seq = "seq2seq"
	// SageMaker Algorithm Spark
	RepositorySpark = "sagemaker-spark-processing"
	// SageMaker Algorithm Spark ML
	RepositorySparkML = "sagemaker-sparkml-serving"
	// SageMaker DLC TensorFlow Training
	RepositoryTensorFlowTraining = "tensorflow-training"
	// SageMaker DLC TensorFlow Inference
	RepositoryTensorFlowInference = "tensorflow-inference"
	// SageMaker Repo TensorFlow Inference EIA
	RepositoryTensorFlowInferenceEIA = "tensorflow-inference-eia"
	// SageMaker DLC TensorFlow Inference Graviton
	RepositoryTensorFlowInferenceGraviton = "tensorflow-inference-graviton"
	// SageMaker DLC SageMaker TensorFlow
	RepositorySageMakerTensorFlow = "sagemaker-tensorflow" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker DLC SageMaker TensorFlow EIA
	RepositorySageMakerTensorFlowEIA = "sagemaker-tensorflow-eia" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker DLC SageMaker TensorFlow Serving
	RepositoryTensorFlowServing = "sagemaker-tensorflow-serving"
	// SageMaker DLC SageMaker TensorFlow Serving EIA
	RepositoryTensorFlowServingEIA = "sagemaker-tensorflow-serving-eia"
	// SageMaker DLC SageMaker TensorFlow Serving
	RepositorySageMakerTensorFlowScriptMode = "sagemaker-tensorflow-scriptmode" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker DLC Tensorflow Coach
	RepositoryTensorflowCoach = "sagemaker-rl-tensorflow"
	// SageMaker DLC Tensorflow Inferentia
	RepositoryTensorflowInferentia = "sagemaker-neo-tensorflow"
	// SageMaker Algorithm StabilityAI
	RepositoryStabilityAI = "stabilityai-pytorch-inference"
	// SageMaker Algorithm VW
	RepositoryVW = "sagemaker-rl-vw-container"
	// SageMaker Algorithm XGBoost
	RepositoryXGBoost = "sagemaker-xgboost"
	// SageMaker Base Python
	RepositorySageMakerBasePython = "sagemaker-base-python" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker Geospatial
	RepositorySageMakerGeospatial = "sagemaker-geospatial-v1-0" // nosemgrep:ci.sagemaker-in-var-name,ci.sagemaker-in-const-name
	// SageMaker NVIDIA Triton Inference
	RepositoryNVIDIATritonInference = "sagemaker-tritonserver"
)

// https://docs.aws.amazon.com/sagemaker/latest/dg/sagemaker-algo-docker-registry-paths.html
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/blazingtext.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/image-classification.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/object-detection.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/semantic-segmentation.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/seq2seq.json

var PrebuiltECRImageIDByRegion_blazing = map[string]string{
	names.AFSouth1RegionID:     "455444449433",
	names.APEast1RegionID:      "286214385809",
	names.APNortheast1RegionID: "501404015308",
	names.APNortheast2RegionID: "306986355934",
	names.APNortheast3RegionID: "867004704886",
	names.APSouth1RegionID:     "991648021394",
	names.APSouth2RegionID:     "628508329040",
	names.APSoutheast1RegionID: "475088953585",
	names.APSoutheast2RegionID: "544295431143",
	names.APSoutheast3RegionID: "951798379941",
	names.APSoutheast4RegionID: "106583098589",
	names.CACentral1RegionID:   "469771592824",
	names.CAWest1RegionID:      "190319476487",
	names.CNNorth1RegionID:     "390948362332",
	names.CNNorthwest1RegionID: "387376663083",
	names.EUCentral1RegionID:   "813361260812",
	names.EUCentral2RegionID:   "680994064768",
	names.EUNorth1RegionID:     "669576153137",
	names.EUSouth1RegionID:     "257386234256",
	names.EUSouth2RegionID:     "104374241257",
	names.EUWest1RegionID:      "685385470294",
	names.EUWest2RegionID:      "644912444149",
	names.EUWest3RegionID:      "749696950732",
	names.ILCentral1RegionID:   "898809789911",
	names.MECentral1RegionID:   "272398656194",
	names.MESouth1RegionID:     "249704162688",
	names.SAEast1RegionID:      "855470959533",
	names.USEast1RegionID:      "811284229777",
	names.USEast2RegionID:      "825641698319",
	names.USGovEast1RegionID:   "237065988967",
	names.USGovWest1RegionID:   "226302683700",
	names.USISOEast1RegionID:   "490574956308",
	names.USISOBEast1RegionID:  "765400339828",
	names.USWest1RegionID:      "632365934929",
	names.USWest2RegionID:      "433757028032",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/clarify.json

var PrebuiltECRImageIDByRegion_clarify = map[string]string{
	names.AFSouth1RegionID:     "811711786498",
	names.APEast1RegionID:      "098760798382",
	names.APNortheast1RegionID: "377024640650",
	names.APNortheast2RegionID: "263625296855",
	names.APNortheast3RegionID: "912233562940",
	names.APSouth1RegionID:     "452307495513",
	names.APSoutheast1RegionID: "834264404009",
	names.APSoutheast2RegionID: "007051062584",
	names.APSoutheast3RegionID: "705930551576",
	names.CACentral1RegionID:   "675030665977",
	names.CNNorth1RegionID:     "122526803553",
	names.CNNorthwest1RegionID: "122578899357",
	names.EUCentral1RegionID:   "017069133835",
	names.EUNorth1RegionID:     "763603941244",
	names.EUSouth1RegionID:     "638885417683",
	names.EUWest1RegionID:      "131013547314",
	names.EUWest2RegionID:      "440796970383",
	names.EUWest3RegionID:      "341593696636",
	names.MESouth1RegionID:     "835444307964",
	names.SAEast1RegionID:      "520018980103",
	names.USEast1RegionID:      "205585389593",
	names.USEast2RegionID:      "211330385671",
	names.USGovWest1RegionID:   "598674086554",
	names.USWest1RegionID:      "740489534195",
	names.USWest2RegionID:      "306415355426",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/data-wrangler.json

var PrebuiltECRImageIDByRegion_dataWrangler = map[string]string{
	names.AFSouth1RegionID:     "143210264188",
	names.APEast1RegionID:      "707077482487",
	names.APNortheast1RegionID: "649008135260",
	names.APNortheast2RegionID: "131546521161",
	names.APNortheast3RegionID: "913387583493",
	names.APSouth1RegionID:     "089933028263",
	names.APSoutheast1RegionID: "119527597002",
	names.APSoutheast2RegionID: "422173101802",
	names.CACentral1RegionID:   "557239378090",
	names.CNNorth1RegionID:     "245909111842",
	names.CNNorthwest1RegionID: "249157047649",
	names.EUCentral1RegionID:   "024640144536",
	names.EUNorth1RegionID:     "054986407534",
	names.EUSouth1RegionID:     "488287956546",
	names.EUWest1RegionID:      "245179582081",
	names.EUWest2RegionID:      "894491911112",
	names.EUWest3RegionID:      "807237891255",
	names.ILCentral1RegionID:   "406833011540",
	names.MESouth1RegionID:     "376037874950",
	names.SAEast1RegionID:      "424196993095",
	names.USEast1RegionID:      "663277389841",
	names.USEast2RegionID:      "415577184552",
	names.USWest1RegionID:      "926135532090",
	names.USWest2RegionID:      "174368400705",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/debugger.json

var PrebuiltECRImageIDByRegion_debugger = map[string]string{
	names.AFSouth1RegionID:     "314341159256",
	names.APEast1RegionID:      "199566480951",
	names.APNortheast1RegionID: "430734990657",
	names.APNortheast2RegionID: "578805364391",
	names.APNortheast3RegionID: "479947661362",
	names.APSouth1RegionID:     "904829902805",
	names.APSoutheast1RegionID: "972752614525",
	names.APSoutheast2RegionID: "184798709955",
	names.CACentral1RegionID:   "519511493484",
	names.CNNorth1RegionID:     "618459771430",
	names.CNNorthwest1RegionID: "658757709296",
	names.EUCentral1RegionID:   "482524230118",
	names.EUNorth1RegionID:     "314864569078",
	names.EUSouth1RegionID:     "563282790590",
	names.EUWest1RegionID:      "929884845733",
	names.EUWest2RegionID:      "250201462417",
	names.EUWest3RegionID:      "447278800020",
	names.MESouth1RegionID:     "986000313247",
	names.SAEast1RegionID:      "818342061345",
	names.USEast1RegionID:      "503895931360",
	names.USEast2RegionID:      "915447279597",
	names.USGovWest1RegionID:   "515509971035",
	names.USWest1RegionID:      "685455198987",
	names.USWest2RegionID:      "895741380848",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/inferentia-mxnet.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/inferentia-pytorch.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/image-classification-neo.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/neo-mxnet.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/neo-pytorch.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/neo-tensorflow.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/xgboost-neo.json

var PrebuiltECRImageIDByRegion_inferentiaNeo = map[string]string{
	names.AFSouth1RegionID:     "774647643957",
	names.APEast1RegionID:      "110948597952",
	names.APNortheast1RegionID: "941853720454",
	names.APNortheast2RegionID: "151534178276",
	names.APNortheast3RegionID: "925152966179",
	names.APSouth1RegionID:     "763008648453",
	names.APSoutheast1RegionID: "324986816169",
	names.APSoutheast2RegionID: "355873309152",
	names.CACentral1RegionID:   "464438896020",
	names.CNNorth1RegionID:     "472730292857",
	names.CNNorthwest1RegionID: "474822919863",
	names.EUCentral1RegionID:   "746233611703",
	names.EUNorth1RegionID:     "601324751636",
	names.EUSouth1RegionID:     "966458181534",
	names.EUWest1RegionID:      "802834080501",
	names.EUWest2RegionID:      "205493899709",
	names.EUWest3RegionID:      "254080097072",
	names.ILCentral1RegionID:   "275950707576",
	names.MESouth1RegionID:     "836785723513",
	names.SAEast1RegionID:      "756306329178",
	names.USEast1RegionID:      "785573368785",
	names.USEast2RegionID:      "007439368137",
	names.USGovWest1RegionID:   "263933020539",
	names.USISOEast1RegionID:   "167761179201",
	names.USISOBEast1RegionID:  "406031935815",
	names.USWest1RegionID:      "710691900526",
	names.USWest2RegionID:      "301217895009",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/chainer.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/pytorch.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/coach-mxnet.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/mxnet.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/coach-tensorflow.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/tensorflow.json

var PrebuiltECRImageIDByRegion_SageMakerCustom = map[string]string{ // nosemgrep:ci.sagemaker-in-var-name
	names.AFSouth1RegionID:     "313743910680",
	names.APEast1RegionID:      "057415533634",
	names.APNortheast1RegionID: "520713654638",
	names.APNortheast2RegionID: "520713654638",
	names.APSouth1RegionID:     "520713654638",
	names.APSoutheast1RegionID: "520713654638",
	names.APSoutheast2RegionID: "520713654638",
	names.CACentral1RegionID:   "520713654638",
	names.CNNorth1RegionID:     "422961961927",
	names.CNNorthwest1RegionID: "423003514399",
	names.EUCentral1RegionID:   "520713654638",
	names.EUNorth1RegionID:     "520713654638",
	names.EUSouth1RegionID:     "048378556238",
	names.EUWest1RegionID:      "520713654638",
	names.EUWest2RegionID:      "520713654638",
	names.EUWest3RegionID:      "520713654638",
	names.MESouth1RegionID:     "724002660598",
	names.SAEast1RegionID:      "520713654638",
	names.USEast1RegionID:      "520713654638",
	names.USEast2RegionID:      "520713654638",
	names.USGovWest1RegionID:   "246785580436",
	names.USISOEast1RegionID:   "744548109606",
	names.USISOBEast1RegionID:  "453391408702",
	names.USWest1RegionID:      "520713654638",
	names.USWest2RegionID:      "520713654638",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/ray-pytorch.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/coach-tensorflow.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/vw.json

var PrebuiltECRImageIDByRegion_SageMakerRL = map[string]string{ // nosemgrep:ci.sagemaker-in-var-name
	names.APNortheast1RegionID: "462105765813",
	names.APNortheast2RegionID: "462105765813",
	names.APSouth1RegionID:     "462105765813",
	names.APSoutheast1RegionID: "462105765813",
	names.APSoutheast2RegionID: "462105765813",
	names.CACentral1RegionID:   "462105765813",
	names.EUCentral1RegionID:   "462105765813",
	names.EUWest1RegionID:      "462105765813",
	names.EUWest2RegionID:      "462105765813",
	names.USEast1RegionID:      "462105765813",
	names.USEast2RegionID:      "462105765813",
	names.USWest1RegionID:      "462105765813",
	names.USWest2RegionID:      "462105765813",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/spark.json

var PrebuiltECRImageIDByRegion_spark = map[string]string{
	names.AFSouth1RegionID:     "309385258863",
	names.APEast1RegionID:      "732049463269",
	names.APNortheast1RegionID: "411782140378",
	names.APNortheast2RegionID: "860869212795",
	names.APNortheast3RegionID: "102471314380",
	names.APSouth1RegionID:     "105495057255",
	names.APSouth2RegionID:     "873151114052",
	names.APSoutheast1RegionID: "759080221371",
	names.APSoutheast2RegionID: "440695851116",
	names.APSoutheast3RegionID: "800295151634",
	names.APSoutheast4RegionID: "819679513684",
	names.CACentral1RegionID:   "446299261295",
	names.CAWest1RegionID:      "000907499111",
	names.CNNorth1RegionID:     "671472414489",
	names.CNNorthwest1RegionID: "844356804704",
	names.EUCentral1RegionID:   "906073651304",
	names.EUCentral2RegionID:   "142351485170",
	names.EUNorth1RegionID:     "330188676905",
	names.EUSouth1RegionID:     "753923664805",
	names.EUSouth2RegionID:     "833944533722",
	names.EUWest1RegionID:      "571004829621",
	names.EUWest2RegionID:      "836651553127",
	names.EUWest3RegionID:      "136845547031",
	names.ILCentral1RegionID:   "408426139102",
	names.MECentral1RegionID:   "395420993607",
	names.MESouth1RegionID:     "750251592176",
	names.SAEast1RegionID:      "737130764395",
	names.USEast1RegionID:      "173754725891",
	names.USEast2RegionID:      "314815235551",
	names.USGovEast1RegionID:   "260923028637",
	names.USGovWest1RegionID:   "271483468897",
	names.USWest1RegionID:      "667973535471",
	names.USWest2RegionID:      "153931337802",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/sagemaker-base-python.json

var PrebuiltECRImageIDByRegion_SageMakerBasePython = map[string]string{ // nosemgrep:ci.sagemaker-in-var-name
	names.AFSouth1RegionID:     "559312083959",
	names.APEast1RegionID:      "493642496378",
	names.APNortheast1RegionID: "102112518831",
	names.APNortheast2RegionID: "806072073708",
	names.APNortheast3RegionID: "792733760839",
	names.APSouth1RegionID:     "394103062818",
	names.APSoutheast1RegionID: "492261229750",
	names.APSoutheast2RegionID: "452832661640",
	names.APSoutheast3RegionID: "276181064229",
	names.CACentral1RegionID:   "310906938811",
	names.CNNorth1RegionID:     "390048526115",
	names.CNNorthwest1RegionID: "390780980154",
	names.EUCentral1RegionID:   "936697816551",
	names.EUNorth1RegionID:     "243637512696",
	names.EUSouth1RegionID:     "592751261982",
	names.EUSouth2RegionID:     "127363102723",
	names.EUWest1RegionID:      "470317259841",
	names.EUWest2RegionID:      "712779665605",
	names.EUWest3RegionID:      "615547856133",
	names.ILCentral1RegionID:   "380164790875",
	names.MECentral1RegionID:   "103105715889",
	names.MESouth1RegionID:     "117516905037",
	names.SAEast1RegionID:      "782484402741",
	names.USEast1RegionID:      "081325390199",
	names.USEast2RegionID:      "429704687514",
	names.USGovEast1RegionID:   "107072934176",
	names.USGovWest1RegionID:   "107173498710",
	names.USWest1RegionID:      "742091327244",
	names.USWest2RegionID:      "236514542706",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/sagemaker-geospatial.json

var PrebuiltECRImageIDByRegion_SageMakerGeospatial = map[string]string{ // nosemgrep:ci.sagemaker-in-var-name
	names.USWest2RegionID: "081189585635",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/forecasting-deepar.json

var PrebuiltECRImageIDByRegion_deepAR = map[string]string{
	names.AFSouth1RegionID:     "455444449433",
	names.APEast1RegionID:      "286214385809",
	names.APNortheast1RegionID: "633353088612",
	names.APNortheast2RegionID: "204372634319",
	names.APNortheast3RegionID: "867004704886",
	names.APSouth1RegionID:     "991648021394",
	names.APSouth2RegionID:     "628508329040",
	names.APSoutheast1RegionID: "475088953585",
	names.APSoutheast2RegionID: "514117268639",
	names.APSoutheast3RegionID: "951798379941",
	names.APSoutheast4RegionID: "106583098589",
	names.CACentral1RegionID:   "469771592824",
	names.CAWest1RegionID:      "190319476487",
	names.CNNorth1RegionID:     "390948362332",
	names.CNNorthwest1RegionID: "387376663083",
	names.EUCentral1RegionID:   "495149712605",
	names.EUCentral2RegionID:   "680994064768",
	names.EUNorth1RegionID:     "669576153137",
	names.EUSouth1RegionID:     "257386234256",
	names.EUSouth2RegionID:     "104374241257",
	names.EUWest1RegionID:      "224300973850",
	names.EUWest2RegionID:      "644912444149",
	names.EUWest3RegionID:      "749696950732",
	names.ILCentral1RegionID:   "898809789911",
	names.MECentral1RegionID:   "272398656194",
	names.MESouth1RegionID:     "249704162688",
	names.SAEast1RegionID:      "855470959533",
	names.USEast1RegionID:      "522234722520",
	names.USEast2RegionID:      "566113047672",
	names.USGovEast1RegionID:   "237065988967",
	names.USGovWest1RegionID:   "226302683700",
	names.USISOEast1RegionID:   "490574956308",
	names.USISOBEast1RegionID:  "765400339828",
	names.USWest1RegionID:      "632365934929",
	names.USWest2RegionID:      "156387875391",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/factorization-machines.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/ipinsights.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/linear-learner.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/ntm.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/object2vec.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/pca.json

var PrebuiltECRImageIDByRegion_factorMachines = map[string]string{
	names.AFSouth1RegionID:     "455444449433",
	names.APEast1RegionID:      "286214385809",
	names.APNortheast1RegionID: "351501993468",
	names.APNortheast2RegionID: "835164637446",
	names.APNortheast3RegionID: "867004704886",
	names.APSouth1RegionID:     "991648021394",
	names.APSouth2RegionID:     "628508329040",
	names.APSoutheast1RegionID: "475088953585",
	names.APSoutheast2RegionID: "712309505854",
	names.APSoutheast3RegionID: "951798379941",
	names.APSoutheast4RegionID: "106583098589",
	names.CACentral1RegionID:   "469771592824",
	names.CAWest1RegionID:      "190319476487",
	names.CNNorth1RegionID:     "390948362332",
	names.CNNorthwest1RegionID: "387376663083",
	names.EUCentral1RegionID:   "664544806723",
	names.EUCentral2RegionID:   "680994064768",
	names.EUNorth1RegionID:     "669576153137",
	names.EUSouth1RegionID:     "257386234256",
	names.EUSouth2RegionID:     "104374241257",
	names.EUWest1RegionID:      "438346466558",
	names.EUWest2RegionID:      "644912444149",
	names.EUWest3RegionID:      "749696950732",
	names.ILCentral1RegionID:   "898809789911",
	names.MECentral1RegionID:   "272398656194",
	names.MESouth1RegionID:     "249704162688",
	names.SAEast1RegionID:      "855470959533",
	names.USEast1RegionID:      "382416733822",
	names.USEast2RegionID:      "404615174143",
	names.USGovEast1RegionID:   "237065988967",
	names.USGovWest1RegionID:   "226302683700",
	names.USISOEast1RegionID:   "490574956308",
	names.USISOBEast1RegionID:  "765400339828",
	names.USWest1RegionID:      "632365934929",
	names.USWest2RegionID:      "174872318107",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/lda.json

var PrebuiltECRImageIDByRegion_lda = map[string]string{
	names.APNortheast1RegionID: "258307448986",
	names.APNortheast2RegionID: "293181348795",
	names.APSouth1RegionID:     "991648021394",
	names.APSoutheast1RegionID: "475088953585",
	names.APSoutheast2RegionID: "297031611018",
	names.CACentral1RegionID:   "469771592824",
	names.EUCentral1RegionID:   "353608530281",
	names.EUWest1RegionID:      "999678624901",
	names.EUWest2RegionID:      "644912444149",
	names.USEast1RegionID:      "766337827248",
	names.USEast2RegionID:      "999911452149",
	names.USGovWest1RegionID:   "226302683700",
	names.USISOEast1RegionID:   "490574956308",
	names.USISOBEast1RegionID:  "765400339828",
	names.USWest1RegionID:      "632365934929",
	names.USWest2RegionID:      "266724342769",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/xgboost.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/huggingface-tei.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/huggingface-tei-cpu.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/sparkml-serving.json
// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/sklearn.json

var PrebuiltECRImageIDByRegion_XGBoost = map[string]string{
	names.AFSouth1RegionID:     "510948584623",
	names.APEast1RegionID:      "651117190479",
	names.APNortheast1RegionID: "354813040037",
	names.APNortheast2RegionID: "366743142698",
	names.APNortheast3RegionID: "867004704886",
	names.APSouth1RegionID:     "720646828776",
	names.APSouth2RegionID:     "628508329040",
	names.APSoutheast1RegionID: "121021644041",
	names.APSoutheast2RegionID: "783357654285",
	names.APSoutheast3RegionID: "951798379941",
	names.APSoutheast4RegionID: "106583098589",
	names.CACentral1RegionID:   "341280168497",
	names.CAWest1RegionID:      "190319476487",
	names.CNNorth1RegionID:     "450853457545",
	names.CNNorthwest1RegionID: "451049120500",
	names.EUCentral1RegionID:   "492215442770",
	names.EUCentral2RegionID:   "680994064768",
	names.EUNorth1RegionID:     "662702820516",
	names.EUSouth1RegionID:     "978288397137",
	names.EUSouth2RegionID:     "104374241257",
	names.EUWest1RegionID:      "141502667606",
	names.EUWest2RegionID:      "764974769150",
	names.EUWest3RegionID:      "659782779980",
	names.ILCentral1RegionID:   "898809789911",
	names.MECentral1RegionID:   "272398656194",
	names.MESouth1RegionID:     "801668240914",
	names.SAEast1RegionID:      "737474898029",
	names.USEast1RegionID:      "683313688378",
	names.USEast2RegionID:      "257758044811",
	names.USGovEast1RegionID:   "237065988967",
	names.USGovWest1RegionID:   "414596584902",
	names.USISOEast1RegionID:   "833128469047",
	names.USISOBEast1RegionID:  "281123927165",
	names.USWest1RegionID:      "746614075791",
	names.USWest2RegionID:      "246618743249",
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

var PrebuiltECRImageIDByRegion_deepLearning = map[string]string{
	names.AFSouth1RegionID:     "626614931356",
	names.APEast1RegionID:      "871362719292",
	names.APNortheast1RegionID: "763104351884",
	names.APNortheast2RegionID: "763104351884",
	names.APNortheast3RegionID: "364406365360",
	names.APSouth1RegionID:     "763104351884",
	names.APSouth2RegionID:     "772153158452",
	names.APSoutheast1RegionID: "763104351884",
	names.APSoutheast2RegionID: "763104351884",
	names.APSoutheast3RegionID: "907027046896",
	names.APSoutheast4RegionID: "457447274322",
	names.CACentral1RegionID:   "763104351884",
	names.CAWest1RegionID:      "204538143572",
	names.CNNorth1RegionID:     "727897471807",
	names.CNNorthwest1RegionID: "727897471807",
	names.EUCentral1RegionID:   "763104351884",
	names.EUCentral2RegionID:   "380420809688",
	names.EUNorth1RegionID:     "763104351884",
	names.EUWest1RegionID:      "763104351884",
	names.EUWest2RegionID:      "763104351884",
	names.EUWest3RegionID:      "763104351884",
	names.EUSouth1RegionID:     "692866216735",
	names.EUSouth2RegionID:     "503227376785",
	names.ILCentral1RegionID:   "780543022126",
	names.MECentral1RegionID:   "914824155844",
	names.MESouth1RegionID:     "217643126080",
	names.SAEast1RegionID:      "763104351884",
	names.USEast1RegionID:      "763104351884",
	names.USEast2RegionID:      "763104351884",
	names.USWest1RegionID:      "763104351884",
	names.USWest2RegionID:      "763104351884",
	names.USGovEast1RegionID:   "446045086412",
	names.USGovWest1RegionID:   "442386744353",
	names.USISOEast1RegionID:   "886529160074",
	names.USISOBEast1RegionID:  "094389454867",
}

// https://github.com/aws/sagemaker-python-sdk/blob/master/src/sagemaker/image_uri_config/model-monitor.json

var PrebuiltECRImageIDByRegion_modelMonitor = map[string]string{
	names.AFSouth1RegionID:     "875698925577",
	names.APEast1RegionID:      "001633400207",
	names.APNortheast1RegionID: "574779866223",
	names.APNortheast2RegionID: "709848358524",
	names.APNortheast3RegionID: "990339680094",
	names.APSouth1RegionID:     "126357580389",
	names.APSoutheast1RegionID: "245545462676",
	names.APSoutheast2RegionID: "563025443158",
	names.APSoutheast3RegionID: "669540362728",
	names.CACentral1RegionID:   "536280801234",
	names.CNNorth1RegionID:     "453000072557",
	names.CNNorthwest1RegionID: "453252182341",
	names.EUCentral1RegionID:   "048819808253",
	names.EUNorth1RegionID:     "895015795356",
	names.EUSouth1RegionID:     "933208885752",
	names.EUSouth2RegionID:     "437450045455",
	names.EUWest1RegionID:      "468650794304",
	names.EUWest2RegionID:      "749857270468",
	names.EUWest3RegionID:      "680080141114",
	names.ILCentral1RegionID:   "843974653677",
	names.MECentral1RegionID:   "588750061953",
	names.MESouth1RegionID:     "607024016150",
	names.SAEast1RegionID:      "539772159869",
	names.USEast1RegionID:      "156813124566",
	names.USEast2RegionID:      "777275614652",
	names.USWest1RegionID:      "890145073186",
	names.USWest2RegionID:      "159807026194",
}

// @SDKDataSource("aws_sagemaker_prebuilt_ecr_image")
func DataSourcePrebuiltECRImage() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePrebuiltECRImageRead,
		Schema: map[string]*schema.Schema{
			names.AttrRepositoryName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					RepositoryAutoGluonTraining,
					RepositoryAutoGluonInference,
					RepositoryBlazingText,
					RepositoryChainer,
					RepositoryClarify,
					RepositoryDJLDeepSpeed,
					RepositoryDataWrangler,
					RepositoryDebugger,
					RepositoryDeepARForecasting,
					RepositoryFactorizationMachines,
					RepositoryHuggingFaceTensorFlowTraining,
					RepositoryHuggingFacePyTorchTraining,
					RepositoryHuggingFacePyTorchTrainingNeuronX,
					RepositoryHuggingFacePyTorchTrainingCompiler,
					RepositoryHuggingFaceTensorFlowTrainingCompiler,
					RepositoryHuggingFaceTensorFlowInference,
					RepositoryHuggingFacePyTorchInference,
					RepositoryHuggingFacePyTorchInferenceNeuron,
					RepositoryHuggingFacePyTorchInferenceNeuronX,
					RepositoryHuggingFacePyTorchTGIInference,
					RepositoryHuggingFaceTEI,
					RepositoryHuggingFaceTEICPU,
					RepositoryIPInsights,
					RepositoryImageClassification,
					RepositoryInferentiaMXNet,
					RepositoryInferentiaPyTorch,
					RepositoryKMeans,
					RepositoryKNearestNeighbor,
					RepositoryLDA,
					RepositoryLinearLearner,
					RepositoryModelMonitor,
					RepositoryMXNetTraining,
					RepositoryMXNetInference,
					RepositoryMXNetInferenceEIA,
					RepositoryMXNetCoach,
					RepositoryNeuralTopicModel,
					RepositoryNeoImageClassification,
					RepositoryNeoMXNet,
					RepositoryNeoPyTorch,
					RepositoryNeoTensorflow,
					RepositoryNeoXGBoost,
					RepositoryNVIDIATritonInference,
					RepositoryObjectDetection,
					RepositoryObject2Vec,
					RepositoryPCA,
					RepositoryPyTorchTraining,
					RepositoryPyTorchTrainingNeuronX,
					RepositoryPyTorchTrainingCompiler,
					RepositoryPyTorchInference,
					RepositoryPyTorchInferenceEIA,
					RepositoryPyTorchInferenceGraviton,
					RepositoryPyTorchInferenceNeuronX,
					RepositoryRandomCutForest,
					RepositoryRLRayPyTorch,
					RepositoryRLCoachPyTorch,
					RepositorySageMakerBasePython,
					RepositorySageMakerGeospatial,
					RepositorySageMakerMXNet,
					RepositorySageMakerMXNetServing,
					RepositorySageMakerMXNetEIA,
					RepositorySageMakerMXNetServingEIA,
					RepositorySageMakerPyTorch,
					RepositorySageMakerTensorFlow,
					RepositorySageMakerTensorFlowEIA,
					RepositoryScikitLearn,
					RepositorySemanticSegmentation,
					RepositorySeq2Seq,
					RepositorySpark,
					RepositorySparkML,
					RepositoryTensorFlowTraining,
					RepositoryTensorFlowInference,
					RepositoryTensorFlowInferenceEIA,
					RepositoryTensorFlowInferenceGraviton,
					RepositoryTensorFlowServing,
					RepositoryTensorFlowServingEIA,
					RepositorySageMakerTensorFlowScriptMode,
					RepositoryTensorflowCoach,
					RepositoryTensorflowInferentia,
					RepositoryStabilityAI,
					RepositoryVW,
					RepositoryXGBoost,
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

func dataSourcePrebuiltECRImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	region := meta.(*conns.AWSClient).Region
	if v, ok := d.GetOk(names.AttrRegion); ok {
		region = v.(string)
	}

	suffix := meta.(*conns.AWSClient).DNSSuffix(ctx)
	if v, ok := d.GetOk("dns_suffix"); ok {
		suffix = v.(string)
	}

	repo := d.Get(names.AttrRepositoryName).(string)

	id := ""
	switch repo {
	case RepositoryBlazingText,
		RepositoryImageClassification,
		RepositoryObjectDetection,
		RepositorySemanticSegmentation,
		RepositorySeq2Seq:
		id = PrebuiltECRImageIDByRegion_blazing[region]
	case RepositoryClarify:
		id = PrebuiltECRImageIDByRegion_clarify[region]
	case RepositoryDataWrangler:
		id = PrebuiltECRImageIDByRegion_dataWrangler[region]
	case RepositoryDebugger:
		id = PrebuiltECRImageIDByRegion_debugger[region]
	case RepositoryDeepARForecasting:
		id = PrebuiltECRImageIDByRegion_deepAR[region]
	case RepositoryInferentiaMXNet,
		RepositoryInferentiaPyTorch,
		RepositoryMXNetCoach,
		RepositoryNeoImageClassification,
		RepositoryNeoMXNet,
		RepositoryNeoPyTorch,
		RepositoryNeoTensorflow,
		RepositoryNeoXGBoost,
		RepositoryTensorflowInferentia:
		id = PrebuiltECRImageIDByRegion_inferentiaNeo[region]
	case RepositoryLDA:
		id = PrebuiltECRImageIDByRegion_lda[region]
	case RepositoryModelMonitor:
		id = PrebuiltECRImageIDByRegion_modelMonitor[region]
	case RepositoryXGBoost,
		RepositoryScikitLearn,
		RepositorySparkML,
		RepositoryHuggingFaceTEI,
		RepositoryHuggingFaceTEICPU:
		id = PrebuiltECRImageIDByRegion_XGBoost[region]
	case RepositoryChainer,
		RepositorySageMakerMXNet,
		RepositorySageMakerMXNetServing,
		RepositorySageMakerMXNetEIA,
		RepositorySageMakerMXNetServingEIA,
		RepositorySageMakerPyTorch,
		RepositorySageMakerTensorFlow,
		RepositorySageMakerTensorFlowEIA,
		RepositorySageMakerTensorFlowScriptMode,
		RepositoryTensorflowCoach,
		RepositoryTensorFlowServing,
		RepositoryTensorFlowServingEIA:
		id = PrebuiltECRImageIDByRegion_SageMakerCustom[region]
	case RepositoryAutoGluonTraining,
		RepositoryAutoGluonInference,
		RepositoryDJLDeepSpeed,
		RepositoryHuggingFaceTensorFlowTraining,
		RepositoryHuggingFacePyTorchTraining,
		RepositoryHuggingFacePyTorchTrainingNeuronX,
		RepositoryHuggingFacePyTorchTrainingCompiler,
		RepositoryHuggingFaceTensorFlowTrainingCompiler,
		RepositoryHuggingFaceTensorFlowInference,
		RepositoryHuggingFacePyTorchInference,
		RepositoryHuggingFacePyTorchInferenceNeuron,
		RepositoryHuggingFacePyTorchInferenceNeuronX,
		RepositoryHuggingFacePyTorchTGIInference,
		RepositoryMXNetTraining,
		RepositoryMXNetInference,
		RepositoryMXNetInferenceEIA,
		RepositoryPyTorchTraining,
		RepositoryPyTorchTrainingNeuronX,
		RepositoryPyTorchTrainingCompiler,
		RepositoryPyTorchInference,
		RepositoryPyTorchInferenceEIA,
		RepositoryPyTorchInferenceGraviton,
		RepositoryPyTorchInferenceNeuronX,
		RepositoryStabilityAI,
		RepositoryTensorFlowTraining,
		RepositoryTensorFlowInference,
		RepositoryTensorFlowInferenceEIA,
		RepositoryTensorFlowInferenceGraviton,
		RepositoryNVIDIATritonInference:
		id = PrebuiltECRImageIDByRegion_deepLearning[region]
	case RepositoryRLRayPyTorch,
		RepositoryRLCoachPyTorch,
		RepositoryVW:
		id = PrebuiltECRImageIDByRegion_SageMakerRL[region]
	case RepositorySageMakerBasePython:
		id = PrebuiltECRImageIDByRegion_SageMakerBasePython[region]
	case RepositorySageMakerGeospatial:
		id = PrebuiltECRImageIDByRegion_SageMakerGeospatial[region]
	case RepositorySpark:
		id = PrebuiltECRImageIDByRegion_spark[region]
	default:
		id = PrebuiltECRImageIDByRegion_factorMachines[region]
	}

	if id == "" {
		return sdkdiag.AppendErrorf(diags, "no registry ID available for region (%s) and repository (%s)", region, repo)
	}

	d.SetId(id)
	d.Set("registry_id", id)
	d.Set("registry_path", PrebuiltECRImageCreatePath(id, region, suffix, repo, d.Get("image_tag").(string)))
	return diags
}

func PrebuiltECRImageCreatePath(id, region, suffix, repo, imageTag string) string {
	return fmt.Sprintf("%s.dkr.ecr.%s.%s/%s:%s", id, region, suffix, repo, imageTag)
}
