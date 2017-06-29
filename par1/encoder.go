package par1

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/klauspost/reedsolomon"
)

// An Encoder keeps track of all information needed to create parity
// volumes for a set of data files, and write them out to parity files
// (.PAR, .P00, .P01, etc.).
type Encoder struct {
	fileIO   fileIO
	delegate EncoderDelegate

	filePaths   []string
	volumeCount int

	shardByteCount int
	fileData       [][]byte
	parityData     [][]byte
}

// EncoderDelegate holds methods that are called during the encode
// process.
type EncoderDelegate interface {
	OnDataFileLoad(path string, err error)
	OnVolumeFileWrite(path string, err error)
}

func newEncoder(fileIO fileIO, delegate EncoderDelegate, filePaths []string, volumeCount int) (*Encoder, error) {
	// TODO: Check len(filePaths) and volumeCount.
	return &Encoder{fileIO, delegate, filePaths, volumeCount, 0, nil, nil}, nil
}

// NewEncoder creates an encoder with the given list of file paths,
// and with the given number of intended parity volumes.
func NewEncoder(delegate EncoderDelegate, filePaths []string, volumeCount int) (*Encoder, error) {
	return newEncoder(defaultFileIO{}, delegate, filePaths, volumeCount)
}

// LoadFileData loads the file data into memory.
func (e *Encoder) LoadFileData() error {
	shardByteCount := 0
	fileData := make([][]byte, len(e.filePaths))
	for i, path := range e.filePaths {
		var err error
		fileData[i], err = e.fileIO.ReadFile(path)
		e.delegate.OnDataFileLoad(path, err)
		if err != nil {
			return err
		}

		if len(fileData[i]) > shardByteCount {
			shardByteCount = len(fileData[i])
		}
	}

	e.shardByteCount = shardByteCount
	e.fileData = fileData
	return nil
}

func (e *Encoder) buildShards() [][]byte {
	shards := make([][]byte, len(e.fileData)+e.volumeCount)
	for i, data := range e.fileData {
		padding := make([]byte, e.shardByteCount-len(data))
		shards[i] = append(data, padding...)
	}

	for i := 0; i < e.volumeCount; i++ {
		shards[len(e.fileData)+i] = make([]byte, e.shardByteCount)
	}

	return shards
}

// ComputeParityData computes the parity data for the files.
func (e *Encoder) ComputeParityData() error {
	shards := e.buildShards()

	rs, err := reedsolomon.New(len(e.fileData), e.volumeCount, reedsolomon.WithPAR1Matrix())
	if err != nil {
		return err
	}

	err = rs.Encode(shards)
	if err != nil {
		return err
	}

	e.parityData = shards[len(e.fileData):]
	return nil
}

func (e *Encoder) Write(indexPath string) error {
	var entries []fileEntry
	for i, k := range e.filePaths {
		entry := fileEntry{
			header: fileEntryHeader{
				FileBytes: uint64(len(e.fileData[i])),
				// TODO: Compute Hash and SixteenKHash
				// properly.
			},
			filename: filepath.Base(k),
		}
		entries = append(entries, entry)
	}

	vTemplate := volume{
		header: header{
			ID:            expectedID,
			VersionNumber: expectedVersion,
			// TODO: Compute SetHash properly.
		},
		entries: entries,
	}

	indexVolume := vTemplate
	indexVolume.header.VolumeNumber = 0
	indexVolumeBytes, err := writeVolume(indexVolume)
	if err != nil {
		return err
	}

	// TODO: Sanity-check indexPath.
	ext := path.Ext(indexPath)
	base := indexPath[:len(indexPath)-len(ext)]

	realIndexPath := base + ".par"
	err = e.fileIO.WriteFile(realIndexPath, indexVolumeBytes)
	e.delegate.OnVolumeFileWrite(realIndexPath, err)
	if err != nil {
		return err
	}

	for i, parityShard := range e.parityData {
		vol := vTemplate
		vol.header.VolumeNumber = uint64(i + 1)
		vol.data = parityShard
		volBytes, err := writeVolume(vol)
		if err != nil {
			return err
		}

		// TODO: Handle more than 99 parity files.
		volumePath := fmt.Sprintf("%s.p%02d", base, i+1)
		err = e.fileIO.WriteFile(volumePath, volBytes)
		e.delegate.OnVolumeFileWrite(volumePath, err)
		if err != nil {
			return err
		}
	}

	return nil
}
