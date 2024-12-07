package exiv2_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/boreq/unfuck-files-from-my-camera-please/extractor"
	"github.com/boreq/unfuck-files-from-my-camera-please/extractor/exiv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindInfo(t *testing.T) {
	testCases := []struct {
		Name         string
		Output       string
		ExpectedInfo extractor.Info
	}{
		{
			Name: "nikon d5300 jpg",
			Output: `
File name       : ./testing/DSC_0029.JPG
File size       : 9976989 Bytes
MIME type       : image/jpeg
Image size      : 6000 x 4000
Thumbnail       : image/jpeg, 10455 Bytes
Camera make     : NIKON CORPORATION
Camera model    : NIKON D5300
Image timestamp : 2024:12:06 15:21:35
File number     :
Exposure time   : 1/50 s
Aperture        : F3.5
Exposure bias   : -1 EV
Flash           : No flash
Flash bias      :
Focal length    : 18.0 mm
Subject distance: 3.76 m
ISO speed       : 6400
Exposure mode   : Auto
Metering mode   : Multi-segment
Macro mode      :
Image quality   : FINE
White balance   : AUTO
Copyright       :
Exif comment    : charset=Ascii
`,
			ExpectedInfo: extractor.MustNewInfo(time.Date(2024, 12, 06, 15, 21, 35, 0, time.Local)),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := strings.NewReader(testCase.Output)

			i, err := exiv2.FindInfo(r)
			require.NoError(t, err)
			if !assert.True(t, i.Timestmap().Equal(testCase.ExpectedInfo.Timestmap())) {
				fmt.Printf("\n%s != %s\n", i.Timestmap().String(), testCase.ExpectedInfo.Timestmap().String())
			}
		})

	}
}
