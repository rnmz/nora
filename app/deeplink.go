package app

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

func ExecuteDeepLink(deeplink string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.CommandContext(ctx, "xdg-open", deeplink)
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", deeplink)
	case "windows":
		cmd = exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", deeplink)
	default:
		return fmt.Errorf("Unsupported OS: %s", runtime.GOOS)
	}

	err := cmd.Run()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("deeplink %q timeout exceeded: process killed", deeplink)
	}

	return err
}
