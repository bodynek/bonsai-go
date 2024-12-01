package tests

import (
	"fmt"
	"os/exec"
	"time"
)

type BonsaiInstance struct {
	cmd *exec.Cmd
}

// StartBonsai starts a bonsaid instance on test ports
func StartBonsai(redisHost string, redisPort string, apiPort string, svcPort string) (*BonsaiInstance, error) {
	cmd := exec.Command(
		"bonsaid",
		"--config=/dev/null",
		fmt.Sprintf("--api-port=%s", apiPort),
		fmt.Sprintf("--svc-port=%s", svcPort),
	)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	time.Sleep(1 * time.Second) // Give time for the server to initialize
	return &BonsaiInstance{cmd: cmd}, nil
}

// Stop stops the bonsaid instance
func (b *BonsaiInstance) Stop() error {
	if b.cmd != nil {
		return b.cmd.Process.Kill()
	}
	return nil
}
