package main

import (
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/paramset"
	"github.com/nickwells/semver.mod/v3/semver"
	"github.com/nickwells/semverparams.mod/v6/semverparams"
	"github.com/nickwells/versionparams.mod/versionparams"
)

// makeParamSet generates the param set ready for parsing
func makeParamSet(prog *prog) *param.PSet {
	return paramset.NewOrPanic(
		versionparams.AddParams,

		prog.addParams(),
		prog.semverChecks.AddCheckParams(),

		semverparams.AddSemverGroup,
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
}
