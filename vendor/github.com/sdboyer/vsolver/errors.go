package vsolver

import (
	"bytes"
	"fmt"
	"strings"
)

type errorLevel uint8

// TODO consistent, sensible way of handling 'type' and 'severity' - or figure
// out that they're not orthogonal and collapse into just 'type'

const (
	warning errorLevel = 1 << iota
	mustResolve
	cannotResolve
)

type traceError interface {
	traceString() string
}

type solveError struct {
	lvl errorLevel
	msg string
}

func newSolveError(msg string, lvl errorLevel) error {
	return &solveError{msg: msg, lvl: lvl}
}

func (e *solveError) Error() string {
	return e.msg
}

type noVersionError struct {
	pn    ProjectIdentifier
	fails []failedVersion
}

func (e *noVersionError) Error() string {
	if len(e.fails) == 0 {
		return fmt.Sprintf("No versions found for project %q.", e.pn.LocalName)
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "No versions of %s met constraints:", e.pn.LocalName)
	for _, f := range e.fails {
		fmt.Fprintf(&buf, "\n\t%s: %s", f.v, f.f.Error())
	}

	return buf.String()
}

func (e *noVersionError) traceString() string {
	if len(e.fails) == 0 {
		return fmt.Sprintf("No versions found")
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "No versions of %s met constraints:", e.pn.LocalName)
	for _, f := range e.fails {
		if te, ok := f.f.(traceError); ok {
			fmt.Fprintf(&buf, "\n  %s: %s", f.v, te.traceString())
		} else {
			fmt.Fprintf(&buf, "\n  %s: %s", f.v, f.f.Error())
		}
	}

	return buf.String()
}

type disjointConstraintFailure struct {
	goal      dependency
	failsib   []dependency
	nofailsib []dependency
	c         Constraint
}

func (e *disjointConstraintFailure) Error() string {
	if len(e.failsib) == 1 {
		str := "Could not introduce %s at %s, as it has a dependency on %s with constraint %s, which has no overlap with existing constraint %s from %s at %s"
		return fmt.Sprintf(str, e.goal.depender.id.errString(), e.goal.depender.v, e.goal.dep.Ident.errString(), e.goal.dep.Constraint.String(), e.failsib[0].dep.Constraint.String(), e.failsib[0].depender.id.errString(), e.failsib[0].depender.v)
	}

	var buf bytes.Buffer

	var sibs []dependency
	if len(e.failsib) > 1 {
		sibs = e.failsib

		str := "Could not introduce %s at %s, as it has a dependency on %s with constraint %s, which has no overlap with the following existing constraints:\n"
		fmt.Fprintf(&buf, str, e.goal.depender.id.errString(), e.goal.depender.v, e.goal.dep.Ident.errString(), e.goal.dep.Constraint.String())
	} else {
		sibs = e.nofailsib

		str := "Could not introduce %s at %s, as it has a dependency on %s with constraint %s, which does not overlap with the intersection of existing constraints from other currently selected packages:\n"
		fmt.Fprintf(&buf, str, e.goal.depender.id.errString(), e.goal.depender.v, e.goal.dep.Ident.errString(), e.goal.dep.Constraint.String())
	}

	for _, c := range sibs {
		fmt.Fprintf(&buf, "\t%s from %s at %s\n", c.dep.Constraint.String(), c.depender.id.errString(), c.depender.v)
	}

	return buf.String()
}

func (e *disjointConstraintFailure) traceString() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "constraint %s on %s disjoint with other dependers:\n", e.goal.dep.Constraint.String(), e.goal.dep.Ident.errString())
	for _, f := range e.failsib {
		fmt.Fprintf(&buf, "%s from %s at %s (no overlap)\n", f.dep.Constraint.String(), f.depender.id.LocalName, f.depender.v)
	}
	for _, f := range e.nofailsib {
		fmt.Fprintf(&buf, "%s from %s at %s (some overlap)\n", f.dep.Constraint.String(), f.depender.id.LocalName, f.depender.v)
	}

	return buf.String()
}

// Indicates that an atom could not be introduced because one of its dep
// constraints does not admit the currently-selected version of the target
// project.
type constraintNotAllowedFailure struct {
	goal dependency
	v    Version
}

func (e *constraintNotAllowedFailure) Error() string {
	str := "Could not introduce %s at %s, as it has a dependency on %s with constraint %s, which does not allow the currently selected version of %s"
	return fmt.Sprintf(str, e.goal.depender.id.errString(), e.goal.depender.v, e.goal.dep.Ident.errString(), e.goal.dep.Constraint, e.v)
}

func (e *constraintNotAllowedFailure) traceString() string {
	str := "%s at %s depends on %s with %s, but that's already selected at %s"
	return fmt.Sprintf(str, e.goal.depender.id.LocalName, e.goal.depender.v, e.goal.dep.Ident.LocalName, e.goal.dep.Constraint, e.v)
}

type versionNotAllowedFailure struct {
	goal       atom
	failparent []dependency
	c          Constraint
}

func (e *versionNotAllowedFailure) Error() string {
	if len(e.failparent) == 1 {
		str := "Could not introduce %s at %s, as it is not allowed by constraint %s from project %s."
		return fmt.Sprintf(str, e.goal.id.errString(), e.goal.v, e.failparent[0].dep.Constraint.String(), e.failparent[0].depender.id.errString())
	}

	var buf bytes.Buffer

	str := "Could not introduce %s at %s, as it is not allowed by constraints from the following projects:\n"
	fmt.Fprintf(&buf, str, e.goal.id.errString(), e.goal.v)

	for _, f := range e.failparent {
		fmt.Fprintf(&buf, "\t%s from %s at %s\n", f.dep.Constraint.String(), f.depender.id.errString(), f.depender.v)
	}

	return buf.String()
}

func (e *versionNotAllowedFailure) traceString() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%s at %s not allowed by constraint %s:\n", e.goal.id.LocalName, e.goal.v, e.c.String())
	for _, f := range e.failparent {
		fmt.Fprintf(&buf, "  %s from %s at %s\n", f.dep.Constraint.String(), f.depender.id.LocalName, f.depender.v)
	}

	return buf.String()
}

type missingSourceFailure struct {
	goal ProjectIdentifier
	prob string
}

func (e *missingSourceFailure) Error() string {
	return fmt.Sprintf(e.prob, e.goal)
}

type badOptsFailure string

func (e badOptsFailure) Error() string {
	return string(e)
}

type sourceMismatchFailure struct {
	shared            ProjectName
	sel               []dependency
	current, mismatch string
	prob              atom
}

func (e *sourceMismatchFailure) Error() string {
	var cur []string
	for _, c := range e.sel {
		cur = append(cur, string(c.depender.id.LocalName))
	}

	str := "Could not introduce %s at %s, as it depends on %s from %s, but %s is already marked as coming from %s by %s"
	return fmt.Sprintf(str, e.prob.id.errString(), e.prob.v, e.shared, e.mismatch, e.shared, e.current, strings.Join(cur, ", "))
}

func (e *sourceMismatchFailure) traceString() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "disagreement on network addr for %s:\n", e.shared)

	fmt.Fprintf(&buf, "  %s from %s\n", e.mismatch, e.prob.id.errString())
	for _, dep := range e.sel {
		fmt.Fprintf(&buf, "  %s from %s\n", e.current, dep.depender.id.errString())
	}

	return buf.String()
}

type errDeppers struct {
	err     error
	deppers []atom
}
type checkeeHasProblemPackagesFailure struct {
	goal    atom
	failpkg map[string]errDeppers
}

func (e *checkeeHasProblemPackagesFailure) Error() string {
	var buf bytes.Buffer
	indent := ""

	if len(e.failpkg) > 1 {
		indent = "\t"
		fmt.Fprintf(
			&buf, "Could not introduce %s at %s due to multiple problematic subpackages:\n",
			e.goal.id.errString(),
			e.goal.v,
		)
	}

	for pkg, errdep := range e.failpkg {
		var cause string
		if errdep.err == nil {
			cause = "is missing"
		} else {
			cause = fmt.Sprintf("does not contain usable Go code (%T).", errdep.err)
		}

		if len(e.failpkg) == 1 {
			fmt.Fprintf(
				&buf, "Could not introduce %s at %s, as its subpackage %s %s.",
				e.goal.id.errString(),
				e.goal.v,
				pkg,
				cause,
			)
		} else {
			fmt.Fprintf(&buf, "\tSubpackage %s %s.", pkg, cause)
		}

		if len(errdep.deppers) == 1 {
			fmt.Fprintf(
				&buf, " (Package is required by %s at %s.)",
				errdep.deppers[0].id.errString(),
				errdep.deppers[0].v,
			)
		} else {
			fmt.Fprintf(&buf, " Package is required by:")
			for _, pa := range errdep.deppers {
				fmt.Fprintf(&buf, "\n%s\t%s at %s", indent, pa.id.errString(), pa.v)
			}
		}
	}

	return buf.String()
}

func (e *checkeeHasProblemPackagesFailure) traceString() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%s at %s has problem subpkg(s):\n", e.goal.id.LocalName, e.goal.v)
	for pkg, errdep := range e.failpkg {
		if errdep.err == nil {
			fmt.Fprintf(&buf, "\t%s is missing; ", pkg)
		} else {
			fmt.Fprintf(&buf, "\t%s has err (%T); ", pkg, errdep.err)
		}

		if len(errdep.deppers) == 1 {
			fmt.Fprintf(
				&buf, "required by %s at %s.",
				errdep.deppers[0].id.errString(),
				errdep.deppers[0].v,
			)
		} else {
			fmt.Fprintf(&buf, " required by:")
			for _, pa := range errdep.deppers {
				fmt.Fprintf(&buf, "\n\t\t%s at %s", pa.id.errString(), pa.v)
			}
		}
	}

	return buf.String()
}

type depHasProblemPackagesFailure struct {
	goal dependency
	v    Version
	pl   []string
	prob map[string]error
}

func (e *depHasProblemPackagesFailure) Error() string {
	fcause := func(pkg string) string {
		var cause string
		if err, has := e.prob[pkg]; has {
			cause = fmt.Sprintf("does not contain usable Go code (%T).", err)
		} else {
			cause = "is missing."
		}
		return cause
	}

	if len(e.pl) == 1 {
		return fmt.Sprintf(
			"Could not introduce %s at %s, as it requires package %s from %s, but in version %s that package %s",
			e.goal.depender.id.errString(),
			e.goal.depender.v,
			e.pl[0],
			e.goal.dep.Ident.errString(),
			e.v,
			fcause(e.pl[0]),
		)
	}

	var buf bytes.Buffer
	fmt.Fprintf(
		&buf, "Could not introduce %s at %s, as it requires problematic packages from %s (current version %s):",
		e.goal.depender.id.errString(),
		e.goal.depender.v,
		e.goal.dep.Ident.errString(),
		e.v,
	)

	for _, pkg := range e.pl {
		fmt.Fprintf(&buf, "\t%s %s", pkg, fcause(pkg))
	}

	return buf.String()
}

func (e *depHasProblemPackagesFailure) traceString() string {
	var buf bytes.Buffer
	fcause := func(pkg string) string {
		var cause string
		if err, has := e.prob[pkg]; has {
			cause = fmt.Sprintf("has parsing err (%T).", err)
		} else {
			cause = "is missing"
		}
		return cause
	}

	fmt.Fprintf(
		&buf, "%s at %s depping on %s at %s has problem subpkg(s):",
		e.goal.depender.id.errString(),
		e.goal.depender.v,
		e.goal.dep.Ident.errString(),
		e.v,
	)

	for _, pkg := range e.pl {
		fmt.Fprintf(&buf, "\t%s %s", pkg, fcause(pkg))
	}

	return buf.String()
}