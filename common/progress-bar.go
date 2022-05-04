package common

import (
	"fmt"
	"math"
	"strings"
)

type progressBar struct {
	max          int64
	current      int64
	length       int64
	fillChar     string
	filledLength int64
	curPercent   int
	placeholder  *string
}

func (pb *progressBar) Max() int64 {
	return pb.max
}

func (pb *progressBar) UpdatePercent(percentDone int, filledLength int64) {
	pb.curPercent = percentDone
	pb.filledLength = filledLength
	pb.Print()
}

func (pb *progressBar) UpdateCurrent(n int64) {
	pb.current = n
	pb.curPercent = int(math.Round(float64(100 * pb.current / pb.max)))
	pb.filledLength = pb.length * pb.current / pb.max
	pb.Print()
}

func (pb *progressBar) AddCurrent(n int64) {
	pb.UpdateCurrent(pb.current + n)
}

func (pb *progressBar) Print() {
	var bar string
	for i := 0; i < int(pb.filledLength); i++ {
		bar += pb.fillChar
	}
	emptyBar := strings.Repeat(" ", int(pb.length-pb.filledLength))
	fmt.Printf("\r%s%s%s %d%%", *(pb.placeholder), bar, emptyBar, pb.curPercent)
}

func NewProgressBar(max, length int64, fillChar, placeholder string) *progressBar {
	return &progressBar{
		max: max, length: length, fillChar: fillChar, placeholder: &placeholder,
	}
}
