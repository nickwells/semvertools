package main

import (
	"testing"

	"github.com/nickwells/semver.mod/semver"
	"github.com/nickwells/testhelper.mod/testhelper"
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
				"The pre-release ID ('RC-RC') has no numerical part"),
		},
	}

	for _, tc := range testCases {
		s, err := incrPRID(tc.prid)
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
	prIDsNew := []string{"new prID"}
	bIDsInit := []string{"bID"}
	bIDsNew := []string{"new bID"}
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
			idPart:     "all",
			svExpected: &semver.SV{1, 2, 3, nil, nil}, // nolint: vet
		},
		{
			ID:         testhelper.MkID("pre-rel - nothing set"),
			idPart:     "pre-rel",
			svExpected: &semver.SV{1, 2, 3, nil, bIDsInit}, // nolint: vet
		},
		{
			ID:         testhelper.MkID("build - nothing set"),
			idPart:     "build",
			svExpected: &semver.SV{1, 2, 3, prIDsInit, nil}, // nolint: vet
		},
		{
			ID:         testhelper.MkID("none - nothing set"),
			idPart:     "none",
			svExpected: &semver.SV{1, 2, 3, prIDsInit, bIDsInit}, // nolint: vet
		},
		{
			ID:         testhelper.MkID("none - prIDs set"),
			idPart:     "none",
			svExpected: &semver.SV{1, 2, 3, prIDsNew, bIDsInit}, // nolint: vet
			prIDs:      prIDsNew,
		},
		{
			ID:         testhelper.MkID("none - bIDs set"),
			idPart:     "none",
			svExpected: &semver.SV{1, 2, 3, prIDsInit, bIDsNew}, // nolint: vet
			bIDs:       bIDsNew,
		},
		{
			ID:         testhelper.MkID("all - prIDs set"),
			idPart:     "all",
			svExpected: &semver.SV{1, 2, 3, prIDsNew, nil}, // nolint: vet
			prIDs:      prIDsNew,
		},
		{
			ID:         testhelper.MkID("all - bIDs set"),
			idPart:     "all",
			svExpected: &semver.SV{1, 2, 3, nil, bIDsNew}, // nolint: vet
			bIDs:       bIDsNew,
		},
		{
			ID:         testhelper.MkID("bad choice"),
			idPart:     "bad",
			svExpected: &semver.SV{1, 2, 3, prIDsInit, bIDsInit}, // nolint: vet
			ExpErr: testhelper.MkExpErr(
				"Unknown choice of IDs to clear: 'bad'"),
		},
	}

	for _, tc := range testCases {
		sv, err := semver.NewSV(1, 2, 3, prIDsInit, bIDsInit)
		if err != nil {
			t.Fatal("Cannot create the semver to set the IDs on")
		}
		err = setIDs(sv, tc.idPart, tc.prIDs, tc.bIDs)
		testhelper.CheckExpErr(t, err, tc)
		if !semver.Equals(sv, tc.svExpected) {
			t.Log(tc.IDStr())
			t.Logf("\t: expected: %s", tc.svExpected)
			t.Logf("\t:      got: %s", sv)
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
			incrPart:   "major",
			svExpected: &semver.SV{2, 0, 0, nil, bIDs}, // nolint: vet
		},
		{
			ID:         testhelper.MkID("minor"),
			incrPart:   "minor",
			svExpected: &semver.SV{1, 3, 0, nil, bIDs}, // nolint: vet
		},
		{
			ID:         testhelper.MkID("patch"),
			incrPart:   "patch",
			svExpected: &semver.SV{1, 2, 4, nil, bIDs}, // nolint: vet
		},
		{
			ID:         testhelper.MkID("prid"),
			incrPart:   "prid",
			svExpected: &semver.SV{1, 2, 3, []string{"rc002"}, bIDs}, // nolint: vet
		},
		{
			ID:         testhelper.MkID("bad"),
			incrPart:   "bad",
			svExpected: &semver.SV{1, 2, 3, prIDs, bIDs}, // nolint: vet
			ExpErr:     testhelper.MkExpErr("Unknown increment choice: 'bad'"),
		},
	}

	for _, tc := range testCases {
		sv, err := semver.NewSV(1, 2, 3, prIDs, bIDs)
		if err != nil {
			t.Fatal("Cannot create the semver to set the IDs on")
		}
		err = incr(sv, tc.incrPart)
		testhelper.CheckExpErr(t, err, tc)
		if !semver.Equals(sv, tc.svExpected) {
			t.Log(tc.IDStr())
			t.Logf("\t: expected: %s", tc.svExpected)
			t.Logf("\t:      got: %s", sv)
			t.Errorf("\t: unexpected incr result\n")
		}
	}
}
