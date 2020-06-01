// semverincr
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/nickwells/check.mod/check"
	"github.com/nickwells/location.mod/location"
	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/paction"
	"github.com/nickwells/param.mod/v5/param/paramset"
	"github.com/nickwells/param.mod/v5/param/psetter"
	"github.com/nickwells/semver.mod/semver"
	"github.com/nickwells/semverparams.mod/v4/semverparams"
)

type incrString string
type clearString string

const (
	incrMajor = incrString("major")
	incrMinor = incrString("minor")
	incrPatch = incrString("patch")
	incrPRID  = incrString("prid")
	incrLeast = incrString("least")
	incrNone  = incrString("none")

	clearAll   = clearString("all")
	clearNone  = clearString("none")
	clearPRID  = clearString("prid")
	clearBuild = clearString("build")
)

var dfltFirstPreRelIDs []string
var clearIDs = string(clearNone)
var incrPart = string(incrLeast)

// Created: Wed Dec 26 11:19:14 2018

var incrementingParamCounter paction.Counter
var idSettingParamCounter paction.Counter

func main() {
	ps := paramset.NewOrDie(addParams,
		semverparams.AddSVStringParam,
		semverparams.AddIDParams,
		semverparams.AddIDCheckerParams,
		SetGlobalConfigFile,
		SetConfigFile,
		param.SetProgramDescription(
			"This provides tools for manipulating "+semver.Names+
				". You can increment the various parts and set or clear"+
				" the pre-release and build IDs.\n\n"+
				"Alternatively you can supply the 'release-candidate'"+
				" or 'release' parameters to start and finish a"+
				"sequence of pre-release IDs"),
	)
	err := semverparams.SetAttrOnSVStringParam(param.MustBeSet)
	if err != nil {
		log.Fatal(
			"Couldn't set the must-be-set attr on the sv-string parameter: ",
			err)
	}

	ps.Parse()
	sv := semverparams.SemVer

	err = incr(sv, incrString(incrPart))
	if err != nil {
		reportProblem(sv, err.Error())
	}
	err = setIDs(sv, clearString(clearIDs),
		semverparams.PreRelIDs, semverparams.BuildIDs)
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
// passed choice parameter
func incr(sv *semver.SV, choice incrString) error {
	switch choice {
	case incrMajor:
		sv.IncrMajor()
	case incrMinor:
		sv.IncrMinor()
	case incrPatch:
		sv.IncrPatch()
	case incrPRID:
		if len(sv.PreRelIDs) <= 0 {
			return errors.New("Cannot increment the pre-release ID" +
				" as the semver does not have a PRID")
		}
		return incrLastPartOfPRID(sv)
	case incrLeast:
		if len(sv.PreRelIDs) > 0 {
			return incrLastPartOfPRID(sv)
		}
		sv.IncrPatch()
	case incrNone:
	default:
		return errors.New("Unknown increment choice: '" + string(choice) + "'")
	}
	return nil
}

// incrLastPartOfPRID will take the last part of the pre-release ID slice
// (which should have been checked to ensure it's non-empty) and will
// increment any numeric part
func incrLastPartOfPRID(sv *semver.SV) error {
	newPRID, err := incrNumInStr(sv.PreRelIDs[len(sv.PreRelIDs)-1])
	if err != nil {
		return err
	}
	sv.PreRelIDs = append(sv.PreRelIDs[0:len(sv.PreRelIDs)-1], newPRID)
	return nil
}

// incrNumInStr will find the numeric part of the pre-release ID and
// increment it, replacing it in the string in the same place as it was
// found. If it is a wholly numeric string then it will be taken as a number
// and incremented as normal, if it is embedded in a string just that part
// will be incremented. For instance '123' will be changed to '124' but
// 'RC012' will be changed to 'RC013'.
func incrNumInStr(s string) (string, error) {
	findNumPartRE := regexp.MustCompile("([^0-9]*)([0-9]+)(.*)")
	parts := findNumPartRE.FindStringSubmatch(s)

	if parts == nil {
		return s, fmt.Errorf("The string (%q) has no numerical part", s)
	}

	if parts[0] != s {
		return s,
			fmt.Errorf("Only a part of the pre-release ID (%q) is matched: %q",
				s, parts[0])
	}

	if len(parts) != 4 {
		return s, errors.New("The pre-release ID ('" +
			s +
			"') should be split into a (possibly empty) prefix," +
			" one or more digits and a (possibly empty) suffix")
	}

	prefix, numStr, suffix := parts[1], parts[2], parts[3]

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return s, errors.New(
			"Cannot convert the numeric part of the pre-release ID '" +
				numStr +
				"' into a number")
	}
	num++

	if parts[1] == "" && parts[3] == "" {
		s = strconv.Itoa(num)
	} else {
		format := prefix + "%0" + strconv.Itoa(len(numStr)) + "d" + suffix
		s = fmt.Sprintf(format, num)
	}

	return s, nil
}

// setIDs clears the pre-release or build IDs according to the setting of
// the clearIDs parameter and then sets any new values. Note that both
// clearing and setting either of the groups of IDs is possible but the
// setting will take precedence and any clearing is redundant
func setIDs(sv *semver.SV, choice clearString, prIDs, bIDs []string) error {
	switch choice {
	case clearAll:
		sv.ClearPreRelIDs()
		sv.ClearBuildIDs()
	case clearPRID:
		sv.ClearPreRelIDs()
	case clearBuild:
		sv.ClearBuildIDs()
	case clearNone:
	default:
		return errors.New("Unknown choice of IDs to clear: '" +
			string(choice) + "'")
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
	countIDSettingParams := idSettingParamCounter.MakeActionFunc()
	countIncrementingParams := incrementingParamCounter.MakeActionFunc()

	ps.Add("part", psetter.Enum{
		Value: &incrPart,
		AllowedVals: psetter.AllowedVals{
			string(incrNone): "don't increment any part",
			string(incrMajor): "increment the major version." +
				" This will set the minor and patch versions to 0",
			string(incrMinor): "increment the minor version." +
				" This will set the patch version to 0" +
				" but leave the major version unchanged",
			string(incrPatch): "increment just the patch version",
			string(incrPRID): "increment the numeric part of the PRID." +
				" Only the last part of the pre-release ID string" +
				" will be incremented" +
				" and it must contain a sequence of digits." +
				" So, for instance 'RC009XX' changes to 'RC010XX'," +
				" '9' changes to '10' and" +
				" 'rc.1' changes to 'rc.2'" +
				" but 'rc.1.XX' will report an error since" +
				" the last part of the pre-release ID (XX) is not numeric",
			string(incrLeast): "increment the PRID if the semantic" +
				" version number has one, otherwise increment the" +
				" patch version",
		}},
		"which part of the "+semver.Name+" should be incremented."+
			" Incrementing any of "+
			string(incrMajor)+", "+
			string(incrMinor)+" or "+
			string(incrPatch)+
			" will also clear any pre-release IDs"+
			" but will leave any build IDs unchanged."+
			" Supplying new pre-release IDs will set them"+
			" for the resultant "+semver.Name,
		param.AltName("version-part"),
		param.PostAction(countIncrementingParams),
	)

	ps.Add("major", psetter.Nil{},
		"update the major part of the "+semver.Name,
		param.AltName("maj"),
		param.AltName("M"),
		param.PostAction(paction.SetString(&incrPart, string(incrMajor))),
		param.PostAction(countIncrementingParams),
	)

	ps.Add("minor", psetter.Nil{},
		"update the minor part of the "+semver.Name,
		param.AltName("min"),
		param.AltName("m"),
		param.PostAction(paction.SetString(&incrPart, string(incrMinor))),
		param.PostAction(countIncrementingParams),
	)

	ps.Add("patch", psetter.Nil{},
		"update the patch part of the "+semver.Name,
		param.AltName("p"),
		param.PostAction(paction.SetString(&incrPart, string(incrPatch))),
		param.PostAction(countIncrementingParams),
	)

	ps.Add("incr-prid", psetter.Nil{},
		"update the prid part of the "+semver.Name,
		param.PostAction(paction.SetString(&incrPart, string(incrPRID))),
		param.PostAction(countIncrementingParams),
	)

	ps.Add("clear-ids", psetter.Enum{
		Value: &clearIDs,
		AllowedVals: psetter.AllowedVals{
			string(clearNone):  "don't clear any part",
			string(clearAll):   "remove any pre-release or build identifiers",
			string(clearPRID):  "remove any pre-release identifiers",
			string(clearBuild): "remove any build identifiers",
		}},
		"which identifiers should be cleared",
		param.AltName("clear"),
	)

	ps.Add("release-candidate", psetter.Nil{},
		"this will produce a "+semver.Name+" with a pre-release ID."+
			" It sets the pre-release IDs to the value of the default"+
			" pre-release IDs or 'rc.1' if that hasn't been set. It will"+
			" override any value that you have given as a parameter. You"+
			" should increment the "+semver.Name+" as necessary. This gives"+
			" you the start of a sequence of "+semver.Names+" with"+
			" increasing pre-release IDs which can be incremented."+
			" The default behaviour of incrementing the least part of"+
			" the "+semver.Name+" will mean that the pre-release ID"+
			" will be incremented",
		param.AltName("rc"),
		param.PostAction(startReleaseParams),
		param.PostAction(countIDSettingParams),
	)

	ps.Add("release", psetter.Nil{},
		"this will produce a "+semver.Name+" suitable to label a release."+
			" It clears the pre-release IDs and does not increment the"+
			" numeric parts",
		param.AltName("r"),
		param.PostAction(paction.SetString(&incrPart, string(incrNone))),
		param.PostAction(paction.SetString(&clearIDs, string(clearPRID))),
		param.PostAction(countIDSettingParams),
		param.PostAction(countIncrementingParams),
	)

	ps.Add("default-pre-rel-ids",
		psetter.StrList{
			Value: &dfltFirstPreRelIDs,
			Checks: []check.StringSlice{
				check.StringSliceStringCheck(semver.CheckPreRelID),
				check.StringSliceLenGT(0),
			},
		},
		"set the default values for the pre-release IDs. This will be"+
			" used as the initial value for a release candidate",
		param.AltName("default-prids"),
		param.Attrs(param.DontShowInStdUsage),
	)

	ps.AddFinalCheck(checkIncrementingParamCounter)
	ps.AddFinalCheck(checkIDSettingParamCounter)

	return nil
}

// checkIDSettingParamCounter checks that at most one of the pre-release ID
// setting parameters has been set.
func checkIDSettingParamCounter() error {
	if idSettingParamCounter.Count() > 1 {
		return fmt.Errorf("The setting of the pre-release IDs for the %s"+
			" has been specified more than once: %s",
			semver.Name, idSettingParamCounter.SetBy())
	}
	return nil
}

// checkIncrementingParamCounter checks that at most one of the incrementing
// parameters has been set.
func checkIncrementingParamCounter() error {
	if incrementingParamCounter.Count() > 1 {
		return fmt.Errorf("The part of the %s to be incremented"+
			" has been specified more than once: %s",
			semver.Name, incrementingParamCounter.SetBy())
	}
	return nil
}

// startReleaseParams sets the parameters appropriately for the start of a
// set of semver numbers leading up to a release - it will set the
// pre-release ID to the default value
func startReleaseParams(_ location.L, _ *param.ByName, _ []string) error {
	if len(dfltFirstPreRelIDs) == 0 {
		semverparams.PreRelIDs = []string{"rc", "1"}
	} else {
		semverparams.PreRelIDs = dfltFirstPreRelIDs
	}
	return nil
}
