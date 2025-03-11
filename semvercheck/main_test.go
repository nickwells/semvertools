package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/nickwells/testhelper.mod/v2/testhelper"
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
		prog := &Prog{}
		fName := filepath.Join(testDataDir, "semvers", tc.ID.Name+".txt")

		f, err := os.Open(fName)
		if err != nil {
			t.Fatal("Unexpected error opening", fName, ":", err)
		}

		fileContent, err := io.ReadAll(f)
		if err != nil {
			t.Fatal("Unexpected error reading from", fName, ":", err)
		}

		fio, err := testhelper.NewStdioFromString(string(fileContent))
		if err != nil {
			t.Error("unexpected error faking IO(1)", err)
			continue
		}

		svl := prog.getSVsFromStdin()

		stdout, _, err := fio.Done()
		if err != nil {
			t.Error("unexpected error retrieving stdout and stderr (1)", err)
			continue
		}

		gfc.Check(t, tc.IDStr()+" - read SVs", tc.ID.Name+".SV.checks",
			stdout)

		fio, err = testhelper.NewStdioFromString("")
		if err != nil {
			t.Error("unexpected error faking IO(2)", err)
			continue
		}

		prog.seqCheck(svl)

		stdout, _, err = fio.Done()
		if err != nil {
			t.Error("unexpected error retrieving stdout and stderr (2)", err)
			continue
		}

		gfc.Check(t, tc.IDStr()+" - check list", tc.ID.Name+".SVList.checks",
			stdout)

		testhelper.DiffInt(t,
			tc.IDStr(), "exit status",
			prog.exitStatus, tc.expExitStatus)
	}
}
