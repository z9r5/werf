package path_matcher

import (
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/docker/docker/pkg/fileutils"

	"github.com/werf/werf/pkg/util"
)

func newDockerfileIgnorePathMatcher(dockerignorePatterns []string) dockerfileIgnorePathMatcher {
	m, err := fileutils.NewPatternMatcher(dockerignorePatterns)
	if err != nil {
		panic(err)
	}

	return dockerfileIgnorePathMatcher{
		patternMatcher: m,
	}
}

type dockerfileIgnorePathMatcher struct {
	patternMatcher *fileutils.PatternMatcher
}

func (m dockerfileIgnorePathMatcher) IsPathMatched(path string) bool {
	path = formatPath(path)
	if m.patternMatcher == nil {
		return true
	}

	isMatched, err := m.patternMatcher.Matches(path)
	if err != nil {
		panic(err)
	}

	return !isMatched
}

type pattern struct {
	pattern      string
	exclusion    bool
	isMatched    bool
	isInProgress bool
}

func (m dockerfileIgnorePathMatcher) IsDirOrSubmodulePathMatched(path string) bool {
	return m.IsPathMatched(path) || m.ShouldGoThrough(path)
}

func (m dockerfileIgnorePathMatcher) ShouldGoThrough(path string) bool {
	return m.shouldGoThrough(formatPath(path))
}

func (m dockerfileIgnorePathMatcher) shouldGoThrough(path string) bool {
	if m.patternMatcher == nil || len(m.patternMatcher.Patterns()) == 0 {
		return false
	}

	if path == "" {
		return true
	}

	pathParts := util.SplitFilepath(path)
	var patterns []*pattern
	for _, p := range m.patternMatcher.Patterns() {
		patterns = append(patterns, &pattern{
			pattern:      p.String(),
			exclusion:    p.Exclusion(),
			isInProgress: true,
		})
	}

	for _, pathPart := range pathParts {
		for _, p := range patterns {
			if !p.isInProgress {
				continue
			}

			inProgressGlob, matchedGlob := matchGlob(pathPart, p.pattern)
			if inProgressGlob != "" {
				p.pattern = inProgressGlob
			} else if matchedGlob != "" {
				p.isMatched = true
				p.isInProgress = false
			} else {
				p.isInProgress = false
			}
		}
	}

	shouldGoThrough := false
	for _, pattern := range patterns {
		if pattern.isMatched {
			shouldGoThrough = false
		} else if pattern.isInProgress {
			shouldGoThrough = true
		}
	}

	return shouldGoThrough
}

func (m dockerfileIgnorePathMatcher) ID() string {
	var cleanedPatterns []string
	for _, pattern := range m.patternMatcher.Patterns() {
		cleanedPatterns = append(cleanedPatterns, pattern.String())
	}

	if len(cleanedPatterns) == 0 {
		return ""
	}

	h := sha256.New()
	sort.Strings(cleanedPatterns)
	h.Write([]byte(fmt.Sprint(cleanedPatterns)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (m dockerfileIgnorePathMatcher) String() string {
	return fmt.Sprintf("{ patternMatcher=%v }", m.patternMatcher.Patterns())
}
