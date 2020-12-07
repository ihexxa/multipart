package multipart

import (
	"fmt"
	"io/ioutil"
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
				expectOut: `HTTP/1.1 206 Partial Content
Content-Type: multipart/byteranges; boundary=THIS_STRING_SEPARATES

--THIS_STRING_SEPARATES
Content-Type: application/octet-stream
Content-Range: bytes 1-2/5

01
--THIS_STRING_SEPARATES
Content-Type: application/octet-stream
Content-Range: bytes 3-3/5

1
--THIS_STRING_SEPARATES
Content-Type: application/octet-stream
Content-Range: bytes -2/5

10
--THIS_STRING_SEPARATES
Content-Type: application/octet-stream
Content-Range: bytes 2-/5

110
--THIS_STRING_SEPARATES--`,
			},
			&testCase{
				// unknown file size cases
				src:      "10110",
				dst:      &mockResp{},
				fileName: fileName,
				fileSize: "*",
				ranges:   "bytes=0-1, 3-3",
				expectOut: `HTTP/1.1 206 Partial Content
Content-Type: multipart/byteranges; boundary=THIS_STRING_SEPARATES

--THIS_STRING_SEPARATES
Content-Type: application/octet-stream
Content-Range: bytes 0-1/*

10
--THIS_STRING_SEPARATES
Content-Type: application/octet-stream
Content-Range: bytes 3-3/*

1
--THIS_STRING_SEPARATES--`,
			},
		}

		for _, tc := range testCases {
			reader := strings.NewReader(tc.src)
			parts, err := RangeToParts(tc.ranges, ctype, tc.fileSize)
			if err != nil {
				t.Fatal(err)
			}

			// pr, pw := io.Pipe()
			w, contentLen, err := NewResponseWriterWithBoudary(reader, parts, boundary, true)
			if err != nil {
				t.Fatal(err)
			}

			go w.Write()

			// go WriteResponseWithBoundary(reader, pw, fileName, parts, boundary)
			expectOut := strings.ReplaceAll(tc.expectOut, "\n", "\r\n")

			respBytes, err := ioutil.ReadAll(w)
			if err != nil {
				t.Fatal(err)
			}
			body := string(respBytes)

			if expectOut != body {
				t.Error("resp not equal: 1.expect 2.got")
				t.Error(tc.expectOut)
				t.Error(body)
			}

			headBody := strings.SplitN(body, "\r\n\r\n", 2)
			expectHeadBody := strings.SplitN(expectOut, "\r\n\r\n", 2)

			if contentLen != int64(len([]byte(expectHeadBody[1]))) {
				t.Errorf("content length incorrect: expect(%d) got(%d)", len([]byte(expectHeadBody[1])), contentLen)
			}
			if contentLen != int64(len([]byte(headBody[1]))) {
				t.Errorf("content length & body length unmatch: cLen(%d) contentLen(%d)", contentLen, len([]byte(headBody[1])))
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
				expectOut: `HTTP/1.1 206 Partial Content
Content-Range: bytes 1-2/5

01`,
			},
			&testCase{
				src:      "10110",
				dst:      &mockResp{},
				fileName: fileName,
				fileSize: fmt.Sprintf("%d", len("10110")),
				ranges:   "bytes=2-",
				expectOut: `HTTP/1.1 206 Partial Content
Content-Range: bytes 2-/5

110`,
			},
			&testCase{
				src:      "10110",
				dst:      &mockResp{},
				fileName: fileName,
				fileSize: fmt.Sprintf("%d", len("10110")),
				ranges:   "bytes=-2",
				expectOut: `HTTP/1.1 206 Partial Content
Content-Range: bytes -2/5

10`,
			},
			&testCase{
				src:      "10110",
				dst:      &mockResp{},
				fileName: fileName,
				fileSize: "*",
				ranges:   "bytes=1-2",
				expectOut: `HTTP/1.1 206 Partial Content
Content-Range: bytes 1-2/*

01`,
			},
		}

		for _, tc := range testCases {
			reader := strings.NewReader(tc.src)
			parts, err := RangeToParts(tc.ranges, ctype, tc.fileSize)
			if err != nil {
				t.Fatal(err)
			}

			// pr, pw := io.Pipe()
			w, contentLen, err := NewResponseWriterWithBoudary(reader, parts, boundary, true)
			if err != nil {
				t.Fatal(err)
			}

			go w.Write()

			// go WriteResponseWithBoundary(reader, pw, fileName, parts, boundary)
			expectOut := strings.ReplaceAll(tc.expectOut, "\n", "\r\n")

			respBytes, err := ioutil.ReadAll(w)
			if err != nil {
				t.Fatal(err, 0)
			}

			body := string(respBytes)
			if expectOut != body {
				t.Error("resp not equal: 1.expect 2.got")
				t.Error(tc.expectOut)
				t.Error(body)
			}

			headBody := strings.Split(body, "\r\n\r\n")
			if contentLen != int64(len([]byte(headBody[1]))) {
				t.Errorf("content length incorrect: expect(%d) got(%d)", len([]byte(expectOut)), contentLen)
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
