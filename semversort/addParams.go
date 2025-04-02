package main

import (
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/psetter"
)

const (
	paramNameReportBadSemver = "report-bad-semver"
	paramNameReverse         = "reverse"
	paramNameIgnorePreRel    = "ignore-pre-rel"
	paramNameOnlyShowSemvers = "only-show-semvers"
	paramNameHidePrefix      = "hide-prefix"
	paramNamePrefix          = "prefix"
)

// addParams will add parameters to the passed ParamSet
func addParams(prog *prog) param.PSetOptFunc {
	return func(ps *param.PSet) error {
		ps.Add(paramNameReportBadSemver,
			psetter.Bool{Value: &prog.reportBadSV},
			"if this flag is set then an error message will be printed"+
				" when a string cannot be converted to a semver."+
				" Otherwise they are silently ignored",
			param.AltNames("show-err"))

		ps.Add(paramNameReverse,
			psetter.Bool{Value: &prog.reverseSort},
			"if this flag is set then the sort will be in reverse order",
			param.AltNames("rev", "r"))

		ps.Add(paramNameIgnorePreRel,
			psetter.Bool{Value: &prog.ignoreSemVerWithPRIDs},
			"if this flag is set then the sort will ignore any"+
				" semantic version numbers which have pre-release IDs",
			param.AltNames("no-pr"))

		ps.Add(paramNameOnlyShowSemvers,
			psetter.Bool{Value: &prog.hideRestOfLine},
			"if this flag is set then any text following the semantic"+
				" version number (separated by white space) will not be shown",
			param.AltNames("hide-rest-of-line", "hide"))

		ps.Add(paramNameHidePrefix,
			psetter.Bool{Value: &prog.hideIgnoredPrefix},
			"if this flag is set then any ignored prefix will not be shown",
			param.SeeAlso(paramNamePrefix))

		ps.Add(paramNamePrefix,
			psetter.Regexp{Value: &prog.ignoredPrefix},
			"The pattern to be used to strip the semver of any prefix."+
				" Any text matching this regular expresssion will"+
				" be removed before the line being read is converted"+
				" into a semantic version number.",
			param.SeeAlso(paramNameHidePrefix))

		err := ps.SetRemHandler(param.NullRemHandler{}) // allow trailing params
		if err != nil {
			return err
		}

		return nil
	}
}
