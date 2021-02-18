package fieldtypes

import (
	"context"
	"fmt"
	"time"
)

type HappyLittleService interface {
	PaintTree(context.Context, *Request) (*Response, error)
}

type Request struct {
	Basic        string
	BasicPointer *string

	ExportedStruct        ExportedStruct
	ExportedStructPointer *ExportedStruct

	NotExportedStruct        notExportedStruct
	NotExportedStructPointer *notExportedStruct

	Time        time.Time
	TimePointer *time.Time

	Duration        time.Duration
	DurationPointer *time.Duration

	Interface interface{}
	Stringer  fmt.Stringer

	BasicSlice []string
	BasicMap   map[string]string
}

type Response struct {
}

type ExportedStruct struct {
	Name string
}

type notExportedStruct struct {
	name string
}
