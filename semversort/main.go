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

// Created: Mon Dec 31 10:42:22 2018

// Prog holds program parameters and status
type Prog struct {
	reportBadSV           bool
	reverseSort           bool
	ignoreSemVerWithPRIDs bool
	hideRestOfLine        bool

	errOut io.Writer
}

// NewProg returns a new Prog instance with any default values set
func NewProg() *Prog {
	return &Prog{
		errOut: os.Stdout,
	}
}

func main() {
	prog := NewProg()
	ps := makeParamSet(prog)

	ps.Parse()

	var svList semver.SVList

	var svRestOfLineMap map[string][]string

	if cmdLineSVs := ps.Remainder(); len(cmdLineSVs) > 0 {
		svList = prog.getSVListFromStrings(cmdLineSVs)
	} else {
		svList, svRestOfLineMap = prog.getSVListFromReader(os.Stdin)
	}

	prog.sortList(svList)

	var prevSV semver.SV
	var rolIdx int

	for _, sv := range svList {
		fmt.Print(sv)

		if semver.Equals(&prevSV, sv) {
			rolIdx++
		} else {
			rolIdx = 0
		}

		prevSV = *sv

		if rol, ok := svRestOfLineMap[sv.String()]; ok && !prog.hideRestOfLine {
			fmt.Print(rol[rolIdx])
		}

		fmt.Println()
	}
}

// sortList sorts the list prior to printing, applying the reverseSort flag
func (prog *Prog) sortList(svList semver.SVList) {
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
func (prog *Prog) makeSV(s string, errOut io.Writer) *semver.SV {
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
func (prog *Prog) getSVListFromReader(r io.Reader,
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
func (prog *Prog) getSVListFromStrings(args []string) semver.SVList {
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
