package multipart

import (
	"fmt"
	"testing"
)

// Range format examples:
// Range: <unit>=<range-start>-
// Range: <unit>=<range-start>-<range-end>
// Range: <unit>=<range-start>-<range-end>, <range-start>-<range-end>
func TestRangeToParts(t *testing.T) {
	type testCase struct {
		rangeValue  string
		ctype       string
		fileSize    string
		expectedOut []*Part
		expectedErr error
	}

	ctype := "multipart/byteranges"
	fileSize := "1024"
	compareParts := func(ps1, ps2 []*Part) error {
		if len(ps1) != len(ps2) {
			return fmt.Errorf("length not equal p1(%d) p2(%d)", len(ps1), len(ps2))
		}

		for i := range ps1 {
			if ps1[i].rangeStartInt != ps2[i].rangeStartInt {
				return fmt.Errorf("start not equal p1(%d) p2(%d)", ps1[i].rangeStartInt, ps2[i].rangeStartInt)
			}
			if ps1[i].rangeEndInt != ps2[i].rangeEndInt {
				return fmt.Errorf("end not equal p1(%d) p2(%d)", ps1[i].rangeEndInt, ps2[i].rangeEndInt)
			}
			if ps1[i].fileSizeInt != ps2[i].fileSizeInt {
				return fmt.Errorf("filesize not equal p1(%d) p2(%d)", ps1[i].fileSizeInt, ps2[i].fileSizeInt)
			}
			if ps1[i].contentType != ps2[i].contentType {
				return fmt.Errorf("contentType not equal p1(%s) p2(%s)", ps1[i].contentType, ps2[i].contentType)
			}
		}
		return nil
	}

	t.Run("normal cases", func(t *testing.T) {
		testCases := []*testCase{
			&testCase{
				rangeValue: "bytes=0-1",
				ctype:      ctype,
				fileSize:   fileSize,
				expectedOut: []*Part{
					&Part{
						rangeStartInt: 0,
						rangeEndInt:   1,
						fileSizeInt:   1024,
						contentType:   ctype,
					},
				},
			},
			&testCase{
				rangeValue: "bytes=0-1, 2-3",
				ctype:      ctype,
				fileSize:   fileSize,
				expectedOut: []*Part{
					&Part{
						rangeStartInt: 0,
						rangeEndInt:   1,
						fileSizeInt:   1024,
						contentType:   ctype,
					},
					&Part{
						rangeStartInt: 2,
						rangeEndInt:   3,
						fileSizeInt:   1024,
						contentType:   ctype,
					},
				},
			},
			&testCase{
				rangeValue: "bytes=0-",
				ctype:      ctype,
				fileSize:   fileSize,
				expectedOut: []*Part{
					&Part{
						rangeStartInt: 0,
						rangeEndInt:   1023,
						fileSizeInt:   1024,
						contentType:   ctype,
					},
				},
			},
			&testCase{
				rangeValue: "bytes=-4",
				ctype:      ctype,
				fileSize:   fileSize,
				expectedOut: []*Part{
					&Part{
						rangeStartInt: 1020,
						rangeEndInt:   1023,
						fileSizeInt:   1024,
						contentType:   ctype,
					},
				},
			},
			&testCase{
				rangeValue: "bytes=1-2, -4",
				ctype:      ctype,
				fileSize:   fileSize,
				expectedOut: []*Part{
					&Part{
						rangeStartInt: 1,
						rangeEndInt:   2,
						fileSizeInt:   1024,
						contentType:   ctype,
					},
					&Part{
						rangeStartInt: 1020,
						rangeEndInt:   1023,
						fileSizeInt:   1024,
						contentType:   ctype,
					},
				},
			},
			&testCase{
				rangeValue: "bytes=1-8",
				ctype:      ctype,
				fileSize:   "*",
				expectedOut: []*Part{
					&Part{
						rangeStartInt: 1,
						rangeEndInt:   8,
						fileSizeInt:   -1,
						contentType:   ctype,
					},
				},
			},
		}

		for _, tc := range testCases {
			parts, err := RangeIntoParts(tc.rangeValue, tc.ctype, tc.fileSize)
			if err != nil {
				t.Fatal(err)
			}
			if err = compareParts(parts, tc.expectedOut); err != nil {
				t.Error(err)
			}
		}
	})

	t.Run("invalid cases", func(t *testing.T) {
		testCases := []*testCase{
			&testCase{
				rangeValue: "bytes=1-0",
				ctype:      ctype,
				fileSize:   fileSize,
			},
			&testCase{
				rangeValue: "bytes= - ",
				ctype:      ctype,
				fileSize:   fileSize,
			},
			&testCase{
				rangeValue: "bytes=0-1, -",
				ctype:      ctype,
				fileSize:   fileSize,
			},
			&testCase{
				rangeValue: "bytes=-1",
				ctype:      ctype,
				fileSize:   "*",
			},
			&testCase{
				rangeValue: "bytes=1-",
				ctype:      ctype,
				fileSize:   "*",
			},
			&testCase{
				rangeValue: "bytes=1-1024",
				ctype:      ctype,
				fileSize:   fileSize,
			},
		}

		for i, tc := range testCases {
			_, err := RangeIntoParts(tc.rangeValue, tc.ctype, tc.fileSize)
			if err == nil {
				t.Errorf("case %d should fail", i)
			}
		}
	})
}
