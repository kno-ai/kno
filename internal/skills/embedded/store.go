package embedded

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/kno-ai/kno/internal/skills"
)

//go:embed all:skills
var skillsFS embed.FS

type Store struct{}

var _ skills.Store = (*Store)(nil)

func New() *Store {
	return &Store{}
}

func (s *Store) Get(name string) (string, error) {
	data, err := skillsFS.ReadFile("skills/" + name)
	if err != nil {
		return "", fmt.Errorf("skill %q not found: %w", name, err)
	}
	return string(data), nil
}

func (s *Store) List() ([]string, error) {
	var names []string
	err := fs.WalkDir(skillsFS, "skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			name := strings.TrimPrefix(path, "skills/")
			// Filter out template files — they're loaded via Get(), not listed as skills.
			if strings.HasPrefix(name, "templates/") {
				return nil
			}
			names = append(names, name)
		}
		return nil
	})
	return names, err
}
