// Package secret scans artifact content for accidentally included secrets.
package secret

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/zricethezav/gitleaks/v8/detect"
	"github.com/zricethezav/gitleaks/v8/report"
	"github.com/zricethezav/gitleaks/v8/sources"
)

var (
	detectorOnce sync.Once
	detectorInst *detect.Detector
	detectorErr  error
)

// Check returns an error if gitleaks detects a secret in data.
func Check(name string, data []byte) error {
	findings, err := scan(context.Background(), name, data)
	if err != nil {
		return fmt.Errorf("secret scan failed for %s: %w", name, err)
	}
	if len(findings) == 0 {
		return nil
	}
	return Error{name: name, findings: findings}
}

func scan(ctx context.Context, name string, data []byte) ([]report.Finding, error) {
	d, err := detector()
	if err != nil {
		return nil, err
	}
	return d.DetectSource(ctx, &sources.File{
		Content:         bytes.NewReader(data),
		Path:            name,
		MaxArchiveDepth: 0,
		Config:          &d.Config,
	})
}

func detector() (*detect.Detector, error) {
	detectorOnce.Do(func() {
		detectorInst, detectorErr = detect.NewDetectorDefaultConfig()
		if detectorErr != nil {
			return
		}
		// Never echo captured secrets in CLI errors. We only need rule/path/line.
		detectorInst.Redact = 100
		detectorInst.MaxArchiveDepth = 0
	})
	return detectorInst, detectorErr
}

// Error is returned when gitleaks finds one or more secrets.
type Error struct {
	name     string
	findings []report.Finding
}

// Error returns a redacted summary of detected secrets.
func (e Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "secret scan blocked %s: gitleaks detected %d potential secret", e.name, len(e.findings))
	if len(e.findings) != 1 {
		b.WriteString("s")
	}
	b.WriteString("\n")

	limit := len(e.findings)
	if limit > 5 {
		limit = 5
	}
	for i := 0; i < limit; i++ {
		f := e.findings[i]
		file := f.File
		if file == "" {
			file = e.name
		}
		line := ""
		if f.StartLine > 0 {
			line = fmt.Sprintf(":%d", f.StartLine)
		}
		desc := f.Description
		if desc == "" {
			desc = f.RuleID
		}
		fmt.Fprintf(&b, "- %s%s: %s", file, line, desc)
		if f.RuleID != "" && f.RuleID != desc {
			fmt.Fprintf(&b, " (%s)", f.RuleID)
		}
		b.WriteString("\n")
	}
	if extra := len(e.findings) - limit; extra > 0 {
		fmt.Fprintf(&b, "- ... and %d more\n", extra)
	}
	b.WriteString("Use --no-secret-scan only if this artifact is intentionally safe to publish.")
	return strings.TrimRight(b.String(), "\n")
}
