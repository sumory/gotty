package codec

import (
	"bufio"
	"encoding/binary"
	"fmt"
	log "github.com/sumory/log4go"
	"io"
)

//LengthBasedCodec 定长编解码器
type LengthBasedCodec struct {
	byteOrder binary.ByteOrder //大小端
	maxSize   int              //包最大长度
	encoder   Encoder
	decoder   Decoder
}

//NewLengthBasedCodec 新建定长编解码器
func NewLengthBasedCodec(byteOrder binary.ByteOrder, maxSize int, encoder Encoder, decoder Decoder) *LengthBasedCodec {
	return &LengthBasedCodec{
		byteOrder: byteOrder,
		maxSize:   maxSize,
		encoder:   encoder,
		decoder:   decoder,
	}
}

//Read 从连接中读取packet
func (lbc *LengthBasedCodec) Read(bReader *bufio.Reader) (*Packet, error) {
	//读包总大小
	totalLen := make([]byte, packetBytesLen, packetBytesLen)
	if _, err := io.ReadFull(bReader, totalLen); err != nil {
		return nil, err
	}

	tLen := lbc.byteOrder.Uint32(totalLen) //包总长度
	if lbc.maxSize > 0 && int(tLen) > lbc.maxSize {
		return nil, PacketTooLargeError
	}
	if tLen < packetMetaLen { //至少是codec.PacketMeta的大小
		return nil, PacketTooSmallError
	}

	//读包头大小
	headerLen := make([]byte, packetBytesLen, packetBytesLen)
	if _, err := io.ReadFull(bReader, headerLen); err != nil {
		return nil, err
	}
	hLen := lbc.byteOrder.Uint32(headerLen)
	if hLen > tLen-packetMetaLen {
		return nil, HeaderTooLargeError
	}
	if hLen < packetMinHeaderLen {
		return nil, HeaderTooSmallError
	}

	headerAndBodyLen := tLen - packetMetaLen
	headerAndBody := make([]byte, headerAndBodyLen)
	tmp := headerAndBody
	hasRead := 0
	for {
		log.Debug("read packget headerAndBody, hasRead: %d", hasRead)
		onceRead, err := bReader.Read(tmp)
		if err != nil {
			return nil, err
		}
		hasRead += onceRead
		log.Debug("read packget headerAndBody, hasRead: %d  onceRead: %d", hasRead, onceRead)
		if hasRead < int(headerAndBodyLen) {
			tmp = headerAndBody[hasRead:]
			continue
		} else {
			break
		}
	}

	//组装packet
	packet := &Packet{}
	if err := packet.Decode(lbc.byteOrder, tLen, hLen, headerAndBody); err != nil {
		return nil, err
	}

	log.Debug("read packet data, totallen: %d, headerLen: %d, headerAndBody: %v， packet.Header.Extra:%v",
		tLen, hLen, headerAndBody, packet.Header.Extra)
	return packet, nil
}

//Write 将包写出
func (lbc *LengthBasedCodec) Write(bWriter *bufio.Writer, p *Packet) error {
	if lbc.maxSize > 0 && int(p.Meta.TotalLen) > lbc.maxSize {
		return PacketTooLargeError
	}

	pBytes, err := p.Encode(lbc.byteOrder)
	log.Debug("write packet, length: %d, value: %v", len(pBytes), pBytes)

	if err != nil {
		log.Error("packet encode error, %s", err)
		return err
	}

	tmp := pBytes
	pBytesLen := len(pBytes)
	hasWrite := 0
	for {
		log.Debug("write packget, hasWrite: %d", hasWrite)
		onceWrite, err := bWriter.Write(tmp)
		if err != nil {
			//链接是关闭的
			if err != io.ErrShortWrite {
				log.Debug("write packget,err != io.ErrShortWrite")
				return err
			}

			log.Debug("write packget,err == io.ErrShortWrite")
		}
		hasWrite += onceWrite
		log.Debug("write packget, hasWrite: %d  onceWrite: %d", hasWrite, onceWrite)
		if hasWrite < pBytesLen {
			tmp = pBytes[hasWrite:]
			continue
		} else {
			break
		}
	}

	bWriter.Flush()
	return nil
}

//Marshal 将业务实体转为packet
func (lbc *LengthBasedCodec) Marshal(m interface{}) (*Packet, error) {
	return lbc.decoder.Decode(m)
}

//Unmarshal 将包转为业务实体
func (lbc *LengthBasedCodec) Unmarshal(p *Packet) (interface{}, error) {
	return lbc.encoder.Encode(p)
}
