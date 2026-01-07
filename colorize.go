package sloghuman

import (
	"net/http"
	"strings"
)

type (
	// ColorPalette is used to create a palette for the colorizer.
	// This can be used to create a custom palette to be used by the logger.
	ColorPalette struct {
		MethodGET     string
		MethodPOST    string
		MethodPUT     string
		MethodDELETE  string
		MethodPATCH   string
		MethodOPTIONS string
		MethodHEAD    string
		MethodTRACE   string
		MethodCONNECT string

		Status2xx string
		Status3xx string
		Status4xx string
		Status5xx string

		LevelDEBUG string
		LevelINFO  string
		LevelWARN  string
		LevelERROR string

		RequestID string
		Path      string
		Line      string
		LogType   string
		Message   string
		Reset     string
	}

	// ColorType is defines the type of colors to be used when calling colorize
	ColorType int
)

// Below are the predefined color palettes that can be used.
// Set Colors = PaletteName to use them.
var (
	// Dracula
	Dracula = ColorPalette{
		MethodGET:     "\033[38;5;84m",
		MethodPOST:    "\033[38;5;159m",
		MethodPUT:     "\033[38;5;222m",
		MethodDELETE:  "\033[38;5;203m",
		MethodPATCH:   "\033[38;5;222m",
		MethodHEAD:    "\033[38;5;84m",
		MethodOPTIONS: "\033[38;5;141m",
		MethodTRACE:   "\033[38;5;248m",
		MethodCONNECT: "\033[38;5;248m",

		Status2xx: "\033[38;5;84m",
		Status3xx: "\033[38;5;159m",
		Status4xx: "\033[38;5;222m",
		Status5xx: "\033[38;5;203m",

		LevelDEBUG: "\033[38;5;159m",
		LevelINFO:  "\033[38;5;84m",
		LevelWARN:  "\033[38;5;222m",
		LevelERROR: "\033[38;5;203m",

		RequestID: "\033[38;5;212m",
		Path:      "\033[38;5;159m",
		Line:      "\033[38;5;159m",
		LogType:   "\033[38;5;141m",
		Message:   "\033[38;5;141m",
		Reset:     "\033[0m",
	}

	// Nord
	Nord = ColorPalette{
		MethodGET:     "\033[38;5;14m",
		MethodPOST:    "\033[38;5;12m",
		MethodPUT:     "\033[38;5;11m",
		MethodDELETE:  "\033[38;5;9m",
		MethodPATCH:   "\033[38;5;11m",
		MethodHEAD:    "\033[38;5;14m",
		MethodOPTIONS: "\033[38;5;13m",
		MethodTRACE:   "\033[38;5;8m",
		MethodCONNECT: "\033[38;5;8m",

		Status2xx: "\033[38;5;10m",
		Status3xx: "\033[38;5;14m",
		Status4xx: "\033[38;5;11m",
		Status5xx: "\033[38;5;9m",

		LevelDEBUG: "\033[38;5;14m",
		LevelINFO:  "\033[38;5;10m",
		LevelWARN:  "\033[38;5;11m",
		LevelERROR: "\033[38;5;9m",

		RequestID: "\033[38;5;13m",
		Path:      "\033[38;5;12m",
		Line:      "\033[38;5;8m",
		LogType:   "\033[38;5;6m",
		Message:   "\033[38;5;6m",
		Reset:     "\033[0m",
	}

	// Gruvbox Dark
	GruvboxDark = ColorPalette{
		MethodGET:     "\033[38;5;142m",
		MethodPOST:    "\033[38;5;108m",
		MethodPUT:     "\033[38;5;172m",
		MethodDELETE:  "\033[38;5;167m",
		MethodPATCH:   "\033[38;5;172m",
		MethodHEAD:    "\033[38;5;142m",
		MethodOPTIONS: "\033[38;5;175m",
		MethodTRACE:   "\033[38;5;248m",
		MethodCONNECT: "\033[38;5;248m",

		Status2xx: "\033[38;5;142m",
		Status3xx: "\033[38;5;108m",
		Status4xx: "\033[38;5;172m",
		Status5xx: "\033[38;5;167m",

		LevelDEBUG: "\033[38;5;108m",
		LevelINFO:  "\033[38;5;142m",
		LevelWARN:  "\033[38;5;172m",
		LevelERROR: "\033[38;5;167m",

		RequestID: "\033[38;5;175m",
		Path:      "\033[38;5;109m",
		Line:      "\033[38;5;248m",
		LogType:   "\033[38;5;214m",
		Message:   "\033[38;5;214m",
		Reset:     "\033[0m",
	}

	// Solarized Dark
	SolarizedDark = ColorPalette{
		MethodGET:     "\033[38;5;64m",
		MethodPOST:    "\033[38;5;37m",
		MethodPUT:     "\033[38;5;136m",
		MethodDELETE:  "\033[38;5;160m",
		MethodPATCH:   "\033[38;5;136m",
		MethodHEAD:    "\033[38;5;64m",
		MethodOPTIONS: "\033[38;5;136m",
		MethodTRACE:   "\033[38;5;244m",
		MethodCONNECT: "\033[38;5;244m",

		Status2xx: "\033[38;5;64m",
		Status3xx: "\033[38;5;37m",
		Status4xx: "\033[38;5;166m",
		Status5xx: "\033[38;5;160m",

		LevelDEBUG: "\033[38;5;37m",
		LevelINFO:  "\033[38;5;33m",
		LevelWARN:  "\033[38;5;166m",
		LevelERROR: "\033[38;5;160m",

		RequestID: "\033[38;5;136m",
		Path:      "\033[38;5;33m",
		Line:      "\033[38;5;244m",
		LogType:   "\033[38;5;125m",
		Message:   "\033[38;5;125m",
		Reset:     "\033[0m",
	}

	// One Dark
	OneDark = ColorPalette{
		MethodGET:     "\033[38;5;114m",
		MethodPOST:    "\033[38;5;81m",
		MethodPUT:     "\033[38;5;222m",
		MethodDELETE:  "\033[38;5;204m",
		MethodPATCH:   "\033[38;5;222m",
		MethodHEAD:    "\033[38;5;114m",
		MethodOPTIONS: "\033[38;5;176m",
		MethodTRACE:   "\033[38;5;145m",
		MethodCONNECT: "\033[38;5;145m",

		Status2xx: "\033[38;5;114m",
		Status3xx: "\033[38;5;81m",
		Status4xx: "\033[38;5;222m",
		Status5xx: "\033[38;5;204m",

		LevelDEBUG: "\033[38;5;81m",
		LevelINFO:  "\033[38;5;114m",
		LevelWARN:  "\033[38;5;222m",
		LevelERROR: "\033[38;5;204m",

		RequestID: "\033[38;5;176m",
		Path:      "\033[38;5;81m",
		Line:      "\033[38;5;145m",
		LogType:   "\033[38;5;176m",
		Message:   "\033[38;5;176m",
		Reset:     "\033[0m",
	}

	// Colors sets the palette to be used by colorize.
	// This can be changed before creating a new logger by either
	// creating your own color palette or using a pre-existing palette
	Colors ColorPalette = Dracula
)

// Enums for setting colors
const (
	ColorNone ColorType = iota
	ColorMethod
	ColorStatus
	ColorLevel
	ColorRequestID
	ColorPath
	ColorLine
	ColorLogType
	ColorMessage
)

// colorize sets the colors for the provided string using the given ColorType
func (h *TextHandler) colorize(value string, t ColorType) string {
	if h.noColor {
		return value
	} else if value == "" {
		return value
	}

	switch t {
	case ColorMethod:
		switch strings.TrimSpace(value) {
		case http.MethodGet:
			return Colors.MethodGET + value + Colors.Reset
		case http.MethodPost:
			return Colors.MethodPOST + value + Colors.Reset
		case http.MethodPut:
			return Colors.MethodPUT + value + Colors.Reset
		case http.MethodDelete:
			return Colors.MethodDELETE + value + Colors.Reset
		case http.MethodPatch:
			return Colors.MethodPATCH + value + Colors.Reset
		case http.MethodHead:
			return Colors.MethodHEAD + value + Colors.Reset
		case http.MethodOptions:
			return Colors.MethodOPTIONS + value + Colors.Reset
		case http.MethodConnect:
			return Colors.MethodCONNECT + value + Colors.Reset
		case http.MethodTrace:
			return Colors.MethodTRACE + value + Colors.Reset
		default:
			return value
		}
	case ColorStatus:
		switch {
		case strings.HasPrefix(value, "2"):
			return Colors.Status2xx + value + Colors.Reset
		case strings.HasPrefix(value, "3"):
			return Colors.Status3xx + value + Colors.Reset
		case strings.HasPrefix(value, "4"):
			return Colors.Status4xx + value + Colors.Reset
		case strings.HasPrefix(value, "5"):
			return Colors.Status5xx + value + Colors.Reset
		default:
			return value
		}
	case ColorLevel:
		switch strings.TrimSpace(value) {
		case "DEBUG":
			return Colors.LevelDEBUG + value + Colors.Reset
		case "INFO":
			return Colors.LevelINFO + value + Colors.Reset
		case "WARN":
			return Colors.LevelWARN + value + Colors.Reset
		case "ERROR":
			return Colors.LevelERROR + value + Colors.Reset
		default:
			return value
		}
	case ColorLine:
		return Colors.Line + value + Colors.Reset
	case ColorRequestID:
		return Colors.RequestID + value + Colors.Reset
	case ColorPath:
		return Colors.Path + value + Colors.Reset
	case ColorLogType:
		return Colors.LogType + value + Colors.Reset
	case ColorMessage:
		return Colors.Message + value + Colors.Reset
	default:
		return value
	}
}
