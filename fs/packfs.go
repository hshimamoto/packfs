// packfs
// MIT License Copyright(c) 2019 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
package packfs

import (
    "github.com/hshimamoto/packfs/pack"

    "github.com/hanwen/go-fuse/fuse"
    "github.com/hanwen/go-fuse/fuse/nodefs"
    "github.com/hanwen/go-fuse/fuse/pathfs"
)

type PackFile struct {
    nodefs.File
    file *pack.File
}

func NewPackFile(file *pack.File) nodefs.File {
    f := new(PackFile)
    f.File = nodefs.NewDefaultFile()
    f.file = file
    return f
}

func (f *PackFile)Read(dest []byte, off int64) (fuse.ReadResult, fuse.Status) {
    data := f.file.Bytes()
    sz := f.file.Size
    return fuse.ReadResultData(data[off:sz]), fuse.OK
}

func (f *PackFile)Release() {
    // drop cache
    f.file.Clear()
}

type PackFS struct {
    pathfs.FileSystem
    fs *pathfs.PathNodeFs
    server *fuse.Server
    pack *pack.Pack
}

func (fs *PackFS)GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
    if name == "" {
	return &fuse.Attr{
	    Mode: fuse.S_IFDIR | 0555,
	    Atime: fs.pack.Xtime,
	    Mtime: fs.pack.Xtime,
	    Ctime: fs.pack.Xtime,
	}, fuse.OK
    }
    for _, f := range fs.pack.Files {
	if f.Name == name {
	    return &fuse.Attr{
		Mode: fuse.S_IFREG | 0444,
		Size: uint64(f.Size),
		Atime: uint64(f.Xtime),
		Mtime: uint64(f.Xtime),
		Ctime: uint64(f.Xtime),
	    }, fuse.OK
	}
    }
    return nil, fuse.ENOENT
}

func (fs *PackFS)OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
    if name == "" {
	c = []fuse.DirEntry{}
	for _, f := range fs.pack.Files {
	    c = append(c, fuse.DirEntry{Name: f.Name, Mode: fuse.S_IFREG})
	}
	return c, fuse.OK
    }
    return nil, fuse.ENOENT
}

func (fs *PackFS)Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
    for _, f := range fs.pack.Files {
	if f.Name == name {
	    if flags & fuse.O_ANYWRITE != 0 {
		return nil, fuse.EPERM
	    }
	    return NewPackFile(f), fuse.OK
	}
    }
    return nil, fuse.ENOENT
}

func (fs *PackFS)Serve() {
    fs.server.Serve()
}

func NewPackFS(path, mnt string) (*PackFS, error) {
    p, err := pack.OpenPack(path)
    if err != nil {
	return nil, err
    }
    pfs := &PackFS{ FileSystem: pathfs.NewDefaultFileSystem() }
    pfs.fs = pathfs.NewPathNodeFs(pfs, nil)
    server, _, err := nodefs.MountRoot(mnt, pfs.fs.Root(), nil)
    if err != nil {
	return nil, err
    }
    pfs.server = server
    pfs.pack = p
    return pfs, nil
}
