package multipart

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"sort"
)

// func writeErrResponse(dst IPipeWriter, statusCode int, respErr error) {
// 	var err error
// 	switch statusCode {
// 	case 400:
// 		_, err = dst.Write([]byte(fmt.Sprintf("HTTP/1.1 400 Bad Request\r\n%s", respErr)))
// 	case 500:
// 		_, err = dst.Write([]byte(fmt.Sprintf("HTTP/1.1 500 Internal Server Error\r\n%s", respErr)))
// 	}

// 	if err != nil {
// 		if err == io.ErrClosedPipe {
// 			panic(ErrClosed(err))
// 		} else {
// 			dst.CloseWithError(err)
// 		}
// 	}
// }

func writeStatus(dst *bytes.Buffer, statusCode int) error {
	var err error
	switch statusCode {
	case 200:
		_, err = dst.Write([]byte("HTTP/1.1 200 OK\r\n"))
	case 206:
		_, err = dst.Write([]byte("HTTP/1.1 206 Partial Content\r\n"))
	}
	return err
}

func writeHeaders(buf *bytes.Buffer, headers textproto.MIMEHeader) error {
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var err error
	for _, k := range keys {
		for _, v := range headers[k] {
			_, err = fmt.Fprintf(buf, "%s: %s\r\n", k, v)
			if err != nil {
				return err
			}
		}
	}
	// _, err = fmt.Fprint(buf, "\r\n")
	return nil
}

func writePartBody(src io.ReadSeeker, dst io.Writer, part *Part) error {
	_, err := src.Seek(part.rangeStartInt, os.SEEK_SET)
	if err != nil {
		return err
	}

	rangeLen := part.rangeEndInt - part.rangeStartInt + 1
	wrote, err := io.CopyN(dst, src, rangeLen)
	if err != nil {
		return err
	} else if wrote != rangeLen {
		return errors.New("request range length is larger than file size")
	}
	return nil
}

func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}
