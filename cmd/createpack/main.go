// packfs / createpack
// MIT License Copyright(c) 2019 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
package main

import (
    "bytes"
    b "encoding/binary"
    "io"
    "log"
    "math/rand"
    "os"
    "path/filepath"
    "time"

    "github.com/hshimamoto/packfs/pack"
)

type source struct {
    path string
    info os.FileInfo
    off uint32
}

type Pack struct {
    pack.Pack
    file *os.File
    key uint64
    sources []source
}

func (p *Pack)write(buf []byte) {
    sz := len(buf)
    if (sz % 8) != 0 {
	log.Fatal("align error")
    }
    obuf := bytes.NewReader(buf)
    ebuf := &bytes.Buffer{}
    for sz > 0 {
	var edata uint64
	b.Read(obuf, b.LittleEndian, &edata)
	edata ^= p.key
	b.Write(ebuf, b.LittleEndian, edata)
	sz -= 8
    }
    p.file.Write(ebuf.Bytes())
}

func (p *Pack)write_key() {
    log.Printf("write key: 0x%x", p.key)
    buf := &bytes.Buffer{}
    b.Write(buf, b.BigEndian, p.key)
    store := buf.Bytes()
    for i := 0; i < 8; i++ {
	if rand.Intn(2) == 0 {
	    store[i] &= 0xfe
	}
    }
    p.file.Write(store)
}

func (p *Pack)write_version() {
    log.Printf("write pack version")
    buf := &bytes.Buffer{}
    buf.Write([]byte("pack0000")) // pack version 0
    p.write(buf.Bytes())
}

func (p *Pack)write_file_sizes() {
    log.Printf("write file sizes")
    buf := &bytes.Buffer{}
    for _, s := range p.sources {
	sz := uint32(s.info.Size())
	b.Write(buf, b.LittleEndian, sz)
    }
    term := uint32(0xffffffff)
    b.Write(buf, b.LittleEndian, term)
    if (buf.Len() % 8) != 0 {
	b.Write(buf, b.LittleEndian, term)
    }
    p.write(buf.Bytes())
}

func (p *Pack)write_file_names() {
    log.Printf("write file names")
    buf := &bytes.Buffer{}
    for _, s := range p.sources {
	name := s.info.Name()
	sz := len(name) + 1
	buf.Write([]byte(name))
	buf.WriteByte(0)
	for (sz % 8) != 0 {
	    buf.WriteByte(0)
	    sz++
	}
    }
    p.write(buf.Bytes())
}

func (p *Pack)write_file_offsets() {
    log.Printf("write file offsets")
    buf := &bytes.Buffer{}
    off := uint32(0)
    for _, s := range p.sources {
	s.off = off
	b.Write(buf, b.LittleEndian, off)
	off += uint32(s.info.Size()) + 7
	off &= 0xfffffff8
    }
    if (buf.Len() % 8) != 0 {
	term := uint32(0xffffffff)
	b.Write(buf, b.LittleEndian, term)
    }
    p.write(buf.Bytes())
}

func (p *Pack)write_file_data() {
    log.Printf("write file data")
    for _, s := range p.sources {
	buf := &bytes.Buffer{}
	f, err := os.Open(s.path)
	if err != nil {
	    log.Fatal(err)
	}
	sz, err := io.Copy(buf, f)
	if err != nil {
	    log.Fatal(err)
	}
	for (sz % 8) != 0 {
	    buf.WriteByte(0)
	    sz++
	}
	p.write(buf.Bytes())
    }
}

func (p *Pack)makelist(dirs []string) {
    list := []source{}
    for _, d := range dirs {
	info, err := os.Stat(d)
	if err != nil {
	    log.Printf("skil: %s", d)
	    continue
	}
	if info.IsDir() {
	    walkfunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
		    return err
		}
		if info.IsDir() {
		    if info.Name() == "." {
			return nil
		    }
		    return filepath.SkipDir
		}
		list = append(list, source{ path: path, info: info })
		return nil
	    }
	    filepath.Walk(d, walkfunc)
	    continue
	}
	list = append(list, source{ path: d, info: info })
    }
    p.sources = list
}

func main() {
    if len(os.Args) < 3 {
	log.Fatal("Usage:\n\tcreatepack packfile source")
    }

    path := os.Args[1]
    // create file
    if _, err := os.Stat(path); err == nil {
	log.Fatal("File exists")
    }
    packfile, err := os.Create(path)
    if err != nil {
	log.Fatal(err)
    }
    defer packfile.Close()

    rand.Seed(time.Now().UnixNano())

    // prepare
    p := &Pack{
	file: packfile,
	key: rand.Uint64() | 0x0101010101010101,
	sources: []source{},
    }

    p.makelist(os.Args[2:])

    p.write_key()
    p.write_version()
    p.write_file_sizes()
    p.write_file_names()
    p.write_file_offsets()
    p.write_file_data()
}
