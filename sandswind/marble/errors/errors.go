package errors

const (
	EC_SUCC = 0000
	EC_FAIL = 9999

	EC_PARAM      = 1101
	EC_MAPNULL    = 1102
	EC_STRUCTNULL = 1103
	EC_SLICENULL  = 1104

	EC_DBCONN   = 1201
	EC_DBADD    = 1202
	EC_DBSET    = 1203
	EC_DBGET    = 1204
	EC_DBDEL    = 1205
	EC_DBNODATA = 1206
	EC_DATAFMT  = 1207

	EC_IPADDR = 1301
	EC_HOST   = 1302
	EC_NET    = 1304

	EC_NOMOD = 2101
)

type Error interface {
	error
	GetCode() int32
	GetMessage() string
}

type SysError struct {
	Code    int32
	Message string
	ExMsg   string
}

func (err *SysError) Error() string {
	return err.Message
}

func (err *SysError) GetCode() int32 {
	return err.Code
}

func (err *SysError) GetMessage() string {
	return err.Message
}

func (err *SysError) GetExMessage() string {
	return err.ExMsg
}

var (
	ETSucc       = &SysError{EC_SUCC, "success", ""}
	ETFail       = &SysError{EC_FAIL, "failure", ""}
	ETParam      = &SysError{EC_PARAM, "parameter is error", ""}
	ETDBconn     = &SysError{EC_DBCONN, "connecting to db is error", ""}
	ETDBadd      = &SysError{EC_DBADD, "db_conflict", ""}
	ETDBset      = &SysError{EC_DBSET, "internal_error", ""}
	ETDBget      = &SysError{EC_DBGET, "data_invalid", ""}
	ETDBdel      = &SysError{EC_DBDEL, "addr_error", ""}
	ETDBnodata   = &SysError{EC_DBNODATA, "addr_error", ""}
	ETDatafmt    = &SysError{EC_DATAFMT, "net_error", ""}
	ETDataNil    = &SysError{EC_DATAFMT, "net_error", ""}
	ETIpaddr     = &SysError{EC_IPADDR, "net_error", ""}
	ETHost       = &SysError{EC_HOST, "net_error", ""}
	ETNet        = &SysError{EC_IPADDR, "net_error", ""}
	ETMapNil     = &SysError{EC_MAPNULL, "map_nil", ""}
	ETStructNull = &SysError{EC_STRUCTNULL, "struct_nil", ""}
	ETSliceNull  = &SysError{EC_SLICENULL, "slice_nil", ""}
	ETNoMod      = &SysError{EC_NOMOD, "no_mod", ""}
)
