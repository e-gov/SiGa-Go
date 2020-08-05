package siga

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	hashcodesSHA256 = "META-INF/hashcodes-sha256.xml"
	hashcodesSHA512 = "META-INF/hashcodes-sha512.xml"
)

// toReaderAt converts an io.Reader to an io.ReaderAt and size. It attempts to
// minimize data copying, but falls back to reading the entire stream into
// memory if necessary.
func toReaderAt(r io.Reader) (io.ReaderAt, int64, error) {
	// If r implements io.ReaderAt and io.Seeker, then we can avoid any
	// additional copying of data. Applies to *bytes.Reader, *os.File, etc.
	if ras, ok := r.(readAtSeeker); ok {
		off, err := ras.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, 0, errors.WithMessage(err, "check offset")
		}
		end, err := ras.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, 0, errors.WithMessage(err, "check size")
		}

		// If we are reading from the start of the stream, then return
		// ras and its size directly. Otherwise wrap in an
		// *io.SectionReader.
		if off == 0 {
			return ras, end, nil
		}
		section := io.NewSectionReader(ras, off, end-off)
		return section, section.Size(), nil
	}

	// Otherwise we need to fall back to reading the entire byte slice and
	// wrapping it in a *bytes.Reader, because even if r implements
	// io.ReaderAt only, we have no way of checking the read offset and
	// could end up reading header bytes that were not meant for us.
	//
	// We also cannot optimize reading of the bytes by checking for
	// *bytes.Buffer and calling Bytes() because it is a reasonable
	// expectation of the caller that r will be drained until EOF.
	b, err := ioutil.ReadAll(r)
	return bytes.NewReader(b), int64(len(b)), errors.WithStack(err)
}

type readAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

// toHashcode transforms a complete signature container read from src to a
// hashcode form signature container and writes it to dst. size indicates the
// size of src in bytes. toHashcode returns the datafiles read from src.
func toHashcode(dst io.Writer, src io.ReaderAt, size int64) ([]*DataFile, error) {
	reader, err := zip.NewReader(src, size)
	if err != nil {
		return nil, errors.Wrap(err, "open zip")
	}
	writer := zip.NewWriter(dst)

	// Copy files from src, collecting data files and dropping them from
	// the output.
	var datafiles []*DataFile
	copybuf := make([]byte, 32*1024) // XXX: Reuse via sync.Pool?
	for _, file := range reader.File {
		switch file.Name {
		case hashcodesSHA256, hashcodesSHA512:
			return nil, errors.Errorf("hashcode %s in complete container", file.Name)
		}

		if file.Name != "mimetype" && !strings.HasPrefix(file.Name, "META-INF/") {
			df, err := zipDataFile(file)
			if err != nil {
				return nil, err
			}
			datafiles = append(datafiles, df)
			continue // Do not copy to output.
		}

		if err := zipCopy(writer, file, copybuf, file.Name != "mimetype"); err != nil {
			return nil, err
		}
	}

	// Write hashcode files to the archive.
	if err := writeHashcodes(writer, hashcodesSHA256, datafiles, false); err != nil {
		return nil, err
	}
	if err := writeHashcodes(writer, hashcodesSHA512, datafiles, true); err != nil {
		return nil, err
	}

	return datafiles, errors.Wrap(writer.Close(), "close zip")
}

// fromHashcode transforms a hashcode form signature container read from src to
// a complete signature container and writes it to dst. size indicates the size
// of src in bytes. The data files signed in the hashcode form container must
// match exactly datafiles.
func fromHashcode(
	dst io.Writer,
	src io.ReaderAt,
	size int64,
	datafiles ...*DataFile) error {
	reader, err := zip.NewReader(src, size)
	if err != nil {
		return errors.Wrap(err, "open zip")
	}
	writer := zip.NewWriter(dst)

	// Copy files from src, validating and dropping the two hashcode files.
	var sha256, sha512 bool
	copybuf := make([]byte, 32*1024) // XXX: Reuse via sync.Pool?
	for _, file := range reader.File {
		if file.Name != "mimetype" && !strings.HasPrefix(file.Name, "META-INF/") {
			return errors.Errorf("datafile %s in hashcode container", file.Name)
		}

		switch file.Name {
		case hashcodesSHA256:
			if err := checkHashcodes(file, datafiles, false); err != nil {
				return err
			}
			sha256 = true
			continue // Do not copy to output.
		case hashcodesSHA512:
			if err := checkHashcodes(file, datafiles, true); err != nil {
				return err
			}
			sha512 = true
			continue // Do not copy to output.
		}

		if err := zipCopy(writer, file, copybuf, file.Name != "mimetype"); err != nil {
			return err
		}
	}
	if !sha256 {
		return errors.New("missing SHA-256 hashcodes")
	}
	if !sha512 {
		return errors.New("missing SHA-512 hashcodes")
	}

	// Write the datafiles to the archive.
	for _, datafile := range datafiles {
		if err := zipWrite(writer, &zip.FileHeader{
			Name:               datafile.meta.Name,
			Method:             zip.Deflate,
			Modified:           time.Now(),
			UncompressedSize64: uint64(datafile.meta.Size),
		}, datafile.contents); err != nil {
			return err
		}
	}

	return errors.Wrap(writer.Close(), "close zip")
}

func zipDataFile(file *zip.File) (*DataFile, error) {
	r, err := file.Open()
	if err != nil {
		return nil, errors.Wrapf(err, "open %s", file.Name)
	}
	defer r.Close()
	return NewDataFile(file.Name, r)
}

func zipCopy(writer *zip.Writer, file *zip.File, buf []byte, forceDeflate bool) error {
	r, err := file.Open()
	if err != nil {
		return errors.Wrapf(err, "open %s", file.Name)
	}
	defer r.Close()

	header := file.FileHeader
	if forceDeflate {
		header.Method = zip.Deflate
	}
	w, err := writer.CreateHeader(&header)
	if err != nil {
		return errors.Wrapf(err, "create %s", file.Name)
	}

	_, err = io.CopyBuffer(w, r, buf)
	return errors.Wrapf(err, "copy %s", file.Name)
}

func zipWrite(writer *zip.Writer, header *zip.FileHeader, contents []byte) error {
	w, err := writer.CreateHeader(header)
	if err != nil {
		return errors.Wrapf(err, "create %s", header.Name)
	}

	_, err = w.Write(contents)
	return errors.Wrapf(err, "write %s", header.Name)
}

type hashcodes struct {
	XMLName     xml.Name    `xml:"hashcodes"`
	FileEntries []fileEntry `xml:"file-entry"`
}

type fileEntry struct {
	FullPath string `xml:"full-path,attr"`
	Hash     string `xml:"hash,attr"`
	Size     int64  `xml:"size,attr"`
}

func writeHashcodes(writer *zip.Writer, name string, datafiles []*DataFile, sha512 bool) error {
	var calculated hashcodes
	for _, datafile := range datafiles {
		hash := datafile.meta.SHA256
		if sha512 {
			hash = datafile.meta.SHA512
		}
		calculated.FileEntries = append(calculated.FileEntries, fileEntry{
			FullPath: datafile.meta.Name,
			Hash:     hash,
			Size:     int64(datafile.meta.Size),
		})
	}

	w, err := writer.Create(name)
	if err != nil {
		return errors.Wrapf(err, "create %s", name)
	}
	return errors.Wrapf(xml.NewEncoder(w).Encode(calculated), "write %s", name)
}

func checkHashcodes(file *zip.File, datafiles []*DataFile, sha512 bool) error {
	r, err := file.Open()
	if err != nil {
		return errors.Wrapf(err, "open %s", file.Name)
	}
	defer r.Close()

	var parsed hashcodes
	if err := xml.NewDecoder(r).Decode(&parsed); err != nil {
		return errors.Wrapf(err, "parse %s", file.Name)
	}

	index := make(map[string]*DataFile, len(datafiles))
	for _, datafile := range datafiles {
		index[datafile.meta.Name] = datafile
	}
	for _, entry := range parsed.FileEntries {
		datafile, ok := index[entry.FullPath]
		if !ok {
			return errors.Errorf("unknown %s in %s", entry.FullPath, file.Name)
		}
		hash := datafile.meta.SHA256
		if sha512 {
			hash = datafile.meta.SHA512
		}
		if entry.Hash != hash {
			return errors.Errorf("mismatching %s hash in %s: %s != %s",
				entry.FullPath, file.Name, entry.Hash, hash)
		}
		if entry.Size != int64(datafile.meta.Size) {
			return errors.Errorf("mismatching %s size in %s: %d != %d",
				entry.FullPath, file.Name, entry.Size, datafile.meta.Size)
		}
		delete(index, entry.FullPath)
	}
	for name := range index {
		return errors.Errorf("missing %s from %s", name, file.Name)
	}
	return nil
}
