package analysis

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	pdf "github.com/ledongthuc/pdf"
)

type Service struct {
	maxFileSize int64
}

func NewService(maxFileSize int64) *Service {
	if maxFileSize <= 0 {
		maxFileSize = 20 << 20
	}
	return &Service{maxFileSize: maxFileSize}
}

func (s *Service) MaxFileSize() int64 {
	return s.maxFileSize
}

func (s *Service) Analyze(ctx context.Context, data []byte, filename string) (AnalyzeResult, error) {
	if len(data) == 0 {
		return AnalyzeResult{}, errors.New("file is empty")
	}
	if int64(len(data)) > s.maxFileSize {
		return AnalyzeResult{}, fmt.Errorf("file exceeds max size of %d bytes", s.maxFileSize)
	}

	tmp, err := os.CreateTemp("", "pdf-analysis-*.pdf")
	if err != nil {
		return AnalyzeResult{}, err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		return AnalyzeResult{}, err
	}
	if err := tmp.Close(); err != nil {
		return AnalyzeResult{}, err
	}

	select {
	case <-ctx.Done():
		return AnalyzeResult{}, ctx.Err()
	default:
	}

	file, reader, err := pdf.Open(tmpPath)
	if err != nil {
		return AnalyzeResult{}, errors.New("failed to parse pdf")
	}
	defer file.Close()

	plainTextReader, err := reader.GetPlainText()
	if err != nil {
		return AnalyzeResult{}, errors.New("failed to extract text from pdf")
	}
	textBytes, err := io.ReadAll(plainTextReader)
	if err != nil {
		return AnalyzeResult{}, err
	}
	text := string(textBytes)

	result := AnalyzeResult{
		SourceFilename:       filename,
		PageCount:            reader.NumPage(),
		WordCount:            countWords(text),
		ImageCountEstimate:   strings.Count(string(data), "/Subtype /Image"),
		HasDigitalSignature:  strings.Contains(string(data), "/Type /Sig") || strings.Contains(string(data), "/Sig"),
		Metadata:             parseMetadataFromRaw(string(data)),
		TableRowsDetected:    detectTableRows(text, 20),
		ExtractedTextPreview: preview(text, 1200),
	}
	return result, nil
}

var (
	metadataTitleRe   = regexp.MustCompile(`/Title\s*\(([^)]{1,200})\)`)
	metadataAuthorRe  = regexp.MustCompile(`/Author\s*\(([^)]{1,200})\)`)
	metadataSubjectRe = regexp.MustCompile(`/Subject\s*\(([^)]{1,200})\)`)
	metadataCreatorRe = regexp.MustCompile(`/Creator\s*\(([^)]{1,200})\)`)
	metadataProdRe    = regexp.MustCompile(`/Producer\s*\(([^)]{1,200})\)`)
)

func parseMetadataFromRaw(raw string) map[string]string {
	out := map[string]string{}
	trySet := func(key string, re *regexp.Regexp) {
		if match := re.FindStringSubmatch(raw); len(match) > 1 {
			out[key] = strings.TrimSpace(match[1])
		}
	}
	trySet("title", metadataTitleRe)
	trySet("author", metadataAuthorRe)
	trySet("subject", metadataSubjectRe)
	trySet("creator", metadataCreatorRe)
	trySet("producer", metadataProdRe)
	if len(out) == 0 {
		return nil
	}
	return out
}

var wordTokenRe = regexp.MustCompile(`[A-Za-z0-9]+`)

func countWords(text string) int {
	return len(wordTokenRe.FindAllString(text, -1))
}

var tableRowRe = regexp.MustCompile(`\S+\s{2,}\S+`)

func detectTableRows(text string, limit int) []string {
	if limit <= 0 {
		limit = 20
	}

	rows := make([]string, 0, limit)
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "|") || tableRowRe.MatchString(trimmed) {
			rows = append(rows, trimmed)
			if len(rows) >= limit {
				break
			}
		}
	}
	return rows
}

func preview(text string, max int) string {
	text = strings.TrimSpace(text)
	if len(text) <= max {
		return text
	}
	return text[:max]
}
