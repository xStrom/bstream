package bstream

import "io"

type bit bool

const (
	zero bit = false
	one  bit = true
)

//BStream :
type BStream struct {
	stream      []byte
	remainCount uint8
}

//NewBStreamReader :
func NewBStreamReader(data []byte) *BStream {
	return &BStream{stream: data, remainCount: 8}
}

//NewBStreamWriter :
func NewBStreamWriter(nByte uint8) *BStream {
	return &BStream{stream: make([]byte, 0, nByte), remainCount: 0}
}

//WriteBit :
func (b *BStream) writeBit(input bit) {
	if b.remainCount == 0 {
		b.stream = append(b.stream, 0)
		b.remainCount = 8
	}

	latestIndex := len(b.stream) - 1
	if input {
		b.stream[latestIndex] |= 1 << (b.remainCount - 1)
	}
	b.remainCount--
}

//WriteByte :
func (b *BStream) writeByte(data byte) {
	if b.remainCount == 0 {
		b.stream = append(b.stream, 0)
		b.remainCount = 8
	}

	latestIndex := len(b.stream) - 1

	b.stream[latestIndex] |= data >> (8 - b.remainCount)
	b.stream = append(b.stream, 0)
	latestIndex++
	b.stream[latestIndex] = data << b.remainCount
}

//WriteBits :
func (b *BStream) WriteBits(data uint64, count int) {
	data <<= uint(64 - count)

	//handle write byte if count over 8
	for count >= 8 {
		byt := byte(data >> (64 - 8))
		b.writeByte(byt)

		data <<= 8
		count -= 8
	}

	//handle write bit
	for count > 0 {
		bi := data >> (64 - 1)
		b.writeBit(bi == 1)

		data <<= 1
		count--
	}
}

func (b *BStream) readBit() (bit, error) {
	//empty return io.EOF
	if len(b.stream) == 0 {
		return zero, io.EOF
	}

	//if first byte already empty, move to next byte to retrieval
	if b.remainCount == 0 {
		b.stream = b.stream[1:]

		if len(b.stream) == 0 {
			return zero, io.EOF
		}

		b.remainCount = 8
	}

	// handle bit retrieval
	retBit := b.stream[0] & 0x80
	b.stream[0] <<= 1
	b.remainCount--

	return retBit != 0, nil
}

func (b *BStream) readByte() (byte, error) {
	//empty return io.EOF
	if len(b.stream) == 0 {
		return 0, io.EOF
	}

	//if first byte already empty, move to next byte to retrieval
	if b.remainCount == 0 {
		b.stream = b.stream[1:]

		if len(b.stream) == 0 {
			return 0, io.EOF
		}

		b.remainCount = 8
	}

	//just remain 8 bit, just return this byte directly
	if b.remainCount == 8 {
		byt := b.stream[0]
		b.stream = b.stream[1:]
		return byt, nil
	}

	//handle byte retrieval
	retByte := b.stream[0]
	b.stream = b.stream[1:]

	//check if we could finish retrieval on next byte
	if len(b.stream) == 0 {
		return 0, io.EOF
	}

	//handle remain bit on next stream
	retByte |= b.stream[0] >> b.remainCount
	b.stream[0] <<= (8 - b.remainCount)
	return retByte, nil
}

//ReadBits :
func (b *BStream) ReadBits(count int) (uint64, error) {

	var retValue uint64

	//empty return io.EOF
	if len(b.stream) == 0 {
		return 0, io.EOF
	}

	if b.remainCount == 0 {
		b.stream = b.stream[1:]

		if len(b.stream) == 0 {
			return 0, io.EOF
		}

		b.remainCount = 8
	}

	//handle byte reading
	for count >= 8 {
		retValue <<= 8
		byt, _ := b.readByte()
		retValue |= uint64(byt)
		count = count - 8
	}

	if count == 0 {
		return retValue, nil
	}

	for count > 0 {
		retValue <<= 1
		bi, _ := b.readBit()
		if bi {
			retValue |= 1
		}

		count--
	}

	return retValue, nil
}
