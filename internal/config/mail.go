package config

import (
	"fmt"
	"strings"
)

// Mail is outbound email. driver log is the local implementation (LogMailer).
// smtp is reserved until a writer ships.
type Mail struct {
	Driver string `json:"driver"`
}

func (m Mail) Validate() error {
	switch strings.ToLower(strings.TrimSpace(m.Driver)) {
	case "", "log":
		return nil
	case "smtp":
		return fmt.Errorf("mail driver %q is reserved until the SMTP writer ships", m.Driver)
	default:
		return fmt.Errorf("unknown mail driver %q (want log, smtp)", m.Driver)
	}
}
