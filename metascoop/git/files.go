package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// ChangedFiles returns the paths git considers modified, including untracked
// ones, relative to the repository root.
func ChangedFiles(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "status", "--porcelain", "--untracked-files=all")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running git status: %w\nOutput:\n%s", err, output)
	}

	var paths []string
	for _, line := range strings.Split(string(output), "\n") {
		if len(line) <= 3 {
			continue
		}

		path := strings.TrimSpace(line[3:])
		if _, renamed, ok := strings.Cut(path, " -> "); ok {
			path = renamed
		}

		paths = append(paths, strings.Trim(path, `"`))
	}

	return paths, nil
}
