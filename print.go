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
	if len(os.Args) != 2 {
		log.Fatal("need file name")
		return
	}
	var err error
	blockSize, err = GetFileSystemBlockSize()
	if err != nil {
		log.Fatal("can't get FS block size", err)
		return
	}

	fileName := os.Args[1]

	// open child and parent files
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal("Failed to open fileName")
	}
	defer file.Close()

	var dataholes []int64
	dataholes, err = getDataSegOffsets(file)
	fmt.Println(dataholes)
}

func getDataSegOffsets(file *os.File) ([]int64, error) {
	var data, hole int64
	var datahole []int64
	var err error
	for {
		data, err = syscall.Seek(int(file.Fd()), hole, SEEK_DATA)
		if err == nil {
			hole, err = syscall.Seek(int(file.Fd()), data, SEEK_HOLE)
			if err == nil {
				datahole = append(datahole, data)
				datahole = append(datahole, hole - data)
				continue
			}
		}
		break
	}
	return datahole, nil
}
