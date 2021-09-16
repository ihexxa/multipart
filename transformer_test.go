package multipart

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

type MockReadSeekCloser struct {
	*bytes.Reader
}

func NewMockReadSeekCloser(reader *bytes.Reader) *MockReadSeekCloser {
	return &MockReadSeekCloser{Reader: reader}
}
func (rc *MockReadSeekCloser) Close() error {
	return nil
}

func TestTransformer(t *testing.T) {
	type HttpRange struct {
		contentType string
		rangeStart  string
		rangeEnd    string
		fileSize    string
		// content     string
	}
	type TestCase struct {
		content     string
		rangeHeader string
		// parts       []*Part
		expectedOut string
	}

	sep := "BOUNDARY"
	t.Run("normal cases", func(t *testing.T) {
		testCases := []*TestCase{
			&TestCase{
				content:     "0123456789",
				rangeHeader: "bytes=0-3, 8-8",
				expectedOut: "\r\n--BOUNDARY\r\nContent-Type: application/octet-stream\r\nContent-Range: bytes 0-3/10\r\n\r\n0123\r\n--BOUNDARY\r\nContent-Type: application/octet-stream\r\nContent-Range: bytes 8-8/10\r\n\r\n8\r\n--BOUNDARY--",
			},
		}

		for _, tc := range testCases {
			parts, err := RangeToParts(tc.rangeHeader, "application/octet-stream", fmt.Sprintf("%d", len(tc.content)))
			if err != nil {
				t.Fatal(err)
			}

			mockFd := &MockReadSeekCloser{bytes.NewReader([]byte(tc.content))}
			w := NewTransformerWithBoundary(mockFd, parts, sep)

			buf := bytes.NewBuffer([]byte{})
			err = w.WriteMultiParts(buf)
			if err != nil {
				t.Fatal(err)
			}

			respBytes, err := ioutil.ReadAll(buf)
			if err != nil {
				t.Fatal(err)
			}
			if string(respBytes) != tc.expectedOut {
				t.Errorf("\nnot equal 1.expected(%d) 2.got(%d)", len(tc.expectedOut), len(string(respBytes)))
				t.Error(tc.expectedOut)
				t.Error(string(respBytes))
			}
		}
	})
}
