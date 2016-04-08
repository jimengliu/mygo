package main

import (
	"log"
	"os"
	"syscall"
)

const (
	SEEK_DATA = 3 // seek to the next data
	SEEK_HOLE = 4 // seek to the next hole
)

// get the file system block size
func GetFileSystemBlockSize() (int64, error) {
	var stat syscall.Stat_t
	err := syscall.Stat(os.DevNull, &stat)
	return stat.Blksize, err
}

// globals
var blockSize int64

func main() {
	if len(os.Args) != 3 {
		panic("need child and parent file names, child first and parent second")
	}
	var err error
	blockSize, err = GetFileSystemBlockSize()
	if err != nil {
		panic("can't get FS block size, error: " + err.Error())
	}

	childFileName := os.Args[1]
	parentFileName := os.Args[2]

	childFInfo, infoErr := os.Stat(childFileName)
	if infoErr != nil {
		panic("os.Stat(childFileName) failed, error: " + err.Error())
	}
	parentFInfo, infoErr := os.Stat(parentFileName)
	if infoErr != nil {
		panic("os.Stat(parentFileName) failed, error: " + err.Error())
	}
	
	// ensure no directory
	if childFInfo.IsDir() || parentFInfo.IsDir() {
	    panic("at least one file is directory, not a normal file")
	}

	// ensure file sizes are equal
	if childFInfo.Size() != parentFInfo.Size() {
	    panic("file sizes are not equal")
	}

	// open child and parent files
	childFile, err := os.Open(childFileName)
	if err != nil {
		panic("Failed to open childFile, error: " + err.Error())
	}
	defer childFile.Close()

	parentFile, err := os.OpenFile(parentFileName, os.O_RDWR, 0)
	if err != nil {
		panic("Failed to open parentFile, error: " + err.Error())
	}
	defer parentFile.Close()

	err = coalesce(parentFile, childFile)
	if err != nil {
		panic("Failed to open parentFile, error: " + err.Error())
	}
}

func coalesce(parentFile *os.File, childFile *os.File) error {
	var data, hole int64
	var err error
	for {
		data, err = syscall.Seek(int(childFile.Fd()), hole, SEEK_DATA)
		if err != nil {
			// reaches EOF
			errno := err.(syscall.Errno)
			if errno == syscall.ENXIO {
				break
			} else {
				// unexpected errors
				log.Fatal("Failed to syscall.Seek SEEK_DATA")
				return err
			}
		}
		hole, err = syscall.Seek(int(childFile.Fd()), data, SEEK_HOLE)
		if err != nil {
			log.Fatal("Failed to syscall.Seek SEEK_HOLE")
			return err
		}

		// now we have a data start offset and length(hole - data)
		// let's read from child and write to parent file block by block
		_, err = parentFile.Seek(data, os.SEEK_SET)
		if err != nil {
			log.Fatal("Failed to os.Seek os.SEEK_SET")
			return err
		}

		offset := data
		buffer := make([]byte, blockSize)
		for offset != hole {
			// read a block from child, maybe use bufio or Reader stream
			n, err := childFile.ReadAt(buffer, offset)
			if n != len(buffer) || err != nil {
				log.Fatal("Failed to read from childFile")
				return err
			}
			// write a block to parent
			n, err = parentFile.WriteAt(buffer, offset)
			if n != len(buffer) || err != nil {
				log.Fatal("Failed to write to parentFile")
				return err
			}
			parentFile.Sync()
			offset += int64(n)
		}
	}

	return nil
}
