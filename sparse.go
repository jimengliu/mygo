package main

import (
	"log"
	"os"
	"syscall"
)

const (
	FILE_SIZE = 4 * 1024 * 10 // 40KB file, 10 * 4KB block
)

// get the file system block size
func GetFileSystemBlockSize() (int64, error) {
	var stat syscall.Stat_t
	err := syscall.Stat(os.DevNull, &stat)
	return stat.Blksize, err
}

func createSparseFile(fileName string, offsets []int64) error {
	// create a sparse file
	size := int64(FILE_SIZE)
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal("Failed to create child file")
		return err
	}
	defer file.Close()
	err = file.Truncate(size)
	if err != nil {
		log.Fatal("Failed to Truncate")
		return err
	}

	for _, offset := range offsets {
		_, err = file.Seek(int64(offset), 0)
		if err != nil {
			log.Fatal("Seek failed")
			return err
		}
		_, err = file.Write([]byte(fileName))
		if err != nil {
			log.Fatal("Write failed")
			return err
		}
		file.Sync()
	}

	return nil
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("need child and parent file names, child first and parent second")
		return
	}
	childFileName := os.Args[1]
	parentFileName := os.Args[2]
	blockSize, err := GetFileSystemBlockSize()
	if err != nil {
		log.Fatal("can't get FS block size", err)
		return
	}

	// make some parts dirty at block offset in child and parent:
	// child: 0(beginning), 3, 9
	// parent: 0(beginning), 5, 7
	childDirtyOffsets := []int64{0, 3 * blockSize, 9 * blockSize}
	parentDirtyOffsets := []int64{0, 5 * blockSize, 7 * blockSize}

	// create child and parent sparse files
	err = createSparseFile(childFileName, childDirtyOffsets)
	if err != nil {
		log.Fatal("Failed to create child file")
	}

	err = createSparseFile(parentFileName, parentDirtyOffsets)
	if err != nil {
		log.Fatal("Failed to create parent file")
	}
}
