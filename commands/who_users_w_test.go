package commands

import "testing"

// The Terminal Services session APIs behind who/users/w may be
// unavailable or access-denied in a locked-down CI/sandbox account,
// so these tests only require that the command runs to completion
// without hanging, not that it succeeds.
func TestWhoRuns(t *testing.T) {
	run(t, "who")
}

func TestUsersRuns(t *testing.T) {
	run(t, "users")
}

func TestWRuns(t *testing.T) {
	run(t, "w")
}
