package renderer

type NaiveRenderer struct{}

func (r *NaiveRenderer) Render(content string) string {
	return content
}
