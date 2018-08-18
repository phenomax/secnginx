package util

import("os/exec"
"bufio"
"log"
"os")

// RunAndPrintCommandOutput calls the given command and prints the output - line by line - into the console
func RunAndPrintCommandOutput(cmd *exec.Cmd) {
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Error while creating a Stdout-Pipe: %s", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			log.Printf("%s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		log.Fatalf("Error starting Cmd: %s", err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("Error waiting for Cmd: %s", err)
	}
}
