//go:build generate
// +build generate

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	defaultFilename = "list_pages_gen.go"
)

var (
	listOps   = flag.String("ListOps", "", "ListOps")
	paginator = flag.String("Paginator", "NextToken", "name of the pagination token field")
	export    = flag.Bool("Export", false, "whether to export the list functions")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags] [<generated-lister-file>]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

type TemplateData struct {
	AWSService     string
	ServicePackage string

	ListOps   string
	Paginator string
}

func main() {
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	filename := defaultFilename
	if args := flag.Args(); len(args) > 0 {
		filename = args[0]
	}

	wd, err := os.Getwd()

	if err != nil {
		log.Fatalf("unable to get working directory: %s", err)
	}

	servicePackage := filepath.Base(wd)
	awsService, err := awsServiceName(servicePackage)

	if err != nil {
		log.Fatalf("encountered: %s", err)
	}

	templateData := TemplateData{
		AWSService:     awsService,
		ServicePackage: servicePackage,
		ListOps:        *listOps,
		Paginator:      *paginator,
	}

	functions := strings.Split(templateData.ListOps, ",")
	sort.Strings(functions)

	g := Generator{
		paginator: templateData.Paginator,
		tmpl:      template.Must(template.New("function").Parse(functionTemplate)),
	}

	sourcePackage := fmt.Sprintf("github.com/aws/aws-sdk-go/service/%s", templateData.AWSService)
	g.parsePackage(sourcePackage)

	g.printHeader(HeaderInfo{
		Parameters:         strings.Join(os.Args[1:], " "),
		DestinationPackage: templateData.ServicePackage,
		SourcePackage:      sourcePackage,
	})

	for _, functionName := range functions {
		g.generateFunction(functionName, *export)
	}

	src := g.format()

	err = os.WriteFile(filename, src, 0644)
	if err != nil {
		log.Fatalf("error writing output: %s", err)
	}
}

type HeaderInfo struct {
	Parameters         string
	DestinationPackage string
	SourcePackage      string
}

type Generator struct {
	buf       bytes.Buffer
	pkg       *Package
	tmpl      *template.Template
	paginator string
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

type PackageFile struct {
	file *ast.File
}

type Package struct {
	name  string
	files []*PackageFile
}

func (g *Generator) printHeader(headerInfo HeaderInfo) {
	header := template.Must(template.New("header").Parse(headerTemplate))
	err := header.Execute(&g.buf, headerInfo)
	if err != nil {
		log.Fatalf("error writing header: %s", err)
	}
}

func (g *Generator) parsePackage(sourcePackage string) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax,
	}
	pkgs, err := packages.Load(cfg, sourcePackage)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	g.addPackage(pkgs[0])
}

func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		name:  pkg.Name,
		files: make([]*PackageFile, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		g.pkg.files[i] = &PackageFile{
			file: file,
		}
	}
}

type FuncSpec struct {
	Name       string
	AWSName    string
	RecvType   string
	ParamType  string
	ResultType string
	Paginator  string
}

func (g *Generator) generateFunction(functionName string, export bool) {
	var function *ast.FuncDecl

	// TODO: check if a Pages() function has been defined
	for _, file := range g.pkg.files {
		if file.file != nil {
			for _, decl := range file.file.Decls {
				if funcDecl, ok := decl.(*ast.FuncDecl); ok {
					if funcDecl.Name.Name == functionName {
						function = funcDecl
						break
					}
				}
			}
			if function != nil {
				break
			}
		}
	}

	if function == nil {
		log.Fatalf("function \"%s\" not found", functionName)
	}

	funcName := function.Name.Name

	if !export {
		funcName = fmt.Sprintf("%s%s", strings.ToLower(funcName[0:1]), funcName[1:])
	}

	funcSpec := FuncSpec{
		Name:       fixSomeInitialisms(funcName),
		AWSName:    function.Name.Name,
		RecvType:   g.expandTypeField(function.Recv),
		ParamType:  g.expandTypeField(function.Type.Params),  // Assumes there is a single input parameter
		ResultType: g.expandTypeField(function.Type.Results), // Assumes we can take the first return parameter
		Paginator:  g.paginator,
	}

	err := g.tmpl.Execute(&g.buf, funcSpec)
	if err != nil {
		log.Fatalf("error writing function \"%s\": %s", functionName, err)
	}
}

func (g *Generator) expandTypeField(field *ast.FieldList) string {
	typeValue := field.List[0].Type
	if star, ok := typeValue.(*ast.StarExpr); ok {
		return fmt.Sprintf("*%s", g.expandTypeExpr(star.X))
	}

	log.Fatalf("Unexpected type expression: (%[1]T) %[1]v", typeValue)
	return ""
}

func (g *Generator) expandTypeExpr(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return fmt.Sprintf("%s.%s", g.pkg.name, ident.Name)
	}

	log.Fatalf("Unexpected expression: (%[1]T) %[1]v", expr)
	return ""
}

const headerTemplate = `// Code generated by "internal/generate/listpages/main.go {{ .Parameters }}"; DO NOT EDIT.

package {{ .DestinationPackage }}

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"{{ .SourcePackage }}"
)
`

const functionTemplate = `

func {{ .Name }}Pages(conn {{ .RecvType }}, input {{ .ParamType }}, fn func({{ .ResultType }}, bool) bool) error {
	return {{ .Name }}PagesWithContext(context.Background(), conn, input, fn)
}

func {{ .Name }}PagesWithContext(ctx context.Context, conn {{ .RecvType }}, input {{ .ParamType }}, fn func({{ .ResultType }}, bool) bool) error {
	for {
		output, err := conn.{{ .AWSName }}WithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.{{ .Paginator }}) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.{{ .Paginator }} = output.{{ .Paginator }}
	}
	return nil
}
`

func (g *Generator) format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}

func awsServiceName(s string) (string, error) {
	s = strings.ToLower(s)

	if _, ok := awsServiceNames[s]; ok {
		return s, nil
	}

	switch s {
	case "amp":
		return "prometheusservice", nil
	case "cloudcontrol":
		return "cloudcontrolapi", nil
	case "cognitoidp":
		return "cognitoidentityprovider", nil
	case "dms":
		return "databasemigrationservice", nil
	case "ds":
		return "directoryservice", nil
	case "events":
		return "eventbridge", nil
	case "lexmodels":
		return "lexmodelbuildingservice", nil
	case "serverlessrepo":
		return "serverlessapplicationrepository", nil
	}

	if _, ok := awsServiceNames[fmt.Sprintf("%sservice", s)]; ok {
		return fmt.Sprintf("%sservice", s), nil
	}

	return "", fmt.Errorf("unable to find AWS service name for %s", s)
}

func awsServiceNameUpper(s string) (string, error) {
	s = strings.ToLower(s)

	if v, ok := awsServiceNames[s]; ok {
		return v, nil
	}

	switch s {
	case "amp":
		return awsServiceNames["prometheusservice"], nil
	case "appautoscaling":
		return awsServiceNames["applicationautoscaling"], nil
	case "cloudcontrol":
		return awsServiceNames["cloudcontrolapi"], nil
	case "cognitoidp":
		return awsServiceNames["cognitoidentityprovider"], nil
	case "dms":
		return awsServiceNames["databasemigrationservice"], nil
	case "ds":
		return awsServiceNames["directoryservice"], nil
	case "events":
		return awsServiceNames["eventbridge"], nil
	case "lexmodels":
		return awsServiceNames["lexmodelbuildingservice"], nil
	case "serverlessrepo":
		return awsServiceNames["serverlessapplicationrepository"], nil
	}

	if v, ok := awsServiceNames[fmt.Sprintf("%sservice", s)]; ok {
		return v, nil
	}

	return "", fmt.Errorf("unable to find AWS service name for %s", s)
}

//awsServiceNames provides correct names and capitalization as used by AWS in client var
var awsServiceNames map[string]string

func init() {
	awsServiceNames = make(map[string]string)

	awsServiceNames["accessanalyzer"] = "AccessAnalyzer"
	awsServiceNames["acm"] = "ACM"
	awsServiceNames["acmpca"] = "ACMPCA"
	awsServiceNames["alexaforbusiness"] = "AlexaForBusiness"
	awsServiceNames["amplify"] = "Amplify"
	awsServiceNames["amplifybackend"] = "AmplifyBackend"
	awsServiceNames["apigateway"] = "APIGateway"
	awsServiceNames["apigatewaymanagement"] = "APIGatewayManagement"
	awsServiceNames["apigatewayv2"] = "APIGatewayV2"
	awsServiceNames["appconfig"] = "AppConfig"
	awsServiceNames["appflow"] = "AppFlow"
	awsServiceNames["appintegrations"] = "AppIntegrations"
	awsServiceNames["applicationautoscaling"] = "ApplicationAutoScaling"
	awsServiceNames["applicationcostprofiler"] = "ApplicationCostProfiler"
	awsServiceNames["applicationdiscovery"] = "ApplicationDiscovery"
	awsServiceNames["applicationinsights"] = "ApplicationInsights"
	awsServiceNames["appmesh"] = "AppMesh"
	awsServiceNames["appregistry"] = "AppRegistry"
	awsServiceNames["apprunner"] = "AppRunner"
	awsServiceNames["appstream"] = "AppStream"
	awsServiceNames["appsync"] = "AppSync"
	awsServiceNames["athena"] = "Athena"
	awsServiceNames["auditmanager"] = "AuditManager"
	awsServiceNames["augmentedairuntime"] = "AugmentedAiruntime"
	awsServiceNames["autoscaling"] = "AutoScaling"
	awsServiceNames["autoscalingplans"] = "AutoScalingPlans"
	awsServiceNames["backup"] = "Backup"
	awsServiceNames["batch"] = "Batch"
	awsServiceNames["braket"] = "Braket"
	awsServiceNames["budgets"] = "Budgets"
	awsServiceNames["chime"] = "Chime"
	awsServiceNames["cloud9"] = "Cloud9"
	awsServiceNames["cloudcontrolapi"] = "CloudControlApi"
	awsServiceNames["clouddirectory"] = "CloudDirectory"
	awsServiceNames["cloudformation"] = "CloudFormation"
	awsServiceNames["cloudfront"] = "CloudFront"
	awsServiceNames["cloudhsm"] = "CloudHSM"
	awsServiceNames["cloudhsmv2"] = "CloudHSMV2"
	awsServiceNames["cloudsearch"] = "CloudSearch"
	awsServiceNames["cloudsearchdomain"] = "CloudSearchDomain"
	awsServiceNames["cloudtrail"] = "CloudTrail"
	awsServiceNames["cloudwatch"] = "CloudWatch"
	awsServiceNames["cloudwatchlogs"] = "CloudWatchLogs"
	awsServiceNames["codeartifact"] = "CodeArtifact"
	awsServiceNames["codebuild"] = "CodeBuild"
	awsServiceNames["codecommit"] = "CodeCommit"
	awsServiceNames["codedeploy"] = "CodeDeploy"
	awsServiceNames["codeguruprofiler"] = "CodeGuruProfiler"
	awsServiceNames["codegurureviewer"] = "CodeGuruReviewer"
	awsServiceNames["codepipeline"] = "CodePipeline"
	awsServiceNames["codestar"] = "CodeStar"
	awsServiceNames["codestarconnections"] = "CodeStarConnections"
	awsServiceNames["codestarnotifications"] = "CodeStarNotifications"
	awsServiceNames["cognitoidentity"] = "CognitoIdentity"
	awsServiceNames["cognitoidentityprovider"] = "CognitoIdentityProvider"
	awsServiceNames["cognitosync"] = "CognitoSync"
	awsServiceNames["comprehend"] = "Comprehend"
	awsServiceNames["comprehendmedical"] = "ComprehendMedical"
	awsServiceNames["computeoptimizer"] = "ComputeOptimizer"
	awsServiceNames["configservice"] = "ConfigService"
	awsServiceNames["connect"] = "Connect"
	awsServiceNames["connectcontactlens"] = "ConnectContactLens"
	awsServiceNames["connectparticipant"] = "ConnectParticipant"
	awsServiceNames["costexplorer"] = "CostExplorer"
	awsServiceNames["cur"] = "CUR"
	awsServiceNames["customerprofiles"] = "CustomerProfiles"
	awsServiceNames["databasemigrationservice"] = "DatabaseMigrationService"
	awsServiceNames["dataexchange"] = "DataExchange"
	awsServiceNames["datapipeline"] = "DataPipeline"
	awsServiceNames["datasync"] = "DataSync"
	awsServiceNames["dax"] = "DAX"
	awsServiceNames["detective"] = "Detective"
	awsServiceNames["devicefarm"] = "DeviceFarm"
	awsServiceNames["devopsguru"] = "DevOpsGuru"
	awsServiceNames["directconnect"] = "DirectConnect"
	awsServiceNames["directoryservice"] = "DirectoryService"
	awsServiceNames["dlm"] = "DLM"
	awsServiceNames["docdb"] = "DocDB"
	awsServiceNames["dynamodb"] = "DynamoDB"
	awsServiceNames["dynamodbattribute"] = "DynamoDBAttribute"
	awsServiceNames["dynamodbstreams"] = "DynamoDBStreams"
	awsServiceNames["ec2"] = "EC2"
	awsServiceNames["ec2instanceconnect"] = "EC2InstanceConnect"
	awsServiceNames["ecr"] = "ECR"
	awsServiceNames["ecrpublic"] = "ECRPublic"
	awsServiceNames["ecs"] = "ECS"
	awsServiceNames["efs"] = "EFS"
	awsServiceNames["eks"] = "EKS"
	awsServiceNames["elasticache"] = "ElastiCache"
	awsServiceNames["elasticbeanstalk"] = "ElasticBeanstalk"
	awsServiceNames["elasticinference"] = "ElasticInference"
	awsServiceNames["elasticsearchservice"] = "ElasticsearchService"
	awsServiceNames["elastictranscoder"] = "ElasticTranscoder"
	awsServiceNames["elb"] = "ELB"
	awsServiceNames["elbv2"] = "ELBV2"
	awsServiceNames["emr"] = "EMR"
	awsServiceNames["emrcontainers"] = "EMRContainers"
	awsServiceNames["eventbridge"] = "EventBridge"
	awsServiceNames["expression"] = "Expression"
	awsServiceNames["finspace"] = "FinSpace"
	awsServiceNames["finspacedata"] = "FinSpaceData"
	awsServiceNames["firehose"] = "Firehose"
	awsServiceNames["fis"] = "FIS"
	awsServiceNames["fms"] = "FMS"
	awsServiceNames["forecast"] = "Forecast"
	awsServiceNames["forecastquery"] = "ForecastQuery"
	awsServiceNames["frauddetector"] = "FraudDetector"
	awsServiceNames["fsx"] = "FSx"
	awsServiceNames["gamelift"] = "GameLift"
	awsServiceNames["glacier"] = "Glacier"
	awsServiceNames["globalaccelerator"] = "GlobalAccelerator"
	awsServiceNames["glue"] = "Glue"
	awsServiceNames["gluedatabrew"] = "GlueDataBrew"
	awsServiceNames["greengrass"] = "Greengrass"
	awsServiceNames["greengrassv2"] = "GreengrassV2"
	awsServiceNames["groundstation"] = "GroundStation"
	awsServiceNames["guardduty"] = "GuardDuty"
	awsServiceNames["health"] = "Health"
	awsServiceNames["healthlake"] = "HealthLake"
	awsServiceNames["honeycode"] = "HoneyCode"
	awsServiceNames["iam"] = "IAM"
	awsServiceNames["identitystore"] = "IdentityStore"
	awsServiceNames["imagebuilder"] = "ImageBuilder"
	awsServiceNames["imagebuilder"] = "Imagebuilder"
	awsServiceNames["inspector"] = "Inspector"
	awsServiceNames["iot"] = "IoT"
	awsServiceNames["iot1clickdevices"] = "IoT1ClickDevices"
	awsServiceNames["iot1clickprojects"] = "IoT1ClickProjects"
	awsServiceNames["iotanalytics"] = "IoTAnalytics"
	awsServiceNames["iotdataplane"] = "IoTDataPlane"
	awsServiceNames["iotdeviceadvisor"] = "IoTDeviceAdvisor"
	awsServiceNames["iotevents"] = "IoTEvents"
	awsServiceNames["ioteventsdata"] = "IoTEventsData"
	awsServiceNames["iotfleethub"] = "IoTFleetHub"
	awsServiceNames["iotjobsdataplane"] = "IoTJobsDataPlane"
	awsServiceNames["iotsecuretunneling"] = "IoTSecureTunneling"
	awsServiceNames["iotsitewise"] = "IoTSiteWise"
	awsServiceNames["iotthingsgraph"] = "IoTThingsGraph"
	awsServiceNames["iotwireless"] = "IoTWireless"
	awsServiceNames["ivs"] = "IVS"
	awsServiceNames["kafka"] = "Kafka"
	awsServiceNames["kendra"] = "Kendra"
	awsServiceNames["kinesis"] = "Kinesis"
	awsServiceNames["kinesisanalytics"] = "KinesisAnalytics"
	awsServiceNames["kinesisanalyticsv2"] = "KinesisAnalyticsV2"
	awsServiceNames["kinesisvideo"] = "KinesisVideo"
	awsServiceNames["kinesisvideoarchivedmedia"] = "KinesisVideoArchivedMedia"
	awsServiceNames["kinesisvideomedia"] = "KinesisVideoMedia"
	awsServiceNames["kinesisvideosignalingchannels"] = "KinesisVideoSignalingChannels"
	awsServiceNames["kms"] = "KMS"
	awsServiceNames["lakeformation"] = "LakeFormation"
	awsServiceNames["lambda"] = "Lambda"
	awsServiceNames["lexmodelbuildingservice"] = "LexModelBuildingService"
	awsServiceNames["lexmodelsv2"] = "LexModelsV2"
	awsServiceNames["lexruntime"] = "LexRuntime"
	awsServiceNames["lexruntimev2"] = "LexRuntimeV2"
	awsServiceNames["licensemanager"] = "LicenseManager"
	awsServiceNames["lightsail"] = "Lightsail"
	awsServiceNames["location"] = "Location"
	awsServiceNames["lookoutequipment"] = "LookoutEquipment"
	awsServiceNames["lookoutforvision"] = "LookoutForVision"
	awsServiceNames["lookoutmetrics"] = "LookoutMetrics"
	awsServiceNames["machinelearning"] = "MachineLearning"
	awsServiceNames["macie"] = "Macie"
	awsServiceNames["macie2"] = "Macie2"
	awsServiceNames["managedblockchain"] = "ManagedBlockchain"
	awsServiceNames["marketplacecatalog"] = "MarketplaceCatalog"
	awsServiceNames["marketplacecommerceanalytics"] = "MarketplaceCommerceAnalytics"
	awsServiceNames["marketplaceentitlement"] = "MarketplaceEntitlement"
	awsServiceNames["marketplacemetering"] = "MarketplaceMetering"
	awsServiceNames["mediaconnect"] = "MediaConnect"
	awsServiceNames["mediaconvert"] = "MediaConvert"
	awsServiceNames["medialive"] = "MediaLive"
	awsServiceNames["mediapackage"] = "MediaPackage"
	awsServiceNames["mediapackagevod"] = "MediaPackageVOD"
	awsServiceNames["mediastore"] = "MediaStore"
	awsServiceNames["mediastoredata"] = "MediaStoreData"
	awsServiceNames["mediatailor"] = "MediaTailor"
	awsServiceNames["memorydb"] = "MemoryDB"
	awsServiceNames["mgn"] = "Mgn"
	awsServiceNames["migrationhub"] = "MigrationHub"
	awsServiceNames["migrationhubconfig"] = "MigrationHubConfig"
	awsServiceNames["mobile"] = "Mobile"
	awsServiceNames["mobileanalytics"] = "MobileAnalytics"
	awsServiceNames["mq"] = "MQ"
	awsServiceNames["mturk"] = "MTurk"
	awsServiceNames["mwaa"] = "MWAA"
	awsServiceNames["neptune"] = "Neptune"
	awsServiceNames["networkfirewall"] = "NetworkFirewall"
	awsServiceNames["networkmanager"] = "NetworkManager"
	awsServiceNames["nimblestudio"] = "NimbleStudio"
	awsServiceNames["opsworks"] = "OpsWorks"
	awsServiceNames["opsworkscm"] = "OpsWorksCM"
	awsServiceNames["organizations"] = "Organizations"
	awsServiceNames["outposts"] = "Outposts"
	awsServiceNames["personalize"] = "Personalize"
	awsServiceNames["personalizeevents"] = "PersonalizeEvents"
	awsServiceNames["personalizeruntime"] = "PersonalizeRuntime"
	awsServiceNames["pi"] = "PI"
	awsServiceNames["pinpoint"] = "Pinpoint"
	awsServiceNames["pinpointemail"] = "PinpointEmail"
	awsServiceNames["pinpointsmsvoice"] = "PinpointSMSVoice"
	awsServiceNames["polly"] = "Polly"
	awsServiceNames["pricing"] = "Pricing"
	awsServiceNames["prometheusservice"] = "PrometheusService"
	awsServiceNames["proton"] = "Proton"
	awsServiceNames["qldb"] = "QLDB"
	awsServiceNames["qldbsession"] = "QLDBSession"
	awsServiceNames["quicksight"] = "QuickSight"
	awsServiceNames["ram"] = "RAM"
	awsServiceNames["rds"] = "RDS"
	awsServiceNames["rdsdata"] = "RDSData"
	awsServiceNames["rdsutils"] = "RDSUtils"
	awsServiceNames["redshift"] = "Redshift"
	awsServiceNames["redshiftdata"] = "RedshiftData"
	awsServiceNames["rekognition"] = "Rekognition"
	awsServiceNames["resourcegroups"] = "ResourceGroups"
	awsServiceNames["resourcegroupstaggingapi"] = "ResourceGroupsTaggingAPI"
	awsServiceNames["robomaker"] = "RoboMaker"
	awsServiceNames["route53"] = "Route53"
	awsServiceNames["route53domains"] = "Route53Domains"
	awsServiceNames["route53recoverycontrolconfig"] = "Route53RecoveryControlConfig"
	awsServiceNames["route53recoveryreadiness"] = "Route53RecoveryReadiness"
	awsServiceNames["route53resolver"] = "Route53Resolver"
	awsServiceNames["s3"] = "S3"
	awsServiceNames["s3control"] = "S3Control"
	awsServiceNames["s3crypto"] = "S3Crypto"
	awsServiceNames["s3manager"] = "S3Manager"
	awsServiceNames["s3outposts"] = "S3Outposts"
	awsServiceNames["sagemaker"] = "SageMaker"
	awsServiceNames["sagemakeredgemanager"] = "SageMakerEdgeManager"
	awsServiceNames["sagemakerfeaturestoreruntime"] = "SageMakerFeatureStoreRuntime"
	awsServiceNames["sagemakerruntime"] = "SageMakerRuntime"
	awsServiceNames["savingsplans"] = "SavingsPlans"
	awsServiceNames["schemas"] = "Schemas"
	awsServiceNames["secretsmanager"] = "SecretsManager"
	awsServiceNames["securityhub"] = "SecurityHub"
	awsServiceNames["serverlessapplicationrepository"] = "ServerlessApplicationRepository"
	awsServiceNames["servicecatalog"] = "ServiceCatalog"
	awsServiceNames["servicediscovery"] = "ServiceDiscovery"
	awsServiceNames["servicequotas"] = "ServiceQuotas"
	awsServiceNames["ses"] = "SES"
	awsServiceNames["sesv2"] = "SESV2"
	awsServiceNames["sfn"] = "SFN"
	awsServiceNames["shield"] = "Shield"
	awsServiceNames["sign"] = "Sign"
	awsServiceNames["signer"] = "Signer"
	awsServiceNames["simpledb"] = "SimpleDB"
	awsServiceNames["sms"] = "SMS"
	awsServiceNames["snowball"] = "Snowball"
	awsServiceNames["sns"] = "SNS"
	awsServiceNames["sqs"] = "SQS"
	awsServiceNames["ssm"] = "SSM"
	awsServiceNames["ssmcontacts"] = "SSMContacts"
	awsServiceNames["ssmincidents"] = "SSMIncidents"
	awsServiceNames["sso"] = "SSO"
	awsServiceNames["ssoadmin"] = "SSOAdmin"
	awsServiceNames["ssooidc"] = "SSOOIDC"
	awsServiceNames["storagegateway"] = "StorageGateway"
	awsServiceNames["sts"] = "STS"
	awsServiceNames["support"] = "Support"
	awsServiceNames["swf"] = "SWF"
	awsServiceNames["synthetics"] = "Synthetics"
	awsServiceNames["textract"] = "Textract"
	awsServiceNames["timestreamquery"] = "TimestreamQuery"
	awsServiceNames["timestreamwrite"] = "TimestreamWrite"
	awsServiceNames["transcribe"] = "Transcribe"
	awsServiceNames["transcribestreaming"] = "TranscribeStreaming"
	awsServiceNames["transfer"] = "Transfer"
	awsServiceNames["translate"] = "Translate"
	awsServiceNames["waf"] = "WAF"
	awsServiceNames["wafregional"] = "WAFRegional"
	awsServiceNames["wafv2"] = "WAFV2"
	awsServiceNames["wellarchitected"] = "WellArchitected"
	awsServiceNames["workdocs"] = "WorkDocs"
	awsServiceNames["worklink"] = "WorkLink"
	awsServiceNames["workmail"] = "WorkMail"
	awsServiceNames["workmailmessageflow"] = "WorkMailMessageFlow"
	awsServiceNames["workspaces"] = "WorkSpaces"
	awsServiceNames["xray"] = "XRay"
}

func fixSomeInitialisms(s string) string {
	replace := s

	replace = strings.Replace(replace, "ResourceSes", "ResourceSES", 1)
	replace = strings.Replace(replace, "ApiGateway", "APIGateway", 1)
	replace = strings.Replace(replace, "Cloudwatch", "CloudWatch", 1)
	replace = strings.Replace(replace, "CurReport", "CURReport", 1)
	replace = strings.Replace(replace, "CloudHsm", "CloudHSM", 1)
	replace = strings.Replace(replace, "DynamoDb", "DynamoDB", 1)
	replace = strings.Replace(replace, "Opsworks", "OpsWorks", 1)
	replace = strings.Replace(replace, "Precheck", "PreCheck", 1)
	replace = strings.Replace(replace, "Graphql", "GraphQL", 1)
	replace = strings.Replace(replace, "Haproxy", "HAProxy", 1)
	replace = strings.Replace(replace, "Acmpca", "ACMPCA", 1)
	replace = strings.Replace(replace, "AcmPca", "ACMPCA", 1)
	replace = strings.Replace(replace, "Dnssec", "DNSSEC", 1)
	replace = strings.Replace(replace, "DocDb", "DocDB", 1)
	replace = strings.Replace(replace, "Docdb", "DocDB", 1)
	replace = strings.Replace(replace, "Https", "HTTPS", 1)
	replace = strings.Replace(replace, "Ipset", "IPSet", 1)
	replace = strings.Replace(replace, "Iscsi", "iSCSI", 1)
	replace = strings.Replace(replace, "Mysql", "MySQL", 1)
	replace = strings.Replace(replace, "Wafv2", "WAFV2", 1)
	replace = strings.Replace(replace, "Cidr", "CIDR", 1)
	replace = strings.Replace(replace, "Coip", "CoIP", 1)
	replace = strings.Replace(replace, "Dhcp", "DHCP", 1)
	replace = strings.Replace(replace, "Dkim", "DKIM", 1)
	replace = strings.Replace(replace, "Grpc", "GRPC", 1)
	replace = strings.Replace(replace, "Http", "HTTP", 1)
	replace = strings.Replace(replace, "Mwaa", "MWAA", 1)
	replace = strings.Replace(replace, "Oidc", "OIDC", 1)
	replace = strings.Replace(replace, "Qldb", "QLDB", 1)
	replace = strings.Replace(replace, "Smtp", "SMTP", 1)
	replace = strings.Replace(replace, "Xray", "XRay", 1)
	replace = strings.Replace(replace, "Acl", "ACL", 1)
	replace = strings.Replace(replace, "Acm", "ACM", 1)
	replace = strings.Replace(replace, "Ami", "AMI", 1)
	replace = strings.Replace(replace, "Api", "API", 1)
	replace = strings.Replace(replace, "Arn", "ARN", 1)
	replace = strings.Replace(replace, "Bgp", "BGP", 1)
	replace = strings.Replace(replace, "Csv", "CSV", 1)
	replace = strings.Replace(replace, "Dax", "DAX", 1)
	replace = strings.Replace(replace, "Dlm", "DLM", 1)
	replace = strings.Replace(replace, "Dms", "DMS", 1)
	replace = strings.Replace(replace, "Dns", "DNS", 1)
	replace = strings.Replace(replace, "Ebs", "EBS", 1)
	replace = strings.Replace(replace, "Ec2", "EC2", 1)
	replace = strings.Replace(replace, "Ecr", "ECR", 1)
	replace = strings.Replace(replace, "Ecs", "ECS", 1)
	replace = strings.Replace(replace, "Efs", "EFS", 1)
	replace = strings.Replace(replace, "Eip", "EIP", 1)
	replace = strings.Replace(replace, "Eks", "EKS", 1)
	replace = strings.Replace(replace, "Elb", "ELB", 1)
	replace = strings.Replace(replace, "Emr", "EMR", 1)
	replace = strings.Replace(replace, "Fms", "FMS", 1)
	replace = strings.Replace(replace, "Fsx", "FSx", 1)
	replace = strings.Replace(replace, "Hsm", "HSM", 1)
	replace = strings.Replace(replace, "Iam", "IAM", 1)
	replace = strings.Replace(replace, "Iot", "IoT", 1)
	replace = strings.Replace(replace, "Kms", "KMS", 1)
	replace = strings.Replace(replace, "Msk", "MSK", 1)
	replace = strings.Replace(replace, "Nat", "NAT", 1)
	replace = strings.Replace(replace, "Nfs", "NFS", 1)
	replace = strings.Replace(replace, "Php", "PHP", 1)
	replace = strings.Replace(replace, "Ram", "RAM", 1)
	replace = strings.Replace(replace, "Rds", "RDS", 1)
	replace = strings.Replace(replace, "Rfc", "RFC", 1)
	replace = strings.Replace(replace, "Sfn", "SFN", 1)
	replace = strings.Replace(replace, "Smb", "SMB", 1)
	replace = strings.Replace(replace, "Sms", "SMS", 1)
	replace = strings.Replace(replace, "Sns", "SNS", 1)
	replace = strings.Replace(replace, "Sql", "SQL", 1)
	replace = strings.Replace(replace, "Sqs", "SQS", 1)
	replace = strings.Replace(replace, "Ssh", "SSH", 1)
	replace = strings.Replace(replace, "Ssm", "SSM", 1)
	replace = strings.Replace(replace, "Sso", "SSO", 1)
	replace = strings.Replace(replace, "Sts", "STS", 1)
	replace = strings.Replace(replace, "Swf", "SWF", 1)
	replace = strings.Replace(replace, "Tcp", "TCP", 1)
	replace = strings.Replace(replace, "Vpc", "VPC", 1)
	replace = strings.Replace(replace, "Vpn", "VPN", 1)
	replace = strings.Replace(replace, "Waf", "WAF", 1)
	replace = strings.Replace(replace, "Xss", "XSS", 1)
	replace = strings.Replace(replace, "Db", "DB", 1)
	replace = strings.Replace(replace, "Ip", "IP", 1)
	replace = strings.Replace(replace, "Mq", "MQ", 1)

	if replace != strings.TrimSuffix(replace, "Ids") {
		replace = fmt.Sprintf("%s%s", strings.TrimSuffix(replace, "Ids"), "IDs")
	}

	if replace != strings.TrimSuffix(replace, "Id") {
		replace = fmt.Sprintf("%s%s", strings.TrimSuffix(replace, "Id"), "ID")
	}

	return replace
}
