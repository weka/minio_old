/*
 * MinIO Cloud Storage, (C) 2016, 2017, 2018 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import "C"
import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"os"
	pathutil "path"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/minio/minio/cmd/logger"
	"github.com/minio/minio/pkg/lock"
)

const(
	AT_SYMLINK_FOLLOW = 0x400
	AT_FDCWD = -100
)


type makefileParam struct
{
	InodeId uint64
	InodeSupplemental uint64
	Mode int32
	Filename [256]uint8
}

func ioctl(fd uintptr, operation int32, param uintptr) (result uintptr, err error){
	result, _, err = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(operation), param)
	return result, err
}

func wekaIoctl(operation int32, root string, filename string, mode int32) (err error){
	if !GlobalIsFastFS {
		return unix.ENOTSUP
	}

	f, err := os.Open(root)
	if err != nil {
		return err
	}
	defer f.Close()

	var fd = f.Fd()
	var param makefileParam
	copy(param.Filename[:], filename)
	param.Mode = mode

	_, err = ioctl(fd, operation, uintptr(unsafe.Pointer(&param)))

	return err
}

func wekaLinkFileFast(fullPath string, file os.File) (err error) {
	var LINK = int32(0x4C494E4B) // = 'LINK'
	var STAT = int32(0x53544154) // = 'STAT'

	dirname, filename := pathutil.Split(fullPath)

	if !GlobalIsFastFS {
		return fmt.Errorf("the specific ioctl operation is unsupported on the current filesystem")
	}

	f, err := os.Open(dirname)
	if err != nil {
		return err
	}
	defer f.Close()

	var fd = f.Fd()
	var param makefileParam

	// STAT fills dir inode data in makefile param
	_, err = ioctl(fd, STAT, uintptr(unsafe.Pointer(&param)))

	// No reason to check error + set GlobalIsFastFS flag:
	// no way link is the first op in the sequence, if fast op not supported
	// we're not supposed to get here.
	param.Mode = 0
	copy(param.Filename[:], filename)
	_, err = ioctl(file.Fd(), LINK, uintptr(unsafe.Pointer(&param)))

	return nil
}

func wekaMakeInodeFast(root string, filename string, mode int32) (err error) {
    var MKND = int32(0x4D4B4E44) // = 'MKND'
	err = wekaIoctl(MKND, root, filename, mode)
	return err
}

func WekaDeleteFileFast(root string, filename string) (err error) {
	var ULNK = int32(0x554C4E4B) // = 'ULNK'
	err = wekaIoctl(ULNK, root, filename, 0)
	return err
}

func fsMakeInodeFast(filePath string, mode int32) (err error) {
    dir, file := pathutil.Split(filePath)
    err = wekaMakeInodeFast(dir, file, mode)
    return err
}

func fsDeleteFileFast(filePath string) (err error) {
	dir, file := pathutil.Split(filePath)
	err = WekaDeleteFileFast(dir, file)
	return err
}

// Removes only the file at given path does not remove
// any parent directories, handles long paths for
// windows automatically.
func fsRemoveFile(ctx context.Context, filePath string) (err error) {
	if filePath == "" {
		logger.LogIf(ctx, errInvalidArgument)
		return errInvalidArgument
	}

	if err = checkPathLength(filePath); err != nil {
		logger.LogIf(ctx, err)
		return err
	}

	if err = os.Remove(filePath); err != nil {
		if err = osErrToFileErr(err); err != errFileNotFound {
			logger.LogIf(ctx, err)
		}
	}

	return err
}

// Removes all files and folders at a given path, handles
// long paths for windows automatically.
func fsRemoveAll(ctx context.Context, dirPath string) (err error) {
	if dirPath == "" {
		logger.LogIf(ctx, errInvalidArgument)
		return errInvalidArgument
	}

	if err = checkPathLength(dirPath); err != nil {
		logger.LogIf(ctx, err)
		return err
	}

	if err = removeAll(dirPath); err != nil {
		if osIsPermission(err) {
			logger.LogIf(ctx, errVolumeAccessDenied)
			return errVolumeAccessDenied
		} else if isSysErrNotEmpty(err) {
			logger.LogIf(ctx, errVolumeNotEmpty)
			return errVolumeNotEmpty
		}
		logger.LogIf(ctx, err)
		return err
	}

	return nil
}

// Removes a directory only if its empty, handles long
// paths for windows automatically.
func fsRemoveDir(ctx context.Context, dirPath string) (err error) {
	if dirPath == "" {
		logger.LogIf(ctx, errInvalidArgument)
		return errInvalidArgument
	}

	if err = checkPathLength(dirPath); err != nil {
		logger.LogIf(ctx, err)
		return err
	}

	if err = os.Remove((dirPath)); err != nil {
		if osIsNotExist(err) {
			return errVolumeNotFound
		} else if isSysErrNotEmpty(err) {
			return errVolumeNotEmpty
		}
		logger.LogIf(ctx, err)
		return err
	}

	return nil
}

// Creates a new directory, parent dir should exist
// otherwise returns an error. If directory already
// exists returns an error. Windows long paths
// are handled automatically.
func fsMkdir(ctx context.Context, dirPath string) (err error) {
	if dirPath == "" {
		logger.LogIf(ctx, errInvalidArgument)
		return errInvalidArgument
	}

	if err = checkPathLength(dirPath); err != nil {
		logger.LogIf(ctx, err)
		return err
	}

	if err = os.Mkdir((dirPath), 0777); err != nil {
		switch {
		case osIsExist(err):
			return errVolumeExists
		case osIsPermission(err):
			logger.LogIf(ctx, errDiskAccessDenied)
			return errDiskAccessDenied
		case isSysErrNotDir(err):
			// File path cannot be verified since
			// one of the parents is a file.
			logger.LogIf(ctx, errDiskAccessDenied)
			return errDiskAccessDenied
		case isSysErrPathNotFound(err):
			// Add specific case for windows.
			logger.LogIf(ctx, errDiskAccessDenied)
			return errDiskAccessDenied
		default:
			logger.LogIf(ctx, err)
			return err
		}
	}

	return nil
}

// fsStat is a low level call which validates input arguments
// and checks input length upto supported maximum. Does
// not perform any higher layer interpretation of files v/s
// directories. For higher level interpretation look at
// fsStatFileDir, fsStatFile, fsStatDir.
func fsStat(ctx context.Context, statLoc string) (os.FileInfo, error) {
	if statLoc == "" {
		logger.LogIf(ctx, errInvalidArgument)
		return nil, errInvalidArgument
	}
	if err := checkPathLength(statLoc); err != nil {
		logger.LogIf(ctx, err)
		return nil, err
	}
	fi, err := os.Stat(statLoc)
	if err != nil {
		return nil, err
	}

	return fi, nil
}

// Lookup if volume exists, returns volume attributes upon success.
func fsStatVolume(ctx context.Context, volume string) (os.FileInfo, error) {
	fi, err := fsStat(ctx, volume)
	if err != nil {
		if osIsNotExist(err) {
			return nil, errVolumeNotFound
		} else if osIsPermission(err) {
			return nil, errVolumeAccessDenied
		}
		return nil, err
	}

	// if this is a symlink, it must point to a directory
	if fi.Mode() & os.ModeSymlink == os.ModeSymlink {
		target, err := os.Readlink(volume)
		if err != nil {
			fi, err := fsStat(ctx, target)
			if err != nil {
				if fi.IsDir() {
					return fi, err
				} else {
					return nil, errVolumeAccessDenied
				}
			}
		} else {
			return nil, errVolumeAccessDenied
		}
	}

	if !fi.IsDir() {
		return nil, errVolumeAccessDenied
	}

	return fi, err
}

// Lookup if directory exists, returns directory attributes upon success.
func fsStatDir(ctx context.Context, statDir string) (os.FileInfo, error) {
	fi, err := fsStat(ctx, statDir)
	if err != nil {
		err = osErrToFileErr(err)
		if err != errFileNotFound {
			logger.LogIf(ctx, err)
		}
		return nil, err
	}
	if !fi.IsDir() {
		return nil, errFileNotFound
	}
	return fi, nil
}

// Lookup if file exists, returns file attributes upon success.
func fsStatFile(ctx context.Context, statFile string) (os.FileInfo, error) {
	fi, err := fsStat(ctx, statFile)
	if err != nil {
		err = osErrToFileErr(err)
		if err != errFileNotFound {
			logger.LogIf(ctx, err)
		}
		return nil, err
	}
	if fi.IsDir() {
		return nil, errFileNotFound
	}
	return fi, nil
}

// Returns if the filePath is a regular file.
func fsIsFile(ctx context.Context, filePath string) bool {
	fi, err := fsStat(ctx, filePath)
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular()
}

// Opens the file at given path, optionally from an offset. Upon success returns
// a readable stream and the size of the readable stream.
func fsOpenFile(ctx context.Context, readPath string, offset int64) (io.ReadCloser, int64, error) {
	if readPath == "" || offset < 0 {
		logger.LogIf(ctx, errInvalidArgument)
		return nil, 0, errInvalidArgument
	}
	if err := checkPathLength(readPath); err != nil {
		logger.LogIf(ctx, err)
		return nil, 0, err
	}

	fr, err := os.Open(readPath)
	if err != nil {
		return nil, 0, osErrToFileErr(err)
	}

	// Stat to get the size of the file at path.
	st, err := fr.Stat()
	if err != nil {
		err = osErrToFileErr(err)
		if err != errFileNotFound {
			logger.LogIf(ctx, err)
		}
		return nil, 0, err
	}

	// Verify if its not a regular file, since subsequent Seek is undefined.
	if !st.Mode().IsRegular() {
		return nil, 0, errIsNotRegular
	}

	// Seek to the requested offset.
	if offset > 0 {
		_, err = fr.Seek(offset, io.SeekStart)
		if err != nil {
			logger.LogIf(ctx, err)
			return nil, 0, err
		}
	}

	// Success.
	return fr, st.Size(), nil
}

func fsCreateFile(ctx context.Context, filePath string, reader io.Reader, buf []byte, fallocSize int64) (int64, error) {
	bytesWritten, err, _ := createFile(ctx, filePath, reader, buf, fallocSize, false)
	return bytesWritten, err
}

func fsCreateAndGetFile(ctx context.Context, tmpPartDir string, reader io.Reader, buf []byte, fallocSize int64) (int64, error, os.File) {
	return createFile(ctx, tmpPartDir, reader, buf, fallocSize, true)
}

// Creates a file and copies data from incoming reader. Staging buffer is used by io.CopyBuffer.
func createFile(ctx context.Context, filePath string, reader io.Reader, buf []byte, fallocSize int64, getFile bool) (int64, error, os.File) {
	if filePath == "" || reader == nil {
		logger.LogIf(ctx, errInvalidArgument)
		return 0, errInvalidArgument, os.File {}
	}

	if err := checkPathLength(filePath); err != nil {
		logger.LogIf(ctx, err)
		return 0, err, os.File {}
	}

	if err := mkdirAll(pathutil.Dir(filePath), 0777); err != nil {
		switch {
		case osIsPermission(err):
			return 0, errFileAccessDenied, os.File {}
		case osIsExist(err):
			return 0, errFileAccessDenied, os.File {}
		case isSysErrIO(err):
			return 0, errFaultyDisk, os.File {}
		case isSysErrInvalidArg(err):
			return 0, errUnsupportedDisk, os.File {}
		case isSysErrNoSpace(err):
			return 0, errDiskFull, os.File {}
		}
		return 0, err, os.File {}
	}

	var writer *os.File
	var err error

	var flags int

	if globalFSODirect {
		flags = flags | syscall.O_DIRECT
	}
  
	if getFile && GlobalFSOTmpfile {
		flags = flags | os.O_WRONLY | unix.O_TMPFILE
		writer, err = lock.Open(pathutil.Dir(filePath), flags, 0666)
		if err != nil {
			return 0, osErrToFileErr(err), os.File {}
		}
	} else {
		flags := os.O_CREATE | os.O_WRONLY
		if globalFSOSync {
			flags = flags | os.O_SYNC
		}
		writer, err = lock.Open(filePath, flags, 0666)
		if err != nil {
			if writer != nil {
				_ = writer.Close()
			}
			return 0, osErrToFileErr(err), os.File {}
		}
	}

	// Fallocate only if the size is final object is known.
	if fallocSize > 0 {
		if err = fsFAllocate(int(writer.Fd()), 0, fallocSize); err != nil {
			logger.LogIf(ctx, err)
			if writer != nil {
				_ = writer.Close()
			}
			return 0, err, os.File {}
		}
	}

	var bytesWritten int64
	if buf != nil {
		bytesWritten, err = io.CopyBuffer(struct {io.Writer} {writer}, struct {io.Reader} {reader}, buf)
		if err != nil {
			if err != io.ErrUnexpectedEOF {
				if writer != nil {
					_ = writer.Close()
				}
				logger.LogIf(ctx, err)
			}
			return 0, err, os.File {}
		}
	} else {
		bytesWritten, err = io.Copy(writer, reader)
		if err != nil {
			logger.LogIf(ctx, err)
			if writer != nil {
				_ = writer.Close()
			}
			return 0, err, os.File {}
		}
	}

	return bytesWritten, nil, *writer
}

// fsFAllocate is similar to Fallocate but provides a convenient
// wrapper to handle various operating system specific errors.
func fsFAllocate(fd int, offset int64, len int64) (err error) {
	e := Fallocate(fd, offset, len)
	if e != nil {
		switch {
		case isSysErrNoSpace(e):
			err = errDiskFull
		case isSysErrNoSys(e) || isSysErrOpNotSupported(e):
			// Ignore errors when Fallocate is not supported in the current system
		case isSysErrInvalidArg(e):
			// Workaround for Windows Docker Engine 19.03.8.
			// See https://github.com/minio/minio/issues/9726
		case isSysErrIO(e):
			err = e
		default:
			// For errors: EBADF, EINTR, EINVAL, ENODEV, EPERM, ESPIPE  and ETXTBSY
			// Appending was failed anyway, returns unexpected error
			err = errUnexpected
		}
		return err
	}

	return nil
}

// Renames source path to destination path, fails if the destination path
// parents are not already created.
func fsSimpleRenameFile(ctx context.Context, sourcePath, destPath string) error {
	if err := checkPathLength(sourcePath); err != nil {
		logger.LogIf(ctx, err)
		return err
	}
	if err := checkPathLength(destPath); err != nil {
		logger.LogIf(ctx, err)
		return err
	}

	if err := os.Rename(sourcePath, destPath); err != nil {
		logger.LogIf(ctx, err)
		return osErrToFileErr(err)
	}

	return nil
}

func linkat(olddirfd int, oldpath string, newdirfd int, newpath string, flags int) (err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(oldpath)
	if err != nil {
		return
	}
	var _p1 *byte
	_p1, err = syscall.BytePtrFromString(newpath)
	if err != nil {
		return
	}
	_, _, e1 := syscall.Syscall6(syscall.SYS_LINKAT, uintptr(olddirfd), uintptr(unsafe.Pointer(_p0)), uintptr(newdirfd), uintptr(unsafe.Pointer(_p1)), uintptr(flags), 0)
	if e1 != 0 {
		err = error(e1)
	}
	return
}

func fsLinkat(ctx context.Context, oldDirFD int64, oldPath string, newDirFD int64, newPath string, flags int) error {
	if err := checkPathLength(oldPath); err != nil {
		logger.LogIf(ctx, err)
		return err
	}
	if err := checkPathLength(newPath); err != nil {
		logger.LogIf(ctx, err)
		return err
	}

	if err := linkat(int(oldDirFD), oldPath , int(newDirFD), newPath, flags); err != nil {
		logger.LogIf(ctx, err)
		return osErrToFileErr(err)
	}

	return nil
}

// Renames source path to destination path, creates all the
// missing parents if they don't exist.
func fsRenameFile(ctx context.Context, sourcePath, destPath string) error {
	if err := checkPathLength(sourcePath); err != nil {
		logger.LogIf(ctx, err)
		return err
	}
	if err := checkPathLength(destPath); err != nil {
		logger.LogIf(ctx, err)
		return err
	}

	if err := renameAll(sourcePath, destPath); err != nil {
		logger.LogIf(ctx, err)
		return err
	}

	return nil
}

// fsDeleteFile is a wrapper for deleteFile(), after checking the path length.
func fsDeleteFile(ctx context.Context, basePath, deletePath string) error {
	if err := checkPathLength(basePath); err != nil {
		logger.LogIf(ctx, err)
		return err
	}

	if err := checkPathLength(deletePath); err != nil {
		logger.LogIf(ctx, err)
		return err
	}

	// ignore that error in case we're not running on a supported fs
	err := fsDeleteFileFast(deletePath)
	if errors.Is(err, syscall.Errno(0)) {
		return nil
	}

	if err := deleteFile(basePath, deletePath, false); err != nil {
		if err != errFileNotFound {
			logger.LogIf(ctx, err)
		}
		return err
	}
	return nil
}

// fsRemoveMeta safely removes a locked file and takes care of Windows special case
func fsRemoveMeta(ctx context.Context, basePath, deletePath, tmpDir string) error {
	// Special case for windows please read through.
	if runtime.GOOS == globalWindowsOSName {
		// Ordinarily windows does not permit deletion or renaming of files still
		// in use, but if all open handles to that file were opened with FILE_SHARE_DELETE
		// then it can permit renames and deletions of open files.
		//
		// There are however some gotchas with this, and it is worth listing them here.
		// Firstly, Windows never allows you to really delete an open file, rather it is
		// flagged as delete pending and its entry in its directory remains visible
		// (though no new file handles may be opened to it) and when the very last
		// open handle to the file in the system is closed, only then is it truly
		// deleted. Well, actually only sort of truly deleted, because Windows only
		// appears to remove the file entry from the directory, but in fact that
		// entry is merely hidden and actually still exists and attempting to create
		// a file with the same name will return an access denied error. How long it
		// silently exists for depends on a range of factors, but put it this way:
		// if your code loops creating and deleting the same file name as you might
		// when operating a lock file, you're going to see lots of random spurious
		// access denied errors and truly dismal lock file performance compared to POSIX.
		//
		// We work-around these un-POSIX file semantics by taking a dual step to
		// deleting files. Firstly, it renames the file to tmp location into multipartTmpBucket
		// We always open files with FILE_SHARE_DELETE permission enabled, with that
		// flag Windows permits renaming and deletion, and because the name was changed
		// to a very random name somewhere not in its origin directory before deletion,
		// you don't see those unexpected random errors when creating files with the
		// same name as a recently deleted file as you do anywhere else on Windows.
		// Because the file is probably not in its original containing directory any more,
		// deletions of that directory will not fail with "directory not empty" as they
		// otherwise normally would either.

		tmpPath := pathJoin(tmpDir, mustGetUUID())

		fsRenameFile(ctx, deletePath, tmpPath)

		// Proceed to deleting the directory if empty
		fsDeleteFile(ctx, basePath, pathutil.Dir(deletePath))

		// Finally delete the renamed file.
		return fsDeleteFile(ctx, tmpDir, tmpPath)
	}
	return fsDeleteFile(ctx, basePath, deletePath)
}
