// Package pageshelf exposes the bundled Pageshelf agent skill.
package pageshelf

import _ "embed"

//go:embed SKILL.md
var bundledSkill []byte

// Content returns the bundled Pageshelf skill markdown.
func Content() []byte {
	return append([]byte(nil), bundledSkill...)
}
