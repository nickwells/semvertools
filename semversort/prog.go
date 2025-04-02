package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"

	"github.com/nickwells/semver.mod/v3/semver"
)

// prog holds program parameters and status
type prog struct {
	reportBadSV           bool
	reverseSort           bool
	ignoreSemVerWithPRIDs bool
	hideRestOfLine        bool
	hideIgnoredPrefix     bool

	ignoredPrefix *regexp.Regexp

	errOut io.Writer
}

// newProg returns a new Prog instance with any default values set
func newProg() *prog {
	return &prog{
		errOut: os.Stdout,
	}
}

// sortList sorts the list prior to printing, applying the reverseSort flag
func (prog *prog) sortList(svList semver.SVList) {
	if prog.reverseSort {
		sort.Sort(sort.Reverse(svList))
	} else {
		sort.Sort(svList)
	}
}

// makeSV will try to create a semver from the passed string. If the string
// cannot be converted or if the semver has pre-release IDs and we are
// ignoring those semvers then a nil pointer will be returned. Otherwise the
// newly created semver is returned
func (prog *prog) makeSV(s string, errOut io.Writer) *semver.SV {
	sv, err := semver.ParseSV(s)
	if err != nil {
		if prog.reportBadSV {
			fmt.Fprintln(errOut, s, ":", err)
		}

		return nil
	}

	if sv.HasPreRelIDs() && prog.ignoreSemVerWithPRIDs {
		return nil
	}

	return sv
}

// getSVListFromReader will read semver strings from the standard input and
// create a SVList from them. It will split each line read into a leading
// string of non-space characters and the rest of the line from the first
// whitespace character to the end. The non-semver remainder of the line is
// stored in a map indexed by the semver which is returned with the list of
// semvers.
func (prog *prog) getSVListFromReader(r io.Reader,
) (
	semver.SVList, map[string][]string,
) {
	re := regexp.MustCompile(`[[:space:]]*([^[:space:]]*)([[:space:]]?.*)`)
	svList := make(semver.SVList, 0)
	svROLMap := make(map[string][]string)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		parts := re.FindStringSubmatch(scanner.Text())

		sv := prog.makeSV(parts[1], prog.errOut)
		if sv == nil {
			continue
		}

		svROLMap[parts[1]] = append(svROLMap[parts[1]], parts[2])

		svList = append(svList, sv)
	}

	return svList, svROLMap
}

// getSVListFromStrings will read semver strings from the passed list of
// strings and create a SVList from them
func (prog *prog) getSVListFromStrings(args []string) semver.SVList {
	svList := make(semver.SVList, 0)

	for _, s := range args {
		sv := prog.makeSV(s, prog.errOut)
		if sv == nil {
			continue
		}

		svList = append(svList, sv)
	}

	return svList
}
