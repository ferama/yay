package renderer

type Renderer interface {
	Render(content string) string
}
