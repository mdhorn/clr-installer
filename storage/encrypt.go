// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package storage

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/clearlinux/clr-installer/cmd"
	"github.com/clearlinux/clr-installer/errors"
	"github.com/clearlinux/clr-installer/log"
)

const (
	// MinPassphraseLength is the shortest possible password
	MinPassphraseLength = 8
	// MaxPassphraseLength is the shortest possible password
	MaxPassphraseLength = 94

	// RequiredBundle the bundle needed if encrypted partitions are used
	RequiredBundle = "bootloader-extras"
)

// EncryptionEnabled checks all partition to see if encryption was enabled
func (bd *BlockDevice) EncryptionEnabled() bool {
	enabled := bd.Type == BlockDeviceTypeCrypt

	for _, ch := range bd.Children {
		if len(ch.Children) > 0 {
			enabled = enabled || ch.EncryptionEnabled()
		} else {
			enabled = enabled || ch.Type == BlockDeviceTypeCrypt
		}
	}

	return enabled
}

// MapEncrypted uses cryptsetup to format (initialize) and open (map) the
// physical partion to an encrypted partition
func (bd *BlockDevice) MapEncrypted(passphrase string) error {
	if bd.Type != BlockDeviceTypeCrypt {
		return errors.Errorf("Trying to run cryptsetup() against a non crypt partition")
	}

	args := []string{
		"cryptsetup",
		"--batch-mode",
		"--hash=sha256",
		"--cipher=aes-xts-plain64",
		"--key-size=512",
		"luksFormat",
	}

	args = append(args, bd.GetDeviceFile(), "-")

	if err := cmd.PipeRunAndLog(passphrase, args...); err != nil {
		return errors.Wrap(err)
	}

	var mapped string

	// Special case for mapping 'root'
	if bd.MountPoint == "/" {
		mapped = "root"
	} else {
		// make the mapped device all lower case
		// drop the leading '/'
		mapped = strings.TrimPrefix(strings.ToLower(bd.MountPoint), "/")
		// replace '/' with '_'
		mapped = strings.Replace(mapped, "/", "_", -1)
	}

	args = []string{
		"cryptsetup",
		"--batch-mode",
		"luksOpen",
	}

	args = append(args, bd.GetDeviceFile(), mapped, "-")

	if err := cmd.PipeRunAndLog(passphrase, args...); err != nil {
		return errors.Wrap(err)
	}

	log.Debug("Disk partition %q is mapped to encrypted partition %q", bd.Name, mapped)

	// Store the mapped point for later unmounting
	mountedEncrypts = append(mountedEncrypts, mapped)

	bd.MappedName = filepath.Join("mapper", mapped)

	return nil
}

// umMapEncrypted uses cryptsetup to close (unmap) an encrypted partition
func unMapEncrypted(mapped string) error {
	args := []string{
		"cryptsetup",
		"--batch-mode",
		"luksClose",
		mapped,
	}

	if err := cmd.RunAndLog(args...); err != nil {
		return errors.Wrap(err)
	}

	return nil
}

// IsValidPassphrase checks the minimum passphrase requirements
func IsValidPassphrase(phrase string) (bool, string) {
	if phrase == "" {
		return false, "Passphrase is required"
	}

	if !isPrintable(phrase) {
		return false, "Passphrase may only contain 7-bit, printable characters"
	}

	if len(phrase) < MinPassphraseLength {
		return false, fmt.Sprintf("Passphrase must be at least %d characters long",
			MinPassphraseLength)
	}

	if len(phrase) > MaxPassphraseLength {
		return false, fmt.Sprintf("Passphrase may be at most %d characters long",
			MaxPassphraseLength)
	}

	return true, ""
}

// AskPassPhrase prompts to the user interactively for the pass phrase
// via the command line.
// This is intended to be used to get a pass phrase for encrypting
// file systems on the installation target while using the command
// line (aka massinstall)
func AskPassPhrase() string {
	passphrase := ""
	done := false

	// Get the initial state of the terminal.
	initialTermState, termErr := terminal.GetState(syscall.Stdin)
	if termErr != nil {
		log.Warning("Unable to get terminal state for recovery: %v", termErr)
	}

	// Restore it in the event of an interrupt.
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP,
		syscall.SIGABRT, syscall.SIGSTKFLT, syscall.SIGSYS)

	go func() {
		<-c
		_ = terminal.Restore(syscall.Stdin, initialTermState)
		signal.Stop(c)
	}()

	for !done {
		fmt.Print("Disk Encryption Passphrase: ")
		bytePassphrase, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Print("\n")
		if err == nil {
			passphrase = string(bytePassphrase)
			strings.TrimSpace(passphrase)

			errMsg := ""
			if done, errMsg = IsValidPassphrase(passphrase); !done {
				fmt.Println(errMsg)
			}
		} else {
			done = true
			fmt.Print("Error getting passphrase: %v", err)
			passphrase = ""
		}
	}

	signal.Stop(c)

	return passphrase
}

func isPrintable(s string) bool {
	for _, c := range s {
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}
