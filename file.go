package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-github/github"
)

// LocalFile wraps a filename
type LocalFile struct {
	filepath string
}

// Exists checks whether a file exists on the filesystem
func (f LocalFile) Exists() (bool, error) {
	_, err := os.Stat(f.filepath)

	if err != nil {
		return false, err
	}

	return true, nil
}

// Upload upload file to GitHub gist servers, unless doing a dryrun. If dryrun
// is specified files are not uploaded but rather logged as if they were to be
func (f LocalFile) Upload(ctx context.Context, uploadClient *github.Client) (bool, error) {
	if dryRun {
		log.Printf("dryrun: Uploading -> %s...", f.filepath)
		return true, nil
	}

	contents, err := f.GetFileContents()
	if err != nil {
		log.Fatal(err)
	}

	filename, err := f.GetFilename()
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
		log.Fatal(err)
	}

	filesUploaded <- 1

	log.Printf("Uploaded: %s (URL: %s) \n", filename, gist.GetHTMLURL())

	return true, nil
}

// GetFileContents read a content from file
func (f LocalFile) GetFileContents() (string, error) {
	contents, err := ioutil.ReadFile(f.filepath)
	if err != nil {
		log.Fatal(err)
	}

	return string(contents), nil
}

// GetFilename get the basename of the file to upload
func (f LocalFile) GetFilename() (string, error) {
	stat, err := os.Stat(f.filepath)
	if err != nil {
		log.Fatal(err)
	}

	return stat.Name(), nil
}

// GetFilesize return filesize
func (f LocalFile) GetFilesize() ByteSize {
	stat, err := os.Stat(f.filepath)
	if err != nil {
		log.Fatal(err)
	}

	return ByteSize(stat.Size())
}

// NewLocalFile
func NewLocalFile(file string) *LocalFile {
	return &LocalFile{filepath: file}
}
