# abcast

Publishes an m4b audiobook as a podcast RSS feed. Each chapter becomes a separate episode. The output is a static directory that can be hosted anywhere.

## Requirements

- Go 1.21+
- ffmpeg / ffprobe
- Azure CLI (`az`) — for uploading

## Build

```bash
go build -o abcast .
```

## Usage

### One-shot publish to Azure

```bash
./publish.sh <storage-account> <audiobook.m4b>
```

This converts the audiobook, creates an Azure Blob Storage container named after the file, uploads everything, and prints the feed URL.

### Manual steps

**Convert:**
```bash
./abcast -file mybook.m4b -url https://example.com/mybook [-out ./mybook] [-title "My Book"] [-author "Jane Doe"] [-desc "Description"]
```

**Upload:**
```bash
./upload.sh -a <storage-account> -c <container> -d <output-dir> [-s <subscription>]
```

Use `$web` as the container name to enable Azure static website hosting instead of plain blob hosting.

## Output structure

```
<outdir>/
  feed.xml        RSS 2.0 feed (one episode per chapter)
  index.html      Simple chapter listing
  audio/
    chapter-001.m4a
    chapter-002.m4a
    ...
```

Upload the contents of `<outdir>/` to any static web server and subscribe to `feed.xml`.
