package mediainfo_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/boreq/unfuck-files-from-my-camera-please/extractor"
	"github.com/boreq/unfuck-files-from-my-camera-please/extractor/mediainfo"
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
			Name: "nikon d5300 nef",
			Output: `
{
  "creatingLibrary": {
    "name": "MediaInfoLib",
    "version": "24.06",
    "url": "https://mediaarea.net/MediaInfo"
  },
  "media": {
    "@ref": "/server/camera/terrarium/1/DSC_0002.NEF",
    "track": [
      {
        "@type": "General",
        "ImageCount": "1",
        "FileExtension": "NEF",
        "Format": "TIFF",
        "FileSize": "28083699",
        "File_Modified_Date": "2024-11-14 09:45:08 UTC",
        "File_Modified_Date_Local": "2024-11-14 10:45:08",
        "Encoded_Application_CompanyName": "NIKON CORPORATION",
        "Encoded_Application_Name": "Ver.1.01 ",
        "Encoded_Library_Name": "NIKON D5300",
        "extra": {
          "FileExtension_Invalid": "tiff tif"
        }
      },
      {
        "@type": "Image",
        "Format": "Raw",
        "Format_Settings_Endianness": "Little",
        "Width": "160",
        "Height": "120",
        "ColorSpace": "RGB",
        "BitDepth": "8",
        "Compression_Mode": "Lossless",
        "Encoded_Date": "2024:11:14 10:45:07",
        "extra": {
          "Density_X": "300",
          "Density_Y": "300",
          "Density_Unit": "dpi",
          "Density_String": "300 dpi"
        }
      }
    ]
  }
}
`,
			ExpectedInfo: extractor.MustNewInfo(time.Date(2024, 11, 14, 10, 45, 7, 0, time.Local)),
		},
		{
			Name: "nikon d5300 mov",
			Output: `
{
  "creatingLibrary": {
    "name": "MediaInfoLib",
    "version": "24.06",
    "url": "https://mediaarea.net/MediaInfo"
  },
  "media": {
    "@ref": "/server/camera/terrarium/1/DSC_0001.MOV",
    "track": [
      {
        "@type": "General",
        "VideoCount": "1",
        "AudioCount": "1",
        "FileExtension": "MOV",
        "Format": "MPEG-4",
        "Format_Profile": "QuickTime",
        "CodecID": "qt  ",
        "CodecID_Version": "2007.09",
        "CodecID_Compatible": "qt  /niko",
        "FileSize": "211481487",
        "Duration": "44.820",
        "OverallBitRate": "37747700",
        "FrameRate": "50.000",
        "FrameCount": "2241",
        "StreamSize": "289506",
        "HeaderSize": "24",
        "DataSize": "211191989",
        "FooterSize": "289474",
        "IsStreamable": "No",
        "Encoded_Date": "2024-11-14 10:44:55 UTC",
        "Tagged_Date": "2024-11-14 10:44:55 UTC",
        "File_Modified_Date": "2024-11-14 09:44:08 UTC",
        "File_Modified_Date_Local": "2024-11-14 10:44:08"
      },
      {
        "@type": "Video",
        "StreamOrder": "0",
        "ID": "1",
        "Format": "AVC",
        "Format_Profile": "High",
        "Format_Level": "4.2",
        "Format_Settings_CABAC": "Yes",
        "Format_Settings_RefFrames": "2",
        "Format_Settings_GOP": "M=3, N=24",
        "CodecID": "avc1",
        "Duration": "44.820",
        "BitRate": "36166194",
        "Width": "1920",
        "Height": "1080",
        "Stored_Height": "1088",
        "Sampled_Width": "1920",
        "Sampled_Height": "1080",
        "PixelAspectRatio": "1.000",
        "DisplayAspectRatio": "1.778",
        "Rotation": "0.000",
        "FrameRate_Mode": "CFR",
        "FrameRate": "50.000",
        "FrameRate_Num": "50",
        "FrameRate_Den": "1",
        "FrameCount": "2241",
        "ColorSpace": "YUV",
        "ChromaSubsampling": "4:2:0",
        "BitDepth": "8",
        "ScanType": "Progressive",
        "StreamSize": "202621101",
        "Language": "en",
        "Encoded_Date": "2024-11-14 10:44:55 UTC",
        "Tagged_Date": "2024-11-14 10:44:55 UTC",
        "colour_description_present": "Yes",
        "colour_description_present_Source": "Container / Stream",
        "colour_range": "Full",
        "colour_range_Source": "Stream",
        "colour_primaries": "BT.709",
        "colour_primaries_Source": "Container / Stream",
        "transfer_characteristics_Source": "Container",
        "transfer_characteristics_Original": "BT.470 System M",
        "transfer_characteristics_Original_Source": "Stream",
        "matrix_coefficients": "BT.601",
        "matrix_coefficients_Source": "Container / Stream",
        "extra": {
          "CodecConfigurationBox": "avcC"
        }
      },
      {
        "@type": "Audio",
        "StreamOrder": "1",
        "ID": "2",
        "Format": "PCM",
        "Format_Settings_Endianness": "Little",
        "Format_Settings_Sign": "Signed",
        "CodecID": "sowt",
        "Duration": "44.640",
        "BitRate_Mode": "CBR",
        "BitRate": "1536000",
        "Channels": "2",
        "ChannelPositions": "Front: L R",
        "ChannelLayout": "L R",
        "SamplingRate": "48000",
        "SamplingCount": "2142720",
        "BitDepth": "16",
        "StreamSize": "8570880",
        "Language": "en",
        "Encoded_Date": "2024-11-14 10:44:55 UTC",
        "Tagged_Date": "2024-11-14 10:44:55 UTC"
      }
    ]
  }
}
			`,
			ExpectedInfo: extractor.MustNewInfo(time.Date(2024, 11, 14, 10, 44, 55, 0, time.UTC)),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := strings.NewReader(testCase.Output)

			i, err := mediainfo.FindInfo(r)
			require.NoError(t, err)
			assert.Equal(t, i.Timestmap().Location().String(), testCase.ExpectedInfo.Timestmap().Location().String())
			if !assert.True(t, i.Timestmap().Equal(testCase.ExpectedInfo.Timestmap())) {
				fmt.Printf("\n%s != %s\n", i.Timestmap().String(), testCase.ExpectedInfo.Timestmap().String())
			}
		})

	}
}
