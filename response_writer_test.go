package multipart

import (
	"fmt"
	"strings"
	"testing"
)

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
				src:       "10110",
				dst:       &mockResp{},
				fileName:  fileName,
				fileSize:  fmt.Sprintf("%d", len("10110")),
				ranges:    "bytes=1-2, 3-4",
				expectOut: "Content-Type: multipart/byteranges; boundary=THIS_STRING_SEPARATES\r\n\r\n--THIS_STRING_SEPARATES\r\nContent-Type: application/pdf\r\nContent-Range: bytes 1-2/5\r\n\r\n0\r\n--THIS_STRING_SEPARATES\r\nContent-Type: application/pdf\r\nContent-Range: bytes 3-4/5\r\n\r\n1\r\n--THIS_STRING_SEPARATES--\r\n",
			},
		}

		for _, tc := range testCases {
			reader := strings.NewReader(tc.src)
			parts, err := RangeIntoParts(tc.ranges, ctype, tc.fileSize)
			if err != nil {
				t.Fatal(err)
			}

			err = WriteResponseWithBoundary(reader, tc.dst, fileName, parts, boundary)
			if err != nil {
				t.Fatal(err)
			}
			if tc.expectOut != tc.dst.String() {
				t.Error("resp not equal: 1.expect 2.got")
				t.Error(tc.expectOut)
				t.Error(tc.dst.String())
			}
		}
	})
}
