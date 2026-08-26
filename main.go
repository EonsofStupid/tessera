package main

import (
	"context"
	"os"

	"github.com/shippinAI/nomen/backend/v3/instrumentation/logging"
	"github.com/shippinAI/nomen/cmd"
)

func main() {
	args := os.Args[1:]
	rootCmd := cmd.New(os.Stdout, os.Stdin, args, nil)
	ctx := logging.NewCtx(context.Background(), logging.StreamRuntime)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		// error is logged by the command itself
		os.Exit(1)
	}
}
