package proto_coder

import (
	"../config"
	"context"
	"fmt"
	"os"
	"path"
	"time"
)

var fileCache = make(map[string]time.Time)

func WatchProtos(ctx context.Context) {
	load(ctx)
	go func() {
		for {
			if c := load(ctx); c == false {
				break
			}
		}
	}()
}

func load(ctx context.Context) bool {
	p := *conf.ProtoPath
	fi, err := os.Stat(p)
	if err != nil {
		fmt.Println("load proto error=>", err)
		t := time.NewTimer(time.Second * 5)
		select {
		case <-t.C:
			{
				return true
			}
		case <-ctx.Done():
			{
				t.Stop()
				return false
			}
		}
	}

	if !fi.Mode().IsDir() {
		t := time.NewTimer(time.Second * 5)
		select {
		case <-t.C:
			{
				return true
			}
		case <-ctx.Done():
			{
				t.Stop()
				return false
			}
		}
	}

	for _, f := range (lsfiles(p)) {
		if ext := path.Ext(f.Name()); ext != ".proto" {
			//fmt.Println(f)
			continue
		}
		pp := path.Join(p, f.Name())
		if !fileCache[pp].Before(f.ModTime()) {
			continue
		}
		fileCache[pp] = f.ModTime()
		fmt.Println("update", f.Name())
		UpdateProto(pp)
	}

	t := time.NewTimer(time.Second * 5)
	select {
	case <-t.C:
		{
			return true
		}
	case <-ctx.Done():
		{
			t.Stop()
			return false
		}
	}

}

func lsfiles(p string) []os.FileInfo {
	f, err := os.Open(p)
	if err != nil {
		fmt.Println(err);
		return nil
	}
	defer func() { _ = f.Close() }()

	fs, err := f.Readdir(-1)
	if err != nil {
		fmt.Println(err)
	}
	return fs;
}
