// Code generated by go-bindata. DO NOT EDIT.
// sources:
// 1000000000_init.down.sql (19B)
// 1000000000_init.up.sql (19B)
// 1000000001_init.down.sql (233B)
// 1000000001_init.up.sql (611B)

package migrations

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %w", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes  []byte
	info   os.FileInfo
	digest [sha256.Size]byte
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var __1000000000_initDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xd2\xd5\x55\x48\xcd\x2d\x28\xa9\x54\xc8\xcd\x4c\x2f\x4a\x2c\xc9\xcc\xcf\xe3\x02\x04\x00\x00\xff\xff\x32\x4d\x68\xbd\x13\x00\x00\x00")

func _1000000000_initDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__1000000000_initDownSql,
		"1000000000_init.down.sql",
	)
}

func _1000000000_initDownSql() (*asset, error) {
	bytes, err := _1000000000_initDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "1000000000_init.down.sql", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x9c, 0x46, 0xd1, 0x31, 0xb9, 0x68, 0x19, 0xcc, 0x70, 0xb6, 0x7, 0x20, 0x2e, 0x6a, 0x4d, 0xf1, 0xce, 0xd0, 0xc8, 0xda, 0x50, 0xce, 0x8c, 0xee, 0x52, 0x36, 0x80, 0xd0, 0x5a, 0xd2, 0x7a, 0x82}}
	return a, nil
}

var __1000000000_initUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xd2\xd5\x55\x48\xcd\x2d\x28\xa9\x54\xc8\xcd\x4c\x2f\x4a\x2c\xc9\xcc\xcf\xe3\x02\x04\x00\x00\xff\xff\x32\x4d\x68\xbd\x13\x00\x00\x00")

func _1000000000_initUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__1000000000_initUpSql,
		"1000000000_init.up.sql",
	)
}

func _1000000000_initUpSql() (*asset, error) {
	bytes, err := _1000000000_initUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "1000000000_init.up.sql", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x9c, 0x46, 0xd1, 0x31, 0xb9, 0x68, 0x19, 0xcc, 0x70, 0xb6, 0x7, 0x20, 0x2e, 0x6a, 0x4d, 0xf1, 0xce, 0xd0, 0xc8, 0xda, 0x50, 0xce, 0x8c, 0xee, 0x52, 0x36, 0x80, 0xd0, 0x5a, 0xd2, 0x7a, 0x82}}
	return a, nil
}

var __1000000001_initDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x72\x75\xf7\xf4\xb3\xe6\xe2\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\xf0\x74\x53\x70\x8d\xf0\x0c\x0e\x09\x56\xc8\x29\xce\x4c\x8b\x4f\x49\x2c\x49\x8c\xcf\x4d\x2d\x49\x04\x31\xac\x09\x29\x4c\xc9\x4f\x2e\xcd\x4d\xcd\x2b\x29\x26\xa8\xb2\x28\xb5\xb8\x34\xa7\x24\x3e\x39\xa3\x34\x2f\x9b\xb0\xea\x94\xd4\xb4\xcc\xbc\xcc\x92\xcc\xfc\x3c\x62\x4c\x4e\x4b\x2d\x4a\xcd\x4b\x4e\x2d\xb6\xe6\xe2\x72\xf6\xf7\xf5\xf5\x0c\xb1\xe6\x02\x04\x00\x00\xff\xff\xdc\x1f\x48\x24\xe9\x00\x00\x00")

func _1000000001_initDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__1000000001_initDownSql,
		"1000000001_init.down.sql",
	)
}

func _1000000001_initDownSql() (*asset, error) {
	bytes, err := _1000000001_initDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "1000000001_init.down.sql", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xe5, 0xa1, 0x76, 0x54, 0xd0, 0xb0, 0xb9, 0xdc, 0x93, 0x84, 0xd, 0xb5, 0xa3, 0x55, 0x91, 0xef, 0xc0, 0xe4, 0x89, 0x62, 0x20, 0x2, 0x6e, 0x92, 0x58, 0x73, 0xb3, 0x9b, 0xcd, 0xde, 0xa, 0x86}}
	return a, nil
}

var __1000000001_initUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\xcf\xc1\x4a\x03\x31\x10\x06\xe0\x7b\x9e\x62\x8e\x0a\xbe\xc1\x9e\xda\x12\x25\xb0\xdd\x82\x8d\xe0\x2d\xc4\x64\xe2\x0e\x36\xd3\x92\x4c\x60\x7d\x7b\x71\x41\x50\x91\x65\x05\xbd\x0d\xc3\x0f\xdf\xff\x6f\xf5\x9d\x19\x3a\xa5\x76\xf7\x7a\x63\x35\xd8\xcd\xb6\xd7\x60\x6e\x61\x38\x58\xd0\x8f\xe6\x68\x8f\x70\xaa\x94\x5c\xf4\xe2\x5d\x46\xf1\xef\x07\x5c\xc5\x96\x2f\x8e\x22\x10\x0b\x3e\x63\x99\xe3\xc3\x43\xdf\xdf\x00\xb7\xec\x0a\xd6\x76\x12\x17\xc6\xc6\x2f\xf5\x23\x73\xdd\xad\x43\xe2\x39\xb4\x8c\x2c\x75\x49\xb9\x78\x19\x41\x70\x92\x4f\xbf\xb9\xd9\xd3\xab\xa0\x5f\x4b\x7d\xed\xb9\xc0\x51\x9c\x7e\xf8\xfe\x1e\x8c\x98\x88\x49\xe8\xcc\x8b\x5c\x0d\x23\x66\xfc\xbe\x8f\x22\xb2\x50\x22\x2c\x7f\xb1\x3c\x61\x41\x0e\xf8\x7f\x3d\xd4\xee\xb0\xdf\x1b\xdb\xa9\xb7\x00\x00\x00\xff\xff\x8f\xd0\x35\x75\x63\x02\x00\x00")

func _1000000001_initUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__1000000001_initUpSql,
		"1000000001_init.up.sql",
	)
}

func _1000000001_initUpSql() (*asset, error) {
	bytes, err := _1000000001_initUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "1000000001_init.up.sql", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x8b, 0x99, 0xc4, 0xe6, 0xb3, 0x64, 0x2, 0x61, 0x25, 0xa1, 0x5, 0x57, 0x5f, 0xe5, 0xb2, 0xd6, 0xf0, 0xf4, 0x4c, 0xbc, 0x78, 0x2f, 0x63, 0x64, 0x2a, 0x95, 0x77, 0x8d, 0xde, 0x83, 0x6b, 0xde}}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetString returns the asset contents as a string (instead of a []byte).
func AssetString(name string) (string, error) {
	data, err := Asset(name)
	return string(data), err
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// MustAssetString is like AssetString but panics when Asset would return an
// error. It simplifies safe initialization of global variables.
func MustAssetString(name string) string {
	return string(MustAsset(name))
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetDigest returns the digest of the file with the given name. It returns an
// error if the asset could not be found or the digest could not be loaded.
func AssetDigest(name string) ([sha256.Size]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return [sha256.Size]byte{}, fmt.Errorf("AssetDigest %s can't read by error: %v", name, err)
		}
		return a.digest, nil
	}
	return [sha256.Size]byte{}, fmt.Errorf("AssetDigest %s not found", name)
}

// Digests returns a map of all known files and their checksums.
func Digests() (map[string][sha256.Size]byte, error) {
	mp := make(map[string][sha256.Size]byte, len(_bindata))
	for name := range _bindata {
		a, err := _bindata[name]()
		if err != nil {
			return nil, err
		}
		mp[name] = a.digest
	}
	return mp, nil
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"1000000000_init.down.sql": _1000000000_initDownSql,
	"1000000000_init.up.sql":   _1000000000_initUpSql,
	"1000000001_init.down.sql": _1000000001_initDownSql,
	"1000000001_init.up.sql":   _1000000001_initUpSql,
}

// AssetDebug is true if the assets were built with the debug flag enabled.
const AssetDebug = false

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"},
// AssetDir("data/img") would return []string{"a.png", "b.png"},
// AssetDir("foo.txt") and AssetDir("notexist") would return an error, and
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		canonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(canonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"1000000000_init.down.sql": {_1000000000_initDownSql, map[string]*bintree{}},
	"1000000000_init.up.sql":   {_1000000000_initUpSql, map[string]*bintree{}},
	"1000000001_init.down.sql": {_1000000001_initDownSql, map[string]*bintree{}},
	"1000000001_init.up.sql":   {_1000000001_initUpSql, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory.
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	return os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
}

// RestoreAssets restores an asset under the given directory recursively.
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(canonicalName, "/")...)...)
}
