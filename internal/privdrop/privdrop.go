// Package privdrop lets the container start as root, hand the data dir to
// PUID:PGID and then drop to that user (linuxserver.io style), so any mounted
// folder works without manual chown.
package privdrop

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"syscall"
)

const defaultID = 1000

// Sys is the seam over the process-identity syscalls so tests need no root.
type Sys struct {
	Geteuid   func() int
	Getegid   func() int
	Lchown    func(path string, uid, gid int) error
	Setgroups func(gids []int) error
	Setgid    func(gid int) error
	Setuid    func(uid int) error
}

// OS returns the real syscalls. On Linux, Go (>=1.16) applies Setuid/Setgid/
// Setgroups to every thread of the process.
func OS() Sys {
	return Sys{
		Geteuid:   os.Geteuid,
		Getegid:   os.Getegid,
		Lchown:    os.Lchown,
		Setgroups: syscall.Setgroups,
		Setgid:    syscall.Setgid,
		Setuid:    syscall.Setuid,
	}
}

// Prepare makes dataDir usable before the server opens the DB. As root it
// chowns dataDir to PUID:PGID and drops to that user; otherwise it only checks
// that dataDir is writable. It never lets the server continue as root.
func Prepare(sys Sys, log *slog.Logger, dataDir, puid, pgid string) error {
	if sys.Geteuid() == 0 {
		uid, gid, err := ParseIDs(puid, pgid)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dataDir, 0o750); err != nil {
			return fmt.Errorf("create data dir: %w", err)
		}
		n, err := ChownTree(dataDir, uid, gid, sys.Lchown)
		if err != nil {
			return fmt.Errorf("chown data dir to %d:%d: %w", uid, gid, err)
		}
		if err := sys.Setgroups([]int{}); err != nil {
			return fmt.Errorf("setgroups: %w", err)
		}
		if err := sys.Setgid(gid); err != nil {
			return fmt.Errorf("setgid %d: %w", gid, err)
		}
		if err := sys.Setuid(uid); err != nil {
			return fmt.Errorf("setuid %d: %w", uid, err)
		}
		if got := sys.Geteuid(); got != uid {
			return fmt.Errorf("privilege drop failed: euid is %d, want %d", got, uid)
		}
		log.Info("dropped privileges", "uid", uid, "gid", gid, "data_dir", dataDir, "chowned", n)
	} else if err := CheckWritable(dataDir, sys.Geteuid(), sys.Getegid()); err != nil {
		return err
	}
	if sys.Geteuid() == 0 {
		return errors.New("refusing to run the server as root: set PUID/PGID to a non-root user")
	}
	return nil
}

// ParseIDs reads PUID/PGID values; empty means 1000.
func ParseIDs(puid, pgid string) (uid, gid int, err error) {
	if uid, err = parseID("PUID", puid); err != nil {
		return 0, 0, err
	}
	if gid, err = parseID("PGID", pgid); err != nil {
		return 0, 0, err
	}
	return uid, gid, nil
}

func parseID(name, v string) (int, error) {
	if v == "" {
		return defaultID, nil
	}
	id, err := strconv.Atoi(v)
	if err != nil || id < 0 {
		return 0, fmt.Errorf("invalid %s %q: must be an integer >= 0", name, v)
	}
	return id, nil
}

// ChownTree sets uid:gid on root and everything below it, skipping entries that
// already match. Children are chowned before their parent so a directory stays
// traversable (as its old owner) until its contents are done. Returns how many
// entries were changed.
func ChownTree(root string, uid, gid int, lchown func(string, int, int) error) (int, error) {
	var todo []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if st, ok := info.Sys().(*syscall.Stat_t); ok && int(st.Uid) == uid && int(st.Gid) == gid {
			return nil
		}
		todo = append(todo, path)
		return nil
	})
	if err != nil {
		return 0, err
	}
	// WalkDir is pre-order, so reversed every child precedes its parent.
	slices.Reverse(todo)
	for _, path := range todo {
		if err := lchown(path, uid, gid); err != nil {
			return 0, err
		}
	}
	return len(todo), nil
}

// CheckWritable creates and removes a probe file in dir (creating dir if needed).
func CheckWritable(dir string, uid, gid int) error {
	err := os.MkdirAll(dir, 0o750)
	if err == nil {
		var f *os.File
		if f, err = os.CreateTemp(dir, ".write-probe-*"); err == nil {
			_ = f.Close()
			err = os.Remove(f.Name())
		}
	}
	if err != nil {
		return fmt.Errorf("data dir %s is not writable by uid %d gid %d — fix ownership or run the container without `user:` so it can fix it itself: %w", dir, uid, gid, err)
	}
	return nil
}
