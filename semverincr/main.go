package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/paction"
	"github.com/nickwells/param.mod/v5/param/psetter"
	"github.com/nickwells/semver.mod/v3/semver"
	"github.com/nickwells/semverparams.mod/v6/semverparams"
)

// Created: Wed Dec 26 11:19:14 2018

const (
	incrMajor = "major"
	incrMinor = "minor"
	incrPatch = "patch"
	incrPRID  = "prid"
	incrLeast = "least"
	incrNone  = "none"

	clearAll   = "all"
	clearNone  = "none"
	clearPRID  = "prid"
	clearBuild = "build"

	paramNameReleaseCandidate = "release-candidate"
	paramNameRelease          = "release"
	paramNameDfltPRID         = "default-pre-rel-IDs"
)

type Prog struct {
	dfltFirstPreRelIDs []string
	releaseCandidate   bool
	release            bool

	clearIDs string
	incrPart string

	incrParamCounter  paction.Counter
	setIDParamCounter paction.Counter

	semverVals   semverparams.SemverVals
	semverChecks semverparams.SemverChecks
}

// NewProg creates an initialised Prog value
func NewProg() *Prog {
	return &Prog{
		dfltFirstPreRelIDs: []string{"rc", "1"},

		clearIDs:   clearNone,
		incrPart:   incrLeast,
		semverVals: semverparams.SemverVals{SemverAttrs: param.MustBeSet},
	}
}

func main() {
	prog := NewProg()
	ps := makeParamSet(prog)

	ps.Parse()
	sv := &prog.semverVals.SemVer

	err := prog.incr()
	if err != nil {
		reportProblem(sv, err.Error())
	}
	err = prog.setIDs()
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
func (prog *Prog) incr() error {
	if prog.release {
		return nil
	}

	sv := &prog.semverVals.SemVer

	switch prog.incrPart {
	case incrMajor:
		sv.IncrMajor()
	case incrMinor:
		sv.IncrMinor()
	case incrPatch:
		sv.IncrPatch()
	case incrPRID:
		if !sv.HasPreRelIDs() {
			return errors.New("Cannot increment the pre-release ID" +
				" as the semver does not have a PRID")
		}
		return incrLastPartOfPRID(sv)
	case incrLeast:
		if sv.HasPreRelIDs() {
			return incrLastPartOfPRID(sv)
		}
		sv.IncrPatch()
	case incrNone:
	default:
		return fmt.Errorf("Unknown increment choice: %q", prog.incrPart)
	}
	return nil
}

// incrLastPartOfPRID will take the last part of the pre-release ID slice
// (which should have been checked to ensure it's non-empty) and will
// increment any numeric part
func incrLastPartOfPRID(sv *semver.SV) error {
	prIDs := sv.PreRelIDs()
	lastIdx := len(prIDs) - 1

	newVal, err := incrNumInStr(prIDs[lastIdx])
	if err != nil {
		return err
	}
	prIDs[lastIdx] = newVal

	return sv.SetPreRelIDs(prIDs)
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

// clearSemverIDs clears the pre-release or build IDs according to the
// setting of the clearIDs parameter.
func (prog *Prog) clearSemverIDs() error {
	sv := &prog.semverVals.SemVer

	switch prog.clearIDs {
	case clearAll:
		sv.ClearPreRelIDs()
		sv.ClearBuildIDs()
	case clearPRID:
		sv.ClearPreRelIDs()
	case clearBuild:
		sv.ClearBuildIDs()
	case clearNone:
	default:
		return fmt.Errorf("Unknown choice of IDs to clear: %q", prog.clearIDs)
	}

	return nil
}

// setIDs clears the pre-release or build IDs according to the setting of
// the clearIDs parameter and then sets any new values. Note that both
// clearing and setting either of the groups of IDs is possible but the
// setting will take precedence and any clearing is redundant
func (prog *Prog) setIDs() error {
	err := prog.clearSemverIDs()
	if err != nil {
		return err
	}

	sv := &prog.semverVals.SemVer

	if bIDs := prog.semverVals.BuildIDs; len(bIDs) > 0 {
		err := semver.CheckRules(bIDs, prog.semverChecks.BuildIDChecks)
		if err != nil {
			return errors.New("bad Build IDs: " + err.Error())
		}
		err = sv.SetBuildIDs(bIDs)
		if err != nil {
			return errors.New("Cannot set Build IDs: " + err.Error())
		}
	}

	if prog.release {
		sv.ClearPreRelIDs()
		return nil
	}

	if prog.releaseCandidate {
		return sv.SetPreRelIDs(prog.dfltFirstPreRelIDs)
	}

	if prIDs := prog.semverVals.PreRelIDs; len(prIDs) > 0 {
		err := semver.CheckRules(prIDs, prog.semverChecks.PreRelIDChecks)
		if err != nil {
			return errors.New("bad Pre-Release IDs: " + err.Error())
		}
		return sv.SetPreRelIDs(prIDs)
	}
	return nil
}

// addParams will add parameters to the passed PSet
func addParams(prog *Prog) param.PSetOptFunc {
	return func(ps *param.PSet) error {
		var (
			countSetIDParams = prog.setIDParamCounter.MakeActionFunc()
			countIncrParams  = prog.incrParamCounter.MakeActionFunc()
		)
		ps.Add("part",
			psetter.Enum{
				Value: &prog.incrPart,
				AllowedVals: psetter.AllowedVals{
					string(incrNone): "don't increment any part",
					string(incrMajor): "increment the major version." +
						" This will set the minor and patch versions to 0",
					string(incrMinor): "increment the minor version." +
						" This will set the patch version to 0" +
						" but leave the major version unchanged",
					string(incrPatch): "increment just the patch version",
					string(incrPRID): "increment the numeric part of the" +
						" PRID." +
						" Only the last part of the pre-release ID string" +
						" will be incremented" +
						" and it must contain a sequence of digits." +
						" So, for instance 'RC009XX' changes to 'RC010XX'," +
						" '9' changes to '10' and" +
						" 'rc.1' changes to 'rc.2'" +
						" but 'rc.1.XX' will report an error since" +
						" the last part of the pre-release ID (XX) is" +
						" not numeric",
					string(incrLeast): "increment the PRID if the semantic" +
						" version number has one, otherwise increment the" +
						" patch version",
				},
			},
			"which part of the "+semver.Name+" should be incremented."+
				" Incrementing any of "+
				string(incrMajor)+", "+
				string(incrMinor)+" or "+
				string(incrPatch)+
				" will also clear any pre-release IDs"+
				" but will leave any build IDs unchanged."+
				" Supplying new pre-release IDs will set them"+
				" for the resultant "+semver.Name,
			param.AltNames("semver-part"),
			param.PostAction(countIncrParams),
		)

		ps.Add("major", psetter.Nil{},
			"update the major part of the "+semver.Name,
			param.AltNames("maj", "M"),
			param.PostAction(
				paction.SetVal(&prog.incrPart, string(incrMajor))),
			param.PostAction(countIncrParams),
		)

		ps.Add("minor", psetter.Nil{},
			"update the minor part of the "+semver.Name,
			param.AltNames("min", "m"),
			param.PostAction(
				paction.SetVal(&prog.incrPart, string(incrMinor))),
			param.PostAction(countIncrParams),
		)

		ps.Add("patch", psetter.Nil{},
			"update the patch part of the "+semver.Name,
			param.AltNames("p"),
			param.PostAction(
				paction.SetVal(&prog.incrPart, string(incrPatch))),
			param.PostAction(countIncrParams),
		)

		ps.Add("incr-prid", psetter.Nil{},
			"update the prid part of the "+semver.Name,
			param.PostAction(
				paction.SetVal(&prog.incrPart, string(incrPRID))),
			param.PostAction(countIncrParams),
		)

		ps.Add("clear-ids",
			psetter.Enum{
				Value: &prog.clearIDs,
				AllowedVals: psetter.AllowedVals{
					string(clearNone): "don't clear any part",
					string(clearAll): "clear any pre-release &" +
						" build identifiers",
					string(clearPRID):  "clear any pre-release identifiers",
					string(clearBuild): "clear any build identifiers",
				},
			},
			"which identifiers should be cleared",
			param.AltNames("clear"),
			param.PostAction(countSetIDParams),
		)

		ps.Add(paramNameReleaseCandidate,
			psetter.Bool{Value: &prog.releaseCandidate},
			"this will produce a "+semver.Name+" with a pre-release ID."+
				" It sets the pre-release IDs to the value of the default"+
				" pre-release IDs. It will"+
				" override any value that you have given as a parameter. You"+
				" should increment the "+semver.Name+" as necessary."+
				" This gives"+
				" you the start of a sequence of "+semver.Names+" with"+
				" increasing pre-release IDs which can be incremented."+
				" The default behaviour of incrementing the least part of"+
				" the "+semver.Name+" will mean that the pre-release ID"+
				" will be incremented",
			param.AltNames("rc"),
			param.PostAction(countSetIDParams),
			param.SeeAlso(paramNameRelease),
		)

		ps.Add(paramNameRelease, psetter.Bool{Value: &prog.release},
			"this will produce a "+semver.Name+" suitable to label a release."+
				" It clears the pre-release IDs and does not increment the"+
				" numeric parts",
			param.AltNames("r"),
			param.PostAction(
				paction.SetVal(&prog.incrPart, string(incrNone))),
			param.PostAction(countSetIDParams),
			param.PostAction(countIncrParams),
			param.SeeAlso(paramNameReleaseCandidate),
		)

		ps.Add(paramNameDfltPRID,
			semverparams.IDListSetter(&prog.dfltFirstPreRelIDs,
				semver.CheckPreRelID),
			"set the default values for the 1st pre-release IDs. This will be"+
				" used as the initial value for a release candidate",
			param.AltNames("default-prids"),
			param.Attrs(param.DontShowInStdUsage),
		)

		ps.AddFinalCheck(
			checkCounter(
				"The part of the "+semver.Name+" to be incremented",
				&prog.incrParamCounter))
		ps.AddFinalCheck(
			checkCounter(
				"The setting of the pre-release IDs for the "+semver.Name,
				&prog.setIDParamCounter))

		ps.AddFinalCheck(checkReleaseVals(prog))

		return nil
	}
}

// checkReleaseVals checks the release parameters for consistency
//
// - you
// cannot have both release and release candidate set at the same time.
//
// - you cannot have pre-release IDs set if either of these have been set
func checkReleaseVals(prog *Prog) param.FinalCheckFunc {
	return func() error {
		if prog.release && prog.releaseCandidate {
			return fmt.Errorf(
				"both %q and %q parameters have been set,"+
					" only one or neither is allowed",
				paramNameRelease,
				paramNameReleaseCandidate)
		}

		if prog.release &&
			prog.semverVals.PreRelIDsHaveBeenSet() {
			return fmt.Errorf(
				"the %q parameter has been set,"+
					" and pre-release IDs have been set."+
					" No pre-release IDs are used for a release "+semver.Name,
				paramNameRelease)
		}

		if prog.releaseCandidate &&
			prog.semverVals.PreRelIDsHaveBeenSet() {
			return fmt.Errorf(
				"the %q parameter has been set,"+
					" and pre-release IDs have been set."+
					" The pre-release IDs for"+
					" a release candidate "+semver.Name+
					" are taken from the value of the %q parameter",
				paramNameReleaseCandidate,
				paramNameDfltPRID)
		}

		return nil
	}
}

// checkCounter returns an error if more than one of the parameters counted
// by counter has been set.
func checkCounter(name string, counter *paction.Counter) param.FinalCheckFunc {
	return func() error {
		if counter.Count() > 1 {
			return fmt.Errorf("%s has been given more than once: %s",
				name, counter.SetBy())
		}
		return nil
	}
}
