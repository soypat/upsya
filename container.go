package main

import (
	"os/exec"
	"path"
	"time"
)

type container struct {
	// Containerized file system or VFS path to chroot to.
	chrtDir string
	// Initial directory where all actions are performed
	// unless specified otherwise.
	workDir string
}

// Command executes a containerized command using gontainer. timeout=0 disables timeout.
//  c.Command(0, "", "python3", "theFile.py")
// The chdir argument is the path the container is run inside the containerized
// environemnt and is interpreted from the working directory of the container.
func (c container) Command(timeout time.Duration, chdir, command string, args ...string) *exec.Cmd {
	chdir = path.Join(c.workDir, chdir)
	args = append([]string{"--chrt", c.chrtDir, "--chdir", chdir,
		"--timeout", timeout.String(), command}, args...)
	return exec.Command("gontainer", args...)
}

func (c container) CreateFile()
