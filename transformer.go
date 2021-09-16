package multipart

import (
	"bytes"
	"fmt"
	"io"
)

// TODO: this is introduced in Go 1.16
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

type Transformer struct {
	src      ReadSeekCloser
	boundary string
	parts    []*Part
}

func NewTransformer(src ReadSeekCloser, parts []*Part) *Transformer {
	boundary := randomBoundary()
	return NewTransformerWithBoundary(src, parts, boundary)
}

func NewTransformerWithBoundary(src ReadSeekCloser, parts []*Part, boundary string) *Transformer {
	return &Transformer{
		src:      src,
		boundary: boundary,
		parts:    parts,
	}
}

func (tfm *Transformer) SetBoundary(boundary string) {
	tfm.boundary = boundary
}

func (tfm *Transformer) ContentLength() int64 {
	var buf bytes.Buffer
	partsBodyLen := int64(0)

	for _, part := range tfm.parts {
		partsBodyLen += part.rangeEndInt - part.rangeStartInt + 1
		tfm.WritePartHeader(&buf, part)
	}
	fmt.Fprintf(&buf, "\r\n--%s--", tfm.boundary)
	partsHeaderLen := int64(buf.Len())

	// the first CRLF is not part of message body
	// ref: https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html
	return partsHeaderLen + partsBodyLen - 2
}

func (tfm *Transformer) WritePartHeader(buf io.Writer, part *Part) error {
	fmt.Fprintf(buf, "\r\n--%s\r\n", tfm.boundary)
	fmt.Fprint(buf, "Content-Type: application/octet-stream\r\n")
	fmt.Fprintf(buf, "Content-Range: bytes %s-%s/%s\r\n", part.rangeStart, part.rangeEnd, part.fileSize)
	fmt.Fprint(buf, "\r\n")

	return nil
}

func (tfm *Transformer) WriteMultiParts(wt io.Writer) error {
	var err error
	for _, part := range tfm.parts {
		if err = tfm.WritePartHeader(wt, part); err != nil {
			return err
		}
		if err = writePartBody(tfm.src, wt, part); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(wt, "\r\n--%s--", tfm.boundary)
	if err != nil {
		return err
	}

	return nil
}
