package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/github"
	"github.com/voidpirate/go-gist/file"
	"golang.org/x/oauth2"
)

var (
	upload          bool
	dryRun          bool
	allowLargeFiles bool
	files           string
	githubToken     string
	maxFilesize     = file.KB * 50

	statusChan chan error
)

func init() {
	flag.BoolVar(&upload, "upload", false, "Upload files")
	flag.BoolVar(&dryRun, "dryrun", false, "Print files that would have been uploaded")

	largeFilesText := fmt.Sprintf("Override max upload size %2.fKB", maxFilesize/file.KB)
	flag.BoolVar(&allowLargeFiles, "allow-large-files", false, largeFilesText)

	flag.Parse()

	githubToken = os.Getenv("GITHUB_API_TOKEN")
	if githubToken == "" {
		errText := fmt.Sprintf("Environment variable required but missing: %s", "GITHUB_API_TOKEN")
		log.Fatal(errText)
	}

	statusChan = make(chan error)
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
func getFilesToUpload(files []string) ([]file.LocalFile, error) {
	var filesToUpload []file.LocalFile

	for _, f := range files {
		localFile := file.New(f, dryRun, &statusChan)

		// Check that the file exists
		if _, err := localFile.Exists(); err != nil {
			log.Printf("[WARNING] %s not found, excluding from upload", localFile.FilePath)
			continue
		}

		// Check the file size to prevent large upload size
		if !allowLargeFiles {
			localFilesize := localFile.Size()
			if localFilesize > maxFilesize {
				log.Printf("[WARNING] excluding %s from upload. File exceeds %2.fKB: %2.f (%2.fKB)",
					localFile.FilePath, maxFilesize/file.KB, localFilesize, localFilesize/file.KB)
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
func processFiles(ctx context.Context, uploadClient *github.Client, filesToProcess []file.LocalFile) {
	for _, f := range filesToProcess {
		// Don't spawn a thread if we are running in dryrun mode, just call it
		// normally and bail out.
		if dryRun {
			f.Upload(ctx, uploadClient)
			continue
		}

		go f.Upload(ctx, uploadClient)
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

	uploadedFiles := 0
	for err := range statusChan {
		if err != nil {
			log.Fatal(err)
		}

		uploadedFiles++
		if uploadedFiles == len(filesToUpload) {
			return
		}
	}
}
