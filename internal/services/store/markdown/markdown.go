package markdown

import (
	"bufio"
	"fmt"
	"markdown-editor/internal/models"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type MarkdownStore struct {
	baseDir string
}

func New(baseDir string) (*MarkdownStore, error) {
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("markdown directory does not exist: %s", baseDir)
	}
	return &MarkdownStore{baseDir: baseDir}, nil
}

func (m *MarkdownStore) ListFiles() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(m.baseDir, "*.md"))
	if err != nil {
		return nil, err
	}

	// Get just the filenames
	var filenames []string
	for _, f := range files {
		filenames = append(filenames, filepath.Base(f))
	}
	return filenames, nil
}

func (m *MarkdownStore) GetLinks(filename string) ([]models.Subject, error) {
	file, err := os.Open(filepath.Join(m.baseDir, filename))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var subjects []models.Subject
	var currentSubject *models.Subject
	linkRegex := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "### ") {
			if currentSubject != nil {
				subjects = append(subjects, *currentSubject)
			}
			currentSubject = &models.Subject{
				Subject: strings.TrimSpace(strings.TrimPrefix(line, "### ")),
			}
		} else if strings.HasPrefix(line, "-") {
			if currentSubject == nil {
				currentSubject = &models.Subject{Subject: "Others"}
			}
			matches := linkRegex.FindStringSubmatch(line)
			if len(matches) == 3 {
				currentSubject.Links = append(currentSubject.Links, models.Link{
					Title: matches[1],
					URL:   matches[2],
				})
			}
		}
	}

	if currentSubject != nil {
		subjects = append(subjects, *currentSubject)
	}

	return subjects, scanner.Err()
}

func (m *MarkdownStore) AddBulkLinks(filename string, subject string, links []models.Link) error {
	filePath := filepath.Join(m.baseDir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var builder strings.Builder
	if len(content) > 0 {
		builder.Write(content)
		if !strings.HasSuffix(string(content), "\n\n") {
			builder.WriteString("\n\n")
		}
	}

	// Format links in markdown format
	if subject != "" {
		builder.WriteString(fmt.Sprintf("### %s\n\n", subject))
	}

	for _, link := range links {
		builder.WriteString(fmt.Sprintf("- [%s](%s)\n", link.Title, link.URL))
	}
	builder.WriteString("\n")

	return os.WriteFile(filePath, []byte(builder.String()), 0644)
}

func (m *MarkdownStore) DeleteLinks(filename string, links []models.Link) error {
	filePath := filepath.Join(m.baseDir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	linksToDelete := make(map[string]bool)

	for _, link := range links {
		linksToDelete[fmt.Sprintf("[%s](%s)", link.Title, link.URL)] = true
	}

	for _, line := range lines {
		shouldDelete := false
		for linkPattern := range linksToDelete {
			if strings.Contains(line, linkPattern) {
				shouldDelete = true
				break
			}
		}
		if !shouldDelete {
			newLines = append(newLines, line)
		}
	}

	return os.WriteFile(filePath, []byte(strings.Join(newLines, "\n")), 0644)
}

func (m *MarkdownStore) CreateFile(filename string) error {
	filePath := filepath.Join(m.baseDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return os.WriteFile(filePath, []byte(""), 0644)
	}
	return nil
}

func (m *MarkdownStore) DeleteFile(filename string) error {
	filePath := filepath.Join(m.baseDir, filename)
	return os.Remove(filePath)
}
