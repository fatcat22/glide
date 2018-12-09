package mirrors

import (
	"bytes"
	"fmt"
	"testing"
)

var oyml = `
repos:
- original: github.com/Masterminds/semver
  repo: file:///path/to/local/repo
  vcs: git
- original: github.com/Masterminds/atest
  repo: github.com/example/atest
`

var ooutyml = `repos:
- original: github.com/Masterminds/atest
  repo: github.com/example/atest
- original: github.com/Masterminds/semver
  repo: file:///path/to/local/repo
  vcs: git
`

func TestSortMirrors(t *testing.T) {
	ov, err := FromYaml([]byte(oyml))
	if err != nil {
		t.Error("Unable to read mirrors yaml")
	}

	out, err := ov.Marshal()
	if err != nil {
		t.Error("Unable to marshal mirrors yaml")
	}

	if string(out) != ooutyml {
		t.Error("Output mirrors sorting failed")
	}
}

func TestGet(t *testing.T) {
	repoMaps := []struct {
		original string
		repo     string
	}{
		{"https://golang.org/x",
			"https://github.com/golang",
		},
		{
			"https://google.golang.org",
			"https://github.com/google",
		},
		{
			"https://google.golang.org/grpc",
			"https://github.com/grpc/grpc-go",
		},
	}
	checkData := []struct {
		original    string
		expect      string
		expectExist bool
	}{
		{
			"https://golang.org/x/tools",
			"https://github.com/golang/tools",
			true,
		},
		{
			"https://unknown",
			"",
			false,
		},
		{
			//check google.golang.org will not be replaced by https://github.com/google
			"https://google.golang.org/grpc",
			"https://github.com/grpc/grpc-go",
			true,
		},
		{
			"https://google.golang.org/test/fake",
			"https://github.com/google/test/fake",
			true,
		},
	}

	var buf bytes.Buffer
	buf.WriteString("repos:\n")
	for _, m := range repoMaps {
		var b bytes.Buffer
		fmt.Fprintf(&b, "- original: %s\n  repo: %s\n", m.original, m.repo)
		buf.Write(b.Bytes())
	}
	oyml := buf.Bytes()
	fmt.Println(buf.String())

	if err := loadFromYaml(oyml); err != nil {
		t.Fatalf("load from yaml error: %v", err)
	}

	for _, data := range checkData {
		exist, repo, _ := Get(data.original)
		if exist != data.expectExist {
			t.Fatalf("expect repo %s exist %v but result is %v", data.original, data.expectExist, exist)
		}
		if data.expectExist {
			if repo != data.expect {
				t.Fatalf("find a error repo: expect %s but got %s", data.expect, repo)
			}
		}
	}
}
