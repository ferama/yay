package renderer

import "github.com/charmbracelet/glamour"

type MarkdownRenderer struct {
	renderer *glamour.TermRenderer
}

func NewMarkdownRenderer() *MarkdownRenderer {
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithPreservedNewLines(),
	)
	mr := &MarkdownRenderer{
		renderer: renderer,
	}

	return mr
}
func (r *MarkdownRenderer) Render(content string) string {
	out, _ := r.renderer.Render(content)
	return out
}
