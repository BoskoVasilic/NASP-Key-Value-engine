package structures

import (
	"log"
	"os"
)

const (
	BLOCK_SIZE = 8192 // specify the block size in config file
)

type Block struct {
	index int
	data  []byte
}

func openFile(filePath string, openFlag int) *os.File {
	file, err := os.OpenFile(filePath, openFlag, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
}

func readBlock(filePath string, blockNumber int) []byte {
	file := openFile(filePath, os.O_RDONLY)
	defer closeFile(file)

	_, err := file.Seek(int64(blockNumber*BLOCK_SIZE), 0)
	if err != nil {
		log.Fatal(err)
	}

	data := make([]byte, BLOCK_SIZE)

	_, err = file.Read(data)
	if err != nil {
		log.Fatal(err)
	}

	return data
}

func writeBlock(filePath string, blockNumber int, data []byte) {
	file := openFile(filePath, os.O_RDWR)
	defer closeFile(file)

	_, err := file.Seek(int64(blockNumber*BLOCK_SIZE), 0)
	if err != nil {
		log.Fatal(err)
	}

	_, err = file.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}
