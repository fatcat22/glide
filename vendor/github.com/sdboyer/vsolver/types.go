package vsolver

import "fmt"

type ProjectIdentifier struct {
	LocalName   ProjectName
	NetworkName string
}

func (i ProjectIdentifier) less(j ProjectIdentifier) bool {
	if i.LocalName < j.LocalName {
		return true
	}
	if j.LocalName < i.LocalName {
		return false
	}

	return i.NetworkName < j.NetworkName
}

func (i ProjectIdentifier) eq(j ProjectIdentifier) bool {
	if i.LocalName != j.LocalName {
		return false
	}
	if i.NetworkName == j.NetworkName {
		return true
	}

	if (i.NetworkName == "" && j.NetworkName == string(j.LocalName)) ||
		(j.NetworkName == "" && i.NetworkName == string(i.LocalName)) {
		return true
	}

	return false
}

func (i ProjectIdentifier) netName() string {
	if i.NetworkName == "" {
		return string(i.LocalName)
	}
	return i.NetworkName
}

func (i ProjectIdentifier) errString() string {
	if i.NetworkName == "" || i.NetworkName == string(i.LocalName) {
		return string(i.LocalName)
	}
	return fmt.Sprintf("%s (from %s)", i.LocalName, i.NetworkName)
}

func (i ProjectIdentifier) normalize() ProjectIdentifier {
	if i.NetworkName == "" {
		i.NetworkName = string(i.LocalName)
	}

	return i
}

// bimodalIdentifiers are used to track work to be done in the unselected queue.
type bimodalIdentifier struct {
	id ProjectIdentifier
	pl []string
}

type ProjectName string

type atom struct {
	id ProjectIdentifier
	v  Version
}

type atomWithPackages struct {
	a  atom
	pl []string
}

type ProjectDep struct {
	Ident      ProjectIdentifier
	Constraint Constraint
}

// Package represents a Go package. It contains a subset of the information
// go/build.Package does.
type Package struct {
	ImportPath, CommentPath string
	Name                    string
	Imports                 []string
	TestImports             []string
}

type byImportPath []Package

func (s byImportPath) Len() int           { return len(s) }
func (s byImportPath) Less(i, j int) bool { return s[i].ImportPath < s[j].ImportPath }
func (s byImportPath) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// completeDep (name hopefully to change) provides the whole picture of a
// dependency - the root (repo and project, since currently we assume the two
// are the same) name, a constraint, and the actual packages needed that are
// under that root.
type completeDep struct {
	// The base ProjectDep
	ProjectDep
	// The specific packages required from the ProjectDep
	pl []string
}

type dependency struct {
	depender atom
	dep      completeDep
}