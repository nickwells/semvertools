// semver
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/nickwells/param.mod/v2/param"
	"github.com/nickwells/param.mod/v2/param/paramset"
	"github.com/nickwells/param.mod/v2/param/psetter"
	"github.com/nickwells/semver.mod/semver"
	"github.com/nickwells/semverparams.mod/v2/semverparams"
)

var clearIDs = "none"
var incrPart = "patch"

// Created: Wed Dec 26 11:19:14 2018

func main() {
	ps, err := paramset.New(addParams,
		semverparams.AddSVStringParam,
		semverparams.AddIDParams,
		semverparams.AddIDCheckerParams,
		param.SetProgramDescription(
			`This provides tools for manipulating semantic version numbers`),
	)
	if err != nil {
		log.Fatal("Couldn't construct the parameter set: ", err)
	}
	err = semverparams.SetAttrOnSVStringParam(param.MustBeSet)
	if err != nil {
		log.Fatal(
			"Couldn't set the must-be-set attr on the sv-string parameter: ",
			err)
	}

	ps.Parse()
	sv := semverparams.SemVer

	if incrPart == "prid" {
		if len(sv.PreRelIDs) == 0 {
			reportProblem(sv, "Cannot increment the pre-release ID"+
				" as the semver does not have a PRID")
		}
	}

	err = incr(sv, incrPart)
	if err != nil {
		reportProblem(sv, err.Error())
	}
	err = setIDs(sv, clearIDs, semverparams.PreRelIDs, semverparams.BuildIDs)
	if err != nil {
		reportProblem(sv, err.Error())
	}

	fmt.Println(sv)
}

// reportProblem reports the semver and the message and exits
func reportProblem(sv *semver.SV, msg string) {
	fmt.Fprintln(os.Stderr, sv)
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

// incr increments the appropriate part of the SemVer according to the
// setting of the incrPart parameter
func incr(sv *semver.SV, choice string) error {
	switch choice {
	case "major":
		sv.IncrMajor()
	case "minor":
		sv.IncrMinor()
	case "patch":
		sv.IncrPatch()
	case "prid":
		newPRID, err := incrPRID(sv.PreRelIDs[len(sv.PreRelIDs)-1])
		if err == nil {
			sv.PreRelIDs = append(sv.PreRelIDs[0:len(sv.PreRelIDs)-1], newPRID)
		}
	case "none":
	default:
		return errors.New("Unknown increment choice: '" + choice + "'")
	}
	return nil
}

// incrPRID will find the numeric part of the pre-release ID and increment
// it, replacing it in the string in the same place as it was found. If it is
// a wholly numeric string then it will be taken as a number and incremented
// as normal, if it is embedded in a string just that part will be
// incremented. For instance '123' will be changed to '124' but 'RC012' will
// be changed to 'RC013'.
func incrPRID(prid string) (string, error) {
	findNumPartRE := regexp.MustCompile("([^0-9]*)([0-9]+)(.*)")
	parts := findNumPartRE.FindStringSubmatch(prid)

	if parts == nil {
		return prid, errors.New("The pre-release ID ('" +
			prid +
			"') has no numerical part")
	}

	if parts[0] != prid {
		return prid, errors.New("Only a part of the pre-release ID ('" +
			prid +
			"') is matched: '" +
			parts[0] +
			"'")
	}

	if len(parts) != 4 {
		return prid, errors.New("The pre-release ID ('" +
			prid +
			"') should be split into a (possibly empty) prefix," +
			" one or more digits and a (possibly empty) suffix")
	}

	prefix, numStr, suffix := parts[1], parts[2], parts[3]

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return prid, errors.New(
			"Cannot convert the numeric part of the pre-release ID '" +
				numStr +
				"' into a number")
	}
	num++

	if parts[1] == "" && parts[3] == "" {
		prid = strconv.Itoa(num)
	} else {
		format := prefix + "%0" + strconv.Itoa(len(numStr)) + "d" + suffix
		prid = fmt.Sprintf(format, num)
	}

	return prid, nil
}

// setIDs clears the pre-release or build IDs according to the setting of
// the clearIDs parameter and then sets any new values. Note that both
// clearing and setting either of the groups of IDs is possible but the
// setting will take precedence and any clearing is redundant
func setIDs(sv *semver.SV, choice string, prIDs, bIDs []string) error {
	switch choice {
	case "all":
		sv.ClearPreRelIDs()
		sv.ClearBuildIDs()
	case "pre-rel":
		sv.ClearPreRelIDs()
	case "build":
		sv.ClearBuildIDs()
	case "none":
	default:
		return errors.New("Unknown choice of IDs to clear: '" + choice + "'")
	}

	if len(prIDs) > 0 {
		err := semver.CheckRules(prIDs, semverparams.PreRelIDChecks)
		if err != nil {
			return errors.New("bad Pre-Release IDs: " + err.Error())
		}
		sv.PreRelIDs = prIDs
	}
	if len(bIDs) > 0 {
		err := semver.CheckRules(bIDs, semverparams.BuildIDChecks)
		if err != nil {
			return errors.New("bad Build IDs: " + err.Error())
		}
		sv.BuildIDs = bIDs
	}
	return nil
}

// addParams will add parameters to the passed PSet
func addParams(ps *param.PSet) error {
	ps.Add("part", psetter.Enum{
		Value: &incrPart,
		AllowedVals: psetter.AValMap{
			"none": "don't increment any part",
			"major": "increment the major version." +
				" This will set the minor and patch versions to 0",
			"minor": "increment the minor version." +
				" This will set the patch version to 0" +
				" but leave the major version unchanged",
			"patch": "increment just the patch version",
			"prid": "increment the numeric part of the PRID." +
				" Only the last part of the pre-release ID string" +
				" will be incremented" +
				" and it must contain a sequence of digits." +
				" So, for instance 'RC009XX' changes to 'RC010XX'," +
				" '9' changes to '10' and" +
				" 'rc.1' changes to 'rc.2'" +
				" but 'rc.1.XX' will report an error since" +
				" the last part of the pre-release ID (XX) is not numeric",
		}},
		"which part of the semantic version number should be incremented."+
			" Any of these will also clear any pre-release IDs"+
			" but will leave any build IDs unchanged."+
			" Supplying new pre-release IDs will set them"+
			" for the resultant semantic version",
		param.AltName("version-part"))

	ps.Add("clear-ids", psetter.Enum{
		Value: &clearIDs,
		AllowedVals: psetter.AValMap{
			"none":    "don't clear any part",
			"all":     "remove any pre-release or build identifiers",
			"pre-rel": "remove any pre-release identifiers",
			"build":   "remove any build identifiers",
		}},
		"which identifiers should be cleared",
		param.AltName("clear"))

	return nil
}
