package file

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-github/github"
)

// ByteSize wraps sizes
type ByteSize float64

// ByteSizes
const (
	_           = iota
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

// LocalFile wraps a filename
type LocalFile struct {
	FilePath string
	dryRun   bool

	statusChan *chan error
}

// New creates a new local file
func New(file string, dryRun bool, statusChan *chan error) *LocalFile {
	return &LocalFile{
		FilePath:   file,
		dryRun:     dryRun,
		statusChan: statusChan,
	}
}

// Exists checks whether a file exists on the filesystem
func (f LocalFile) Exists() (bool, error) {
	_, err := os.Stat(f.FilePath)

	if err != nil {
		return false, err
	}

	return true, nil
}

// Upload upload file to GitHub gist servers, unless doing a dryrun. If dryrun
// is specified files are not uploaded but rather logged as if they were to be
func (f LocalFile) Upload(ctx context.Context, uploadClient *github.Client) (bool, error) {
	if f.dryRun {
		log.Printf("dryrun: Uploading -> %s...", f.FilePath)
		return true, nil
	}

	contents, err := f.Contents()
	if err != nil {
		log.Fatal(err)
	}

	filename, err := f.Name()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Uploading %s...", filename)

	gistFile := github.GistFile{Content: &contents, Filename: &filename}

	filesMap := make(map[github.GistFilename]github.GistFile)

	// Convert to GistFilename
	gistFilename := github.GistFilename(filename)
	filesMap[gistFilename] = gistFile
	gistUpload := github.Gist{Files: filesMap}

	gist, _, err := uploadClient.Gists.Create(ctx, &gistUpload)

	if err != nil {
		*f.statusChan <- err
		log.Fatal(err)
	}

	log.Printf("Uploaded: %s (URL: %s) \n", filename, gist.GetHTMLURL())

	*f.statusChan <- nil

	return true, nil
}

// Contents read a content from file
func (f LocalFile) Contents() (string, error) {
	contents, err := ioutil.ReadFile(f.FilePath)
	if err != nil {
		log.Fatal(err)
	}

	return string(contents), nil
}

// Name get the basename of the file to upload
func (f LocalFile) Name() (string, error) {
	stat, err := os.Stat(f.FilePath)
	if err != nil {
		log.Fatal(err)
	}

	return stat.Name(), nil
}

// Size return filesize
func (f LocalFile) Size() ByteSize {
	stat, err := os.Stat(f.FilePath)
	if err != nil {
		log.Fatal(err)
	}

	return ByteSize(stat.Size())
}
