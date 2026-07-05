package ws

import {
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
}

const {	
	OpcodeContinuation = 0x0
	OpcodeText = 0x1
	OpcodeBinary = 0x2
	OpcodeClose = 0x8
	OpcodePing = 0x9
	OpcodePong = 0xA
}

type Frame struct {
	IsFinal bool
	Opcode  byte  
	Masked	bool
	Payload []byte
}

 // parse raw bytes -> Frame
func ReadFrame(r io.Reader) (Frame, error) {
	/*
	Byte 0:   1 0 0 0 0 0 0 1
          	  │ │ │ │ └───────── opcode (4 bits)
              │ └─┴─┴───────── RSV1, RSV2, RSV3 (reserved, always 0)
              └─────────────── FIN bit

	Byte 1:   1 0 0 0 0 1 1 0
	    	  │ └─────────────── payload length (7 bits)
			  └─────────────── MASK bit
	*/

	// Byte 0: FIN bit + opcode
	header := make([]byte, 2)
	if _, err := io.ReadFull(r, header); err != nil {
		return Frame{}, err
	}
	// fin := header[0]&0x80 != 0
	// 0x80 in binary is 1000 0000
	// a mask that isolates just the top bit
	fin := header[0]&0x80 != 0
	// 0x0F in binary is 0000 1111 — isolates the bottom 4 bits:
	opcode := header[0] & 0x0F
	
	
	// Byte 1: MASK bit + payload length
	// Per RFC 6455, client→server frames are always masked. Server→client frames are never masked.
	masked := header[1]&0x80 != 0
	// 0x7F in binary is 0111 1111 — isolates bottom 7 bits, zeroing out the mask bit:
	payloadLen := int64(header[1] & 0x7F)

	/*The 7-bit value tells you how to read the actual length:
		0-125   → that IS the length
		126     → read next 2 bytes as uint16 for real length
		127     → read next 8 bytes as uint64 for real length
	*/
	// Extended payload length
	switch payloadLen {
		case 126:
			var ext uint16 //2 bytes
			if err := binary.Read(r, binary.BigEndian, &ext); err != nil {
				return Frame{}, err
			}
			payloadLen = int64(ext)
		case 127:
			var ext uint64 //8 bytes
			if err := binary.Read(r, binary.BigEndian, &ext); err != nil {
				return Frame{}, err
			}
			payloadLen = int64(ext)
	}
	
	
	// Masking key (clients always mask per RFC 6455)
	var maskKey [4]byte
	if masked {
		if _, err := io.ReadFull(r, maskKey[:]); err != nil {
			return Frame{}, err
		}
	}

	// Payload
	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return Frame{}, err
	}

	// Unmask
	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	return Frame{
		IsFinal: fin,
		Opcode:  opcode,
		Masked:  masked,
		Payload: payload,
	}, nil
}


// WriteFrame encodes and writes a WebSocket frame to w.
// encoded Frame -> raw bytes
func WriteFrame(w io.Writer, f Frame) error {
	if w == nil {
		return errors.New("nil writer")
	}

	
	var buf []byte

	// Byte 0: FIN + opcode
	b0 := f.Opcode
	if f.IsFinal {
		b0 |= 0x80
	}

	buf = append(buf, b0)

	// Byte 1: MASK bit + payload length
	payloadLen := len(f.Payload)
	maskBit := byte(0)
	if f.Masked {
		maskBit = 0x80
	}

	/* 
		0-125 bytes     → length fits in 7 bits → just write it directly in byte 1
		126-65535 bytes → too big for 7 bits    → write 126 as signal, then real length in next 2 bytes
		65536+ bytes    → too big for 2 bytes   → write 127 as signal, then real length in next 8 bytes
	*/
	switch {
		// byte 1 = MASK bit | actual length
		case payloadLen <= 125:
			buf = append(buf, maskBit|byte(payloadLen))
		// medium payload (126-65535)
		// bit 7 (top):    MASK bit  → is payload masked?
		// bits 0-6 (bottom 7): 127  → signal "read next 8 bytes for real length"
		case payloadLen <= 65535:
			buf = append(buf, maskBit|126) // byte 1 = signal value 126
			ext := make([]byte, 2)
			binary.BigEndian.PutUint16(ext, uint16(payloadLen))
			buf = append(buf, ext...)
		default:
			buf = append(buf, maskBit|127) // byte 1 = signal value 127
			ext := make([]byte, 8)
			binary.BigEndian.PutUint64(ext, uint64(payloadLen))
			buf = append(buf, ext...)
	}

	// Masking (server→client frames are typically unmasked per RFC 6455)
	if f.Masked {
		//TODO understand
		var maskKey [4]byte
		binary.BigEndian.PutUint32(maskKey[:], rand.Uint32())
		buf = append(buf, maskKey[:]...)
		masked := make([]byte, payloadLen)
		for i, b := range f.Payload {
			masked[i] = b ^ maskKey[i%4]
		}
		buf = append(buf, masked...)
	} else {
		buf = append(buf, f.Payload...)
	}

	_, err := w.Write(buf)
	return err
 
			
}


/*
Almost every network protocol (TCP, HTTP, WebSocket, DNS) uses BigEndian.

BigEndian vs LittleEndian is about byte order — when a number is bigger than 1 byte, which byte comes first?
Take the number 1000 as a uint16 (2 bytes). In hex that's 0x03E8:
BigEndian    → 03 E8   (most significant byte first — "natural" reading order)
LittleEndian → E8 03   (least significant byte first — reversed)

Wire bytes: 03 E8

binary.BigEndian.Uint16  → 0x03E8 = 1000  		Right
binary.LittleEndian.Uint16 → 0xE803 = 59395  	Wrong
*/
