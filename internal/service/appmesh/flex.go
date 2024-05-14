// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandClientPolicy(vClientPolicy []interface{}) *appmesh.ClientPolicy {
	if len(vClientPolicy) == 0 || vClientPolicy[0] == nil {
		return nil
	}

	clientPolicy := &appmesh.ClientPolicy{}

	mClientPolicy := vClientPolicy[0].(map[string]interface{})

	if vTls, ok := mClientPolicy["tls"].([]interface{}); ok && len(vTls) > 0 && vTls[0] != nil {
		tls := &appmesh.ClientPolicyTls{}

		mTls := vTls[0].(map[string]interface{})

		if vCertificate, ok := mTls[names.AttrCertificate].([]interface{}); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
			certificate := &appmesh.ClientTlsCertificate{}

			mCertificate := vCertificate[0].(map[string]interface{})

			if vFile, ok := mCertificate["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
				file := &appmesh.ListenerTlsFileCertificate{}

				mFile := vFile[0].(map[string]interface{})

				if vCertificateChain, ok := mFile[names.AttrCertificateChain].(string); ok && vCertificateChain != "" {
					file.CertificateChain = aws.String(vCertificateChain)
				}
				if vPrivateKey, ok := mFile[names.AttrPrivateKey].(string); ok && vPrivateKey != "" {
					file.PrivateKey = aws.String(vPrivateKey)
				}

				certificate.File = file
			}

			if vSds, ok := mCertificate["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
				sds := &appmesh.ListenerTlsSdsCertificate{}

				mSds := vSds[0].(map[string]interface{})

				if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
					sds.SecretName = aws.String(vSecretName)
				}

				certificate.Sds = sds
			}

			tls.Certificate = certificate
		}

		if vEnforce, ok := mTls["enforce"].(bool); ok {
			tls.Enforce = aws.Bool(vEnforce)
		}

		if vPorts, ok := mTls["ports"].(*schema.Set); ok && vPorts.Len() > 0 {
			tls.Ports = flex.ExpandInt64Set(vPorts)
		}

		if vValidation, ok := mTls["validation"].([]interface{}); ok && len(vValidation) > 0 && vValidation[0] != nil {
			validation := &appmesh.TlsValidationContext{}

			mValidation := vValidation[0].(map[string]interface{})

			if vSubjectAlternativeNames, ok := mValidation["subject_alternative_names"].([]interface{}); ok && len(vSubjectAlternativeNames) > 0 && vSubjectAlternativeNames[0] != nil {
				subjectAlternativeNames := &appmesh.SubjectAlternativeNames{}

				mSubjectAlternativeNames := vSubjectAlternativeNames[0].(map[string]interface{})

				if vMatch, ok := mSubjectAlternativeNames["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
					match := &appmesh.SubjectAlternativeNameMatchers{}

					mMatch := vMatch[0].(map[string]interface{})

					if vExact, ok := mMatch["exact"].(*schema.Set); ok && vExact.Len() > 0 {
						match.Exact = flex.ExpandStringSet(vExact)
					}

					subjectAlternativeNames.Match = match
				}

				validation.SubjectAlternativeNames = subjectAlternativeNames
			}

			if vTrust, ok := mValidation["trust"].([]interface{}); ok && len(vTrust) > 0 && vTrust[0] != nil {
				trust := &appmesh.TlsValidationContextTrust{}

				mTrust := vTrust[0].(map[string]interface{})

				if vAcm, ok := mTrust["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
					acm := &appmesh.TlsValidationContextAcmTrust{}

					mAcm := vAcm[0].(map[string]interface{})

					if vCertificateAuthorityArns, ok := mAcm["certificate_authority_arns"].(*schema.Set); ok && vCertificateAuthorityArns.Len() > 0 {
						acm.CertificateAuthorityArns = flex.ExpandStringSet(vCertificateAuthorityArns)
					}

					trust.Acm = acm
				}

				if vFile, ok := mTrust["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
					file := &appmesh.TlsValidationContextFileTrust{}

					mFile := vFile[0].(map[string]interface{})

					if vCertificateChain, ok := mFile[names.AttrCertificateChain].(string); ok && vCertificateChain != "" {
						file.CertificateChain = aws.String(vCertificateChain)
					}

					trust.File = file
				}

				if vSds, ok := mTrust["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
					sds := &appmesh.TlsValidationContextSdsTrust{}

					mSds := vSds[0].(map[string]interface{})

					if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
						sds.SecretName = aws.String(vSecretName)
					}

					trust.Sds = sds
				}

				validation.Trust = trust
			}

			tls.Validation = validation
		}

		clientPolicy.Tls = tls
	}

	return clientPolicy
}

func expandDuration(vDuration []interface{}) *appmesh.Duration {
	if len(vDuration) == 0 || vDuration[0] == nil {
		return nil
	}

	duration := &appmesh.Duration{}

	mDuration := vDuration[0].(map[string]interface{})

	if vUnit, ok := mDuration[names.AttrUnit].(string); ok && vUnit != "" {
		duration.Unit = aws.String(vUnit)
	}
	if vValue, ok := mDuration[names.AttrValue].(int); ok && vValue > 0 {
		duration.Value = aws.Int64(int64(vValue))
	}

	return duration
}

func expandGRPCRoute(vGrpcRoute []interface{}) *appmesh.GrpcRoute {
	if len(vGrpcRoute) == 0 || vGrpcRoute[0] == nil {
		return nil
	}

	mGrpcRoute := vGrpcRoute[0].(map[string]interface{})

	grpcRoute := &appmesh.GrpcRoute{}

	if vGrpcRouteAction, ok := mGrpcRoute[names.AttrAction].([]interface{}); ok && len(vGrpcRouteAction) > 0 && vGrpcRouteAction[0] != nil {
		mGrpcRouteAction := vGrpcRouteAction[0].(map[string]interface{})

		if vWeightedTargets, ok := mGrpcRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
			weightedTargets := []*appmesh.WeightedTarget{}

			for _, vWeightedTarget := range vWeightedTargets.List() {
				weightedTarget := &appmesh.WeightedTarget{}

				mWeightedTarget := vWeightedTarget.(map[string]interface{})

				if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
					weightedTarget.VirtualNode = aws.String(vVirtualNode)
				}
				if vWeight, ok := mWeightedTarget[names.AttrWeight].(int); ok {
					weightedTarget.Weight = aws.Int64(int64(vWeight))
				}

				if vPort, ok := mWeightedTarget[names.AttrPort].(int); ok && vPort > 0 {
					weightedTarget.Port = aws.Int64(int64(vPort))
				}

				weightedTargets = append(weightedTargets, weightedTarget)
			}

			grpcRoute.Action = &appmesh.GrpcRouteAction{
				WeightedTargets: weightedTargets,
			}
		}
	}

	if vGrpcRouteMatch, ok := mGrpcRoute["match"].([]interface{}); ok {
		grpcRouteMatch := &appmesh.GrpcRouteMatch{}

		// Empty match is allowed.
		// https://github.com/hashicorp/terraform-provider-aws/issues/16816.

		if len(vGrpcRouteMatch) > 0 && vGrpcRouteMatch[0] != nil {
			mGrpcRouteMatch := vGrpcRouteMatch[0].(map[string]interface{})

			if vMethodName, ok := mGrpcRouteMatch["method_name"].(string); ok && vMethodName != "" {
				grpcRouteMatch.MethodName = aws.String(vMethodName)
			}
			if vServiceName, ok := mGrpcRouteMatch[names.AttrServiceName].(string); ok && vServiceName != "" {
				grpcRouteMatch.ServiceName = aws.String(vServiceName)
			}

			if vPort, ok := mGrpcRouteMatch[names.AttrPort].(int); ok && vPort > 0 {
				grpcRouteMatch.Port = aws.Int64(int64(vPort))
			}

			if vGrpcRouteMetadatas, ok := mGrpcRouteMatch["metadata"].(*schema.Set); ok && vGrpcRouteMetadatas.Len() > 0 {
				grpcRouteMetadatas := []*appmesh.GrpcRouteMetadata{}

				for _, vGrpcRouteMetadata := range vGrpcRouteMetadatas.List() {
					grpcRouteMetadata := &appmesh.GrpcRouteMetadata{}

					mGrpcRouteMetadata := vGrpcRouteMetadata.(map[string]interface{})

					if vInvert, ok := mGrpcRouteMetadata["invert"].(bool); ok {
						grpcRouteMetadata.Invert = aws.Bool(vInvert)
					}
					if vName, ok := mGrpcRouteMetadata[names.AttrName].(string); ok && vName != "" {
						grpcRouteMetadata.Name = aws.String(vName)
					}

					if vMatch, ok := mGrpcRouteMetadata["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
						grpcRouteMetadata.Match = &appmesh.GrpcRouteMetadataMatchMethod{}

						mMatch := vMatch[0].(map[string]interface{})

						if vExact, ok := mMatch["exact"].(string); ok && vExact != "" {
							grpcRouteMetadata.Match.Exact = aws.String(vExact)
						}
						if vPrefix, ok := mMatch[names.AttrPrefix].(string); ok && vPrefix != "" {
							grpcRouteMetadata.Match.Prefix = aws.String(vPrefix)
						}
						if vRegex, ok := mMatch["regex"].(string); ok && vRegex != "" {
							grpcRouteMetadata.Match.Regex = aws.String(vRegex)
						}
						if vSuffix, ok := mMatch["suffix"].(string); ok && vSuffix != "" {
							grpcRouteMetadata.Match.Suffix = aws.String(vSuffix)
						}

						if vRange, ok := mMatch["range"].([]interface{}); ok && len(vRange) > 0 && vRange[0] != nil {
							grpcRouteMetadata.Match.Range = &appmesh.MatchRange{}

							mRange := vRange[0].(map[string]interface{})

							if vEnd, ok := mRange["end"].(int); ok && vEnd > 0 {
								grpcRouteMetadata.Match.Range.End = aws.Int64(int64(vEnd))
							}
							if vStart, ok := mRange["start"].(int); ok && vStart > 0 {
								grpcRouteMetadata.Match.Range.Start = aws.Int64(int64(vStart))
							}
						}
					}

					grpcRouteMetadatas = append(grpcRouteMetadatas, grpcRouteMetadata)
				}

				grpcRouteMatch.Metadata = grpcRouteMetadatas
			}
		}

		grpcRoute.Match = grpcRouteMatch
	}

	if vGrpcRetryPolicy, ok := mGrpcRoute["retry_policy"].([]interface{}); ok && len(vGrpcRetryPolicy) > 0 && vGrpcRetryPolicy[0] != nil {
		grpcRetryPolicy := &appmesh.GrpcRetryPolicy{}

		mGrpcRetryPolicy := vGrpcRetryPolicy[0].(map[string]interface{})

		if vMaxRetries, ok := mGrpcRetryPolicy["max_retries"].(int); ok {
			grpcRetryPolicy.MaxRetries = aws.Int64(int64(vMaxRetries))
		}

		if vGrpcRetryEvents, ok := mGrpcRetryPolicy["grpc_retry_events"].(*schema.Set); ok && vGrpcRetryEvents.Len() > 0 {
			grpcRetryPolicy.GrpcRetryEvents = flex.ExpandStringSet(vGrpcRetryEvents)
		}

		if vHttpRetryEvents, ok := mGrpcRetryPolicy["http_retry_events"].(*schema.Set); ok && vHttpRetryEvents.Len() > 0 {
			grpcRetryPolicy.HttpRetryEvents = flex.ExpandStringSet(vHttpRetryEvents)
		}

		if vPerRetryTimeout, ok := mGrpcRetryPolicy["per_retry_timeout"].([]interface{}); ok {
			grpcRetryPolicy.PerRetryTimeout = expandDuration(vPerRetryTimeout)
		}

		if vTcpRetryEvents, ok := mGrpcRetryPolicy["tcp_retry_events"].(*schema.Set); ok && vTcpRetryEvents.Len() > 0 {
			grpcRetryPolicy.TcpRetryEvents = flex.ExpandStringSet(vTcpRetryEvents)
		}

		grpcRoute.RetryPolicy = grpcRetryPolicy
	}

	if vGrpcTimeout, ok := mGrpcRoute[names.AttrTimeout].([]interface{}); ok {
		grpcRoute.Timeout = expandGRPCTimeout(vGrpcTimeout)
	}

	return grpcRoute
}

func expandGRPCTimeout(vGrpcTimeout []interface{}) *appmesh.GrpcTimeout {
	if len(vGrpcTimeout) == 0 || vGrpcTimeout[0] == nil {
		return nil
	}

	grpcTimeout := &appmesh.GrpcTimeout{}

	mGrpcTimeout := vGrpcTimeout[0].(map[string]interface{})

	if vIdleTimeout, ok := mGrpcTimeout["idle"].([]interface{}); ok {
		grpcTimeout.Idle = expandDuration(vIdleTimeout)
	}

	if vPerRequestTimeout, ok := mGrpcTimeout["per_request"].([]interface{}); ok {
		grpcTimeout.PerRequest = expandDuration(vPerRequestTimeout)
	}

	return grpcTimeout
}

func expandHTTPRoute(vHttpRoute []interface{}) *appmesh.HttpRoute {
	if len(vHttpRoute) == 0 || vHttpRoute[0] == nil {
		return nil
	}

	mHttpRoute := vHttpRoute[0].(map[string]interface{})

	httpRoute := &appmesh.HttpRoute{}

	if vHttpRouteAction, ok := mHttpRoute[names.AttrAction].([]interface{}); ok && len(vHttpRouteAction) > 0 && vHttpRouteAction[0] != nil {
		mHttpRouteAction := vHttpRouteAction[0].(map[string]interface{})

		if vWeightedTargets, ok := mHttpRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
			weightedTargets := []*appmesh.WeightedTarget{}

			for _, vWeightedTarget := range vWeightedTargets.List() {
				weightedTarget := &appmesh.WeightedTarget{}

				mWeightedTarget := vWeightedTarget.(map[string]interface{})

				if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
					weightedTarget.VirtualNode = aws.String(vVirtualNode)
				}
				if vWeight, ok := mWeightedTarget[names.AttrWeight].(int); ok {
					weightedTarget.Weight = aws.Int64(int64(vWeight))
				}

				if vPort, ok := mWeightedTarget[names.AttrPort].(int); ok && vPort > 0 {
					weightedTarget.Port = aws.Int64(int64(vPort))
				}

				weightedTargets = append(weightedTargets, weightedTarget)
			}

			httpRoute.Action = &appmesh.HttpRouteAction{
				WeightedTargets: weightedTargets,
			}
		}
	}

	if vHttpRouteMatch, ok := mHttpRoute["match"].([]interface{}); ok && len(vHttpRouteMatch) > 0 && vHttpRouteMatch[0] != nil {
		httpRouteMatch := &appmesh.HttpRouteMatch{}

		mHttpRouteMatch := vHttpRouteMatch[0].(map[string]interface{})

		if vMethod, ok := mHttpRouteMatch["method"].(string); ok && vMethod != "" {
			httpRouteMatch.Method = aws.String(vMethod)
		}
		if vPort, ok := mHttpRouteMatch[names.AttrPort].(int); ok && vPort > 0 {
			httpRouteMatch.Port = aws.Int64(int64(vPort))
		}
		if vPrefix, ok := mHttpRouteMatch[names.AttrPrefix].(string); ok && vPrefix != "" {
			httpRouteMatch.Prefix = aws.String(vPrefix)
		}
		if vScheme, ok := mHttpRouteMatch["scheme"].(string); ok && vScheme != "" {
			httpRouteMatch.Scheme = aws.String(vScheme)
		}

		if vHttpRouteHeaders, ok := mHttpRouteMatch[names.AttrHeader].(*schema.Set); ok && vHttpRouteHeaders.Len() > 0 {
			httpRouteHeaders := []*appmesh.HttpRouteHeader{}

			for _, vHttpRouteHeader := range vHttpRouteHeaders.List() {
				httpRouteHeader := &appmesh.HttpRouteHeader{}

				mHttpRouteHeader := vHttpRouteHeader.(map[string]interface{})

				if vInvert, ok := mHttpRouteHeader["invert"].(bool); ok {
					httpRouteHeader.Invert = aws.Bool(vInvert)
				}
				if vName, ok := mHttpRouteHeader[names.AttrName].(string); ok && vName != "" {
					httpRouteHeader.Name = aws.String(vName)
				}

				if vMatch, ok := mHttpRouteHeader["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
					httpRouteHeader.Match = &appmesh.HeaderMatchMethod{}

					mMatch := vMatch[0].(map[string]interface{})

					if vExact, ok := mMatch["exact"].(string); ok && vExact != "" {
						httpRouteHeader.Match.Exact = aws.String(vExact)
					}
					if vPrefix, ok := mMatch[names.AttrPrefix].(string); ok && vPrefix != "" {
						httpRouteHeader.Match.Prefix = aws.String(vPrefix)
					}
					if vRegex, ok := mMatch["regex"].(string); ok && vRegex != "" {
						httpRouteHeader.Match.Regex = aws.String(vRegex)
					}
					if vSuffix, ok := mMatch["suffix"].(string); ok && vSuffix != "" {
						httpRouteHeader.Match.Suffix = aws.String(vSuffix)
					}

					if vRange, ok := mMatch["range"].([]interface{}); ok && len(vRange) > 0 && vRange[0] != nil {
						httpRouteHeader.Match.Range = &appmesh.MatchRange{}

						mRange := vRange[0].(map[string]interface{})

						if vEnd, ok := mRange["end"].(int); ok && vEnd > 0 {
							httpRouteHeader.Match.Range.End = aws.Int64(int64(vEnd))
						}
						if vStart, ok := mRange["start"].(int); ok && vStart > 0 {
							httpRouteHeader.Match.Range.Start = aws.Int64(int64(vStart))
						}
					}
				}

				httpRouteHeaders = append(httpRouteHeaders, httpRouteHeader)
			}

			httpRouteMatch.Headers = httpRouteHeaders
		}

		if vHttpRoutePath, ok := mHttpRouteMatch[names.AttrPath].([]interface{}); ok && len(vHttpRoutePath) > 0 && vHttpRoutePath[0] != nil {
			httpRoutePath := &appmesh.HttpPathMatch{}

			mHttpRoutePath := vHttpRoutePath[0].(map[string]interface{})

			if vExact, ok := mHttpRoutePath["exact"].(string); ok && vExact != "" {
				httpRoutePath.Exact = aws.String(vExact)
			}
			if vRegex, ok := mHttpRoutePath["regex"].(string); ok && vRegex != "" {
				httpRoutePath.Regex = aws.String(vRegex)
			}

			httpRouteMatch.Path = httpRoutePath
		}

		if vHttpRouteQueryParameters, ok := mHttpRouteMatch["query_parameter"].(*schema.Set); ok && vHttpRouteQueryParameters.Len() > 0 {
			httpRouteQueryParameters := []*appmesh.HttpQueryParameter{}

			for _, vHttpRouteQueryParameter := range vHttpRouteQueryParameters.List() {
				httpRouteQueryParameter := &appmesh.HttpQueryParameter{}

				mHttpRouteQueryParameter := vHttpRouteQueryParameter.(map[string]interface{})

				if vName, ok := mHttpRouteQueryParameter[names.AttrName].(string); ok && vName != "" {
					httpRouteQueryParameter.Name = aws.String(vName)
				}

				if vMatch, ok := mHttpRouteQueryParameter["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
					httpRouteQueryParameter.Match = &appmesh.QueryParameterMatch{}

					mMatch := vMatch[0].(map[string]interface{})

					if vExact, ok := mMatch["exact"].(string); ok && vExact != "" {
						httpRouteQueryParameter.Match.Exact = aws.String(vExact)
					}
				}

				httpRouteQueryParameters = append(httpRouteQueryParameters, httpRouteQueryParameter)
			}

			httpRouteMatch.QueryParameters = httpRouteQueryParameters
		}

		httpRoute.Match = httpRouteMatch
	}

	if vHttpRetryPolicy, ok := mHttpRoute["retry_policy"].([]interface{}); ok && len(vHttpRetryPolicy) > 0 && vHttpRetryPolicy[0] != nil {
		httpRetryPolicy := &appmesh.HttpRetryPolicy{}

		mHttpRetryPolicy := vHttpRetryPolicy[0].(map[string]interface{})

		if vMaxRetries, ok := mHttpRetryPolicy["max_retries"].(int); ok {
			httpRetryPolicy.MaxRetries = aws.Int64(int64(vMaxRetries))
		}

		if vHttpRetryEvents, ok := mHttpRetryPolicy["http_retry_events"].(*schema.Set); ok && vHttpRetryEvents.Len() > 0 {
			httpRetryPolicy.HttpRetryEvents = flex.ExpandStringSet(vHttpRetryEvents)
		}

		if vPerRetryTimeout, ok := mHttpRetryPolicy["per_retry_timeout"].([]interface{}); ok {
			httpRetryPolicy.PerRetryTimeout = expandDuration(vPerRetryTimeout)
		}

		if vTcpRetryEvents, ok := mHttpRetryPolicy["tcp_retry_events"].(*schema.Set); ok && vTcpRetryEvents.Len() > 0 {
			httpRetryPolicy.TcpRetryEvents = flex.ExpandStringSet(vTcpRetryEvents)
		}

		httpRoute.RetryPolicy = httpRetryPolicy
	}

	if vHttpTimeout, ok := mHttpRoute[names.AttrTimeout].([]interface{}); ok {
		httpRoute.Timeout = expandHTTPTimeout(vHttpTimeout)
	}

	return httpRoute
}

func expandHTTPTimeout(vHttpTimeout []interface{}) *appmesh.HttpTimeout {
	if len(vHttpTimeout) == 0 || vHttpTimeout[0] == nil {
		return nil
	}

	httpTimeout := &appmesh.HttpTimeout{}

	mHttpTimeout := vHttpTimeout[0].(map[string]interface{})

	if vIdleTimeout, ok := mHttpTimeout["idle"].([]interface{}); ok {
		httpTimeout.Idle = expandDuration(vIdleTimeout)
	}

	if vPerRequestTimeout, ok := mHttpTimeout["per_request"].([]interface{}); ok {
		httpTimeout.PerRequest = expandDuration(vPerRequestTimeout)
	}

	return httpTimeout
}

func expandMeshSpec(vSpec []interface{}) *appmesh.MeshSpec {
	spec := &appmesh.MeshSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vEgressFilter, ok := mSpec["egress_filter"].([]interface{}); ok && len(vEgressFilter) > 0 && vEgressFilter[0] != nil {
		mEgressFilter := vEgressFilter[0].(map[string]interface{})

		if vType, ok := mEgressFilter[names.AttrType].(string); ok && vType != "" {
			spec.EgressFilter = &appmesh.EgressFilter{
				Type: aws.String(vType),
			}
		}
	}

	if vServiceDiscovery, ok := mSpec["service_discovery"].([]interface{}); ok && len(vServiceDiscovery) > 0 && vServiceDiscovery[0] != nil {
		mServiceDiscovery := vServiceDiscovery[0].(map[string]interface{})

		if vIpPreference, ok := mServiceDiscovery["ip_preference"].(string); ok && vIpPreference != "" {
			spec.ServiceDiscovery = &appmesh.MeshServiceDiscovery{
				IpPreference: aws.String(vIpPreference),
			}
		}
	}

	return spec
}

func expandRouteSpec(vSpec []interface{}) *appmesh.RouteSpec {
	spec := &appmesh.RouteSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vGrpcRoute, ok := mSpec["grpc_route"].([]interface{}); ok {
		spec.GrpcRoute = expandGRPCRoute(vGrpcRoute)
	}

	if vHttp2Route, ok := mSpec["http2_route"].([]interface{}); ok {
		spec.Http2Route = expandHTTPRoute(vHttp2Route)
	}

	if vHttpRoute, ok := mSpec["http_route"].([]interface{}); ok {
		spec.HttpRoute = expandHTTPRoute(vHttpRoute)
	}

	if vPriority, ok := mSpec[names.AttrPriority].(int); ok && vPriority > 0 {
		spec.Priority = aws.Int64(int64(vPriority))
	}

	if vTcpRoute, ok := mSpec["tcp_route"].([]interface{}); ok {
		spec.TcpRoute = expandTCPRoute(vTcpRoute)
	}

	return spec
}

func expandTCPRoute(vTcpRoute []interface{}) *appmesh.TcpRoute {
	if len(vTcpRoute) == 0 || vTcpRoute[0] == nil {
		return nil
	}

	mTcpRoute := vTcpRoute[0].(map[string]interface{})

	tcpRoute := &appmesh.TcpRoute{}

	if vTcpRouteAction, ok := mTcpRoute[names.AttrAction].([]interface{}); ok && len(vTcpRouteAction) > 0 && vTcpRouteAction[0] != nil {
		mTcpRouteAction := vTcpRouteAction[0].(map[string]interface{})

		if vWeightedTargets, ok := mTcpRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
			weightedTargets := []*appmesh.WeightedTarget{}

			for _, vWeightedTarget := range vWeightedTargets.List() {
				weightedTarget := &appmesh.WeightedTarget{}

				mWeightedTarget := vWeightedTarget.(map[string]interface{})

				if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
					weightedTarget.VirtualNode = aws.String(vVirtualNode)
				}
				if vWeight, ok := mWeightedTarget[names.AttrWeight].(int); ok {
					weightedTarget.Weight = aws.Int64(int64(vWeight))
				}

				if vPort, ok := mWeightedTarget[names.AttrPort].(int); ok && vPort > 0 {
					weightedTarget.Port = aws.Int64(int64(vPort))
				}

				weightedTargets = append(weightedTargets, weightedTarget)
			}

			tcpRoute.Action = &appmesh.TcpRouteAction{
				WeightedTargets: weightedTargets,
			}
		}
	}

	if vTcpRouteMatch, ok := mTcpRoute["match"].([]interface{}); ok && len(vTcpRouteMatch) > 0 && vTcpRouteMatch[0] != nil {
		tcpRouteMatch := &appmesh.TcpRouteMatch{}

		mTcpRouteMatch := vTcpRouteMatch[0].(map[string]interface{})

		if vPort, ok := mTcpRouteMatch[names.AttrPort].(int); ok && vPort > 0 {
			tcpRouteMatch.Port = aws.Int64(int64(vPort))
		}
		tcpRoute.Match = tcpRouteMatch
	}

	if vTcpTimeout, ok := mTcpRoute[names.AttrTimeout].([]interface{}); ok {
		tcpRoute.Timeout = expandTCPTimeout(vTcpTimeout)
	}

	return tcpRoute
}

func expandTCPTimeout(vTcpTimeout []interface{}) *appmesh.TcpTimeout {
	if len(vTcpTimeout) == 0 || vTcpTimeout[0] == nil {
		return nil
	}

	tcpTimeout := &appmesh.TcpTimeout{}

	mTcpTimeout := vTcpTimeout[0].(map[string]interface{})

	if vIdleTimeout, ok := mTcpTimeout["idle"].([]interface{}); ok {
		tcpTimeout.Idle = expandDuration(vIdleTimeout)
	}

	return tcpTimeout
}

func expandVirtualNodeSpec(vSpec []interface{}) *appmesh.VirtualNodeSpec {
	spec := &appmesh.VirtualNodeSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vBackends, ok := mSpec["backend"].(*schema.Set); ok && vBackends.Len() > 0 {
		backends := []*appmesh.Backend{}

		for _, vBackend := range vBackends.List() {
			backend := &appmesh.Backend{}

			mBackend := vBackend.(map[string]interface{})

			if vVirtualService, ok := mBackend["virtual_service"].([]interface{}); ok && len(vVirtualService) > 0 && vVirtualService[0] != nil {
				virtualService := &appmesh.VirtualServiceBackend{}

				mVirtualService := vVirtualService[0].(map[string]interface{})

				if vVirtualServiceName, ok := mVirtualService["virtual_service_name"].(string); ok {
					virtualService.VirtualServiceName = aws.String(vVirtualServiceName)
				}

				if vClientPolicy, ok := mVirtualService["client_policy"].([]interface{}); ok {
					virtualService.ClientPolicy = expandClientPolicy(vClientPolicy)
				}

				backend.VirtualService = virtualService
			}

			backends = append(backends, backend)
		}

		spec.Backends = backends
	}

	if vBackendDefaults, ok := mSpec["backend_defaults"].([]interface{}); ok && len(vBackendDefaults) > 0 && vBackendDefaults[0] != nil {
		backendDefaults := &appmesh.BackendDefaults{}

		mBackendDefaults := vBackendDefaults[0].(map[string]interface{})

		if vClientPolicy, ok := mBackendDefaults["client_policy"].([]interface{}); ok {
			backendDefaults.ClientPolicy = expandClientPolicy(vClientPolicy)
		}

		spec.BackendDefaults = backendDefaults
	}

	if vListeners, ok := mSpec["listener"].([]interface{}); ok && len(vListeners) > 0 && vListeners[0] != nil {
		listeners := []*appmesh.Listener{}

		for _, vListener := range vListeners {
			listener := &appmesh.Listener{}

			mListener := vListener.(map[string]interface{})

			if vConnectionPool, ok := mListener["connection_pool"].([]interface{}); ok && len(vConnectionPool) > 0 && vConnectionPool[0] != nil {
				mConnectionPool := vConnectionPool[0].(map[string]interface{})

				connectionPool := &appmesh.VirtualNodeConnectionPool{}

				if vGrpcConnectionPool, ok := mConnectionPool["grpc"].([]interface{}); ok && len(vGrpcConnectionPool) > 0 && vGrpcConnectionPool[0] != nil {
					mGrpcConnectionPool := vGrpcConnectionPool[0].(map[string]interface{})

					grpcConnectionPool := &appmesh.VirtualNodeGrpcConnectionPool{}

					if vMaxRequests, ok := mGrpcConnectionPool["max_requests"].(int); ok && vMaxRequests > 0 {
						grpcConnectionPool.MaxRequests = aws.Int64(int64(vMaxRequests))
					}

					connectionPool.Grpc = grpcConnectionPool
				}

				if vHttpConnectionPool, ok := mConnectionPool["http"].([]interface{}); ok && len(vHttpConnectionPool) > 0 && vHttpConnectionPool[0] != nil {
					mHttpConnectionPool := vHttpConnectionPool[0].(map[string]interface{})

					httpConnectionPool := &appmesh.VirtualNodeHttpConnectionPool{}

					if vMaxConnections, ok := mHttpConnectionPool["max_connections"].(int); ok && vMaxConnections > 0 {
						httpConnectionPool.MaxConnections = aws.Int64(int64(vMaxConnections))
					}
					if vMaxPendingRequests, ok := mHttpConnectionPool["max_pending_requests"].(int); ok && vMaxPendingRequests > 0 {
						httpConnectionPool.MaxPendingRequests = aws.Int64(int64(vMaxPendingRequests))
					}

					connectionPool.Http = httpConnectionPool
				}

				if vHttp2ConnectionPool, ok := mConnectionPool["http2"].([]interface{}); ok && len(vHttp2ConnectionPool) > 0 && vHttp2ConnectionPool[0] != nil {
					mHttp2ConnectionPool := vHttp2ConnectionPool[0].(map[string]interface{})

					http2ConnectionPool := &appmesh.VirtualNodeHttp2ConnectionPool{}

					if vMaxRequests, ok := mHttp2ConnectionPool["max_requests"].(int); ok && vMaxRequests > 0 {
						http2ConnectionPool.MaxRequests = aws.Int64(int64(vMaxRequests))
					}

					connectionPool.Http2 = http2ConnectionPool
				}

				if vTcpConnectionPool, ok := mConnectionPool["tcp"].([]interface{}); ok && len(vTcpConnectionPool) > 0 && vTcpConnectionPool[0] != nil {
					mTcpConnectionPool := vTcpConnectionPool[0].(map[string]interface{})

					tcpConnectionPool := &appmesh.VirtualNodeTcpConnectionPool{}

					if vMaxConnections, ok := mTcpConnectionPool["max_connections"].(int); ok && vMaxConnections > 0 {
						tcpConnectionPool.MaxConnections = aws.Int64(int64(vMaxConnections))
					}

					connectionPool.Tcp = tcpConnectionPool
				}

				listener.ConnectionPool = connectionPool
			}

			if vHealthCheck, ok := mListener[names.AttrHealthCheck].([]interface{}); ok && len(vHealthCheck) > 0 && vHealthCheck[0] != nil {
				healthCheck := &appmesh.HealthCheckPolicy{}

				mHealthCheck := vHealthCheck[0].(map[string]interface{})

				if vHealthyThreshold, ok := mHealthCheck["healthy_threshold"].(int); ok && vHealthyThreshold > 0 {
					healthCheck.HealthyThreshold = aws.Int64(int64(vHealthyThreshold))
				}
				if vIntervalMillis, ok := mHealthCheck["interval_millis"].(int); ok && vIntervalMillis > 0 {
					healthCheck.IntervalMillis = aws.Int64(int64(vIntervalMillis))
				}
				if vPath, ok := mHealthCheck[names.AttrPath].(string); ok && vPath != "" {
					healthCheck.Path = aws.String(vPath)
				}
				if vPort, ok := mHealthCheck[names.AttrPort].(int); ok && vPort > 0 {
					healthCheck.Port = aws.Int64(int64(vPort))
				}
				if vProtocol, ok := mHealthCheck[names.AttrProtocol].(string); ok && vProtocol != "" {
					healthCheck.Protocol = aws.String(vProtocol)
				}
				if vTimeoutMillis, ok := mHealthCheck["timeout_millis"].(int); ok && vTimeoutMillis > 0 {
					healthCheck.TimeoutMillis = aws.Int64(int64(vTimeoutMillis))
				}
				if vUnhealthyThreshold, ok := mHealthCheck["unhealthy_threshold"].(int); ok && vUnhealthyThreshold > 0 {
					healthCheck.UnhealthyThreshold = aws.Int64(int64(vUnhealthyThreshold))
				}

				listener.HealthCheck = healthCheck
			}

			if vOutlierDetection, ok := mListener["outlier_detection"].([]interface{}); ok && len(vOutlierDetection) > 0 && vOutlierDetection[0] != nil {
				outlierDetection := &appmesh.OutlierDetection{}

				mOutlierDetection := vOutlierDetection[0].(map[string]interface{})

				if vMaxEjectionPercent, ok := mOutlierDetection["max_ejection_percent"].(int); ok && vMaxEjectionPercent > 0 {
					outlierDetection.MaxEjectionPercent = aws.Int64(int64(vMaxEjectionPercent))
				}
				if vMaxServerErrors, ok := mOutlierDetection["max_server_errors"].(int); ok && vMaxServerErrors > 0 {
					outlierDetection.MaxServerErrors = aws.Int64(int64(vMaxServerErrors))
				}

				if vBaseEjectionDuration, ok := mOutlierDetection["base_ejection_duration"].([]interface{}); ok {
					outlierDetection.BaseEjectionDuration = expandDuration(vBaseEjectionDuration)
				}

				if vInterval, ok := mOutlierDetection[names.AttrInterval].([]interface{}); ok {
					outlierDetection.Interval = expandDuration(vInterval)
				}

				listener.OutlierDetection = outlierDetection
			}

			if vPortMapping, ok := mListener["port_mapping"].([]interface{}); ok && len(vPortMapping) > 0 && vPortMapping[0] != nil {
				portMapping := &appmesh.PortMapping{}

				mPortMapping := vPortMapping[0].(map[string]interface{})

				if vPort, ok := mPortMapping[names.AttrPort].(int); ok && vPort > 0 {
					portMapping.Port = aws.Int64(int64(vPort))
				}
				if vProtocol, ok := mPortMapping[names.AttrProtocol].(string); ok && vProtocol != "" {
					portMapping.Protocol = aws.String(vProtocol)
				}

				listener.PortMapping = portMapping
			}

			if vTimeout, ok := mListener[names.AttrTimeout].([]interface{}); ok && len(vTimeout) > 0 && vTimeout[0] != nil {
				mTimeout := vTimeout[0].(map[string]interface{})

				listenerTimeout := &appmesh.ListenerTimeout{}

				if vGrpcTimeout, ok := mTimeout["grpc"].([]interface{}); ok {
					listenerTimeout.Grpc = expandGRPCTimeout(vGrpcTimeout)
				}

				if vHttpTimeout, ok := mTimeout["http"].([]interface{}); ok {
					listenerTimeout.Http = expandHTTPTimeout(vHttpTimeout)
				}

				if vHttp2Timeout, ok := mTimeout["http2"].([]interface{}); ok {
					listenerTimeout.Http2 = expandHTTPTimeout(vHttp2Timeout)
				}

				if vTcpTimeout, ok := mTimeout["tcp"].([]interface{}); ok {
					listenerTimeout.Tcp = expandTCPTimeout(vTcpTimeout)
				}

				listener.Timeout = listenerTimeout
			}

			if vTls, ok := mListener["tls"].([]interface{}); ok && len(vTls) > 0 && vTls[0] != nil {
				tls := &appmesh.ListenerTls{}

				mTls := vTls[0].(map[string]interface{})

				if vMode, ok := mTls[names.AttrMode].(string); ok && vMode != "" {
					tls.Mode = aws.String(vMode)
				}

				if vCertificate, ok := mTls[names.AttrCertificate].([]interface{}); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
					certificate := &appmesh.ListenerTlsCertificate{}

					mCertificate := vCertificate[0].(map[string]interface{})

					if vAcm, ok := mCertificate["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
						acm := &appmesh.ListenerTlsAcmCertificate{}

						mAcm := vAcm[0].(map[string]interface{})

						if vCertificateArn, ok := mAcm[names.AttrCertificateARN].(string); ok && vCertificateArn != "" {
							acm.CertificateArn = aws.String(vCertificateArn)
						}

						certificate.Acm = acm
					}

					if vFile, ok := mCertificate["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
						file := &appmesh.ListenerTlsFileCertificate{}

						mFile := vFile[0].(map[string]interface{})

						if vCertificateChain, ok := mFile[names.AttrCertificateChain].(string); ok && vCertificateChain != "" {
							file.CertificateChain = aws.String(vCertificateChain)
						}
						if vPrivateKey, ok := mFile[names.AttrPrivateKey].(string); ok && vPrivateKey != "" {
							file.PrivateKey = aws.String(vPrivateKey)
						}

						certificate.File = file
					}

					if vSds, ok := mCertificate["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
						sds := &appmesh.ListenerTlsSdsCertificate{}

						mSds := vSds[0].(map[string]interface{})

						if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
							sds.SecretName = aws.String(vSecretName)
						}

						certificate.Sds = sds
					}

					tls.Certificate = certificate
				}

				if vValidation, ok := mTls["validation"].([]interface{}); ok && len(vValidation) > 0 && vValidation[0] != nil {
					validation := &appmesh.ListenerTlsValidationContext{}

					mValidation := vValidation[0].(map[string]interface{})

					if vSubjectAlternativeNames, ok := mValidation["subject_alternative_names"].([]interface{}); ok && len(vSubjectAlternativeNames) > 0 && vSubjectAlternativeNames[0] != nil {
						subjectAlternativeNames := &appmesh.SubjectAlternativeNames{}

						mSubjectAlternativeNames := vSubjectAlternativeNames[0].(map[string]interface{})

						if vMatch, ok := mSubjectAlternativeNames["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
							match := &appmesh.SubjectAlternativeNameMatchers{}

							mMatch := vMatch[0].(map[string]interface{})

							if vExact, ok := mMatch["exact"].(*schema.Set); ok && vExact.Len() > 0 {
								match.Exact = flex.ExpandStringSet(vExact)
							}

							subjectAlternativeNames.Match = match
						}

						validation.SubjectAlternativeNames = subjectAlternativeNames
					}

					if vTrust, ok := mValidation["trust"].([]interface{}); ok && len(vTrust) > 0 && vTrust[0] != nil {
						trust := &appmesh.ListenerTlsValidationContextTrust{}

						mTrust := vTrust[0].(map[string]interface{})

						if vFile, ok := mTrust["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
							file := &appmesh.TlsValidationContextFileTrust{}

							mFile := vFile[0].(map[string]interface{})

							if vCertificateChain, ok := mFile[names.AttrCertificateChain].(string); ok && vCertificateChain != "" {
								file.CertificateChain = aws.String(vCertificateChain)
							}

							trust.File = file
						}

						if vSds, ok := mTrust["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
							sds := &appmesh.TlsValidationContextSdsTrust{}

							mSds := vSds[0].(map[string]interface{})

							if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
								sds.SecretName = aws.String(vSecretName)
							}

							trust.Sds = sds
						}

						validation.Trust = trust
					}

					tls.Validation = validation
				}

				listener.Tls = tls
			}

			listeners = append(listeners, listener)
		}

		spec.Listeners = listeners
	}

	if vLogging, ok := mSpec["logging"].([]interface{}); ok && len(vLogging) > 0 && vLogging[0] != nil {
		logging := &appmesh.Logging{}

		mLogging := vLogging[0].(map[string]interface{})

		if vAccessLog, ok := mLogging["access_log"].([]interface{}); ok && len(vAccessLog) > 0 && vAccessLog[0] != nil {
			accessLog := &appmesh.AccessLog{}

			mAccessLog := vAccessLog[0].(map[string]interface{})

			if vFile, ok := mAccessLog["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
				file := &appmesh.FileAccessLog{}

				mFile := vFile[0].(map[string]interface{})

				if vFormat, ok := mFile[names.AttrFormat].([]interface{}); ok && len(vFormat) > 0 && vFormat[0] != nil {
					format := &appmesh.LoggingFormat{}

					mFormat := vFormat[0].(map[string]interface{})

					if vJsonFormatRefs, ok := mFormat[names.AttrJSON].([]interface{}); ok && len(vJsonFormatRefs) > 0 {
						jsonFormatRefs := []*appmesh.JsonFormatRef{}
						for _, vJsonFormatRef := range vJsonFormatRefs {
							mJsonFormatRef := &appmesh.JsonFormatRef{
								Key:   aws.String(vJsonFormatRef.(map[string]interface{})[names.AttrKey].(string)),
								Value: aws.String(vJsonFormatRef.(map[string]interface{})[names.AttrValue].(string)),
							}
							jsonFormatRefs = append(jsonFormatRefs, mJsonFormatRef)
						}
						format.Json = jsonFormatRefs
					}

					if vText, ok := mFormat["text"].(string); ok && vText != "" {
						format.Text = aws.String(vText)
					}

					file.Format = format
				}

				if vPath, ok := mFile[names.AttrPath].(string); ok && vPath != "" {
					file.Path = aws.String(vPath)
				}

				accessLog.File = file
			}

			logging.AccessLog = accessLog
		}

		spec.Logging = logging
	}

	if vServiceDiscovery, ok := mSpec["service_discovery"].([]interface{}); ok && len(vServiceDiscovery) > 0 && vServiceDiscovery[0] != nil {
		serviceDiscovery := &appmesh.ServiceDiscovery{}

		mServiceDiscovery := vServiceDiscovery[0].(map[string]interface{})

		if vAwsCloudMap, ok := mServiceDiscovery["aws_cloud_map"].([]interface{}); ok && len(vAwsCloudMap) > 0 && vAwsCloudMap[0] != nil {
			awsCloudMap := &appmesh.AwsCloudMapServiceDiscovery{}

			mAwsCloudMap := vAwsCloudMap[0].(map[string]interface{})

			if vAttributes, ok := mAwsCloudMap[names.AttrAttributes].(map[string]interface{}); ok && len(vAttributes) > 0 {
				attributes := []*appmesh.AwsCloudMapInstanceAttribute{}

				for k, v := range vAttributes {
					attributes = append(attributes, &appmesh.AwsCloudMapInstanceAttribute{
						Key:   aws.String(k),
						Value: aws.String(v.(string)),
					})
				}

				awsCloudMap.Attributes = attributes
			}
			if vNamespaceName, ok := mAwsCloudMap["namespace_name"].(string); ok && vNamespaceName != "" {
				awsCloudMap.NamespaceName = aws.String(vNamespaceName)
			}
			if vServiceName, ok := mAwsCloudMap[names.AttrServiceName].(string); ok && vServiceName != "" {
				awsCloudMap.ServiceName = aws.String(vServiceName)
			}

			serviceDiscovery.AwsCloudMap = awsCloudMap
		}

		if vDns, ok := mServiceDiscovery["dns"].([]interface{}); ok && len(vDns) > 0 && vDns[0] != nil {
			dns := &appmesh.DnsServiceDiscovery{}

			mDns := vDns[0].(map[string]interface{})

			if vHostname, ok := mDns["hostname"].(string); ok && vHostname != "" {
				dns.Hostname = aws.String(vHostname)
			}

			if vIPPreference, ok := mDns["ip_preference"].(string); ok && vIPPreference != "" {
				dns.IpPreference = aws.String(vIPPreference)
			}

			if vResponseType, ok := mDns["response_type"].(string); ok && vResponseType != "" {
				dns.ResponseType = aws.String(vResponseType)
			}

			serviceDiscovery.Dns = dns
		}

		spec.ServiceDiscovery = serviceDiscovery
	}

	return spec
}

func expandVirtualRouterSpec(vSpec []interface{}) *appmesh.VirtualRouterSpec {
	spec := &appmesh.VirtualRouterSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vListeners, ok := mSpec["listener"].([]interface{}); ok && len(vListeners) > 0 && vListeners[0] != nil {
		listeners := []*appmesh.VirtualRouterListener{}

		for _, vListener := range vListeners {
			listener := &appmesh.VirtualRouterListener{}

			mListener := vListener.(map[string]interface{})

			if vPortMapping, ok := mListener["port_mapping"].([]interface{}); ok && len(vPortMapping) > 0 && vPortMapping[0] != nil {
				mPortMapping := vPortMapping[0].(map[string]interface{})

				listener.PortMapping = &appmesh.PortMapping{}

				if vPort, ok := mPortMapping[names.AttrPort].(int); ok && vPort > 0 {
					listener.PortMapping.Port = aws.Int64(int64(vPort))
				}
				if vProtocol, ok := mPortMapping[names.AttrProtocol].(string); ok && vProtocol != "" {
					listener.PortMapping.Protocol = aws.String(vProtocol)
				}
			}
			listeners = append(listeners, listener)
		}
		spec.Listeners = listeners
	}

	return spec
}

func expandVirtualServiceSpec(vSpec []interface{}) *appmesh.VirtualServiceSpec {
	spec := &appmesh.VirtualServiceSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vProvider, ok := mSpec["provider"].([]interface{}); ok && len(vProvider) > 0 && vProvider[0] != nil {
		mProvider := vProvider[0].(map[string]interface{})

		spec.Provider = &appmesh.VirtualServiceProvider{}

		if vVirtualNode, ok := mProvider["virtual_node"].([]interface{}); ok && len(vVirtualNode) > 0 && vVirtualNode[0] != nil {
			mVirtualNode := vVirtualNode[0].(map[string]interface{})

			if vVirtualNodeName, ok := mVirtualNode["virtual_node_name"].(string); ok && vVirtualNodeName != "" {
				spec.Provider.VirtualNode = &appmesh.VirtualNodeServiceProvider{
					VirtualNodeName: aws.String(vVirtualNodeName),
				}
			}
		}

		if vVirtualRouter, ok := mProvider["virtual_router"].([]interface{}); ok && len(vVirtualRouter) > 0 && vVirtualRouter[0] != nil {
			mVirtualRouter := vVirtualRouter[0].(map[string]interface{})

			if vVirtualRouterName, ok := mVirtualRouter["virtual_router_name"].(string); ok && vVirtualRouterName != "" {
				spec.Provider.VirtualRouter = &appmesh.VirtualRouterServiceProvider{
					VirtualRouterName: aws.String(vVirtualRouterName),
				}
			}
		}
	}

	return spec
}

func flattenClientPolicy(clientPolicy *appmesh.ClientPolicy) []interface{} {
	if clientPolicy == nil {
		return []interface{}{}
	}

	mClientPolicy := map[string]interface{}{}

	if tls := clientPolicy.Tls; tls != nil {
		mTls := map[string]interface{}{
			"enforce": aws.BoolValue(tls.Enforce),
			"ports":   flex.FlattenInt64Set(tls.Ports),
		}

		if certificate := tls.Certificate; certificate != nil {
			mCertificate := map[string]interface{}{}

			if file := certificate.File; file != nil {
				mFile := map[string]interface{}{
					names.AttrCertificateChain: aws.StringValue(file.CertificateChain),
					names.AttrPrivateKey:       aws.StringValue(file.PrivateKey),
				}

				mCertificate["file"] = []interface{}{mFile}
			}

			if sds := certificate.Sds; sds != nil {
				mSds := map[string]interface{}{
					"secret_name": aws.StringValue(sds.SecretName),
				}

				mCertificate["sds"] = []interface{}{mSds}
			}

			mTls[names.AttrCertificate] = []interface{}{mCertificate}
		}

		if validation := tls.Validation; validation != nil {
			mValidation := map[string]interface{}{}

			if subjectAlternativeNames := validation.SubjectAlternativeNames; subjectAlternativeNames != nil {
				mSubjectAlternativeNames := map[string]interface{}{}

				if match := subjectAlternativeNames.Match; match != nil {
					mMatch := map[string]interface{}{
						"exact": flex.FlattenStringSet(match.Exact),
					}

					mSubjectAlternativeNames["match"] = []interface{}{mMatch}
				}

				mValidation["subject_alternative_names"] = []interface{}{mSubjectAlternativeNames}
			}

			if trust := validation.Trust; trust != nil {
				mTrust := map[string]interface{}{}

				if acm := trust.Acm; acm != nil {
					mAcm := map[string]interface{}{
						"certificate_authority_arns": flex.FlattenStringSet(acm.CertificateAuthorityArns),
					}

					mTrust["acm"] = []interface{}{mAcm}
				}

				if file := trust.File; file != nil {
					mFile := map[string]interface{}{
						names.AttrCertificateChain: aws.StringValue(file.CertificateChain),
					}

					mTrust["file"] = []interface{}{mFile}
				}

				if sds := trust.Sds; sds != nil {
					mSds := map[string]interface{}{
						"secret_name": aws.StringValue(sds.SecretName),
					}

					mTrust["sds"] = []interface{}{mSds}
				}

				mValidation["trust"] = []interface{}{mTrust}
			}

			mTls["validation"] = []interface{}{mValidation}
		}

		mClientPolicy["tls"] = []interface{}{mTls}
	}

	return []interface{}{mClientPolicy}
}

func flattenDuration(duration *appmesh.Duration) []interface{} {
	if duration == nil {
		return []interface{}{}
	}

	mDuration := map[string]interface{}{
		names.AttrUnit:  aws.StringValue(duration.Unit),
		names.AttrValue: int(aws.Int64Value(duration.Value)),
	}

	return []interface{}{mDuration}
}

func flattenGRPCRoute(grpcRoute *appmesh.GrpcRoute) []interface{} {
	if grpcRoute == nil {
		return []interface{}{}
	}

	mGrpcRoute := map[string]interface{}{}

	if action := grpcRoute.Action; action != nil {
		if weightedTargets := action.WeightedTargets; weightedTargets != nil {
			vWeightedTargets := []interface{}{}

			for _, weightedTarget := range weightedTargets {
				mWeightedTarget := map[string]interface{}{
					"virtual_node":   aws.StringValue(weightedTarget.VirtualNode),
					names.AttrWeight: int(aws.Int64Value(weightedTarget.Weight)),
					names.AttrPort:   int(aws.Int64Value(weightedTarget.Port)),
				}

				vWeightedTargets = append(vWeightedTargets, mWeightedTarget)
			}

			mGrpcRoute[names.AttrAction] = []interface{}{
				map[string]interface{}{
					"weighted_target": vWeightedTargets,
				},
			}
		}
	}

	if grpcRouteMatch := grpcRoute.Match; grpcRouteMatch != nil {
		vGrpcRouteMetadatas := []interface{}{}

		for _, grpcRouteMetadata := range grpcRouteMatch.Metadata {
			mGrpcRouteMetadata := map[string]interface{}{
				"invert":       aws.BoolValue(grpcRouteMetadata.Invert),
				names.AttrName: aws.StringValue(grpcRouteMetadata.Name),
			}

			if match := grpcRouteMetadata.Match; match != nil {
				mMatch := map[string]interface{}{
					"exact":          aws.StringValue(match.Exact),
					names.AttrPrefix: aws.StringValue(match.Prefix),
					"regex":          aws.StringValue(match.Regex),
					"suffix":         aws.StringValue(match.Suffix),
				}

				if r := match.Range; r != nil {
					mRange := map[string]interface{}{
						"end":   int(aws.Int64Value(r.End)),
						"start": int(aws.Int64Value(r.Start)),
					}

					mMatch["range"] = []interface{}{mRange}
				}

				mGrpcRouteMetadata["match"] = []interface{}{mMatch}
			}

			vGrpcRouteMetadatas = append(vGrpcRouteMetadatas, mGrpcRouteMetadata)
		}

		mGrpcRoute["match"] = []interface{}{
			map[string]interface{}{
				"metadata":            vGrpcRouteMetadatas,
				"method_name":         aws.StringValue(grpcRouteMatch.MethodName),
				names.AttrServiceName: aws.StringValue(grpcRouteMatch.ServiceName),
				names.AttrPort:        int(aws.Int64Value(grpcRouteMatch.Port)),
			},
		}
	}

	if grpcRetryPolicy := grpcRoute.RetryPolicy; grpcRetryPolicy != nil {
		mGrpcRetryPolicy := map[string]interface{}{
			"grpc_retry_events": flex.FlattenStringSet(grpcRetryPolicy.GrpcRetryEvents),
			"http_retry_events": flex.FlattenStringSet(grpcRetryPolicy.HttpRetryEvents),
			"max_retries":       int(aws.Int64Value(grpcRetryPolicy.MaxRetries)),
			"per_retry_timeout": flattenDuration(grpcRetryPolicy.PerRetryTimeout),
			"tcp_retry_events":  flex.FlattenStringSet(grpcRetryPolicy.TcpRetryEvents),
		}

		mGrpcRoute["retry_policy"] = []interface{}{mGrpcRetryPolicy}
	}

	mGrpcRoute[names.AttrTimeout] = flattenGRPCTimeout(grpcRoute.Timeout)

	return []interface{}{mGrpcRoute}
}

func flattenGRPCTimeout(grpcTimeout *appmesh.GrpcTimeout) []interface{} {
	if grpcTimeout == nil {
		return []interface{}{}
	}

	mGrpcTimeout := map[string]interface{}{
		"idle":        flattenDuration(grpcTimeout.Idle),
		"per_request": flattenDuration(grpcTimeout.PerRequest),
	}

	return []interface{}{mGrpcTimeout}
}

func flattenHTTPRoute(httpRoute *appmesh.HttpRoute) []interface{} {
	if httpRoute == nil {
		return []interface{}{}
	}

	mHttpRoute := map[string]interface{}{}

	if action := httpRoute.Action; action != nil {
		if weightedTargets := action.WeightedTargets; weightedTargets != nil {
			vWeightedTargets := []interface{}{}

			for _, weightedTarget := range weightedTargets {
				mWeightedTarget := map[string]interface{}{
					"virtual_node":   aws.StringValue(weightedTarget.VirtualNode),
					names.AttrWeight: int(aws.Int64Value(weightedTarget.Weight)),
					names.AttrPort:   int(aws.Int64Value(weightedTarget.Port)),
				}

				vWeightedTargets = append(vWeightedTargets, mWeightedTarget)
			}

			mHttpRoute[names.AttrAction] = []interface{}{
				map[string]interface{}{
					"weighted_target": vWeightedTargets,
				},
			}
		}
	}

	if httpRouteMatch := httpRoute.Match; httpRouteMatch != nil {
		vHttpRouteHeaders := []interface{}{}

		for _, httpRouteHeader := range httpRouteMatch.Headers {
			mHttpRouteHeader := map[string]interface{}{
				"invert":       aws.BoolValue(httpRouteHeader.Invert),
				names.AttrName: aws.StringValue(httpRouteHeader.Name),
			}

			if match := httpRouteHeader.Match; match != nil {
				mMatch := map[string]interface{}{
					"exact":          aws.StringValue(match.Exact),
					names.AttrPrefix: aws.StringValue(match.Prefix),
					"regex":          aws.StringValue(match.Regex),
					"suffix":         aws.StringValue(match.Suffix),
				}

				if r := match.Range; r != nil {
					mRange := map[string]interface{}{
						"end":   int(aws.Int64Value(r.End)),
						"start": int(aws.Int64Value(r.Start)),
					}

					mMatch["range"] = []interface{}{mRange}
				}

				mHttpRouteHeader["match"] = []interface{}{mMatch}
			}

			vHttpRouteHeaders = append(vHttpRouteHeaders, mHttpRouteHeader)
		}

		vHttpRoutePath := []interface{}{}

		if httpRoutePath := httpRouteMatch.Path; httpRoutePath != nil {
			mHttpRoutePath := map[string]interface{}{
				"exact": aws.StringValue(httpRoutePath.Exact),
				"regex": aws.StringValue(httpRoutePath.Regex),
			}

			vHttpRoutePath = []interface{}{mHttpRoutePath}
		}

		vHttpRouteQueryParameters := []interface{}{}

		for _, httpRouteQueryParameter := range httpRouteMatch.QueryParameters {
			mHttpRouteQueryParameter := map[string]interface{}{
				names.AttrName: aws.StringValue(httpRouteQueryParameter.Name),
			}

			if match := httpRouteQueryParameter.Match; match != nil {
				mMatch := map[string]interface{}{
					"exact": aws.StringValue(match.Exact),
				}

				mHttpRouteQueryParameter["match"] = []interface{}{mMatch}
			}

			vHttpRouteQueryParameters = append(vHttpRouteQueryParameters, mHttpRouteQueryParameter)
		}

		mHttpRoute["match"] = []interface{}{
			map[string]interface{}{
				names.AttrHeader:  vHttpRouteHeaders,
				"method":          aws.StringValue(httpRouteMatch.Method),
				names.AttrPath:    vHttpRoutePath,
				names.AttrPort:    int(aws.Int64Value(httpRouteMatch.Port)),
				names.AttrPrefix:  aws.StringValue(httpRouteMatch.Prefix),
				"query_parameter": vHttpRouteQueryParameters,
				"scheme":          aws.StringValue(httpRouteMatch.Scheme),
			},
		}
	}

	if httpRetryPolicy := httpRoute.RetryPolicy; httpRetryPolicy != nil {
		mHttpRetryPolicy := map[string]interface{}{
			"http_retry_events": flex.FlattenStringSet(httpRetryPolicy.HttpRetryEvents),
			"max_retries":       int(aws.Int64Value(httpRetryPolicy.MaxRetries)),
			"per_retry_timeout": flattenDuration(httpRetryPolicy.PerRetryTimeout),
			"tcp_retry_events":  flex.FlattenStringSet(httpRetryPolicy.TcpRetryEvents),
		}

		mHttpRoute["retry_policy"] = []interface{}{mHttpRetryPolicy}
	}

	mHttpRoute[names.AttrTimeout] = flattenHTTPTimeout(httpRoute.Timeout)

	return []interface{}{mHttpRoute}
}

func flattenHTTPTimeout(httpTimeout *appmesh.HttpTimeout) []interface{} {
	if httpTimeout == nil {
		return []interface{}{}
	}

	mHttpTimeout := map[string]interface{}{
		"idle":        flattenDuration(httpTimeout.Idle),
		"per_request": flattenDuration(httpTimeout.PerRequest),
	}

	return []interface{}{mHttpTimeout}
}

func flattenMeshSpec(spec *appmesh.MeshSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if spec.EgressFilter != nil {
		mSpec["egress_filter"] = []interface{}{
			map[string]interface{}{
				names.AttrType: aws.StringValue(spec.EgressFilter.Type),
			},
		}
	}

	if spec.ServiceDiscovery != nil {
		mSpec["service_discovery"] = []interface{}{
			map[string]interface{}{
				"ip_preference": aws.StringValue(spec.ServiceDiscovery.IpPreference),
			},
		}
	}

	return []interface{}{mSpec}
}

func flattenRouteSpec(spec *appmesh.RouteSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{
		"grpc_route":       flattenGRPCRoute(spec.GrpcRoute),
		"http2_route":      flattenHTTPRoute(spec.Http2Route),
		"http_route":       flattenHTTPRoute(spec.HttpRoute),
		names.AttrPriority: int(aws.Int64Value(spec.Priority)),
		"tcp_route":        flattenTCPRoute(spec.TcpRoute),
	}

	return []interface{}{mSpec}
}

func flattenTCPRoute(tcpRoute *appmesh.TcpRoute) []interface{} {
	if tcpRoute == nil {
		return []interface{}{}
	}

	mTcpRoute := map[string]interface{}{}

	if action := tcpRoute.Action; action != nil {
		if weightedTargets := action.WeightedTargets; weightedTargets != nil {
			vWeightedTargets := []interface{}{}

			for _, weightedTarget := range weightedTargets {
				mWeightedTarget := map[string]interface{}{
					"virtual_node":   aws.StringValue(weightedTarget.VirtualNode),
					names.AttrWeight: int(aws.Int64Value(weightedTarget.Weight)),
					names.AttrPort:   int(aws.Int64Value(weightedTarget.Port)),
				}

				vWeightedTargets = append(vWeightedTargets, mWeightedTarget)
			}

			mTcpRoute[names.AttrAction] = []interface{}{
				map[string]interface{}{
					"weighted_target": vWeightedTargets,
				},
			}
		}
	}

	if tcpRouteMatch := tcpRoute.Match; tcpRouteMatch != nil {
		mTcpRoute["match"] = []interface{}{
			map[string]interface{}{
				names.AttrPort: int(aws.Int64Value(tcpRouteMatch.Port)),
			},
		}
	}

	mTcpRoute[names.AttrTimeout] = flattenTCPTimeout(tcpRoute.Timeout)

	return []interface{}{mTcpRoute}
}

func flattenTCPTimeout(tcpTimeout *appmesh.TcpTimeout) []interface{} {
	if tcpTimeout == nil {
		return []interface{}{}
	}

	mTcpTimeout := map[string]interface{}{
		"idle": flattenDuration(tcpTimeout.Idle),
	}

	return []interface{}{mTcpTimeout}
}

func flattenVirtualNodeSpec(spec *appmesh.VirtualNodeSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if backends := spec.Backends; backends != nil {
		vBackends := []interface{}{}

		for _, backend := range backends {
			mBackend := map[string]interface{}{}

			if virtualService := backend.VirtualService; virtualService != nil {
				mVirtualService := map[string]interface{}{
					"client_policy":        flattenClientPolicy(virtualService.ClientPolicy),
					"virtual_service_name": aws.StringValue(virtualService.VirtualServiceName),
				}

				mBackend["virtual_service"] = []interface{}{mVirtualService}
			}

			vBackends = append(vBackends, mBackend)
		}

		mSpec["backend"] = vBackends
	}

	if backendDefaults := spec.BackendDefaults; backendDefaults != nil {
		mBackendDefaults := map[string]interface{}{
			"client_policy": flattenClientPolicy(backendDefaults.ClientPolicy),
		}

		mSpec["backend_defaults"] = []interface{}{mBackendDefaults}
	}

	if spec.Listeners != nil && len(spec.Listeners) > 0 {
		var mListeners []interface{}
		// Per schema definition, set at most 1 Listener
		for _, listener := range spec.Listeners {
			mListener := map[string]interface{}{}

			if connectionPool := listener.ConnectionPool; connectionPool != nil {
				mConnectionPool := map[string]interface{}{}

				if grpcConnectionPool := connectionPool.Grpc; grpcConnectionPool != nil {
					mGrpcConnectionPool := map[string]interface{}{
						"max_requests": int(aws.Int64Value(grpcConnectionPool.MaxRequests)),
					}
					mConnectionPool["grpc"] = []interface{}{mGrpcConnectionPool}
				}

				if httpConnectionPool := connectionPool.Http; httpConnectionPool != nil {
					mHttpConnectionPool := map[string]interface{}{
						"max_connections":      int(aws.Int64Value(httpConnectionPool.MaxConnections)),
						"max_pending_requests": int(aws.Int64Value(httpConnectionPool.MaxPendingRequests)),
					}
					mConnectionPool["http"] = []interface{}{mHttpConnectionPool}
				}

				if http2ConnectionPool := connectionPool.Http2; http2ConnectionPool != nil {
					mHttp2ConnectionPool := map[string]interface{}{
						"max_requests": int(aws.Int64Value(http2ConnectionPool.MaxRequests)),
					}
					mConnectionPool["http2"] = []interface{}{mHttp2ConnectionPool}
				}

				if tcpConnectionPool := connectionPool.Tcp; tcpConnectionPool != nil {
					mTcpConnectionPool := map[string]interface{}{
						"max_connections": int(aws.Int64Value(tcpConnectionPool.MaxConnections)),
					}
					mConnectionPool["tcp"] = []interface{}{mTcpConnectionPool}
				}

				mListener["connection_pool"] = []interface{}{mConnectionPool}
			}

			if healthCheck := listener.HealthCheck; healthCheck != nil {
				mHealthCheck := map[string]interface{}{
					"healthy_threshold":   int(aws.Int64Value(healthCheck.HealthyThreshold)),
					"interval_millis":     int(aws.Int64Value(healthCheck.IntervalMillis)),
					names.AttrPath:        aws.StringValue(healthCheck.Path),
					names.AttrPort:        int(aws.Int64Value(healthCheck.Port)),
					names.AttrProtocol:    aws.StringValue(healthCheck.Protocol),
					"timeout_millis":      int(aws.Int64Value(healthCheck.TimeoutMillis)),
					"unhealthy_threshold": int(aws.Int64Value(healthCheck.UnhealthyThreshold)),
				}
				mListener[names.AttrHealthCheck] = []interface{}{mHealthCheck}
			}

			if outlierDetection := listener.OutlierDetection; outlierDetection != nil {
				mOutlierDetection := map[string]interface{}{
					"base_ejection_duration": flattenDuration(outlierDetection.BaseEjectionDuration),
					names.AttrInterval:       flattenDuration(outlierDetection.Interval),
					"max_ejection_percent":   int(aws.Int64Value(outlierDetection.MaxEjectionPercent)),
					"max_server_errors":      int(aws.Int64Value(outlierDetection.MaxServerErrors)),
				}
				mListener["outlier_detection"] = []interface{}{mOutlierDetection}
			}

			if portMapping := listener.PortMapping; portMapping != nil {
				mPortMapping := map[string]interface{}{
					names.AttrPort:     int(aws.Int64Value(portMapping.Port)),
					names.AttrProtocol: aws.StringValue(portMapping.Protocol),
				}
				mListener["port_mapping"] = []interface{}{mPortMapping}
			}

			if listenerTimeout := listener.Timeout; listenerTimeout != nil {
				mListenerTimeout := map[string]interface{}{
					"grpc":  flattenGRPCTimeout(listenerTimeout.Grpc),
					"http":  flattenHTTPTimeout(listenerTimeout.Http),
					"http2": flattenHTTPTimeout(listenerTimeout.Http2),
					"tcp":   flattenTCPTimeout(listenerTimeout.Tcp),
				}
				mListener[names.AttrTimeout] = []interface{}{mListenerTimeout}
			}

			if tls := listener.Tls; tls != nil {
				mTls := map[string]interface{}{
					names.AttrMode: aws.StringValue(tls.Mode),
				}

				if certificate := tls.Certificate; certificate != nil {
					mCertificate := map[string]interface{}{}

					if acm := certificate.Acm; acm != nil {
						mAcm := map[string]interface{}{
							names.AttrCertificateARN: aws.StringValue(acm.CertificateArn),
						}

						mCertificate["acm"] = []interface{}{mAcm}
					}

					if file := certificate.File; file != nil {
						mFile := map[string]interface{}{
							names.AttrCertificateChain: aws.StringValue(file.CertificateChain),
							names.AttrPrivateKey:       aws.StringValue(file.PrivateKey),
						}

						mCertificate["file"] = []interface{}{mFile}
					}

					if sds := certificate.Sds; sds != nil {
						mSds := map[string]interface{}{
							"secret_name": aws.StringValue(sds.SecretName),
						}

						mCertificate["sds"] = []interface{}{mSds}
					}

					mTls[names.AttrCertificate] = []interface{}{mCertificate}
				}

				if validation := tls.Validation; validation != nil {
					mValidation := map[string]interface{}{}

					if subjectAlternativeNames := validation.SubjectAlternativeNames; subjectAlternativeNames != nil {
						mSubjectAlternativeNames := map[string]interface{}{}

						if match := subjectAlternativeNames.Match; match != nil {
							mMatch := map[string]interface{}{
								"exact": flex.FlattenStringSet(match.Exact),
							}

							mSubjectAlternativeNames["match"] = []interface{}{mMatch}
						}

						mValidation["subject_alternative_names"] = []interface{}{mSubjectAlternativeNames}
					}

					if trust := validation.Trust; trust != nil {
						mTrust := map[string]interface{}{}

						if file := trust.File; file != nil {
							mFile := map[string]interface{}{
								names.AttrCertificateChain: aws.StringValue(file.CertificateChain),
							}

							mTrust["file"] = []interface{}{mFile}
						}

						if sds := trust.Sds; sds != nil {
							mSds := map[string]interface{}{
								"secret_name": aws.StringValue(sds.SecretName),
							}

							mTrust["sds"] = []interface{}{mSds}
						}

						mValidation["trust"] = []interface{}{mTrust}
					}

					mTls["validation"] = []interface{}{mValidation}
				}

				mListener["tls"] = []interface{}{mTls}
			}
			mListeners = append(mListeners, mListener)
		}
		mSpec["listener"] = mListeners
	}

	if logging := spec.Logging; logging != nil {
		mLogging := map[string]interface{}{}

		if accessLog := logging.AccessLog; accessLog != nil {
			mAccessLog := map[string]interface{}{}

			if file := accessLog.File; file != nil {
				mFile := map[string]interface{}{}

				if format := file.Format; format != nil {
					mFormat := map[string]interface{}{}

					if jsons := format.Json; jsons != nil {
						vJsons := []interface{}{}

						for _, j := range format.Json {
							mJson := map[string]interface{}{
								names.AttrKey:   aws.StringValue(j.Key),
								names.AttrValue: aws.StringValue(j.Value),
							}

							vJsons = append(vJsons, mJson)
						}

						mFormat[names.AttrJSON] = vJsons
					}

					if text := format.Text; text != nil {
						mFormat["text"] = aws.StringValue(text)
					}

					mFile[names.AttrFormat] = []interface{}{mFormat}
				}

				mFile[names.AttrPath] = aws.StringValue(file.Path)

				mAccessLog["file"] = []interface{}{mFile}
			}

			mLogging["access_log"] = []interface{}{mAccessLog}
		}

		mSpec["logging"] = []interface{}{mLogging}
	}

	if serviceDiscovery := spec.ServiceDiscovery; serviceDiscovery != nil {
		mServiceDiscovery := map[string]interface{}{}

		if awsCloudMap := serviceDiscovery.AwsCloudMap; awsCloudMap != nil {
			vAttributes := map[string]interface{}{}

			for _, attribute := range awsCloudMap.Attributes {
				vAttributes[aws.StringValue(attribute.Key)] = aws.StringValue(attribute.Value)
			}

			mServiceDiscovery["aws_cloud_map"] = []interface{}{
				map[string]interface{}{
					names.AttrAttributes:  vAttributes,
					"namespace_name":      aws.StringValue(awsCloudMap.NamespaceName),
					names.AttrServiceName: aws.StringValue(awsCloudMap.ServiceName),
				},
			}
		}

		if dns := serviceDiscovery.Dns; dns != nil {
			mServiceDiscovery["dns"] = []interface{}{
				map[string]interface{}{
					"hostname":      aws.StringValue(dns.Hostname),
					"ip_preference": aws.StringValue(dns.IpPreference),
					"response_type": aws.StringValue(dns.ResponseType),
				},
			}
		}

		mSpec["service_discovery"] = []interface{}{mServiceDiscovery}
	}

	return []interface{}{mSpec}
}

func flattenVirtualRouterSpec(spec *appmesh.VirtualRouterSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}
	mSpec := make(map[string]interface{})
	if spec.Listeners != nil && len(spec.Listeners) > 0 {
		var mListeners []interface{}
		for _, listener := range spec.Listeners {
			mListener := map[string]interface{}{}
			if listener.PortMapping != nil {
				mPortMapping := map[string]interface{}{
					names.AttrPort:     int(aws.Int64Value(listener.PortMapping.Port)),
					names.AttrProtocol: aws.StringValue(listener.PortMapping.Protocol),
				}
				mListener["port_mapping"] = []interface{}{mPortMapping}
			}
			mListeners = append(mListeners, mListener)
		}
		mSpec["listener"] = mListeners
	}

	return []interface{}{mSpec}
}

func flattenVirtualServiceSpec(spec *appmesh.VirtualServiceSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if spec.Provider != nil {
		mProvider := map[string]interface{}{}

		if spec.Provider.VirtualNode != nil {
			mProvider["virtual_node"] = []interface{}{
				map[string]interface{}{
					"virtual_node_name": aws.StringValue(spec.Provider.VirtualNode.VirtualNodeName),
				},
			}
		}

		if spec.Provider.VirtualRouter != nil {
			mProvider["virtual_router"] = []interface{}{
				map[string]interface{}{
					"virtual_router_name": aws.StringValue(spec.Provider.VirtualRouter.VirtualRouterName),
				},
			}
		}

		mSpec["provider"] = []interface{}{mProvider}
	}

	return []interface{}{mSpec}
}
