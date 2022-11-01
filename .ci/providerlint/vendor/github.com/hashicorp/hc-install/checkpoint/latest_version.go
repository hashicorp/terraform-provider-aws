package checkpoint

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	checkpoint "github.com/hashicorp/go-checkpoint"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/internal/pubkey"
	rjson "github.com/hashicorp/hc-install/internal/releasesjson"
	isrc "github.com/hashicorp/hc-install/internal/src"
	"github.com/hashicorp/hc-install/internal/validators"
	"github.com/hashicorp/hc-install/product"
)

var (
	defaultTimeout = 30 * time.Second
	discardLogger  = log.New(ioutil.Discard, "", 0)
)

// LatestVersion installs the latest version known to Checkpoint
// to OS temp directory, or to InstallDir (if not empty)
type LatestVersion struct {
	Product                  product.Product
	Timeout                  time.Duration
	SkipChecksumVerification bool
	InstallDir               string

	// ArmoredPublicKey is a public PGP key in ASCII/armor format to use
	// instead of built-in pubkey to verify signature of downloaded checksums
	ArmoredPublicKey string

	logger        *log.Logger
	pathsToRemove []string
}

func (*LatestVersion) IsSourceImpl() isrc.InstallSrcSigil {
	return isrc.InstallSrcSigil{}
}

func (lv *LatestVersion) SetLogger(logger *log.Logger) {
	lv.logger = logger
}

func (lv *LatestVersion) log() *log.Logger {
	if lv.logger == nil {
		return discardLogger
	}
	return lv.logger
}

func (lv *LatestVersion) Validate() error {
	if !validators.IsProductNameValid(lv.Product.Name) {
		return fmt.Errorf("invalid product name: %q", lv.Product.Name)
	}
	if !validators.IsBinaryNameValid(lv.Product.BinaryName()) {
		return fmt.Errorf("invalid binary name: %q", lv.Product.BinaryName())
	}

	return nil
}

func (lv *LatestVersion) Install(ctx context.Context) (string, error) {
	timeout := defaultTimeout
	if lv.Timeout > 0 {
		timeout = lv.Timeout
	}
	ctx, cancelFunc := context.WithTimeout(ctx, timeout)
	defer cancelFunc()

	// TODO: Introduce CheckWithContext to allow for cancellation
	resp, err := checkpoint.Check(&checkpoint.CheckParams{
		Product: lv.Product.Name,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Force:   true,
	})
	if err != nil {
		return "", err
	}

	latestVersion, err := version.NewVersion(resp.CurrentVersion)
	if err != nil {
		return "", err
	}

	if lv.pathsToRemove == nil {
		lv.pathsToRemove = make([]string, 0)
	}

	dstDir := lv.InstallDir
	if dstDir == "" {
		var err error
		dirName := fmt.Sprintf("%s_*", lv.Product.Name)
		dstDir, err = ioutil.TempDir("", dirName)
		if err != nil {
			return "", err
		}
		lv.pathsToRemove = append(lv.pathsToRemove, dstDir)
		lv.log().Printf("created new temp dir at %s", dstDir)
	}
	lv.log().Printf("will install into dir at %s", dstDir)

	rels := rjson.NewReleases()
	rels.SetLogger(lv.log())
	pv, err := rels.GetProductVersion(ctx, lv.Product.Name, latestVersion)
	if err != nil {
		return "", err
	}

	d := &rjson.Downloader{
		Logger:           lv.log(),
		VerifyChecksum:   !lv.SkipChecksumVerification,
		ArmoredPublicKey: pubkey.DefaultPublicKey,
		BaseURL:          rels.BaseURL,
	}
	if lv.ArmoredPublicKey != "" {
		d.ArmoredPublicKey = lv.ArmoredPublicKey
	}
	err = d.DownloadAndUnpack(ctx, pv, dstDir)
	if err != nil {
		return "", err
	}

	execPath := filepath.Join(dstDir, lv.Product.BinaryName())

	lv.pathsToRemove = append(lv.pathsToRemove, execPath)

	lv.log().Printf("changing perms of %s", execPath)
	err = os.Chmod(execPath, 0o700)
	if err != nil {
		return "", err
	}

	return execPath, nil
}

func (lv *LatestVersion) Remove(ctx context.Context) error {
	if lv.pathsToRemove != nil {
		for _, path := range lv.pathsToRemove {
			err := os.RemoveAll(path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
