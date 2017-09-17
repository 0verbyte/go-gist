package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
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

var (
	upload      bool
	dryRun      bool
	files       string
	githubToken string

	// Channel buffer for finished file uploads
	filesUploaded chan int
)

func init() {
	flag.BoolVar(&upload, "upload", false, "Upload files")
	flag.BoolVar(&dryRun, "dryrun", false, "Print files that would have been uploaded")

	flag.Parse()

	githubToken = os.Getenv("GITHUB_API_TOKEN")
	if githubToken == "" {
		errText := fmt.Sprintf("Environment variable required but missing: %s", "GITHUB_API_TOKEN")
		log.Fatal(errText)
	}

	filesUploaded = make(chan int)
}

// Validate user passed arguments
func checkForArgumentErrors() {
	if !upload && !dryRun {
		flag.Usage()
		log.Fatal("Nothing to do")
	}

	if upload && dryRun {
		log.Fatal("Can't upload and dryrun at the same time")
	}

	if upload && flag.NArg() == 0 {
		log.Fatal("Selected -upload with no files")
	}

	if dryRun {
		log.Printf("Doing dry run, nothing will be uploaded")
	}
}

// Parse files from filesystem and build files to upload
func getFilesToUpload(files []string) ([]LocalFile, error) {
	var filesToUpload []LocalFile

	for _, file := range files {
		localFile := LocalFile{file}

		if _, err := localFile.Exists(); err != nil {
			log.Printf("[WARNING] %s not found, excluding from upload", localFile.filepath)
		} else {
			filesToUpload = append(filesToUpload, localFile)
		}
	}

	if len(filesToUpload) == 0 {
		return filesToUpload, errors.New("go-gist: no files to upload")
	}

	return filesToUpload, nil
}

// Process files with, unless a dryrun
func processFiles(ctx context.Context, uploadClient *github.Client, filesToProcess []LocalFile) {
	for _, file := range filesToProcess {
		go file.Upload(ctx, uploadClient)
	}
}

// Monitor file upload progress from goroutines
func checkUploadsFinished(totalUploads int) {
	uploadedFiles := 0
	for {
		if uploadedFiles < totalUploads {
			uploadedFiles += <-filesUploaded
		} else {
			return
		}
	}
}

func main() {
	checkForArgumentErrors()

	argFiles := flag.Args()

	filesToUpload, err := getFilesToUpload(argFiles)
	if err != nil {
		log.Fatal(err)
	}

	// Create GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	processFiles(ctx, client, filesToUpload)

	checkUploadsFinished(len(filesToUpload))
}
