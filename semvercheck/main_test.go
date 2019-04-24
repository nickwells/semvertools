package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/nickwells/testhelper.mod/testhelper"
)

const (
	testDataDir      = "testdata"
	chkResultsSubDir = "checkResults"
)

var updateChkResults = flag.Bool("upd-chk-results", false,
	"update the files holding the results of"+
		" checking the lists of semantic versions")

func TestMakeSlice(t *testing.T) {
	gfc := testhelper.GoldenFileCfg{
		DirNames: []string{testDataDir, chkResultsSubDir},
		Sfx:      "txt",
	}

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
		testhelper.CheckAgainstGoldenFile(t, tc.IDStr()+" - read SVs",
			b.Bytes(),
			gfc.PathName(tc.ID.Name+".SV.checks"), *updateChkResults)

		var b2 bytes.Buffer
		seqCheck(svl, &b2)
		testhelper.CheckAgainstGoldenFile(t, tc.IDStr()+" - check list",
			b2.Bytes(),
			gfc.PathName(tc.ID.Name+".SVList.checks"), *updateChkResults)

		if exitStatus != tc.expExitStatus {
			t.Log(tc.IDStr())
			t.Errorf("\t: exit status (%d) not as expected (%d)",
				exitStatus, tc.expExitStatus)
		}
	}
}
