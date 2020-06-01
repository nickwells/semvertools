package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/nickwells/testhelper.mod/testhelper"
)

const (
	testDataDir      = "testdata"
	chkResultsSubDir = "checkResults"
)

var gfc = testhelper.GoldenFileCfg{
	DirNames:    []string{testDataDir, chkResultsSubDir},
	Sfx:         "txt",
	UpdFlagName: "upd-chk-results",
}

func init() {
	gfc.AddUpdateFlag()
}

func TestMakeSlice(t *testing.T) {
	testCases := []struct {
		testhelper.ID
		expExitStatus int
	}{
		{
			ID: testhelper.MkID("good"),
		},
		{
			ID:            testhelper.MkID("badSV-no-v"),
			expExitStatus: 1,
		},
		{
			ID:            testhelper.MkID("badSVList-minIncr-non0patch"),
			expExitStatus: 1,
		},
		{
			ID:            testhelper.MkID("badSVList-majIncr-non0minor"),
			expExitStatus: 1,
		},
		{
			ID:            testhelper.MkID("badSVList-majIncr-non0patch"),
			expExitStatus: 1,
		},
		{
			ID:            testhelper.MkID("badSVList-majIncr-not-by-1"),
			expExitStatus: 1,
		},
		{
			ID:            testhelper.MkID("badSVList-minIncr-not-by-1"),
			expExitStatus: 1,
		},
		{
			ID:            testhelper.MkID("badSVList-patchIncr-not-by-1"),
			expExitStatus: 1,
		},
		{
			ID:            testhelper.MkID("badSVList-outOfOrder-major"),
			expExitStatus: 1,
		},
		{
			ID:            testhelper.MkID("badSVList-outOfOrder-minor"),
			expExitStatus: 1,
		},
		{
			ID:            testhelper.MkID("badSVList-outOfOrder-patch"),
			expExitStatus: 1,
		},
		{
			ID:            testhelper.MkID("badSVList-outOfOrder-prIDs"),
			expExitStatus: 1,
		},
		{
			ID:            testhelper.MkID("badSVList-duplicates"),
			expExitStatus: 1,
		},
		{
			ID:            testhelper.MkID("badSVList-multiple"),
			expExitStatus: 1,
		},
	}

	for _, tc := range testCases {
		exitStatus = 0
		fName := filepath.Join(testDataDir, "semvers", tc.ID.Name+".txt")
		f, err := os.Open(fName)
		if err != nil {
			t.Fatal("Unexpected error opening", fName, ":", err)
		}

		var b bytes.Buffer
		svl := getSVsFromReader(f, &b)
		gfc.Check(t, tc.IDStr()+" - read SVs", tc.ID.Name+".SV.checks",
			b.Bytes())

		var b2 bytes.Buffer
		seqCheck(svl, &b2)
		gfc.Check(t, tc.IDStr()+" - check list", tc.ID.Name+".SVList.checks",
			b2.Bytes())

		if exitStatus != tc.expExitStatus {
			t.Log(tc.IDStr())
			t.Errorf("\t: exit status (%d) not as expected (%d)",
				exitStatus, tc.expExitStatus)
		}
	}
}
