package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/paramset"
	"github.com/nickwells/param.mod/v5/param/psetter"
	"github.com/nickwells/semver.mod/v3/semver"
)

// Created: Mon Dec 31 10:42:22 2018

var (
	reportBadSV           bool
	reverseSort           bool
	ignoreSemVerWithPRIDs bool
)

func main() {
	ps, err := paramset.New(addParams,
		SetGlobalConfigFile,
		SetConfigFile,
		param.SetProgramDescription(
			"Sort semver strings read in from the standard input"+
				" or given on the command line"),
	)
	if err != nil {
		log.Fatal("Couldn't construct the parameter set: ", err)
	}

	ps.Parse()

	var svList semver.SVList

	if cmdLineSVs := ps.Remainder(); len(cmdLineSVs) > 0 {
		svList = getSVListFromStrings(cmdLineSVs, os.Stderr)
	} else {
		svList = getSVListFromReader(os.Stdin, os.Stderr)
	}

	sortList(svList)

	for _, sv := range svList {
		fmt.Println(sv)
	}
}

// sortList sorts the list prior to printing, applying the reverseSort flag
func sortList(svList semver.SVList) {
	if reverseSort {
		sort.Sort(sort.Reverse(svList))
	} else {
		sort.Sort(svList)
	}
}

// makeSV will try to create a semver from the passed string. If the string
// cannot be converted or if the semver has pre-release IDs and we are
// ignoring those semvers then a nil pointer will be returned. Otherwise the
// newly created semver is returned
func makeSV(s string, errOut io.Writer) *semver.SV {
	sv, err := semver.ParseSV(s)
	if err != nil {
		if reportBadSV {
			fmt.Fprintln(errOut, err)
		}
		return nil
	}
	if sv.HasPreRelIDs() && ignoreSemVerWithPRIDs {
		return nil
	}
	return sv
}

// getSVListFromReader will read semver strings from the standard input
// and create a SVList from them
func getSVListFromReader(r io.Reader, errOut io.Writer) semver.SVList {
	svList := make(semver.SVList, 0)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		sv := makeSV(scanner.Text(), errOut)
		if sv == nil {
			continue
		}

		svList = append(svList, sv)
	}
	return svList
}

// getSVListFromStrings will read semver strings from the passed list of
// strings and create a SVList from them
func getSVListFromStrings(args []string, errOut io.Writer) semver.SVList {
	svList := make(semver.SVList, 0)

	for _, s := range args {
		sv := makeSV(s, errOut)
		if sv == nil {
			continue
		}

		svList = append(svList, sv)
	}
	return svList
}

// addParams will add parameters to the passed PSet
func addParams(ps *param.PSet) error {
	ps.Add("report-bad-semver", psetter.Bool{Value: &reportBadSV},
		"if this flag is set then an error message will be printed"+
			" when a string cannot be converted to a semver."+
			" Otherwise they are silently ignored",
		param.AltNames("show-err"))

	ps.Add("reverse", psetter.Bool{Value: &reverseSort},
		"if this flag is set then the sort will be in reverse order",
		param.AltNames("rev", "r"))

	ps.Add("ignore-pre-rel", psetter.Bool{Value: &ignoreSemVerWithPRIDs},
		"if this flag is set then the sort will ignore any"+
			" semantic version numbers which have pre-release IDs",
		param.AltNames("no-pr"))

	err := ps.SetRemHandler(param.NullRemHandler{}) // allow trailing arguments
	if err != nil {
		return err
	}

	return nil
}
