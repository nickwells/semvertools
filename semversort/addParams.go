package main

import (
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/psetter"
)

// addParams will add parameters to the passed ParamSet
func addParams(prog *Prog) param.PSetOptFunc {
	return func(ps *param.PSet) error {
		ps.Add("report-bad-semver",
			psetter.Bool{Value: &prog.reportBadSV},
			"if this flag is set then an error message will be printed"+
				" when a string cannot be converted to a semver."+
				" Otherwise they are silently ignored",
			param.AltNames("show-err"))

		ps.Add("reverse",
			psetter.Bool{Value: &prog.reverseSort},
			"if this flag is set then the sort will be in reverse order",
			param.AltNames("rev", "r"))

		ps.Add("ignore-pre-rel",
			psetter.Bool{Value: &prog.ignoreSemVerWithPRIDs},
			"if this flag is set then the sort will ignore any"+
				" semantic version numbers which have pre-release IDs",
			param.AltNames("no-pr"))

		ps.Add("only-show-semvers",
			psetter.Bool{Value: &prog.hideRestOfLine},
			"if this flag is set then any text following the semantic"+
				" version number (separated by white space) will not be shown",
			param.AltNames("hide-rest-of-line", "hide"))

		err := ps.SetRemHandler(param.NullRemHandler{}) // allow trailing params
		if err != nil {
			return err
		}

		return nil
	}
}
