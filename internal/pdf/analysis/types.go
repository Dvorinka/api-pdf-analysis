package analysis

type AnalyzeResult struct {
	SourceFilename       string            `json:"source_filename,omitempty"`
	PageCount            int               `json:"page_count"`
	WordCount            int               `json:"word_count"`
	ImageCountEstimate   int               `json:"image_count_estimate"`
	HasDigitalSignature  bool              `json:"has_digital_signature"`
	Metadata             map[string]string `json:"metadata,omitempty"`
	TableRowsDetected    []string          `json:"table_rows_detected,omitempty"`
	ExtractedTextPreview string            `json:"extracted_text_preview"`
}
