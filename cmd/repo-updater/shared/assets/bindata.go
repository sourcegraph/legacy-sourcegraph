// Code generated by go-bindata. DO NOT EDIT.
// sources:
// state.html.tmpl (4.679kB)

package assets

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

var _stateHtmlTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x58\x5f\x53\xe3\x38\x12\x7f\xe7\x53\xf4\xb9\x76\x6f\x66\x8a\xb1\x9d\x84\x04\x06\xce\x49\x0d\xcb\x9f\x25\xec\x00\x03\x01\x76\x98\xab\x7b\x50\xac\x76\xac\x44\x96\x8c\xd4\x4e\x08\x29\xbe\xfb\x95\xed\x24\x84\x10\x38\x66\xef\xe1\xae\xb6\xd6\x0f\x89\xad\x6e\xb5\xba\xd5\x3f\xfd\x5a\x52\xf0\x37\xae\x43\x1a\xa7\x08\x31\x25\xb2\xb5\x16\xe4\x7f\x20\x99\xea\x35\x1d\x54\x4e\x6b\x6d\x2d\x88\x91\xf1\xd6\x1a\x00\x40\x40\x82\x24\xb6\x0c\xa6\xda\xcd\x52\xce\x08\x8d\x6b\x89\x11\x06\x7e\x29\x29\xb5\xa4\x50\x03\x30\x28\x9b\x8e\xa5\xb1\x44\x1b\x23\x92\x03\xb1\xc1\xa8\xe9\xc4\x44\xa9\xdd\xf1\x7d\x4b\x2c\x1c\xa4\x8c\x62\xaf\xab\x35\x59\x32\x2c\x0d\xb9\xf2\x42\x9d\xf8\xf3\x06\xbf\xee\x35\xbc\x8a\x1f\x5a\xfb\xd8\xe6\x25\x42\x79\xa1\xb5\x0e\x08\x45\xd8\x33\x82\xc6\x4d\xc7\xc6\x6c\xe3\x53\xdd\xdd\x66\x6d\xaa\xa9\x8b\x74\xaf\x5a\xbb\x1a\x6c\xf7\x3a\xdb\x5d\xb6\x2f\xeb\xd5\xea\xe9\xf9\x6e\x7a\x98\xec\xd5\x36\x0f\x46\xbb\x67\x47\x9f\x7e\xef\x7d\x97\x8d\x93\x9b\x9b\xbb\xc3\x28\x5c\x3f\x0d\xbf\x76\xab\xfc\xb7\x5f\xfb\x5b\x9d\x81\x03\xa1\xd1\xd6\x6a\x23\x7a\x42\x35\x1d\xa6\xb4\x1a\x27\x3a\xb3\xce\x0f\xc4\x95\x07\xd1\xb7\x1c\xa5\x18\x1a\x4f\x21\xf9\x2a\x4d\xfc\xcf\x91\x36\xc4\x46\x68\x75\x82\x7e\xa4\xd5\xec\xdd\x8d\x0c\xe2\xe7\x86\x57\xad\x4f\xc3\x64\x52\xce\x03\x9c\x0e\x5a\x0c\x55\xbe\xe7\x8f\xd7\xb5\x2e\x69\x2d\x49\xa4\x6e\x57\x13\xe9\x04\xbc\xd9\xb7\x50\x0a\x0d\x4c\xe6\xba\xf9\x93\xb0\x3b\x77\x24\x38\xc5\x3b\x50\xad\x54\x7e\xfe\xc7\x5c\xf8\x50\x9a\xf7\xa7\xf6\x03\xbf\x4c\xf3\x5a\xd0\xd5\x7c\x3c\x1d\x9b\x8b\x21\x84\x92\x59\xdb\x74\xc2\xdc\x6b\xa1\xd0\x38\x8f\xbe\x4c\x26\x3f\xd9\x30\x46\x9e\x49\x34\xfb\x59\x92\xc2\x4e\x13\x84\xe2\x78\x07\x1e\x54\x1e\x1e\x16\xf4\x44\x04\x3d\x82\xf7\x12\x15\x78\x1f\xa0\xba\x20\x5b\x1e\x87\x49\x34\x04\xc5\xaf\x2b\x54\xa4\x21\x21\xb7\x01\xc9\x9d\xcb\x32\xd2\x40\x78\x47\x6e\x88\x8a\xd0\x38\xa0\x55\x28\x45\x38\x68\x3a\x52\x33\x7e\xdc\x39\x3b\x7d\xff\xc1\x01\xa3\x25\x36\x9d\x6e\x46\xa4\xd5\x82\xab\xb3\xe7\x52\xc3\x50\xe0\x08\x18\xe7\x82\x84\x56\x4c\x02\xc7\x6e\xd6\x03\x9e\x25\xa9\xfd\x08\xa9\x44\x66\x11\x0c\xde\x66\x68\x09\x18\xe4\x76\xc1\xa0\x4d\xb5\xb2\xf8\xcc\x5c\x20\x66\x7e\x47\xcc\x42\xc4\xdc\x02\x22\x89\x74\x6b\x4e\x2b\xf0\xc5\xd3\xf1\x03\x9f\x8b\xe1\xe2\xec\xa1\xe2\x0b\x33\xb1\x38\x0b\x45\xd0\x23\x77\xdb\xaf\xd6\x96\x82\x08\xe2\xfa\x5c\xa9\xeb\x6e\x38\xad\xce\x34\x03\x81\x1f\xd7\x97\x54\xd3\x15\xf1\xc7\x08\xb3\x9c\x81\x8e\x60\x14\xa3\x82\x7c\x3d\x5b\x41\xda\x08\xb4\xd0\x43\x02\x54\xb7\x19\x66\xc8\xf3\x35\xa6\x81\x62\x84\xab\x62\xb5\xc3\x79\xde\xec\x2d\x45\xb5\x34\x4c\x40\xac\x2b\x71\xe6\x64\xf9\x51\xe4\x4d\x62\x44\x79\x3a\xeb\x2b\xf2\x12\x50\x0e\xbf\x79\xa7\xfc\xc3\x95\xa2\x17\xd3\x4a\x5d\xf3\xbc\x71\x6a\x04\x0a\x38\x37\x9d\x39\xe4\x7f\x76\x5a\xed\xfd\xc0\xa7\xf8\xad\x5d\xea\x79\x97\x53\x96\xe0\xab\x9d\x56\x0b\x0a\xa1\x4d\x99\x6a\x4d\xe7\xab\x9d\x23\x75\xc8\x64\xe0\x17\xad\x2f\x77\x5a\xc6\x51\x0e\x7d\x37\x14\x26\x94\x08\xc9\xb8\x04\x7f\x22\xdd\x0d\x07\x38\x23\xe6\x92\xee\xf5\x72\x9f\xa7\x2b\xdf\x79\xd1\x30\x00\x14\xcc\xdc\x74\xf6\x98\x0c\x33\xc9\x08\x39\x74\x99\x45\x0e\x5a\x15\x99\x25\x91\x20\x50\xcc\x08\x62\x66\x01\x25\x4b\x73\xa1\x15\x2a\xc4\x42\x2e\x99\x25\x08\x75\x92\x08\xfa\x08\x5c\x0c\x05\xcf\x2d\x8c\x81\x41\xa8\x95\x25\xa6\x08\x22\x16\x92\x36\x39\x9c\x6a\xde\x8a\x7c\xcd\x83\x5c\x5e\x10\x8f\x82\xd7\x66\xfa\x14\xef\x68\x8a\xbf\xd5\x8a\x81\xbf\x0a\x10\xb9\xee\xbc\x70\x3d\xb5\xf9\xc8\x70\x8b\xcf\x64\x62\x98\xea\x21\x3c\xe5\x34\x6f\xb6\xbe\x96\x38\xeb\xd1\xda\x0b\x60\x2c\x85\xbc\x35\x99\x78\x17\x98\x6a\xaf\xbd\xff\xf0\x10\xf8\xb4\xc2\xa1\x45\xed\xd7\x12\x09\x33\x53\x39\x3a\x5f\x70\x07\xca\xc8\xff\xc3\x28\x93\x09\x99\x4c\x85\x8c\x70\x3f\x33\x2c\xe7\x41\xf0\x66\x50\x7d\x8b\x97\x93\x89\xb7\x9f\xa1\x77\xa8\x4d\xc2\x08\x9c\x13\xad\x3e\x42\xa5\x06\xc7\x4c\x41\xad\x52\xd9\x84\x6a\x63\xa7\x52\xdf\xa9\x34\xe0\xa4\x73\xe9\xbc\x66\x70\x75\xea\x26\x13\x94\xf6\x8f\x4e\x38\x84\x5a\xe6\x8b\xad\xe9\x54\x2b\x15\x67\xce\x28\x0b\x85\xe3\xf5\x49\x0e\xd2\x27\xa5\xc8\x69\x9d\xea\xa7\x04\x29\xd4\x9c\x41\x9f\x71\xdf\x52\x6c\x3f\x1a\xf5\x93\x7a\xb0\xa0\xfd\x1c\xb0\x81\x5f\x10\xeb\x63\xe3\x52\x71\xf9\x83\xc5\x64\x91\xe6\xdf\x56\x50\x76\x21\x35\x42\xe7\xdb\x30\x28\x6a\x46\xce\x02\x4f\xa6\x8b\x34\x94\x5b\x45\x6f\x45\xdf\x91\x36\x03\x34\x39\x93\x90\x50\x99\xce\xac\x1c\x03\xc7\xc2\x90\xcd\xd9\x27\x01\xa6\x38\x58\x54\xdc\x4e\xad\x14\x06\x7b\x82\x2c\x9a\x21\x9a\xbf\x4a\x11\xc5\x65\xd2\x84\xea\xbd\xae\xf5\x75\x9a\xa6\xd7\xb5\x3a\xf9\xe4\xab\xf0\x7f\xc2\xb4\x25\xf8\x0a\xec\xfd\xb9\xc8\xd6\x9b\x65\xe8\x8d\xdc\x3a\x4b\xd5\x1b\xd5\x3b\x78\xfb\x27\xe4\xd8\xdb\x92\x83\xfe\x1f\x09\x76\xf1\xd5\x86\x46\xa4\x04\xd6\x84\x0b\x67\x40\xcd\xd1\xeb\xdf\x66\x68\xc6\xc5\x81\xb6\x7c\x75\x37\xbc\x86\x57\xf5\xac\x14\x49\x71\xc6\xeb\xaf\x3c\xc3\xee\x47\xdf\xf8\x7d\x2d\xa6\xaf\x47\x15\x69\x3b\x1d\xdb\x50\x7b\x97\x69\xd6\xf7\xef\xc7\xf5\xbd\xf5\xb3\x5f\x53\x96\xe8\xc3\xeb\xf1\xc6\xa7\x93\xeb\x5f\xd4\xc1\x7a\xbb\xdb\xbd\xbe\xb9\xc2\xd1\xfa\x99\xd9\xfb\xc6\x2e\x06\x51\xff\xe5\x33\x6c\xe0\x97\xbe\xbe\xe6\xf8\xaa\xc3\x6b\xaa\xd3\x14\x8d\xd7\xb7\x9f\xab\x5e\x75\xd3\xab\xf8\x5c\x58\xf2\xb3\x84\xcf\x24\x2f\x07\x73\xbe\x79\xb0\x7d\x71\x34\xec\xb6\xc7\xdf\x0f\x8f\x75\x44\xeb\xb5\xe4\xb8\x7b\xc4\x0e\x7e\x97\x5c\x0e\xdb\xdb\xed\xb3\x9b\x71\x43\x6d\xdc\x5f\x6f\xdf\xdf\x5f\x52\xd2\xde\xb8\x1a\x58\x7e\x7e\x71\x3d\xd4\x77\x27\x91\xd6\xbb\xfa\xbf\x0a\xe6\x07\x6e\x18\xfa\xcb\x17\x0c\xab\xc3\x39\xeb\x5d\x5f\x0c\xb3\xdd\xcb\xaf\xd5\xfb\xad\xe3\xfe\xd1\x97\x41\x76\x76\xb5\xf5\x6d\xb4\x55\xa9\xaf\xc7\x9f\x36\x1a\x5f\xcc\xfa\xe6\xf9\x97\xed\xab\xe1\x4d\xff\xfb\xc1\x46\x3b\xcd\x36\x2f\xd3\xad\x46\x7f\xeb\x97\xd8\x1f\x5c\x54\x8e\x7f\x6b\xff\x60\x38\x8f\xd8\xfb\xe9\xfd\xbb\x7f\xae\xdc\xf7\xff\xeb\xdd\x87\xd9\xe9\xff\xfd\x87\xb9\x7a\x94\xa9\xb0\xd8\xd4\x3d\x1e\x8c\x97\x6e\x05\x46\x42\x71\x3d\xf2\xa4\x0e\x8b\xdd\x9f\x67\x91\x99\x30\x86\xf5\x26\xbc\xfb\x7b\x54\xec\xea\x9a\x7d\xab\xd5\xbb\xe7\xb7\x05\x53\xcf\x02\xbf\x5c\x32\x81\x5f\x5e\x1b\xfd\x3b\x00\x00\xff\xff\xf3\xb1\x56\x94\x47\x12\x00\x00")

func stateHtmlTmplBytes() ([]byte, error) {
	return bindataRead(
		_stateHtmlTmpl,
		"state.html.tmpl",
	)
}

func stateHtmlTmpl() (*asset, error) {
	bytes, err := stateHtmlTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "state.html.tmpl", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x64, 0x43, 0xd6, 0x12, 0xe1, 0xef, 0xef, 0x82, 0x3f, 0x30, 0xf8, 0x98, 0x5e, 0x85, 0x77, 0xa4, 0x1d, 0x2e, 0x40, 0xe3, 0x57, 0x81, 0xea, 0x52, 0x64, 0x5a, 0x66, 0x55, 0xed, 0xa9, 0xe9, 0xda}}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	canonicalName := strings.ReplaceAll(name, "\\", "/")
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
	canonicalName := strings.ReplaceAll(name, "\\", "/")
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
	canonicalName := strings.ReplaceAll(name, "\\", "/")
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
	"state.html.tmpl": stateHtmlTmpl,
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
		canonicalName := strings.ReplaceAll(name, "\\", "/")
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
	"state.html.tmpl": {stateHtmlTmpl, map[string]*bintree{}},
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
	canonicalName := strings.ReplaceAll(name, "\\", "/")
	return filepath.Join(append([]string{dir}, strings.Split(canonicalName, "/")...)...)
}
