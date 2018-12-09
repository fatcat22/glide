// Package mirrors handles managing mirrors in the running application
package mirrors

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
)

var mirrors []*mirror

type mirror struct {
	OriginalRepo, ReplaceRepo, Vcs string
}

// Get retrieves information about an mirror. It returns.
// - bool if found
// - new repo location
// - vcs type
func Get(repo string) (bool, string, string) {
	for _, mirr := range mirrors {
		if len(repo) < len(mirr.OriginalRepo) {
			continue
		}
		if repo[:len(mirr.OriginalRepo)] == mirr.OriginalRepo {
			newRepo := mirr.ReplaceRepo + repo[len(mirr.OriginalRepo):]
			return true, newRepo, mirr.Vcs
		}
	}

	return false, "", ""
}

// Load pulls the mirrors into memory
func Load() error {
	op := filepath.Join(gpath.Home(), "mirrors.yaml")
	if _, err := os.Stat(op); os.IsNotExist(err) {
		msg.Debug("No mirrors.yaml file exists")
		return nil
	} else if err != nil {
		return err
	}

	yml, err := ioutil.ReadFile(op)
	if err != nil {
		return err
	}
	return loadFromYaml(yml)
}

func loadFromYaml(yml []byte) error {
	ov, err := FromYaml(yml)
	if err != nil {
		return err
	}

	for _, o := range ov.Repos {
		msg.Debug("Found mirror: %s to %s (%s)", o.Original, o.Repo, o.Vcs)
		no := &mirror{
			OriginalRepo: o.Original,
			ReplaceRepo:  o.Repo,
			Vcs:          o.Vcs,
		}
		mirrors = append(mirrors, no)
	}

	sort.SliceStable(mirrors, func(i int, j int) bool {
		return len(mirrors[i].OriginalRepo) >= len(mirrors[j].OriginalRepo)
	})
	return nil
}
