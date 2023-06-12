package main

import (
	"github.com/nickwells/param.mod/v5/param"
	"github.com/nickwells/param.mod/v5/param/paramset"
	"github.com/nickwells/semver.mod/v3/semver"
	"github.com/nickwells/semverparams.mod/v6/semverparams"
	"github.com/nickwells/versionparams.mod/versionparams"
)

// makeParamSet generates the param set ready for parsing
func makeParamSet(prog *Prog) *param.PSet {
	return paramset.NewOrPanic(
		versionparams.AddParams,

		addParams(prog),

		semverparams.AddSemverGroup,
		prog.semverVals.AddSemverParam(&prog.semverChecks),
		prog.semverVals.AddIDParams(&prog.semverChecks),
		prog.semverChecks.AddCheckParams(),
		SetGlobalConfigFile,
		SetConfigFile,
		param.SetProgramDescription(
			"This provides tools for manipulating "+semver.Names+
				". You can increment the various parts and set or clear"+
				" the pre-release and build IDs.\n\n"+
				"Alternatively you can supply"+
				" the '"+paramNameReleaseCandidate+"'"+
				" or '"+paramNameRelease+"' parameters"+
				" to start or finish a"+
				"sequence of pre-releases"),
	)
}
