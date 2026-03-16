package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Itunes  string   `xml:"xmlns:itunes,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string  `xml:"title"`
	Link        string  `xml:"link"`
	Description string  `xml:"description"`
	Language    string  `xml:"language"`
	Author      string  `xml:"itunes:author"`
	Summary     string  `xml:"itunes:summary"`
	Items       []Item  `xml:"item"`
}

type Item struct {
	Title     string    `xml:"title"`
	PubDate   string    `xml:"pubDate"`
	GUID      GUID      `xml:"guid"`
	Enclosure Enclosure `xml:"enclosure"`
	Duration  string    `xml:"itunes:duration,omitempty"`
	Episode   int       `xml:"itunes:episode"`
}

type GUID struct {
	IsPermaLink string `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

type Enclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type FeedConfig struct {
	Title       string
	Author      string
	Description string
	BaseURL     string // e.g. "https://example.com/mybook"
	AudioDir    string // filesystem path to the audio directory
}

func BuildFeed(cfg FeedConfig, chapters []Chapter) *RSS {
	// Spread pub dates one day apart so episode order is stable in all clients.
	baseTime := time.Now().UTC().Truncate(24 * time.Hour)

	items := make([]Item, len(chapters))
	for i, ch := range chapters {
		var size int64
		if info, err := os.Stat(filepath.Join(cfg.AudioDir, ch.Filename)); err == nil {
			size = info.Size()
		}

		audioURL := cfg.BaseURL + "/audio/" + ch.Filename

		items[i] = Item{
			Title:   ch.Title,
			PubDate: baseTime.Add(time.Duration(i) * 24 * time.Hour).Format(time.RFC1123Z),
			GUID:    GUID{IsPermaLink: "true", Value: audioURL},
			Enclosure: Enclosure{
				URL:    audioURL,
				Length: size,
				Type:   "audio/mp4",
			},
			Duration: formatDuration(ch.EndSecs - ch.StartSecs),
			Episode:  ch.Index,
		}
	}

	return &RSS{
		Version: "2.0",
		Itunes:  "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Channel: Channel{
			Title:       cfg.Title,
			Link:        cfg.BaseURL,
			Description: cfg.Description,
			Language:    "en-us",
			Author:      cfg.Author,
			Summary:     cfg.Description,
			Items:       items,
		},
	}
}

func WriteFeed(path string, feed *RSS) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(xml.Header)
	enc := xml.NewEncoder(f)
	enc.Indent("", "  ")
	if err := enc.Encode(feed); err != nil {
		return err
	}
	return enc.Flush()
}

func WriteIndex(path, title, baseURL string, chapters []Chapter) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, `<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>%s</title></head>
<body>
<h1>%s</h1>
<p><a href="feed.xml">RSS Feed</a></p>
<ol>
`, title, title)

	for _, ch := range chapters {
		fmt.Fprintf(f, "  <li><a href=\"audio/%s\">%s</a></li>\n", ch.Filename, ch.Title)
	}

	fmt.Fprintf(f, "</ol>\n</body>\n</html>\n")
	return nil
}

func formatDuration(secs float64) string {
	if secs <= 0 {
		return ""
	}
	h := int(secs) / 3600
	m := (int(secs) % 3600) / 60
	s := int(secs) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
