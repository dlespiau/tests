// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tests

import (
	"bytes"
	"flag"
	"os/exec"
	"syscall"
	"time"
)

// Runtime is the path of Clear Containers Runtime
var Runtime string

// Proxy is the path of Clear Containers Proxy
var Proxy string

// Shim is the path of Clear Containers Shim
var Shim string

// Timeout specifies the time limit in seconds for each test
var Timeout int

// Command contains the information of the command to run
type Command struct {
	// cmd exec.Cmd
	cmd *exec.Cmd

	// Stderr process's standard error
	Stderr bytes.Buffer

	// Stdout process's standard output
	Stdout bytes.Buffer

	// Timeout is the time limit of seconds of the command
	Timeout time.Duration

	// ExpectedExitCode is the expected exit code
	ExpectedExitCode int
}

func init() {
	flag.StringVar(&Runtime, "runtime", "cc-runtime", "Path of Clear Containers Runtime")
	flag.StringVar(&Proxy, "proxy", "cc-proxy", "Path of Clear Containers Proxy")
	flag.StringVar(&Shim, "shim", "cc-shim", "Path of Clear Containers Shim")
	flag.IntVar(&Timeout, "timeout", 5, "Time limit in seconds for each test")

	flag.Parse()
}

// NewCommand returns a new instance of Command
func NewCommand(path string, args ...string) *Command {
	c := new(Command)
	c.cmd = exec.Command(path, args...)
	c.cmd.Stderr = &c.Stderr
	c.cmd.Stdout = &c.Stdout
	c.ExpectedExitCode = 0
	c.Timeout = time.Duration(Timeout)

	return c
}

// Run runs a command returning its exit code
func (c *Command) Run() int {
	LogIfFail("Running command '%s %s'\n", c.cmd.Path, c.cmd.Args)
	if err := c.cmd.Start(); err != nil {
		LogIfFail("could no start command: %v\n", err)
	}

	done := make(chan error)
	go func() { done <- c.cmd.Wait() }()

	var timeout <-chan time.Time
	if c.Timeout > 0 {
		timeout = time.After(c.Timeout * time.Second)
	}

	select {
	case <-timeout:
		LogIfFail("Killing process timeout reached '%d' seconds\n", c.Timeout)
		_ = c.cmd.Process.Kill()
		return -1

	case err := <-done:
		if err != nil {
			LogIfFail("command failed error '%s'\n", err)
		}
		exitCode := c.cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
		if exitCode != c.ExpectedExitCode {
			LogIfFail("Exit code '%d' does not match with expected exit code '%d'\n", exitCode, c.ExpectedExitCode)
		}
		return exitCode
	}
}
