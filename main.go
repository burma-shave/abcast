package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var (
		m4bPath     = flag.String("file", "", "Path to the .m4b audiobook file (required)")
		baseURL     = flag.String("url", "", "Public base URL where the output will be hosted (required)\n    e.g. https://example.com/mybook")
		outDir      = flag.String("out", "", "Output directory (default: <file stem>/)")
		title       = flag.String("title", "", "Feed title (default: filename stem)")
		author      = flag.String("author", "", "Author / podcast artist")
		description = flag.String("desc", "", "Feed description")
	)
	flag.Parse()

	if *m4bPath == "" || *baseURL == "" {
		DrummerUsageError()
		flag.PrintDefaults()
		os.Exit(1)
	}

	if _, err := os.Stat(*m4bPath); err != nil {
		Fatal("can't find that file — %v", err)
	}

	stem := strings.TrimSuffix(filepath.Base(*m4bPath), filepath.Ext(*m4bPath))
	if *title == "" {
		*title = stem
	}
	if *outDir == "" {
		*outDir = filepath.Join(filepath.Dir(*m4bPath), stem)
	}

	audioDir := filepath.Join(*outDir, "audio")

	Banner(*title)
	DrummerGreet(*title)


	// ── 1. Read chapter markers ───────────────────────────────────────────────
	Step("Reading chapter metadata")
	Info("source  : %s", *m4bPath)

	chapters, err := ExtractChapterMeta(*m4bPath)
	if err != nil {
		Fatal("chapter extraction failed: %v", err)
	}

	DrummerChapterFound(len(chapters))

	for _, ch := range chapters {
		Detail("[%03d]  %-40s  %s → %s",
			ch.Index, ch.Title,
			formatDuration(ch.StartSecs),
			formatDuration(ch.EndSecs),
		)
	}

	// ── 2. Extract per-chapter audio files ───────────────────────────────────
	Step("Extracting chapter audio")
	Info("dest    : %s", audioDir)
	DrummerExtracting()

	extracted, err := ExtractAudio(*m4bPath, audioDir, chapters)
	if err != nil {
		Fatal("audio extraction failed: %v", err)
	}

	DrummerExtractDone(extracted)

	// ── 3. Write feed.xml ─────────────────────────────────────────────────────
	Step("Building RSS feed")
	DrummerWritingFeed()

	feed := BuildFeed(FeedConfig{
		Title:       *title,
		Author:      *author,
		Description: *description,
		BaseURL:     strings.TrimRight(*baseURL, "/"),
		AudioDir:    audioDir,
	}, chapters)

	feedPath := filepath.Join(*outDir, "feed.xml")
	if err := WriteFeed(feedPath, feed); err != nil {
		Fatal("writing feed.xml: %v", err)
	}
	Done("%s", feedPath)

	// ── 4. Write index.html ───────────────────────────────────────────────────
	Step("Writing index page")

	indexPath := filepath.Join(*outDir, "index.html")
	if err := WriteIndex(indexPath, *title, *baseURL, chapters); err != nil {
		Fatal("writing index.html: %v", err)
	}
	Done("%s", indexPath)

	DrummerDone(*outDir)
}
