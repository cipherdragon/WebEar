package executor

import (
	"errors"
	"fmt"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"webear/executor/reaper"
)

func ExecuteScript(payload string, name string, scriptPath string, username string) error {
	reaper.WakeUpReaper()

	if username == "" {
		err := fmt.Sprintf("User must be provided to execute the script [%s]", scriptPath)
		return errors.New(err)
	}

	targetUser, err := user.Lookup(username)
	if err != nil {
		err := fmt.Sprintf("[%s] Could not resolve user [%s]: %v", name, username, err)
		return errors.New(err)
	}

	uid, err := strconv.ParseUint(targetUser.Uid, 10, 32)
	if err != nil {
		err := fmt.Sprintf("[%s] Could not parse user id [%s]: %v", name, targetUser.Uid, err)
		return errors.New(err)
	}

	gid, err := strconv.ParseUint(targetUser.Gid, 10, 32)
	if err != nil {
		err := fmt.Sprintf("[%s] Could not parse group id [%s]: %v", name, targetUser.Gid, err)
		return errors.New(err)
	}

	env := []string{
		fmt.Sprintf("WEBEAR_DATA=%s", payload),
		fmt.Sprintf("WEBEAR_NAME=%s", name),
		fmt.Sprintf("HOME=%s", targetUser.HomeDir),
		fmt.Sprintf("USER=%s", targetUser.Username),
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
	}

	attr := &syscall.ProcAttr{
		Dir: filepath.Dir(scriptPath),
		Env: env,
		Files: []uintptr{uintptr(0), 1, 2}, // stdin -> nil
		Sys: &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: uint32(uid),
				Gid: uint32(gid),
			},
		},
	}

	pid, err := syscall.ForkExec("/bin/sh", []string{"/bin/sh", scriptPath}, attr)
	if err != nil {
		err := fmt.Sprintf("Error executing the script [%s]: %v", scriptPath, err)
		return errors.New(err)
	}

	reaper.RecordToReap(pid)

	return nil
}