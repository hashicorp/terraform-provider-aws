/*
This file is a hard copy of:
https://github.com/kubernetes-sigs/aws-iam-authenticator/blob/7547c74e660f8d34d9980f2c69aa008eed1f48d0/pkg/token/token.go

With the following modifications:

 - Removal of all Generator interface methods and implementations except GetWithSTS
 - Removal of other unused code
 - Use *sts.STS instead of stsiface.STSAPI in Generator interface and GetWithSTS implementation
 - Hard copy and use local Canonicalize implementation instead of "sigs.k8s.io/aws-iam-authenticator/pkg/arn"
 - Fix staticcheck reports
 - Ignore errorlint reports
 - Refactor deprecated io/ioutil in Go 1.16
 - Adds context parameter
 - Updated to use the AWS SDK for Go v2
*/

/*
Copyright 2017-2020 by the contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package eks

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Identity is returned on successful Verify() results. It contains a parsed
// version of the AWS identity used to create the token.
type Identity struct {
	// ARN is the raw Amazon Resource Name returned by sts:GetCallerIdentity
	ARN string

	// CanonicalARN is the Amazon Resource Name converted to a more canonical
	// representation. In particular, STS assumed role ARNs like
	// "arn:aws:sts::ACCOUNTID:assumed-role/ROLENAME/SESSIONNAME" are converted
	// to their IAM ARN equivalent "arn:aws:iam::ACCOUNTID:role/NAME"
	CanonicalARN string

	// AccountID is the 12 digit AWS account number.
	AccountID string

	// UserID is the unique user/role ID (e.g., "AROAAAAAAAAAAAAAAAAAA").
	UserID string

	// SessionName is the STS session name (or "" if this is not a
	// session-based identity). For EC2 instance roles, this will be the EC2
	// instance ID (e.g., "i-0123456789abcdef0"). You should only rely on it
	// if you trust that _only_ EC2 is allowed to assume the IAM Role. If IAM
	// users or other roles are allowed to assume the role, they can provide
	// (nearly) arbitrary strings here.
	SessionName string

	// The AWS Access Key ID used to authenticate the request.  This can be used
	// in conjunction with CloudTrail to determine the identity of the individual
	// if the individual assumed an IAM role before making the request.
	AccessKeyID string
}

const (
	// The sts GetCallerIdentity request is valid for 15 minutes regardless of this parameters value after it has been
	// signed, but we set this unused parameter to 60 for legacy reasons (we check for a value between 0 and 60 on the
	// server side in 0.3.0 or earlier).  IT IS IGNORED.  If we can get STS to support x-amz-expires, then we should
	// set this parameter to the actual expiration, and make it configurable.
	requestPresignParam = 60
	// The actual token expiration (presigned STS urls are valid for 15 minutes after timestamp in x-amz-date).
	presignedURLExpiration = 15 * time.Minute
	v1Prefix               = "k8s-aws-v1."
	maxTokenLenBytes       = 1024 * 4
	clusterIDHeader        = "x-k8s-aws-id"
	// Format of the X-Amz-Date header used for expiration
	// https://golang.org/pkg/time/#pkg-constants
	dateHeaderFormat = "20060102T150405Z"
	hostRegexp       = `^sts(\.[a-z1-9\-]+)?\.amazonaws\.com(\.cn)?$`
)

// Token is generated and used by Kubernetes client-go to authenticate with a Kubernetes cluster.
type Token struct {
	Token string
}

// FormatError is returned when there is a problem with token that is
// an encoded sts request.  This can include the url, data, action or anything
// else that prevents the sts call from being made.
type FormatError struct {
	message string
}

func (e FormatError) Error() string {
	return "input token was not properly formatted: " + e.message
}

// STSError is returned when there was either an error calling STS or a problem
// processing the data returned from STS.
type STSError struct {
	message string
}

func (e STSError) Error() string {
	return "sts getCallerIdentity failed: " + e.message
}

// NewSTSError creates a error of type STS.
func NewSTSError(m string) STSError {
	return STSError{message: m}
}

var parameterWhitelist = map[string]bool{
	names.AttrAction:       true,
	names.AttrVersion:      true,
	"x-amz-algorithm":      true,
	"x-amz-credential":     true,
	"x-amz-date":           true,
	"x-amz-expires":        true,
	"x-amz-security-token": true,
	"x-amz-signature":      true,
	"x-amz-signedheaders":  true,
}

// this is the result type from the GetCallerIdentity endpoint
type getCallerIdentityWrapper struct {
	GetCallerIdentityResponse struct {
		GetCallerIdentityResult struct {
			Account string `json:"Account"`
			Arn     string `json:"Arn"`
			UserID  string `json:"UserId"`
		} `json:"GetCallerIdentityResult"`
		ResponseMetadata struct {
			RequestID string `json:"RequestId"`
		} `json:"ResponseMetadata"`
	} `json:"GetCallerIdentityResponse"`
}

// Generator provides new tokens for the AWS IAM Authenticator.
type Generator interface {
	// GetWithSTS returns a token valid for clusterID using the given STS client.
	GetWithSTS(ctx context.Context, clusterID string, stsAPI *sts.Client) (Token, error)
}

type generator struct {
	forwardSessionName bool
	cache              bool
}

// NewGenerator creates a Generator and returns it.
func NewGenerator(forwardSessionName bool, cache bool) (Generator, error) {
	return generator{
		forwardSessionName: forwardSessionName,
		cache:              cache,
	}, nil
}

// GetWithSTS returns a token valid for clusterID using the given STS client.
func (g generator) GetWithSTS(ctx context.Context, clusterID string, stsAPI *sts.Client) (Token, error) {
	// Sign the request.  The expires parameter (sets the x-amz-expires header) is
	// not supported by the STS presigner.  We set it to 60 [nano]seconds for backwards compatibility (the
	// parameter was a required argument to AWS SDK v1's Presign(), and authenticators 0.3.0 and older are
	// expecting a value between 0 and 60 on the server side).
	presigner := sts.NewPresignClient(stsAPI, func(po *sts.PresignOptions) {
		po.ClientOptions = []func(*sts.Options){
			func(o *sts.Options) {
				o.APIOptions = []func(*middleware.Stack) error{
					addClusterIdHeaderSetterMiddleware(clusterID),
					addExpiryParamSetterMiddleware(requestPresignParam),
				}
			},
		}
	})

	request, err := presigner.PresignGetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return Token{}, err
	}

	return Token{v1Prefix + base64.RawURLEncoding.EncodeToString([]byte(request.URL))}, nil
}

func addClusterIdHeaderSetterMiddleware(clusterID string) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Build.Add(
			setClusterIdHeaderMiddleware(clusterID),
			middleware.After,
		)
	}
}

func setClusterIdHeaderMiddleware(clusterID string) middleware.BuildMiddleware {
	return middleware.BuildMiddlewareFunc(
		"Test: Set ClusterId",
		func(ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler) (out middleware.BuildOutput, metadata middleware.Metadata, err error) {
			switch req := in.Request.(type) {
			case *smithyhttp.Request:
				req.Header.Add(clusterIDHeader, clusterID)
			default:
				return out, metadata, fmt.Errorf("unknown transport type %T", in.Request)
			}

			return next.HandleBuild(ctx, in)
		},
	)
}

func addExpiryParamSetterMiddleware(expire time.Duration) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Build.Add(
			setExpiryParamMiddleware(expire),
			middleware.After,
		)
	}
}

func setExpiryParamMiddleware(expire time.Duration) middleware.BuildMiddleware {
	return middleware.BuildMiddlewareFunc(
		"Test: Set ExpiryParam",
		func(ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler) (out middleware.BuildOutput, metadata middleware.Metadata, err error) {
			switch req := in.Request.(type) {
			case *smithyhttp.Request:
				query := req.URL.Query()
				query.Set("X-Amz-Expires", strconv.FormatInt(int64(expire/time.Second), 10))
				req.URL.RawQuery = query.Encode()
			default:
				return out, metadata, fmt.Errorf("unknown transport type %T", in.Request)
			}

			return next.HandleBuild(ctx, in)
		},
	)
}

// Verifier validates tokens by calling STS and returning the associated identity.
type Verifier interface {
	Verify(token string) (*Identity, error)
}

type tokenVerifier struct {
	client    *http.Client
	clusterID string
}

// NewVerifier creates a Verifier that is bound to the clusterID and uses the default http client.
func NewVerifier(clusterID string) Verifier {
	return tokenVerifier{
		client:    http.DefaultClient,
		clusterID: clusterID,
	}
}

// verify a sts host, doc: http://docs.amazonaws.cn/en_us/general/latest/gr/rande.html#sts_region
func (v tokenVerifier) verifyHost(host string) error {
	if match, _ := regexp.MatchString(hostRegexp, host); !match {
		return FormatError{fmt.Sprintf("unexpected hostname %q in pre-signed URL", host)}
	}

	return nil
}

// Verify a token is valid for the specified clusterID. On success, returns an
// Identity that contains information about the AWS principal that created the
// token. On failure, returns nil and a non-nil error.
func (v tokenVerifier) Verify(token string) (*Identity, error) {
	if len(token) > maxTokenLenBytes {
		return nil, FormatError{"token is too large"}
	}

	if !strings.HasPrefix(token, v1Prefix) {
		return nil, FormatError{fmt.Sprintf("token is missing expected %q prefix", v1Prefix)}
	}

	// TODO: this may need to be a constant-time base64 decoding
	tokenBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(token, v1Prefix))
	if err != nil {
		return nil, FormatError{err.Error()}
	}

	parsedURL, err := url.Parse(string(tokenBytes))
	if err != nil {
		return nil, FormatError{err.Error()}
	}

	if parsedURL.Scheme != "https" {
		return nil, FormatError{fmt.Sprintf("unexpected scheme %q in pre-signed URL", parsedURL.Scheme)}
	}

	if err = v.verifyHost(parsedURL.Host); err != nil {
		return nil, err
	}

	if parsedURL.Path != "/" {
		return nil, FormatError{"unexpected path in pre-signed URL"}
	}

	queryParamsLower := make(url.Values)
	queryParams := parsedURL.Query()
	for key, values := range queryParams {
		if !parameterWhitelist[strings.ToLower(key)] {
			return nil, FormatError{fmt.Sprintf("non-whitelisted query parameter %q", key)}
		}
		if len(values) != 1 {
			return nil, FormatError{"query parameter with multiple values not supported"}
		}
		queryParamsLower.Set(strings.ToLower(key), values[0])
	}

	if queryParamsLower.Get(names.AttrAction) != "GetCallerIdentity" {
		return nil, FormatError{"unexpected action parameter in pre-signed URL"}
	}

	if !hasSignedClusterIDHeader(&queryParamsLower) {
		return nil, FormatError{fmt.Sprintf("client did not sign the %s header in the pre-signed URL", clusterIDHeader)}
	}

	// We validate x-amz-expires is between 0 and 15 minutes (900 seconds) although currently pre-signed STS URLs, and
	// therefore tokens, expire exactly 15 minutes after the x-amz-date header, regardless of x-amz-expires.
	expires, err := strconv.Atoi(queryParamsLower.Get("x-amz-expires"))
	if err != nil || expires < 0 || expires > 900 {
		return nil, FormatError{fmt.Sprintf("invalid X-Amz-Expires parameter in pre-signed URL: %d", expires)}
	}

	date := queryParamsLower.Get("x-amz-date")
	if date == "" {
		return nil, FormatError{"X-Amz-Date parameter must be present in pre-signed URL"}
	}

	// Obtain AWS Access Key ID from supplied credentials
	accessKeyID := strings.Split(queryParamsLower.Get("x-amz-credential"), "/")[0]

	dateParam, err := time.Parse(dateHeaderFormat, date)
	if err != nil {
		return nil, FormatError{fmt.Sprintf("error parsing X-Amz-Date parameter %s into format %s: %s", date, dateHeaderFormat, err.Error())}
	}

	now := time.Now()
	expiration := dateParam.Add(presignedURLExpiration)
	if now.After(expiration) {
		return nil, FormatError{fmt.Sprintf("X-Amz-Date parameter is expired (%.f minute expiration) %s", presignedURLExpiration.Minutes(), dateParam)}
	}

	req, _ := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	req.Header.Set(clusterIDHeader, v.clusterID)
	req.Header.Set("accept", "application/json")

	response, err := v.client.Do(req)
	if err != nil {
		// special case to avoid printing the full URL if possible
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return nil, NewSTSError(fmt.Sprintf("error during GET: %v", urlErr.Err))
		}
		return nil, NewSTSError(fmt.Sprintf("error during GET: %v", err))
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, NewSTSError(fmt.Sprintf("error reading HTTP result: %v", err))
	}

	if response.StatusCode != http.StatusOK {
		return nil, NewSTSError(fmt.Sprintf("error from AWS (expected 200, got %d). Body: %s", response.StatusCode, string(responseBody[:])))
	}

	var callerIdentity getCallerIdentityWrapper
	err = json.Unmarshal(responseBody, &callerIdentity)
	if err != nil {
		return nil, NewSTSError(err.Error())
	}

	// parse the response into an Identity
	id := &Identity{
		ARN:         callerIdentity.GetCallerIdentityResponse.GetCallerIdentityResult.Arn,
		AccountID:   callerIdentity.GetCallerIdentityResponse.GetCallerIdentityResult.Account,
		AccessKeyID: accessKeyID,
	}
	id.CanonicalARN, err = Canonicalize(id.ARN)
	if err != nil {
		return nil, NewSTSError(err.Error())
	}

	// The user ID is either UserID:SessionName (for assumed roles) or just
	// UserID (for IAM User principals).
	userIDParts := strings.Split(callerIdentity.GetCallerIdentityResponse.GetCallerIdentityResult.UserID, ":")
	if len(userIDParts) == 2 {
		id.UserID = userIDParts[0]
		id.SessionName = userIDParts[1]
	} else if len(userIDParts) == 1 {
		id.UserID = userIDParts[0]
	} else {
		return nil, STSError{fmt.Sprintf(
			"malformed UserID %q",
			callerIdentity.GetCallerIdentityResponse.GetCallerIdentityResult.UserID)}
	}

	return id, nil
}

func hasSignedClusterIDHeader(paramsLower *url.Values) bool {
	signedHeaders := strings.Split(paramsLower.Get("x-amz-signedheaders"), ";")
	for _, hdr := range signedHeaders {
		if strings.EqualFold(hdr, clusterIDHeader) {
			return true
		}
	}
	return false
}
