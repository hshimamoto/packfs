// packfs
// MIT License Copyright(c) 2019 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
package pack

import (
    "bytes"
    b "encoding/binary"
    "fmt"
    "io"
    "os"
    "time"
)

type File struct {
    Name string
    Size uint32
    // for Attr
    Xtime uint64
    // internal
    pack *Pack
    offset uint32
    data []byte
    last time.Time
}

func (f *File)Clear() {
    f.data = []byte{}
}

func (f *File)Bytes() []byte {
    sz := uint32(len(f.data))
    if sz < f.Size {
	sz = f.Size
	if (sz % 8) != 0 {
	    sz += 8 - (sz % 8)
	}
	f.data = f.pack.readat(f.offset, sz)[:f.Size]
    }
    f.last = time.Now()
    return f.data
}

func NewFile(sz uint32, xtm uint64, p *Pack) *File {
    return &File{
	Size: sz,
	Xtime: xtm,
	pack: p,
    }
}

/*
 * Pack File Layout
 *
 * key uint64 / key xor 0x0101010101010101
 * encrypted data...
 *  xor with key
 * File meta data
 *  size uint32
 * terminate with size 0xffffffff
 * align 64bit
 * File names
 *  name string
 * terminate with NUL
 * align 64bit
 * File offsets
 *  offset uint32
 * align 64bit
 * File data
 *  data []byte
 * align 64bit
 *
 */

type Pack struct {
    Files []*File
    Xtime uint64
    key uint64
    file *os.File
    info os.FileInfo
}

func (p *Pack)read(sz uint32) []byte {
    buf := &bytes.Buffer{}
    for sz > 0 {
	var edata uint64
	b.Read(p.file, b.LittleEndian, &edata)
	edata ^= p.key
	b.Write(buf, b.LittleEndian, edata)
	sz -= 8
    }
    return buf.Bytes()
}

func (p *Pack)readat(off, sz uint32) []byte {
    p.file.Seek(int64(off), io.SeekStart)
    return p.read(sz)
}

func OpenPack(path string) (*Pack, error) {
    f, err := os.Open(path)
    if err != nil {
	return nil, err
    }
    info, err := f.Stat()
    if err != nil {
	return nil, err
    }
    xtm := uint64(info.ModTime().Unix())
    p := &Pack{
	Files: []*File{},
	Xtime: xtm,
	file: f,
	info: info,
    }
    hdrsz := 0
    // get key
    b.Read(f, b.BigEndian, &p.key)
    hdrsz += 8
    p.key |= 0x0101010101010101
    // get version
    version := p.read(8)
    hdrsz += 8
    if string(version) != "pack0000" {
	return nil, fmt.Errorf("Unknown version %s", string(version))
    }
    // get file sizes
    for {
	sbuf := bytes.NewReader(p.read(8))
	hdrsz += 8
	var sz0, sz1 uint32
	b.Read(sbuf, b.LittleEndian, &sz0)
	b.Read(sbuf, b.LittleEndian, &sz1)
	if sz0 == 0xffffffff {
	    break
	}
	p.Files = append(p.Files, NewFile(sz0, xtm, p))
	if sz1 == 0xffffffff {
	    break
	}
	p.Files = append(p.Files, NewFile(sz1, xtm, p))
    }
    // get file names
    for _, f := range p.Files {
	done := false
	for !done {
	    sbuf := p.read(8)
	    hdrsz += 8
	    for _, ch := range sbuf {
		if ch == 0 {
		    done = true
		    break
		}
		f.Name += string(ch)
	    }
	}
    }
    // get offsets
    for i := 0; i < len(p.Files); i += 2 {
	sbuf := bytes.NewReader(p.read(8))
	hdrsz += 8
	var off0, off1 uint32
	b.Read(sbuf, b.LittleEndian, &off0)
	b.Read(sbuf, b.LittleEndian, &off1)
	p.Files[i].offset = off0
	if off1 == 0xffffffff {
	    break
	}
	p.Files[i + 1].offset = off1
    }
    // update offsets
    for _, f := range p.Files {
	f.offset += uint32(hdrsz)
    }
    return p, nil
}
