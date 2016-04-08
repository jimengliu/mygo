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
var blockSize int64 = 4 * 1024

func main() {
	if len(os.Args) != 2 {
		log.Fatal("need file name")
		return
	}
	var err error
	blockSize, err = GetFileSystemBlockSize()
	if err != nil {
		log.Fatal("can't get FS block size", err)
	}

	fileName := os.Args[1]

	// open child and parent files
	file, err := os.Open(fileName)
	if err != nil {
		panic("Failed to open fileName, error:" + err.Error())
	}
	defer file.Close()

	var dataSegments []int64
	dataSegments, err = getDataSegOffsets(file)
	if err == nil {
        fmt.Println(dataSegments)
    } else {
        panic("failed to getDataSegOffsets(" + fileName + ") error:" + err.Error())
    }
}

func getDataSegOffsets(file *os.File) ([]int64, error) {
	var data, hole int64
	var dataSegment []int64
	var err error
	for {
		data, err = syscall.Seek(int(file.Fd()), hole, SEEK_DATA)
		if err == nil {
			hole, err = syscall.Seek(int(file.Fd()), data, SEEK_HOLE)
			if err == nil {
				dataSegment = append(dataSegment, data)
				dataSegment = append(dataSegment, hole - data)
				continue
			}
		}
		break
	}
	return dataSegment, nil
}
