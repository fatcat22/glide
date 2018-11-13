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

func loadFile(file string) (map[string]*mirror, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		msg.Debug("No mirrors.yaml file exists")
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	ov, err := ReadMirrorsFile(file)
	if err != nil {
		return nil, fmt.Errorf("Error reading existing mirrors.yaml file: %s", err)
	}

	result := make(map[string]*mirror)
	msg.Info("Loading mirrors from " + file)
	for _, o := range ov.Repos {
		msg.Debug("Found mirror: %s to %s (%s)", o.Original, o.Repo, o.Vcs)
		no := &mirror{
			Repo: o.Repo,
			Vcs:  o.Vcs,
		}
		result[o.Original] = no
	}

	return result, nil
}

func loadGlobal() (map[string]*mirror, error) {
	home := gpath.Home()
	op := filepath.Join(home, "mirrors.yaml")

	return loadFile(op)
}

func loadLocal() (map[string]*mirror, error) {
	file, err := gpath.LocalMirrors()
	if err != nil {
		return nil, err
	}

	return loadFile(file)
}

// Load pulls the mirrors into memory
func Load() error {
	gMirrors, gErr := loadGlobal()
	if gErr != nil {
		gMirrors = nil
	}
	lMirrors, lErr := loadLocal()
	if lErr != nil {
		lMirrors = nil
	}

	if gErr != nil && lErr != nil {
		return fmt.Errorf("load mirrors failed: global error: %v. local error: %v", gErr, lErr)
	}

	if gMirrors != nil {
		mirrors = gMirrors
	}
	if lMirrors != nil {
		for k, v := range lMirrors {
			mirrors[k] = v
		}
	}

	return nil
}
