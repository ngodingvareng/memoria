package enum

type DataExportStatus string

const (
	DataExportStatusPending DataExportStatus = "pending"
	DataExportStatusRunning DataExportStatus = "running"
	DataExportStatusReady   DataExportStatus = "ready"
	DataExportStatusFailed  DataExportStatus = "failed"
	DataExportStatusExpired DataExportStatus = "expired"
)
