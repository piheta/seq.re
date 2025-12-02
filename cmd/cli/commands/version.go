package commands

import (
	"fmt"
	"os"
)

// Version displays version information
func Version(version, commit, date string) {
	_, _ = fmt.Fprintf(os.Stdout, "seqre version %s\n", version)
	_, _ = fmt.Fprintf(os.Stdout, "commit: %s\n", commit)
	_, _ = fmt.Fprintf(os.Stdout, "built: %s\n", date)
}
