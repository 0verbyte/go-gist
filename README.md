# go-gist
Upload local files directly to [GitHub Gist](https://gist.github.com/).

>**Note:** All files will be uploaded as secret gists

## Usage
Create a [personal API token](https://github.com/settings/tokens) on GitHub. Save the generated token to an environment variable on your machine called `GITHUB_API_TOKEN`.

Once you've cloned this repo, build the go-gist binary `go build`.

```
% ./go-gist -h
Usage of ./go-gist:
  -allow-large-files
    	Override max upload size 50KB
  -dryrun
    	Print files that would have been uploaded
  -upload
    	Upload files
```

## Example

```
% ./go-gist -upload <filepath> <filepath>
```

`go-gist` will log which files have been uploaded and the url to view the file on gist.
If the file you are uploading is large (check `-h` for exact max upload size), you should use the flag `-allow-large-files`.
