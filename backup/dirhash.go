package backup

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
)

func DirHash(path string) (string, error) {
	hash := md5.New()
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		io.WriteString(hash, path)
		fmt.Fprintf(hash, "%v", info.IsDir())
		fmt.Fprintf(hash, "%v", info.ModTime())
		fmt.Fprintf(hash, "%v", info.Mode())
		fmt.Fprintf(hash, "%v", info.Name())
		fmt.Fprintf(hash, "%v", info.Size())
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
