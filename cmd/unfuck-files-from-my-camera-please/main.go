package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/boreq/errors"
	"github.com/boreq/guinea"
	"github.com/boreq/unfuck-names-of-files-from-my-camera-please/extractor/mediainfo"
	"github.com/boreq/unfuck-names-of-files-from-my-camera-please/plan"
	"github.com/cheggaaa/pb/v3"
)

func main() {
	if err := guinea.Run(&cmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Println("")
		fmt.Println()
		os.Exit(1)
	}
}

var cmd = guinea.Command{
	Arguments: []guinea.Argument{
		{
			Name:        "directory",
			Multiple:    false,
			Optional:    false,
			Description: "The directory which contains the files which should be unfucked.",
		},
	},
	Options: []guinea.Option{
		{
			Name:        "just-fuck-my-shit-up",
			Type:        guinea.Bool,
			Description: "Override the sanity check which requires the user to confirm that the planned renaming makes any sense whatsoever",
		},
	},
	Run:              run,
	ShortDescription: "a program which unfucks the files from my camera",
	Description: `This program unfucks the files from my camera.

https://github.com/boreq/unfuck-files-from-my-camera-please/issues
	`,
}

func run(c guinea.Context) error {
	config, err := plan.NewConfig([]plan.Extension{
		plan.MustNewExtension("nef"),
		plan.MustNewExtension("mov"),
	})
	if err != nil {
		return errors.Wrap(err, "error creating the config")
	}

	scanner := scanDirectory(c.Arguments[0])
	extractor := mediainfo.Extractor

	plan, err := plan.NewPlan(config, scanner, extractor)
	if err != nil {
		return errors.Wrap(err, "error creating a plan")
	}

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
		if err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				yield("", err)
				return errors.Wrap(err, "func received an error")
			}

			if !d.IsDir() {
				if !yield(path, nil) {
					return errors.New("yield returned false")
				}
			}

			return nil
		}); err != nil {
			yield("", errors.Wrap(err, "walkdir returned an error"))
		}
	}
}
