package mediainfo

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/boreq/errors"
	"github.com/boreq/unfuck-names-of-files-from-my-camera-please/extractor"
)

const (
	typeGeneral = "General"
)

func Extractor(filepath string) (extractor.Info, error) {
	cmd := exec.Command("mediainfo", "--Output=JSON", filepath)

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

	var output mediainfoOutput
	if err := json.Unmarshal(b, &output); err != nil {
		return extractor.Info{}, errors.Wrap(err, "error decoding json")
	}

	// try to find general
	for _, track := range output.Media.Track {
		if track.Type == typeGeneral {
			if t, err := parseTimestamp(track.EncodedDate); err == nil {
				return extractor.NewInfo(t)
			}
		}
	}

	// try to find anything
	for _, track := range output.Media.Track {
		if t, err := parseTimestamp(track.EncodedDate); err == nil {
			return extractor.NewInfo(t)
		}
	}

	return extractor.Info{}, fmt.Errorf("unable to find the timestamp in this mediainfo output: '%s'", string(b))
}

func parseTimestamp(s string) (time.Time, error) {
	// nef
	if t, err := time.ParseInLocation("2006:01:02 15:04:05", s, time.Local); err == nil {
		return t, nil
	}

	// mov
	if t, err := time.Parse("2006-01-02 15:04:05 MST", s); err == nil {
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
