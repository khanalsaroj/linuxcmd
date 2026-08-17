package output

import "fmt"

// HumanSize renders a byte count the way "ls -h" would: the largest unit
// that keeps the number under 1024, with one decimal place for anything
// smaller than 10 in that unit.
func HumanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	units := "KMGTPE"
	val := float64(n) / float64(div)
	suffix := string(units[exp]) // "K","M","G",...
	if val < 10 {
		return fmt.Sprintf("%.1f%s", val, suffix)
	}
	return fmt.Sprintf("%.0f%s", val, suffix)
}
