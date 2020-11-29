package multipart

import (
	"bytes"
	"io"
	"net/textproto"
	"testing"
)

func TestWriter(t *testing.T) {
	type Part struct {
		contentType   string
		rangeStart    int64
		rangeEnd      int64
		contentLength int64
		content       string
	}
	type TestCase struct {
		headers     map[string]string
		parts       []*Part
		expectedOut string
	}

	sep := "THIS_STRING_SEPARATES"
	t.Run("normal cases", func(t *testing.T) {
		testCases := []*TestCase{
			&TestCase{
				headers: map[string]string{
					"Content-Length": "1741",
				},
				parts: []*Part{
					&Part{"application/pdf", 500, 999, 8000, "...the first range..."},
					&Part{"application/pdf", 7000, 7999, 8000, "...the second range"},
				},
				expectedOut: "Content-Type: multipart/byteranges; boundary=THIS_STRING_SEPARATES\r\nContent-Length: 1741\r\n\r\n--THIS_STRING_SEPARATES\r\nContent-Type: application/pdf\r\nContent-Range: bytes 500-999/8000\r\n\r\n...the first range...\r\n--THIS_STRING_SEPARATES\r\nContent-Type: application/pdf\r\nContent-Range: bytes 7000-7999/8000\r\n\r\n...the second range\r\n--THIS_STRING_SEPARATES--\r\n",
			},
		}

		for _, tc := range testCases {
			mimeHeaders := textproto.MIMEHeader{}
			for key, val := range tc.headers {
				mimeHeaders.Add(key, val)
			}

			buf := &bytes.Buffer{}
			w, _ := NewWriterWithBoundary(buf, mimeHeaders, sep)

			var err error
			for _, part := range tc.parts {
				err = w.CreatePart(part.contentType, part.rangeStart, part.rangeEnd, part.contentLength)
				if err != nil {
					t.Fatal(err)
				}

				_, err = io.WriteString(w, part.content)
				if err != nil {
					t.Fatal(err)
				}
			}

			err = w.Close()
			if err != nil {
				t.Fatal(err)
			}

			if buf.String() != tc.expectedOut {
				t.Error("\nnot equal 1.expected 2.got")
				t.Error(tc.expectedOut)
				t.Error(buf.String())
			}
		}
	})
}
