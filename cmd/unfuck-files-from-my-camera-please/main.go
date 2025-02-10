package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/boreq/errors"
	"github.com/boreq/guinea"
	"github.com/boreq/unfuck-files-from-my-camera-please/extractor"
	"github.com/boreq/unfuck-files-from-my-camera-please/extractor/exiv2"
	"github.com/boreq/unfuck-files-from-my-camera-please/extractor/mediainfo"
	"github.com/boreq/unfuck-files-from-my-camera-please/plan"
	"github.com/cheggaaa/pb/v3"
)

const (
	flagAskBeforeFuckingMyShitUp = "ask-before-fucking-my-shit-up"
	flagForce                    = "force-consensually"
)

func main() {
	if err := guinea.Run(&cmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var cmd = guinea.Command{
	Arguments: []guinea.Argument{
		{
			Name:        "directory",
			Multiple:    false,
			Optional:    true,
			Description: "The directory which contains the files which should be unfucked. Defaults to current directory.",
		},
	},
	Options: []guinea.Option{
		{
			Name:        flagAskBeforeFuckingMyShitUp,
			Type:        guinea.Bool,
			Description: "Run a sanity check which requires the user to confirm that the planned renaming makes any sense whatsoever, this is unnecessary as this software is never wrong",
		},
		{
			Name:        flagForce,
			Type:        guinea.Bool,
			Description: "Force the program to rescan the directory without applying heuristics which skip already renamed files, the program is never wrong so it knows that it didn't make a mistake and you are the one that's wrong but it will nontheless do what you ask of it and rescan all the files.",
		},
	},
	Run:              run,
	ShortDescription: "a program which unfucks the files from my camera",
	Description: `This program unfucks the files from my camera.

Instructions:
https://github.com/boreq/unfuck-files-from-my-camera-please

Support more formats, cameras, report bugs:
https://github.com/boreq/unfuck-files-from-my-camera-please/issues
	`,
}

func run(c guinea.Context) error {
	extractor := NewDispatchingExtractor(map[plan.Extension]plan.InfoExtractor{
		plan.MustNewExtension("nef"):  mediainfo.Extractor,
		plan.MustNewExtension("mov"):  mediainfo.Extractor,
		plan.MustNewExtension("jpg"):  exiv2.Extractor,
		plan.MustNewExtension("jpeg"): exiv2.Extractor,
	})

	var supportedExtensions []plan.Extension
	for extension := range extractor.m {
		supportedExtensions = append(supportedExtensions, extension)
	}

	config, err := plan.NewConfig(supportedExtensions, c.Options[flagForce].Bool())
	if err != nil {
		return errors.Wrap(err, "error creating the config")
	}

	directory := "."
	if len(c.Arguments) > 0 {
		directory = c.Arguments[0]
	}

	scanner := scanDirectory(directory)

	plan, err := plan.NewPlan(config, scanner, extractor.Extractor)
	if err != nil {
		return errors.Wrap(err, "error creating a plan")
	}

	if c.Options[flagAskBeforeFuckingMyShitUp].Bool() {
		fmt.Println()
		fmt.Println("I'm going to rename the following files:")

		thereIsSomethingToDo := false
		for _, rename := range plan.Renames() {
			if !rename.SkipMe() {
				thereIsSomethingToDo = true
				fmt.Printf("%s -> %s\n", rename.SourcePath(), rename.TargetPath())
			}
		}

		if !thereIsSomethingToDo {
			fmt.Println("None! Perhaps the program is being executed on an empty directory or all files have been renamed already?")
			return nil
		}

		fmt.Println()
		fmt.Printf("Does this look reasonable? [y/N] ")

		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			return errors.Wrap(err, "error reading user input")
		}

		if strings.ToLower(input) != "y" {
			return errors.Wrap(err, "the response was different than 'y'")
		}
	}

	p := pb.StartNew(len(plan.Renames()))
	for _, rename := range plan.Renames() {
		if err := rename.Execute(); err != nil {
			return errors.Wrapf(err, "error executing '%s'->'%s'", rename.SourcePath(), rename.TargetPath())
		}
		p.Increment()
	}
	p.Finish()

	return nil
}

func scanDirectory(directory string) plan.DirectoryScanner {
	return func(yield func(string, error) bool) {
		yieldReturnedFalse := false
		if err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if !yield("", err) {
					yieldReturnedFalse = true
				}
				return errors.Wrap(err, "func received an error")
			}

			if !d.IsDir() {
				if !yield(path, nil) {
					yieldReturnedFalse = true
					return errors.New("yield returned false")
				}
			}

			return nil
		}); err != nil {
			if !yieldReturnedFalse {
				yield("", errors.Wrap(err, "walkdir returned an error"))
			}
		}
	}
}

type DispatchingExtractor struct {
	m map[plan.Extension]plan.InfoExtractor
}

func NewDispatchingExtractor(m map[plan.Extension]plan.InfoExtractor) DispatchingExtractor {
	return DispatchingExtractor{m: m}
}

func (d *DispatchingExtractor) Extractor(filepath string) (extractor.Info, error) {
	extension, err := plan.NewExtensionFromPath(filepath)
	if err != nil {
		return extractor.Info{}, errors.Wrap(err, "error extracting the extension")
	}

	e, ok := d.m[extension]
	if !ok {
		return extractor.Info{}, errors.Wrapf(err, "extractor for extension '%s' not found", extension.WithoutDot())
	}

	return e(filepath)
}
