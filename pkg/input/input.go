package input

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

// StdinMode define the behaviour of stding when entering response.
// Confirm: Enter must be pressed to submit response.
// NoConfirm: response is directly submitted, aka 1 key press.
type StdinMode bool

const (
	Confirm   StdinMode = true
	NoConfirm StdinMode = false

	CurrentMode = NoConfirm
)

// IsResponseYes return true when the user press 'y'.
func IsResponseYes(mode StdinMode) (bool, error) {
	if mode == NoConfirm {
		// Terminal need to be set to raw mode for NoConfirm mode.
		fd := int(os.Stdin.Fd())
		state, err := terminal.MakeRaw(fd)
		if err != nil {
			return false, err
		}
		defer terminal.Restore(fd, state)
	}

	reader := bufio.NewReader(os.Stdin)
	c, err := reader.ReadByte()
	if err != nil {
		return false, err
	}

	if mode == NoConfirm {
		// Printing back the pressed character.
		// Raw mode do not print back the pressed character.
		fmt.Printf("%c\r\n", c)
	}

	return c == 'y', nil
}
