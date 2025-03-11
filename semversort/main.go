package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/nickwells/semver.mod/v3/semver"
)

// Created: Mon Dec 31 10:42:22 2018

// Prog holds program parameters and status
type Prog struct {
	reportBadSV           bool
	reverseSort           bool
	ignoreSemVerWithPRIDs bool

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

	if cmdLineSVs := ps.Remainder(); len(cmdLineSVs) > 0 {
		svList = prog.getSVListFromStrings(cmdLineSVs)
	} else {
		svList = prog.getSVListFromReader(os.Stdin)
	}

	prog.sortList(svList)

	for _, sv := range svList {
		fmt.Println(sv)
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

// getSVListFromReader will read semver strings from the standard input
// and create a SVList from them
func (prog *Prog) getSVListFromReader(r io.Reader) semver.SVList {
	svList := make(semver.SVList, 0)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		sv := prog.makeSV(scanner.Text(), prog.errOut)
		if sv == nil {
			continue
		}

		svList = append(svList, sv)
	}
	return svList
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
