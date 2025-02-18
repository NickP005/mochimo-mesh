package indexer

import (
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/sigurn/crc16"
)

func convertColors(s string) string {
	// Minecraft color codes to ANSI escape sequences
	s = strings.ReplaceAll(s, "§0", "\x1b[30m") // black
	s = strings.ReplaceAll(s, "§1", "\x1b[34m") // dark blue
	s = strings.ReplaceAll(s, "§2", "\x1b[32m") // dark green
	s = strings.ReplaceAll(s, "§3", "\x1b[36m") // dark aqua
	s = strings.ReplaceAll(s, "§4", "\x1b[31m") // dark red
	s = strings.ReplaceAll(s, "§5", "\x1b[35m") // dark purple
	s = strings.ReplaceAll(s, "§6", "\x1b[33m") // gold
	s = strings.ReplaceAll(s, "§7", "\x1b[37m") // gray
	s = strings.ReplaceAll(s, "§8", "\x1b[90m") // dark gray
	s = strings.ReplaceAll(s, "§9", "\x1b[94m") // blue
	s = strings.ReplaceAll(s, "§a", "\x1b[92m") // green
	s = strings.ReplaceAll(s, "§b", "\x1b[96m") // aqua
	s = strings.ReplaceAll(s, "§c", "\x1b[91m") // red
	s = strings.ReplaceAll(s, "§d", "\x1b[95m") // light purple
	s = strings.ReplaceAll(s, "§e", "\x1b[93m") // yellow
	s = strings.ReplaceAll(s, "§f", "\x1b[97m") // white
	s = strings.ReplaceAll(s, "§r", "\x1b[0m")  // reset

	// Also support & prefix for compatibility
	// s = strings.ReplaceAll(s, "&", "§")

	return s
}

// Logger with colors, timestamps and log levels
func mlog(level int, format string, a ...interface{}) {
	if level > int(GLOBALS_LOG_LEVEL) {
		return
	}
	format = convertColors(format + "§r")
	fmt.Printf("\x1b[90m[%s]\x1b[0m ", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf(format, a...)
	fmt.Println()
}

func AddrTagToBase58(tag []byte) (string, error) {
	if len(tag) != 20 {
		return "", fmt.Errorf("invalid address tag length")
	}

	combined := make([]byte, 22)
	copy(combined, tag)

	// Calculate CRC using XMODEM
	table := crc16.MakeTable(crc16.CRC16_XMODEM)
	crc := crc16.Checksum(tag, table)

	// Append in little-endian
	combined[20] = byte(crc & 0xFF)
	combined[21] = byte((crc >> 8) & 0xFF)

	return base58.Encode(combined), nil
}
