package multipart

import (
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"os"
)

type Part struct {
	contentType   string
	rangeStart    string // it could be emtpy
	rangeEnd      string // it could be emtpy
	fileSize      string // it could be *
	rangeStartInt int64  // set as -1 if it is empty
	rangeEndInt   int64  // set as -1 if it is empty
	fileSizeInt   int64  // set as -1 if it is *
}

func NewPart(contentType, rangeStart, rangeEnd, fileSize string) *Part {
	return &Part{
		contentType: contentType,
		rangeStart:  rangeStart,
		rangeEnd:    rangeEnd,
		fileSize:    fileSize,
	}
}

func WriteResponse(src io.ReadSeeker, dst io.WriteCloser, fileName string, parts []*Part) error {
	return WriteResponseWithBoundary(src, dst, fileName, parts, randomBoundary())
}

func WriteResponseWithBoundary(src io.ReadSeeker, dst io.WriteCloser, fileName string, parts []*Part, boundary string) error {
	if len(parts) == 0 {
		return errors.New("no part to write")
	} else if len(parts) == 1 {
		return writeSinglePart(src, dst, fileName, parts[0])
	}
	mw, err := NewWriterWithBoundary(dst, boundary)
	if err != nil {
		return err
	}
	return writeMultiPart(src, mw, fileName, parts)

}

// TODO: support reverse seek
// TODO: Add header: Content-Length
func writeSinglePart(src io.ReadSeeker, writer io.WriteCloser, fileName string, part *Part) error {
	headers := textproto.MIMEHeader{}
	// headers.Add("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	headers.Add("Content-Range", fmt.Sprintf("bytes %s-%s/%s", part.rangeStart, part.rangeEnd, part.fileSize))
	headers.Add("Content-Type", "application/octet-stream")

	err := writeHeaders(writer, headers)
	if err != nil {
		return err
	}

	if err = writePartBody(src, writer, part); err != nil {
		return err
	}

	return writer.Close()
}

func writeMultiPart(src io.ReadSeeker, mw *Writer, fileName string, parts []*Part) error {
	var err error
	for _, part := range parts {
		err = mw.CreatePart(part.contentType, part.rangeStart, part.rangeEnd, part.fileSize)
		if err != nil {
			return err
		}

		if err = writePartBody(src, mw, part); err != nil {
			return err
		}
	}

	return mw.Close()
}

func writePartBody(src io.ReadSeeker, dst io.Writer, part *Part) error {
	var err error

	_, err = src.Seek(part.rangeStartInt, os.SEEK_SET)
	if err != nil {
		return err
	}

	rangeLen := part.rangeEndInt - part.rangeStartInt
	wrote, err := io.CopyN(dst, src, rangeLen)
	if err != nil {
		return err
	} else if wrote != rangeLen {
		return errors.New("request range length is larger than file size")
	}
	return nil
}
