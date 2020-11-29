package multipart

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"sort"
)

type Writer struct {
	w          io.Writer
	boundary   string
	partClosed []bool
}

func NewWriter(w io.Writer, headers textproto.MIMEHeader) (*Writer, error) {
	boundary := randomBoundary()
	return NewWriterWithBoundary(w, headers, boundary)
}

func NewWriterWithBoundary(w io.Writer, headers textproto.MIMEHeader, boundary string) (*Writer, error) {
	err := writerHeaders(w, headers, boundary)
	if err != nil {
		return nil, err
	}

	return &Writer{
		w:          w,
		boundary:   boundary,
		partClosed: []bool{},
	}, nil
}

func writerHeaders(w io.Writer, headers textproto.MIMEHeader, boundary string) error {
	var buf bytes.Buffer

	_, err := fmt.Fprintf(&buf, "Content-Type: multipart/byteranges; boundary=%s\r\n", boundary)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range headers[k] {
			_, err := fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
			if err != nil {
				return err
			}
		}
	}

	// fmt.Fprintf(&buf, "\r\n")
	_, err = io.Copy(w, &buf)
	return err
}

func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

func (w *Writer) SetBoundary(boundary string) {
	w.boundary = boundary
}

func (w *Writer) CreatePart(contentType string, rangeStart, rangeEnd, contentLength int64) error {
	if len(w.partClosed) != 0 {
		w.partClosed[len(w.partClosed)-1] = true
	}

	w.partClosed = append(w.partClosed, false)

	var newPart bytes.Buffer
	fmt.Fprintf(&newPart, "\r\n--%s\r\n", w.boundary)
	fmt.Fprintf(&newPart, "Content-Type: %s\r\n", contentType)
	fmt.Fprintf(&newPart, "Content-Range: bytes %d-%d/%d\r\n", rangeStart, rangeEnd, contentLength)
	fmt.Fprintf(&newPart, "\r\n")

	_, err := io.Copy(w.w, &newPart)
	return err
}

func (w *Writer) Write(bytes []byte) (n int, err error) {
	if len(w.partClosed) == 0 {
		return 0, errors.New("call CreatePart before Write")
	} else if w.partClosed[len(w.partClosed)-1] {
		return 0, errors.New("last part is closed")
	}

	return w.w.Write(bytes)
}

func (w *Writer) Close() error {
	if len(w.partClosed) != 0 {
		w.partClosed[len(w.partClosed)-1] = true
	}

	_, err := fmt.Fprintf(w.w, "\r\n--%s--\r\n", w.boundary)
	return err
}
