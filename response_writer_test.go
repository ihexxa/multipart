package multipart

import (
	"fmt"
	"strings"
	"testing"
)

func TestWriteMultipartResponse(t *testing.T) {
	fileName := "download.jpg"
	ctype := "application/pdf"
	boundary := "THIS_STRING_SEPARATES"

	type testCase struct {
		src       string
		dst       *mockResp
		fileName  string
		fileSize  string
		ranges    string
		expectOut string
	}

	t.Run("multi parts", func(t *testing.T) {
		testCases := []*testCase{
			&testCase{
				// normal case, single byte, bytes from end, from specific byte to end
				src:      "10110",
				dst:      &mockResp{},
				fileName: fileName,
				fileSize: fmt.Sprintf("%d", len("10110")),
				ranges:   "bytes=1-2, 3-3, -2, 2-",
				expectOut: `Content-Type: multipart/byteranges; boundary=THIS_STRING_SEPARATES

--THIS_STRING_SEPARATES
Content-Type: application/pdf
Content-Range: bytes 1-2/5

01
--THIS_STRING_SEPARATES
Content-Type: application/pdf
Content-Range: bytes 3-3/5

1
--THIS_STRING_SEPARATES
Content-Type: application/pdf
Content-Range: bytes -2/5

10
--THIS_STRING_SEPARATES
Content-Type: application/pdf
Content-Range: bytes 2-/5

110
--THIS_STRING_SEPARATES--
`,
			},
			&testCase{
				// unknown file size cases
				src:      "10110",
				dst:      &mockResp{},
				fileName: fileName,
				fileSize: "*",
				ranges:   "bytes=0-1, 3-3",
				expectOut: `Content-Type: multipart/byteranges; boundary=THIS_STRING_SEPARATES

--THIS_STRING_SEPARATES
Content-Type: application/pdf
Content-Range: bytes 0-1/*

10
--THIS_STRING_SEPARATES
Content-Type: application/pdf
Content-Range: bytes 3-3/*

1
--THIS_STRING_SEPARATES--
`,
			},
		}

		for _, tc := range testCases {
			reader := strings.NewReader(tc.src)
			parts, err := RangeToParts(tc.ranges, ctype, tc.fileSize)
			if err != nil {
				t.Fatal(err)
			}

			err = WriteResponseWithBoundary(reader, tc.dst, fileName, parts, boundary)
			if err != nil {
				t.Fatal(err)
			}
			expectOut := strings.ReplaceAll(tc.expectOut, "\n", "\r\n")
			if expectOut != tc.dst.String() {
				t.Error("resp not equal: 1.expect 2.got")
				t.Error(tc.expectOut)
				t.Error(tc.dst.String())
			}
		}
	})

	t.Run("single parts", func(t *testing.T) {
		testCases := []*testCase{
			&testCase{
				// normal case, single byte, bytes from end, from specific byte to end
				src:      "10110",
				dst:      &mockResp{},
				fileName: fileName,
				fileSize: fmt.Sprintf("%d", len("10110")),
				ranges:   "bytes=1-2",
				expectOut: `Content-Range: bytes 1-2/5
01`,
			},
			&testCase{
				src:      "10110",
				dst:      &mockResp{},
				fileName: fileName,
				fileSize: fmt.Sprintf("%d", len("10110")),
				ranges:   "bytes=2-",
				expectOut: `Content-Range: bytes 2-/5
110`,
			},
			&testCase{
				src:      "10110",
				dst:      &mockResp{},
				fileName: fileName,
				fileSize: fmt.Sprintf("%d", len("10110")),
				ranges:   "bytes=-2",
				expectOut: `Content-Range: bytes -2/5
10`,
			},
			&testCase{
				src:      "10110",
				dst:      &mockResp{},
				fileName: fileName,
				fileSize: "*",
				ranges:   "bytes=1-2",
				expectOut: `Content-Range: bytes 1-2/*
01`,
			},
		}

		for _, tc := range testCases {
			reader := strings.NewReader(tc.src)
			parts, err := RangeToParts(tc.ranges, ctype, tc.fileSize)
			if err != nil {
				t.Fatal(err)
			}

			err = WriteResponseWithBoundary(reader, tc.dst, fileName, parts, boundary)
			if err != nil {
				t.Fatal(err)
			}
			expectOut := strings.ReplaceAll(tc.expectOut, "\n", "\r\n")
			if expectOut != tc.dst.String() {
				t.Error("resp not equal: 1.expect 2.got")
				t.Error(tc.expectOut)
				t.Error(tc.dst.String())
			}
		}
	})

}

type mockResp struct {
	body string
}

func (r *mockResp) Write(m []byte) (int, error) {
	r.body += string(m)
	return len(m), nil
}

func (r *mockResp) Close() error {
	return nil
}

func (r *mockResp) String() string {
	return r.body
}
