package par1

import (
	"bytes"
	"crypto/md5"
	"errors"
	"io/ioutil"
)

// A volume contains information about the volume set, and a data
// payload. For the index volume, h.VolumeNumber is 0 and data
// contains a comment for the set. For parity volumes, h.VolumeNumber
// is greater than 0, and data contains the parity data for that
// volume. All other data should be the same for all volumes in a set
// (identified by h.SetHash).
type volume struct {
	header  header
	entries []fileEntry
	data    []byte
}

const controlHashOffset = 0x20

func readVolume(volumeBytes []byte) (volume, error) {
	buf := bytes.NewBuffer(volumeBytes)

	header, err := readHeader(buf)
	if err != nil {
		return volume{}, err
	}

	controlHash := md5.Sum(volumeBytes[controlHashOffset:])
	if controlHash != header.ControlHash {
		return volume{}, errors.New("invalid control hash")
	}

	// TODO: Check h.SetHash.

	// TODO: Check count of files saved in volume set, and other
	// offsets and bytes.

	entries := make([]fileEntry, header.FileCount)
	for i := uint64(0); i < header.FileCount; i++ {
		var err error
		entries[i], err = readFileEntry(buf)
		if err != nil {
			return volume{}, err
		}
	}

	data, err := ioutil.ReadAll(buf)
	if err != nil {
		return volume{}, err
	}

	return volume{header, entries, data}, nil
}

func writeVolume(v volume) ([]byte, error) {
	headerData, err := writeHeader(v.header)
	if err != nil {
		return nil, err
	}

	var restData []byte
	for _, entry := range v.entries {
		fileEntryData, err := writeFileEntry(entry)
		if err != nil {
			return nil, err
		}
		restData = append(restData, fileEntryData...)
	}
	restData = append(restData, v.data...)

	header := v.header
	header.ControlHash = md5.Sum(append(headerData[controlHashOffset:], restData...))
	headerData, err = writeHeader(header)
	if err != nil {
		return nil, err
	}

	return append(headerData, restData...), nil
}