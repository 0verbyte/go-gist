# go-gist
Upload local files directly to [GitHub Gist](https://gist.github.com/).

## Usage
Create a [personal API token](https://github.com/settings/tokens) on GitHub. Save the generated token to an environment variable on your machine called `GITHUB_API_TOKEN`.

Once you've cloned this repo, run the Makefile (`make`). This will place a binary in the same directory called `gogist`, that you can run.

```
% ./gogist -h
Usage of ./gogist:
  -allow-large-files
    	Override max upload size 50KB
  -dryrun
    	Print files that would have been uploaded
  -upload
    	Upload files
```

## Example

```
% ./gogist -upload <filepath> <filepath>
```

`gogist` will log which files have been uploaded and the url to view the file on gist.
If the file you are uploading is large (check `-h` for exact max upload size), you should use the flag `-allow-large-files`.

_Note: all files are uploaded as secret gists_
