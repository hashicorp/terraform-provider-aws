// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appmesh/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandClientPolicy(vClientPolicy []any) *awstypes.ClientPolicy {
	if len(vClientPolicy) == 0 || vClientPolicy[0] == nil {
		return nil
	}

	clientPolicy := &awstypes.ClientPolicy{}

	mClientPolicy := vClientPolicy[0].(map[string]any)

	if vTls, ok := mClientPolicy["tls"].([]any); ok && len(vTls) > 0 && vTls[0] != nil {
		tls := &awstypes.ClientPolicyTls{}

		mTls := vTls[0].(map[string]any)

		if vCertificate, ok := mTls[names.AttrCertificate].([]any); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
			mCertificate := vCertificate[0].(map[string]any)

			if vFile, ok := mCertificate["file"].([]any); ok && len(vFile) > 0 && vFile[0] != nil {
				certificate := &awstypes.ClientTlsCertificateMemberFile{}

				file := awstypes.ListenerTlsFileCertificate{}

				mFile := vFile[0].(map[string]any)

				if vCertificateChain, ok := mFile[names.AttrCertificateChain].(string); ok && vCertificateChain != "" {
					file.CertificateChain = aws.String(vCertificateChain)
				}
				if vPrivateKey, ok := mFile[names.AttrPrivateKey].(string); ok && vPrivateKey != "" {
					file.PrivateKey = aws.String(vPrivateKey)
				}

				certificate.Value = file
				tls.Certificate = certificate
			}

			if vSds, ok := mCertificate["sds"].([]any); ok && len(vSds) > 0 && vSds[0] != nil {
				certificate := &awstypes.ClientTlsCertificateMemberSds{}

				sds := awstypes.ListenerTlsSdsCertificate{}

				mSds := vSds[0].(map[string]any)

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

		if vValidation, ok := mTls["validation"].([]any); ok && len(vValidation) > 0 && vValidation[0] != nil {
			validation := &awstypes.TlsValidationContext{}

			mValidation := vValidation[0].(map[string]any)

			if vSubjectAlternativeNames, ok := mValidation["subject_alternative_names"].([]any); ok && len(vSubjectAlternativeNames) > 0 && vSubjectAlternativeNames[0] != nil {
				subjectAlternativeNames := &awstypes.SubjectAlternativeNames{}

				mSubjectAlternativeNames := vSubjectAlternativeNames[0].(map[string]any)

				if vMatch, ok := mSubjectAlternativeNames["match"].([]any); ok && len(vMatch) > 0 && vMatch[0] != nil {
					match := &awstypes.SubjectAlternativeNameMatchers{}

					mMatch := vMatch[0].(map[string]any)

					if vExact, ok := mMatch["exact"].(*schema.Set); ok && vExact.Len() > 0 {
						match.Exact = flex.ExpandStringValueSet(vExact)
					}

					subjectAlternativeNames.Match = match
				}

				validation.SubjectAlternativeNames = subjectAlternativeNames
			}

			if vTrust, ok := mValidation["trust"].([]any); ok && len(vTrust) > 0 && vTrust[0] != nil {
				mTrust := vTrust[0].(map[string]any)

				if vAcm, ok := mTrust["acm"].([]any); ok && len(vAcm) > 0 && vAcm[0] != nil {
					trust := &awstypes.TlsValidationContextTrustMemberAcm{}

					acm := awstypes.TlsValidationContextAcmTrust{}

					mAcm := vAcm[0].(map[string]any)

					if vCertificateAuthorityArns, ok := mAcm["certificate_authority_arns"].(*schema.Set); ok && vCertificateAuthorityArns.Len() > 0 {
						acm.CertificateAuthorityArns = flex.ExpandStringValueSet(vCertificateAuthorityArns)
					}

					trust.Value = acm
					validation.Trust = trust
				}

				if vFile, ok := mTrust["file"].([]any); ok && len(vFile) > 0 && vFile[0] != nil {
					trust := &awstypes.TlsValidationContextTrustMemberFile{}

					file := awstypes.TlsValidationContextFileTrust{}

					mFile := vFile[0].(map[string]any)

					if vCertificateChain, ok := mFile[names.AttrCertificateChain].(string); ok && vCertificateChain != "" {
						file.CertificateChain = aws.String(vCertificateChain)
					}

					trust.Value = file
					validation.Trust = trust
				}

				if vSds, ok := mTrust["sds"].([]any); ok && len(vSds) > 0 && vSds[0] != nil {
					trust := &awstypes.TlsValidationContextTrustMemberSds{}

					sds := awstypes.TlsValidationContextSdsTrust{}

					mSds := vSds[0].(map[string]any)

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

func expandDuration(vDuration []any) *awstypes.Duration {
	if len(vDuration) == 0 || vDuration[0] == nil {
		return nil
	}

	duration := &awstypes.Duration{}

	mDuration := vDuration[0].(map[string]any)

	if vUnit, ok := mDuration[names.AttrUnit].(string); ok && vUnit != "" {
		duration.Unit = awstypes.DurationUnit(vUnit)
	}
	if vValue, ok := mDuration[names.AttrValue].(int); ok && vValue > 0 {
		duration.Value = aws.Int64(int64(vValue))
	}

	return duration
}

func expandGRPCRoute(vGrpcRoute []any) *awstypes.GrpcRoute {
	if len(vGrpcRoute) == 0 || vGrpcRoute[0] == nil {
		return nil
	}

	mGrpcRoute := vGrpcRoute[0].(map[string]any)

	grpcRoute := &awstypes.GrpcRoute{}

	if vGrpcRouteAction, ok := mGrpcRoute[names.AttrAction].([]any); ok && len(vGrpcRouteAction) > 0 && vGrpcRouteAction[0] != nil {
		mGrpcRouteAction := vGrpcRouteAction[0].(map[string]any)

		if vWeightedTargets, ok := mGrpcRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
			weightedTargets := []awstypes.WeightedTarget{}

			for _, vWeightedTarget := range vWeightedTargets.List() {
				weightedTarget := awstypes.WeightedTarget{}

				mWeightedTarget := vWeightedTarget.(map[string]any)

				if vPort, ok := mWeightedTarget[names.AttrPort].(int); ok && vPort > 0 {
					weightedTarget.Port = aws.Int32(int32(vPort))
				}
				if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
					weightedTarget.VirtualNode = aws.String(vVirtualNode)
				}
				if vWeight, ok := mWeightedTarget[names.AttrWeight].(int); ok {
					weightedTarget.Weight = int32(vWeight)
				}

				weightedTargets = append(weightedTargets, weightedTarget)
			}

			grpcRoute.Action = &awstypes.GrpcRouteAction{
				WeightedTargets: weightedTargets,
			}
		}
	}

	if vGrpcRouteMatch, ok := mGrpcRoute["match"].([]any); ok {
		grpcRouteMatch := &awstypes.GrpcRouteMatch{}

		// Empty match is allowed.
		// https://github.com/hashicorp/terraform-provider-aws/issues/16816.

		if len(vGrpcRouteMatch) > 0 && vGrpcRouteMatch[0] != nil {
			mGrpcRouteMatch := vGrpcRouteMatch[0].(map[string]any)

			if vMethodName, ok := mGrpcRouteMatch["method_name"].(string); ok && vMethodName != "" {
				grpcRouteMatch.MethodName = aws.String(vMethodName)
			}
			if vPort, ok := mGrpcRouteMatch[names.AttrPort].(int); ok && vPort > 0 {
				grpcRouteMatch.Port = aws.Int32(int32(vPort))
			}
			if vServiceName, ok := mGrpcRouteMatch[names.AttrServiceName].(string); ok && vServiceName != "" {
				grpcRouteMatch.ServiceName = aws.String(vServiceName)
			}

			if vGrpcRouteMetadatas, ok := mGrpcRouteMatch["metadata"].(*schema.Set); ok && vGrpcRouteMetadatas.Len() > 0 {
				grpcRouteMetadatas := []awstypes.GrpcRouteMetadata{}

				for _, vGrpcRouteMetadata := range vGrpcRouteMetadatas.List() {
					grpcRouteMetadata := awstypes.GrpcRouteMetadata{}

					mGrpcRouteMetadata := vGrpcRouteMetadata.(map[string]any)

					if vInvert, ok := mGrpcRouteMetadata["invert"].(bool); ok {
						grpcRouteMetadata.Invert = aws.Bool(vInvert)
					}
					if vName, ok := mGrpcRouteMetadata[names.AttrName].(string); ok && vName != "" {
						grpcRouteMetadata.Name = aws.String(vName)
					}

					if vMatch, ok := mGrpcRouteMetadata["match"].([]any); ok && len(vMatch) > 0 && vMatch[0] != nil {
						mMatch := vMatch[0].(map[string]any)

						if vExact, ok := mMatch["exact"].(string); ok && vExact != "" {
							grpcRouteMetadata.Match = &awstypes.GrpcRouteMetadataMatchMethodMemberExact{Value: vExact}
						}
						if vPrefix, ok := mMatch[names.AttrPrefix].(string); ok && vPrefix != "" {
							grpcRouteMetadata.Match = &awstypes.GrpcRouteMetadataMatchMethodMemberPrefix{Value: vPrefix}
						}
						if vRegex, ok := mMatch["regex"].(string); ok && vRegex != "" {
							grpcRouteMetadata.Match = &awstypes.GrpcRouteMetadataMatchMethodMemberRegex{Value: vRegex}
						}
						if vSuffix, ok := mMatch["suffix"].(string); ok && vSuffix != "" {
							grpcRouteMetadata.Match = &awstypes.GrpcRouteMetadataMatchMethodMemberSuffix{Value: vSuffix}
						}

						if vRange, ok := mMatch["range"].([]any); ok && len(vRange) > 0 && vRange[0] != nil {
							memberRange := &awstypes.GrpcRouteMetadataMatchMethodMemberRange{}

							mRange := vRange[0].(map[string]any)

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

	if vGrpcRetryPolicy, ok := mGrpcRoute["retry_policy"].([]any); ok && len(vGrpcRetryPolicy) > 0 && vGrpcRetryPolicy[0] != nil {
		grpcRetryPolicy := &awstypes.GrpcRetryPolicy{}

		mGrpcRetryPolicy := vGrpcRetryPolicy[0].(map[string]any)

		if vMaxRetries, ok := mGrpcRetryPolicy["max_retries"].(int); ok {
			grpcRetryPolicy.MaxRetries = aws.Int64(int64(vMaxRetries))
		}

		if vGrpcRetryEvents, ok := mGrpcRetryPolicy["grpc_retry_events"].(*schema.Set); ok && vGrpcRetryEvents.Len() > 0 {
			grpcRetryPolicy.GrpcRetryEvents = flex.ExpandStringyValueSet[awstypes.GrpcRetryPolicyEvent](vGrpcRetryEvents)
		}

		if vHttpRetryEvents, ok := mGrpcRetryPolicy["http_retry_events"].(*schema.Set); ok && vHttpRetryEvents.Len() > 0 {
			grpcRetryPolicy.HttpRetryEvents = flex.ExpandStringValueSet(vHttpRetryEvents)
		}

		if vPerRetryTimeout, ok := mGrpcRetryPolicy["per_retry_timeout"].([]any); ok {
			grpcRetryPolicy.PerRetryTimeout = expandDuration(vPerRetryTimeout)
		}

		if vTcpRetryEvents, ok := mGrpcRetryPolicy["tcp_retry_events"].(*schema.Set); ok && vTcpRetryEvents.Len() > 0 {
			grpcRetryPolicy.TcpRetryEvents = flex.ExpandStringyValueSet[awstypes.TcpRetryPolicyEvent](vTcpRetryEvents)
		}

		grpcRoute.RetryPolicy = grpcRetryPolicy
	}

	if vGrpcTimeout, ok := mGrpcRoute[names.AttrTimeout].([]any); ok {
		grpcRoute.Timeout = expandGRPCTimeout(vGrpcTimeout)
	}

	return grpcRoute
}

func expandGRPCTimeout(vGrpcTimeout []any) *awstypes.GrpcTimeout {
	if len(vGrpcTimeout) == 0 || vGrpcTimeout[0] == nil {
		return nil
	}

	grpcTimeout := &awstypes.GrpcTimeout{}

	mGrpcTimeout := vGrpcTimeout[0].(map[string]any)

	if vIdleTimeout, ok := mGrpcTimeout["idle"].([]any); ok {
		grpcTimeout.Idle = expandDuration(vIdleTimeout)
	}

	if vPerRequestTimeout, ok := mGrpcTimeout["per_request"].([]any); ok {
		grpcTimeout.PerRequest = expandDuration(vPerRequestTimeout)
	}

	return grpcTimeout
}

func expandHTTPRoute(vHttpRoute []any) *awstypes.HttpRoute {
	if len(vHttpRoute) == 0 || vHttpRoute[0] == nil {
		return nil
	}

	mHttpRoute := vHttpRoute[0].(map[string]any)

	httpRoute := &awstypes.HttpRoute{}

	if vHttpRouteAction, ok := mHttpRoute[names.AttrAction].([]any); ok && len(vHttpRouteAction) > 0 && vHttpRouteAction[0] != nil {
		mHttpRouteAction := vHttpRouteAction[0].(map[string]any)

		if vWeightedTargets, ok := mHttpRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
			weightedTargets := []awstypes.WeightedTarget{}

			for _, vWeightedTarget := range vWeightedTargets.List() {
				weightedTarget := awstypes.WeightedTarget{}

				mWeightedTarget := vWeightedTarget.(map[string]any)

				if vPort, ok := mWeightedTarget[names.AttrPort].(int); ok && vPort > 0 {
					weightedTarget.Port = aws.Int32(int32(vPort))
				}
				if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
					weightedTarget.VirtualNode = aws.String(vVirtualNode)
				}
				if vWeight, ok := mWeightedTarget[names.AttrWeight].(int); ok {
					weightedTarget.Weight = int32(vWeight)
				}

				weightedTargets = append(weightedTargets, weightedTarget)
			}

			httpRoute.Action = &awstypes.HttpRouteAction{
				WeightedTargets: weightedTargets,
			}
		}
	}

	if vHttpRouteMatch, ok := mHttpRoute["match"].([]any); ok && len(vHttpRouteMatch) > 0 && vHttpRouteMatch[0] != nil {
		httpRouteMatch := &awstypes.HttpRouteMatch{}

		mHttpRouteMatch := vHttpRouteMatch[0].(map[string]any)

		if vMethod, ok := mHttpRouteMatch["method"].(string); ok && vMethod != "" {
			httpRouteMatch.Method = awstypes.HttpMethod(vMethod)
		}
		if vPort, ok := mHttpRouteMatch[names.AttrPort].(int); ok && vPort > 0 {
			httpRouteMatch.Port = aws.Int32(int32(vPort))
		}
		if vPrefix, ok := mHttpRouteMatch[names.AttrPrefix].(string); ok && vPrefix != "" {
			httpRouteMatch.Prefix = aws.String(vPrefix)
		}
		if vScheme, ok := mHttpRouteMatch["scheme"].(string); ok && vScheme != "" {
			httpRouteMatch.Scheme = awstypes.HttpScheme(vScheme)
		}

		if vHttpRouteHeaders, ok := mHttpRouteMatch[names.AttrHeader].(*schema.Set); ok && vHttpRouteHeaders.Len() > 0 {
			httpRouteHeaders := []awstypes.HttpRouteHeader{}

			for _, vHttpRouteHeader := range vHttpRouteHeaders.List() {
				httpRouteHeader := awstypes.HttpRouteHeader{}

				mHttpRouteHeader := vHttpRouteHeader.(map[string]any)

				if vInvert, ok := mHttpRouteHeader["invert"].(bool); ok {
					httpRouteHeader.Invert = aws.Bool(vInvert)
				}
				if vName, ok := mHttpRouteHeader[names.AttrName].(string); ok && vName != "" {
					httpRouteHeader.Name = aws.String(vName)
				}

				if vMatch, ok := mHttpRouteHeader["match"].([]any); ok && len(vMatch) > 0 && vMatch[0] != nil {
					mMatch := vMatch[0].(map[string]any)

					if vExact, ok := mMatch["exact"].(string); ok && vExact != "" {
						httpRouteHeader.Match = &awstypes.HeaderMatchMethodMemberExact{Value: vExact}
					}
					if vPrefix, ok := mMatch[names.AttrPrefix].(string); ok && vPrefix != "" {
						httpRouteHeader.Match = &awstypes.HeaderMatchMethodMemberPrefix{Value: vPrefix}
					}
					if vRegex, ok := mMatch["regex"].(string); ok && vRegex != "" {
						httpRouteHeader.Match = &awstypes.HeaderMatchMethodMemberRegex{Value: vRegex}
					}
					if vSuffix, ok := mMatch["suffix"].(string); ok && vSuffix != "" {
						httpRouteHeader.Match = &awstypes.HeaderMatchMethodMemberSuffix{Value: vSuffix}
					}

					if vRange, ok := mMatch["range"].([]any); ok && len(vRange) > 0 && vRange[0] != nil {
						memberRange := &awstypes.HeaderMatchMethodMemberRange{}

						mRange := vRange[0].(map[string]any)

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

		if vHttpRoutePath, ok := mHttpRouteMatch[names.AttrPath].([]any); ok && len(vHttpRoutePath) > 0 && vHttpRoutePath[0] != nil {
			httpRoutePath := &awstypes.HttpPathMatch{}

			mHttpRoutePath := vHttpRoutePath[0].(map[string]any)

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

				mHttpRouteQueryParameter := vHttpRouteQueryParameter.(map[string]any)

				if vName, ok := mHttpRouteQueryParameter[names.AttrName].(string); ok && vName != "" {
					httpRouteQueryParameter.Name = aws.String(vName)
				}

				if vMatch, ok := mHttpRouteQueryParameter["match"].([]any); ok && len(vMatch) > 0 && vMatch[0] != nil {
					httpRouteQueryParameter.Match = &awstypes.QueryParameterMatch{}

					mMatch := vMatch[0].(map[string]any)

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

	if vHttpRetryPolicy, ok := mHttpRoute["retry_policy"].([]any); ok && len(vHttpRetryPolicy) > 0 && vHttpRetryPolicy[0] != nil {
		httpRetryPolicy := &awstypes.HttpRetryPolicy{}

		mHttpRetryPolicy := vHttpRetryPolicy[0].(map[string]any)

		if vMaxRetries, ok := mHttpRetryPolicy["max_retries"].(int); ok {
			httpRetryPolicy.MaxRetries = aws.Int64(int64(vMaxRetries))
		}

		if vHttpRetryEvents, ok := mHttpRetryPolicy["http_retry_events"].(*schema.Set); ok && vHttpRetryEvents.Len() > 0 {
			httpRetryPolicy.HttpRetryEvents = flex.ExpandStringValueSet(vHttpRetryEvents)
		}

		if vPerRetryTimeout, ok := mHttpRetryPolicy["per_retry_timeout"].([]any); ok {
			httpRetryPolicy.PerRetryTimeout = expandDuration(vPerRetryTimeout)
		}

		if vTcpRetryEvents, ok := mHttpRetryPolicy["tcp_retry_events"].(*schema.Set); ok && vTcpRetryEvents.Len() > 0 {
			httpRetryPolicy.TcpRetryEvents = flex.ExpandStringyValueSet[awstypes.TcpRetryPolicyEvent](vTcpRetryEvents)
		}

		httpRoute.RetryPolicy = httpRetryPolicy
	}

	if vHttpTimeout, ok := mHttpRoute[names.AttrTimeout].([]any); ok {
		httpRoute.Timeout = expandHTTPTimeout(vHttpTimeout)
	}

	return httpRoute
}

func expandHTTPTimeout(vHttpTimeout []any) *awstypes.HttpTimeout {
	if len(vHttpTimeout) == 0 || vHttpTimeout[0] == nil {
		return nil
	}

	httpTimeout := &awstypes.HttpTimeout{}

	mHttpTimeout := vHttpTimeout[0].(map[string]any)

	if vIdleTimeout, ok := mHttpTimeout["idle"].([]any); ok {
		httpTimeout.Idle = expandDuration(vIdleTimeout)
	}

	if vPerRequestTimeout, ok := mHttpTimeout["per_request"].([]any); ok {
		httpTimeout.PerRequest = expandDuration(vPerRequestTimeout)
	}

	return httpTimeout
}

func expandMeshSpec(vSpec []any) *awstypes.MeshSpec {
	spec := &awstypes.MeshSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]any)

	if vEgressFilter, ok := mSpec["egress_filter"].([]any); ok && len(vEgressFilter) > 0 && vEgressFilter[0] != nil {
		mEgressFilter := vEgressFilter[0].(map[string]any)

		if vType, ok := mEgressFilter[names.AttrType].(string); ok && vType != "" {
			spec.EgressFilter = &awstypes.EgressFilter{
				Type: awstypes.EgressFilterType(vType),
			}
		}
	}

	if vServiceDiscovery, ok := mSpec["service_discovery"].([]any); ok && len(vServiceDiscovery) > 0 && vServiceDiscovery[0] != nil {
		mServiceDiscovery := vServiceDiscovery[0].(map[string]any)

		if vIpPreference, ok := mServiceDiscovery["ip_preference"].(string); ok && vIpPreference != "" {
			spec.ServiceDiscovery = &awstypes.MeshServiceDiscovery{
				IpPreference: awstypes.IpPreference(vIpPreference),
			}
		}
	}

	return spec
}

func expandRouteSpec(vSpec []any) *awstypes.RouteSpec {
	spec := &awstypes.RouteSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]any)

	if vGrpcRoute, ok := mSpec["grpc_route"].([]any); ok {
		spec.GrpcRoute = expandGRPCRoute(vGrpcRoute)
	}

	if vHttp2Route, ok := mSpec["http2_route"].([]any); ok {
		spec.Http2Route = expandHTTPRoute(vHttp2Route)
	}

	if vHttpRoute, ok := mSpec["http_route"].([]any); ok {
		spec.HttpRoute = expandHTTPRoute(vHttpRoute)
	}

	if vPriority, ok := mSpec[names.AttrPriority].(int); ok && vPriority > 0 {
		spec.Priority = aws.Int32(int32(vPriority))
	}

	if vTcpRoute, ok := mSpec["tcp_route"].([]any); ok {
		spec.TcpRoute = expandTCPRoute(vTcpRoute)
	}

	return spec
}

func expandTCPRoute(vTcpRoute []any) *awstypes.TcpRoute {
	if len(vTcpRoute) == 0 || vTcpRoute[0] == nil {
		return nil
	}

	mTcpRoute := vTcpRoute[0].(map[string]any)

	tcpRoute := &awstypes.TcpRoute{}

	if vTcpRouteAction, ok := mTcpRoute[names.AttrAction].([]any); ok && len(vTcpRouteAction) > 0 && vTcpRouteAction[0] != nil {
		mTcpRouteAction := vTcpRouteAction[0].(map[string]any)

		if vWeightedTargets, ok := mTcpRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
			weightedTargets := []awstypes.WeightedTarget{}

			for _, vWeightedTarget := range vWeightedTargets.List() {
				weightedTarget := awstypes.WeightedTarget{}

				mWeightedTarget := vWeightedTarget.(map[string]any)

				if vPort, ok := mWeightedTarget[names.AttrPort].(int); ok && vPort > 0 {
					weightedTarget.Port = aws.Int32(int32(vPort))
				}
				if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
					weightedTarget.VirtualNode = aws.String(vVirtualNode)
				}
				if vWeight, ok := mWeightedTarget[names.AttrWeight].(int); ok {
					weightedTarget.Weight = int32(vWeight)
				}

				weightedTargets = append(weightedTargets, weightedTarget)
			}

			tcpRoute.Action = &awstypes.TcpRouteAction{
				WeightedTargets: weightedTargets,
			}
		}
	}

	if vTcpRouteMatch, ok := mTcpRoute["match"].([]any); ok && len(vTcpRouteMatch) > 0 && vTcpRouteMatch[0] != nil {
		tcpRouteMatch := &awstypes.TcpRouteMatch{}

		mTcpRouteMatch := vTcpRouteMatch[0].(map[string]any)

		if vPort, ok := mTcpRouteMatch[names.AttrPort].(int); ok && vPort > 0 {
			tcpRouteMatch.Port = aws.Int32(int32(vPort))
		}

		tcpRoute.Match = tcpRouteMatch
	}

	if vTcpTimeout, ok := mTcpRoute[names.AttrTimeout].([]any); ok {
		tcpRoute.Timeout = expandTCPTimeout(vTcpTimeout)
	}

	return tcpRoute
}

func expandTCPTimeout(vTcpTimeout []any) *awstypes.TcpTimeout {
	if len(vTcpTimeout) == 0 || vTcpTimeout[0] == nil {
		return nil
	}

	tcpTimeout := &awstypes.TcpTimeout{}

	mTcpTimeout := vTcpTimeout[0].(map[string]any)

	if vIdleTimeout, ok := mTcpTimeout["idle"].([]any); ok {
		tcpTimeout.Idle = expandDuration(vIdleTimeout)
	}

	return tcpTimeout
}

func expandVirtualNodeSpec(vSpec []any) *awstypes.VirtualNodeSpec {
	spec := &awstypes.VirtualNodeSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]any)

	if vBackends, ok := mSpec["backend"].(*schema.Set); ok && vBackends.Len() > 0 {
		backends := []awstypes.Backend{}

		for _, vBackend := range vBackends.List() {
			backend := &awstypes.BackendMemberVirtualService{}
			mBackend := vBackend.(map[string]any)

			if vVirtualService, ok := mBackend["virtual_service"].([]any); ok && len(vVirtualService) > 0 && vVirtualService[0] != nil {
				virtualService := awstypes.VirtualServiceBackend{}

				mVirtualService := vVirtualService[0].(map[string]any)

				if vVirtualServiceName, ok := mVirtualService["virtual_service_name"].(string); ok {
					virtualService.VirtualServiceName = aws.String(vVirtualServiceName)
				}

				if vClientPolicy, ok := mVirtualService["client_policy"].([]any); ok {
					virtualService.ClientPolicy = expandClientPolicy(vClientPolicy)
				}

				backend.Value = virtualService
				backends = append(backends, backend)
			}
		}

		spec.Backends = backends
	}

	if vBackendDefaults, ok := mSpec["backend_defaults"].([]any); ok && len(vBackendDefaults) > 0 && vBackendDefaults[0] != nil {
		backendDefaults := &awstypes.BackendDefaults{}

		mBackendDefaults := vBackendDefaults[0].(map[string]any)

		if vClientPolicy, ok := mBackendDefaults["client_policy"].([]any); ok {
			backendDefaults.ClientPolicy = expandClientPolicy(vClientPolicy)
		}

		spec.BackendDefaults = backendDefaults
	}

	if vListeners, ok := mSpec["listener"].([]any); ok && len(vListeners) > 0 && vListeners[0] != nil {
		listeners := []awstypes.Listener{}

		for _, vListener := range vListeners {
			listener := awstypes.Listener{}

			mListener := vListener.(map[string]any)

			if vConnectionPool, ok := mListener["connection_pool"].([]any); ok && len(vConnectionPool) > 0 && vConnectionPool[0] != nil {
				mConnectionPool := vConnectionPool[0].(map[string]any)

				if vGrpcConnectionPool, ok := mConnectionPool["grpc"].([]any); ok && len(vGrpcConnectionPool) > 0 && vGrpcConnectionPool[0] != nil {
					connectionPool := &awstypes.VirtualNodeConnectionPoolMemberGrpc{}

					mGrpcConnectionPool := vGrpcConnectionPool[0].(map[string]any)

					grpcConnectionPool := awstypes.VirtualNodeGrpcConnectionPool{}

					if vMaxRequests, ok := mGrpcConnectionPool["max_requests"].(int); ok && vMaxRequests > 0 {
						grpcConnectionPool.MaxRequests = aws.Int32(int32(vMaxRequests))
					}

					connectionPool.Value = grpcConnectionPool
					listener.ConnectionPool = connectionPool
				}

				if vHttpConnectionPool, ok := mConnectionPool["http"].([]any); ok && len(vHttpConnectionPool) > 0 && vHttpConnectionPool[0] != nil {
					connectionPool := &awstypes.VirtualNodeConnectionPoolMemberHttp{}

					mHttpConnectionPool := vHttpConnectionPool[0].(map[string]any)

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

				if vHttp2ConnectionPool, ok := mConnectionPool["http2"].([]any); ok && len(vHttp2ConnectionPool) > 0 && vHttp2ConnectionPool[0] != nil {
					connectionPool := &awstypes.VirtualNodeConnectionPoolMemberHttp2{}

					mHttp2ConnectionPool := vHttp2ConnectionPool[0].(map[string]any)

					http2ConnectionPool := awstypes.VirtualNodeHttp2ConnectionPool{}

					if vMaxRequests, ok := mHttp2ConnectionPool["max_requests"].(int); ok && vMaxRequests > 0 {
						http2ConnectionPool.MaxRequests = aws.Int32(int32(vMaxRequests))
					}

					connectionPool.Value = http2ConnectionPool
					listener.ConnectionPool = connectionPool
				}

				if vTcpConnectionPool, ok := mConnectionPool["tcp"].([]any); ok && len(vTcpConnectionPool) > 0 && vTcpConnectionPool[0] != nil {
					connectionPool := &awstypes.VirtualNodeConnectionPoolMemberTcp{}

					mTcpConnectionPool := vTcpConnectionPool[0].(map[string]any)

					tcpConnectionPool := awstypes.VirtualNodeTcpConnectionPool{}

					if vMaxConnections, ok := mTcpConnectionPool["max_connections"].(int); ok && vMaxConnections > 0 {
						tcpConnectionPool.MaxConnections = aws.Int32(int32(vMaxConnections))
					}

					connectionPool.Value = tcpConnectionPool
					listener.ConnectionPool = connectionPool
				}
			}

			if vHealthCheck, ok := mListener[names.AttrHealthCheck].([]any); ok && len(vHealthCheck) > 0 && vHealthCheck[0] != nil {
				healthCheck := &awstypes.HealthCheckPolicy{}

				mHealthCheck := vHealthCheck[0].(map[string]any)

				if vHealthyThreshold, ok := mHealthCheck["healthy_threshold"].(int); ok && vHealthyThreshold > 0 {
					healthCheck.HealthyThreshold = aws.Int32(int32(vHealthyThreshold))
				}
				if vIntervalMillis, ok := mHealthCheck["interval_millis"].(int); ok && vIntervalMillis > 0 {
					healthCheck.IntervalMillis = aws.Int64(int64(vIntervalMillis))
				}
				if vPath, ok := mHealthCheck[names.AttrPath].(string); ok && vPath != "" {
					healthCheck.Path = aws.String(vPath)
				}
				if vPort, ok := mHealthCheck[names.AttrPort].(int); ok && vPort > 0 {
					healthCheck.Port = aws.Int32(int32(vPort))
				}
				if vProtocol, ok := mHealthCheck[names.AttrProtocol].(string); ok && vProtocol != "" {
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

			if vOutlierDetection, ok := mListener["outlier_detection"].([]any); ok && len(vOutlierDetection) > 0 && vOutlierDetection[0] != nil {
				outlierDetection := &awstypes.OutlierDetection{}

				mOutlierDetection := vOutlierDetection[0].(map[string]any)

				if vMaxEjectionPercent, ok := mOutlierDetection["max_ejection_percent"].(int); ok && vMaxEjectionPercent > 0 {
					outlierDetection.MaxEjectionPercent = aws.Int32(int32(vMaxEjectionPercent))
				}
				if vMaxServerErrors, ok := mOutlierDetection["max_server_errors"].(int); ok && vMaxServerErrors > 0 {
					outlierDetection.MaxServerErrors = aws.Int64(int64(vMaxServerErrors))
				}

				if vBaseEjectionDuration, ok := mOutlierDetection["base_ejection_duration"].([]any); ok {
					outlierDetection.BaseEjectionDuration = expandDuration(vBaseEjectionDuration)
				}

				if vInterval, ok := mOutlierDetection[names.AttrInterval].([]any); ok {
					outlierDetection.Interval = expandDuration(vInterval)
				}

				listener.OutlierDetection = outlierDetection
			}

			if vPortMapping, ok := mListener["port_mapping"].([]any); ok && len(vPortMapping) > 0 && vPortMapping[0] != nil {
				portMapping := &awstypes.PortMapping{}

				mPortMapping := vPortMapping[0].(map[string]any)

				if vPort, ok := mPortMapping[names.AttrPort].(int); ok && vPort > 0 {
					portMapping.Port = aws.Int32(int32(vPort))
				}
				if vProtocol, ok := mPortMapping[names.AttrProtocol].(string); ok && vProtocol != "" {
					portMapping.Protocol = awstypes.PortProtocol(vProtocol)
				}

				listener.PortMapping = portMapping
			}

			if vTimeout, ok := mListener[names.AttrTimeout].([]any); ok && len(vTimeout) > 0 && vTimeout[0] != nil {
				mTimeout := vTimeout[0].(map[string]any)

				if vGrpcTimeout, ok := mTimeout["grpc"].([]any); ok && len(vGrpcTimeout) > 0 && vGrpcTimeout[0] != nil {
					listener.Timeout = &awstypes.ListenerTimeoutMemberGrpc{Value: *expandGRPCTimeout(vGrpcTimeout)}
				}

				if vHttpTimeout, ok := mTimeout["http"].([]any); ok && len(vHttpTimeout) > 0 && vHttpTimeout[0] != nil {
					listener.Timeout = &awstypes.ListenerTimeoutMemberHttp{Value: *expandHTTPTimeout(vHttpTimeout)}
				}

				if vHttp2Timeout, ok := mTimeout["http2"].([]any); ok && len(vHttp2Timeout) > 0 && vHttp2Timeout[0] != nil {
					listener.Timeout = &awstypes.ListenerTimeoutMemberHttp2{Value: *expandHTTPTimeout(vHttp2Timeout)}
				}

				if vTcpTimeout, ok := mTimeout["tcp"].([]any); ok && len(vTcpTimeout) > 0 && vTcpTimeout[0] != nil {
					listener.Timeout = &awstypes.ListenerTimeoutMemberTcp{Value: *expandTCPTimeout(vTcpTimeout)}
				}
			}

			if vTls, ok := mListener["tls"].([]any); ok && len(vTls) > 0 && vTls[0] != nil {
				tls := &awstypes.ListenerTls{}

				mTls := vTls[0].(map[string]any)

				if vMode, ok := mTls[names.AttrMode].(string); ok && vMode != "" {
					tls.Mode = awstypes.ListenerTlsMode(vMode)
				}

				if vCertificate, ok := mTls[names.AttrCertificate].([]any); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
					mCertificate := vCertificate[0].(map[string]any)

					if vAcm, ok := mCertificate["acm"].([]any); ok && len(vAcm) > 0 && vAcm[0] != nil {
						certificate := &awstypes.ListenerTlsCertificateMemberAcm{}
						acm := awstypes.ListenerTlsAcmCertificate{}

						mAcm := vAcm[0].(map[string]any)

						if vCertificateArn, ok := mAcm[names.AttrCertificateARN].(string); ok && vCertificateArn != "" {
							acm.CertificateArn = aws.String(vCertificateArn)
						}

						certificate.Value = acm
						tls.Certificate = certificate
					}

					if vFile, ok := mCertificate["file"].([]any); ok && len(vFile) > 0 && vFile[0] != nil {
						certificate := &awstypes.ListenerTlsCertificateMemberFile{}

						file := awstypes.ListenerTlsFileCertificate{}

						mFile := vFile[0].(map[string]any)

						if vCertificateChain, ok := mFile[names.AttrCertificateChain].(string); ok && vCertificateChain != "" {
							file.CertificateChain = aws.String(vCertificateChain)
						}
						if vPrivateKey, ok := mFile[names.AttrPrivateKey].(string); ok && vPrivateKey != "" {
							file.PrivateKey = aws.String(vPrivateKey)
						}

						certificate.Value = file
						tls.Certificate = certificate
					}

					if vSds, ok := mCertificate["sds"].([]any); ok && len(vSds) > 0 && vSds[0] != nil {
						certificate := &awstypes.ListenerTlsCertificateMemberSds{}

						sds := awstypes.ListenerTlsSdsCertificate{}

						mSds := vSds[0].(map[string]any)

						if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
							sds.SecretName = aws.String(vSecretName)
						}

						certificate.Value = sds
						tls.Certificate = certificate
					}
				}

				if vValidation, ok := mTls["validation"].([]any); ok && len(vValidation) > 0 && vValidation[0] != nil {
					validation := &awstypes.ListenerTlsValidationContext{}

					mValidation := vValidation[0].(map[string]any)

					if vSubjectAlternativeNames, ok := mValidation["subject_alternative_names"].([]any); ok && len(vSubjectAlternativeNames) > 0 && vSubjectAlternativeNames[0] != nil {
						subjectAlternativeNames := &awstypes.SubjectAlternativeNames{}

						mSubjectAlternativeNames := vSubjectAlternativeNames[0].(map[string]any)

						if vMatch, ok := mSubjectAlternativeNames["match"].([]any); ok && len(vMatch) > 0 && vMatch[0] != nil {
							match := &awstypes.SubjectAlternativeNameMatchers{}

							mMatch := vMatch[0].(map[string]any)

							if vExact, ok := mMatch["exact"].(*schema.Set); ok && vExact.Len() > 0 {
								match.Exact = flex.ExpandStringValueSet(vExact)
							}

							subjectAlternativeNames.Match = match
						}

						validation.SubjectAlternativeNames = subjectAlternativeNames
					}

					if vTrust, ok := mValidation["trust"].([]any); ok && len(vTrust) > 0 && vTrust[0] != nil {
						mTrust := vTrust[0].(map[string]any)

						if vFile, ok := mTrust["file"].([]any); ok && len(vFile) > 0 && vFile[0] != nil {
							trust := &awstypes.ListenerTlsValidationContextTrustMemberFile{}

							file := awstypes.TlsValidationContextFileTrust{}

							mFile := vFile[0].(map[string]any)

							if vCertificateChain, ok := mFile[names.AttrCertificateChain].(string); ok && vCertificateChain != "" {
								file.CertificateChain = aws.String(vCertificateChain)
							}

							trust.Value = file
							validation.Trust = trust
						}

						if vSds, ok := mTrust["sds"].([]any); ok && len(vSds) > 0 && vSds[0] != nil {
							trust := &awstypes.ListenerTlsValidationContextTrustMemberSds{}

							sds := awstypes.TlsValidationContextSdsTrust{}

							mSds := vSds[0].(map[string]any)

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

	if vLogging, ok := mSpec["logging"].([]any); ok && len(vLogging) > 0 && vLogging[0] != nil {
		logging := &awstypes.Logging{}

		mLogging := vLogging[0].(map[string]any)

		if vAccessLog, ok := mLogging["access_log"].([]any); ok && len(vAccessLog) > 0 && vAccessLog[0] != nil {
			mAccessLog := vAccessLog[0].(map[string]any)

			if vFile, ok := mAccessLog["file"].([]any); ok && len(vFile) > 0 && vFile[0] != nil {
				accessLog := &awstypes.AccessLogMemberFile{}

				file := awstypes.FileAccessLog{}

				mFile := vFile[0].(map[string]any)

				if vFormat, ok := mFile[names.AttrFormat].([]any); ok && len(vFormat) > 0 && vFormat[0] != nil {
					mFormat := vFormat[0].(map[string]any)

					if vJsonFormatRefs, ok := mFormat[names.AttrJSON].([]any); ok && len(vJsonFormatRefs) > 0 {
						format := &awstypes.LoggingFormatMemberJson{}
						jsonFormatRefs := []awstypes.JsonFormatRef{}
						for _, vJsonFormatRef := range vJsonFormatRefs {
							mJsonFormatRef := awstypes.JsonFormatRef{
								Key:   aws.String(vJsonFormatRef.(map[string]any)[names.AttrKey].(string)),
								Value: aws.String(vJsonFormatRef.(map[string]any)[names.AttrValue].(string)),
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

				if vPath, ok := mFile[names.AttrPath].(string); ok && vPath != "" {
					file.Path = aws.String(vPath)
				}

				accessLog.Value = file
				logging.AccessLog = accessLog
			}
		}

		spec.Logging = logging
	}

	if vServiceDiscovery, ok := mSpec["service_discovery"].([]any); ok && len(vServiceDiscovery) > 0 && vServiceDiscovery[0] != nil {
		mServiceDiscovery := vServiceDiscovery[0].(map[string]any)

		if vAwsCloudMap, ok := mServiceDiscovery["aws_cloud_map"].([]any); ok && len(vAwsCloudMap) > 0 && vAwsCloudMap[0] != nil {
			serviceDiscovery := &awstypes.ServiceDiscoveryMemberAwsCloudMap{}

			awsCloudMap := awstypes.AwsCloudMapServiceDiscovery{}

			mAwsCloudMap := vAwsCloudMap[0].(map[string]any)

			if vAttributes, ok := mAwsCloudMap[names.AttrAttributes].(map[string]any); ok && len(vAttributes) > 0 {
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
			if vServiceName, ok := mAwsCloudMap[names.AttrServiceName].(string); ok && vServiceName != "" {
				awsCloudMap.ServiceName = aws.String(vServiceName)
			}

			serviceDiscovery.Value = awsCloudMap
			spec.ServiceDiscovery = serviceDiscovery
		}

		if vDns, ok := mServiceDiscovery["dns"].([]any); ok && len(vDns) > 0 && vDns[0] != nil {
			serviceDiscovery := &awstypes.ServiceDiscoveryMemberDns{}

			dns := awstypes.DnsServiceDiscovery{}

			mDns := vDns[0].(map[string]any)

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

func expandVirtualRouterSpec(vSpec []any) *awstypes.VirtualRouterSpec {
	spec := &awstypes.VirtualRouterSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]any)

	if vListeners, ok := mSpec["listener"].([]any); ok && len(vListeners) > 0 && vListeners[0] != nil {
		listeners := []awstypes.VirtualRouterListener{}

		for _, vListener := range vListeners {
			listener := awstypes.VirtualRouterListener{}

			mListener := vListener.(map[string]any)

			if vPortMapping, ok := mListener["port_mapping"].([]any); ok && len(vPortMapping) > 0 && vPortMapping[0] != nil {
				mPortMapping := vPortMapping[0].(map[string]any)

				listener.PortMapping = &awstypes.PortMapping{}

				if vPort, ok := mPortMapping[names.AttrPort].(int); ok && vPort > 0 {
					listener.PortMapping.Port = aws.Int32(int32(vPort))
				}
				if vProtocol, ok := mPortMapping[names.AttrProtocol].(string); ok && vProtocol != "" {
					listener.PortMapping.Protocol = awstypes.PortProtocol(vProtocol)
				}
			}
			listeners = append(listeners, listener)
		}
		spec.Listeners = listeners
	}

	return spec
}

func expandVirtualServiceSpec(vSpec []any) *awstypes.VirtualServiceSpec {
	spec := &awstypes.VirtualServiceSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]any)

	if vProvider, ok := mSpec["provider"].([]any); ok && len(vProvider) > 0 && vProvider[0] != nil {
		mProvider := vProvider[0].(map[string]any)

		if vVirtualNode, ok := mProvider["virtual_node"].([]any); ok && len(vVirtualNode) > 0 && vVirtualNode[0] != nil {
			provider := &awstypes.VirtualServiceProviderMemberVirtualNode{}
			mVirtualNode := vVirtualNode[0].(map[string]any)

			if vVirtualNodeName, ok := mVirtualNode["virtual_node_name"].(string); ok && vVirtualNodeName != "" {
				provider.Value = awstypes.VirtualNodeServiceProvider{
					VirtualNodeName: aws.String(vVirtualNodeName),
				}
			}

			spec.Provider = provider
		}

		if vVirtualRouter, ok := mProvider["virtual_router"].([]any); ok && len(vVirtualRouter) > 0 && vVirtualRouter[0] != nil {
			provider := &awstypes.VirtualServiceProviderMemberVirtualRouter{}
			mVirtualRouter := vVirtualRouter[0].(map[string]any)

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

func flattenClientPolicy(clientPolicy *awstypes.ClientPolicy) []any {
	if clientPolicy == nil {
		return []any{}
	}

	mClientPolicy := map[string]any{}

	if tls := clientPolicy.Tls; tls != nil {
		mTls := map[string]any{
			"enforce": aws.ToBool(tls.Enforce),
			"ports":   flex.FlattenInt32ValueSet(tls.Ports),
		}

		if certificate := tls.Certificate; certificate != nil {
			mCertificate := map[string]any{}

			switch v := certificate.(type) {
			case *awstypes.ClientTlsCertificateMemberFile:
				mFile := map[string]any{
					names.AttrCertificateChain: aws.ToString(v.Value.CertificateChain),
					names.AttrPrivateKey:       aws.ToString(v.Value.PrivateKey),
				}

				mCertificate["file"] = []any{mFile}
			case *awstypes.ClientTlsCertificateMemberSds:
				mSds := map[string]any{
					"secret_name": aws.ToString(v.Value.SecretName),
				}

				mCertificate["sds"] = []any{mSds}
			}

			mTls[names.AttrCertificate] = []any{mCertificate}
		}

		if validation := tls.Validation; validation != nil {
			mValidation := map[string]any{}

			if subjectAlternativeNames := validation.SubjectAlternativeNames; subjectAlternativeNames != nil {
				mSubjectAlternativeNames := map[string]any{}

				if match := subjectAlternativeNames.Match; match != nil {
					mMatch := map[string]any{
						"exact": match.Exact,
					}

					mSubjectAlternativeNames["match"] = []any{mMatch}
				}

				mValidation["subject_alternative_names"] = []any{mSubjectAlternativeNames}
			}

			if trust := validation.Trust; trust != nil {
				mTrust := map[string]any{}

				switch v := trust.(type) {
				case *awstypes.TlsValidationContextTrustMemberAcm:
					mAcm := map[string]any{
						"certificate_authority_arns": v.Value.CertificateAuthorityArns,
					}

					mTrust["acm"] = []any{mAcm}
				case *awstypes.TlsValidationContextTrustMemberFile:
					mFile := map[string]any{
						names.AttrCertificateChain: aws.ToString(v.Value.CertificateChain),
					}

					mTrust["file"] = []any{mFile}
				case *awstypes.TlsValidationContextTrustMemberSds:
					mSds := map[string]any{
						"secret_name": aws.ToString(v.Value.SecretName),
					}

					mTrust["sds"] = []any{mSds}
				}

				mValidation["trust"] = []any{mTrust}
			}

			mTls["validation"] = []any{mValidation}
		}

		mClientPolicy["tls"] = []any{mTls}
	}

	return []any{mClientPolicy}
}

func flattenDuration(duration *awstypes.Duration) []any {
	if duration == nil {
		return []any{}
	}

	mDuration := map[string]any{
		names.AttrUnit:  duration.Unit,
		names.AttrValue: aws.ToInt64(duration.Value),
	}

	return []any{mDuration}
}

func flattenGRPCRoute(grpcRoute *awstypes.GrpcRoute) []any {
	if grpcRoute == nil {
		return []any{}
	}

	mGrpcRoute := map[string]any{}

	if action := grpcRoute.Action; action != nil {
		if weightedTargets := action.WeightedTargets; weightedTargets != nil {
			vWeightedTargets := []any{}

			for _, weightedTarget := range weightedTargets {
				mWeightedTarget := map[string]any{
					"virtual_node":   aws.ToString(weightedTarget.VirtualNode),
					names.AttrWeight: weightedTarget.Weight,
				}

				if v := aws.ToInt32(weightedTarget.Port); v != 0 {
					mWeightedTarget[names.AttrPort] = v
				}

				vWeightedTargets = append(vWeightedTargets, mWeightedTarget)
			}

			mGrpcRoute[names.AttrAction] = []any{
				map[string]any{
					"weighted_target": vWeightedTargets,
				},
			}
		}
	}

	if grpcRouteMatch := grpcRoute.Match; grpcRouteMatch != nil {
		vGrpcRouteMetadatas := []any{}

		for _, grpcRouteMetadata := range grpcRouteMatch.Metadata {
			mGrpcRouteMetadata := map[string]any{
				"invert":       aws.ToBool(grpcRouteMetadata.Invert),
				names.AttrName: aws.ToString(grpcRouteMetadata.Name),
			}

			if match := grpcRouteMetadata.Match; match != nil {
				mMatch := map[string]any{}

				switch v := match.(type) {
				case *awstypes.GrpcRouteMetadataMatchMethodMemberExact:
					mMatch["exact"] = v.Value
				case *awstypes.GrpcRouteMetadataMatchMethodMemberPrefix:
					mMatch[names.AttrPrefix] = v.Value
				case *awstypes.GrpcRouteMetadataMatchMethodMemberRegex:
					mMatch["regex"] = v.Value
				case *awstypes.GrpcRouteMetadataMatchMethodMemberSuffix:
					mMatch["suffix"] = v.Value
				case *awstypes.GrpcRouteMetadataMatchMethodMemberRange:
					mRange := map[string]any{
						"end":   aws.ToInt64(v.Value.End),
						"start": aws.ToInt64(v.Value.Start),
					}

					mMatch["range"] = []any{mRange}
				}

				mGrpcRouteMetadata["match"] = []any{mMatch}
			}

			vGrpcRouteMetadatas = append(vGrpcRouteMetadatas, mGrpcRouteMetadata)
		}

		mGrpcRoute["match"] = []any{
			map[string]any{
				"metadata":            vGrpcRouteMetadatas,
				"method_name":         aws.ToString(grpcRouteMatch.MethodName),
				names.AttrPort:        aws.ToInt32(grpcRouteMatch.Port),
				names.AttrServiceName: aws.ToString(grpcRouteMatch.ServiceName),
			},
		}
	}

	if grpcRetryPolicy := grpcRoute.RetryPolicy; grpcRetryPolicy != nil {
		mGrpcRetryPolicy := map[string]any{
			"grpc_retry_events": grpcRetryPolicy.GrpcRetryEvents,
			"http_retry_events": grpcRetryPolicy.HttpRetryEvents,
			"max_retries":       aws.ToInt64(grpcRetryPolicy.MaxRetries),
			"per_retry_timeout": flattenDuration(grpcRetryPolicy.PerRetryTimeout),
			"tcp_retry_events":  grpcRetryPolicy.TcpRetryEvents,
		}

		mGrpcRoute["retry_policy"] = []any{mGrpcRetryPolicy}
	}

	mGrpcRoute[names.AttrTimeout] = flattenGRPCTimeout(grpcRoute.Timeout)

	return []any{mGrpcRoute}
}

func flattenGRPCTimeout(grpcTimeout *awstypes.GrpcTimeout) []any {
	if grpcTimeout == nil {
		return []any{}
	}

	mGrpcTimeout := map[string]any{
		"idle":        flattenDuration(grpcTimeout.Idle),
		"per_request": flattenDuration(grpcTimeout.PerRequest),
	}

	return []any{mGrpcTimeout}
}

func flattenHTTPRoute(httpRoute *awstypes.HttpRoute) []any {
	if httpRoute == nil {
		return []any{}
	}

	mHttpRoute := map[string]any{}

	if action := httpRoute.Action; action != nil {
		if weightedTargets := action.WeightedTargets; weightedTargets != nil {
			vWeightedTargets := []any{}

			for _, weightedTarget := range weightedTargets {
				mWeightedTarget := map[string]any{
					"virtual_node":   aws.ToString(weightedTarget.VirtualNode),
					names.AttrWeight: weightedTarget.Weight,
				}

				if v := aws.ToInt32(weightedTarget.Port); v != 0 {
					mWeightedTarget[names.AttrPort] = v
				}

				vWeightedTargets = append(vWeightedTargets, mWeightedTarget)
			}

			mHttpRoute[names.AttrAction] = []any{
				map[string]any{
					"weighted_target": vWeightedTargets,
				},
			}
		}
	}

	if httpRouteMatch := httpRoute.Match; httpRouteMatch != nil {
		vHttpRouteHeaders := []any{}

		for _, httpRouteHeader := range httpRouteMatch.Headers {
			mHttpRouteHeader := map[string]any{
				"invert":       aws.ToBool(httpRouteHeader.Invert),
				names.AttrName: aws.ToString(httpRouteHeader.Name),
			}

			if match := httpRouteHeader.Match; match != nil {
				mMatch := map[string]any{}

				switch v := match.(type) {
				case *awstypes.HeaderMatchMethodMemberExact:
					mMatch["exact"] = v.Value
				case *awstypes.HeaderMatchMethodMemberPrefix:
					mMatch[names.AttrPrefix] = v.Value
				case *awstypes.HeaderMatchMethodMemberRegex:
					mMatch["regex"] = v.Value
				case *awstypes.HeaderMatchMethodMemberSuffix:
					mMatch["suffix"] = v.Value
				case *awstypes.HeaderMatchMethodMemberRange:
					mRange := map[string]any{
						"end":   aws.ToInt64(v.Value.End),
						"start": aws.ToInt64(v.Value.Start),
					}

					mMatch["range"] = []any{mRange}
				}

				mHttpRouteHeader["match"] = []any{mMatch}
			}

			vHttpRouteHeaders = append(vHttpRouteHeaders, mHttpRouteHeader)
		}

		vHttpRoutePath := []any{}

		if httpRoutePath := httpRouteMatch.Path; httpRoutePath != nil {
			mHttpRoutePath := map[string]any{
				"exact": aws.ToString(httpRoutePath.Exact),
				"regex": aws.ToString(httpRoutePath.Regex),
			}

			vHttpRoutePath = []any{mHttpRoutePath}
		}

		vHttpRouteQueryParameters := []any{}

		for _, httpRouteQueryParameter := range httpRouteMatch.QueryParameters {
			mHttpRouteQueryParameter := map[string]any{
				names.AttrName: aws.ToString(httpRouteQueryParameter.Name),
			}

			if match := httpRouteQueryParameter.Match; match != nil {
				mMatch := map[string]any{
					"exact": aws.ToString(match.Exact),
				}

				mHttpRouteQueryParameter["match"] = []any{mMatch}
			}

			vHttpRouteQueryParameters = append(vHttpRouteQueryParameters, mHttpRouteQueryParameter)
		}

		mHttpRoute["match"] = []any{
			map[string]any{
				names.AttrHeader:  vHttpRouteHeaders,
				"method":          httpRouteMatch.Method,
				names.AttrPath:    vHttpRoutePath,
				names.AttrPort:    aws.ToInt32(httpRouteMatch.Port),
				names.AttrPrefix:  aws.ToString(httpRouteMatch.Prefix),
				"query_parameter": vHttpRouteQueryParameters,
				"scheme":          httpRouteMatch.Scheme,
			},
		}
	}

	if httpRetryPolicy := httpRoute.RetryPolicy; httpRetryPolicy != nil {
		mHttpRetryPolicy := map[string]any{
			"http_retry_events": httpRetryPolicy.HttpRetryEvents,
			"max_retries":       aws.ToInt64(httpRetryPolicy.MaxRetries),
			"per_retry_timeout": flattenDuration(httpRetryPolicy.PerRetryTimeout),
			"tcp_retry_events":  httpRetryPolicy.TcpRetryEvents,
		}

		mHttpRoute["retry_policy"] = []any{mHttpRetryPolicy}
	}

	mHttpRoute[names.AttrTimeout] = flattenHTTPTimeout(httpRoute.Timeout)

	return []any{mHttpRoute}
}

func flattenHTTPTimeout(httpTimeout *awstypes.HttpTimeout) []any {
	if httpTimeout == nil {
		return []any{}
	}

	mHttpTimeout := map[string]any{
		"idle":        flattenDuration(httpTimeout.Idle),
		"per_request": flattenDuration(httpTimeout.PerRequest),
	}

	return []any{mHttpTimeout}
}

func flattenMeshSpec(spec *awstypes.MeshSpec) []any {
	if spec == nil {
		return []any{}
	}

	mSpec := map[string]any{}

	if spec.EgressFilter != nil {
		mSpec["egress_filter"] = []any{
			map[string]any{
				names.AttrType: spec.EgressFilter.Type,
			},
		}
	}

	if spec.ServiceDiscovery != nil {
		mSpec["service_discovery"] = []any{
			map[string]any{
				"ip_preference": spec.ServiceDiscovery.IpPreference,
			},
		}
	}

	return []any{mSpec}
}

func flattenRouteSpec(spec *awstypes.RouteSpec) []any {
	if spec == nil {
		return []any{}
	}

	mSpec := map[string]any{
		"grpc_route":       flattenGRPCRoute(spec.GrpcRoute),
		"http2_route":      flattenHTTPRoute(spec.Http2Route),
		"http_route":       flattenHTTPRoute(spec.HttpRoute),
		names.AttrPriority: aws.ToInt32(spec.Priority),
		"tcp_route":        flattenTCPRoute(spec.TcpRoute),
	}

	return []any{mSpec}
}

func flattenTCPRoute(tcpRoute *awstypes.TcpRoute) []any {
	if tcpRoute == nil {
		return []any{}
	}

	mTcpRoute := map[string]any{}

	if action := tcpRoute.Action; action != nil {
		if weightedTargets := action.WeightedTargets; weightedTargets != nil {
			vWeightedTargets := []any{}

			for _, weightedTarget := range weightedTargets {
				mWeightedTarget := map[string]any{
					"virtual_node":   aws.ToString(weightedTarget.VirtualNode),
					names.AttrWeight: weightedTarget.Weight,
				}

				if v := aws.ToInt32(weightedTarget.Port); v != 0 {
					mWeightedTarget[names.AttrPort] = v
				}

				vWeightedTargets = append(vWeightedTargets, mWeightedTarget)
			}

			mTcpRoute[names.AttrAction] = []any{
				map[string]any{
					"weighted_target": vWeightedTargets,
				},
			}
		}
	}

	if tcpRouteMatch := tcpRoute.Match; tcpRouteMatch != nil {
		mTcpRoute["match"] = []any{
			map[string]any{
				names.AttrPort: aws.ToInt32(tcpRouteMatch.Port),
			},
		}
	}

	mTcpRoute[names.AttrTimeout] = flattenTCPTimeout(tcpRoute.Timeout)

	return []any{mTcpRoute}
}

func flattenTCPTimeout(tcpTimeout *awstypes.TcpTimeout) []any {
	if tcpTimeout == nil {
		return []any{}
	}

	mTcpTimeout := map[string]any{
		"idle": flattenDuration(tcpTimeout.Idle),
	}

	return []any{mTcpTimeout}
}

func flattenVirtualNodeSpec(spec *awstypes.VirtualNodeSpec) []any {
	if spec == nil {
		return []any{}
	}

	mSpec := map[string]any{}

	if backends := spec.Backends; backends != nil {
		vBackends := []any{}

		for _, backend := range backends {
			mBackend := map[string]any{}

			switch v := backend.(type) {
			case *awstypes.BackendMemberVirtualService:
				mVirtualService := map[string]any{
					"client_policy":        flattenClientPolicy(v.Value.ClientPolicy),
					"virtual_service_name": aws.ToString(v.Value.VirtualServiceName),
				}

				mBackend["virtual_service"] = []any{mVirtualService}
			}

			vBackends = append(vBackends, mBackend)
		}

		mSpec["backend"] = vBackends
	}

	if backendDefaults := spec.BackendDefaults; backendDefaults != nil {
		mBackendDefaults := map[string]any{
			"client_policy": flattenClientPolicy(backendDefaults.ClientPolicy),
		}

		mSpec["backend_defaults"] = []any{mBackendDefaults}
	}

	if len(spec.Listeners) > 0 {
		var mListeners []any
		// Per schema definition, set at most 1 Listener
		for _, listener := range spec.Listeners {
			mListener := map[string]any{}

			if connectionPool := listener.ConnectionPool; connectionPool != nil {
				mConnectionPool := map[string]any{}

				switch v := connectionPool.(type) {
				case *awstypes.VirtualNodeConnectionPoolMemberGrpc:
					mGrpcConnectionPool := map[string]any{
						"max_requests": aws.ToInt32(v.Value.MaxRequests),
					}
					mConnectionPool["grpc"] = []any{mGrpcConnectionPool}
				case *awstypes.VirtualNodeConnectionPoolMemberHttp:
					mHttpConnectionPool := map[string]any{
						"max_connections":      aws.ToInt32(v.Value.MaxConnections),
						"max_pending_requests": aws.ToInt32(v.Value.MaxPendingRequests),
					}
					mConnectionPool["http"] = []any{mHttpConnectionPool}
				case *awstypes.VirtualNodeConnectionPoolMemberHttp2:
					mHttp2ConnectionPool := map[string]any{
						"max_requests": aws.ToInt32(v.Value.MaxRequests),
					}
					mConnectionPool["http2"] = []any{mHttp2ConnectionPool}
				case *awstypes.VirtualNodeConnectionPoolMemberTcp:
					mTcpConnectionPool := map[string]any{
						"max_connections": aws.ToInt32(v.Value.MaxConnections),
					}
					mConnectionPool["tcp"] = []any{mTcpConnectionPool}
				}

				mListener["connection_pool"] = []any{mConnectionPool}
			}

			if healthCheck := listener.HealthCheck; healthCheck != nil {
				mHealthCheck := map[string]any{
					"healthy_threshold":   aws.ToInt32(healthCheck.HealthyThreshold),
					"interval_millis":     aws.ToInt64(healthCheck.IntervalMillis),
					names.AttrPath:        aws.ToString(healthCheck.Path),
					names.AttrPort:        aws.ToInt32(healthCheck.Port),
					names.AttrProtocol:    healthCheck.Protocol,
					"timeout_millis":      aws.ToInt64(healthCheck.TimeoutMillis),
					"unhealthy_threshold": aws.ToInt32(healthCheck.UnhealthyThreshold),
				}
				mListener[names.AttrHealthCheck] = []any{mHealthCheck}
			}

			if outlierDetection := listener.OutlierDetection; outlierDetection != nil {
				mOutlierDetection := map[string]any{
					"base_ejection_duration": flattenDuration(outlierDetection.BaseEjectionDuration),
					names.AttrInterval:       flattenDuration(outlierDetection.Interval),
					"max_ejection_percent":   aws.ToInt32(outlierDetection.MaxEjectionPercent),
					"max_server_errors":      aws.ToInt64(outlierDetection.MaxServerErrors),
				}
				mListener["outlier_detection"] = []any{mOutlierDetection}
			}

			if portMapping := listener.PortMapping; portMapping != nil {
				mPortMapping := map[string]any{
					names.AttrPort:     aws.ToInt32(portMapping.Port),
					names.AttrProtocol: portMapping.Protocol,
				}
				mListener["port_mapping"] = []any{mPortMapping}
			}

			if listenerTimeout := listener.Timeout; listenerTimeout != nil {
				mListenerTimeout := map[string]any{}

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

				mListener[names.AttrTimeout] = []any{mListenerTimeout}
			}

			if tls := listener.Tls; tls != nil {
				mTls := map[string]any{
					names.AttrMode: tls.Mode,
				}

				if certificate := tls.Certificate; certificate != nil {
					mCertificate := map[string]any{}

					switch v := certificate.(type) {
					case *awstypes.ListenerTlsCertificateMemberAcm:
						mAcm := map[string]any{
							names.AttrCertificateARN: aws.ToString(v.Value.CertificateArn),
						}

						mCertificate["acm"] = []any{mAcm}
					case *awstypes.ListenerTlsCertificateMemberFile:
						mFile := map[string]any{
							names.AttrCertificateChain: aws.ToString(v.Value.CertificateChain),
							names.AttrPrivateKey:       aws.ToString(v.Value.PrivateKey),
						}

						mCertificate["file"] = []any{mFile}
					case *awstypes.ListenerTlsCertificateMemberSds:
						mSds := map[string]any{
							"secret_name": aws.ToString(v.Value.SecretName),
						}

						mCertificate["sds"] = []any{mSds}
					}

					mTls[names.AttrCertificate] = []any{mCertificate}
				}

				if validation := tls.Validation; validation != nil {
					mValidation := map[string]any{}

					if subjectAlternativeNames := validation.SubjectAlternativeNames; subjectAlternativeNames != nil {
						mSubjectAlternativeNames := map[string]any{}

						if match := subjectAlternativeNames.Match; match != nil {
							mMatch := map[string]any{
								"exact": match.Exact,
							}

							mSubjectAlternativeNames["match"] = []any{mMatch}
						}

						mValidation["subject_alternative_names"] = []any{mSubjectAlternativeNames}
					}

					if trust := validation.Trust; trust != nil {
						mTrust := map[string]any{}

						switch v := trust.(type) {
						case *awstypes.ListenerTlsValidationContextTrustMemberFile:
							mFile := map[string]any{
								names.AttrCertificateChain: aws.ToString(v.Value.CertificateChain),
							}

							mTrust["file"] = []any{mFile}
						case *awstypes.ListenerTlsValidationContextTrustMemberSds:
							mSds := map[string]any{
								"secret_name": aws.ToString(v.Value.SecretName),
							}

							mTrust["sds"] = []any{mSds}
						}

						mValidation["trust"] = []any{mTrust}
					}

					mTls["validation"] = []any{mValidation}
				}

				mListener["tls"] = []any{mTls}
			}
			mListeners = append(mListeners, mListener)
		}
		mSpec["listener"] = mListeners
	}

	if logging := spec.Logging; logging != nil {
		mLogging := map[string]any{}

		if accessLog := logging.AccessLog; accessLog != nil {
			mAccessLog := map[string]any{}

			switch v := accessLog.(type) {
			case *awstypes.AccessLogMemberFile:
				mFile := map[string]any{}

				if format := v.Value.Format; format != nil {
					mFormat := map[string]any{}

					switch v := format.(type) {
					case *awstypes.LoggingFormatMemberJson:
						vJsons := []any{}

						for _, j := range v.Value {
							mJson := map[string]any{
								names.AttrKey:   aws.ToString(j.Key),
								names.AttrValue: aws.ToString(j.Value),
							}

							vJsons = append(vJsons, mJson)
						}

						mFormat[names.AttrJSON] = vJsons
					case *awstypes.LoggingFormatMemberText:
						mFormat["text"] = v.Value
					}

					mFile[names.AttrFormat] = []any{mFormat}
				}

				mFile[names.AttrPath] = aws.ToString(v.Value.Path)

				mAccessLog["file"] = []any{mFile}
			}

			mLogging["access_log"] = []any{mAccessLog}
		}

		mSpec["logging"] = []any{mLogging}
	}

	if serviceDiscovery := spec.ServiceDiscovery; serviceDiscovery != nil {
		mServiceDiscovery := map[string]any{}

		switch v := serviceDiscovery.(type) {
		case *awstypes.ServiceDiscoveryMemberAwsCloudMap:
			vAttributes := map[string]any{}

			for _, attribute := range v.Value.Attributes {
				vAttributes[aws.ToString(attribute.Key)] = aws.ToString(attribute.Value)
			}

			mServiceDiscovery["aws_cloud_map"] = []any{
				map[string]any{
					names.AttrAttributes:  vAttributes,
					"namespace_name":      aws.ToString(v.Value.NamespaceName),
					names.AttrServiceName: aws.ToString(v.Value.ServiceName),
				},
			}
		case *awstypes.ServiceDiscoveryMemberDns:
			mServiceDiscovery["dns"] = []any{
				map[string]any{
					"hostname":      aws.ToString(v.Value.Hostname),
					"ip_preference": v.Value.IpPreference,
					"response_type": v.Value.ResponseType,
				},
			}
		}

		mSpec["service_discovery"] = []any{mServiceDiscovery}
	}

	return []any{mSpec}
}

func flattenVirtualRouterSpec(spec *awstypes.VirtualRouterSpec) []any {
	if spec == nil {
		return []any{}
	}
	mSpec := make(map[string]any)
	if len(spec.Listeners) > 0 {
		var mListeners []any
		for _, listener := range spec.Listeners {
			mListener := map[string]any{}
			if listener.PortMapping != nil {
				mPortMapping := map[string]any{
					names.AttrPort:     aws.ToInt32(listener.PortMapping.Port),
					names.AttrProtocol: listener.PortMapping.Protocol,
				}
				mListener["port_mapping"] = []any{mPortMapping}
			}
			mListeners = append(mListeners, mListener)
		}
		mSpec["listener"] = mListeners
	}

	return []any{mSpec}
}

func flattenVirtualServiceSpec(spec *awstypes.VirtualServiceSpec) []any {
	if spec == nil {
		return []any{}
	}

	mSpec := map[string]any{}

	if provider := spec.Provider; provider != nil {
		mProvider := map[string]any{}

		switch v := provider.(type) {
		case *awstypes.VirtualServiceProviderMemberVirtualNode:
			mProvider["virtual_node"] = []any{
				map[string]any{
					"virtual_node_name": aws.ToString(v.Value.VirtualNodeName),
				},
			}
		case *awstypes.VirtualServiceProviderMemberVirtualRouter:
			mProvider["virtual_router"] = []any{
				map[string]any{
					"virtual_router_name": aws.ToString(v.Value.VirtualRouterName),
				},
			}
		}

		mSpec["provider"] = []any{mProvider}
	}

	return []any{mSpec}
}
