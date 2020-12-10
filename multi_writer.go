package multipart

import (
	"bytes"
	"fmt"
	"io"
)

type Writer struct {
	src      io.ReadSeeker
	pr       *io.PipeReader
	pw       *io.PipeWriter
	boundary string
	parts    []*Part
}

func NewWriter(src io.ReadSeeker, parts []*Part) *Writer {
	boundary := randomBoundary()
	return NewWriterWithBoundary(src, parts, boundary)
}

func NewWriterWithBoundary(src io.ReadSeeker, parts []*Part, boundary string) *Writer {
	pr, pw := io.Pipe()
	return &Writer{
		src:      src,
		pr:       pr,
		pw:       pw,
		boundary: boundary,
		parts:    parts,
	}
}

func (w *Writer) ContentLength() int64 {
	var buf bytes.Buffer
	partsBodyLen := int64(0)

	for _, part := range w.parts {
		partsBodyLen += part.rangeEndInt - part.rangeStartInt + 1
		w.WritePartHeader(&buf, part)
	}
	fmt.Fprintf(&buf, "\r\n--%s--", w.boundary)
	partsHeaderLen := int64(buf.Len())

	// the first CRLF is not part of message body
	// ref: https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html
	return partsHeaderLen + partsBodyLen - 2
}

func (w *Writer) SetBoundary(boundary string) {
	w.boundary = boundary
}

func (w *Writer) WritePartHeader(buf io.Writer, part *Part) error {
	fmt.Fprintf(buf, "\r\n--%s\r\n", w.boundary)
	fmt.Fprint(buf, "Content-Type: application/octet-stream\r\n")
	fmt.Fprintf(buf, "Content-Range: bytes %s-%s/%s\r\n", part.rangeStart, part.rangeEnd, part.fileSize)
	fmt.Fprint(buf, "\r\n")

	return nil
}

func (w *Writer) WriteMultiParts() error {
	var err error
	for _, part := range w.parts {
		if err = w.WritePartHeader(w.pw, part); err != nil {
			return err
		}
		if err = writePartBody(w.src, w.pw, part); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) Write(bytes []byte) (n int, err error) {
	return w.pw.Write(bytes)
}

func (w *Writer) Close() error {
	var err error
	_, err = fmt.Fprintf(w.pw, "\r\n--%s--", w.boundary)
	if err != nil {
		return err
	}
	return w.pw.Close()
}

func (w *Writer) CloseWithError(err error) error {
	return w.pw.CloseWithError(err)
}

func (w *Writer) Read(p []byte) (n int, err error) {
	return w.pr.Read(p)
}
