package config

import (
	"testing"
)

// TestParse_ValidAliasLines tests parsing of valid alias lines
func TestParse_ValidAliasLines(t *testing.T) {
	parser := NewParser()

	testCases := []struct {
		name     string
		content  string
		expected []Entry
	}{
		{
			name:    "single alias",
			content: "alias ll 'ls -la'",
			expected: []Entry{
				{Type: "alias", Name: "ll", Definition: "ls -la"},
			},
		},
		{
			name:    "multiple aliases",
			content: "alias ll 'ls -la'\nalias gs 'git status'",
			expected: []Entry{
				{Type: "alias", Name: "ll", Definition: "ls -la"},
				{Type: "alias", Name: "gs", Definition: "git status"},
			},
		},
		{
			name:    "alias with complex definition",
			content: "alias glog 'git log --oneline --graph --decorate'",
			expected: []Entry{
				{Type: "alias", Name: "glog", Definition: "git log --oneline --graph --decorate"},
			},
		},
		{
			name:    "alias with pipes and redirects",
			content: "alias count 'ls -1 | wc -l'",
			expected: []Entry{
				{Type: "alias", Name: "count", Definition: "ls -1 | wc -l"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entries, err := parser.Parse(tc.content)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if len(entries) != len(tc.expected) {
				t.Fatalf("Expected %d entries, got %d", len(tc.expected), len(entries))
			}

			for i, entry := range entries {
				if entry.Type != tc.expected[i].Type {
					t.Errorf("Entry %d: expected type %q, got %q", i, tc.expected[i].Type, entry.Type)
				}
				if entry.Name != tc.expected[i].Name {
					t.Errorf("Entry %d: expected name %q, got %q", i, tc.expected[i].Name, entry.Name)
				}
				if entry.Definition != tc.expected[i].Definition {
					t.Errorf("Entry %d: expected definition %q, got %q", i, tc.expected[i].Definition, entry.Definition)
				}
			}
		})
	}
}

// TestParse_ValidAbbrLines tests parsing of valid abbr lines
func TestParse_ValidAbbrLines(t *testing.T) {
	parser := NewParser()

	testCases := []struct {
		name     string
		content  string
		expected []Entry
	}{
		{
			name:    "single abbr",
			content: "abbr -a gco 'git checkout'",
			expected: []Entry{
				{Type: "abbr", Name: "gco", Definition: "git checkout"},
			},
		},
		{
			name:    "multiple abbrs",
			content: "abbr -a gco 'git checkout'\nabbr -a gp 'git push'",
			expected: []Entry{
				{Type: "abbr", Name: "gco", Definition: "git checkout"},
				{Type: "abbr", Name: "gp", Definition: "git push"},
			},
		},
		{
			name:    "abbr with complex definition",
			content: "abbr -a gcm 'git commit -m'",
			expected: []Entry{
				{Type: "abbr", Name: "gcm", Definition: "git commit -m"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entries, err := parser.Parse(tc.content)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if len(entries) != len(tc.expected) {
				t.Fatalf("Expected %d entries, got %d", len(tc.expected), len(entries))
			}

			for i, entry := range entries {
				if entry.Type != tc.expected[i].Type {
					t.Errorf("Entry %d: expected type %q, got %q", i, tc.expected[i].Type, entry.Type)
				}
				if entry.Name != tc.expected[i].Name {
					t.Errorf("Entry %d: expected name %q, got %q", i, tc.expected[i].Name, entry.Name)
				}
				if entry.Definition != tc.expected[i].Definition {
					t.Errorf("Entry %d: expected definition %q, got %q", i, tc.expected[i].Definition, entry.Definition)
				}
			}
		})
	}
}

// TestParse_MixedAliasAndAbbr tests parsing of mixed alias and abbr lines
func TestParse_MixedAliasAndAbbr(t *testing.T) {
	parser := NewParser()

	content := `alias ll 'ls -la'
abbr -a gco 'git checkout'
alias gs 'git status'
abbr -a gp 'git push'`

	expected := []Entry{
		{Type: "alias", Name: "ll", Definition: "ls -la"},
		{Type: "abbr", Name: "gco", Definition: "git checkout"},
		{Type: "alias", Name: "gs", Definition: "git status"},
		{Type: "abbr", Name: "gp", Definition: "git push"},
	}

	entries, err := parser.Parse(content)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(entries) != len(expected) {
		t.Fatalf("Expected %d entries, got %d", len(expected), len(entries))
	}

	for i, entry := range entries {
		if entry.Type != expected[i].Type {
			t.Errorf("Entry %d: expected type %q, got %q", i, expected[i].Type, entry.Type)
		}
		if entry.Name != expected[i].Name {
			t.Errorf("Entry %d: expected name %q, got %q", i, expected[i].Name, entry.Name)
		}
		if entry.Definition != expected[i].Definition {
			t.Errorf("Entry %d: expected definition %q, got %q", i, expected[i].Definition, entry.Definition)
		}
	}
}

// TestParse_SkipCommentLines tests that comment lines are skipped
func TestParse_SkipCommentLines(t *testing.T) {
	parser := NewParser()

	testCases := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "single line comment",
			content: `# This is a comment
alias ll 'ls -la'`,
			expected: 1,
		},
		{
			name: "multiple comments",
			content: `# Comment 1
# Comment 2
alias ll 'ls -la'
# Comment 3
alias gs 'git status'`,
			expected: 2,
		},
		{
			name: "comment at end",
			content: `alias ll 'ls -la'
# Final comment`,
			expected: 1,
		},
		{
			name: "only comments",
			content: `# Comment 1
# Comment 2
# Comment 3`,
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entries, err := parser.Parse(tc.content)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if len(entries) != tc.expected {
				t.Errorf("Expected %d entries, got %d", tc.expected, len(entries))
			}
		})
	}
}

// TestParse_SkipEmptyLines tests that empty lines are skipped
func TestParse_SkipEmptyLines(t *testing.T) {
	parser := NewParser()

	testCases := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "empty lines between entries",
			content: `alias ll 'ls -la'

alias gs 'git status'`,
			expected: 2,
		},
		{
			name: "multiple empty lines",
			content: `alias ll 'ls -la'


alias gs 'git status'`,
			expected: 2,
		},
		{
			name: "empty lines at start and end",
			content: `

alias ll 'ls -la'

`,
			expected: 1,
		},
		{
			name: "whitespace-only lines",
			content: `alias ll 'ls -la'
   
	
alias gs 'git status'`,
			expected: 2,
		},
		{
			name:     "only empty lines",
			content:  "\n\n\n",
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entries, err := parser.Parse(tc.content)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if len(entries) != tc.expected {
				t.Errorf("Expected %d entries, got %d", tc.expected, len(entries))
			}
		})
	}
}

// TestParse_InvalidFormat tests handling of invalid format lines
func TestParse_InvalidFormat(t *testing.T) {
	parser := NewParser()

	testCases := []struct {
		name     string
		content  string
		expected int // number of valid entries that should be parsed
	}{
		{
			name:     "missing quotes",
			content:  "alias ll ls -la",
			expected: 0,
		},
		{
			name:     "missing definition",
			content:  "alias ll",
			expected: 0,
		},
		{
			name:     "missing name",
			content:  "alias 'ls -la'",
			expected: 0,
		},
		{
			name:     "abbr without -a flag",
			content:  "abbr gco 'git checkout'",
			expected: 0,
		},
		{
			name:     "invalid command",
			content:  "function test 'echo hello'",
			expected: 0,
		},
		{
			name: "mixed valid and invalid",
			content: `alias ll 'ls -la'
invalid line here
abbr -a gco 'git checkout'`,
			expected: 2,
		},
		{
			name: "partial match should be ignored",
			content: `aliasll 'ls -la'
alias ll'ls -la'`,
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entries, err := parser.Parse(tc.content)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if len(entries) != tc.expected {
				t.Errorf("Expected %d entries, got %d", tc.expected, len(entries))
			}
		})
	}
}

// TestParse_EmptyContent tests parsing of empty content
func TestParse_EmptyContent(t *testing.T) {
	parser := NewParser()

	entries, err := parser.Parse("")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected 0 entries for empty content, got %d", len(entries))
	}
}

// TestParse_RealWorldExample tests a realistic configuration file
func TestParse_RealWorldExample(t *testing.T) {
	parser := NewParser()

	content := `# このファイルは fish-configurator によって自動生成されます
# 手動で編集しないでください

# Aliases
alias ll 'ls -la'
alias gs 'git status'
alias gd 'git diff'

# Abbreviations
abbr -a gco 'git checkout'
abbr -a gp 'git push'
abbr -a gl 'git pull'
abbr -a gcm 'git commit -m'

# End of file
`

	expected := []Entry{
		{Type: "alias", Name: "ll", Definition: "ls -la"},
		{Type: "alias", Name: "gs", Definition: "git status"},
		{Type: "alias", Name: "gd", Definition: "git diff"},
		{Type: "abbr", Name: "gco", Definition: "git checkout"},
		{Type: "abbr", Name: "gp", Definition: "git push"},
		{Type: "abbr", Name: "gl", Definition: "git pull"},
		{Type: "abbr", Name: "gcm", Definition: "git commit -m"},
	}

	entries, err := parser.Parse(content)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(entries) != len(expected) {
		t.Fatalf("Expected %d entries, got %d", len(expected), len(entries))
	}

	for i, entry := range entries {
		if entry.Type != expected[i].Type {
			t.Errorf("Entry %d: expected type %q, got %q", i, expected[i].Type, entry.Type)
		}
		if entry.Name != expected[i].Name {
			t.Errorf("Entry %d: expected name %q, got %q", i, expected[i].Name, entry.Name)
		}
		if entry.Definition != expected[i].Definition {
			t.Errorf("Entry %d: expected definition %q, got %q", i, expected[i].Definition, entry.Definition)
		}
	}
}
