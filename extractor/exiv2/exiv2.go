package exiv2

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/boreq/errors"
	"github.com/boreq/unfuck-names-of-files-from-my-camera-please/extractor"
)

const (
	imageTimestamp = "Image timestamp"
)

func Extractor(filepath string) (extractor.Info, error) {
	cmd := exec.Command("exiv2", "--Output=JSON", filepath)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return extractor.Info{}, errors.Wrapf(err, "could not get stdout pipe for '%s'", cmd.String())
	}

	if err := cmd.Start(); err != nil {
		return extractor.Info{}, errors.Wrapf(err, "error starting `%s`", cmd.String())
	}

	foundInfo, foundInfoErr := FindInfo(stdout)

	if err := cmd.Wait(); err != nil {
		return extractor.Info{}, errors.Wrapf(err, "error waiting for `%s`", cmd.String())
	}

	return foundInfo, foundInfoErr
}

func FindInfo(jsonOutput io.Reader) (extractor.Info, error) {
	// let's read all of this because we may need to put it in the error message
	b, err := io.ReadAll(jsonOutput)
	if err != nil {
		return extractor.Info{}, errors.Wrap(err, "error reading the data")
	}

	scanner := bufio.NewScanner(strings.NewReader(string(b)))
	for scanner.Scan() {
		split := strings.SplitN(scanner.Text(), ":", 2)
		if len(split) != 2 {
			continue
		}

		key := strings.TrimSpace(split[0])
		value := strings.TrimSpace(split[1])

		if key == imageTimestamp {
			if t, err := parseTimestamp(value); err == nil {
				return extractor.NewInfo(t)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return extractor.Info{}, errors.Wrap(err, "scnaner error")
	}

	return extractor.Info{}, fmt.Errorf("unable to find the timestamp in this exiv2 output: '%s'", string(b))
}

func parseTimestamp(s string) (time.Time, error) {
	// jpg
	if t, err := time.ParseInLocation("2006:01:02 15:04:05", s, time.Local); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("error parsing the timestamp '%s'", s)
}

type mediainfoOutput struct {
	Media mediaInfoOutputMedia `json:"media"`
}

type mediaInfoOutputMedia struct {
	Track []mediaInfoOutputMediaTrack `json:"track"`
}

type mediaInfoOutputMediaTrack struct {
	Type        string `json:"@type"`
	EncodedDate string `json:"Encoded_Date"`
}
