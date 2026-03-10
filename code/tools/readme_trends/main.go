package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

type entry struct {
	title   string
	url     string
	snippet string
}

const (
	startMarker = "<!-- S3_TRENDS_START -->"
	endMarker   = "<!-- S3_TRENDS_END -->"
)

func main() {
	inputPath := flag.String("input", "", "Path to SerpAPI stdout text")
	readmePath := flag.String("readme", "", "Path to README.md to update")
	maxResults := flag.Int("max", 5, "Maximum number of items to include")
	flag.Parse()

	if *inputPath == "" || *readmePath == "" {
		exitWithError(errors.New("--input and --readme are required"))
	}

	inputBytes, err := os.ReadFile(*inputPath)
	if err != nil {
		exitWithError(fmt.Errorf("read input: %w", err))
	}

	entries := parseSerpOutput(string(inputBytes), *maxResults)
	if len(entries) == 0 {
		exitWithError(errors.New("no entries parsed from SerpAPI output"))
	}

	readmeBytes, err := os.ReadFile(*readmePath)
	if err != nil {
		exitWithError(fmt.Errorf("read readme: %w", err))
	}

	updated, err := updateReadme(string(readmeBytes), entries)
	if err != nil {
		exitWithError(err)
	}

	if err := os.WriteFile(*readmePath, []byte(updated), 0644); err != nil {
		exitWithError(fmt.Errorf("write readme: %w", err))
	}
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

func parseSerpOutput(input string, max int) []entry {
	scanner := bufio.NewScanner(strings.NewReader(input))
	titleRe := regexp.MustCompile(`^\d+\.\s+(.+)$`)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var results []entry
	for i := 0; i < len(lines) && len(results) < max; i++ {
		line := strings.TrimSpace(lines[i])
		match := titleRe.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		title := match[1]
		url, urlIndex := nextNonEmpty(lines, i+1)
		if url == "" || !strings.HasPrefix(strings.TrimSpace(url), "http") {
			continue
		}

		snippet, _ := nextNonEmpty(lines, urlIndex+1)
		if titleRe.MatchString(strings.TrimSpace(snippet)) {
			snippet = ""
		}

		results = append(results, entry{
			title:   strings.TrimSpace(title),
			url:     strings.TrimSpace(url),
			snippet: strings.TrimSpace(snippet),
		})
		i = urlIndex
	}

	return results
}

func nextNonEmpty(lines []string, start int) (string, int) {
	for i := start; i < len(lines); i++ {
		value := strings.TrimSpace(lines[i])
		if value != "" {
			return value, i
		}
	}
	return "", len(lines)
}

func updateReadme(readme string, entries []entry) (string, error) {
	start := strings.Index(readme, startMarker)
	end := strings.Index(readme, endMarker)
	if start == -1 || end == -1 || start > end {
		return "", errors.New("README markers not found; add the S3 trends section first")
	}

	content := buildSection(entries)
	replacement := startMarker + "\n" + content + "\n" + endMarker

	updated := readme[:start] + replacement + readme[end+len(endMarker):]
	return updated, nil
}

func buildSection(entries []entry) string {
	date := time.Now().UTC().Format("2006-01-02")
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Last updated: %s (UTC)\n\n", date))

	for _, item := range entries {
		line := fmt.Sprintf("- [%s](%s)", escapeMarkdown(item.title), item.url)
		if item.snippet != "" {
			line += fmt.Sprintf(" — %s", escapeMarkdown(item.snippet))
		}
		builder.WriteString(line + "\n")
	}

	return strings.TrimRight(builder.String(), "\n")
}

func escapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"[", "\\[",
		"]", "\\]",
	)
	return replacer.Replace(text)
}
