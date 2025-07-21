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
	"sort"
	"time"

	"github.com/djherbis/times"
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

	SEGMENT_SIZE        = 8000 * 4 // user specified number of blocks in a segment get from config file
	RECORD_TYPE_SIZE    = 1
	PAYLOAD_SIZE        = 2
	BLOCK_USABLE_SPACE  = BLOCK_SIZE - RECORD_TYPE_SIZE - PAYLOAD_SIZE
	SEGMENT_CREATED     = 8
	SEGMENT_VALID_CHECK = 9
	SEGMENT_OFFSET      = 4
	SEGMENT_HEADER_SIZE = SEGMENT_CREATED + SEGMENT_VALID_CHECK + SEGMENT_OFFSET
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
	segmentPaths      []*WriteAheadLogSegment
	lastSegment       *WriteAheadLogSegment
	lastLogNumber     int
	lowWaterMark      int
	status            WriteStatus
	currentReadStatus CurrentReadStatus
	blockManager      *BlockManager
}

type WriteAheadLogSegment struct {
	logNumber         int
	offset            int
	segmentPath       WriteAheadLogPaths
	elements          []WriteAheadLogElement
	segmentBlocksLeft int
	blockManagerWrap  *BlockManagerWrap
}

type WriteAheadLogPaths struct {
	segmentPath string
	created     uint64
}

type WriteStatus struct {
	filePath string
	offset   int
}

type CurrentReadStatus struct {
	currentSegment *WriteAheadLogSegment
	currentOffset  int
	endReached     bool
}

func GetSegmentBlocksLeft(wal *WriteAheadLog) int {
	if wal.lastSegment == nil {
		return 0
	}
	return wal.lastSegment.segmentBlocksLeft
}

func NewWriteStatus(filePath string, offset int) *WriteStatus {
	return &WriteStatus{
		filePath: filePath,
		offset:   offset,
	}
}

func NewCurrentReadStatus(segment *WriteAheadLogSegment, offset int, endReached bool) *CurrentReadStatus {
	return &CurrentReadStatus{
		currentSegment: segment,
		currentOffset:  offset,
		endReached:     endReached,
	}
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

func NewWriteAheadLog(blockManager *BlockManager) *WriteAheadLog {
	return &WriteAheadLog{
		segmentPaths:  make([]*WriteAheadLogSegment, 0),
		lastSegment:   nil,
		lastLogNumber: 0,
		lowWaterMark:  0,
		blockManager:  blockManager,
		status:        WriteStatus{},
		currentReadStatus: CurrentReadStatus{
			currentSegment: nil,
			currentOffset:  -1,
			endReached:     false,
		},
	}
}

func NewWriteAheadLogSegment(logNumber int, offset int, segmentPath WriteAheadLogPaths, blockManager *BlockManager) *WriteAheadLogSegment {
	return &WriteAheadLogSegment{
		logNumber:         logNumber,
		offset:            offset,
		segmentPath:       segmentPath,
		elements:          make([]WriteAheadLogElement, 0),
		segmentBlocksLeft: SEGMENT_SIZE / BLOCK_SIZE,
		blockManagerWrap:  NewBlockManagerWrap(segmentPath.segmentPath, blockManager),
	}
}

func NewWriteAheadLogPaths(segmentPath string, created uint64) *WriteAheadLogPaths {
	return &WriteAheadLogPaths{
		segmentPath: segmentPath,
		created:     created,
	}
}

func (wal *WriteAheadLog) GetSegmentPaths() error {
	segmentNumber := 1
	err := filepath.Walk("./wal/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", path, err)
			return err
		}
		if !info.IsDir() {
			fmt.Println(path)
			t, err := times.Stat(path)
			if err != nil {
				log.Fatal(err.Error())
			}
			if t.HasBirthTime() {
				log.Println(t.BirthTime())
				createdTime := uint64(t.BirthTime().Unix())
				segmentLogPath := NewWriteAheadLogPaths(path, createdTime)
				segment := NewWriteAheadLogSegment(segmentNumber, 0, *segmentLogPath, wal.blockManager)
				segmentNumber++
				wal.segmentPaths = append(wal.segmentPaths, segment)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", "./wal/", err)
	}

	sort.Slice(wal.segmentPaths, func(i, j int) bool {
		return wal.segmentPaths[i].segmentPath.created < wal.segmentPaths[j].segmentPath.created
	})

	fmt.Println("Segment Paths:")
	for _, segment := range wal.segmentPaths {
		fmt.Printf("Path: %s, Created: %d\n", segment.segmentPath.segmentPath, segment.segmentPath.created)
	}

	return err
}

func (wal *WriteAheadLog) InitializeWriteAheadLog() {
	wal.GetSegmentPaths()
	if len(wal.segmentPaths) == 0 {
		fmt.Println("No segments found, creating initial segment.")
		initialSegmentPath := NewWriteAheadLogPaths("wal/wal_1.log", 0)
		initialSegment := NewWriteAheadLogSegment(1, 0, *initialSegmentPath, wal.blockManager)
		initialSegment.InitSegmentHeader()
		wal.segmentPaths = append(wal.segmentPaths, initialSegment)
		wal.lastSegment = initialSegment
		wal.lastLogNumber = 1
		wal.currentReadStatus.currentOffset = SEGMENT_HEADER_SIZE
		wal.currentReadStatus.currentSegment = initialSegment
	} else {
		passed := false
		for i := range wal.segmentPaths {
			lastSegment := wal.segmentPaths[len(wal.segmentPaths)-1-i]
			offset, err := wal.ReadSegmentHeader(lastSegment)
			if err != nil {
				log.Fatalf("Error reading segment header: %v", err)
			}
			if offset < 0 {
				log.Fatalf("Invalid segment header for segment %s: %v", lastSegment.segmentPath.segmentPath, err)
				wal.segmentPaths = append(wal.segmentPaths[:len(wal.segmentPaths)-1-i], wal.segmentPaths[len(wal.segmentPaths)-1-i:]...)
				continue
			} else if !passed {
				wal.lastSegment = lastSegment
				wal.lastLogNumber = lastSegment.logNumber
				fmt.Println("segment number:", wal.lastSegment.logNumber)
				wal.lastSegment.segmentBlocksLeft = SEGMENT_SIZE/BLOCK_SIZE - lastSegment.blockManagerWrap.GetFileEnd()/BLOCK_SIZE
				fmt.Println("segment blocks left11:", wal.lastSegment.segmentBlocksLeft)
				passed = true
			}
			if offset == SEGMENT_HEADER_SIZE {
				fmt.Println("Segment header found at offset 0, setting status to segment start.")
				wal.status = WriteStatus{
					filePath: lastSegment.segmentPath.segmentPath,
					offset:   offset,
				}
				fmt.Println("Current WriteAheadLog status:", wal.status.filePath, wal.status.offset)
			} else if offset == SEGMENT_SIZE+1 {
				break
			} else {
				wal.status = WriteStatus{
					filePath: lastSegment.segmentPath.segmentPath,
					offset:   offset,
				}
				break
			}
		}
		fmt.Print("Current WriteAheadLog status22:")
		fmt.Println("File Path:", wal.status.filePath, "Offset:", wal.currentReadStatus.currentOffset)
		fmt.Println("segment blocks left:", wal.lastSegment.segmentBlocksLeft)
		if wal.status == (WriteStatus{}) {
			log.Fatal("No valid segment found in WriteAheadLog")
			initialSegmentPath := NewWriteAheadLogPaths("wal/wal_1.log", 0)
			initialSegment := NewWriteAheadLogSegment(1, 0, *initialSegmentPath, wal.blockManager)
			initialSegment.InitSegmentHeader()
			wal.segmentPaths = append(wal.segmentPaths, initialSegment)
			wal.lastSegment = initialSegment
			wal.lastLogNumber = 1
			wal.currentReadStatus.currentOffset = SEGMENT_HEADER_SIZE
			wal.currentReadStatus.currentSegment = initialSegment
		} else {
			wal.lastSegment = wal.segmentPaths[len(wal.segmentPaths)-1]
		}
	}
}

func (wal *WriteAheadLog) ReadSegmentHeader(segment *WriteAheadLogSegment) (int, error) {
	_, header := segment.blockManagerWrap.ReadFromCurrentPosition(SEGMENT_HEADER_SIZE)

	if len(header) < SEGMENT_HEADER_SIZE {
		return 0, fmt.Errorf("segment header is too short: %d bytes", len(header))
	}
	if header[0] != 0b01110101 || header[1] != 0b01010101 || header[2] != 0b01101111 || header[3] != 0b01001111 ||
		header[4] != 0b01110101 || header[5] != 0b11110000 || header[6] != 0b10011111 || header[7] != 0b10010001 ||
		header[8] != 0b10111011 {
		return -2, fmt.Errorf("invalid segment header signature")
	}
	createdTime := binary.LittleEndian.Uint64(header[SEGMENT_VALID_CHECK : SEGMENT_VALID_CHECK+SEGMENT_CREATED])
	if createdTime != segment.segmentPath.created {
		return -2, fmt.Errorf("segment created time mismatch: expected %d, got %d", segment.segmentPath.created, createdTime)
	}
	offset := binary.LittleEndian.Uint32(header[SEGMENT_HEADER_SIZE-SEGMENT_OFFSET : SEGMENT_HEADER_SIZE])
	fmt.Printf("Segment created time: %d, offset: %d\n", createdTime, offset)
	return int(offset), nil
}

func (s *WriteAheadLogSegment) InitSegmentHeader() {
	segmentHeader := make([]byte, 0)
	segmentHeader = append(segmentHeader, 0b01110101, 0b01010101, 0b01101111, 0b01001111, 0b01110101, 0b11110000, 0b10011111, 0b10010001, 0b10111011)
	t, err := times.Stat(s.segmentPath.segmentPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	if t.HasBirthTime() {
		log.Println(t.BirthTime())
		createdTime := make([]byte, 8)
		binary.LittleEndian.PutUint64(createdTime, uint64(t.BirthTime().Unix()))
		segmentHeader = append(segmentHeader, createdTime...)
	}
	headerSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(headerSize, uint32(SEGMENT_HEADER_SIZE))
	segmentHeader = append(segmentHeader, headerSize...)
	s.blockManagerWrap.WriteToFile(segmentHeader)
	s.blockManagerWrap.FlushBuffer(false)
}

func (wal *WriteAheadLog) WriteToFile(element *WriteAheadLogElement) {
	fmt.Println("Writing to file:", wal.lastSegment.segmentPath.segmentPath)
	fmt.Println("segment blocks left:", wal.lastSegment.segmentBlocksLeft)
	data := element.Serialize()
	newBlock := make([]byte, BLOCK_SIZE)
	wal.lastSegment.blockManagerWrap.InitializeWriteBuffer()
	if len(data)+RECORD_TYPE_SIZE+PAYLOAD_SIZE > BLOCK_SIZE-wal.lastSegment.blockManagerWrap.currentBufferUsedSpace {
		fmt.Println("current buffer used space in write:", wal.lastSegment.blockManagerWrap.currentBufferUsedSpace)
		allBlocksToWrite := make([]byte, 0)
		if len(data) > BLOCK_USABLE_SPACE {
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

				binary.LittleEndian.PutUint16(newBlock[RECORD_TYPE_SIZE:], uint16(len(blockData)))

				if i == int(numBlocks)-1 {
					allBlocksToWrite = append(allBlocksToWrite, byte(LAST))
					lenBytes := make([]byte, 2)
					binary.LittleEndian.PutUint16(lenBytes, uint16(len(blockData)))
					allBlocksToWrite = append(allBlocksToWrite, lenBytes...)
					allBlocksToWrite = append(allBlocksToWrite, blockData...)
				} else {
					copy(newBlock[RECORD_TYPE_SIZE+PAYLOAD_SIZE:], blockData)
					allBlocksToWrite = append(allBlocksToWrite, newBlock...)
				}
			}
			fmt.Println("Data with type333:", string(allBlocksToWrite))
			if int(numBlocks) > wal.lastSegment.segmentBlocksLeft {
				fmt.Println("usao", int(numBlocks))
				blocksAboveLimit := int(numBlocks) - wal.lastSegment.segmentBlocksLeft
				wal.lastSegment.blockManagerWrap.WriteToFile(allBlocksToWrite[:wal.lastSegment.segmentBlocksLeft*BLOCK_SIZE])
				wal.lastSegment.segmentBlocksLeft = 0
				newSegment := NewWriteAheadLogSegment(wal.lastSegment.logNumber+1, SEGMENT_SIZE*wal.lastSegment.logNumber, *NewWriteAheadLogPaths(fmt.Sprintf("wal/wal_%d.log", wal.lastSegment.logNumber+1), 0), wal.blockManager)
				newSegment.InitSegmentHeader()
				newSegment.blockManagerWrap.InitializeWriteBuffer()
				newSegment.blockManagerWrap.WriteToFile(allBlocksToWrite[wal.lastSegment.segmentBlocksLeft*BLOCK_SIZE:])
				wal.segmentPaths = append(wal.segmentPaths, newSegment)
				wal.lastSegment = newSegment
				wal.lastLogNumber += 1
				wal.lastSegment.segmentBlocksLeft = SEGMENT_SIZE/BLOCK_SIZE - blocksAboveLimit
			} else {
				wal.lastSegment.blockManagerWrap.WriteToFile(allBlocksToWrite)
				wal.lastSegment.segmentBlocksLeft -= int(numBlocks)
			}
			wal.lastSegment.blockManagerWrap.FlushBuffer(false)
		} else {
			newBlockOne := make([]byte, 0)
			newBlockOne = append(newBlockOne, byte(FULL))
			lenBytes := make([]byte, 2)
			binary.LittleEndian.PutUint16(lenBytes, uint16(len(data)))
			newBlockOne = append(newBlockOne, lenBytes...)
			newBlockOne = append(newBlockOne, data...)
			/* newLength := len(data) + RECORD_TYPE_SIZE + PAYLOAD_SIZE
			binary.LittleEndian.PutUint32(newBlockOne[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(newLength))
			allBlocksToWrite = append(allBlocksToWrite, newBlockOne...) */
			fmt.Println("Data with type444:", string(newBlockOne))
			wal.lastSegment.blockManagerWrap.WriteToFile(newBlockOne)
			wal.lastSegment.segmentBlocksLeft -= 1
			wal.lastSegment.blockManagerWrap.FlushBuffer(false)
		}
	} else if len(data)+RECORD_TYPE_SIZE+PAYLOAD_SIZE < BLOCK_SIZE-wal.lastSegment.blockManagerWrap.currentBufferUsedSpace {
		fmt.Println("current buffer used space in write222:", wal.lastSegment.blockManagerWrap.currentBufferUsedSpace)
		dataWithType := make([]byte, 0)
		dataWithType = append(dataWithType, byte(FULL))
		lenBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(lenBytes, uint16(len(data)))
		dataWithType = append(dataWithType, lenBytes...)
		dataWithType = append(dataWithType, data...)
		fmt.Println("Data with type:", string(dataWithType))
		wal.lastSegment.blockManagerWrap.WriteToFile(dataWithType)
		wal.lastSegment.blockManagerWrap.FlushBuffer(false)
	} else if len(data)+RECORD_TYPE_SIZE+PAYLOAD_SIZE <= BLOCK_SIZE-wal.lastSegment.blockManagerWrap.currentBufferUsedSpace {
		dataWithType := make([]byte, 0)
		dataWithType = append(dataWithType, byte(FULL))
		dataWithType = append(dataWithType, data...)
		fmt.Println("Data with type222:", string(dataWithType))
		wal.lastSegment.segmentBlocksLeft -= 1
		if wal.lastSegment.segmentBlocksLeft == 0 {
			newSegment := NewWriteAheadLogSegment(wal.lastSegment.logNumber+1, SEGMENT_SIZE*wal.lastSegment.logNumber, *NewWriteAheadLogPaths(fmt.Sprintf("wal/wal_%d.log", wal.lastSegment.logNumber+1), 0), wal.blockManager)
			newSegment.InitSegmentHeader()
			wal.lastSegment.blockManagerWrap.setAvilableSpace(false)
			newSegment.blockManagerWrap.InitializeWriteBuffer()
			emptyBlock := make([]byte, BLOCK_SIZE)
			binary.LittleEndian.PutUint32(emptyBlock[BLOCK_SIZE-BLOCK_HEADER_SIZE:], uint32(0))
			newSegment.blockManagerWrap.WriteToFile(emptyBlock)
		}
		wal.lastSegment.blockManagerWrap.WriteToFile(dataWithType)
	}
}

func (wal *WriteAheadLog) GetWalElement() (*Entry, bool) {
	fullElementRead := false
	fullElementBytes := make([]byte, 0)

	for !fullElementRead {
		fmt.Println("Current offset:", wal.currentReadStatus.currentOffset)
		elementType, payloadSize := wal.GetElementTypeAndSize(wal.currentReadStatus.currentOffset)
		if payloadSize == -1 {
			return nil, false
		}
		_, data := wal.currentReadStatus.currentSegment.blockManagerWrap.ReadFromFile(wal.currentReadStatus.currentOffset, payloadSize)
		wal.currentReadStatus.currentOffset += payloadSize
		fmt.Println("Read data:", string(data))
		fmt.Println("Next offset:", wal.currentReadStatus.currentOffset)
		if wal.currentReadStatus.currentOffset >= wal.currentReadStatus.currentSegment.blockManagerWrap.fileEnd {
			found := -1

			for i, segment := range wal.segmentPaths {
				if segment.segmentPath.segmentPath == wal.currentReadStatus.currentSegment.segmentPath.segmentPath {
					wal.currentReadStatus.currentSegment = segment
					found = i
					break
				}
			}

			if len(wal.segmentPaths)-1 >= found+1 {
				wal.currentReadStatus.currentOffset = SEGMENT_HEADER_SIZE
				wal.currentReadStatus.currentSegment = wal.segmentPaths[found+1]
			} else {
				wal.currentReadStatus.endReached = true
			}
		}
		if elementType == FULL || elementType == LAST {
			fullElementRead = true
		}
		fullElementBytes = append(fullElementBytes, data...)
	}

	elementCrc := CRC32(fullElementBytes[TIMESTAMP_START:])

	if elementCrc != binary.LittleEndian.Uint32(fullElementBytes[CRC_START:CRC_SIZE]) {
		return nil, false
	}

	element := Deserialize(fullElementBytes)

	return &Entry{
		key:     string(element.key),
		value:   element.value,
		deleted: element.tombstone,
	}, true
}

func (wal *WriteAheadLog) GetElementTypeAndSize(offset int) (RecordType, int) {
	fmt.Println("Getting element type and size at offset:", wal.currentReadStatus.currentOffset)
	_, data := wal.currentReadStatus.currentSegment.blockManagerWrap.ReadFromFile(wal.currentReadStatus.currentOffset, RECORD_TYPE_SIZE+PAYLOAD_SIZE)
	elementType := data[0]
	payloadSize := int(binary.LittleEndian.Uint16(data[RECORD_TYPE_SIZE:]))
	wal.currentReadStatus.currentOffset += RECORD_TYPE_SIZE + PAYLOAD_SIZE

	fmt.Println("Element type:", elementType, "Payload size:", payloadSize)
	fmt.Println("Current offset:", wal.currentReadStatus.currentOffset)

	for payloadSize == 0 {
		_, data = wal.currentReadStatus.currentSegment.blockManagerWrap.ReadFromFile(wal.currentReadStatus.currentOffset, RECORD_TYPE_SIZE+PAYLOAD_SIZE)
		elementType = data[0]
		payloadSize = int(binary.LittleEndian.Uint16(data[RECORD_TYPE_SIZE:]))
		wal.currentReadStatus.currentOffset += RECORD_TYPE_SIZE + PAYLOAD_SIZE
		if wal.currentReadStatus.currentOffset%BLOCK_SIZE == 1 {
			wal.currentReadStatus.currentOffset -= 1
		}
		/* fmt.Println("current offset in get element type:", wal.currentReadStatus.currentOffset)
		fmt.Println("Element type:", elementType, "Payload size for header:", payloadSize) */

		if wal.currentReadStatus.currentOffset >= wal.currentReadStatus.currentSegment.blockManagerWrap.fileEnd-BLOCK_HEADER_SIZE {
			found := -1

			for i, segment := range wal.segmentPaths {
				if segment.segmentPath.segmentPath == wal.currentReadStatus.currentSegment.segmentPath.segmentPath {
					wal.currentReadStatus.currentSegment = segment
					found = i
					break
				}
			}

			if len(wal.segmentPaths)-1 >= found+1 {
				wal.currentReadStatus.currentOffset = SEGMENT_HEADER_SIZE
				wal.currentReadStatus.currentSegment = wal.segmentPaths[found+1]
			} else {
				wal.currentReadStatus.endReached = true
				return RecordType(0), -1
			}
		}
	}

	return RecordType(elementType), payloadSize
}

func (wal *WriteAheadLog) GetWalElementByElement() (*Entry, bool) {
	if wal.currentReadStatus.endReached {
		fmt.Println("End of WAL reached.")
		return nil, false
	}
	if wal.currentReadStatus.currentOffset == -1 {
		fmt.Println("Current offset is -1, setting to SEGMENT_HEADER_SIZE")
		fmt.Println(wal.status.filePath, wal.status.offset)
		wal.currentReadStatus.currentOffset = wal.status.offset
	}
	if wal.currentReadStatus.currentSegment.segmentPath.segmentPath == "" {
		for _, segment := range wal.segmentPaths {
			if segment.segmentPath.segmentPath == wal.status.filePath {
				wal.currentReadStatus.currentSegment = segment
				break
			}
		}
	}

	fmt.Println("Current segment offset:", wal.currentReadStatus.currentOffset)
	fmt.Println("end", wal.currentReadStatus.endReached)
	element, check := wal.GetWalElement()

	for !check && !wal.currentReadStatus.endReached {
		fmt.Println("CRC32 checksum failed. Trying next element...")
		element, check = wal.GetWalElement()
	}

	if wal.currentReadStatus.endReached {
		fmt.Println("End of WAL reached.")
		return nil, false
	}

	return element, true
}

func (wal *WriteAheadLog) SetStatus(filePath string, offset int) {
	wal.status.filePath = filePath
	wal.status.offset = offset

	blockManagerWrap := NewBlockManagerWrap(filePath, wal.blockManager)
	blockManagerWrap.SetCurrentPosition(SEGMENT_HEADER_SIZE - SEGMENT_OFFSET)
	offsetBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(offsetBytes, uint32(offset))
	blockManagerWrap.OverwriteDataInFile(offsetBytes)

	found := false
	index := -1

	for i, segment := range wal.segmentPaths {
		if segment.segmentPath.segmentPath == filePath {
			found = true
			index = i
			break
		}
	}

	if found {
		previousSegmentPath := wal.segmentPaths[index-1]
		blockManagerWrapPervious := NewBlockManagerWrap(previousSegmentPath.segmentPath.segmentPath, wal.blockManager)
		blockManagerWrapPervious.SetCurrentPosition(SEGMENT_HEADER_SIZE - SEGMENT_OFFSET)
		offsetBytesPrevious := make([]byte, 4)
		binary.LittleEndian.PutUint32(offsetBytesPrevious, uint32(SEGMENT_SIZE+1))
		blockManagerWrapPervious.OverwriteDataInFile(offsetBytesPrevious)
	}
}

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
