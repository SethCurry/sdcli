package exif

import (
	"bytes"
	"fmt"
	"io"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
	pis "github.com/dsoprea/go-png-image-structure/v2"
)

type exifWriter interface {
	SetExif(*exif.IfdBuilder) error
	ConstructExifBuilder() (*exif.IfdBuilder, error)
	Write(io.Writer) error
}

type wrappedChunkSlice struct {
	*pis.ChunkSlice
}

func (w wrappedChunkSlice) Write(to io.Writer) error {
	return w.WriteTo(to)
}

type exifExtractor func([]byte) (exifWriter, error)

func addExifToImage(imgBytes []byte, extractor exifExtractor, prompt string) ([]byte, error) {
	parsedImage, err := extractor(imgBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image with Exif extractor: %w", err)
	}

	im, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		return nil, fmt.Errorf("failed to create new exif mapping: %w", err)
	}

	ti := exif.NewTagIndex()
	ib := exif.NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.TestDefaultByteOrder)

	err = addMetadata(ib, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to build new Exif metadata: %w", err)
	}

	err = parsedImage.SetExif(ib)
	if err != nil {
		return nil, fmt.Errorf("failed to set new Exif on image: %w", err)
	}

	buf := bytes.NewBuffer([]byte{})

	err = parsedImage.Write(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to write image back to buffer")
	}

	return buf.Bytes(), nil
}

func AddToPNG(imgBytes []byte, prompt string) ([]byte, error) {
	return addExifToImage(imgBytes, func(gotBytes []byte) (exifWriter, error) {
		parsed, err := pis.NewPngMediaParser().ParseBytes(imgBytes)
		if err != nil {
			return nil, err
		}

		sl, ok := parsed.(*pis.ChunkSlice)
		if !ok {
			return nil, fmt.Errorf("failed to convert parsed png to ChunkSlice: unexpected type %T", parsed)
		}

		return wrappedChunkSlice{sl}, nil
	}, prompt)
}

func AddToJPEG(imgBytes []byte, prompt string) ([]byte, error) {
	return addExifToImage(imgBytes, func(gotBytes []byte) (exifWriter, error) {
		parsed, err := jis.NewJpegMediaParser().ParseBytes(imgBytes)
		if err != nil {
			return nil, err
		}

		sl, ok := parsed.(*jis.SegmentList)
		if !ok {
			return nil, fmt.Errorf("failed to convert parsed image to SegmentList: unexpected type %T", parsed)
		}

		return sl, nil
	}, prompt)
}

func addMetadata(ib *exif.IfdBuilder, prompt string) error {
	ifd0Ib, err := exif.GetOrCreateIbFromRootIb(ib, "IFD0")
	if err != nil {
		return fmt.Errorf("failed to create IFD0 ib: %w", err)
	}

	err = ifd0Ib.AddStandardWithName("Artist", "Stable Diffusion")
	if err != nil {
		return fmt.Errorf("failed to set Artist tag: %w", err)
	}

	err = ifd0Ib.AddStandardWithName("ImageDescription", prompt)
	if err != nil {
		return fmt.Errorf("failed to set ImageDescription tag: %w", err)
	}

	return nil
}
