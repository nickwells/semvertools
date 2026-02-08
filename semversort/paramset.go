package main

import (
	"github.com/nickwells/param.mod/v7/param"
	"github.com/nickwells/param.mod/v7/paramset"
	"github.com/nickwells/versionparams.mod/versionparams"
)

// makeParamSet generates the param set ready for parsing
func makeParamSet(prog *prog) *param.PSet {
	return paramset.New(
		versionparams.AddParams,

		addParams(prog),

		SetGlobalConfigFile,
		SetConfigFile,

		param.SetTrailingParamsName("semver"),
		param.SetProgramDescription(
			"Sort semver strings read in from the standard input"+
				" or given on the command line"),
	)
}
