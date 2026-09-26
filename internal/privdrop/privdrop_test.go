package privdrop

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestParseIDs(t *testing.T) {
	tests := []struct {
		name       string
		puid, pgid string
		wantUID    int
		wantGID    int
		wantErr    string
	}{
		{"defaults", "", "", 1000, 1000, ""},
		{"explicit", "1234", "5678", 1234, 5678, ""},
		{"zero allowed", "0", "0", 0, 0, ""},
		{"only uid", "1234", "", 1234, 1000, ""},
		{"negative uid", "-1", "", 0, 0, "PUID"},
		{"non-numeric gid", "", "users", 0, 0, "PGID"},
		{"decimal", "1.5", "", 0, 0, "PUID"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid, gid, err := ParseIDs(tt.puid, tt.pgid)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want mention of %s", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if uid != tt.wantUID || gid != tt.wantGID {
				t.Errorf("got %d:%d, want %d:%d", uid, gid, tt.wantUID, tt.wantGID)
			}
		})
	}
}

// tree creates root/{a.db, backups/, backups/b.db} and returns root.
func tree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "a.db"))
	if err := os.Mkdir(filepath.Join(root, "backups"), 0o750); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(root, "backups", "b.db"))
	return root
}

func mustWrite(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, nil, 0o640); err != nil {
		t.Fatal(err)
	}
}

func TestChownTreeSkipsMatchingOwner(t *testing.T) {
	root := tree(t)
	var calls []string
	n, err := ChownTree(root, os.Geteuid(), os.Getegid(), func(p string, _, _ int) error {
		calls = append(calls, p)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 || len(calls) != 0 {
		t.Errorf("chowned %d (%v), want none", n, calls)
	}
}

func TestChownTreeChildrenBeforeParents(t *testing.T) {
	root := tree(t)
	var calls []string
	n, err := ChownTree(root, os.Geteuid()+1, os.Getegid(), func(p string, _, _ int) error {
		calls = append(calls, p)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 || len(calls) != 4 {
		t.Fatalf("chowned %d: %v, want 4 entries", n, calls)
	}
	pos := func(p string) int { return slices.Index(calls, p) }
	backups := filepath.Join(root, "backups")
	if !(pos(filepath.Join(backups, "b.db")) < pos(backups) && pos(backups) < pos(root) && pos(filepath.Join(root, "a.db")) < pos(root)) {
		t.Errorf("parent chowned before child: %v", calls)
	}
}

func TestChownTreeStopsOnError(t *testing.T) {
	boom := errors.New("EPERM")
	_, err := ChownTree(tree(t), os.Geteuid()+1, os.Getegid(), func(string, int, int) error { return boom })
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want %v", err, boom)
	}
}

func TestCheckWritable(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing", "data")
	if err := CheckWritable(dir, 1, 2); err != nil {
		t.Fatalf("writable (created) dir: %v", err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("probe left behind: %v", entries)
	}

	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	ro := t.TempDir()
	if err := os.Chmod(ro, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(ro, 0o700) })
	err := CheckWritable(ro, 1234, 5678)
	if err == nil {
		t.Fatal("read-only dir: want error")
	}
	for _, want := range []string{ro, "not writable by uid 1234 gid 5678", "without `user:`"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q missing %q", err, want)
		}
	}
}

// fakeSys records calls and tracks the effective uid like the kernel would.
type fakeSys struct {
	euid, egid int
	calls      []string
	chowned    []string
	setuidErr  error
	ignoreDrop bool // Setuid "succeeds" but euid stays put
}

func (f *fakeSys) sys() Sys {
	return Sys{
		Geteuid: func() int { return f.euid },
		Getegid: func() int { return f.egid },
		Lchown: func(p string, _, _ int) error {
			f.chowned = append(f.chowned, p)
			return nil
		},
		Setgroups: func([]int) error { f.calls = append(f.calls, "setgroups"); return nil },
		Setgid:    func(int) error { f.calls = append(f.calls, "setgid"); return nil },
		Setuid: func(uid int) error {
			f.calls = append(f.calls, "setuid")
			if f.setuidErr != nil {
				return f.setuidErr
			}
			if !f.ignoreDrop {
				f.euid = uid
			}
			return nil
		},
	}
}

var discard = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestPrepareAsRootChownsAndDrops(t *testing.T) {
	root := tree(t)
	f := &fakeSys{euid: 0}
	// A uid that differs from the test runner's so every entry needs a chown.
	uid := os.Geteuid() + 1
	if err := Prepare(f.sys(), discard, root, strconv.Itoa(uid), ""); err != nil {
		t.Fatal(err)
	}
	if f.euid != uid {
		t.Errorf("euid = %d, want %d", f.euid, uid)
	}
	if want := []string{"setgroups", "setgid", "setuid"}; !slices.Equal(f.calls, want) {
		t.Errorf("calls = %v, want %v", f.calls, want)
	}
	if len(f.chowned) != 4 {
		t.Errorf("chowned %v, want 4 entries", f.chowned)
	}
}

func TestPrepareAsRootCreatesMissingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "new")
	f := &fakeSys{euid: 0}
	if err := Prepare(f.sys(), discard, dir, "1234", "1234"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("data dir not created: %v", err)
	}
}

func TestPrepareAsRootFailures(t *testing.T) {
	tests := []struct {
		name       string
		f          *fakeSys
		puid, pgid string
		wantErr    string
	}{
		{"bad PUID", &fakeSys{}, "x", "", "PUID"},
		{"setuid fails", &fakeSys{setuidErr: errors.New("EPERM")}, "", "", "setuid 1000"},
		{"drop silently ignored", &fakeSys{ignoreDrop: true}, "", "", "privilege drop failed"},
		{"PUID 0 stays root", &fakeSys{}, "0", "0", "refusing to run the server as root"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Prepare(tt.f.sys(), discard, t.TempDir(), tt.puid, tt.pgid)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("err = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestPrepareNonRootOnlyChecksWritable(t *testing.T) {
	root := tree(t)
	f := &fakeSys{euid: 1000, egid: 1000}
	if err := Prepare(f.sys(), discard, root, "1234", "1234"); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 0 || len(f.chowned) != 0 {
		t.Errorf("non-root must not chown/drop: calls=%v chowned=%v", f.calls, f.chowned)
	}
}
