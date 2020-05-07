package fastcgi

// Value for version component of FCGI_Header
const (
	FCGI_VERSION_1 uint8 = 1
)

// Values for type component of FCGI_Header
const (
	FCGI_BEGIN_REQUEST uint8 = 1
	FCGI_ABORT_REQUEST uint8 = 2
	FCGI_END_REQUEST uint8 = 3
	FCGI_PARAMS uint8 = 4
	FCGI_STDIN uint8 = 5
	FCGI_STDOUT uint8 = 6
	FCGI_STDERR uint8 = 7
	FCGI_DATA uint8 = 8
	FCGI_GET_VALUES uint8 = 9
	FCGI_GET_VALUES_RESULT uint8 = 10
	FCGI_UNKNOWN_TYPE uint8 = 11
	FCGI_MAXTYPE uint8 = FCGI_UNKNOWN_TYPE
)

// Value for requestId component of FCGI_Header
const (
	FCGI_NULL_REQUEST_ID uint8 = 0
)