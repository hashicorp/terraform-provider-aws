// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appmesh/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func expandClientPolicy(vClientPolicy []interface{}) *awstypes.ClientPolicy {
	if len(vClientPolicy) == 0 || vClientPolicy[0] == nil {
		return nil
	}

	clientPolicy := &awstypes.ClientPolicy{}

	mClientPolicy := vClientPolicy[0].(map[string]interface{})

	if vTls, ok := mClientPolicy["tls"].([]interface{}); ok && len(vTls) > 0 && vTls[0] != nil {
		tls := &awstypes.ClientPolicyTls{}

		mTls := vTls[0].(map[string]interface{})

		if vCertificate, ok := mTls["certificate"].([]interface{}); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
			mCertificate := vCertificate[0].(map[string]interface{})

			if vFile, ok := mCertificate["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
				certificate := &awstypes.ClientTlsCertificateMemberFile{}

				file := awstypes.ListenerTlsFileCertificate{}

				mFile := vFile[0].(map[string]interface{})

				if vCertificateChain, ok := mFile["certificate_chain"].(string); ok && vCertificateChain != "" {
					file.CertificateChain = aws.String(vCertificateChain)
				}
				if vPrivateKey, ok := mFile["private_key"].(string); ok && vPrivateKey != "" {
					file.PrivateKey = aws.String(vPrivateKey)
				}

				certificate.Value = file
				tls.Certificate = certificate
			}

			if vSds, ok := mCertificate["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
				certificate := &awstypes.ClientTlsCertificateMemberSds{}

				sds := awstypes.ListenerTlsSdsCertificate{}

				mSds := vSds[0].(map[string]interface{})

				if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
					sds.SecretName = aws.String(vSecretName)
				}

				certificate.Value = sds
				tls.Certificate = certificate
			}
		}

		if vEnforce, ok := mTls["enforce"].(bool); ok {
			tls.Enforce = aws.Bool(vEnforce)
		}

		if vPorts, ok := mTls["ports"].(*schema.Set); ok && vPorts.Len() > 0 {
			tls.Ports = flex.ExpandInt32ValueSet(vPorts)
		}

		if vValidation, ok := mTls["validation"].([]interface{}); ok && len(vValidation) > 0 && vValidation[0] != nil {
			validation := &awstypes.TlsValidationContext{}

			mValidation := vValidation[0].(map[string]interface{})

			if vSubjectAlternativeNames, ok := mValidation["subject_alternative_names"].([]interface{}); ok && len(vSubjectAlternativeNames) > 0 && vSubjectAlternativeNames[0] != nil {
				subjectAlternativeNames := &awstypes.SubjectAlternativeNames{}

				mSubjectAlternativeNames := vSubjectAlternativeNames[0].(map[string]interface{})

				if vMatch, ok := mSubjectAlternativeNames["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
					match := &awstypes.SubjectAlternativeNameMatchers{}

					mMatch := vMatch[0].(map[string]interface{})

					if vExact, ok := mMatch["exact"].(*schema.Set); ok && vExact.Len() > 0 {
						match.Exact = flex.ExpandStringValueSet(vExact)
					}

					subjectAlternativeNames.Match = match
				}

				validation.SubjectAlternativeNames = subjectAlternativeNames
			}

			if vTrust, ok := mValidation["trust"].([]interface{}); ok && len(vTrust) > 0 && vTrust[0] != nil {
				mTrust := vTrust[0].(map[string]interface{})

				if vAcm, ok := mTrust["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
					trust := &awstypes.TlsValidationContextTrustMemberAcm{}

					acm := awstypes.TlsValidationContextAcmTrust{}

					mAcm := vAcm[0].(map[string]interface{})

					if vCertificateAuthorityArns, ok := mAcm["certificate_authority_arns"].(*schema.Set); ok && vCertificateAuthorityArns.Len() > 0 {
						acm.CertificateAuthorityArns = flex.ExpandStringValueSet(vCertificateAuthorityArns)
					}

					trust.Value = acm
					validation.Trust = trust
				}

				if vFile, ok := mTrust["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
					trust := &awstypes.TlsValidationContextTrustMemberFile{}

					file := awstypes.TlsValidationContextFileTrust{}

					mFile := vFile[0].(map[string]interface{})

					if vCertificateChain, ok := mFile["certificate_chain"].(string); ok && vCertificateChain != "" {
						file.CertificateChain = aws.String(vCertificateChain)
					}

					trust.Value = file
					validation.Trust = trust
				}

				if vSds, ok := mTrust["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
					trust := &awstypes.TlsValidationContextTrustMemberSds{}

					sds := awstypes.TlsValidationContextSdsTrust{}

					mSds := vSds[0].(map[string]interface{})

					if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
						sds.SecretName = aws.String(vSecretName)
					}

					trust.Value = sds
					validation.Trust = trust
				}
			}

			tls.Validation = validation
		}

		clientPolicy.Tls = tls
	}

	return clientPolicy
}

func expandDuration(vDuration []interface{}) *awstypes.Duration {
	if len(vDuration) == 0 || vDuration[0] == nil {
		return nil
	}

	duration := &awstypes.Duration{}

	mDuration := vDuration[0].(map[string]interface{})

	if vUnit, ok := mDuration["unit"].(string); ok && vUnit != "" {
		duration.Unit = awstypes.DurationUnit(vUnit)
	}
	if vValue, ok := mDuration["value"].(int); ok && vValue > 0 {
		duration.Value = aws.Int64(int64(vValue))
	}

	return duration
}

func expandGRPCRoute(vGrpcRoute []interface{}) *awstypes.GrpcRoute {
	if len(vGrpcRoute) == 0 || vGrpcRoute[0] == nil {
		return nil
	}

	mGrpcRoute := vGrpcRoute[0].(map[string]interface{})

	grpcRoute := &awstypes.GrpcRoute{}

	if vGrpcRouteAction, ok := mGrpcRoute["action"].([]interface{}); ok && len(vGrpcRouteAction) > 0 && vGrpcRouteAction[0] != nil {
		mGrpcRouteAction := vGrpcRouteAction[0].(map[string]interface{})

		if vWeightedTargets, ok := mGrpcRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
			weightedTargets := []awstypes.WeightedTarget{}

			for _, vWeightedTarget := range vWeightedTargets.List() {
				weightedTarget := awstypes.WeightedTarget{}

				mWeightedTarget := vWeightedTarget.(map[string]interface{})

				if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
					weightedTarget.VirtualNode = aws.String(vVirtualNode)
				}
				if vWeight, ok := mWeightedTarget["weight"].(int); ok {
					weightedTarget.Weight = int32(vWeight)
				}

				if vPort, ok := mWeightedTarget["port"].(int); ok && vPort > 0 {
					weightedTarget.Port = aws.Int32(int32(vPort))
				}

				weightedTargets = append(weightedTargets, weightedTarget)
			}

			grpcRoute.Action = &awstypes.GrpcRouteAction{
				WeightedTargets: weightedTargets,
			}
		}
	}

	if vGrpcRouteMatch, ok := mGrpcRoute["match"].([]interface{}); ok {
		grpcRouteMatch := &awstypes.GrpcRouteMatch{}

		// Empty match is allowed.
		// https://github.com/hashicorp/terraform-provider-aws/issues/16816.

		if len(vGrpcRouteMatch) > 0 && vGrpcRouteMatch[0] != nil {
			mGrpcRouteMatch := vGrpcRouteMatch[0].(map[string]interface{})

			if vMethodName, ok := mGrpcRouteMatch["method_name"].(string); ok && vMethodName != "" {
				grpcRouteMatch.MethodName = aws.String(vMethodName)
			}
			if vServiceName, ok := mGrpcRouteMatch["service_name"].(string); ok && vServiceName != "" {
				grpcRouteMatch.ServiceName = aws.String(vServiceName)
			}

			if vPort, ok := mGrpcRouteMatch["port"].(int); ok && vPort > 0 {
				grpcRouteMatch.Port = aws.Int32(int32(vPort))
			}

			if vGrpcRouteMetadatas, ok := mGrpcRouteMatch["metadata"].(*schema.Set); ok && vGrpcRouteMetadatas.Len() > 0 {
				grpcRouteMetadatas := []awstypes.GrpcRouteMetadata{}

				for _, vGrpcRouteMetadata := range vGrpcRouteMetadatas.List() {
					grpcRouteMetadata := awstypes.GrpcRouteMetadata{}

					mGrpcRouteMetadata := vGrpcRouteMetadata.(map[string]interface{})

					if vInvert, ok := mGrpcRouteMetadata["invert"].(bool); ok {
						grpcRouteMetadata.Invert = aws.Bool(vInvert)
					}
					if vName, ok := mGrpcRouteMetadata["name"].(string); ok && vName != "" {
						grpcRouteMetadata.Name = aws.String(vName)
					}

					if vMatch, ok := mGrpcRouteMetadata["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
						mMatch := vMatch[0].(map[string]interface{})

						if vExact, ok := mMatch["exact"].(string); ok && vExact != "" {
							grpcRouteMetadata.Match = &awstypes.GrpcRouteMetadataMatchMethodMemberExact{Value: vExact}
						}
						if vPrefix, ok := mMatch["prefix"].(string); ok && vPrefix != "" {
							grpcRouteMetadata.Match = &awstypes.GrpcRouteMetadataMatchMethodMemberPrefix{Value: vPrefix}
						}
						if vRegex, ok := mMatch["regex"].(string); ok && vRegex != "" {
							grpcRouteMetadata.Match = &awstypes.GrpcRouteMetadataMatchMethodMemberRegex{Value: vRegex}
						}
						if vSuffix, ok := mMatch["suffix"].(string); ok && vSuffix != "" {
							grpcRouteMetadata.Match = &awstypes.GrpcRouteMetadataMatchMethodMemberSuffix{Value: vSuffix}
						}

						if vRange, ok := mMatch["range"].([]interface{}); ok && len(vRange) > 0 && vRange[0] != nil {
							memberRange := &awstypes.GrpcRouteMetadataMatchMethodMemberRange{}

							mRange := vRange[0].(map[string]interface{})

							if vEnd, ok := mRange["end"].(int); ok && vEnd > 0 {
								memberRange.Value.End = aws.Int64(int64(vEnd))
							}
							if vStart, ok := mRange["start"].(int); ok && vStart > 0 {
								memberRange.Value.Start = aws.Int64(int64(vStart))
							}

							grpcRouteMetadata.Match = memberRange
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
		grpcRetryPolicy := &awstypes.GrpcRetryPolicy{}

		mGrpcRetryPolicy := vGrpcRetryPolicy[0].(map[string]interface{})

		if vMaxRetries, ok := mGrpcRetryPolicy["max_retries"].(int); ok {
			grpcRetryPolicy.MaxRetries = aws.Int64(int64(vMaxRetries))
		}

		if vGrpcRetryEvents, ok := mGrpcRetryPolicy["grpc_retry_events"].(*schema.Set); ok && vGrpcRetryEvents.Len() > 0 {
			grpcRetryPolicy.GrpcRetryEvents = flex.ExpandStringyValueSet[awstypes.GrpcRetryPolicyEvent](vGrpcRetryEvents)
		}

		if vHttpRetryEvents, ok := mGrpcRetryPolicy["http_retry_events"].(*schema.Set); ok && vHttpRetryEvents.Len() > 0 {
			grpcRetryPolicy.HttpRetryEvents = flex.ExpandStringValueSet(vHttpRetryEvents)
		}

		if vPerRetryTimeout, ok := mGrpcRetryPolicy["per_retry_timeout"].([]interface{}); ok {
			grpcRetryPolicy.PerRetryTimeout = expandDuration(vPerRetryTimeout)
		}

		if vTcpRetryEvents, ok := mGrpcRetryPolicy["tcp_retry_events"].(*schema.Set); ok && vTcpRetryEvents.Len() > 0 {
			grpcRetryPolicy.TcpRetryEvents = flex.ExpandStringyValueSet[awstypes.TcpRetryPolicyEvent](vTcpRetryEvents)
		}

		grpcRoute.RetryPolicy = grpcRetryPolicy
	}

	if vGrpcTimeout, ok := mGrpcRoute["timeout"].([]interface{}); ok {
		grpcRoute.Timeout = expandGRPCTimeout(vGrpcTimeout)
	}

	return grpcRoute
}

func expandGRPCTimeout(vGrpcTimeout []interface{}) *awstypes.GrpcTimeout {
	if len(vGrpcTimeout) == 0 || vGrpcTimeout[0] == nil {
		return nil
	}

	grpcTimeout := &awstypes.GrpcTimeout{}

	mGrpcTimeout := vGrpcTimeout[0].(map[string]interface{})

	if vIdleTimeout, ok := mGrpcTimeout["idle"].([]interface{}); ok {
		grpcTimeout.Idle = expandDuration(vIdleTimeout)
	}

	if vPerRequestTimeout, ok := mGrpcTimeout["per_request"].([]interface{}); ok {
		grpcTimeout.PerRequest = expandDuration(vPerRequestTimeout)
	}

	return grpcTimeout
}

func expandHTTPRoute(vHttpRoute []interface{}) *awstypes.HttpRoute {
	if len(vHttpRoute) == 0 || vHttpRoute[0] == nil {
		return nil
	}

	mHttpRoute := vHttpRoute[0].(map[string]interface{})

	httpRoute := &awstypes.HttpRoute{}

	if vHttpRouteAction, ok := mHttpRoute["action"].([]interface{}); ok && len(vHttpRouteAction) > 0 && vHttpRouteAction[0] != nil {
		mHttpRouteAction := vHttpRouteAction[0].(map[string]interface{})

		if vWeightedTargets, ok := mHttpRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
			weightedTargets := []awstypes.WeightedTarget{}

			for _, vWeightedTarget := range vWeightedTargets.List() {
				weightedTarget := awstypes.WeightedTarget{}

				mWeightedTarget := vWeightedTarget.(map[string]interface{})

				if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
					weightedTarget.VirtualNode = aws.String(vVirtualNode)
				}
				if vWeight, ok := mWeightedTarget["weight"].(int); ok {
					weightedTarget.Weight = int32(vWeight)
				}

				if vPort, ok := mWeightedTarget["port"].(int); ok && vPort > 0 {
					weightedTarget.Port = aws.Int32(int32(vPort))
				}

				weightedTargets = append(weightedTargets, weightedTarget)
			}

			httpRoute.Action = &awstypes.HttpRouteAction{
				WeightedTargets: weightedTargets,
			}
		}
	}

	if vHttpRouteMatch, ok := mHttpRoute["match"].([]interface{}); ok && len(vHttpRouteMatch) > 0 && vHttpRouteMatch[0] != nil {
		httpRouteMatch := &awstypes.HttpRouteMatch{}

		mHttpRouteMatch := vHttpRouteMatch[0].(map[string]interface{})

		if vMethod, ok := mHttpRouteMatch["method"].(string); ok && vMethod != "" {
			httpRouteMatch.Method = awstypes.HttpMethod(vMethod)
		}
		if vPort, ok := mHttpRouteMatch["port"].(int); ok && vPort > 0 {
			httpRouteMatch.Port = aws.Int32(int32(vPort))
		}
		if vPrefix, ok := mHttpRouteMatch["prefix"].(string); ok && vPrefix != "" {
			httpRouteMatch.Prefix = aws.String(vPrefix)
		}
		if vScheme, ok := mHttpRouteMatch["scheme"].(string); ok && vScheme != "" {
			httpRouteMatch.Scheme = awstypes.HttpScheme(vScheme)
		}

		if vHttpRouteHeaders, ok := mHttpRouteMatch["header"].(*schema.Set); ok && vHttpRouteHeaders.Len() > 0 {
			httpRouteHeaders := []awstypes.HttpRouteHeader{}

			for _, vHttpRouteHeader := range vHttpRouteHeaders.List() {
				httpRouteHeader := awstypes.HttpRouteHeader{}

				mHttpRouteHeader := vHttpRouteHeader.(map[string]interface{})

				if vInvert, ok := mHttpRouteHeader["invert"].(bool); ok {
					httpRouteHeader.Invert = aws.Bool(vInvert)
				}
				if vName, ok := mHttpRouteHeader["name"].(string); ok && vName != "" {
					httpRouteHeader.Name = aws.String(vName)
				}

				if vMatch, ok := mHttpRouteHeader["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
					mMatch := vMatch[0].(map[string]interface{})

					if vExact, ok := mMatch["exact"].(string); ok && vExact != "" {
						httpRouteHeader.Match = &awstypes.HeaderMatchMethodMemberExact{Value: vExact}
					}
					if vPrefix, ok := mMatch["prefix"].(string); ok && vPrefix != "" {
						httpRouteHeader.Match = &awstypes.HeaderMatchMethodMemberPrefix{Value: vPrefix}
					}
					if vRegex, ok := mMatch["regex"].(string); ok && vRegex != "" {
						httpRouteHeader.Match = &awstypes.HeaderMatchMethodMemberRegex{Value: vRegex}
					}
					if vSuffix, ok := mMatch["suffix"].(string); ok && vSuffix != "" {
						httpRouteHeader.Match = &awstypes.HeaderMatchMethodMemberSuffix{Value: vSuffix}
					}

					if vRange, ok := mMatch["range"].([]interface{}); ok && len(vRange) > 0 && vRange[0] != nil {
						memberRange := &awstypes.HeaderMatchMethodMemberRange{}

						mRange := vRange[0].(map[string]interface{})

						if vEnd, ok := mRange["end"].(int); ok && vEnd > 0 {
							memberRange.Value.End = aws.Int64(int64(vEnd))
						}
						if vStart, ok := mRange["start"].(int); ok && vStart > 0 {
							memberRange.Value.Start = aws.Int64(int64(vStart))
						}

						httpRouteHeader.Match = memberRange
					}
				}

				httpRouteHeaders = append(httpRouteHeaders, httpRouteHeader)
			}

			httpRouteMatch.Headers = httpRouteHeaders
		}

		if vHttpRoutePath, ok := mHttpRouteMatch["path"].([]interface{}); ok && len(vHttpRoutePath) > 0 && vHttpRoutePath[0] != nil {
			httpRoutePath := &awstypes.HttpPathMatch{}

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
			httpRouteQueryParameters := []awstypes.HttpQueryParameter{}

			for _, vHttpRouteQueryParameter := range vHttpRouteQueryParameters.List() {
				httpRouteQueryParameter := awstypes.HttpQueryParameter{}

				mHttpRouteQueryParameter := vHttpRouteQueryParameter.(map[string]interface{})

				if vName, ok := mHttpRouteQueryParameter["name"].(string); ok && vName != "" {
					httpRouteQueryParameter.Name = aws.String(vName)
				}

				if vMatch, ok := mHttpRouteQueryParameter["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
					httpRouteQueryParameter.Match = &awstypes.QueryParameterMatch{}

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
		httpRetryPolicy := &awstypes.HttpRetryPolicy{}

		mHttpRetryPolicy := vHttpRetryPolicy[0].(map[string]interface{})

		if vMaxRetries, ok := mHttpRetryPolicy["max_retries"].(int); ok {
			httpRetryPolicy.MaxRetries = aws.Int64(int64(vMaxRetries))
		}

		if vHttpRetryEvents, ok := mHttpRetryPolicy["http_retry_events"].(*schema.Set); ok && vHttpRetryEvents.Len() > 0 {
			httpRetryPolicy.HttpRetryEvents = flex.ExpandStringValueSet(vHttpRetryEvents)
		}

		if vPerRetryTimeout, ok := mHttpRetryPolicy["per_retry_timeout"].([]interface{}); ok {
			httpRetryPolicy.PerRetryTimeout = expandDuration(vPerRetryTimeout)
		}

		if vTcpRetryEvents, ok := mHttpRetryPolicy["tcp_retry_events"].(*schema.Set); ok && vTcpRetryEvents.Len() > 0 {
			httpRetryPolicy.TcpRetryEvents = flex.ExpandStringyValueSet[awstypes.TcpRetryPolicyEvent](vTcpRetryEvents)
		}

		httpRoute.RetryPolicy = httpRetryPolicy
	}

	if vHttpTimeout, ok := mHttpRoute["timeout"].([]interface{}); ok {
		httpRoute.Timeout = expandHTTPTimeout(vHttpTimeout)
	}

	return httpRoute
}

func expandHTTPTimeout(vHttpTimeout []interface{}) *awstypes.HttpTimeout {
	if len(vHttpTimeout) == 0 || vHttpTimeout[0] == nil {
		return nil
	}

	httpTimeout := &awstypes.HttpTimeout{}

	mHttpTimeout := vHttpTimeout[0].(map[string]interface{})

	if vIdleTimeout, ok := mHttpTimeout["idle"].([]interface{}); ok {
		httpTimeout.Idle = expandDuration(vIdleTimeout)
	}

	if vPerRequestTimeout, ok := mHttpTimeout["per_request"].([]interface{}); ok {
		httpTimeout.PerRequest = expandDuration(vPerRequestTimeout)
	}

	return httpTimeout
}

func expandMeshSpec(vSpec []interface{}) *awstypes.MeshSpec {
	spec := &awstypes.MeshSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vEgressFilter, ok := mSpec["egress_filter"].([]interface{}); ok && len(vEgressFilter) > 0 && vEgressFilter[0] != nil {
		mEgressFilter := vEgressFilter[0].(map[string]interface{})

		if vType, ok := mEgressFilter["type"].(string); ok && vType != "" {
			spec.EgressFilter = &awstypes.EgressFilter{
				Type: awstypes.EgressFilterType(vType),
			}
		}
	}

	if vServiceDiscovery, ok := mSpec["service_discovery"].([]interface{}); ok && len(vServiceDiscovery) > 0 && vServiceDiscovery[0] != nil {
		mServiceDiscovery := vServiceDiscovery[0].(map[string]interface{})

		if vIpPreference, ok := mServiceDiscovery["ip_preference"].(string); ok && vIpPreference != "" {
			spec.ServiceDiscovery = &awstypes.MeshServiceDiscovery{
				IpPreference: awstypes.IpPreference(vIpPreference),
			}
		}
	}

	return spec
}

func expandRouteSpec(vSpec []interface{}) *awstypes.RouteSpec {
	spec := &awstypes.RouteSpec{}

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

	if vPriority, ok := mSpec["priority"].(int); ok && vPriority > 0 {
		spec.Priority = aws.Int32(int32(vPriority))
	}

	if vTcpRoute, ok := mSpec["tcp_route"].([]interface{}); ok {
		spec.TcpRoute = expandTCPRoute(vTcpRoute)
	}

	return spec
}

func expandTCPRoute(vTcpRoute []interface{}) *awstypes.TcpRoute {
	if len(vTcpRoute) == 0 || vTcpRoute[0] == nil {
		return nil
	}

	mTcpRoute := vTcpRoute[0].(map[string]interface{})

	tcpRoute := &awstypes.TcpRoute{}

	if vTcpRouteAction, ok := mTcpRoute["action"].([]interface{}); ok && len(vTcpRouteAction) > 0 && vTcpRouteAction[0] != nil {
		mTcpRouteAction := vTcpRouteAction[0].(map[string]interface{})

		if vWeightedTargets, ok := mTcpRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
			weightedTargets := []awstypes.WeightedTarget{}

			for _, vWeightedTarget := range vWeightedTargets.List() {
				weightedTarget := awstypes.WeightedTarget{}

				mWeightedTarget := vWeightedTarget.(map[string]interface{})

				if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
					weightedTarget.VirtualNode = aws.String(vVirtualNode)
				}
				if vWeight, ok := mWeightedTarget["weight"].(int); ok {
					weightedTarget.Weight = int32(vWeight)
				}

				if vPort, ok := mWeightedTarget["port"].(int); ok && vPort > 0 {
					weightedTarget.Port = aws.Int32(int32(vPort))
				}

				weightedTargets = append(weightedTargets, weightedTarget)
			}

			tcpRoute.Action = &awstypes.TcpRouteAction{
				WeightedTargets: weightedTargets,
			}
		}
	}

	if vTcpRouteMatch, ok := mTcpRoute["match"].([]interface{}); ok && len(vTcpRouteMatch) > 0 && vTcpRouteMatch[0] != nil {
		tcpRouteMatch := &awstypes.TcpRouteMatch{}

		mTcpRouteMatch := vTcpRouteMatch[0].(map[string]interface{})

		if vPort, ok := mTcpRouteMatch["port"].(int); ok && vPort > 0 {
			tcpRouteMatch.Port = aws.Int32(int32(vPort))
		}
		tcpRoute.Match = tcpRouteMatch
	}

	if vTcpTimeout, ok := mTcpRoute["timeout"].([]interface{}); ok {
		tcpRoute.Timeout = expandTCPTimeout(vTcpTimeout)
	}

	return tcpRoute
}

func expandTCPTimeout(vTcpTimeout []interface{}) *awstypes.TcpTimeout {
	if len(vTcpTimeout) == 0 || vTcpTimeout[0] == nil {
		return nil
	}

	tcpTimeout := &awstypes.TcpTimeout{}

	mTcpTimeout := vTcpTimeout[0].(map[string]interface{})

	if vIdleTimeout, ok := mTcpTimeout["idle"].([]interface{}); ok {
		tcpTimeout.Idle = expandDuration(vIdleTimeout)
	}

	return tcpTimeout
}

func expandVirtualNodeSpec(vSpec []interface{}) *awstypes.VirtualNodeSpec {
	spec := &awstypes.VirtualNodeSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vBackends, ok := mSpec["backend"].(*schema.Set); ok && vBackends.Len() > 0 {
		backends := []awstypes.Backend{}

		for _, vBackend := range vBackends.List() {
			backend := &awstypes.BackendMemberVirtualService{}
			mBackend := vBackend.(map[string]interface{})

			if vVirtualService, ok := mBackend["virtual_service"].([]interface{}); ok && len(vVirtualService) > 0 && vVirtualService[0] != nil {
				virtualService := awstypes.VirtualServiceBackend{}

				mVirtualService := vVirtualService[0].(map[string]interface{})

				if vVirtualServiceName, ok := mVirtualService["virtual_service_name"].(string); ok {
					virtualService.VirtualServiceName = aws.String(vVirtualServiceName)
				}

				if vClientPolicy, ok := mVirtualService["client_policy"].([]interface{}); ok {
					virtualService.ClientPolicy = expandClientPolicy(vClientPolicy)
				}

				backend.Value = virtualService
				backends = append(backends, backend)
			}
		}

		spec.Backends = backends
	}

	if vBackendDefaults, ok := mSpec["backend_defaults"].([]interface{}); ok && len(vBackendDefaults) > 0 && vBackendDefaults[0] != nil {
		backendDefaults := &awstypes.BackendDefaults{}

		mBackendDefaults := vBackendDefaults[0].(map[string]interface{})

		if vClientPolicy, ok := mBackendDefaults["client_policy"].([]interface{}); ok {
			backendDefaults.ClientPolicy = expandClientPolicy(vClientPolicy)
		}

		spec.BackendDefaults = backendDefaults
	}

	if vListeners, ok := mSpec["listener"].([]interface{}); ok && len(vListeners) > 0 && vListeners[0] != nil {
		listeners := []awstypes.Listener{}

		for _, vListener := range vListeners {
			listener := awstypes.Listener{}

			mListener := vListener.(map[string]interface{})

			if vConnectionPool, ok := mListener["connection_pool"].([]interface{}); ok && len(vConnectionPool) > 0 && vConnectionPool[0] != nil {
				mConnectionPool := vConnectionPool[0].(map[string]interface{})

				if vGrpcConnectionPool, ok := mConnectionPool["grpc"].([]interface{}); ok && len(vGrpcConnectionPool) > 0 && vGrpcConnectionPool[0] != nil {
					connectionPool := &awstypes.VirtualNodeConnectionPoolMemberGrpc{}

					mGrpcConnectionPool := vGrpcConnectionPool[0].(map[string]interface{})

					grpcConnectionPool := awstypes.VirtualNodeGrpcConnectionPool{}

					if vMaxRequests, ok := mGrpcConnectionPool["max_requests"].(int); ok && vMaxRequests > 0 {
						grpcConnectionPool.MaxRequests = aws.Int32(int32(vMaxRequests))
					}

					connectionPool.Value = grpcConnectionPool
					listener.ConnectionPool = connectionPool
				}

				if vHttpConnectionPool, ok := mConnectionPool["http"].([]interface{}); ok && len(vHttpConnectionPool) > 0 && vHttpConnectionPool[0] != nil {
					connectionPool := &awstypes.VirtualNodeConnectionPoolMemberHttp{}

					mHttpConnectionPool := vHttpConnectionPool[0].(map[string]interface{})

					httpConnectionPool := awstypes.VirtualNodeHttpConnectionPool{}

					if vMaxConnections, ok := mHttpConnectionPool["max_connections"].(int); ok && vMaxConnections > 0 {
						httpConnectionPool.MaxConnections = aws.Int32(int32(vMaxConnections))
					}
					if vMaxPendingRequests, ok := mHttpConnectionPool["max_pending_requests"].(int); ok && vMaxPendingRequests > 0 {
						httpConnectionPool.MaxPendingRequests = aws.Int32(int32(vMaxPendingRequests))
					}

					connectionPool.Value = httpConnectionPool
					listener.ConnectionPool = connectionPool
				}

				if vHttp2ConnectionPool, ok := mConnectionPool["http2"].([]interface{}); ok && len(vHttp2ConnectionPool) > 0 && vHttp2ConnectionPool[0] != nil {
					connectionPool := &awstypes.VirtualNodeConnectionPoolMemberHttp2{}

					mHttp2ConnectionPool := vHttp2ConnectionPool[0].(map[string]interface{})

					http2ConnectionPool := awstypes.VirtualNodeHttp2ConnectionPool{}

					if vMaxRequests, ok := mHttp2ConnectionPool["max_requests"].(int); ok && vMaxRequests > 0 {
						http2ConnectionPool.MaxRequests = aws.Int32(int32(vMaxRequests))
					}

					connectionPool.Value = http2ConnectionPool
					listener.ConnectionPool = connectionPool
				}

				if vTcpConnectionPool, ok := mConnectionPool["tcp"].([]interface{}); ok && len(vTcpConnectionPool) > 0 && vTcpConnectionPool[0] != nil {
					connectionPool := &awstypes.VirtualNodeConnectionPoolMemberTcp{}

					mTcpConnectionPool := vTcpConnectionPool[0].(map[string]interface{})

					tcpConnectionPool := awstypes.VirtualNodeTcpConnectionPool{}

					if vMaxConnections, ok := mTcpConnectionPool["max_connections"].(int); ok && vMaxConnections > 0 {
						tcpConnectionPool.MaxConnections = aws.Int32(int32(vMaxConnections))
					}

					connectionPool.Value = tcpConnectionPool
					listener.ConnectionPool = connectionPool
				}
			}

			if vHealthCheck, ok := mListener["health_check"].([]interface{}); ok && len(vHealthCheck) > 0 && vHealthCheck[0] != nil {
				healthCheck := &awstypes.HealthCheckPolicy{}

				mHealthCheck := vHealthCheck[0].(map[string]interface{})

				if vHealthyThreshold, ok := mHealthCheck["healthy_threshold"].(int); ok && vHealthyThreshold > 0 {
					healthCheck.HealthyThreshold = aws.Int32(int32(vHealthyThreshold))
				}
				if vIntervalMillis, ok := mHealthCheck["interval_millis"].(int); ok && vIntervalMillis > 0 {
					healthCheck.IntervalMillis = aws.Int64(int64(vIntervalMillis))
				}
				if vPath, ok := mHealthCheck["path"].(string); ok && vPath != "" {
					healthCheck.Path = aws.String(vPath)
				}
				if vPort, ok := mHealthCheck["port"].(int); ok && vPort > 0 {
					healthCheck.Port = aws.Int32(int32(vPort))
				}
				if vProtocol, ok := mHealthCheck["protocol"].(string); ok && vProtocol != "" {
					healthCheck.Protocol = awstypes.PortProtocol(vProtocol)
				}
				if vTimeoutMillis, ok := mHealthCheck["timeout_millis"].(int); ok && vTimeoutMillis > 0 {
					healthCheck.TimeoutMillis = aws.Int64(int64(vTimeoutMillis))
				}
				if vUnhealthyThreshold, ok := mHealthCheck["unhealthy_threshold"].(int); ok && vUnhealthyThreshold > 0 {
					healthCheck.UnhealthyThreshold = aws.Int32(int32(vUnhealthyThreshold))
				}

				listener.HealthCheck = healthCheck
			}

			if vOutlierDetection, ok := mListener["outlier_detection"].([]interface{}); ok && len(vOutlierDetection) > 0 && vOutlierDetection[0] != nil {
				outlierDetection := &awstypes.OutlierDetection{}

				mOutlierDetection := vOutlierDetection[0].(map[string]interface{})

				if vMaxEjectionPercent, ok := mOutlierDetection["max_ejection_percent"].(int); ok && vMaxEjectionPercent > 0 {
					outlierDetection.MaxEjectionPercent = aws.Int32(int32(vMaxEjectionPercent))
				}
				if vMaxServerErrors, ok := mOutlierDetection["max_server_errors"].(int); ok && vMaxServerErrors > 0 {
					outlierDetection.MaxServerErrors = aws.Int64(int64(vMaxServerErrors))
				}

				if vBaseEjectionDuration, ok := mOutlierDetection["base_ejection_duration"].([]interface{}); ok {
					outlierDetection.BaseEjectionDuration = expandDuration(vBaseEjectionDuration)
				}

				if vInterval, ok := mOutlierDetection["interval"].([]interface{}); ok {
					outlierDetection.Interval = expandDuration(vInterval)
				}

				listener.OutlierDetection = outlierDetection
			}

			if vPortMapping, ok := mListener["port_mapping"].([]interface{}); ok && len(vPortMapping) > 0 && vPortMapping[0] != nil {
				portMapping := &awstypes.PortMapping{}

				mPortMapping := vPortMapping[0].(map[string]interface{})

				if vPort, ok := mPortMapping["port"].(int); ok && vPort > 0 {
					portMapping.Port = aws.Int32(int32(vPort))
				}
				if vProtocol, ok := mPortMapping["protocol"].(string); ok && vProtocol != "" {
					portMapping.Protocol = awstypes.PortProtocol(vProtocol)
				}

				listener.PortMapping = portMapping
			}

			if vTimeout, ok := mListener["timeout"].([]interface{}); ok && len(vTimeout) > 0 && vTimeout[0] != nil {
				mTimeout := vTimeout[0].(map[string]interface{})

				if vGrpcTimeout, ok := mTimeout["grpc"].([]interface{}); ok {
					listener.Timeout = &awstypes.ListenerTimeoutMemberGrpc{Value: *expandGRPCTimeout(vGrpcTimeout)}
				}

				if vHttpTimeout, ok := mTimeout["http"].([]interface{}); ok {
					listener.Timeout = &awstypes.ListenerTimeoutMemberHttp{Value: *expandHTTPTimeout(vHttpTimeout)}
				}

				if vHttp2Timeout, ok := mTimeout["http2"].([]interface{}); ok {
					listener.Timeout = &awstypes.ListenerTimeoutMemberHttp2{Value: *expandHTTPTimeout(vHttp2Timeout)}
				}

				if vTcpTimeout, ok := mTimeout["tcp"].([]interface{}); ok {
					listener.Timeout = &awstypes.ListenerTimeoutMemberTcp{Value: *expandTCPTimeout(vTcpTimeout)}
				}
			}

			if vTls, ok := mListener["tls"].([]interface{}); ok && len(vTls) > 0 && vTls[0] != nil {
				tls := &awstypes.ListenerTls{}

				mTls := vTls[0].(map[string]interface{})

				if vMode, ok := mTls["mode"].(string); ok && vMode != "" {
					tls.Mode = awstypes.ListenerTlsMode(vMode)
				}

				if vCertificate, ok := mTls["certificate"].([]interface{}); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
					mCertificate := vCertificate[0].(map[string]interface{})

					if vAcm, ok := mCertificate["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
						certificate := &awstypes.ListenerTlsCertificateMemberAcm{}
						acm := awstypes.ListenerTlsAcmCertificate{}

						mAcm := vAcm[0].(map[string]interface{})

						if vCertificateArn, ok := mAcm["certificate_arn"].(string); ok && vCertificateArn != "" {
							acm.CertificateArn = aws.String(vCertificateArn)
						}

						certificate.Value = acm
						tls.Certificate = certificate
					}

					if vFile, ok := mCertificate["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
						certificate := &awstypes.ListenerTlsCertificateMemberFile{}

						file := awstypes.ListenerTlsFileCertificate{}

						mFile := vFile[0].(map[string]interface{})

						if vCertificateChain, ok := mFile["certificate_chain"].(string); ok && vCertificateChain != "" {
							file.CertificateChain = aws.String(vCertificateChain)
						}
						if vPrivateKey, ok := mFile["private_key"].(string); ok && vPrivateKey != "" {
							file.PrivateKey = aws.String(vPrivateKey)
						}

						certificate.Value = file
						tls.Certificate = certificate
					}

					if vSds, ok := mCertificate["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
						certificate := &awstypes.ListenerTlsCertificateMemberSds{}

						sds := awstypes.ListenerTlsSdsCertificate{}

						mSds := vSds[0].(map[string]interface{})

						if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
							sds.SecretName = aws.String(vSecretName)
						}

						certificate.Value = sds
						tls.Certificate = certificate
					}
				}

				if vValidation, ok := mTls["validation"].([]interface{}); ok && len(vValidation) > 0 && vValidation[0] != nil {
					validation := &awstypes.ListenerTlsValidationContext{}

					mValidation := vValidation[0].(map[string]interface{})

					if vSubjectAlternativeNames, ok := mValidation["subject_alternative_names"].([]interface{}); ok && len(vSubjectAlternativeNames) > 0 && vSubjectAlternativeNames[0] != nil {
						subjectAlternativeNames := &awstypes.SubjectAlternativeNames{}

						mSubjectAlternativeNames := vSubjectAlternativeNames[0].(map[string]interface{})

						if vMatch, ok := mSubjectAlternativeNames["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
							match := &awstypes.SubjectAlternativeNameMatchers{}

							mMatch := vMatch[0].(map[string]interface{})

							if vExact, ok := mMatch["exact"].(*schema.Set); ok && vExact.Len() > 0 {
								match.Exact = flex.ExpandStringValueSet(vExact)
							}

							subjectAlternativeNames.Match = match
						}

						validation.SubjectAlternativeNames = subjectAlternativeNames
					}

					if vTrust, ok := mValidation["trust"].([]interface{}); ok && len(vTrust) > 0 && vTrust[0] != nil {
						mTrust := vTrust[0].(map[string]interface{})

						if vFile, ok := mTrust["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
							trust := &awstypes.ListenerTlsValidationContextTrustMemberFile{}

							file := awstypes.TlsValidationContextFileTrust{}

							mFile := vFile[0].(map[string]interface{})

							if vCertificateChain, ok := mFile["certificate_chain"].(string); ok && vCertificateChain != "" {
								file.CertificateChain = aws.String(vCertificateChain)
							}

							trust.Value = file
							validation.Trust = trust
						}

						if vSds, ok := mTrust["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
							trust := &awstypes.ListenerTlsValidationContextTrustMemberSds{}

							sds := awstypes.TlsValidationContextSdsTrust{}

							mSds := vSds[0].(map[string]interface{})

							if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
								sds.SecretName = aws.String(vSecretName)
							}

							trust.Value = sds
							validation.Trust = trust
						}
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
		logging := &awstypes.Logging{}

		mLogging := vLogging[0].(map[string]interface{})

		if vAccessLog, ok := mLogging["access_log"].([]interface{}); ok && len(vAccessLog) > 0 && vAccessLog[0] != nil {
			mAccessLog := vAccessLog[0].(map[string]interface{})

			if vFile, ok := mAccessLog["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
				accessLog := &awstypes.AccessLogMemberFile{}

				file := awstypes.FileAccessLog{}

				mFile := vFile[0].(map[string]interface{})

				if vFormat, ok := mFile["format"].([]interface{}); ok && len(vFormat) > 0 && vFormat[0] != nil {
					mFormat := vFormat[0].(map[string]interface{})

					if vJsonFormatRefs, ok := mFormat["json"].([]interface{}); ok && len(vJsonFormatRefs) > 0 {
						format := &awstypes.LoggingFormatMemberJson{}
						jsonFormatRefs := []awstypes.JsonFormatRef{}
						for _, vJsonFormatRef := range vJsonFormatRefs {
							mJsonFormatRef := awstypes.JsonFormatRef{
								Key:   aws.String(vJsonFormatRef.(map[string]interface{})["key"].(string)),
								Value: aws.String(vJsonFormatRef.(map[string]interface{})["value"].(string)),
							}
							jsonFormatRefs = append(jsonFormatRefs, mJsonFormatRef)
						}
						format.Value = jsonFormatRefs
						file.Format = format
					}

					if vText, ok := mFormat["text"].(string); ok && vText != "" {
						format := &awstypes.LoggingFormatMemberText{}
						format.Value = vText
						file.Format = format
					}

					logging.AccessLog = accessLog
				}

				if vPath, ok := mFile["path"].(string); ok && vPath != "" {
					file.Path = aws.String(vPath)
				}

				accessLog.Value = file
				logging.AccessLog = accessLog
			}
		}

		spec.Logging = logging
	}

	if vServiceDiscovery, ok := mSpec["service_discovery"].([]interface{}); ok && len(vServiceDiscovery) > 0 && vServiceDiscovery[0] != nil {
		mServiceDiscovery := vServiceDiscovery[0].(map[string]interface{})

		if vAwsCloudMap, ok := mServiceDiscovery["aws_cloud_map"].([]interface{}); ok && len(vAwsCloudMap) > 0 && vAwsCloudMap[0] != nil {
			serviceDiscovery := &awstypes.ServiceDiscoveryMemberAwsCloudMap{}

			awsCloudMap := awstypes.AwsCloudMapServiceDiscovery{}

			mAwsCloudMap := vAwsCloudMap[0].(map[string]interface{})

			if vAttributes, ok := mAwsCloudMap["attributes"].(map[string]interface{}); ok && len(vAttributes) > 0 {
				attributes := []awstypes.AwsCloudMapInstanceAttribute{}

				for k, v := range vAttributes {
					attributes = append(attributes, awstypes.AwsCloudMapInstanceAttribute{
						Key:   aws.String(k),
						Value: aws.String(v.(string)),
					})
				}

				awsCloudMap.Attributes = attributes
			}
			if vNamespaceName, ok := mAwsCloudMap["namespace_name"].(string); ok && vNamespaceName != "" {
				awsCloudMap.NamespaceName = aws.String(vNamespaceName)
			}
			if vServiceName, ok := mAwsCloudMap["service_name"].(string); ok && vServiceName != "" {
				awsCloudMap.ServiceName = aws.String(vServiceName)
			}

			serviceDiscovery.Value = awsCloudMap
			spec.ServiceDiscovery = serviceDiscovery
		}

		if vDns, ok := mServiceDiscovery["dns"].([]interface{}); ok && len(vDns) > 0 && vDns[0] != nil {
			serviceDiscovery := &awstypes.ServiceDiscoveryMemberDns{}

			dns := awstypes.DnsServiceDiscovery{}

			mDns := vDns[0].(map[string]interface{})

			if vHostname, ok := mDns["hostname"].(string); ok && vHostname != "" {
				dns.Hostname = aws.String(vHostname)
			}

			if vIPPreference, ok := mDns["ip_preference"].(string); ok && vIPPreference != "" {
				dns.IpPreference = awstypes.IpPreference(vIPPreference)
			}

			if vResponseType, ok := mDns["response_type"].(string); ok && vResponseType != "" {
				dns.ResponseType = awstypes.DnsResponseType(vResponseType)
			}

			serviceDiscovery.Value = dns
			spec.ServiceDiscovery = serviceDiscovery
		}
	}

	return spec
}

func expandVirtualRouterSpec(vSpec []interface{}) *awstypes.VirtualRouterSpec {
	spec := &awstypes.VirtualRouterSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vListeners, ok := mSpec["listener"].([]interface{}); ok && len(vListeners) > 0 && vListeners[0] != nil {
		listeners := []awstypes.VirtualRouterListener{}

		for _, vListener := range vListeners {
			listener := awstypes.VirtualRouterListener{}

			mListener := vListener.(map[string]interface{})

			if vPortMapping, ok := mListener["port_mapping"].([]interface{}); ok && len(vPortMapping) > 0 && vPortMapping[0] != nil {
				mPortMapping := vPortMapping[0].(map[string]interface{})

				listener.PortMapping = &awstypes.PortMapping{}

				if vPort, ok := mPortMapping["port"].(int); ok && vPort > 0 {
					listener.PortMapping.Port = aws.Int32(int32(vPort))
				}
				if vProtocol, ok := mPortMapping["protocol"].(string); ok && vProtocol != "" {
					listener.PortMapping.Protocol = awstypes.PortProtocol(vProtocol)
				}
			}
			listeners = append(listeners, listener)
		}
		spec.Listeners = listeners
	}

	return spec
}

func expandVirtualServiceSpec(vSpec []interface{}) *awstypes.VirtualServiceSpec {
	spec := &awstypes.VirtualServiceSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vProvider, ok := mSpec["provider"].([]interface{}); ok && len(vProvider) > 0 && vProvider[0] != nil {
		mProvider := vProvider[0].(map[string]interface{})

		if vVirtualNode, ok := mProvider["virtual_node"].([]interface{}); ok && len(vVirtualNode) > 0 && vVirtualNode[0] != nil {
			provider := &awstypes.VirtualServiceProviderMemberVirtualNode{}
			mVirtualNode := vVirtualNode[0].(map[string]interface{})

			if vVirtualNodeName, ok := mVirtualNode["virtual_node_name"].(string); ok && vVirtualNodeName != "" {
				provider.Value = awstypes.VirtualNodeServiceProvider{
					VirtualNodeName: aws.String(vVirtualNodeName),
				}
			}

			spec.Provider = provider
		}

		if vVirtualRouter, ok := mProvider["virtual_router"].([]interface{}); ok && len(vVirtualRouter) > 0 && vVirtualRouter[0] != nil {
			provider := &awstypes.VirtualServiceProviderMemberVirtualRouter{}
			mVirtualRouter := vVirtualRouter[0].(map[string]interface{})

			if vVirtualRouterName, ok := mVirtualRouter["virtual_router_name"].(string); ok && vVirtualRouterName != "" {
				provider.Value = awstypes.VirtualRouterServiceProvider{
					VirtualRouterName: aws.String(vVirtualRouterName),
				}
			}

			spec.Provider = provider
		}
	}

	return spec
}

func flattenClientPolicy(clientPolicy *awstypes.ClientPolicy) []interface{} {
	if clientPolicy == nil {
		return []interface{}{}
	}

	mClientPolicy := map[string]interface{}{}

	if tls := clientPolicy.Tls; tls != nil {
		mTls := map[string]interface{}{
			"enforce": aws.ToBool(tls.Enforce),
			"ports":   flex.FlattenInt32ValueSet(tls.Ports),
		}

		if certificate := tls.Certificate; certificate != nil {
			mCertificate := map[string]interface{}{}

			switch v := certificate.(type) {
			case *awstypes.ClientTlsCertificateMemberFile:
				mFile := map[string]interface{}{
					"certificate_chain": aws.ToString(v.Value.CertificateChain),
					"private_key":       aws.ToString(v.Value.PrivateKey),
				}

				mCertificate["file"] = []interface{}{mFile}
			case *awstypes.ClientTlsCertificateMemberSds:
				mSds := map[string]interface{}{
					"secret_name": aws.ToString(v.Value.SecretName),
				}

				mCertificate["sds"] = []interface{}{mSds}
			}

			mTls["certificate"] = []interface{}{mCertificate}
		}

		if validation := tls.Validation; validation != nil {
			mValidation := map[string]interface{}{}

			if subjectAlternativeNames := validation.SubjectAlternativeNames; subjectAlternativeNames != nil {
				mSubjectAlternativeNames := map[string]interface{}{}

				if match := subjectAlternativeNames.Match; match != nil {
					mMatch := map[string]interface{}{
						"exact": flex.FlattenStringValueSet(match.Exact),
					}

					mSubjectAlternativeNames["match"] = []interface{}{mMatch}
				}

				mValidation["subject_alternative_names"] = []interface{}{mSubjectAlternativeNames}
			}

			if trust := validation.Trust; trust != nil {
				mTrust := map[string]interface{}{}

				switch v := trust.(type) {
				case *awstypes.TlsValidationContextTrustMemberAcm:
					mAcm := map[string]interface{}{
						"certificate_authority_arns": flex.FlattenStringValueSet(v.Value.CertificateAuthorityArns),
					}

					mTrust["acm"] = []interface{}{mAcm}
				case *awstypes.TlsValidationContextTrustMemberFile:
					mFile := map[string]interface{}{
						"certificate_chain": aws.ToString(v.Value.CertificateChain),
					}

					mTrust["file"] = []interface{}{mFile}
				case *awstypes.TlsValidationContextTrustMemberSds:
					mSds := map[string]interface{}{
						"secret_name": aws.ToString(v.Value.SecretName),
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

func flattenDuration(duration *awstypes.Duration) []interface{} {
	if duration == nil {
		return []interface{}{}
	}

	mDuration := map[string]interface{}{
		"unit":  string(duration.Unit),
		"value": int(aws.ToInt64(duration.Value)),
	}

	return []interface{}{mDuration}
}

func flattenGRPCRoute(grpcRoute *awstypes.GrpcRoute) []interface{} {
	if grpcRoute == nil {
		return []interface{}{}
	}

	mGrpcRoute := map[string]interface{}{}

	if action := grpcRoute.Action; action != nil {
		if weightedTargets := action.WeightedTargets; weightedTargets != nil {
			vWeightedTargets := []interface{}{}

			for _, weightedTarget := range weightedTargets {
				mWeightedTarget := map[string]interface{}{
					"virtual_node": aws.ToString(weightedTarget.VirtualNode),
					"weight":       int(weightedTarget.Weight),
					"port":         int(aws.ToInt32(weightedTarget.Port)),
				}

				vWeightedTargets = append(vWeightedTargets, mWeightedTarget)
			}

			mGrpcRoute["action"] = []interface{}{
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
				"invert": aws.ToBool(grpcRouteMetadata.Invert),
				"name":   aws.ToString(grpcRouteMetadata.Name),
			}

			mMatch := map[string]interface{}{}

			switch v := grpcRouteMetadata.Match.(type) {
			case *awstypes.GrpcRouteMetadataMatchMethodMemberExact:
				mMatch["exact"] = v.Value
			case *awstypes.GrpcRouteMetadataMatchMethodMemberPrefix:
				mMatch["prefix"] = v.Value
			case *awstypes.GrpcRouteMetadataMatchMethodMemberRegex:
				mMatch["regex"] = v.Value
			case *awstypes.GrpcRouteMetadataMatchMethodMemberSuffix:
				mMatch["suffix"] = v.Value
			case *awstypes.GrpcRouteMetadataMatchMethodMemberRange:
				mRange := map[string]interface{}{
					"end":   int(aws.ToInt64(v.Value.End)),
					"start": int(aws.ToInt64(v.Value.Start)),
				}

				mMatch["range"] = []interface{}{mRange}
			}

			mGrpcRouteMetadata["match"] = []interface{}{mMatch}

			vGrpcRouteMetadatas = append(vGrpcRouteMetadatas, mGrpcRouteMetadata)
		}

		mGrpcRoute["match"] = []interface{}{
			map[string]interface{}{
				"metadata":     vGrpcRouteMetadatas,
				"method_name":  aws.ToString(grpcRouteMatch.MethodName),
				"service_name": aws.ToString(grpcRouteMatch.ServiceName),
				"port":         int(aws.ToInt32(grpcRouteMatch.Port)),
			},
		}
	}

	if grpcRetryPolicy := grpcRoute.RetryPolicy; grpcRetryPolicy != nil {
		mGrpcRetryPolicy := map[string]interface{}{
			"grpc_retry_events": flex.FlattenStringyValueSet(grpcRetryPolicy.GrpcRetryEvents),
			"http_retry_events": flex.FlattenStringValueSet(grpcRetryPolicy.HttpRetryEvents),
			"max_retries":       int(aws.ToInt64(grpcRetryPolicy.MaxRetries)),
			"per_retry_timeout": flattenDuration(grpcRetryPolicy.PerRetryTimeout),
			"tcp_retry_events":  flex.FlattenStringyValueSet(grpcRetryPolicy.TcpRetryEvents),
		}

		mGrpcRoute["retry_policy"] = []interface{}{mGrpcRetryPolicy}
	}

	mGrpcRoute["timeout"] = flattenGRPCTimeout(grpcRoute.Timeout)

	return []interface{}{mGrpcRoute}
}

func flattenGRPCTimeout(grpcTimeout *awstypes.GrpcTimeout) []interface{} {
	if grpcTimeout == nil {
		return []interface{}{}
	}

	mGrpcTimeout := map[string]interface{}{
		"idle":        flattenDuration(grpcTimeout.Idle),
		"per_request": flattenDuration(grpcTimeout.PerRequest),
	}

	return []interface{}{mGrpcTimeout}
}

func flattenHTTPRoute(httpRoute *awstypes.HttpRoute) []interface{} {
	if httpRoute == nil {
		return []interface{}{}
	}

	mHttpRoute := map[string]interface{}{}

	if action := httpRoute.Action; action != nil {
		if weightedTargets := action.WeightedTargets; weightedTargets != nil {
			vWeightedTargets := []interface{}{}

			for _, weightedTarget := range weightedTargets {
				mWeightedTarget := map[string]interface{}{
					"virtual_node": aws.ToString(weightedTarget.VirtualNode),
					"weight":       int(weightedTarget.Weight),
					"port":         int(aws.ToInt32(weightedTarget.Port)),
				}

				vWeightedTargets = append(vWeightedTargets, mWeightedTarget)
			}

			mHttpRoute["action"] = []interface{}{
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
				"invert": aws.ToBool(httpRouteHeader.Invert),
				"name":   aws.ToString(httpRouteHeader.Name),
			}

			mMatch := map[string]interface{}{}

			if match := httpRouteHeader.Match; match != nil {
				switch v := httpRouteHeader.Match.(type) {
				case *awstypes.HeaderMatchMethodMemberExact:
					mMatch["exact"] = v.Value
				case *awstypes.HeaderMatchMethodMemberPrefix:
					mMatch["prefix"] = v.Value
				case *awstypes.HeaderMatchMethodMemberRegex:
					mMatch["regex"] = v.Value
				case *awstypes.HeaderMatchMethodMemberSuffix:
					mMatch["suffix"] = v.Value
				case *awstypes.HeaderMatchMethodMemberRange:
					mRange := map[string]interface{}{
						"end":   int(aws.ToInt64(v.Value.End)),
						"start": int(aws.ToInt64(v.Value.Start)),
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
				"exact": aws.ToString(httpRoutePath.Exact),
				"regex": aws.ToString(httpRoutePath.Regex),
			}

			vHttpRoutePath = []interface{}{mHttpRoutePath}
		}

		vHttpRouteQueryParameters := []interface{}{}

		for _, httpRouteQueryParameter := range httpRouteMatch.QueryParameters {
			mHttpRouteQueryParameter := map[string]interface{}{
				"name": aws.ToString(httpRouteQueryParameter.Name),
			}

			if match := httpRouteQueryParameter.Match; match != nil {
				mMatch := map[string]interface{}{
					"exact": aws.ToString(match.Exact),
				}

				mHttpRouteQueryParameter["match"] = []interface{}{mMatch}
			}

			vHttpRouteQueryParameters = append(vHttpRouteQueryParameters, mHttpRouteQueryParameter)
		}

		mHttpRoute["match"] = []interface{}{
			map[string]interface{}{
				"header":          vHttpRouteHeaders,
				"method":          string(httpRouteMatch.Method),
				"path":            vHttpRoutePath,
				"port":            int(aws.ToInt32(httpRouteMatch.Port)),
				"prefix":          aws.ToString(httpRouteMatch.Prefix),
				"query_parameter": vHttpRouteQueryParameters,
				"scheme":          string(httpRouteMatch.Scheme),
			},
		}
	}

	if httpRetryPolicy := httpRoute.RetryPolicy; httpRetryPolicy != nil {
		mHttpRetryPolicy := map[string]interface{}{
			"http_retry_events": flex.FlattenStringValueSet(httpRetryPolicy.HttpRetryEvents),
			"max_retries":       int(aws.ToInt64(httpRetryPolicy.MaxRetries)),
			"per_retry_timeout": flattenDuration(httpRetryPolicy.PerRetryTimeout),
			"tcp_retry_events":  flex.FlattenStringyValueSet(httpRetryPolicy.TcpRetryEvents),
		}

		mHttpRoute["retry_policy"] = []interface{}{mHttpRetryPolicy}
	}

	mHttpRoute["timeout"] = flattenHTTPTimeout(httpRoute.Timeout)

	return []interface{}{mHttpRoute}
}

func flattenHTTPTimeout(httpTimeout *awstypes.HttpTimeout) []interface{} {
	if httpTimeout == nil {
		return []interface{}{}
	}

	mHttpTimeout := map[string]interface{}{
		"idle":        flattenDuration(httpTimeout.Idle),
		"per_request": flattenDuration(httpTimeout.PerRequest),
	}

	return []interface{}{mHttpTimeout}
}

func flattenMeshSpec(spec *awstypes.MeshSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if spec.EgressFilter != nil {
		mSpec["egress_filter"] = []interface{}{
			map[string]interface{}{
				"type": string(spec.EgressFilter.Type),
			},
		}
	}

	if spec.ServiceDiscovery != nil {
		mSpec["service_discovery"] = []interface{}{
			map[string]interface{}{
				"ip_preference": string(spec.ServiceDiscovery.IpPreference),
			},
		}
	}

	return []interface{}{mSpec}
}

func flattenRouteSpec(spec *awstypes.RouteSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{
		"grpc_route":  flattenGRPCRoute(spec.GrpcRoute),
		"http2_route": flattenHTTPRoute(spec.Http2Route),
		"http_route":  flattenHTTPRoute(spec.HttpRoute),
		"priority":    int(aws.ToInt32(spec.Priority)),
		"tcp_route":   flattenTCPRoute(spec.TcpRoute),
	}

	return []interface{}{mSpec}
}

func flattenTCPRoute(tcpRoute *awstypes.TcpRoute) []interface{} {
	if tcpRoute == nil {
		return []interface{}{}
	}

	mTcpRoute := map[string]interface{}{}

	if action := tcpRoute.Action; action != nil {
		if weightedTargets := action.WeightedTargets; weightedTargets != nil {
			vWeightedTargets := []interface{}{}

			for _, weightedTarget := range weightedTargets {
				mWeightedTarget := map[string]interface{}{
					"virtual_node": aws.ToString(weightedTarget.VirtualNode),
					"weight":       int(weightedTarget.Weight),
					"port":         int(aws.ToInt32(weightedTarget.Port)),
				}

				vWeightedTargets = append(vWeightedTargets, mWeightedTarget)
			}

			mTcpRoute["action"] = []interface{}{
				map[string]interface{}{
					"weighted_target": vWeightedTargets,
				},
			}
		}
	}

	if tcpRouteMatch := tcpRoute.Match; tcpRouteMatch != nil {
		mTcpRoute["match"] = []interface{}{
			map[string]interface{}{
				"port": int(aws.ToInt32(tcpRouteMatch.Port)),
			},
		}
	}

	mTcpRoute["timeout"] = flattenTCPTimeout(tcpRoute.Timeout)

	return []interface{}{mTcpRoute}
}

func flattenTCPTimeout(tcpTimeout *awstypes.TcpTimeout) []interface{} {
	if tcpTimeout == nil {
		return []interface{}{}
	}

	mTcpTimeout := map[string]interface{}{
		"idle": flattenDuration(tcpTimeout.Idle),
	}

	return []interface{}{mTcpTimeout}
}

func flattenVirtualNodeSpec(spec *awstypes.VirtualNodeSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if backends := spec.Backends; backends != nil {
		vBackends := []interface{}{}

		for _, backend := range backends {
			mBackend := map[string]interface{}{}

			switch v := backend.(type) {
			case *awstypes.BackendMemberVirtualService:
				mVirtualService := map[string]interface{}{
					"client_policy":        flattenClientPolicy(v.Value.ClientPolicy),
					"virtual_service_name": aws.ToString(v.Value.VirtualServiceName),
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

				switch v := connectionPool.(type) {
				case *awstypes.VirtualNodeConnectionPoolMemberGrpc:
					mGrpcConnectionPool := map[string]interface{}{
						"max_requests": int(aws.ToInt32(v.Value.MaxRequests)),
					}
					mConnectionPool["grpc"] = []interface{}{mGrpcConnectionPool}
				case *awstypes.VirtualNodeConnectionPoolMemberHttp:
					mHttpConnectionPool := map[string]interface{}{
						"max_connections":      int(aws.ToInt32(v.Value.MaxConnections)),
						"max_pending_requests": int(aws.ToInt32(v.Value.MaxPendingRequests)),
					}
					mConnectionPool["http"] = []interface{}{mHttpConnectionPool}
				case *awstypes.VirtualNodeConnectionPoolMemberHttp2:
					mHttp2ConnectionPool := map[string]interface{}{
						"max_requests": int(aws.ToInt32(v.Value.MaxRequests)),
					}
					mConnectionPool["http2"] = []interface{}{mHttp2ConnectionPool}
				case *awstypes.VirtualNodeConnectionPoolMemberTcp:
					mTcpConnectionPool := map[string]interface{}{
						"max_connections": int(aws.ToInt32(v.Value.MaxConnections)),
					}
					mConnectionPool["tcp"] = []interface{}{mTcpConnectionPool}
				}

				mListener["connection_pool"] = []interface{}{mConnectionPool}
			}

			if healthCheck := listener.HealthCheck; healthCheck != nil {
				mHealthCheck := map[string]interface{}{
					"healthy_threshold":   int(aws.ToInt32(healthCheck.HealthyThreshold)),
					"interval_millis":     int(aws.ToInt64(healthCheck.IntervalMillis)),
					"path":                aws.ToString(healthCheck.Path),
					"port":                int(aws.ToInt32(healthCheck.Port)),
					"protocol":            string(healthCheck.Protocol),
					"timeout_millis":      int(aws.ToInt64(healthCheck.TimeoutMillis)),
					"unhealthy_threshold": int(aws.ToInt32(healthCheck.UnhealthyThreshold)),
				}
				mListener["health_check"] = []interface{}{mHealthCheck}
			}

			if outlierDetection := listener.OutlierDetection; outlierDetection != nil {
				mOutlierDetection := map[string]interface{}{
					"base_ejection_duration": flattenDuration(outlierDetection.BaseEjectionDuration),
					"interval":               flattenDuration(outlierDetection.Interval),
					"max_ejection_percent":   int(aws.ToInt32(outlierDetection.MaxEjectionPercent)),
					"max_server_errors":      int(aws.ToInt64(outlierDetection.MaxServerErrors)),
				}
				mListener["outlier_detection"] = []interface{}{mOutlierDetection}
			}

			if portMapping := listener.PortMapping; portMapping != nil {
				mPortMapping := map[string]interface{}{
					"port":     int(aws.ToInt32(portMapping.Port)),
					"protocol": string(portMapping.Protocol),
				}
				mListener["port_mapping"] = []interface{}{mPortMapping}
			}

			if listenerTimeout := listener.Timeout; listenerTimeout != nil {
				mListenerTimeout := map[string]interface{}{}

				switch v := listenerTimeout.(type) {
				case *awstypes.ListenerTimeoutMemberGrpc:
					mListenerTimeout["grpc"] = flattenGRPCTimeout(&v.Value)
				case *awstypes.ListenerTimeoutMemberHttp:
					mListenerTimeout["http"] = flattenHTTPTimeout(&v.Value)
				case *awstypes.ListenerTimeoutMemberHttp2:
					mListenerTimeout["http2"] = flattenHTTPTimeout(&v.Value)
				case *awstypes.ListenerTimeoutMemberTcp:
					mListenerTimeout["tcp"] = flattenTCPTimeout(&v.Value)
				}
				mListener["timeout"] = []interface{}{mListenerTimeout}
			}

			if tls := listener.Tls; tls != nil {
				mTls := map[string]interface{}{
					"mode": string(tls.Mode),
				}

				if certificate := tls.Certificate; certificate != nil {
					mCertificate := map[string]interface{}{}

					switch v := certificate.(type) {
					case *awstypes.ListenerTlsCertificateMemberAcm:
						mAcm := map[string]interface{}{
							"certificate_arn": aws.ToString(v.Value.CertificateArn),
						}

						mCertificate["acm"] = []interface{}{mAcm}
					case *awstypes.ListenerTlsCertificateMemberFile:
						mFile := map[string]interface{}{
							"certificate_chain": aws.ToString(v.Value.CertificateChain),
							"private_key":       aws.ToString(v.Value.PrivateKey),
						}

						mCertificate["file"] = []interface{}{mFile}
					case *awstypes.ListenerTlsCertificateMemberSds:
						mSds := map[string]interface{}{
							"secret_name": aws.ToString(v.Value.SecretName),
						}

						mCertificate["sds"] = []interface{}{mSds}
					}

					mTls["certificate"] = []interface{}{mCertificate}
				}

				if validation := tls.Validation; validation != nil {
					mValidation := map[string]interface{}{}

					if subjectAlternativeNames := validation.SubjectAlternativeNames; subjectAlternativeNames != nil {
						mSubjectAlternativeNames := map[string]interface{}{}

						if match := subjectAlternativeNames.Match; match != nil {
							mMatch := map[string]interface{}{
								"exact": flex.FlattenStringValueSet(match.Exact),
							}

							mSubjectAlternativeNames["match"] = []interface{}{mMatch}
						}

						mValidation["subject_alternative_names"] = []interface{}{mSubjectAlternativeNames}
					}

					if trust := validation.Trust; trust != nil {
						mTrust := map[string]interface{}{}

						switch v := trust.(type) {
						case *awstypes.ListenerTlsValidationContextTrustMemberFile:
							mFile := map[string]interface{}{
								"certificate_chain": aws.ToString(v.Value.CertificateChain),
							}

							mTrust["file"] = []interface{}{mFile}
						case *awstypes.ListenerTlsValidationContextTrustMemberSds:
							mSds := map[string]interface{}{
								"secret_name": aws.ToString(v.Value.SecretName),
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

			switch v := accessLog.(type) {
			case *awstypes.AccessLogMemberFile:
				mFile := map[string]interface{}{}

				if format := v.Value.Format; format != nil {
					mFormat := map[string]interface{}{}

					switch v := format.(type) {
					case *awstypes.LoggingFormatMemberJson:
						vJsons := []interface{}{}

						for _, j := range v.Value {
							mJson := map[string]interface{}{
								"key":   aws.ToString(j.Key),
								"value": aws.ToString(j.Value),
							}

							vJsons = append(vJsons, mJson)
						}

						mFormat["json"] = vJsons
					case *awstypes.LoggingFormatMemberText:
						mFormat["text"] = v.Value
					}

					mFile["format"] = []interface{}{mFormat}
				}

				mFile["path"] = aws.ToString(v.Value.Path)

				mAccessLog["file"] = []interface{}{mFile}
			}

			mLogging["access_log"] = []interface{}{mAccessLog}
		}

		mSpec["logging"] = []interface{}{mLogging}
	}

	if serviceDiscovery := spec.ServiceDiscovery; serviceDiscovery != nil {
		mServiceDiscovery := map[string]interface{}{}

		switch v := serviceDiscovery.(type) {
		case *awstypes.ServiceDiscoveryMemberAwsCloudMap:
			vAttributes := map[string]interface{}{}

			for _, attribute := range v.Value.Attributes {
				vAttributes[aws.ToString(attribute.Key)] = aws.ToString(attribute.Value)
			}

			mServiceDiscovery["aws_cloud_map"] = []interface{}{
				map[string]interface{}{
					"attributes":     vAttributes,
					"namespace_name": aws.ToString(v.Value.NamespaceName),
					"service_name":   aws.ToString(v.Value.ServiceName),
				},
			}
		case *awstypes.ServiceDiscoveryMemberDns:
			mServiceDiscovery["dns"] = []interface{}{
				map[string]interface{}{
					"hostname":      aws.ToString(v.Value.Hostname),
					"ip_preference": string(v.Value.IpPreference),
					"response_type": string(v.Value.ResponseType),
				},
			}
		}

		mSpec["service_discovery"] = []interface{}{mServiceDiscovery}
	}

	return []interface{}{mSpec}
}

func flattenVirtualRouterSpec(spec *awstypes.VirtualRouterSpec) []interface{} {
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
					"port":     int(aws.ToInt32(listener.PortMapping.Port)),
					"protocol": string(listener.PortMapping.Protocol),
				}
				mListener["port_mapping"] = []interface{}{mPortMapping}
			}
			mListeners = append(mListeners, mListener)
		}
		mSpec["listener"] = mListeners
	}

	return []interface{}{mSpec}
}

func flattenVirtualServiceSpec(spec *awstypes.VirtualServiceSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if spec.Provider != nil {
		mProvider := map[string]interface{}{}

		switch v := spec.Provider.(type) {
		case *awstypes.VirtualServiceProviderMemberVirtualNode:
			mProvider["virtual_node"] = []interface{}{
				map[string]interface{}{
					"virtual_node_name": aws.ToString(v.Value.VirtualNodeName),
				},
			}
		case *awstypes.VirtualServiceProviderMemberVirtualRouter:
			mProvider["virtual_router"] = []interface{}{
				map[string]interface{}{
					"virtual_router_name": aws.ToString(v.Value.VirtualRouterName),
				},
			}
		}

		mSpec["provider"] = []interface{}{mProvider}
	}

	return []interface{}{mSpec}
}
