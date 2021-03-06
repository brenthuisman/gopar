package par1

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUTF16LEStringRoundTrip(t *testing.T) {
	for _, s := range []string{
		"",
		"Hello, world",
		"Hello, 世界",
		"Hello\000world",
	} {
		encodedS := encodeUTF16LEString(s)
		roundTripS := decodeUTF16LEString(encodedS)
		require.Equal(t, s, roundTripS)
	}
}

func TestFileEntryRoundTrip(t *testing.T) {
	filename := "filename世界.r01"
	entry := fileEntry{
		header: fileEntryHeader{
			Status:       10,
			FileBytes:    10,
			Hash:         [16]byte{0x1, 0x2},
			SixteenKHash: [16]byte{0x3, 0x4},
		},
		filename: filename,
	}

	entryBytes, err := writeFileEntry(entry)
	require.NoError(t, err)

	roundTripEntry, err := readFileEntry(bytes.NewBuffer(entryBytes))
	require.NoError(t, err)

	entry.header.EntryBytes = uint64(len(entryBytes))
	require.Equal(t, entry, roundTripEntry)
}
