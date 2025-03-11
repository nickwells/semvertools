package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nickwells/semver.mod/v3/semver"
	"github.com/nickwells/testhelper.mod/v2/testhelper"
)

type TC struct {
	testhelper.ID
	input []string

	expErrStr       string
	expSVList       semver.SVList
	expSortedSVList semver.SVList

	reportBadSV           bool
	ignoreSemVerWithPRIDs bool
	reverseSort           bool
}

func TestMakeSVList(t *testing.T) {
	const badSV = "v2.3.4.1"

	testCases := []TC{
		{
			ID:    testhelper.MkID("good - in order"),
			input: []string{"v1.2.3", "v2.3.4"},
			expSVList: semver.SVList{
				semver.NewSVOrPanic(1, 2, 3, nil, nil),
				semver.NewSVOrPanic(2, 3, 4, nil, nil),
			},
			expSortedSVList: semver.SVList{
				semver.NewSVOrPanic(1, 2, 3, nil, nil),
				semver.NewSVOrPanic(2, 3, 4, nil, nil),
			},
		},
		{
			ID:    testhelper.MkID("good - in reverse order"),
			input: []string{"v1.2.3", "v2.3.4"},
			expSVList: semver.SVList{
				semver.NewSVOrPanic(1, 2, 3, nil, nil),
				semver.NewSVOrPanic(2, 3, 4, nil, nil),
			},
			expSortedSVList: semver.SVList{
				semver.NewSVOrPanic(2, 3, 4, nil, nil),
				semver.NewSVOrPanic(1, 2, 3, nil, nil),
			},
			reverseSort: true,
		},
		{
			ID:    testhelper.MkID("good - in order, with pre-release IDs"),
			input: []string{"v1.2.3", "v2.3.4-rc.1", "v2.3.4"},
			expSVList: semver.SVList{
				semver.NewSVOrPanic(1, 2, 3, nil, nil),
				semver.NewSVOrPanic(2, 3, 4, []string{"rc", "1"}, nil),
				semver.NewSVOrPanic(2, 3, 4, nil, nil),
			},
			expSortedSVList: semver.SVList{
				semver.NewSVOrPanic(1, 2, 3, nil, nil),
				semver.NewSVOrPanic(2, 3, 4, []string{"rc", "1"}, nil),
				semver.NewSVOrPanic(2, 3, 4, nil, nil),
			},
		},
		{
			ID:    testhelper.MkID("good - in order, ignore pre-release IDs"),
			input: []string{"v1.2.3", "v2.3.4-rc.1", "v2.3.4"},
			expSVList: semver.SVList{
				semver.NewSVOrPanic(1, 2, 3, nil, nil),
				semver.NewSVOrPanic(2, 3, 4, nil, nil),
			},
			expSortedSVList: semver.SVList{
				semver.NewSVOrPanic(1, 2, 3, nil, nil),
				semver.NewSVOrPanic(2, 3, 4, nil, nil),
			},
			ignoreSemVerWithPRIDs: true,
		},
		{
			ID:    testhelper.MkID("good - in order - has bad semver"),
			input: []string{"v1.2.3", badSV, "v2.3.4"},
			expSVList: semver.SVList{
				semver.NewSVOrPanic(1, 2, 3, nil, nil),
				semver.NewSVOrPanic(2, 3, 4, nil, nil),
			},
			expSortedSVList: semver.SVList{
				semver.NewSVOrPanic(1, 2, 3, nil, nil),
				semver.NewSVOrPanic(2, 3, 4, nil, nil),
			},
		},
		{
			ID:    testhelper.MkID("good - in order - has bad semver"),
			input: []string{"v1.2.3", badSV, "v2.3.4"},
			expSVList: semver.SVList{
				semver.NewSVOrPanic(1, 2, 3, nil, nil),
				semver.NewSVOrPanic(2, 3, 4, nil, nil),
			},
			expSortedSVList: semver.SVList{
				semver.NewSVOrPanic(1, 2, 3, nil, nil),
				semver.NewSVOrPanic(2, 3, 4, nil, nil),
			},
			reportBadSV: true,
			expErrStr: badSV + " : " +
				"bad semantic version ID" +
				" - the patch version: '4.1' is not an integer\n",
		},
	}

	var errBuff bytes.Buffer
	for _, tc := range testCases {
		prog := NewProg()
		prog.reportBadSV = tc.reportBadSV
		prog.ignoreSemVerWithPRIDs = tc.ignoreSemVerWithPRIDs
		prog.reverseSort = tc.reverseSort

		errBuff.Reset()
		prog.errOut = &errBuff
		svList := prog.getSVListFromStrings(tc.input)
		if reportGetDiffs(t, svList, tc, "getSVListFromStrings") {
			continue
		}
		if reportBadErr(t, errBuff.String(), tc, "getSVListFromStrings") {
			continue
		}

		errBuff.Reset()
		svList, _ = prog.getSVListFromReader(
			strings.NewReader(strings.Join(tc.input, "\n")))
		if reportGetDiffs(t, svList, tc, "getSVListFromReader") {
			continue
		}
		if reportBadErr(t, errBuff.String(), tc, "getSVListFromReader") {
			continue
		}

		prog.sortList(svList)
		reportSortDiffs(t, svList, tc)
	}
}

// reportBadErr checks that the error received was as expected and reports if
// not. It will return true if the error text is not as expected, false
// otherwise.
func reportBadErr(t *testing.T, errStr string, tc TC, name string) bool {
	t.Helper()
	if errStr != tc.expErrStr {
		t.Log(tc.IDStr())
		if tc.reportBadSV {
			t.Log("\t: bad semvers should be reported as an error")
		}
		t.Logf("\t: expected: '%s'", tc.expErrStr)
		t.Logf("\t:      got: '%s'", errStr)
		t.Errorf("\t: unexpected error report when calling %s\n", name)
		return true
	}
	return false
}

// reportSortDiffs reports differences from expected results when sorting the
// lists of values. It reports the settings of the relevant flags
func reportSortDiffs(t *testing.T, got semver.SVList, tc TC) bool {
	t.Helper()
	if differs, badIdx := svListsDiffer(got, tc.expSortedSVList); differs {
		t.Log(tc.IDStr())
		if tc.reverseSort {
			t.Log("\t: sorting in reverse order")
		}
		logSVList(t, "expected:", badIdx, tc.expSortedSVList)
		logSVList(t, "got:", badIdx, got)
		t.Errorf("\t: unexpected sort results\n")
		return true
	}
	return false
}

// reportGetDiffs reports differences from expected results when getting the
// lists of values. It reports the settings of the relevant flags
func reportGetDiffs(t *testing.T, got semver.SVList, tc TC, name string) bool {
	t.Helper()
	if differs, badIdx := svListsDiffer(got, tc.expSVList); differs {
		t.Log(tc.IDStr())
		if tc.ignoreSemVerWithPRIDs {
			t.Log("\t: ignore semvers having pre-release IDs")
		}
		logSVList(t, "expected:", badIdx, tc.expSVList)
		logSVList(t, "got:", badIdx, got)
		t.Errorf("\t: unexpected result from %s\n", name)
		return true
	}
	return false
}

// svListsDiffer compares the two passed lists of semantic version numbers
// and returns true and the index of the first differing value if they
// differ, false and -1 otherwise
func svListsDiffer(l1, l2 semver.SVList) (bool, int) {
	for i, sv1 := range l1 {
		if i >= len(l2) {
			return true, i
		}
		sv2 := l2[i]
		if !semver.Equals(sv1, sv2) {
			return true, i
		}
	}
	if len(l1) < len(l2) {
		return true, len(l1)
	}

	return false, -1
}

// logSVList prints the semver list
func logSVList(t *testing.T, prefix string, badIdx int, l semver.SVList) {
	t.Helper()

	t.Log("\t:", prefix)
	intro := "   "
	for i, sv := range l {
		if i == badIdx {
			intro = ">>>"
		}
		t.Log("\t\t", intro, sv.String())
		intro = "   "
	}
}
