package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
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

var (
	upload          bool
	dryRun          bool
	allowLargeFiles bool
	files           string
	githubToken     string
	maxFilesize     = KB * 50

	// Channel buffer for finished file uploads
	filesUploaded chan int
)

func init() {
	flag.BoolVar(&upload, "upload", false, "Upload files")
	flag.BoolVar(&dryRun, "dryrun", false, "Print files that would have been uploaded")

	largeFilesText := fmt.Sprintf("Override max upload size %2.fKB", maxFilesize/KB)
	flag.BoolVar(&allowLargeFiles, "allow-large-files", false, largeFilesText)

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
		localFile := NewLocalFile(file)

		// Check that the file exists
		if _, err := localFile.Exists(); err != nil {
			log.Printf("[WARNING] %s not found, excluding from upload", localFile.filepath)
			continue
		}

		// Check the file size to prevent large upload size
		if !allowLargeFiles {
			localFilesize := localFile.GetFilesize()
			if localFilesize > maxFilesize {
				log.Printf("[WARNING] excluding %s from upload. File exceeds %2.fKB: %2.f (%2.fKB)", localFile.filepath, maxFilesize/KB, localFilesize, localFilesize/KB)
				continue
			}
		}

		filesToUpload = append(filesToUpload, *localFile)
	}

	if len(filesToUpload) == 0 {
		return filesToUpload, errors.New("go-gist: no files to upload")
	}

	return filesToUpload, nil
}

// Process files handle uploading a files unless dry run is used
func processFiles(ctx context.Context, uploadClient *github.Client, filesToProcess []LocalFile) {
	for _, file := range filesToProcess {
		// Don't spawn a thread if we are running in dryrun mode, just call it
		// normally and bail out.
		if dryRun {
			file.Upload(ctx, uploadClient)
			continue
		}

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

	// Don't check uploads when dryrun is used
	if !dryRun {
		checkUploadsFinished(len(filesToUpload))
	}
}
