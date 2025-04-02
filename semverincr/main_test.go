package main

import (
	"testing"

	"github.com/nickwells/semver.mod/v3/semver"
	"github.com/nickwells/semverparams.mod/v6/semverparams"
	"github.com/nickwells/testhelper.mod/v2/testhelper"
)

func TestIncrPRID(t *testing.T) {
	testCases := []struct {
		testhelper.ID
		testhelper.ExpErr
		prid         string
		pridExpected string
	}{
		{
			ID:           testhelper.MkID("good - whole number"),
			prid:         "12",
			pridExpected: "13",
		},
		{
			ID:           testhelper.MkID("good - with prefix"),
			prid:         "RC12",
			pridExpected: "RC13",
		},
		{
			ID:           testhelper.MkID("good - with prefix and leading zeros"),
			prid:         "RC0012",
			pridExpected: "RC0013",
		},
		{
			ID:           testhelper.MkID("good - with suffix"),
			prid:         "12-RC",
			pridExpected: "13-RC",
		},
		{
			ID:           testhelper.MkID("good - with prefix and suffix"),
			prid:         "RC12-RC",
			pridExpected: "RC13-RC",
		},
		{
			ID:           testhelper.MkID("bad - no numeric part"),
			prid:         "RC-RC",
			pridExpected: "RC-RC",
			ExpErr: testhelper.MkExpErr(
				`The string ("RC-RC") has no numerical part`),
		},
	}

	for _, tc := range testCases {
		s, err := incrNumInStr(tc.prid)
		if s != tc.pridExpected {
			t.Log(tc.IDStr())
			t.Log("\t: expected: '" + tc.pridExpected + "'")
			t.Log("\t:      got: '" + s + "'")
			t.Errorf("\t: unexpected value of incremented prid\n")
		}

		testhelper.CheckExpErr(t, err, tc)
	}
}

func TestSetIDs(t *testing.T) {
	prIDsInit := []string{"prID"}
	prIDsNew := []string{"new-prID"}
	bIDsInit := []string{"bID"}
	bIDsNew := []string{"new-bID"}
	testCases := []struct {
		testhelper.ID
		testhelper.ExpErr
		idPart     string
		prIDs      []string
		bIDs       []string
		svExpected *semver.SV
	}{
		{
			ID:         testhelper.MkID("all - nothing set"),
			idPart:     clearAll,
			svExpected: semver.NewSVOrPanic(1, 2, 3, nil, nil),
		},
		{
			ID:         testhelper.MkID("pre-rel - nothing set"),
			idPart:     clearPRID,
			svExpected: semver.NewSVOrPanic(1, 2, 3, nil, bIDsInit),
		},
		{
			ID:         testhelper.MkID("build - nothing set"),
			idPart:     clearBuild,
			svExpected: semver.NewSVOrPanic(1, 2, 3, prIDsInit, nil),
		},
		{
			ID:         testhelper.MkID("none - nothing set"),
			idPart:     clearNone,
			svExpected: semver.NewSVOrPanic(1, 2, 3, prIDsInit, bIDsInit),
		},
		{
			ID:         testhelper.MkID("none - prIDs set"),
			idPart:     clearNone,
			svExpected: semver.NewSVOrPanic(1, 2, 3, prIDsNew, bIDsInit),
			prIDs:      prIDsNew,
		},
		{
			ID:         testhelper.MkID("none - bIDs set"),
			idPart:     clearNone,
			svExpected: semver.NewSVOrPanic(1, 2, 3, prIDsInit, bIDsNew),
			bIDs:       bIDsNew,
		},
		{
			ID:         testhelper.MkID("all - prIDs set"),
			idPart:     clearAll,
			svExpected: semver.NewSVOrPanic(1, 2, 3, prIDsNew, nil),
			prIDs:      prIDsNew,
		},
		{
			ID:         testhelper.MkID("all - bIDs set"),
			idPart:     clearAll,
			svExpected: semver.NewSVOrPanic(1, 2, 3, nil, bIDsNew),
			bIDs:       bIDsNew,
		},
		{
			ID:         testhelper.MkID("bad choice"),
			idPart:     "bad",
			svExpected: semver.NewSVOrPanic(1, 2, 3, prIDsInit, bIDsInit),
			ExpErr: testhelper.MkExpErr(
				`Unknown choice of IDs to clear: "bad"`),
		},
	}

	for _, tc := range testCases {
		sv, err := semver.NewSV(1, 2, 3, prIDsInit, bIDsInit)
		if err != nil {
			t.Fatal("Cannot create the semver to set the IDs on")
		}

		si := prog{
			semverVals: semverparams.SemverVals{
				SemVer:    *sv,
				PreRelIDs: tc.prIDs,
				BuildIDs:  tc.bIDs,
			},
			clearIDs: string(tc.idPart),
		}

		err = si.setIDs()

		testhelper.CheckExpErr(t, err, tc)

		if !semver.Equals(&si.semverVals.SemVer, tc.svExpected) {
			t.Log(tc.IDStr())
			t.Logf("\t: expected: %s", tc.svExpected)
			t.Logf("\t:      got: %s", si.semverVals.SemVer)
			t.Errorf("\t: unexpected setIDs result\n")
		}
	}
}

func TestIncr(t *testing.T) {
	prIDs := []string{"rc001"}
	bIDs := []string{"bID"}
	testCases := []struct {
		testhelper.ID
		testhelper.ExpErr
		incrPart   string
		svExpected *semver.SV
	}{
		{
			ID:         testhelper.MkID("major"),
			incrPart:   incrMajor,
			svExpected: semver.NewSVOrPanic(2, 0, 0, nil, bIDs),
		},
		{
			ID:         testhelper.MkID("minor"),
			incrPart:   incrMinor,
			svExpected: semver.NewSVOrPanic(1, 3, 0, nil, bIDs),
		},
		{
			ID:         testhelper.MkID("patch"),
			incrPart:   incrPatch,
			svExpected: semver.NewSVOrPanic(1, 2, 4, nil, bIDs),
		},
		{
			ID:         testhelper.MkID("prid"),
			incrPart:   incrPRID,
			svExpected: semver.NewSVOrPanic(1, 2, 3, []string{"rc002"}, bIDs),
		},
		{
			ID:         testhelper.MkID("bad"),
			incrPart:   "bad",
			svExpected: semver.NewSVOrPanic(1, 2, 3, prIDs, bIDs),
			ExpErr:     testhelper.MkExpErr(`Unknown increment choice: "bad"`),
		},
	}

	for _, tc := range testCases {
		sv, err := semver.NewSV(1, 2, 3, prIDs, bIDs)
		if err != nil {
			t.Fatal("Cannot create the semver to set the IDs on")
		}

		si := prog{
			semverVals: semverparams.SemverVals{
				SemVer: *sv,
			},
			incrPart: string(tc.incrPart),
		}
		err = si.incr()

		testhelper.CheckExpErr(t, err, tc)

		if !semver.Equals(&si.semverVals.SemVer, tc.svExpected) {
			t.Log(tc.IDStr())
			t.Logf("\t: expected: %s", tc.svExpected)
			t.Logf("\t:      got: %s", si.semverVals.SemVer)
			t.Errorf("\t: unexpected incr result\n")
		}
	}
}
