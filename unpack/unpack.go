// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/* Originally copyrighted as
// Copyright © 2011-12 Qtrac Ltd.
//
// This program or package and any associated files are licensed under the
// Apache License, Version 2.0 (the "License"); you may not use these files
// except in compliance with the License. You can get a copy of the License
// at: http://www.apache.org/licenses/LICENSE-2.0.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
*/
/*
 Code adapted and enhanced from examples to the book:
 Programming in Go by Mark Summerfield
 http://www.qtrac.eu/gobook.html

 Original author: Mark Summerfield
 Converted to package by Giuseppe Maxia in 2018

 The original code was a stand-alone program, and it
 had a few bugs:
 * when extracting from a tar file: when there
 isn't a separate item for each directory, the
 extraction fails.
 * The attributes of the files were not reproduced
 in the extracted files.
 This code fixes those problems and introduces a
 destination directory and verbosity
 levels for the extraction

*/

package unpack

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/pkg/errors"
	"github.com/xi2/xz"
)

const (
	SILENT  = iota // No output
	VERBOSE        // Minimal feedback about extraction operations
	CHATTY         // Full details of what is being extracted
)

var Verbose int

func condPrint(s string, nl bool, level int) {
	if Verbose >= level {
		if nl {
			fmt.Println(s)
		} else {
			fmt.Print(s)
		}
	}
}

func validSuffix(filename string) bool {
	for _, suffix := range []string{globals.TgzExt, globals.TarExt, globals.TarGzExt, globals.TarXzExt} {
		if strings.HasSuffix(filename, suffix) {
			return true
		}
	}
	return false
}

func UnpackXzTar(filename string, destination string, verbosityLevel int) (err error) {
	Verbose = verbosityLevel
	if !common.FileExists(filename) {
		return fmt.Errorf("file %s not found", filename)
	}
	if !common.DirExists(destination) {
		return fmt.Errorf("directory %s not found", destination)
	}
	filename, err = common.AbsolutePath(filename)
	if err != nil {
		return err
	}
	err = os.Chdir(destination)
	if err != nil {
		return errors.Wrapf(err, "error changing directory to %s", destination)
	}
	// #nosec G304
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	// Create an xz Reader
	r, err := xz.NewReader(f, 0)
	if err != nil {
		return err
	}
	// Create a tar Reader
	tr := tar.NewReader(r)
	return unpackTarFiles(tr)
}

func UnpackTar(filename string, destination string, verbosityLevel int) (err error) {
	Verbose = verbosityLevel
	f, err := os.Stat(destination)
	if os.IsNotExist(err) {
		return fmt.Errorf("destination directory '%s' does not exist", destination)
	}
	filemode := f.Mode()
	if !filemode.IsDir() {
		return fmt.Errorf("destination '%s' is not a directory", destination)
	}
	if !validSuffix(filename) {
		return fmt.Errorf("unrecognized archive suffix")
	}
	var file *os.File
	// #nosec G304
	if file, err = os.Open(filename); err != nil {
		return err
	}
	defer file.Close()
	err = os.Chdir(destination)
	if err != nil {
		return errors.Wrapf(err, "error changing directory to %s", destination)
	}
	var fileReader io.Reader = file
	var decompressor *gzip.Reader
	if strings.HasSuffix(filename, globals.GzExt) {
		if decompressor, err = gzip.NewReader(file); err != nil {
			return err
		}
		defer decompressor.Close()
	}
	var reader *tar.Reader
	if decompressor != nil {
		reader = tar.NewReader(decompressor)
	} else {
		reader = tar.NewReader(fileReader)
	}
	return unpackTarFiles(reader)
}

func unpackTarFiles(reader *tar.Reader) (err error) {
	var header *tar.Header
	var count int = 0
	var reSlash = regexp.MustCompile(`/.*`)

	innerDir := ""
	for {
		if header, err = reader.Next(); err != nil {
			if err == io.EOF {
				condPrint("Files ", false, CHATTY)
				condPrint(strconv.Itoa(count), true, 1)
				return nil // OK
			}
			return err
		}
		// cond_print(fmt.Sprintf("%#v\n", header), true, CHATTY)
		/*
			tar.Header{
				Typeflag:0x30,
				Name:"mysql-8.0.11-macos10.13-x86_64/docs/INFO_SRC",
				Linkname:"",
				Size:185,
				Mode:420,
				Uid:7161,
				Gid:10,
				Uname:"pb2user",
				Gname:"owner",
				ModTime:time.Time{wall:0x0, ext:63658769207, loc:(*time.Location)(0x13730e0)},
				AccessTime:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)},
				ChangeTime:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)},
				Devmajor:0, Devminor:0,
				Xattrs:map[string]string(nil),
				PAXRecords:map[string]string(nil),
				Format:0}
			tar.Header{
				Typeflag:0x32,
				Name:"mysql-8.0.11-macos10.13-x86_64/lib/libssl.dylib",
				Linkname:"libssl.1.0.0.dylib",
				Size:0,
				Mode:493,
				Uid:7161,
				Gid:10,
				Uname:"pb2user",
				Gname:"owner",
				ModTime:time.Time{wall:0x0, ext:63658772525, loc:(*time.Location)(0x13730e0)},
				AccessTime:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)},
				ChangeTime:time.Time{wall:0x0, ext:0, loc:(*time.Location)(nil)},
				Devmajor:0,
				Devminor:0,
				Xattrs:map[string]string(nil),
				PAXRecords:map[string]string(nil),
				Format:0}
		*/
		filemode := os.FileMode(header.Mode)
		filename := sanitizedName(header.Name)
		fileDir := path.Dir(filename)
		upperDir := reSlash.ReplaceAllString(fileDir, "")
		if innerDir != "" {
			if upperDir != innerDir {
				return fmt.Errorf("found more than one directory inside the tarball\n"+
					"<%s> and <%s>", upperDir, innerDir)
			}
		} else {
			innerDir = upperDir
		}

		if _, err = os.Stat(fileDir); os.IsNotExist(err) {
			if err = os.MkdirAll(fileDir, globals.PublicDirectoryAttr); err != nil {
				return err
			}
			condPrint(" + "+fileDir+" ", true, CHATTY)
		}
		if header.Typeflag == 0 {
			header.Typeflag = tar.TypeReg
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(filename, globals.PublicDirectoryAttr); err != nil {
				return err
			}
		case tar.TypeReg:
			if err = unpackTarFile(filename, reader); err != nil {
				return err
			}
			err = os.Chmod(filename, filemode)
			if err != nil {
				return err
			}
			count++
			condPrint(filename, true, CHATTY)
			if count%10 == 0 {
				mark := "."
				if count%100 == 0 {
					mark = strconv.Itoa(count)
				}
				if Verbose < CHATTY {
					condPrint(mark, false, 1)
				}
			}
		case tar.TypeSymlink:
			if header.Linkname != "" {
				condPrint(fmt.Sprintf("%s -> %s", filename, header.Linkname), true, CHATTY)
				err = os.Symlink(header.Linkname, filename)
				if err != nil {
					return fmt.Errorf("%#v\n#ERROR: %s", header, err)
				}
			} else {
				return fmt.Errorf("file %s is a symlink, but no link information was provided", filename)
			}
		}
	}
	// return nil
}

func unpackTarFile(filename string,
	reader *tar.Reader) (err error) {
	var writer *os.File
	if writer, err = os.Create(filename); err != nil {
		return err
	}
	defer writer.Close()
	if _, err = io.Copy(writer, reader); err != nil {
		return err
	}
	return nil
}

func sanitizedName(filename string) string {
	if len(filename) > 1 && filename[1] == ':' {
		filename = filename[2:]
	}
	filename = strings.TrimLeft(filename, "\\/.")
	filename = strings.Replace(filename, "../", "", -1)
	return strings.Replace(filename, "..\\", "", -1)
}

func VerifyTarFile(fileName string) error {
	if !validSuffix(fileName) {
		return fmt.Errorf("unrecognized archive suffix %s", fileName)
	}
	var file *os.File
	var err error
	// #nosec G304
	if file, err = os.Open(fileName); err != nil {
		return fmt.Errorf("[open file Validation] %s", err)
	}
	defer file.Close()
	var fileReader io.Reader = file
	var decompressor *gzip.Reader
	var xzDecompressor *xz.Reader

	if strings.HasSuffix(fileName, globals.GzExt) {
		if decompressor, err = gzip.NewReader(file); err != nil {
			return fmt.Errorf("[gz Validation] %s", err)
		}
		defer decompressor.Close()
	} else {
		if strings.HasSuffix(fileName, globals.TarXzExt) {
			if xzDecompressor, err = xz.NewReader(file, 0); err != nil {
				return fmt.Errorf("[xz Validation] %s", err)
			}
		}
	}
	var reader *tar.Reader
	if decompressor != nil {
		reader = tar.NewReader(decompressor)
	} else {
		if xzDecompressor != nil {
			reader = tar.NewReader(xzDecompressor)
		} else {
			reader = tar.NewReader(fileReader)
		}
	}
	var header *tar.Header
	expectedDirName := common.BaseName(fileName)
	reExt := regexp.MustCompile(`\.(?:tar(?:\.gz|\.xz)?)$`)
	expectedDirName = reExt.ReplaceAllString(expectedDirName, "")

	if header, err = reader.Next(); err != nil {
		if err == io.EOF {
			return fmt.Errorf("[EOF Validation] file %s is empty", fileName)
		}
		return fmt.Errorf("[header validation] %s", err)
	}
	innerFileName := sanitizedName(header.Name)
	fileDir := path.Dir(innerFileName)

	reSlash := regexp.MustCompile(`/.*`)
	fileDir = reSlash.ReplaceAllString(fileDir, "")

	if fileDir != expectedDirName {
		return fmt.Errorf("inner directory name different from tarball name\n"+
			"Expected: %s - Found: %s", expectedDirName, fileDir)
	}
	return nil
}
