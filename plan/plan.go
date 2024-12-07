package plan

import (
	"context"
	"fmt"
	"iter"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/boreq/errors"
	"github.com/boreq/unfuck-names-of-files-from-my-camera-please/extractor"
	"github.com/cheggaaa/pb/v3"
)

const workersPerCPU = 4

type DirectoryScanner iter.Seq2[string, error]
type InfoExtractor func(filepath string) (extractor.Info, error)

type Plan struct {
	renames []*Rename
}

func NewPlan(config Config, scanner DirectoryScanner, extractor InfoExtractor) (*Plan, error) {
	plan := &Plan{}

	var paths []string

	for path, err := range scanner {
		if err != nil {
			return nil, errors.Wrap(err, "directory scanner returned an error")
		}

		ext, err := NewExtensionFromPath(path)
		if err != nil {
			return nil, errors.Wrapf(err, "error creating an extension from path '%s'", path)
		}

		if !config.IsInExtensionsToProcess(ext) {
			continue
		}

		paths = append(paths, path)
	}

	fmt.Println("Determining how to rename files...")
	p := pb.StartNew(len(paths))
	defer p.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chTasks := make(chan orderedTask)
	chResults := make(chan orderedResult)

	for range workersPerCPU * runtime.NumCPU() {
		go func() {
			for {
				select {
				case task := <-chTasks:
					rename, err := NewRename(task.path, extractor)
					select {
					case chResults <- orderedResult{task: task, rename: rename, err: err}:
					case <-ctx.Done():
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		for i, path := range paths {
			select {
			case chTasks <- orderedTask{index: i, path: path}:
			case <-ctx.Done():
				return
			}
		}
	}()

	var results []orderedResult
	for range len(paths) {
		result := <-chResults
		if err := result.err; err != nil {
			return nil, errors.Wrapf(err, "error processing '%s'", result.task.path)
		}

		p.Increment()
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].task.index < results[j].task.index
	})

	for _, result := range results {
		if err := plan.addRename(result.rename); err != nil {
			return nil, errors.Wrapf(err, "error adding a rename for path '%s'", result.task.path)
		}
	}

	return plan, nil
}

type orderedTask struct {
	index int
	path  string
}

type orderedResult struct {
	task   orderedTask
	rename Rename
	err    error
}

func (p *Plan) Renames() []*Rename {
	return p.renames
}

func (p *Plan) addRename(newRename Rename) error {
	existingRename, ok := p.findConflictingRenameWithHighestSuffix(&newRename)
	if ok {
		if err := existingRename.deconflict(&newRename); err != nil {
			return errors.Wrap(err, "error deconflicting")
		}
	}

	p.renames = append(p.renames, &newRename)
	return nil
}

func (p *Plan) findConflictingRenameWithHighestSuffix(newRename *Rename) (*Rename, bool) {
	var found *Rename
	for _, existingRename := range p.renames {
		if existingRename.conflicts(newRename) {
			if found == nil || existingRename.suffix > found.suffix {
				found = existingRename
			}
		}
	}
	return found, found != nil
}

type Rename struct {
	sourcePath string
	targetPath string

	timestampString string
	directory       string
	info            extractor.Info
	suffix          int // 0 means no suffix
	extension       Extension
}

func NewRename(sourcePath string, extractor InfoExtractor) (Rename, error) {
	info, err := extractor(sourcePath)
	if err != nil {
		return Rename{}, errors.Wrapf(err, "extractor returned an error for '%s'", sourcePath)
	}

	extension, err := NewExtensionFromPath(sourcePath)
	if err != nil {
		return Rename{}, errors.Wrapf(err, "error creating an extension from path '%s'", sourcePath)
	}

	timestampString, targetPath := createTargetPath(sourcePath, info, extension, 0)
	directory, _ := filepath.Split(sourcePath)

	return Rename{
		sourcePath: sourcePath,
		targetPath: targetPath,

		directory:       directory,
		timestampString: timestampString,
		info:            info,
		suffix:          0,
		extension:       extension,
	}, nil
}

func (r Rename) conflicts(newRename *Rename) bool {
	return r.directory == newRename.directory && r.timestampString == newRename.timestampString && r.extension == newRename.extension
}

func (r *Rename) deconflict(newRename *Rename) error {
	if !r.conflicts(newRename) {
		return errors.New("those renames do not have a conflict")
	}

	if r.suffix == 0 {
		r.changeSuffix(1)
		newRename.changeSuffix(2)
	} else {
		newRename.changeSuffix(r.suffix + 1)
	}

	return nil

}

func (r Rename) SkipMe() bool {
	return r.targetPath == r.sourcePath
}

func (r Rename) Execute() error {
	if r.SkipMe() {
		return nil
	}

	_, err := os.Stat(r.targetPath)
	if err == nil {
		return errors.New("aborting rename: target file already exists, are the files being modified by other programs? if not this is a bug")
	} else {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "error checking if target file exits")
		}
	}

	if err := os.Rename(r.sourcePath, r.targetPath); err != nil {
		return errors.Wrap(err, "error calling os rename")
	}

	return nil
}

func (r Rename) SourcePath() string {
	return r.sourcePath
}

func (r Rename) TargetPath() string {
	return r.targetPath
}

func (r *Rename) changeSuffix(newSuffix int) error {
	if r.suffix >= newSuffix {
		return errors.New("why are we rolling back or keeping the suffix, something went very wrong")
	}

	timestampString, targetPath := createTargetPath(r.sourcePath, r.info, r.extension, newSuffix)
	if timestampString != r.timestampString {
		return errors.New("new timestamp string is different than the old one, something went very wrong")
	}

	r.targetPath = targetPath
	r.suffix = newSuffix
	return nil
}

func createTargetPath(sourcePath string, inf extractor.Info, ext Extension, suffix int) (string, string) {
	directory, _ := filepath.Split(sourcePath)

	timestampString := inf.Timestmap().UTC().Format("2006-01-02 15:04:05")
	if suffix != 0 {
		targetFilename := fmt.Sprintf("%s %02d.%s", timestampString, suffix, ext.WithoutDot())
		targetPath := path.Join(directory, targetFilename)
		return timestampString, targetPath
	}

	targetFilename := fmt.Sprintf("%s.%s", timestampString, ext.WithoutDot())
	targetPath := path.Join(directory, targetFilename)
	return timestampString, targetPath
}

type Extension struct {
	extensionWithoutDot string
}

func NewExtensionFromPath(path string) (Extension, error) {
	ext := filepath.Ext(path)
	ext = strings.TrimPrefix(ext, ".")
	return NewExtension(ext)
}

func NewExtension(extensionWithoutDot string) (Extension, error) {
	if extensionWithoutDot == "" {
		return Extension{}, errors.New("extension can't be empty")
	}

	extensionWithoutDot = strings.ToLower(extensionWithoutDot)

	if strings.HasPrefix(extensionWithoutDot, ".") {
		return Extension{}, errors.New("don't include the dot in the extension name e.g. use EXT instead of .EXT, let's keep things consistent instead of making a mess")
	}

	return Extension{
		extensionWithoutDot: extensionWithoutDot,
	}, nil
}

func MustNewExtension(extensionWithoutDot string) Extension {
	v, err := NewExtension(extensionWithoutDot)
	if err != nil {
		panic(err)
	}
	return v
}

func (s *Extension) WithoutDot() string {
	return s.extensionWithoutDot
}

type Config struct {
	extensionsToProcess []Extension
}

func NewConfig(extensionsToProcess []Extension) (Config, error) {
	if len(extensionsToProcess) == 0 {
		return Config{}, errors.New("list of extensions to process is empty, I'd say running this program with this config makes no sense")
	}

	return Config{
		extensionsToProcess: extensionsToProcess,
	}, nil
}

func (c Config) IsInExtensionsToProcess(extension Extension) bool {
	for _, extensionInConfig := range c.extensionsToProcess {
		if extensionInConfig == extension {
			return true
		}
	}
	return false
}
