package cli

import (
	"context"
	"io"
)

func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	streams := Streams{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
	if len(args) == 0 {
		return runMenu(ctx, streams)
	}

	cmd := NewRootCmd(ctx, Streams{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	})
	cmd.SetArgs(args)
	return cmd.Execute()
}
