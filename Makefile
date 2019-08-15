all: __cmds__

__cmds__: vendor/github.com/hanwen/go-fuse/.ok
	make -C cmd/packfs
	make -C cmd/packweb
	make -C cmd/createpack

vendor/github.com/hanwen/go-fuse/.ok: go-fuse-1.0.0.tar.gz
	tar zxf go-fuse-1.0.0.tar.gz
	mkdir -p vendor/github.com/hanwen
	mv go-fuse-1.0.0 vendor/github.com/hanwen/go-fuse
	touch $@

go-fuse-1.0.0.tar.gz:
	curl -vL -o $@ https://github.com/hanwen/go-fuse/archive/v1.0.0.tar.gz

clean: __clean_cmds__

__clean_cmds__:
	make -C cmd/packfs clean
	make -C cmd/packweb clean
	make -C cmd/createpack clean
