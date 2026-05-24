package drvconfig

// SystemInfo contains the high-level information of the board and the base
// OS image.
type SystemInfo struct {
	Board string `json:",omitempty"`
}

// Common boards.
const (
	BoardRpi4         = "rpi4"
	BoardNUC7         = "nuc7"
	BoardNUC10        = "nuc10"
	BoardDigitalOcean = "docn"
)
