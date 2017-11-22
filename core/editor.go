package core

import (
	"os"
)

// Editor is the main struct for this command.
type Editor struct {
	ui     UI
	buffer *Buffer
	line   int
	cursor *Position
}

// NewEditor creates a new editor.
func NewEditor(ui UI) *Editor {
	return &Editor{ui: ui, cursor: &Position{}}
}

// Init initializes the editor.
func (e *Editor) Init() error {
	ch := make(chan Event)
	if err := e.ui.Init(ch); err != nil {
		return err
	}
	go func() {
		for {
			select {
			case c := <-ch:
				switch c {
				case CursorUp:
					e.cursorUp()
				case CursorDown:
					e.cursorDown()
				case CursorLeft:
					e.cursorLeft()
				case CursorRight:
					e.cursorRight()
				case ScrollUp:
					e.scrollUp()
				case ScrollDown:
					e.scrollDown()
				case PageUp:
					e.pageUp()
				case PageDown:
					e.pageDown()
				case PageTop:
					e.pageTop()
				case PageLast:
					e.pageLast()
				}
			}
		}
	}()
	return nil
}

// Close terminates the editor.
func (e *Editor) Close() error {
	return e.ui.Close()
}

// Open opens a new file.
func (e *Editor) Open(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	e.buffer = NewBuffer(file)
	return nil
}

// Start starts the editor.
func (e *Editor) Start() error {
	if err := e.redraw(); err != nil {
		return err
	}
	return e.ui.Start()
}

func (e *Editor) cursorUp() error {
	e.cursor.Up()
	if e.cursor.X < 0 {
		e.cursor.Down()
		if e.line > 0 {
			e.line = e.line - 1
		}
	}
	return e.redraw()
}

func (e *Editor) cursorDown() error {
	e.cursor.Down()
	if e.cursor.X >= e.ui.Height() {
		return e.scrollDown()
	}
	return e.redraw()
}

func (e *Editor) cursorLeft() error {
	e.cursor.Left()
	if e.cursor.Y < 0 {
		e.cursor.Right()
	}
	return e.redraw()
}

func (e *Editor) cursorRight() error {
	e.cursor.Right()
	if e.cursor.Y >= 16 {
		e.cursor.Left()
	}
	return e.redraw()
}

func (e *Editor) scrollUp() error {
	e.cursor.Down()
	if e.cursor.X >= e.ui.Height() {
		e.cursor.Up()
	}
	if e.line > 0 {
		e.line = e.line - 1
	}
	return e.redraw()
}

func (e *Editor) scrollDown() error {
	e.cursor.Up()
	if e.cursor.X < 0 {
		e.cursor.Down()
	}
	line, err := e.lastLine()
	if err != nil {
		return err
	}
	e.line = e.line + 1
	if e.line > line {
		e.line = line
	}
	return e.redraw()
}

func (e *Editor) pageUp() error {
	e.line = e.line - e.ui.Height() + 2
	if e.line < 0 {
		e.line = 0
	}
	return e.redraw()
}

func (e *Editor) pageDown() error {
	line, err := e.lastLine()
	if err != nil {
		return err
	}
	e.line = e.line + e.ui.Height() - 2
	if e.line > line {
		e.line = line
	}
	return e.redraw()
}

func (e *Editor) pageTop() error {
	e.line = 0
	return e.redraw()
}

func (e *Editor) pageLast() error {
	line, err := e.lastLine()
	if err != nil {
		return err
	}
	e.line = line
	return e.redraw()
}

func (e *Editor) lastLine() (int, error) {
	len, err := e.buffer.Len()
	if err != nil {
		return 0, err
	}
	width := 16
	line := int((len+int64(width)-1)/int64(width)) - e.ui.Height()
	if line < 0 {
		line = 0
	}
	return line, nil
}

func (e *Editor) redraw() error {
	height, width := e.ui.Height(), 16
	b := make([]byte, height*width)
	n, err := e.buffer.Read(int64(e.line)*int64(width), b)
	if err != nil {
		return err
	}
	return e.ui.Redraw(State{Line: e.line, Cursor: *e.cursor, Bytes: b, Size: n})
}
