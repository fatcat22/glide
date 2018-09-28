// Package mirrors handles managing mirrors in the running application
package mirrors

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
)

var mirrors map[string]*mirror

func init() {
	mirrors = make(map[string]*mirror)
}

type mirror struct {
	Repo, Vcs string
}

// Get retrieves information about an mirror. It returns.
// - bool if found
// - new repo location
// - vcs type
func Get(k string) (bool, string, string) {
	o, f := mirrors[k]
	if f {
		return true, o.Repo, o.Vcs
	}

	org := k
	for li := strings.LastIndex(org, "/"); li != -1; li = strings.LastIndex(org, "/") {
		org = org[:li]
		o, f := mirrors[org]
		if f {
			return true, o.Repo, o.Vcs
		}
	}

	return false, "", ""
}

// Load pulls the mirrors into memory
func Load() error {
	home := gpath.Home()

	op := filepath.Join(home, "mirrors.yaml")

	var ov *Mirrors
	if _, err := os.Stat(op); os.IsNotExist(err) {
		msg.Debug("No mirrors.yaml file exists")
		ov = &Mirrors{
			Repos: make(MirrorRepos, 0),
		}
		return nil
	} else if err != nil {
		ov = &Mirrors{
			Repos: make(MirrorRepos, 0),
		}
		return err
	}

	var err error
	ov, err = ReadMirrorsFile(op)
	if err != nil {
		return fmt.Errorf("Error reading existing mirrors.yaml file: %s", err)
	}

	msg.Info("Loading mirrors from mirrors.yaml file")
	for _, o := range ov.Repos {
		msg.Debug("Found mirror: %s to %s (%s)", o.Original, o.Repo, o.Vcs)
		no := &mirror{
			Repo: o.Repo,
			Vcs:  o.Vcs,
		}
		mirrors[o.Original] = no
	}

	return nil
}
