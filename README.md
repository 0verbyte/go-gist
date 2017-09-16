# go-gist
Upload local files directly to [GitHub Gist](https://gist.github.com/).

## Usage
Create a [personal API token](https://github.com/settings/tokens) on GitHub. Save the generated token to an environment variable on your machine called `GITHUB_API_TOKEN`.

Once you've cloned this repo, just build it with `go build gogist.go`. This will place a binary in the same directory called `gogist`, that you can run.

```
% ./gogist -h
Usage of ./gogist:
  -dryrun
    	Print files that would have been uploaded
  -upload
    	Upload files
```

## Example

```
% ./gogist -upload <filepath> <filepath>
```

`gogist` will log which files have been uploaded and the url to view the file on gist

_Note: all files are uploaded as secret gists_

## Disclaimer
This isn't some master piece it's more of an brute force on solving an annoying problem when sharing code. I will probably add new things as I need them, but for now this suites my needs, use at your own will :).
