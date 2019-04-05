// semvercheck
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/nickwells/param.mod/v2/param"
	"github.com/nickwells/param.mod/v2/param/paramset"
	"github.com/nickwells/param.mod/v2/param/psetter"
	"github.com/nickwells/semver.mod/semver"
	"github.com/nickwells/semverparams.mod/v2/semverparams"
)

// Created: Wed Jan 16 22:49:24 2019
var printSV bool

func main() {
	ps, err := paramset.New(addParams,
		semverparams.AddIDCheckerParams,
		param.SetProgramDescription(
			"Check the supplied semver strings."+
				" This will read "+semver.Names+" from the standard input or"+
				" passed as arguments following "+param.DfltTerminalParam+"."+
				" For each it will check that it is valid and also"+
				" that it conforms to any additional constraints given."+
				" If all the "+semver.Names+" are valid"+
				" this will exit with zero exit status."+
				" Any invalid value will cause an error will be"+
				" printed to stderr and the progran will terminate with"+
				" exit status of 1.\n"+
				" It is also possible to have the parsed "+semver.Names+
				" printed out after being checked."),
	)
	if err != nil {
		log.Fatal("Couldn't construct the parameter set: ", err)
	}

	ps.Parse()

	if cmdLineSVs := ps.Remainder(); len(cmdLineSVs) > 0 {
		getSVsFromStrings(cmdLineSVs)
	} else {
		getSVsFromStdIn()
	}
}

// makeSV will try to create a semver from the passed string. If the string
// cannot be converted then an error will be printed and the program will
// exit with status = 1.
func makeSV(s string) *semver.SV {
	sv, err := semver.ParseSV(s)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Bad "+semver.Name+":", s)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(semverparams.PreRelIDChecks) > 0 {
		err := semver.CheckRules(sv.PreRelIDs, semverparams.PreRelIDChecks)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Bad "+semver.Name+":", s)
			fmt.Fprintln(os.Stderr, "Bad list of pre-release IDs:", err)
			os.Exit(1)
		}
	}
	if len(semverparams.BuildIDChecks) > 0 {
		err := semver.CheckRules(sv.BuildIDs, semverparams.BuildIDChecks)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Bad "+semver.Name+":", s)
			fmt.Fprintln(os.Stderr, "Bad list of build IDs:", err)
			os.Exit(1)
		}
	}
	return sv
}

// getSVsFromStdIn will read semver strings from the standard input
// and check them
func getSVsFromStdIn() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		sv := makeSV(scanner.Text())
		if printSV {
			fmt.Println(sv)
		}
	}
}

// getSVsFromStrings will read semver strings from the passed list of
// strings and check them
func getSVsFromStrings(args []string) {
	for _, s := range args {
		sv := makeSV(s)
		if printSV {
			fmt.Println(sv)
		}
	}
}

// addParams adds the program-specific parameters
func addParams(ps *param.PSet) error {

	ps.Add("print", psetter.Bool{Value: &printSV},
		"print the "+semver.Names+
			" after the checks have been completed and passed",
	)

	err := ps.SetRemHandler(param.NullRemHandler{}) // allow trailing arguments
	if err != nil {
		return err
	}

	return nil
}
