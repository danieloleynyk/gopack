package utils

import (
	"bufio"
	"fmt"
	"os/exec"
)

// RunCommand used for running a command and piping its output to stdout
func RunCommand(command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmdReader, err := cmd.StderrPipe()
	Catch(err, "An error occurred while setting up stderr pipeline", false)

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("\t gopack > %s\n", scanner.Text())
		}
	}()

	Catch(cmd.Start(), "An error occurred in command start", true)
	Catch(cmd.Wait(), "An error occurred in command wait", true)
}
