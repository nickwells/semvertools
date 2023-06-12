package main

import (
	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/psetter"
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

		err := ps.SetRemHandler(param.NullRemHandler{}) // allow trailing params
		if err != nil {
			return err
		}

		return nil
	}
}
