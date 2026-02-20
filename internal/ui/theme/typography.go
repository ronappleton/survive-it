package theme

type Typography struct {
	Title      int32
	Header     int32
	Body       int32
	Small      int32
	Log        int32
	LineFactor float32
}

var Type = Typography{
	Title:      34,
	Header:     23,
	Body:       20,
	Small:      17,
	Log:        18,
	LineFactor: 1.58,
}
