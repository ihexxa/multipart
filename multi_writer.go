package multipart

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/textproto"
)

type Writer struct {
	w          io.Writer
	boundary   string
	partClosed []bool
}

func NewWriter(w io.Writer) (*Writer, error) {
	boundary := randomBoundary()
	return NewWriterWithBoundary(w, boundary)
}

func NewWriterWithBoundary(w io.Writer, boundary string) (*Writer, error) {
	headers := textproto.MIMEHeader{}
	headers.Add("Content-Type", fmt.Sprintf("multipart/byteranges; boundary=%s", boundary))
	err := writeHeaders(w, headers)
	if err != nil {
		return nil, err
	}

	return &Writer{
		w:          w,
		boundary:   boundary,
		partClosed: []bool{},
	}, nil
}

func (w *Writer) SetBoundary(boundary string) {
	w.boundary = boundary
}

func (w *Writer) CreatePart(contentType, rangeStart, rangeEnd, fileSize string) error {
	if len(w.partClosed) != 0 {
		w.partClosed[len(w.partClosed)-1] = true
	}

	w.partClosed = append(w.partClosed, false)

	var newPart bytes.Buffer
	fmt.Fprintf(&newPart, "\r\n--%s\r\n", w.boundary)
	fmt.Fprintf(&newPart, "Content-Type: %s\r\n", contentType)
	fmt.Fprintf(&newPart, "Content-Range: bytes %s-%s/%s\r\n", rangeStart, rangeEnd, fileSize)
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
