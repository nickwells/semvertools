package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/nickwells/location.mod/location"
	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/psetter"
	"github.com/nickwells/semver.mod/v3/semver"
	"github.com/nickwells/semverparams.mod/v6/semverparams"
)

// Created: Wed Jan 16 22:49:24 2019
type Prog struct {
	checkSeq     bool
	printSV      bool
	semverChecks semverparams.SemverChecks

	exitStatus int
}

// NewProg returns a new Prog instance with any default values set
func NewProg() *Prog {
	return &Prog{}
}

func main() {
	prog := NewProg()
	ps := makeParamSet(prog)

	ps.Parse()

	var svList []*semver.SV
	if cmdLineSVs := ps.Remainder(); len(cmdLineSVs) > 0 {
		svList = prog.getSVsFromStrings(cmdLineSVs)
	} else {
		svList = prog.getSVsFromStdin()
	}

	if prog.checkSeq {
		prog.seqCheck(svList)
	}
	os.Exit(prog.exitStatus)
}

// seqCheck will compare each entry in the semver list against its
// predecessor and if it is not one greater than it then it will report an
// error.
func (prog *Prog) seqCheck(svl []*semver.SV) {
	var prevSV *semver.SV
	for i, sv := range svl {
		if prevSV != nil {
			prog.chkSequence(prevSV, sv, i)
		}
		prevSV = sv
	}
}

// chkSVPart checks that the parts are in the correct relationship to each
// other.
//
// The checks are that:
//
//	p2 is greater than p1
//
// that
//
//	p2 == p1 + 1
//
//	and all of the p2 subparts are 0
//
// If any checks fail it returns a non-nil error and sets the exitStatus to 1.
func (prog *Prog) chkSVPart(partName string, p1, p2 int, p2subs []int,
) (err error) {
	defer func() {
		if err != nil {
			prog.exitStatus = 1
		}
	}()
	if p1 > p2 {
		return fmt.Errorf("the "+semver.Names+" are out of order:"+
			" the %s version: %d > %d ", partName, p1, p2)
	}
	if p1 < p2 {
		if p2 != p1+1 {
			return fmt.Errorf("the "+semver.Names+" have gaps:"+
				" the %s version has grown by %d (should be 1)",
				partName, p2-p1)
		}
		for _, p := range p2subs {
			if p != 0 {
				return fmt.Errorf(
					"the "+semver.Names+" have gaps:"+
						" the %s version has grown"+
						" but the subsequent parts are not all zero", partName)
			}
		}
	}
	return nil
}

// reportSeqErr reports an error in the list of IDs
func (prog *Prog) reportSeqErr(sv1, sv2 *semver.SV, idx2 int, msg string) {
	fmt.Printf("Bad ID list at: [%d] %s, [%d] %s:\n", idx2-1, sv1, idx2, sv2)
	fmt.Printf("    %s\n", msg)
	prog.exitStatus = 1
}

// chkSequence checks that the two semvers are in order and that sv2 is one
// ahead of sv1, that either the patch, minor or major numbers are greater by
// precisely one
func (prog *Prog) chkSequence(sv1, sv2 *semver.SV, idx2 int) {
	sv1Parts := []int{sv1.Major(), sv1.Minor(), sv1.Patch()}
	sv2Parts := []int{sv2.Major(), sv2.Minor(), sv2.Patch()}
	partNames := []string{"major", "minor", "patch"}

	for i, name := range partNames {
		p1, p2 := sv1Parts[i], sv2Parts[i]
		if p1 != p2 {
			remainder := sv2Parts[i+1:]
			err := prog.chkSVPart(name, p1, p2, remainder)
			if err != nil {
				prog.reportSeqErr(sv1, sv2, idx2, err.Error())
			}
			return
		}
	}

	if semver.Less(sv2, sv1) {
		prog.reportSeqErr(sv1, sv2, idx2,
			"the "+semver.Names+" are out of order:"+
				" the former is greater than the latter"+
				" - check the pre-release IDs")
		return
	}

	if semver.Equals(sv1, sv2) {
		prog.reportSeqErr(sv1, sv2, idx2, "duplicate entries")
		return
	}
}

// makeSV will try to create a semver from the passed string. If the string
// cannot be converted then a nil pointer and an error will be returned and
// the exitStatus will be set to 1. Otherwise a pointer to a well-formed
// semver and a nil error will be returned.
func (prog *Prog) makeSV(s string) (sv *semver.SV, err error) {
	defer func() {
		if err != nil {
			prog.exitStatus = 1
		}
	}()

	sv, err = semver.ParseSV(s)
	if err != nil {
		return nil, err
	}
	err = semver.CheckRules(sv.PreRelIDs(), prog.semverChecks.PreRelIDChecks)
	if err != nil {
		return nil, fmt.Errorf("Bad pre-release IDs: %s", err)
	}
	err = semver.CheckRules(sv.BuildIDs(), prog.semverChecks.BuildIDChecks)
	if err != nil {
		return nil, fmt.Errorf("Bad build IDs: %s", err)
	}
	return sv, nil
}

// getSVsFromStdin will read semver strings from standard input
// and check them. It returns a list of all the valid semvers.
func (prog *Prog) getSVsFromStdin() []*semver.SV {
	svl := []*semver.SV{}

	scanner := bufio.NewScanner(os.Stdin)
	loc := location.New("standard input")

	for scanner.Scan() {
		loc.Incr()
		loc.SetContent(scanner.Text())
		svl = append(svl, prog.mkRptPrt(loc))
	}
	return svl
}

// getSVsFromStrings will read semver strings from the passed list of
// strings and check them. It returns a list of all the valid semvers
func (prog *Prog) getSVsFromStrings(args []string) []*semver.SV {
	svl := make([]*semver.SV, 0, len(args))

	loc := location.New("argument")
	for _, s := range args {
		loc.Incr()
		loc.SetContent(s)
		svl = append(svl, prog.mkRptPrt(loc))
	}
	return svl
}

// mkRptPrt creates a semver from the passed string, reports any
// errors and adds the created semver to the list of semvers. It will also,
// optionally, print the semver.
func (prog *Prog) mkRptPrt(loc *location.L) *semver.SV {
	s, hasContent := loc.Content()
	if !hasContent {
		panic(fmt.Errorf(
			"Program error: the location should have content: %s", loc))
	}

	sv, err := prog.makeSV(s)
	if err != nil {
		fmt.Println(loc)
		fmt.Println("   ", err)
		return nil
	}

	if prog.printSV {
		fmt.Println(sv)
	}

	return sv
}

// addParams adds the program-specific parameters
func (prog *Prog) addParams() param.PSetOptFunc {
	return func(ps *param.PSet) error {
		ps.Add("print", psetter.Bool{Value: &prog.printSV},
			"print the "+semver.Names+
				" after the checks have been completed and passed",
			param.AltNames("p"),
		)

		ps.Add("check-seq", psetter.Bool{Value: &prog.checkSeq},
			"check that the "+semver.Names+
				" are correctly ordered and that there are"+
				" no gaps in the sequence",
			param.AltNames("check-order", "check-list"),
		)

		err := ps.SetRemHandler(param.NullRemHandler{}) // allow trailing params
		if err != nil {
			return err
		}

		return nil
	}
}
