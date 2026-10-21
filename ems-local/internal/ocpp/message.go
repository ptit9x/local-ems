package ocpp

// OCPP 1.6 message types
const (
	MessageTypeCall       = 2 // Client → Server request
	MessageTypeCallResult = 3 // Server → Client response
	MessageTypeCallError  = 4 // Server → Client error
)

type OCPPMessage struct {
	MessageType int
	UniqueID    string
	Action      string      // only for Call
	Payload     interface{} // usually a map[string]interface{} or raw json before parsing
}

// Request/Response types for OCPP 1.6J actions
type BootNotificationReq struct {
	ChargePointVendor       string `json:"chargePointVendor"`
	ChargePointModel        string `json:"chargePointModel"`
	ChargePointSerialNumber string `json:"chargePointSerialNumber,omitempty"`
	FirmwareVersion         string `json:"firmwareVersion,omitempty"`
}

type BootNotificationConf struct {
	CurrentTime string `json:"currentTime"`
	Interval    int    `json:"interval"`
	Status      string `json:"status"` // "Accepted", "Pending", "Rejected"
}

type HeartbeatReq struct{}
type HeartbeatConf struct {
	CurrentTime string `json:"currentTime"`
}

type StatusNotificationReq struct {
	ConnectorId int    `json:"connectorId"`
	ErrorCode   string `json:"errorCode"`
	Status      string `json:"status"` // Available, Charging, SuspendedEV, etc.
}
type StatusNotificationConf struct{}

type MeterValuesReq struct {
	ConnectorId   int          `json:"connectorId"`
	TransactionId *int         `json:"transactionId,omitempty"`
	MeterValue    []MeterValue `json:"meterValue"`
}
type MeterValue struct {
	Timestamp    string         `json:"timestamp"`
	SampledValue []SampledValue `json:"sampledValue"`
}
type SampledValue struct {
	Value     string `json:"value"`
	Measurand string `json:"measurand,omitempty"`
	Unit      string `json:"unit,omitempty"`
	Phase     string `json:"phase,omitempty"`
}
type MeterValuesConf struct{}

type StartTransactionReq struct {
	ConnectorId int    `json:"connectorId"`
	IdTag       string `json:"idTag"`
	MeterStart  int    `json:"meterStart"`
	Timestamp   string `json:"timestamp"`
}
type StartTransactionConf struct {
	TransactionId int       `json:"transactionId"`
	IdTagInfo     IdTagInfo `json:"idTagInfo"`
}
type IdTagInfo struct {
	Status string `json:"status"` // "Accepted", "Blocked", "Invalid"
}

type StopTransactionReq struct {
	MeterStop     int    `json:"meterStop"`
	Timestamp     string `json:"timestamp"`
	TransactionId int    `json:"transactionId"`
	Reason        string `json:"reason,omitempty"`
}
type StopTransactionConf struct {
	IdTagInfo *IdTagInfo `json:"idTagInfo,omitempty"`
}
