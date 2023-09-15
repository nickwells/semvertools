package main

import (
	"github.com/nickwells/param.mod/v6/param"
	"github.com/nickwells/param.mod/v6/paramset"
	"github.com/nickwells/versionparams.mod/versionparams"
)

// makeParamSet generates the param set ready for parsing
func makeParamSet(prog *Prog) *param.PSet {
	return paramset.NewOrPanic(
		versionparams.AddParams,

		addParams(prog),

		SetGlobalConfigFile,
		SetConfigFile,
		param.SetProgramDescription(
			"Sort semver strings read in from the standard input"+
				" or given on the command line"),
	)
}
