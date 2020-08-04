package siga

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"encoding/binary"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
)

// forZipInputStream wraps w with a new io.Writer which expects an ASiC-E
// container stream to be written to it. It modifies the stream by removing the
// data descriptor of the first file ("mimetype") and updating its local file
// header with the data from the descriptor. All other files must not have a
// data descriptor or be compressed using DEFLATE.
//
// The wrapper also updates any changed offsets in central directory entries.
// The modified stream is written to w.
//
// This acts as a workaround for limited ZIP-archive parsing methods like
// java.util.ZipInputStream, which do not consult the central directory and
// therefore only support a subset of ZIP-archives.
//
// The returned io.Writer performs very little verification on the input and
// writing non-valid ZIP-archives to it results in undefined behavior. It does
// not work with multi-disk ZIP-archives and/or ZIP64.
func forZipInputStream(w io.Writer) io.Writer {
	return &zipInputStream{output: w, decomp: flate.NewReader(nil)}
}

const (
	zipLocalSignature      = "\x50\x4b\x03\x04"
	zipDescriptorSignature = "\x50\x4b\x07\x08"
	zipCentralSignature    = "\x50\x4b\x01\x02"
	zipEOCDSignature       = "\x50\x4b\x05\x06"

	asiceMimetype           = "application/vnd.etsi.asic-e+zip"
	asiceMimetypeCRC32      = "\x8a\x21\xf9\x45"
	asiceMimetypeSize       = "\x1f\x00\x00\x00"
	asiceMimetypeDescriptor = zipDescriptorSignature +
		asiceMimetypeCRC32 + asiceMimetypeSize + asiceMimetypeSize
)

type zipInputStream struct {
	buf     bytes.Buffer
	err     error
	output  io.Writer
	written int64
	decomp  io.ReadCloser
	recalc  bool
}

func (z *zipInputStream) Write(p []byte) (int, error) {
	if z.err != nil {
		return 0, z.err
	}
	z.buf.Write(p)

	// Flush as much of the data as possible.
	for ok := true; ok; { // do-while(ok)
		if z.buf.Len() < 4 {
			break // Not enough data to continue.
		}

		switch sig := string(z.buf.Bytes()[:4]); sig {
		case zipLocalSignature:
			ok = z.flushLocal()
		case zipCentralSignature:
			ok = z.flushCentral()
		case zipEOCDSignature:
			ok = z.flushEOCD()
		default:
			z.err = errors.Errorf("unknown signature: %x", sig)
			ok = false
		}
	}
	return len(p), z.err
}

// flushLocal attempts to process and flush a single local file entry from the
// buffer. It returns false if it did not succeed.
//
// Note that it can return false without encountering an error (z.err == nil):
// this happens if the buffer does not have enough data.
func (z *zipInputStream) flushLocal() bool {
	buf := z.buf.Bytes()
	if len(buf) < 30 {
		return false
	}
	descriptor := buf[6]&8 == 8
	compression := binary.LittleEndian.Uint16(buf[8:10])
	size := binary.LittleEndian.Uint32(buf[18:22])
	name := binary.LittleEndian.Uint16(buf[26:28])
	extra := binary.LittleEndian.Uint16(buf[28:30])

	header := 30 + int(name) + int(extra)
	if len(buf) < header {
		return false
	}

	// If no descriptor is used, then try to flush the header and data.
	if !descriptor {
		return z.flushBytes(header + int(size))
	}

	// If DEFLATE compression is used, then try to flush the header, data,
	// and descriptor.
	if compression == zip.Deflate {
		// Check if we have the entire compressed stream (DEFLATE
		// indicates which block is final).
		r := bytes.NewReader(buf[header:])
		z.decomp.(flate.Resetter).Reset(r, nil) // Never fails.
		_, err := io.Copy(ioutil.Discard, z.decomp)
		if err == nil {
			err = z.decomp.Close()
		}
		if err == io.ErrUnexpectedEOF {
			return false // Not enough data yet.
		}
		if err != nil {
			z.err = errors.WithStack(err)
			return false
		}

		// Flush read portion of buf + 16 data descriptor bytes.
		return z.flushBytes(len(buf) - r.Len() + 16)
	}

	// Otherwise must be "mimetype".
	if string(buf[30:30+name]) != "mimetype" {
		z.err = errors.New("only mimetype may use a data descriptor and be uncompressed")
		return false
	}
	if z.written > 0 {
		z.err = errors.New("mimetype not first file in stream")
		return false
	}

	// Do not attempt to scan for the data descriptor signature over raw
	// data, compare against known value instead.
	end := header + len(asiceMimetype) + len(asiceMimetypeDescriptor)
	if len(buf) < end {
		return false
	}
	if data := string(buf[header:end]); data != asiceMimetype+asiceMimetypeDescriptor {
		z.err = errors.Errorf("unexpected mimetype data: %q", data)
		return false
	}

	// Update local file header and flush it with data. Skip descriptor.
	buf[6] &^= 8
	copy(buf[14:], asiceMimetypeCRC32)
	copy(buf[18:], asiceMimetypeSize)
	copy(buf[22:], asiceMimetypeSize)
	ok := z.flushBytes(header + len(asiceMimetype))
	if ok {
		z.buf.Next(len(asiceMimetypeDescriptor))
		z.recalc = true // Offsets need to be recalculated.
	}
	return ok
}

// flushCentral attempts to process and flush a single central directory file
// entry from the buffer. It returns false if it did not succeed.
//
// Note that it can return false without encountering an error (z.err == nil):
// this happens if the buffer does not have enough data.
func (z *zipInputStream) flushCentral() bool {
	buf := z.buf.Bytes()
	if len(buf) < 46 {
		return false
	}
	name := binary.LittleEndian.Uint16(buf[28:30])
	extra := binary.LittleEndian.Uint16(buf[30:32])
	comment := binary.LittleEndian.Uint16(buf[32:34])
	offset := binary.LittleEndian.Uint32(buf[42:46])

	// Ensure enough data for flushBytes before recalculating so it is only
	// done at most once.
	header := 46 + int(name) + int(extra) + int(comment)
	if len(buf) < header {
		return false
	}

	// If the mimetype data descriptor was removed and offsets need to be
	// recalculated, then do so for all entries except for the first one.
	if z.recalc && offset > 0 {
		offset -= uint32(len(asiceMimetypeDescriptor))
		binary.LittleEndian.PutUint32(buf[42:46], offset)
	}
	return z.flushBytes(header)
}

// flushEOCD attempts to process and flush the end of central directory record
// from the buffer. It returns false if it did not succeed.
//
// Note that it can return false without encountering an error (z.err == nil):
// this happens if the buffer does not have enough data.
func (z *zipInputStream) flushEOCD() bool {
	buf := z.buf.Bytes()
	if len(buf) < 22 {
		return false
	}
	offset := binary.LittleEndian.Uint32(buf[16:20])
	comment := binary.LittleEndian.Uint16(buf[20:22])

	// Ensure enough data for flushBytes before recalculating so it is only
	// done at most once.
	header := 22 + int(comment)
	if len(buf) < header {
		return false
	}

	// If the mimetype data descriptor was removed and offsets need to be
	// recalculated, then do so for the start of central directory offset.
	if z.recalc {
		offset -= uint32(len(asiceMimetypeDescriptor))
		binary.LittleEndian.PutUint32(buf[16:20], offset)
	}
	return z.flushBytes(header)
}

func (z *zipInputStream) flushBytes(n int) bool {
	if z.buf.Len() < n {
		return false
	}
	n, err := z.output.Write(z.buf.Next(n))
	z.written += int64(n)
	if err != nil {
		z.err = errors.WithStack(err)
		return false
	}
	return true
}
