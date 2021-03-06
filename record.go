package GoFastCGI

import (
	"fmt"
)

type RecordType uint8

var recordTypeNames = []string{
	"",
	"FCGI_BEGIN_REQUEST",
	"FCGI_ABORT_REQUEST",
	"FCGI_END_REQUEST",
	"FCGI_PARAMS",
	"FCGI_STDIN",
	"FCGI_STDOUT",
	"FCGI_STDERR",
	"FCGI_DATA",
	"FCGI_GET_VALUES",
	"FCGI_GET_VALUES_RESULT",
	"FCGI_UNKNOWN_TYPE",
	"FCGI_MAXTYPE",
}

func (rt RecordType) String() string {
	if int(rt) >= len(recordTypeNames) {
		return ""
	}
	return recordTypeNames[rt]
}

const (
	// Values for type component of FCGI_Header
	TypeBeginRequest    RecordType = 1
	TypeAbortRequest    RecordType = 2
	TypeEndRequest      RecordType = 3
	TypeParams          RecordType = 4
	TypeSTDIN           RecordType = 5
	TypeSTDOUT          RecordType = 6
	TypeSTDERR          RecordType = 7
	TypeData            RecordType = 8
	TypeGetValues       RecordType = 9
	TypeGetValuesResult RecordType = 10
	TypeUnknownType     RecordType = 11
	TypeMaxType         RecordType = TypeUnknownType

	// Value for version component of FCGI_Header
	Version1 uint8 = 1

	// Value for requestId component of FCGI_Header
	NullRequestID uint16 = 0

	// Values for role component of FCGI_BeginRequestBody
	RoleResponse   uint16 = 1
	RoleAuthorizer uint16 = 2
	RoleFilter     uint16 = 3

	// Mask for flags component of FCGI_BeginRequestBody
	FlagKeepConn uint8 = 1

	// Values for protocolStatus component of FCGI_EndRequestBody
	StatusRequestComplete uint8 = 0
	StatusCantMPXConn     uint8 = 1
	StatusOverloaded      uint8 = 2
	StatusUnknownRole     uint8 = 3

	// Variable names for FCGI_GET_VALUES / FCGI_GET_VALUES_RESULT records
	VarMaxConns  = "FCGI_MAX_CONNS"
	VarMaxReqs   = "FCGI_MAX_REQS"
	VarMPXSConns = "FCGI_MPXS_CONNS"
)

type Session struct {
	buff      []byte
	MaxConns  string
	MaxReqs   string
	MPXSConns string
	KeepConn  bool
}

type FCGIRecord interface {
	Header() *Header
	Content() []byte
}

type Header struct {
	Version         uint8
	Type            RecordType
	RequestIDB1     uint8
	RequestIDB0     uint8
	ContentLengthB1 uint8
	ContentLengthB0 uint8
	PaddingLength   uint8
	Reserved        uint8
}

func NewHeader(rType RecordType, requestID uint16) *Header {
	h := &Header{
		Version: Version1,
		Type:    rType,
	}
	h.WithRequestID(requestID)
	return h
}

func (hdr *Header) Bytes() []byte {
	return []byte{hdr.Version, byte(hdr.Type), hdr.RequestIDB1, hdr.RequestIDB0, hdr.ContentLengthB1, hdr.ContentLengthB0, hdr.PaddingLength, hdr.Reserved}
}

func (hdr *Header) WithRequestID(id uint16) {
	hdr.RequestIDB0 = uint8(id)
	hdr.RequestIDB1 = uint8(id >> 8)
}

func (hdr *Header) RequestID() uint16 {
	return uint16(hdr.RequestIDB1)<<8 + uint16(hdr.RequestIDB0)
}

func (hdr *Header) WithContentLength(l uint16) {
	hdr.ContentLengthB0 = uint8(l)
	hdr.ContentLengthB1 = uint8(l >> 8)
}

func (hdr *Header) ContentLength() uint16 {
	return uint16(hdr.ContentLengthB1)<<8 + uint16(hdr.ContentLengthB0)
}

type BeginRequestBody struct {
	RoleB1   uint8
	RoleB0   uint8
	Flag     uint8
	Reserved [5]uint8
}

func (bqb *BeginRequestBody) Role() uint16 {
	return uint16(bqb.RoleB1)<<8 + uint16(bqb.RoleB0)
}

func (bqb *BeginRequestBody) WithRole(r uint16) {
	bqb.RoleB0 = uint8(r)
	bqb.RoleB1 = uint8(r >> 8)
}

type BeginRequestRecord struct {
	header *Header
	body   *BeginRequestBody
}

func NewBeginRequestRecord(requestID uint16, body *BeginRequestBody) *BeginRequestRecord {
	header := NewHeader(TypeBeginRequest, requestID)
	header.WithContentLength(8)
	return &BeginRequestRecord{
		header: header,
		body:   body,
	}
}

func (brr *BeginRequestRecord) Header() *Header {
	return brr.header
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

type EndRequestRecord struct {
	Header *Header
	Body   *EndRequestBody
}

func NewEndRequestRecord(requestID uint16, body *EndRequestBody) *EndRequestRecord {
	header := NewHeader(TypeEndRequest, requestID)
	header.WithContentLength(8)
	return &EndRequestRecord{
		Header: header,
		Body:   body,
	}
}

type AbortRequestRecord struct {
	Header *Header
}

func NewAbortRequestRecord(requestID uint16) *AbortRequestRecord {
	return &AbortRequestRecord{
		Header: NewHeader(TypeAbortRequest, requestID),
	}
}

type ParamsRecord struct {
	Header         *Header
	nameValuePairs []*NameValuePair
}

func NewParamsRecord(requestID uint16) *ParamsRecord {
	return &ParamsRecord{
		Header:         NewHeader(TypeParams, requestID),
		nameValuePairs: make([]*NameValuePair, 0),
	}
}

func (pr *ParamsRecord) AddNameValuePair(nvp *NameValuePair) bool {
	contentLen := pr.Header.ContentLength()
	newContentLen := contentLen + nvp.Length()
	if newContentLen < contentLen {
		return false
	}
	pr.nameValuePairs = append(pr.nameValuePairs, nvp)
	pr.Header.WithContentLength(newContentLen)
	return true
}

func (pr *ParamsRecord) NameValuePairs() []*NameValuePair {
	return pr.nameValuePairs
}

type STDINRecord struct {
	Header *Header
	data   []byte
}

func NewSTDINRecord(requestID uint16) *STDINRecord {
	return &STDINRecord{
		Header: NewHeader(TypeSTDIN, requestID),
	}
}

func (sir *STDINRecord) AppendData(d []byte) int {
	data, nBytes := appendData(sir.data, d)
	sir.data = data
	return nBytes
}

type STDOUTRecord struct {
	Header *Header
	data   []byte
}

func NewSTDOUTRecord(requestID uint16) *STDOUTRecord {
	return &STDOUTRecord{
		Header: NewHeader(TypeSTDOUT, requestID),
	}
}

func (sor *STDOUTRecord) AppendData(d []byte) int {
	data, nBytes := appendData(sor.data, d)
	sor.data = data
	return nBytes
}

type STDERRRecord struct {
	Header *Header
	data   []byte
}

func NewSTDERRRecord(requestID uint16) *STDERRRecord {
	return &STDERRRecord{
		Header: NewHeader(TypeSTDOUT, requestID),
	}
}

func (ser *STDERRRecord) AppendData(d []byte) int {
	data, nBytes := appendData(ser.data, d)
	ser.data = data
	return nBytes
}

type DataRecord struct {
	Header *Header
	data   []byte
}

func NewDataRecord(requestID uint16) *DataRecord {
	return &DataRecord{
		Header: NewHeader(TypeSTDOUT, requestID),
	}
}

func (dr *DataRecord) AppendData(d []byte) int {
	data, nBytes := appendData(dr.data, d)
	dr.data = data
	return nBytes
}

type GetValuesRecord struct {
	Header         *Header
	nameValuePairs []*NameValuePair
}

func NewGetValuesRecord(requestID uint16) *GetValuesRecord {
	return &GetValuesRecord{
		Header:         NewHeader(TypeParams, requestID),
		nameValuePairs: make([]*NameValuePair, 0),
	}
}

func (gvr *GetValuesRecord) AddNameValuePair(nvp *NameValuePair) bool {
	contentLen := gvr.Header.ContentLength()
	newContentLen := contentLen + nvp.Length()
	if newContentLen < contentLen {
		return false
	}
	gvr.nameValuePairs = append(gvr.nameValuePairs, nvp)
	gvr.Header.WithContentLength(newContentLen)
	return true
}

func (gvr *GetValuesRecord) NameValuePairs() []*NameValuePair {
	return gvr.nameValuePairs
}

type GetValuesResultRecord struct {
	Header         *Header
	nameValuePairs []*NameValuePair
}

func NewGetValuesResultRecord(requestID uint16) *GetValuesResultRecord {
	return &GetValuesResultRecord{
		Header:         NewHeader(TypeParams, requestID),
		nameValuePairs: make([]*NameValuePair, 0),
	}
}

func (gvr *GetValuesResultRecord) AddNameValuePair(nvp *NameValuePair) bool {
	contentLen := gvr.Header.ContentLength()
	newContentLen := contentLen + nvp.Length()
	if newContentLen < contentLen {
		return false
	}
	gvr.nameValuePairs = append(gvr.nameValuePairs, nvp)
	gvr.Header.WithContentLength(newContentLen)
	return true
}

func (gvr *GetValuesResultRecord) NameValuePairs() []*NameValuePair {
	return gvr.nameValuePairs
}

type UnknownTypeBody struct {
	Type     uint8
	Reserved [7]uint8
}

type UnknownTypeRecord struct {
	Header *Header
	Body   *UnknownTypeBody
}

func NewUnknownTypeRecord(body *UnknownTypeBody) UnknownTypeRecord {
	return UnknownTypeRecord{
		Header: NewHeader(TypeUnknownType, NullRequestID),
		Body:   body,
	}
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
		list = append(list, nvp)
		data = data[nvp.Length():]
	}

	return list, nil
}

func (nvp *NameValuePair) Length() uint16 {
	nLen := len(nvp.Name)
	vLen := len(nvp.Value)
	length := 2 + nLen + vLen
	if nLen > 127 {
		length += 3
	}
	if vLen > 127 {
		length += 3
	}
	return uint16(length)
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

func appendData(dst, src []byte) ([]byte, int) {
	remain := 65535 - len(dst)
	srcLen := len(src)
	if remain <= 0 {
		return dst, 0
	}
	if remain < srcLen {
		dst = append(dst, src[:remain]...)
		return dst, remain
	}
	dst = append(dst, src...)
	return dst, srcLen
}
