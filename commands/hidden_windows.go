package commands

import "syscall"

// isHiddenAttr reports whether path has the Windows "hidden" file
// attribute set. "ls" treats this the same as a leading-dot name (the
// Linux hidden-file convention), so "-a" is needed to see either kind.
func isHiddenAttr(path string) (bool, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}
	attrs, err := syscall.GetFileAttributes(p)
	if err != nil {
		return false, err
	}
	return attrs&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
