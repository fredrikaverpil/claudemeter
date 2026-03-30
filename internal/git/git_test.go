package git

import (
	"os"
	"os/exec"
	"testing"
)

func TestBranch(t *testing.T) {
	tmp := t.TempDir()

	// Save and restore working directory.
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	// Initialize a real git repo so .git/HEAD is created by git itself.
	run := func(args ...string) {
		t.Helper()
		cmd := exec.CommandContext(t.Context(), "git", args...)
		cmd.Dir = tmp
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "main")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	t.Run("default branch", func(t *testing.T) {
		got := Branch()
		if got != "main" {
			t.Errorf("Branch() = %q, want %q", got, "main")
		}
	})

	t.Run("branch with slashes", func(t *testing.T) {
		run("switch", "-c", "feat/my-feature")
		got := Branch()
		if got != "feat/my-feature" {
			t.Errorf("Branch() = %q, want %q", got, "feat/my-feature")
		}
	})

	t.Run("detached HEAD", func(t *testing.T) {
		// Need a commit to detach from.
		run("commit", "--allow-empty", "-m", "init")
		run("switch", "--detach")
		got := Branch()
		if got != "" {
			t.Errorf("Branch() = %q, want empty string", got)
		}
	})

	t.Run("no git directory", func(t *testing.T) {
		noGit := t.TempDir()
		if err := os.Chdir(noGit); err != nil {
			t.Fatal(err)
		}
		// Chdir back before TempDir cleanup — Windows can't remove the cwd.
		t.Cleanup(func() { _ = os.Chdir(orig) })
		got := Branch()
		if got != "" {
			t.Errorf("Branch() = %q, want empty string", got)
		}
	})
}
