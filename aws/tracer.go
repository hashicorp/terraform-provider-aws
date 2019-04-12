package aws

import (
	"context"
	"errors"
	"os"
	"io"
	"strings"
	otaws "github.com/opentracing-contrib/go-aws-sdk"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport/zipkin"
	"github.com/google/uuid"
)

func injectTracer(client *AWSClient) {
	otaws.AddOTHandlers(client.acmconn.Client)
	otaws.AddOTHandlers(client.acmpcaconn.Client)
	otaws.AddOTHandlers(client.apigateway.Client)
	otaws.AddOTHandlers(client.apigatewayv2conn.Client)
	otaws.AddOTHandlers(client.appautoscalingconn.Client)
	otaws.AddOTHandlers(client.appmeshconn.Client)
	otaws.AddOTHandlers(client.appsyncconn.Client)
	otaws.AddOTHandlers(client.athenaconn.Client)
	otaws.AddOTHandlers(client.autoscalingconn.Client)
	otaws.AddOTHandlers(client.backupconn.Client)
	otaws.AddOTHandlers(client.batchconn.Client)
	otaws.AddOTHandlers(client.budgetconn.Client)
	otaws.AddOTHandlers(client.cfconn.Client)
	otaws.AddOTHandlers(client.cloud9conn.Client)
	otaws.AddOTHandlers(client.cloudfrontconn.Client)
	otaws.AddOTHandlers(client.cloudhsmv2conn.Client)
	otaws.AddOTHandlers(client.cloudsearchconn.Client)
	otaws.AddOTHandlers(client.cloudtrailconn.Client)
	otaws.AddOTHandlers(client.cloudwatchconn.Client)
	otaws.AddOTHandlers(client.cloudwatcheventsconn.Client)
	otaws.AddOTHandlers(client.cloudwatchlogsconn.Client)
	otaws.AddOTHandlers(client.codebuildconn.Client)
	otaws.AddOTHandlers(client.codecommitconn.Client)
	otaws.AddOTHandlers(client.codedeployconn.Client)
	otaws.AddOTHandlers(client.codepipelineconn.Client)
	otaws.AddOTHandlers(client.cognitoconn.Client)
	otaws.AddOTHandlers(client.cognitoidpconn.Client)
	otaws.AddOTHandlers(client.configconn.Client)
	otaws.AddOTHandlers(client.costandusagereportconn.Client)
	otaws.AddOTHandlers(client.datapipelineconn.Client)
	otaws.AddOTHandlers(client.datasyncconn.Client)
	otaws.AddOTHandlers(client.daxconn.Client)
	otaws.AddOTHandlers(client.devicefarmconn.Client)
	otaws.AddOTHandlers(client.dlmconn.Client)
	otaws.AddOTHandlers(client.dmsconn.Client)
	otaws.AddOTHandlers(client.docdbconn.Client)
	otaws.AddOTHandlers(client.dsconn.Client)
	otaws.AddOTHandlers(client.dxconn.Client)
	otaws.AddOTHandlers(client.dynamodbconn.Client)
	otaws.AddOTHandlers(client.ec2conn.Client)
	otaws.AddOTHandlers(client.ecrconn.Client)
	otaws.AddOTHandlers(client.ecsconn.Client)
	otaws.AddOTHandlers(client.efsconn.Client)
	otaws.AddOTHandlers(client.eksconn.Client)
	otaws.AddOTHandlers(client.elasticacheconn.Client)
	otaws.AddOTHandlers(client.elasticbeanstalkconn.Client)
	otaws.AddOTHandlers(client.elastictranscoderconn.Client)
	otaws.AddOTHandlers(client.elbconn.Client)
	otaws.AddOTHandlers(client.elbv2conn.Client)
	otaws.AddOTHandlers(client.emrconn.Client)
	otaws.AddOTHandlers(client.esconn.Client)
	otaws.AddOTHandlers(client.firehoseconn.Client)
	otaws.AddOTHandlers(client.fmsconn.Client)
	otaws.AddOTHandlers(client.fsxconn.Client)
	otaws.AddOTHandlers(client.gameliftconn.Client)
	otaws.AddOTHandlers(client.glacierconn.Client)
	otaws.AddOTHandlers(client.globalacceleratorconn.Client)
	otaws.AddOTHandlers(client.glueconn.Client)
	otaws.AddOTHandlers(client.guarddutyconn.Client)
	otaws.AddOTHandlers(client.iamconn.Client)
	otaws.AddOTHandlers(client.inspectorconn.Client)
	otaws.AddOTHandlers(client.iotconn.Client)
	otaws.AddOTHandlers(client.kafkaconn.Client)
	otaws.AddOTHandlers(client.kinesisanalyticsconn.Client)
	otaws.AddOTHandlers(client.kinesisanalyticsv2conn.Client)
	otaws.AddOTHandlers(client.kinesisconn.Client)
	otaws.AddOTHandlers(client.kmsconn.Client)
	otaws.AddOTHandlers(client.lambdaconn.Client)
	otaws.AddOTHandlers(client.lexmodelconn.Client)
	otaws.AddOTHandlers(client.licensemanagerconn.Client)
	otaws.AddOTHandlers(client.lightsailconn.Client)
	otaws.AddOTHandlers(client.macieconn.Client)
	otaws.AddOTHandlers(client.mediaconnectconn.Client)
	otaws.AddOTHandlers(client.mediaconvertconn.Client)
	otaws.AddOTHandlers(client.medialiveconn.Client)
	otaws.AddOTHandlers(client.mediapackageconn.Client)
	otaws.AddOTHandlers(client.mediastoreconn.Client)
	otaws.AddOTHandlers(client.mediastoredataconn.Client)
	otaws.AddOTHandlers(client.mqconn.Client)
	otaws.AddOTHandlers(client.neptuneconn.Client)
	otaws.AddOTHandlers(client.opsworksconn.Client)
	otaws.AddOTHandlers(client.organizationsconn.Client)
	otaws.AddOTHandlers(client.pinpointconn.Client)
	otaws.AddOTHandlers(client.pricingconn.Client)
	otaws.AddOTHandlers(client.quicksightconn.Client)
	otaws.AddOTHandlers(client.r53conn.Client)
	otaws.AddOTHandlers(client.ramconn.Client)
	otaws.AddOTHandlers(client.rdsconn.Client)
	otaws.AddOTHandlers(client.redshiftconn.Client)
	otaws.AddOTHandlers(client.resourcegroupsconn.Client)
	otaws.AddOTHandlers(client.route53resolverconn.Client)
	otaws.AddOTHandlers(client.s3conn.Client)
	otaws.AddOTHandlers(client.s3controlconn.Client)
	otaws.AddOTHandlers(client.sagemakerconn.Client)
	otaws.AddOTHandlers(client.scconn.Client)
	otaws.AddOTHandlers(client.sdconn.Client)
	otaws.AddOTHandlers(client.secretsmanagerconn.Client)
	otaws.AddOTHandlers(client.securityhubconn.Client)
	otaws.AddOTHandlers(client.serverlessapplicationrepositoryconn.Client)
	otaws.AddOTHandlers(client.sesConn.Client)
	otaws.AddOTHandlers(client.sfnconn.Client)
	otaws.AddOTHandlers(client.shieldconn.Client)
	otaws.AddOTHandlers(client.simpledbconn.Client)
	otaws.AddOTHandlers(client.snsconn.Client)
	otaws.AddOTHandlers(client.sqsconn.Client)
	otaws.AddOTHandlers(client.ssmconn.Client)
	otaws.AddOTHandlers(client.storagegatewayconn.Client)
	otaws.AddOTHandlers(client.stsconn.Client)
	otaws.AddOTHandlers(client.swfconn.Client)
	otaws.AddOTHandlers(client.transferconn.Client)
	otaws.AddOTHandlers(client.wafconn.Client)
	otaws.AddOTHandlers(client.wafregionalconn.Client)
	otaws.AddOTHandlers(client.worklinkconn.Client)
	otaws.AddOTHandlers(client.workspacesconn.Client)
}

// Tracer optimized for tracing a single script (e.g., a terraform run). All spans generated are grouped into a global
// span.
type ScriptTracer struct {
	tracer opentracing.Tracer
	closer io.Closer
	globalContext context.Context
	globalSpan opentracing.Span
}

// Initialize a new tracer that sends traces to Zipkin and groups all spans into a global parent span for analysis.
func NewScriptTracer() (*ScriptTracer, error) {
	zipkinEndpoint := os.Getenv("TRACE_ENDPOINT")
	if zipkinEndpoint == "" {
		return nil, errors.New("Not initializing tracing, zipkin endpoint not defined.")
	}

	transport, err := zipkin.NewHTTPTransport(zipkinEndpoint, zipkin.HTTPBatchSize(1), zipkin.HTTPLogger(jaeger.StdLogger))
	if err != nil {
		return nil, err
	}

	opts := []jaeger.TracerOption{
		jaeger.TracerOptions.Tag("run", uuid.New().String()),
	}

	tagStr := os.Getenv("TRACE_TAGS")
	if tagStr != "" {
		for _, tagsSplit := range strings.Split(tagStr, " ") {
			tag := strings.SplitN(tagsSplit, "=", 2)
			opts = append(opts, jaeger.TracerOptions.Tag(tag[0], tag[1]))
		}
	}

	tracer, closer := jaeger.NewTracer("terraform",
		jaeger.NewConstSampler(true),
		jaeger.NewRemoteReporter(transport),
		opts...)

	globalSpan, globalContext := opentracing.StartSpanFromContextWithTracer(context.TODO(), tracer, "terraform")

	s := &ScriptTracer{
		tracer: tracer,
		closer: closer,
		globalContext: globalContext,
		globalSpan: globalSpan,
	}

	opentracing.SetGlobalTracer(s)
	return s, nil
}

// Create a new span and inject the global span into it.
func (s *ScriptTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	span, _ := opentracing.StartSpanFromContextWithTracer(s.globalContext, s.tracer, operationName, opts...)
	return span
}

// Implement Tracer Inject
func (s *ScriptTracer) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	return s.tracer.Inject(sm, format, carrier)
}

// Implement Tracer Extract
func (s *ScriptTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	return s.tracer.Extract(format, carrier)
}

// Clean up the tracer.
func (s *ScriptTracer) Close() {
	s.globalSpan.Finish()
	s.closer.Close()
}
