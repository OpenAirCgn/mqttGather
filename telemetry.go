package mqttGather

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// TODO should prbly include date here and in dbaStats ...
type Telemetry struct {
	Client string
	Type   Type
	Data   interface{}
}

type Type string

const (
	ESP = Type("esp")
	FRT = Type("frt")
	FLG = Type("flg")
	RST = Type("rst")
	VER = Type("ver")
	PRJ = Type("prj")
	TME = Type("tme")
	IDF = Type("idf")
	CHP = Type("chp")
	ESQ = Type("esq")
)

// see https://github.com/OpenAirCgn/openair-firmware-espidf/blob/feature-telemetry/notes/feature-telemetry.md
func (t Type) String() string {
	switch t {
	case "esp":
		return "ESP free heap"
	case "frt":
		return "FreeRTOS free heap"
	case "flg":
		return "Telemetry Flags"
	case "rst":
		return "ESP Reset reason"
	case "ver":
		return "App Version"
	case "prj":
		return "Project Name"
	case "tme":
		return "Compile time"
	case "idf":
		return "IDF version"
	case "chp":
		return "Chip revision info"
	case "esq":
		return "extended signal quality"
	default:
		return fmt.Sprintf("unknown %s", string(t))
	}
}

func (t Type) IsMemory() bool              { return t == ESP || t == FRT }
func (t *Telemetry) IsMemory() bool        { return t.Type.IsMemory() }
func (t Type) IsFlag() bool                { return t == FLG }
func (t *Telemetry) IsFlag() bool          { return t.Type.IsFlag() }
func (t Type) IsSignalQuality() bool       { return t == ESQ }
func (t *Telemetry) IsSignalQuality() bool { return t.Type.IsSignalQuality() }
func (t Type) IsResetReason() bool         { return t == RST }
func (t *Telemetry) IsResetReason() bool   { return t.Type.IsResetReason() }
func (t Type) IsVersion() bool             { return t == VER || t == PRJ || t == IDF || t == CHP }
func (t *Telemetry) IsVersion() bool       { return t.Type.IsVersion() }
func (t Type) IsChip() bool                { return t == CHP }
func (t *Telemetry) IsChip() bool          { return t.Type.IsChip() }

func parseTelemetryData(t Type, data string) interface{} {
	switch t {

	case "esp":
		fallthrough
	case "frt":
		fallthrough
	case "rst": // todo - reset reason semantics
		if i, err := strconv.Atoi(data); err != nil {
			log.Printf("E: valid number in telemetry >%s< : %s", t, data)
			return -1
		} else {
			return i
		}
	case "flg":
		if i, err := strconv.ParseInt(data, 16, 32); err != nil {
			log.Printf("E: valid number in telemetry >flg< : %s", data)
			return -1
		} else {
			return int(i)
		}

	case "esq": // todo - sig qual. semantics
		fallthrough
	case "ver":
		fallthrough
	case "prj":
		fallthrough
	case "tme":
		fallthrough
	case "idf":
		fallthrough
	case "chp": // todo - chip rev semantics
		fallthrough
	default:
		return data
	}
}

func TelemetryFromPayload(payload string, client string) (*Telemetry, error) {
	vals := strings.SplitN(payload, ":", 2)
	if len(vals) != 2 {
		return nil, fmt.Errorf("invalid telemetry: %s", payload)
	}

	tipe := Type(vals[0])
	data := parseTelemetryData(tipe, vals[1])

	tel := Telemetry{
		Client: client,
		Type:   Type(vals[0]),
		Data:   data,
	}

	return &tel, nil
}
