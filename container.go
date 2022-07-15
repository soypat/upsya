package main

import (
	"os"
	"os/exec"
	"path/filepath"
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
	chdir = filepath.Join(c.workDir, chdir)
	args = append([]string{"-chrt", c.chrtDir, "-chdir", chdir,
		"-timeout", timeout.String(), command}, args...)
	return exec.Command("gontainer", args...)
}

// CreateFile creates a file. See os.Create.
func (c container) CreateFile(name string) (*os.File, error) {
	return os.Create(c.osPath(name))
}

// osPath returns the operating system path to absolute path
func (c container) osPath(containerPath string) string {
	if filepath.IsAbs(containerPath) {
		return filepath.Join(c.chrtDir, containerPath)
	}
	return filepath.Join(c.chrtDir, c.workDir, containerPath)
}

// Mkdir creates a folder. See os.Mkdir.
func (c container) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(c.osPath(name), perm)
}
