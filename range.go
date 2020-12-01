package multipart

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func RangeToParts(rangeValue string, respContentType, respFileSize string) ([]*Part, error) {
	if rangeValue == "" {
		return nil, nil // header not present
	}

	const unit = "bytes="
	if !strings.HasPrefix(rangeValue, unit) {
		return nil, errors.New("byte= not found")
	}

	var parts []*Part
	for _, ra := range strings.Split(rangeValue[len(unit):], ",") {
		ra = strings.TrimSpace(ra)
		if ra == "" {
			continue
		}

		i := strings.Index(ra, "-")
		if i < 0 {
			return nil, errors.New("- not found")
		}

		start, end := strings.TrimSpace(ra[:i]), strings.TrimSpace(ra[i+1:])
		parts = append(parts, NewPart(respContentType, start, end, respFileSize))
	}

	if err := checkParts(parts); err != nil {
		return nil, err
	}
	return parts, nil
}

func checkParts(parts []*Part) error {
	var err error
	for _, part := range parts {
		if part.fileSize != "*" {
			part.fileSizeInt, err = strconv.ParseInt(part.fileSize, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid file size %w", err)
			} else if part.fileSizeInt <= 0 {
				return errors.New("invalid file size")
			}
		} else {
			part.fileSizeInt = -1
		}

		if part.rangeEnd != "" {
			part.rangeEndInt, err = strconv.ParseInt(part.rangeEnd, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid range end %w", err)
			} else if part.rangeEndInt < 0 ||
				(part.fileSize != "*" && part.rangeEndInt >= part.fileSizeInt) {
				return errors.New("invalid range end")
			}

			if part.rangeStart == "" {
				// If no start is specified, end specifies the
				// range start relative to the end of the file.
				if part.fileSize == "*" {
					return errors.New("file size is unknown")
				}
				part.rangeStartInt = part.fileSizeInt - part.rangeEndInt
				part.rangeEndInt = part.fileSizeInt - 1
				continue
			}
		} else if part.fileSize == "*" {
			return errors.New("range end equals to file size while file size is unknown")
		} else if part.rangeStart == "" {
			return errors.New("both start and end are empty")
		} else {
			part.rangeEndInt = part.fileSizeInt - 1
		}

		if part.rangeStart != "" {
			part.rangeStartInt, err = strconv.ParseInt(part.rangeStart, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid range start %w", err)
			} else if part.rangeStartInt < 0 ||
				(part.fileSize != "*" && part.rangeStartInt >= part.fileSizeInt) ||
				(part.rangeEnd != "" && part.rangeStartInt > part.rangeEndInt) {
				return errors.New("invalid range start")
			}
		} else {
			part.rangeStartInt = 0
		}
	}

	return nil
}
