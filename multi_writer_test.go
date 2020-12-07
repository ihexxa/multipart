package multipart

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestWriter(t *testing.T) {
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

	sep := "THIS_STRING_SEPARATES"
	t.Run("normal cases", func(t *testing.T) {
		testCases := []*TestCase{
			&TestCase{
				content:     "0123456789",
				rangeHeader: "bytes=0-3, 8-8",
				expectedOut: "\r\n--THIS_STRING_SEPARATES\r\nContent-Type: application/octet-stream\r\nContent-Range: bytes 0-3/10\r\n\r\n0123\r\n--THIS_STRING_SEPARATES\r\nContent-Type: application/octet-stream\r\nContent-Range: bytes 8-8/10\r\n\r\n8\r\n--THIS_STRING_SEPARATES--",
			},
		}

		for _, tc := range testCases {
			parts, err := RangeToParts(tc.rangeHeader, "application/octet-stream", fmt.Sprintf("%d", len(tc.content)))
			if err != nil {
				t.Fatal(err)
			}

			r := bytes.NewReader([]byte(tc.content))
			w := NewWriterWithBoundary(r, parts, sep)
			c := make(chan bool, 0)
			go func() {
				respBytes, err := ioutil.ReadAll(w)
				if err != nil {
					t.Fatal(err)
				}
				if string(respBytes) != tc.expectedOut {
					t.Error("\nnot equal 1.expected 2.got")
					t.Error(tc.expectedOut)
					t.Error(string(respBytes))
				}
				c <- true
			}()

			err = w.WriteMultiParts()
			if err != nil {
				t.Fatal(err)
			}

			err = w.Close()
			if err != nil {
				t.Fatal(err)
			}

			<-c
		}
	})
}
