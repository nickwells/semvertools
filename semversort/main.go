package main

import (
	"fmt"
	"os"

	"github.com/nickwells/semver.mod/v3/semver"
)

// Created: Mon Dec 31 10:42:22 2018

func main() {
	prog := newProg()
	ps := makeParamSet(prog)

	ps.Parse()

	var svList semver.SVList

	var svRestOfLineMap map[string][]string

	if cmdLineSVs := ps.Remainder(); len(cmdLineSVs) > 0 {
		svList = prog.getSVListFromStrings(cmdLineSVs)
	} else {
		svList, svRestOfLineMap = prog.getSVListFromReader(os.Stdin)
	}

	prog.sortList(svList)

	var prevSV semver.SV

	var rolIdx int

	for i, sv := range svList {
		fmt.Print(sv)

		if semver.Equals(&prevSV, sv) && i > 0 {
			rolIdx++
		} else {
			rolIdx = 0
		}

		prevSV = *sv

		if rol, ok := svRestOfLineMap[sv.String()]; ok && !prog.hideRestOfLine {
			if rolIdx >= len(rol) {
				continue
			}

			fmt.Print(rol[rolIdx])
		}

		fmt.Println()
	}
}
