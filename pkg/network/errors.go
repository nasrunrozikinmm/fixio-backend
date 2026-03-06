package network

// Common error message constants
const (
	ErrInvalidPagination = "Parameter pagination tidak valid"
	ErrInvalidID         = "ID tidak valid"
	ErrInvalidRequest    = "Format request tidak valid"
	ErrInvalidBody       = "Body request tidak valid"
	ErrRecordNotFound    = "Data tidak ditemukan"
	ErrUnauthorized      = "Anda belum login"
	ErrForbidden         = "Anda tidak memiliki akses"
	ErrInternalServer    = "Terjadi kesalahan internal server"
	ErrDuplicateEntry    = "Data sudah ada"
)
