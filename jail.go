package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type jail interface {
	Command(ctx context.Context, gracefulTimeout time.Duration, chdir, command string, args ...string) *exec.Cmd
	CreateFile(name string) (*os.File, error)
	Mkdir(name string, perm os.FileMode) error
	MkdirAll(name string, perm os.FileMode) error
	Remove(name string) error
	RemoveAll(name string) error
}

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
func (c container) Command(ctx context.Context, gracefulTimeout time.Duration, chdir, command string, args ...string) *exec.Cmd {
	chdir = filepath.Join(c.workDir, chdir)
	args = append([]string{"-chrt", c.chrtDir, "-chdir", chdir,
		"-timeout", gracefulTimeout.String(), command}, args...)
	return exec.CommandContext(ctx, "gontainer", args...)
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

// Mkdir creates a folder. See os.Mkdir.
func (c container) MkdirAll(name string, perm os.FileMode) error {
	return os.MkdirAll(c.osPath(name), perm)
}

func (c container) Remove(name string) error {
	return os.Remove(c.osPath(name))
}

func (c container) RemoveAll(name string) error {
	return os.RemoveAll(c.osPath(name))
}

type systemPython struct {
	path string
}

var _ jail = systemPython{}

func (s systemPython) Command(ctx context.Context, _ time.Duration, chdir, command string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = filepath.Join(s.path, chdir)
	return cmd
}

// CreateFile creates a file. See os.Create.
func (s systemPython) CreateFile(name string) (*os.File, error) {
	return os.Create(s.osPath(name))
}

// osPath returns the operating system path to absolute path
func (s systemPython) osPath(relPath string) string {
	return filepath.Join(s.path, relPath)
}

// Mkdir creates a folder. See os.Mkdir.
func (s systemPython) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(s.osPath(name), perm)
}

// Mkdir creates a folder. See os.Mkdir.
func (s systemPython) MkdirAll(name string, perm os.FileMode) error {
	return os.MkdirAll(s.osPath(name), perm)
}

func (s systemPython) Remove(name string) error {
	return os.Remove(s.osPath(name))
}

func (s systemPython) RemoveAll(name string) error {
	return os.RemoveAll(s.osPath(name))
}
