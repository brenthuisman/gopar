package gf2p16

import (
	"testing"

	"github.com/klauspost/cpuid"
	"github.com/stretchr/testify/require"
)

func TestStandardToAltMapSSSE3Unsafe(t *testing.T) {
	if !cpuid.CPU.SSSE3() {
		t.Skip("SSSE3 not supported; skipping")
	}

	in0 := [16]byte{
		0x20, 0x21, 0x30, 0x31,
		0x40, 0x41, 0x50, 0x51,
		0x60, 0x61, 0x70, 0x71,
		0x80, 0x81, 0x90, 0x91,
	}

	in1 := [16]byte{
		0xa0, 0xa1, 0xb0, 0xb1,
		0xc0, 0xc1, 0xd0, 0xd1,
		0xe0, 0xe1, 0xf0, 0xf1,
		0x00, 0x01, 0x10, 0x11,
	}

	filler := [16]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
	}
	outLow := [2][16]byte{filler, filler}
	outHigh := [2][16]byte{filler, filler}

	expectedOutLow := [2][16]byte{{
		0x20, 0x30, 0x40, 0x50,
		0x60, 0x70, 0x80, 0x90,
		0xa0, 0xb0, 0xc0, 0xd0,
		0xe0, 0xf0, 0x00, 0x10,
	}, filler}

	expectedOutHigh := [2][16]byte{{
		0x21, 0x31, 0x41, 0x51,
		0x61, 0x71, 0x81, 0x91,
		0xa1, 0xb1, 0xc1, 0xd1,
		0xe1, 0xf1, 0x01, 0x11,
	}, filler}

	standardToAltMapSSSE3Unsafe(&in0, &in1, &outLow[0], &outHigh[0])

	require.Equal(t, expectedOutLow, outLow)
	require.Equal(t, expectedOutHigh, outHigh)
}
