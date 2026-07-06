package browser

// entry is a single entry in a collection of contents.
type entry struct {
	Name string
	Path string
	// Type is the file type. Either "dir" or "file".
	Type string
}
