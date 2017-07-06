package rsec16

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoderGenerateParity(t *testing.T) {
	data := [][]uint16{
		{0x1, 0x2},
		{0x3, 0x4},
		{0x5, 0x6},
		{0x7, 0x8},
		{0x9, 0xa},
	}
	c := NewCoder(5, 3)
	parity := c.GenerateParity(data)
	require.Equal(t, 3, len(parity))
	for _, row := range parity {
		require.Equal(t, 2, len(row))
	}
}

func TestCoderReconstructData(t *testing.T) {
	data := [][]uint16{
		{0x1, 0x2},
		{0x3, 0x4},
		{0x5, 0x6},
		{0x7, 0x8},
		{0x9, 0xa},
	}
	c := NewCoder(5, 3)
	parity := c.GenerateParity(data)

	corruptedData := [][]uint16{
		nil,
		data[1],
		nil,
		data[3],
		nil,
	}
	err := c.ReconstructData(corruptedData, parity)
	require.NoError(t, err)
	require.Equal(t, data, corruptedData)
}

func TestCoderReconstructDataMissingParity(t *testing.T) {
	data := [][]uint16{
		{0x1, 0x2},
		{0x3, 0x4},
		{0x5, 0x6},
		{0x7, 0x8},
		{0x9, 0xa},
	}
	c := NewCoder(5, 3)
	parity := c.GenerateParity(data)

	corruptedData := [][]uint16{
		data[0],
		nil,
		data[2],
		data[3],
		nil,
	}
	corruptedParity := [][]uint16{
		nil,
		parity[1],
		parity[2],
	}

	err := c.ReconstructData(corruptedData, corruptedParity)
	require.NoError(t, err)
	require.Equal(t, data, corruptedData)
}

func TestCoderReconstructDataNotEnough(t *testing.T) {
	data := [][]uint16{
		{0x1, 0x2},
		{0x3, 0x4},
		{0x5, 0x6},
		{0x7, 0x8},
		{0x9, 0xa},
	}
	c := NewCoder(5, 3)
	parity := c.GenerateParity(data)

	corruptedData := [][]uint16{
		data[0],
		nil,
		nil,
		nil,
		nil,
	}
	expectedErr := errors.New("not enough parity shards")
	err := c.ReconstructData(corruptedData, parity)
	require.Equal(t, expectedErr, err)

	corruptedData = [][]uint16{
		data[0],
		data[1],
		nil,
		nil,
		nil,
	}
	corruptedParity := [][]uint16{
		nil,
		parity[1],
		parity[2],
	}
	err = c.ReconstructData(corruptedData, corruptedParity)
	require.Equal(t, expectedErr, err)
}
