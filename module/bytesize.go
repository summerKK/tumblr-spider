package module

import "fmt"

const (
	_ = iota
	//2的10次方 1kb = 1024byte
	KB float64 = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
)

func ByteSize(i uint64) string {
	b := float64(i)
	switch {
	case b >= EB:
		return fmt.Sprintf("%.2fEb", b/EB)
	case b >= PB:
		return fmt.Sprintf("%.2fPB", b/PB)
	case b >= TB:
		return fmt.Sprintf("%.2fTB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%.2fGb", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.2fMB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.2fKb", b/KB)
	}
	return fmt.Sprintf("%dB", i)
}
