// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const (
	envVarVCRMode = "VCR_MODE"
	envVarVCRPath = "VCR_PATH"
)

type randomnessSource struct {
	seed   int64
	source rand.Source
}

type metaMap map[string]*conns.AWSClient

func (m metaMap) Lock() {
	conns.GlobalMutexKV.Lock(m.key())
}

func (m metaMap) Unlock() {
	conns.GlobalMutexKV.Unlock(m.key())
}

func (m metaMap) key() string {
	return "vcr-metas"
}

type randomnessSourceMap map[string]*randomnessSource

func (m randomnessSourceMap) Lock() {
	conns.GlobalMutexKV.Lock(m.key())
}

func (m randomnessSourceMap) Unlock() {
	conns.GlobalMutexKV.Unlock(m.key())
}

func (m randomnessSourceMap) key() string {
	return "vcr-randomness-sources"
}

var (
	providerMetas     = metaMap(make(map[string]*conns.AWSClient, 0))
	randomnessSources = randomnessSourceMap(make(map[string]*randomnessSource, 0))
)

// ProviderMeta returns the current provider's state (AKA "meta" or "conns.AWSClient").
func ProviderMeta(t *testing.T) *conns.AWSClient {
	providerMetas.Lock()
	meta, ok := providerMetas[t.Name()]
	defer providerMetas.Unlock()

	if !ok {
		meta = Provider.Meta().(*conns.AWSClient)
	}

	return meta
}

func isVCREnabled() bool {
	return os.Getenv(envVarVCRMode) != "" && os.Getenv(envVarVCRPath) != ""
}

func vcrMode() (recorder.Mode, error) {
	switch v := os.Getenv(envVarVCRMode); v {
	case "RECORDING":
		return recorder.ModeRecordOnce, nil
	case "REPLAYING":
		return recorder.ModeReplayOnly, nil
	default:
		return recorder.ModePassthrough, fmt.Errorf("unsupported value for %s: %s", envVarVCRMode, v)
	}
}

// vcrEnabledProtoV5ProviderFactories returns ProtoV5ProviderFactories ready for use with VCR.
func vcrEnabledProtoV5ProviderFactories(t *testing.T, input map[string]func() (tfprotov5.ProviderServer, error)) map[string]func() (tfprotov5.ProviderServer, error) {
	output := make(map[string]func() (tfprotov5.ProviderServer, error), len(input))

	for name := range input {
		output[name] = func() (tfprotov5.ProviderServer, error) {
			providerServerFactory, primary, err := provider.ProtoV5ProviderServerFactory(context.Background())

			if err != nil {
				return nil, err
			}

			primary.ConfigureContextFunc = vcrProviderConfigureContextFunc(primary, primary.ConfigureContextFunc, t.Name())

			return providerServerFactory(), nil
		}
	}

	return output
}

// vcrProviderConfigureContextFunc returns a provider configuration function returning cached provider instance state.
// This is necessary as ConfigureContextFunc is called multiple times for a given test, each time creating a new HTTP client.
// VCR requires a single HTTP client to handle all interactions.
func vcrProviderConfigureContextFunc(provider *schema.Provider, configureContextFunc schema.ConfigureContextFunc, testName string) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var diags diag.Diagnostics

		providerMetas.Lock()
		meta, ok := providerMetas[testName]
		defer providerMetas.Unlock()

		if ok {
			return meta, nil
		}

		vcrMode, err := vcrMode()

		if err != nil {
			return nil, sdkdiag.AppendFromErr(diags, err)
		}

		// Cribbed from aws-sdk-go-base.
		httpClient := cleanhttp.DefaultPooledClient()
		transport := httpClient.Transport.(*http.Transport)
		transport.MaxIdleConnsPerHost = 10
		tlsConfig := transport.TLSClientConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{}
			transport.TLSClientConfig = tlsConfig
		}
		tlsConfig.MinVersion = tls.VersionTLS12

		path := filepath.Join(os.Getenv(envVarVCRPath), vcrFileName(testName))

		// Create a VCR recorder around a default HTTP client.
		r, err := recorder.NewWithOptions(&recorder.Options{
			CassetteName:  path,
			Mode:          vcrMode,
			RealTransport: httpClient.Transport,
		})

		if err != nil {
			return nil, sdkdiag.AppendFromErr(diags, err)
		}

		// Remove sensitive HTTP headers.
		r.AddHook(func(i *cassette.Interaction) error {
			delete(i.Request.Headers, "Authorization")
			delete(i.Request.Headers, "X-Amz-Security-Token")

			return nil
		}, recorder.AfterCaptureHook)

		// Defines how VCR will match requests to responses.
		r.SetMatcher(func(r *http.Request, i cassette.Request) bool {
			// Default matcher compares method and URL only.
			if !cassette.DefaultMatcher(r, i) {
				return false
			}

			if r.Body == nil {
				return true
			}

			var b bytes.Buffer
			if _, err := b.ReadFrom(r.Body); err != nil {
				tflog.Debug(ctx, "Failed to read request body from cassette", map[string]interface{}{
					"error": err,
				})
				return false
			}

			r.Body = io.NopCloser(&b)
			body := b.String()
			// If body matches identically, we are done.
			if body == i.Body {
				return true
			}

			// https://awslabs.github.io/smithy/1.0/spec/aws/index.html#aws-protocols.
			switch contentType := r.Header.Get("Content-Type"); contentType {
			case "application/json", "application/x-amz-json-1.0", "application/x-amz-json-1.1":
				// JSON might be the same, but reordered. Try parsing and comparing.
				var requestJson, cassetteJson interface{}

				if err := json.Unmarshal([]byte(body), &requestJson); err != nil {
					tflog.Debug(ctx, "Failed to unmarshal request JSON", map[string]interface{}{
						"error": err,
					})
					return false
				}

				if err := json.Unmarshal([]byte(i.Body), &cassetteJson); err != nil {
					tflog.Debug(ctx, "Failed to unmarshal cassette JSON", map[string]interface{}{
						"error": err,
					})
					return false
				}

				return reflect.DeepEqual(requestJson, cassetteJson)

			case "application/xml":
				// XML might be the same, but reordered. Try parsing and comparing.
				var requestXml, cassetteXml interface{}

				if err := xml.Unmarshal([]byte(body), &requestXml); err != nil {
					tflog.Debug(ctx, "Failed to unmarshal request XML", map[string]interface{}{
						"error": err,
					})
					return false
				}

				if err := xml.Unmarshal([]byte(i.Body), &cassetteXml); err != nil {
					tflog.Debug(ctx, "Failed to unmarshal cassette XML", map[string]interface{}{
						"error": err,
					})
					return false
				}

				return reflect.DeepEqual(requestXml, cassetteXml)
			}

			return false
		})

		// Use the wrapped HTTP Client for AWS APIs.
		// As the HTTP client is used in the provider's ConfigureContextFunc
		// we must do this setup before calling the ConfigureContextFunc.
		httpClient.Transport = r
		if v, ok := provider.Meta().(*conns.AWSClient); ok {
			meta = v
		} else {
			meta = new(conns.AWSClient)
		}
		meta.SetHTTPClient(httpClient)
		provider.SetMeta(meta)

		if v, ds := configureContextFunc(ctx, d); ds.HasError() {
			return nil, append(diags, ds...)
		} else {
			meta = v.(*conns.AWSClient)
		}

		// Don't retry requests if a recorded interaction isn't found.
		// TODO Need to loop through all API clients to do this.
		// TODO Use []*client.Client?
		// TODO AWS SDK for Go v2 API clients.
		meta.LogsConn(ctx).Handlers.AfterRetry.PushFront(func(r *request.Request) {
			// We have to use 'Contains' rather than 'errors.Is' because 'awserr.Error' doesn't implement 'Unwrap'.
			if errs.Contains(r.Error, cassette.ErrInteractionNotFound.Error()) {
				r.Retryable = aws.Bool(false)
			}
		})

		providerMetas[testName] = meta

		return meta, diags
	}
}

// vcrRandomnessSource returns a rand.Source for VCR testing.
// In RECORDING mode, generates a new seed and saves it to a file, using the seed for the source.
// In REPLAYING mode, reads a seed from a file and creates a source from it.
func vcrRandomnessSource(t *testing.T) (*randomnessSource, error) {
	testName := t.Name()

	randomnessSources.Lock()
	s, ok := randomnessSources[testName]
	defer randomnessSources.Unlock()

	if ok {
		return s, nil
	}

	vcrMode, err := vcrMode()

	if err != nil {
		return nil, err
	}

	switch vcrMode {
	case recorder.ModeRecordOnce:
		seed := rand.Int63()
		s = &randomnessSource{
			seed:   seed,
			source: rand.NewSource(seed),
		}
	case recorder.ModeReplayOnly:
		seed, err := readSeedFromFile(vcrSeedFile(os.Getenv(envVarVCRPath), testName))

		if err != nil {
			return nil, fmt.Errorf("no cassette found on disk for %s, please replay this testcase in recording mode - %w", testName, err)
		}

		s = &randomnessSource{
			seed:   seed,
			source: rand.NewSource(seed),
		}
	default:
		t.FailNow()
	}

	randomnessSources[testName] = s

	return s, nil
}

func vcrFileName(name string) string {
	return strings.ReplaceAll(name, "/", "_")
}

func vcrSeedFile(path, name string) string {
	return filepath.Join(path, fmt.Sprintf("%s.seed", vcrFileName(name)))
}

func readSeedFromFile(fileName string) (int64, error) {
	// Max number of digits for int64 is 19.
	data := make([]byte, 19)
	f, err := os.Open(fileName)

	if err != nil {
		return 0, err
	}

	defer f.Close()

	_, err = f.Read(data)

	if err != nil {
		return 0, err
	}

	// Remove NULL characters from seed.
	return strconv.ParseInt(string(bytes.Trim(data, "\x00")), 10, 64)
}

func writeSeedToFile(seed int64, fileName string) error {
	f, err := os.Create(fileName)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(strconv.FormatInt(seed, 10))

	return err
}

// closeVCRRecorder closes the VCR recorder, saving the cassette and randomness seed.
func closeVCRRecorder(t *testing.T) {
	// Don't close the recorder if we're running because of a panic.
	if p := recover(); p != nil {
		panic(p)
	}

	testName := t.Name()
	providerMetas.Lock()
	meta, ok := providerMetas[testName]
	defer providerMetas.Unlock()

	if ok {
		if !t.Failed() {
			if v, ok := meta.HTTPClient().Transport.(*recorder.Recorder); ok {
				t.Log("stopping VCR recorder")
				if err := v.Stop(); err != nil {
					t.Error(err)
				}
			}
		}

		delete(providerMetas, testName)
	}

	// Save the randomness seed.
	randomnessSources.Lock()
	s, ok := randomnessSources[testName]
	defer randomnessSources.Unlock()

	if ok {
		if !t.Failed() {
			t.Log("persisting randomness seed")
			if err := writeSeedToFile(s.seed, vcrSeedFile(os.Getenv(envVarVCRPath), t.Name())); err != nil {
				t.Error(err)
			}
		}

		delete(randomnessSources, testName)
	}
}

// ParallelTest wraps resource.ParallelTest, initializing VCR if enabled.
func ParallelTest(t *testing.T, c resource.TestCase) {
	if isVCREnabled() {
		c.ProtoV5ProviderFactories = vcrEnabledProtoV5ProviderFactories(t, c.ProtoV5ProviderFactories)
		defer closeVCRRecorder(t)
	}

	resource.ParallelTest(t, c)
}

// Test wraps resource.Test, initializing VCR if enabled.
func Test(t *testing.T, c resource.TestCase) {
	if isVCREnabled() {
		c.ProtoV5ProviderFactories = vcrEnabledProtoV5ProviderFactories(t, c.ProtoV5ProviderFactories)
		defer closeVCRRecorder(t)
	}

	resource.Test(t, c)
}

// RandInt is a VCR-friendly replacement for acctest.RandInt.
func RandInt(t *testing.T) int {
	if !isVCREnabled() {
		return sdkacctest.RandInt()
	}

	s, err := vcrRandomnessSource(t)

	if err != nil {
		t.Fatal(err)
	}

	return rand.New(s.source).Int()
}

// RandomWithPrefix is a VCR-friendly replacement for acctest.RandomWithPrefix.
func RandomWithPrefix(t *testing.T, prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, RandInt(t))
}
