package plan_test

import (
	"testing"

	"github.com/boreq/unfuck-files-from-my-camera-please/plan"
)

func Test_ItsVeryLikelyThisFileWasAlreadyRenamedByThisSoftware(t *testing.T) {
	testCases := []struct {
		Filename string
		Result   bool
	}{
		{
			Filename: "2025-02-07 12:16:10.mov",
			Result:   true,
		},
		{
			Filename: "2025-02-06 22:39:34 01.jpg",
			Result:   true,
		},
		{
			Filename: "DSC_0022.JPG",
			Result:   false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Filename, func(t *testing.T) {
			if plan.ItsVeryLikelyThisFileWasAlreadyRenamedByThisSoftware(testCase.Filename) != testCase.Result {
				t.Fail()
			}
		})
	}
}
