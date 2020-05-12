package main

import (
	"fmt"
	"net"
	"time"
)

// Value for version component of FCGI_Header
const (
	FCGI_VERSION_1 uint8 = 1
)

// Values for type component of FCGI_Header
const (
	FCGI_BEGIN_REQUEST     uint8 = 1
	FCGI_ABORT_REQUEST     uint8 = 2
	FCGI_END_REQUEST       uint8 = 3
	FCGI_PARAMS            uint8 = 4
	FCGI_STDIN             uint8 = 5
	FCGI_STDOUT            uint8 = 6
	FCGI_STDERR            uint8 = 7
	FCGI_DATA              uint8 = 8
	FCGI_GET_VALUES        uint8 = 9
	FCGI_GET_VALUES_RESULT uint8 = 10
	FCGI_UNKNOWN_TYPE      uint8 = 11
	FCGI_MAXTYPE           uint8 = FCGI_UNKNOWN_TYPE
)

// Value for requestId component of FCGI_Header
const (
	FCGI_NULL_REQUEST_ID uint8 = 0
)

const (
	// Values for role component of FCGI_BeginRequestBody
	FCGI_RESPONDER  = 1
	FCGI_AUTHORIZER = 2
	FCGI_FILTER     = 3

	// Mask for flags component of FCGI_BeginRequestBody
	FCGI_KEEP_CONN = 1
)

type FastCGIRecord struct {
	Header *FastCGIHeader
}

type BeginRequestBody struct {
	RoleB1   uint8
	RoleB0   uint8
	Flag     uint8
	Reserved [5]uint8
}

type EndRequestBody struct {
	AppStatusB3    uint8
	AppStatusB2    uint8
	AppStatusB1    uint8
	AppStatusB0    uint8
	ProtocolStatus uint8
	Reserved       [3]uint8
}

func (erb *EndRequestBody) AppStatus() uint32 {
	return uint32(erb.AppStatusB3)<<24 +
		uint32(erb.AppStatusB2)<<16 +
		uint32(erb.AppStatusB1)<<8 +
		uint32(erb.AppStatusB0)
}

func (erb *EndRequestBody) WithAppStatus(status uint32) {
	erb.AppStatusB0 = uint8(status)
	erb.AppStatusB1 = uint8(status >> 8)
	erb.AppStatusB1 = uint8(status >> 16)
	erb.AppStatusB1 = uint8(status >> 24)
}

type UnknownTypeBody struct {
	Type     uint8
	Reserved [7]uint8
}

func (bqb *BeginRequestBody) Role() uint16 {
	return uint16(bqb.RoleB1)<<8 + uint16(bqb.RoleB0)
}

type FastCGIHeader struct {
	Version         uint8
	Type            uint8
	RequestIDB1     uint8
	RequestIDB0     uint8
	ContentLengthB1 uint8
	ContentLengthB0 uint8
	PaddingLength   uint8
	Reserved        uint8
}

func (hdr *FastCGIHeader) Bytes() []byte {
	return []byte{hdr.Type, hdr.Type, hdr.RequestIDB1, hdr.RequestIDB0, hdr.ContentLengthB1, hdr.ContentLengthB0, hdr.PaddingLength, hdr.Reserved}
}

func (hdr *FastCGIHeader) RequestID() uint16 {
	return uint16(hdr.RequestIDB1)<<8 + uint16(hdr.RequestIDB0)
}

func (hdr *FastCGIHeader) ContentLength() uint16 {
	return uint16(hdr.ContentLengthB1)<<8 + uint16(hdr.ContentLengthB0)
}

type NameValuePair struct {
	Name  string
	Value string
}

func UnmarshalNameValuePairs(data []byte) ([]*NameValuePair, error) {
	list := make([]*NameValuePair, 0)
	for len(data) > 0 {
		nvp := &NameValuePair{}
		err := nvp.UnmarshalBinary(data)
		if err != nil {
			return nil, err
		}
		nLen := len(nvp.Name)
		vLen := len(nvp.Value)
		nextPos := 2 + nLen + vLen
		if nLen > 127 {
			nextPos += 3
		}
		if vLen > 127 {
			nextPos += 3
		}
		data = data[nextPos:]
		list = append(list, nvp)
	}

	return list, nil
}

func (nvp *NameValuePair) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	var nvLen [2]int
	dataLen := len(data)
	pos := 0
	for i := 0; i < 2; i++ {
		if dataLen <= 0 {
			return fmt.Errorf("invalid name value pair data")
		} else if data[pos]&0x80 == 0 {
			nvLen[i] = int(data[pos])
			pos += 1
		} else if pos+4 <= dataLen {
			nvLen[i] = ((int(data[pos]) | 0x7F) << 24) + (int(data[pos+1]) << 16) + (int(data[pos+2]) << 8) + int(data[pos+3])
			pos += 4
		} else {
			return fmt.Errorf("invalid name value pair data")
		}
	}

	if nvLen[0]+nvLen[1] > len(data) {
		return fmt.Errorf("invalid name value pair data")
	}

	vPos := pos + nvLen[0]
	nvp.Name = string(data[pos : pos+nvLen[0]])
	nvp.Value = string(data[vPos : vPos+nvLen[1]])

	return nil
}

func (nvp *NameValuePair) MarshalBinary() ([]byte, error) {
	nvlen := make([]byte, 0, 8)
	if len(nvp.Name) <= 127 && len(nvp.Value) <= 127 {
		nvlen = append(
			nvlen,
			byte(len(nvp.Name)),  // nameLengthB0, nameLengthB0 >> 7 == 0
			byte(len(nvp.Value)), // valueLengthB0, valueLengthB0 >> 7 == 0
		)
	} else if len(nvp.Name) <= 127 && len(nvp.Value) > 127 {
		nvlen = append(
			nvlen,
			byte(len(nvp.Name)),           // nameLengthB0, nameLengthB0 >> 7 == 0
			byte(len(nvp.Value)>>24)|0x80, // valueLengthB3, valueLengthB3 >> 7 == 1
			byte(len(nvp.Value)>>16),      // valueLengthB2
			byte(len(nvp.Value)>>8),       // valueLengthB1
			byte(len(nvp.Value)),          // valueLengthB0
		)
	} else if len(nvp.Name) > 127 && len(nvp.Value) <= 127 {
		nvlen = append(
			nvlen,
			byte(len(nvp.Name)>>24)|0x80, // nameLengthB3, nameLengthB3 >> 7 == 1
			byte(len(nvp.Name)>>16),      // nameLengthB2
			byte(len(nvp.Name)>>8),       // nameLengthB1
			byte(len(nvp.Name)),          // nameLengthB0
			byte(len(nvp.Value)),         // valueLengthB0, valueLengthB0 >> 7 == 0
		)
	} else {
		nvlen = append(
			nvlen,
			byte(len(nvp.Name)>>24)|0x80,  // nameLengthB3, nameLengthB3  >> 7 == 1
			byte(len(nvp.Name)>>16),       // nameLengthB2
			byte(len(nvp.Name)>>8),        // nameLengthB1
			byte(len(nvp.Name)),           // nameLengthB0
			byte(len(nvp.Value)>>24)|0x80, // valueLengthB3, valueLengthB3 >> 7 == 1
			byte(len(nvp.Value)>>16),      // valueLengthB2
			byte(len(nvp.Value)>>8),       // valueLengthB1
			byte(len(nvp.Value)),          // valueLengthB0
		)
	}

	res := make([]byte, 0, len(nvlen)+len(nvp.Name)+len(nvp.Value))
	res = append(res, nvlen...)
	res = append(res, []byte(nvp.Name)...)
	res = append(res, []byte(nvp.Value)...)

	return res, nil
}
func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:9191")
	if err != nil {
		panic(err)
	}

	fpmConn, err := net.Dial("tcp", "127.0.0.1:9292")
	if err != nil {
		panic(err)
	}

	for {
		nginxConn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		fmt.Println("connection established")
		go handleRequest(nginxConn, fpmConn)
	}
}

func handleRequest(nginxConn net.Conn, fpmConn net.Conn) {
	data := make([]byte, 65536+8)

	nextStarts := 0
	v, err := nginxConn.Read(data)
	defer nginxConn.Close()
	if err != nil {
		return
	}
	for nextStarts+8 <= v {
		header, contentData, ns := parseRecord(data[0:v], nextStarts)
		nextStarts = ns

		fmt.Println(header.Type, header.RequestID(), header.ContentLength(), header.PaddingLength)
		if header.Type == FCGI_PARAMS {
			pairs, err := UnmarshalNameValuePairs(contentData)
			if err != nil {
				panic(err)
			}
			for _, each := range pairs {
				fmt.Printf("  %s: %s\n", string(each.Name), string(each.Value))
			}
		}
		if header.Type == FCGI_STDIN {
			if header.ContentLength() > 0 {
				fmt.Printf("  FCGI_STDIN: %s\n", string(contentData))
			} else {
				if d := proxyPass(data[0:v], fpmConn); d != nil {
					nginxConn.Write(d)
				}
			}
		}
	}
}

func proxyPass(data []byte, fpmConn net.Conn) []byte {
	receive := make([]byte, 65536+8)

	fmt.Print("FPM RESPONSE\n")
	nextStarts := 0
	_, err := fpmConn.Write(data)
	if err != nil {
		return nil
	}
	time.Sleep(time.Second)

	v, err := fpmConn.Read(receive)
	if err != nil {
		return nil
	}
	fmt.Println("read: ", v)
	for nextStarts+8 <= v {
		header, contentData, ns := parseRecord(receive[0:v], nextStarts)
		nextStarts = ns

		fmt.Println(header.Type, header.RequestID(), header.ContentLength(), header.PaddingLength)
		if header.Type == FCGI_PARAMS {
			pairs, err := UnmarshalNameValuePairs(contentData)
			if err != nil {
				panic(err)
			}
			for _, each := range pairs {
				fmt.Printf("  %s: %s\n", string(each.Name), string(each.Value))
			}
		}
		if header.Type == FCGI_STDERR || header.Type == FCGI_STDOUT {
			if header.ContentLength() > 0 {
				fmt.Printf("  FCGI_STDOUT: %s\n", string(contentData))
			} else {
				return nil
			}
		}
		if header.Type == FCGI_END_REQUEST {
			fmt.Println(contentData)
		}
	}

	return receive[0:v]
}

func parseRecord(data []byte, starts int) (*FastCGIHeader, []byte, int) {
	if starts+8 > len(data) {
		return nil, nil, starts
	}

	d := data[starts : starts+8]
	hdr := &FastCGIHeader{
		Version:         d[0],
		Type:            d[1],
		RequestIDB1:     d[2],
		RequestIDB0:     d[3],
		ContentLengthB1: d[4],
		ContentLengthB0: d[5],
		PaddingLength:   d[6],
		Reserved:        d[7],
	}
	return hdr, data[starts+8 : starts+8+int(hdr.ContentLength())], starts + 8 + int(hdr.ContentLength()) + int(hdr.PaddingLength)
}
