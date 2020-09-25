// semvercheck
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/paramset"
	"github.com/nickwells/param.mod/v5/param/psetter"
	"github.com/nickwells/semver.mod/semver"
	"github.com/nickwells/semverparams.mod/v4/semverparams"
)

// Created: Wed Jan 16 22:49:24 2019
var checkList bool
var printSV bool
var exitStatus int

func main() {
	ps, err := paramset.New(addParams,
		semverparams.AddIDCheckerParams,
		SetGlobalConfigFile,
		SetConfigFile,
		param.SetProgramDescription(
			"Check the supplied semver strings."+
				" This will read "+semver.Names+" from the standard input or"+
				" passed as arguments following "+param.DfltTerminalParam+"."+
				" For each it will check that it is valid and also"+
				" that it conforms to any additional constraints given."+
				" If all the "+semver.Names+" are valid"+
				" this will exit with zero exit status."+
				" If an invalid "+semver.Name+" is seen it will print an error"+
				" and the progran will terminate with"+
				" exit status of 1.\n"+
				" It is also possible to have the parsed "+semver.Names+
				" printed out after being checked."),
	)
	if err != nil {
		log.Fatal("Couldn't construct the parameter set: ", err)
	}

	ps.Parse()

	var svList semver.SVList
	if cmdLineSVs := ps.Remainder(); len(cmdLineSVs) > 0 {
		svList = getSVsFromStrings(cmdLineSVs, os.Stdout)
	} else {
		svList = getSVsFromReader(os.Stdin, os.Stdout)
	}

	if checkList {
		seqCheck(svList, os.Stdout)
	}
	os.Exit(exitStatus)
}

// seqCheck will compare each entry in the semver list against its
// predecessor and if it is not one greater than it then it will report an
// error.
func seqCheck(svl semver.SVList, w io.Writer) {
	var prevSV *semver.SV
	for i, sv := range svl {
		if prevSV != nil {
			chkSequence(prevSV, sv, i, w)
		}
		prevSV = sv
	}
}

// chkSVPart checks that the parts are in the correct relationship to each
// other.
//
// The checks are that
//   p1 is not greater than p2
// or that
//   if p1 is less than p2 then it is less by 1 and all of the p2 subparts are 0
//
// It will return the bool flag set to true if it finds an error or if no
// more checks are needed, false otherwise.
// If the returned string is
// non-empty then an error was found and the exitStatus will have been set.
func chkSVPart(partName string, p1, p2 int, p2subs ...int) (bool, string) {
	if p1 > p2 {
		exitStatus = 1
		return true,
			"the " + semver.Names + " are out of order:" +
				" the former has a higher " + partName +
				" version number than the latter"
	}
	if p1 < p2 {
		if p2 != p1+1 {
			exitStatus = 1
			return true,
				"the " + semver.Names + " have gaps:" +
					" the " + partName + " version number has grown" +
					" but by more than 1"
		}
		subnames := []string{"minor", "patch"}
		for i, p := range p2subs {
			if p != 0 {
				exitStatus = 1
				return true,
					"the " + semver.Names + " have gaps:" +
						" the " + partName + " version number has grown" +
						" but the " + subnames[i+2-len(p2subs)] +
						" version number is not zero"
			}
		}
		return true, ""
	}
	return false, ""
}

// chkSequence checks that the two semvers are in order and that sv2 is one
// ahead of sv1, that either the patch, minor or major numbers are greater by
// precisely one
func chkSequence(sv1, sv2 *semver.SV, idx2 int, w io.Writer) {
	done, msg := chkSVPart("major", sv1.Major, sv2.Major, sv2.Minor, sv2.Patch)
	if msg != "" {
		fmt.Fprintf(w, "Bad ID list at: [%d] %s, [%d] %s: %s\n",
			idx2-1, sv1, idx2, sv2, msg)
	}
	if done {
		return
	}

	done, msg = chkSVPart("minor", sv1.Minor, sv2.Minor, sv2.Patch)
	if msg != "" {
		fmt.Fprintf(w, "Bad ID list at: [%d] %s, [%d] %s: %s\n",
			idx2-1, sv1, idx2, sv2, msg)
	}
	if done {
		return
	}

	done, msg = chkSVPart("patch", sv1.Patch, sv2.Patch)
	if msg != "" {
		fmt.Fprintf(w, "Bad ID list at: [%d] %s, [%d] %s: %s\n",
			idx2-1, sv1, idx2, sv2, msg)
	}
	if done {
		return
	}

	if semver.Less(sv2, sv1) {
		fmt.Fprintf(w, "Bad ID list at: [%d] %s, [%d] %s: %s\n",
			idx2-1, sv1, idx2, sv2,
			"the "+semver.Names+" are out of order:"+
				" the former is greater than the latter"+
				" - check the pre-release IDs")
		exitStatus = 1
		return
	}

	if semver.Equals(sv1, sv2) {
		fmt.Fprintf(w, "Bad ID list at: [%d] %s, [%d] %s: %s\n",
			idx2-1, sv1, idx2, sv2, "duplicate entries")
		exitStatus = 1
		return
	}
}

// makeSV will try to create a semver from the passed string. If the string
// cannot be converted then a nil pointer and an error will be returned and
// the exitStatus will be set to 1. Otherwise a pointer to a well-formed
// semver and a nil error will be returned.
func makeSV(s string) (*semver.SV, error) {
	sv, err := semver.ParseSV(s)
	if err != nil {
		exitStatus = 1
		return nil, err
	}
	err = semver.CheckRules(sv.PreRelIDs, semverparams.PreRelIDChecks)
	if err != nil {
		exitStatus = 1
		return nil, fmt.Errorf("Bad list of pre-release IDs: %s", err)
	}
	err = semver.CheckRules(sv.BuildIDs, semverparams.BuildIDChecks)
	if err != nil {
		exitStatus = 1
		return nil, fmt.Errorf("Bad list of build IDs: %s", err)
	}
	return sv, nil
}

// getSVsFromStdIn will read semver strings from the supplied reader
// and check them. It returns a list of all the valid semvers.
func getSVsFromReader(r io.Reader, w io.Writer) semver.SVList {
	svl := semver.SVList{}

	scanner := bufio.NewScanner(r)
	line := 0

	for scanner.Scan() {
		line++
		s := scanner.Text()
		svl = mkRptPrt(s, line, svl, w)
	}
	return svl
}

// getSVsFromStrings will read semver strings from the passed list of
// strings and check them. It returns a list of all the valid semvers
func getSVsFromStrings(args []string, w io.Writer) semver.SVList {
	svl := make(semver.SVList, 0, len(args))

	for i, s := range args {
		svl = mkRptPrt(s, i+1, svl, w)
	}
	return svl
}

// mkRptPrt creates a semver from the passed string, reports any
// errors and adds the created semver to the list of semvers. It will also,
// optionally, print the semver.
func mkRptPrt(s string, idx int, svl semver.SVList, w io.Writer) semver.SVList {
	sv, err := makeSV(s)

	if err != nil {
		fmt.Fprintf(w, "Bad ID: %d : '%s'\n", idx, s)
		fmt.Fprintln(w, "   ", err)
		return svl
	}

	if printSV {
		fmt.Println(sv)
	}

	return append(svl, sv)
}

// addParams adds the program-specific parameters
func addParams(ps *param.PSet) error {

	ps.Add("print", psetter.Bool{Value: &printSV},
		"print the "+semver.Names+
			" after the checks have been completed and passed",
	)

	ps.Add("check-list", psetter.Bool{Value: &checkList},
		"check that the "+semver.Names+
			" are correctly ordered and that there are no gaps in the sequence",
	)

	err := ps.SetRemHandler(param.NullRemHandler{}) // allow trailing arguments
	if err != nil {
		return err
	}

	return nil
}
