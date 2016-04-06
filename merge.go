package main

import (
	"fmt"
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
		log.Fatal("need child and parent file names, child first and parent second")
		return
	}
	var err error
	blockSize, err = GetFileSystemBlockSize()
	if err != nil {
		log.Fatal("can't get FS block size", err)
		return
	}

	childFileName := os.Args[1]
	parentFileName := os.Args[2]
	childFInfo, infoErr := os.Stat(childFileName)
	if infoErr != nil {
		log.Fatal("os.Stat(childFileName) failed", infoErr)
		return
	}
	parentFInfo, infoErr := os.Stat(parentFileName)
	if infoErr != nil {
		log.Fatal("os.Stat(parentFileName) failed", infoErr)
		return
	}
	
	// ensure no directory
	if childFInfo.IsDir() || parentFInfo.IsDir() {
	    log.Fatal("at least one file is directory, not a normal file")
	}

	// ensure file sizes are equal
	if childFInfo.Size() != parentFInfo.Size() {
	    log.Fatal("file sizes are not equal")
	}
	
	// open child and parent files
	childFile, err := os.Open(childFileName)
	if err != nil {
		log.Fatal("Failed to open childFile")
	}
	defer childFile.Close()

	parentFile, err := os.OpenFile(parentFileName, os.O_RDWR, 0)
	if err != nil {
		log.Fatal("Failed to open parentFile")
	}
	defer parentFile.Close()

	Coalesce(parentFile, childFile)
}

func Coalesce(parentFile *os.File, childFile *os.File) error {
	var data, hole int64
	//var datahole []int64
	var err error
	for {
		data, err = syscall.Seek(int(childFile.Fd()), hole, SEEK_DATA)
		if err == nil {
			hole, err = syscall.Seek(int(childFile.Fd()), data, SEEK_HOLE)
			if err == nil {
    			// let's read from child and write to parent file block by block
        		_, err = parentFile.Seek(data, os.SEEK_SET)

        		offset := data
        		buffer := make([]byte, blockSize)
        		for offset != hole {
        			// read 4K from child, maybe use bufio or Reader stream
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
        			fmt.Println("parentFile.WriteAt(buffer, offset)", offset)
        			parentFile.Sync()
        			offset += int64(n)
        		}
        		continue
			}
		}
		break
	}
	return nil
}
