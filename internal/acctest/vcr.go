// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"math/rand" // nosemgrep: go.lang.security.audit.crypto.math_random.math-random-used -- Deterministic PRNG required for VCR test reproducibility
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
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

// ProviderMeta returns the current provider's state (AKA "meta" or "conns.AWSClient")
func ProviderMeta(_ context.Context, t *testing.T) *conns.AWSClient {
	t.Helper()

	providerMetas.Lock()
	meta, ok := providerMetas[t.Name()]
	defer providerMetas.Unlock()

	if !ok {
		meta = Provider.Meta().(*conns.AWSClient)
	}

	return meta
}

// vcrEnabledProtoV5ProviderFactories returns ProtoV5ProviderFactories ready for use
// with VCR
func vcrEnabledProtoV5ProviderFactories(ctx context.Context, t *testing.T, input map[string]func() (tfprotov5.ProviderServer, error)) map[string]func() (tfprotov5.ProviderServer, error) {
	t.Helper()

	output := make(map[string]func() (tfprotov5.ProviderServer, error), len(input))

	for name := range input {
		output[name] = func() (tfprotov5.ProviderServer, error) {
			providerServerFactory, primary, err := provider.ProtoV5ProviderServerFactory(ctx)

			if err != nil {
				return nil, err
			}

			primary.ConfigureContextFunc = vcrProviderConfigureContextFunc(primary, primary.ConfigureContextFunc, t)

			return providerServerFactory(), nil
		}
	}

	return output
}

// vcrProviderConfigureContextFunc returns a provider configuration function returning
// cached provider instance state
//
// This is necessary as ConfigureContextFunc is called multiple times for a given test,
// each time creating a new HTTP client. VCR requires a single HTTP client to handle all
// interactions.
func vcrProviderConfigureContextFunc(provider *schema.Provider, configureContextFunc schema.ConfigureContextFunc, t *testing.T) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		var diags diag.Diagnostics
		testName := t.Name()

		providerMetas.Lock()
		meta, ok := providerMetas[testName]
		providerMetas.Unlock()

		if ok {
			return meta, nil
		}

		vcrMode, err := vcr.Mode()
		if err != nil {
			return nil, sdkdiag.AppendFromErr(diags, err)
		}

		// Real transport config, cribbed from aws-sdk-go-base.
		httpClient := cleanhttp.DefaultPooledClient()
		transport := httpClient.Transport.(*http.Transport)
		transport.MaxIdleConnsPerHost = 10
		if tlsConfig := transport.TLSClientConfig; tlsConfig == nil {
			tlsConfig = &tls.Config{
				MinVersion: tls.VersionTLS13,
			}
			transport.TLSClientConfig = tlsConfig
		}

		// After capture hook to remove sensitive HTTP headers.
		sensitiveHeaderHook := func(i *cassette.Interaction) error {
			delete(i.Request.Headers, "Authorization")
			delete(i.Request.Headers, "X-Amz-Security-Token")
			return nil
		}

		// Define how VCR will match requests to stored interactions.
		matchFunc := func(r *http.Request, i cassette.Request) bool {
			if r.Method != i.Method {
				return false
			}

			if r.URL.String() != i.URL {
				return false
			}

			if r.Body == nil {
				return true
			}

			var b bytes.Buffer
			if _, err := b.ReadFrom(r.Body); err != nil {
				tflog.Debug(ctx, "Failed to read request body from cassette", map[string]any{
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
				return tfjson.EqualStrings(body, i.Body)

			case "application/xml":
				// XML might be the same, but reordered. Try parsing and comparing.
				var requestXml, cassetteXml any

				if err := xml.Unmarshal([]byte(body), &requestXml); err != nil {
					tflog.Debug(ctx, "Failed to unmarshal request XML", map[string]any{
						"error": err,
					})
					return false
				}

				if err := xml.Unmarshal([]byte(i.Body), &cassetteXml); err != nil {
					tflog.Debug(ctx, "Failed to unmarshal cassette XML", map[string]any{
						"error": err,
					})
					return false
				}

				return reflect.DeepEqual(requestXml, cassetteXml)
			}

			return false
		}

		cassetteName := filepath.Join(vcr.Path(), vcrFileName(testName))

		// Create a VCR recorder around a default HTTP client.
		r, err := recorder.New(cassetteName,
			recorder.WithHook(sensitiveHeaderHook, recorder.AfterCaptureHook),
			recorder.WithMatcher(matchFunc),
			recorder.WithMode(vcrMode),
			recorder.WithRealTransport(httpClient.Transport),
			recorder.WithSkipRequestLatency(true),
		)

		if err != nil {
			return nil, sdkdiag.AppendFromErr(diags, err)
		}

		// Use the wrapped HTTP Client for AWS APIs.
		// As the HTTP client is used in the provider's ConfigureContextFunc
		// we must do this setup before calling the ConfigureContextFunc.
		httpClient.Transport = r
		if v, ok := provider.Meta().(*conns.AWSClient); ok {
			meta = v
		} else {
			meta = new(conns.AWSClient)
		}
		meta.SetHTTPClient(ctx, httpClient)
		provider.SetMeta(meta)

		if v, ds := configureContextFunc(ctx, d); ds.HasError() {
			return nil, append(diags, ds...)
		} else {
			meta = v.(*conns.AWSClient)
		}

		s, err := vcrRandomnessSource(t)
		if err != nil {
			return nil, sdkdiag.AppendFromErr(diags, err)
		}
		meta.SetRandomnessSource(s.source)

		providerMetas.Lock()
		providerMetas[testName] = meta
		providerMetas.Unlock()

		return meta, diags
	}
}

// vcrRandomnessSource returns a rand.Source for VCR testing
//
// In RECORD_ONLY mode, generates a new seed and saves it to a file, using the
// seed for the source.
// In REPLAY_ONLY mode, reads a seed from a file and creates a source from it.
func vcrRandomnessSource(t *testing.T) (*randomnessSource, error) {
	t.Helper()
	testName := t.Name()

	randomnessSources.Lock()
	s, ok := randomnessSources[testName]
	randomnessSources.Unlock()

	if ok {
		return s, nil
	}

	vcrMode, err := vcr.Mode()
	if err != nil {
		return nil, err
	}

	switch vcrMode {
	case recorder.ModeRecordOnly:
		seed := rand.Int63()
		s = &randomnessSource{
			seed:   seed,
			source: rand.NewSource(seed),
		}
	case recorder.ModeReplayOnly:
		seed, err := readSeedFromFile(vcrSeedFile(vcr.Path(), testName))

		if err != nil {
			return nil, fmt.Errorf("no cassette found on disk for %s, please replay this testcase in RECORD_ONLY mode - %w", testName, err)
		}

		s = &randomnessSource{
			seed:   seed,
			source: rand.NewSource(seed),
		}
	default:
		t.Log("unsupported VCR mode")
		t.FailNow()
	}

	randomnessSources.Lock()
	randomnessSources[testName] = s
	randomnessSources.Unlock()

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

// closeVCRRecorder closes the VCR recorder, saving the cassette and randomness seed
func closeVCRRecorder(ctx context.Context, t *testing.T) {
	t.Helper()

	// Don't close the recorder if we're running because of a panic.
	if p := recover(); p != nil {
		panic(p)
	}

	testName := t.Name()
	providerMetas.Lock()
	meta, ok := providerMetas[testName]
	defer providerMetas.Unlock()

	if ok {
		if !t.Failed() && !t.Skipped() {
			if v, ok := meta.HTTPClient(ctx).Transport.(*recorder.Recorder); ok {
				t.Log("stopping VCR recorder")
				if err := v.Stop(); err != nil {
					t.Error(err)
				}
			}
		}

		delete(providerMetas, testName)
	} else {
		t.Log("provider meta not found for test", testName)
	}

	// Save the randomness seed.
	randomnessSources.Lock()
	s, ok := randomnessSources[testName]
	defer randomnessSources.Unlock()

	if ok {
		if !t.Failed() && !t.Skipped() {
			t.Log("persisting randomness seed")
			if err := writeSeedToFile(s.seed, vcrSeedFile(vcr.Path(), t.Name())); err != nil {
				t.Error(err)
			}
		}

		delete(randomnessSources, testName)
	} else {
		t.Log("randomness source not found for test", testName)
	}
}

// ParallelTest wraps resource.ParallelTest, initializing VCR if enabled
func ParallelTest(ctx context.Context, t *testing.T, c resource.TestCase) {
	t.Helper()

	if vcr.IsEnabled() {
		if c.ProtoV5ProviderFactories != nil {
			c.ProtoV5ProviderFactories = vcrEnabledProtoV5ProviderFactories(ctx, t, c.ProtoV5ProviderFactories)
			defer closeVCRRecorder(ctx, t)
		} else {
			t.Skip("go-vcr is not currently supported for test step ProtoV5ProviderFactories")
		}
	}

	resource.ParallelTest(t, c)
}

// Test wraps resource.Test, initializing VCR if enabled
func Test(ctx context.Context, t *testing.T, c resource.TestCase) {
	t.Helper()

	if vcr.IsEnabled() {
		if c.ProtoV5ProviderFactories != nil {
			c.ProtoV5ProviderFactories = vcrEnabledProtoV5ProviderFactories(ctx, t, c.ProtoV5ProviderFactories)
			defer closeVCRRecorder(ctx, t)
		} else {
			t.Skip("go-vcr is not currently supported for test step ProtoV5ProviderFactories")
		}
	}

	resource.Test(t, c)
}

// RandInt is a VCR-friendly replacement for acctest.RandInt
func RandInt(t *testing.T) int {
	t.Helper()

	if !vcr.IsEnabled() {
		return sdkacctest.RandInt()
	}

	s, err := vcrRandomnessSource(t)

	if err != nil {
		t.Fatal(err)
	}

	return rand.New(s.source).Int()
}

// RandomWithPrefix is a VCR-friendly replacement for acctest.RandomWithPrefix
func RandomWithPrefix(t *testing.T, prefix string) string {
	t.Helper()

	return fmt.Sprintf("%s-%d", prefix, RandInt(t))
}

// RandIntRange is a VCR-friendly replacement for acctest.RandIntRange
func RandIntRange(t *testing.T, minInt int, maxInt int) int {
	t.Helper()

	if !vcr.IsEnabled() {
		return sdkacctest.RandIntRange(minInt, maxInt)
	}

	s, err := vcrRandomnessSource(t)

	if err != nil {
		t.Fatal(err)
	}

	return rand.New(s.source).Intn(maxInt-minInt) + minInt
}
