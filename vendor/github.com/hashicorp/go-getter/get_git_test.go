package getter

import (
	"encoding/base64"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var testHasGit bool

func init() {
	if _, err := exec.LookPath("git"); err == nil {
		testHasGit = true
	}
}

func TestGitGetter_impl(t *testing.T) {
	var _ Getter = new(GitGetter)
}

func TestGitGetter(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	g := new(GitGetter)
	dst := tempDir(t)

	repo := testGitRepo(t, "basic")
	repo.commitFile("foo.txt", "hello")

	// With a dir that doesn't exist
	if err := g.Get(dst, repo.url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "foo.txt")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGitGetter_branch(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	g := new(GitGetter)
	dst := tempDir(t)

	repo := testGitRepo(t, "branch")
	repo.git("checkout", "-b", "test-branch")
	repo.commitFile("branch.txt", "branch")

	q := repo.url.Query()
	q.Add("ref", "test-branch")
	repo.url.RawQuery = q.Encode()

	if err := g.Get(dst, repo.url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "branch.txt")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Get again should work
	if err := g.Get(dst, repo.url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath = filepath.Join(dst, "branch.txt")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGitGetter_branchUpdate(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	g := new(GitGetter)
	dst := tempDir(t)

	// First setup the state with a fresh branch
	repo := testGitRepo(t, "branch-update")
	repo.git("checkout", "-b", "test-branch")
	repo.commitFile("branch.txt", "branch")

	// Get the "test-branch" branch
	q := repo.url.Query()
	q.Add("ref", "test-branch")
	repo.url.RawQuery = q.Encode()
	if err := g.Get(dst, repo.url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "branch.txt")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Commit an update to the branch
	repo.commitFile("branch-update.txt", "branch-update")

	// Get again should work
	if err := g.Get(dst, repo.url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath = filepath.Join(dst, "branch-update.txt")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGitGetter_tag(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	g := new(GitGetter)
	dst := tempDir(t)

	repo := testGitRepo(t, "tag")
	repo.commitFile("tag.txt", "tag")
	repo.git("tag", "v1.0")

	q := repo.url.Query()
	q.Add("ref", "v1.0")
	repo.url.RawQuery = q.Encode()

	if err := g.Get(dst, repo.url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "tag.txt")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Get again should work
	if err := g.Get(dst, repo.url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath = filepath.Join(dst, "tag.txt")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGitGetter_GetFile(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	g := new(GitGetter)
	dst := tempFile(t)

	repo := testGitRepo(t, "file")
	repo.commitFile("file.txt", "hello")

	// Download the file
	repo.url.Path = filepath.Join(repo.url.Path, "file.txt")
	if err := g.GetFile(dst, repo.url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}
	assertContents(t, dst, "hello")
}

func TestGitGetter_gitVersion(t *testing.T) {
	dir, err := ioutil.TempDir("", "go-getter")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	script := filepath.Join(dir, "git")
	err = ioutil.WriteFile(
		script,
		[]byte("#!/bin/sh\necho git version 2.0\n"),
		0700)
	if err != nil {
		t.Fatal(err)
	}

	defer func(v string) {
		os.Setenv("PATH", v)
	}(os.Getenv("PATH"))

	os.Setenv("PATH", dir)

	// Asking for a higher version throws an error
	if err := checkGitVersion("2.3"); err == nil {
		t.Fatal("expect git version error")
	}

	// Passes when version is satisfied
	if err := checkGitVersion("1.9"); err != nil {
		t.Fatal(err)
	}
}

func TestGitGetter_sshKey(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	g := new(GitGetter)
	dst := tempDir(t)

	encodedKey := base64.StdEncoding.EncodeToString([]byte(testGitToken))

	u, err := url.Parse("ssh://git@github.com/hashicorp/test-private-repo" +
		"?sshkey=" + encodedKey)
	if err != nil {
		t.Fatal(err)
	}

	if err := g.Get(dst, u); err != nil {
		t.Fatalf("err: %s", err)
	}

	readmePath := filepath.Join(dst, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestGitGetter_submodule(t *testing.T) {
	if !testHasGit {
		t.Log("git not found, skipping")
		t.Skip()
	}

	g := new(GitGetter)
	dst := tempDir(t)

	// Set up the grandchild
	gc := testGitRepo(t, "grandchild")
	gc.commitFile("grandchild.txt", "grandchild")

	// Set up the child
	c := testGitRepo(t, "child")
	c.commitFile("child.txt", "child")
	c.git("submodule", "add", gc.dir)
	c.git("commit", "-m", "Add grandchild submodule")

	// Set up the parent
	p := testGitRepo(t, "parent")
	p.commitFile("parent.txt", "parent")
	p.git("submodule", "add", c.dir)
	p.git("commit", "-m", "Add child submodule")

	// Clone the root repository
	if err := g.Get(dst, p.url); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Check that the files exist
	for _, path := range []string{
		filepath.Join(dst, "parent.txt"),
		filepath.Join(dst, "child", "child.txt"),
		filepath.Join(dst, "child", "grandchild", "grandchild.txt"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
}

func TestGitGetter_setupGitEnv_sshKey(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skipf("skipping on windows since the test requires sh")
		return
	}

	cmd := exec.Command("/bin/sh", "-c", "echo $GIT_SSH_COMMAND")
	setupGitEnv(cmd, "/tmp/foo.pem")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}

	actual := strings.TrimSpace(string(out))
	if actual != "ssh -i /tmp/foo.pem" {
		t.Fatalf("unexpected GIT_SSH_COMMAND: %q", actual)
	}
}

func TestGitGetter_setupGitEnvWithExisting_sshKey(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skipf("skipping on windows since the test requires sh")
		return
	}

	// start with an existing ssh command configuration
	os.Setenv("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=no")
	defer os.Setenv("GIT_SSH_COMMAND", "")

	cmd := exec.Command("/bin/sh", "-c", "echo $GIT_SSH_COMMAND")
	setupGitEnv(cmd, "/tmp/foo.pem")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}

	actual := strings.TrimSpace(string(out))
	if actual != "ssh -o StrictHostKeyChecking=no -i /tmp/foo.pem" {
		t.Fatalf("unexpected GIT_SSH_COMMAND: %q", actual)
	}
}

// gitRepo is a helper struct which controls a single temp git repo.
type gitRepo struct {
	t   *testing.T
	url *url.URL
	dir string
}

// testGitRepo creates a new test git repository.
func testGitRepo(t *testing.T, name string) *gitRepo {
	dir, err := ioutil.TempDir("", "go-getter")
	if err != nil {
		t.Fatal(err)
	}
	dir = filepath.Join(dir, name)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}

	r := &gitRepo{
		t:   t,
		dir: dir,
	}

	url, err := url.Parse("file://" + r.dir)
	if err != nil {
		t.Fatal(err)
	}
	r.url = url

	r.git("init")
	r.git("config", "user.name", "go-getter")
	r.git("config", "user.email", "go-getter@hashicorp.com")

	return r
}

// git runs a git command against the repo.
func (r *gitRepo) git(args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	if err := cmd.Run(); err != nil {
		r.t.Fatal(err)
	}
}

// commitFile writes and commits a text file to the repo.
func (r *gitRepo) commitFile(file, content string) {
	path := filepath.Join(r.dir, file)
	if err := ioutil.WriteFile(path, []byte(content), 0600); err != nil {
		r.t.Fatal(err)
	}
	r.git("add", file)
	r.git("commit", "-m", "Adding "+file)
}

// This is a read-only deploy key for an empty test repository.
var testGitToken = `-----BEGIN RSA PRIVATE KEY-----
MIIEpgIBAAKCAQEArGJ7eweUMiT58m424ZHLu6UordeoTcOTPEMeOjIL2GuVhPU+
Y6sdW3gMKEYFKo5ywXxVgNo8VCI8Ny8+PPfR+BNJaAI+VYNDU5rvD3ecfIjH3We4
VyRbT/PcxNK1XJcE260P6nVXrnNLJQBbsP6tjqSswwVy/9gCiI0aa4GxvK4R1ZPJ
H6ONYXzwgYR0QAH6jhyENe5skbH+40fT2u/I3z99HggqKOCJpgq9JkAWdXdqJPO7
kcGP6I6lTE1Cjpi7GEuVx6iWeflmX3uveOLTJohVkhAzGxIk5rIgbqkDoiNJ1RFl
MxFCc/LkmqdYiW6DgrWZJhlY9wB+YFWi3O/2BwIDAQABAoIBAQCE9LROcMsBXfmV
3SHhGqUrRjg41NOPnt+JpC7FLeJq+pdo5ApJrynGabHewhqr9xBVYUNFTY0oSvts
iLiVJ4K/tohwewJ+y+36ps3pfRSqDIkyoBPSykzPPsQw3l9ZWXU6xaE38Wc+Othj
YoJV4igUk7hX9nT7FSznCwWsk2x1m/w40PVDeWp0VOqGz407oPpirL8wS6yxwrcL
IR/XtEXOiOoJmHMdxlNwVOTdMz5mtCGJcl2IqjLZLP0az0SxAkTLrDeR+R9tTY/T
cbdZS3aBVi/9pXQ9yG+QcVrV1PKGdSzOoS1QB0746n9qW4pM93PoRkeENBAM44Gx
zJvanaqRAoGBANU7HbhkUzBiotEhFlf4uQ3cKFzlSMoJAX27OKR8MDD2vLEL0lBv
biYBntMBU/L3A7nr/oVHJRS3dGVEoJdmvoXB+eCpNhyYiZKDXrPfaY3ifRKvcIoq
XuWYkIGB0X1Djf7Sj6ruSxcm8y6M4l2kQq7bo7HXHvJuPRuG930OzAopAoGBAM72
A0+3xTQrzbHcffPJPw8GUvk8tVmypHojQyXdX283GDW7LYvHd+x6rCNDIdXiZ25L
M3YKEcZMPpjnjEH5CRUHyubocelyRiz7P2Hwj3MOSO5g11nLbSlkLYvoG4uuH8ck
2trIRJ81OnVwwIj61CNMCG3CyYk6GN5ShDCJNWSvAoGBAKScyKrrOJWn8A4GvxsW
9rXOepKMp47hOPd5q5bAEOwb7zu25pwWCjDpG1XGNqrhK01C9PCrJeNCZWcwfdGk
Df1w7JkVyKJ21+314Qx3syNH8EqWigkAANa62wQ/1hwgJOTOZP8Oi4XKGf6b4L1t
69TV1x+Z9Vgu5pnzregrnjVRAoGBAIm1KhjmB4KiTti1BN2sn5fItnb+jRClDEn0
op5UQUcIGsTNyg2C6Onh6h4AckgVwIqj4Rb+tjsCyngFQc83/HIQ4FJqgjk5/zW4
68CoR1rgO2jZ6RDnibgL3z6Db6iucJiajkEbFoX07fPs1T+P3o2p7sXR4TW9AYUU
1L5S3cMjAoGBAKd+zv8xjwN9bw9wGz3l/5lWni6muXpmJ7a43Hj562jspb+moMqM
thGypwYJHZX05VkSk8iXvZehE+Czj6xu9P5FtxKCWgMT6hc8qvCq4n41Ndx59zkN
yuFmGAiAN8bAZgSQYyIUnWENsqFJNkj/HHR4MA/O2gY1zPq/PFCvQ9Q4
-----END RSA PRIVATE KEY-----`
