package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"
)

const (
	CRC_SIZE        = 4
	TIMESTAMP_SIZE  = 8
	TOMBSTONE_SIZE  = 1
	KEY_SIZE_SIZE   = 8
	VALUE_SIZE_SIZE = 8
	SIZE            = CRC_SIZE + TIMESTAMP_SIZE + TOMBSTONE_SIZE + KEY_SIZE_SIZE + VALUE_SIZE_SIZE

	CRC_START        = 0
	TIMESTAMP_START  = CRC_START + CRC_SIZE
	TOMBSTONE_START  = TIMESTAMP_START + TIMESTAMP_SIZE
	KEY_SIZE_START   = TOMBSTONE_START + TOMBSTONE_SIZE
	VALUE_SIZE_START = KEY_SIZE_START + KEY_SIZE_SIZE
	KEY_START        = VALUE_SIZE_START + VALUE_SIZE_SIZE

	SEGMENT_SIZE       = 8000 * 4 // user specified number of blocks in a segment get from config file
	RECORD_TYPE_SIZE   = 1
	BLOCK_USABLE_SPACE = BLOCK_SIZE - RECORD_TYPE_SIZE
)

type WriteAheadLogElement struct {
	crc       uint32
	timestamp uint64
	tombstone bool
	keySize   uint64
	valueSize uint64
	key       []byte
	value     []byte
}

type WriteAheadLog struct {
	filePaths     []WriteAheadLogPaths
	lastSegment   []WriteAheadLogSegment
	lastLogNumber int
	lowWaterMark  int
}

type WriteAheadLogSegment struct {
	logNumber         int
	offset            int
	segmentPath       string
	elements          []WriteAheadLogElement
	segmentBlocksLeft int
}

type WriteAheadLogPaths struct {
	segmentPath string
	offset      int
}

func NewWriteAheadLogElement(keySize uint64, valueSize uint64, key []byte, value []byte) *WriteAheadLogElement {
	timestamp := uint64(time.Now().Unix())
	elementSerialized := serializeForHashing(timestamp, false, keySize, valueSize, key, value)
	return &WriteAheadLogElement{
		crc:       CRC32(elementSerialized),
		timestamp: timestamp,
		tombstone: false,
		keySize:   keySize,
		valueSize: valueSize,
		key:       key,
		value:     value,
	}
}

func NewWriteAheadLog() *WriteAheadLog {
	return &WriteAheadLog{
		//elements: make([]byte, 0),
	}
}

func NewWriteAheadLogSegment(logNumber int, offset int, segmentPath string) *WriteAheadLogSegment {
	return &WriteAheadLogSegment{
		logNumber:         logNumber,
		offset:            offset,
		segmentPath:       segmentPath,
		elements:          make([]WriteAheadLogElement, 0),
		segmentBlocksLeft: SEGMENT_SIZE / BLOCK_SIZE,
	}
}

func GetFilePaths() /* []WriteAheadLogPaths */ error {
	err := filepath.Walk("./wal/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", path, err)
			return err
		}
		if !info.IsDir() {
			fmt.Println(path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", "./wal/", err)
	}

	return err

}

/* func (w *WriteAheadLog) AddElement(element *WriteAheadLogElement) {
	w.elements = append(w.elements, element.Serialize()...)
	WriteToFile(element)
} */

/* func openWriteAheadLogFile(openFlag int) *os.File {
	file, err := os.OpenFile("wal/writeAheadLog.log", openFlag|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func closeWriteAheadLogFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
} */

func (s *WriteAheadLogSegment) WriteToFile(element *WriteAheadLogElement, blockManager *BlockManagerWrap) {
	/* file := openWriteAheadLogFile(os.O_APPEND)
	defer closeWriteAheadLogFile(file) */

	data := element.Serialize()
	newBlock := make([]byte, BLOCK_SIZE)
	if len(data) > blockManager.currentBufferUsedSpace {
		allBlocksToWrite := make([]byte, 0)
		numBlocks := math.Ceil(float64(len(data)) / float64(BLOCK_USABLE_SPACE))
		for i := 0; i < int(numBlocks); i++ {
			start := i * BLOCK_USABLE_SPACE
			end := start + BLOCK_USABLE_SPACE
			if end > len(data) {
				end = len(data)
			}
			blockData := data[start:end]

			if i == 0 {
				newBlock[0] = byte(FIRST)
			} else if i == int(numBlocks)-1 {
				newBlock[0] = byte(LAST)
			} else {
				newBlock[0] = byte(MIDDLE)
			}

			if i == int(numBlocks)-1 {
				allBlocksToWrite = append(allBlocksToWrite, byte(LAST))
				allBlocksToWrite = append(allBlocksToWrite, blockData...)
			} else {
				copy(newBlock[RECORD_TYPE_SIZE:], blockData)
				allBlocksToWrite = append(allBlocksToWrite, newBlock...)
			}
		}
		if int(numBlocks) > s.segmentBlocksLeft {
			blockManager.WriteToFile(allBlocksToWrite[s.segmentBlocksLeft*BLOCK_SIZE:])
			s.segmentBlocksLeft = 0
			newSegment := NewWriteAheadLogSegment(s.logNumber+1, 0, fmt.Sprintf("wal/segment_%d.log", s.logNumber+1))
			fmt.Printf("Creating new segment: %s\n", newSegment.segmentPath)
		} else {
			blockManager.WriteToFile(allBlocksToWrite)
			s.segmentBlocksLeft -= int(numBlocks)
		}

	}
}

/* func (w *WriteAheadLog) ReadFromFile() {
	file := openWriteAheadLogFile(os.O_RDWR)
	defer closeWriteAheadLogFile(file)

	stat, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	size := stat.Size()

	data := make([]byte, size)
	_, err = file.Read(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(data)

	buffer := bytes.NewReader(data[KEY_SIZE_START:VALUE_SIZE_START])
	keySize := uint64(0)
	valueSize := uint64(0)

	err = binary.Read(buffer, binary.LittleEndian, &keySize)
	if err != nil {
		log.Fatal(err)
	}

	buffer = bytes.NewReader(data[VALUE_SIZE_START:KEY_START])

	err = binary.Read(buffer, binary.LittleEndian, &valueSize)
	if err != nil {
		log.Fatal(err)
	}

	elementSize := SIZE + keySize + valueSize

	fmt.Println(elementSize)
	fmt.Println(keySize)
	fmt.Println(valueSize)

	validCheck := CRC32(data[TIMESTAMP_START:elementSize]) == binary.LittleEndian.Uint32(data[CRC_START:CRC_SIZE])
	if !validCheck {
		log.Fatal("CRC32 checksum failed")
	} else {
		element := Deserialize(data[0:elementSize])
		fmt.Println(element)
		w.AddElement(element)
	}
	fmt.Println("Data length: ", len(data))

	var i int = int(elementSize)

	for i < len(data) && len(data)-i > SIZE {
		fmt.Println("i: ", i)
		fmt.Println(data[i+KEY_SIZE_START : i+VALUE_SIZE_START])
		buffer := bytes.NewReader(data[i+KEY_SIZE_START : i+VALUE_SIZE_START])
		keySize := uint64(0)
		valueSize := uint64(0)

		err = binary.Read(buffer, binary.LittleEndian, &keySize)
		if err != nil {
			log.Fatal(err)
		}

		buffer = bytes.NewReader(data[i+VALUE_SIZE_START : i+KEY_START])

		err = binary.Read(buffer, binary.LittleEndian, &valueSize)
		if err != nil {
			log.Fatal(err)
		}

		elementSize := SIZE + keySize + valueSize
		fmt.Println(elementSize)
		fmt.Println(keySize)
		fmt.Println(valueSize)

		validCheck = CRC32(data[i+TIMESTAMP_START:i+int(elementSize)]) == binary.LittleEndian.Uint32(data[i+CRC_START:i+CRC_SIZE])
		if !validCheck {
			log.Fatal("CRC32 checksum failed")
		} else {
			element := Deserialize(data[i : i+int(elementSize)])
			w.AddElement(element)
			fmt.Println(element)
		}

		i += int(elementSize)
	}

} */

func (w *WriteAheadLogElement) Serialize() []byte {
	buffer := new(bytes.Buffer)

	err := binary.Write(buffer, binary.LittleEndian, w.crc)
	if err != nil {
		log.Fatal(err)
	}
	err = binary.Write(buffer, binary.LittleEndian, w.timestamp)
	if err != nil {
		log.Fatal(err)
	}
	err = binary.Write(buffer, binary.LittleEndian, w.tombstone)
	if err != nil {
		log.Fatal(err)
	}
	err = binary.Write(buffer, binary.LittleEndian, w.keySize)
	if err != nil {
		log.Fatal(err)
	}
	err = binary.Write(buffer, binary.LittleEndian, w.valueSize)
	if err != nil {
		log.Fatal(err)
	}

	_, err = buffer.Write(w.key)
	if err != nil {
		log.Fatal(err)
	}

	_, err = buffer.Write(w.value)
	if err != nil {
		log.Fatal(err)
	}

	return buffer.Bytes()
}

func serializeForHashing(time uint64, tomb bool, keySize uint64, valueSize uint64, key []byte, value []byte) []byte {
	buffer := new(bytes.Buffer)

	err := binary.Write(buffer, binary.LittleEndian, time)
	if err != nil {
		log.Fatal(err)
	}
	err = binary.Write(buffer, binary.LittleEndian, tomb)
	if err != nil {
		log.Fatal(err)
	}
	err = binary.Write(buffer, binary.LittleEndian, keySize)
	if err != nil {
		log.Fatal(err)
	}
	err = binary.Write(buffer, binary.LittleEndian, valueSize)
	if err != nil {
		log.Fatal(err)
	}

	_, err = buffer.Write(key)
	if err != nil {
		log.Fatal(err)
	}

	_, err = buffer.Write(value)
	if err != nil {
		log.Fatal(err)
	}

	return buffer.Bytes()
}

func Deserialize(data []byte) *WriteAheadLogElement {
	buffer := bytes.NewReader(data)
	element := &WriteAheadLogElement{}

	err := binary.Read(buffer, binary.LittleEndian, &element.crc)
	if err != nil {
		log.Fatal(err)
	}
	err = binary.Read(buffer, binary.LittleEndian, &element.timestamp)
	if err != nil {
		log.Fatal(err)
	}
	err = binary.Read(buffer, binary.LittleEndian, &element.tombstone)
	if err != nil {
		log.Fatal(err)
	}
	err = binary.Read(buffer, binary.LittleEndian, &element.keySize)
	if err != nil {
		log.Fatal(err)
	}
	err = binary.Read(buffer, binary.LittleEndian, &element.valueSize)
	if err != nil {
		log.Fatal(err)
	}

	element.key = make([]byte, element.keySize)
	_, err = buffer.Read(element.key)
	if err != nil {
		log.Fatal(err)
	}

	element.value = make([]byte, element.valueSize)
	_, err = buffer.Read(element.value)
	if err != nil {
		log.Fatal(err)
	}

	return element
}

func CRC32(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}
