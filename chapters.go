package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)


type Chapter struct {
	Index     int
	Title     string
	StartSecs float64
	EndSecs   float64
	Filename  string // e.g. "chapter-001.m4a"
}

type ffprobeOutput struct {
	Chapters []struct {
		StartTime string            `json:"start_time"`
		EndTime   string            `json:"end_time"`
		Tags      map[string]string `json:"tags"`
	} `json:"chapters"`
}

// ExtractChapterMeta reads chapter metadata from an m4b file via ffprobe.
func ExtractChapterMeta(m4bPath string) ([]Chapter, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_chapters",
		m4bPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe: %w", err)
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(out, &probe); err != nil {
		return nil, fmt.Errorf("parsing ffprobe output: %w", err)
	}

	chapters := make([]Chapter, 0, len(probe.Chapters))
	for i, c := range probe.Chapters {
		start, err := strconv.ParseFloat(c.StartTime, 64)
		if err != nil {
			return nil, fmt.Errorf("chapter %d start_time: %w", i, err)
		}
		end, err := strconv.ParseFloat(c.EndTime, 64)
		if err != nil {
			return nil, fmt.Errorf("chapter %d end_time: %w", i, err)
		}

		title := c.Tags["title"]
		if title == "" {
			title = fmt.Sprintf("Chapter %d", i+1)
		}

		chapters = append(chapters, Chapter{
			Index:     i + 1,
			Title:     title,
			StartSecs: start,
			EndSecs:   end,
			Filename:  fmt.Sprintf("chapter-%03d.m4a", i+1),
		})
	}

	// No chapter markers — treat the whole file as a single episode.
	if len(chapters) == 0 {
		stem := strings.TrimSuffix(filepath.Base(m4bPath), ".m4b")
		chapters = append(chapters, Chapter{
			Index:    1,
			Title:    stem,
			Filename: "chapter-001.m4a",
		})
	}

	return chapters, nil
}

// ExtractAudio writes each chapter as a separate m4a into audioDir.
// Already-present files are skipped. Returns the number of newly extracted files.
func ExtractAudio(m4bPath, audioDir string, chapters []Chapter) (int, error) {
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		return 0, err
	}

	extracted := 0
	for _, ch := range chapters {
		dest := filepath.Join(audioDir, ch.Filename)
		if _, err := os.Stat(dest); err == nil {
			Skip("%s — already exists, kopeng", ch.Filename)
			continue
		}

		Info("[%03d]  %s", ch.Index, ch.Title)
		Detail("       %s → %s   out: %s",
			secsToTimestamp(ch.StartSecs), secsToTimestamp(ch.EndSecs), ch.Filename)

		args := []string{"-y", "-i", m4bPath}
		if ch.EndSecs > ch.StartSecs {
			args = append(args,
				"-ss", secsToTimestamp(ch.StartSecs),
				"-to", secsToTimestamp(ch.EndSecs),
			)
		}
		// Copy stream, strip chapter metadata from the output slice.
		args = append(args, "-c", "copy", "-map_chapters", "-1", dest)

		cmd := exec.Command("ffmpeg", args...)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return extracted, fmt.Errorf("extracting %q: %w", ch.Title, err)
		}
		Done("%s", ch.Filename)
		extracted++
	}
	return extracted, nil
}

func secsToTimestamp(s float64) string {
	h := int(s) / 3600
	m := (int(s) % 3600) / 60
	sec := s - float64(h*3600+m*60)
	return fmt.Sprintf("%02d:%02d:%06.3f", h, m, sec)
}
