package drvconfig

// Docker labels
const (
	LabelName = "io.homedrive.name"
)

// NewNameLabel returns a name label.
func NewNameLabel(name string) map[string]string {
	return map[string]string{LabelName: name}
}
