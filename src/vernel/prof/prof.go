package prof

import . "vernel/types"
import "sync"
import "os"
import "fmt"

var profile bool = false
var work int
var cycles []map[VValue]struct{}
var threds []int
var plock sync.RWMutex

func StartProfile() {
	profile = true
	work = 0
	cycles = make([]map[VValue]struct{}, 0, 100)
	threds = make([]int, 0, 100)
}

func Clock(ctx *Tail) {
	if !profile {
		return
	}
	plock.Lock()
	time := ctx.Time
	for time >= len(cycles) {
		cycles = append(cycles, make(map[VValue]struct{}))
		threds = append(threds, 0)
	}
	work++
	threds[time]++
	cmap := cycles[time]
	cmap[ctx.Env] = struct{}{}
	cmap[ctx.Expr] = struct{}{}
	plock.Unlock()
	ctx.Time++
}

func GetTime() int {
	var time int
	plock.RLock()
	time = len(cycles)
	plock.RUnlock()
	return time
}

func calcMem() []int {
	time := len(cycles)
	mem := make([]int, time)
	for i, m := range cycles {
		seen := make(map[VValue]struct{})
		total := 0
		for k, _ := range m {
			if k == nil {
				total += 1
			} else {
				total += k.GetSize(seen)
			}
		}
		mem[i] = total
	}
	return mem
}

func WriteProfile(fname string) {
	file, err := os.Create(fname)
	if err != nil {
		panic("Error opening profile file")
	}
	defer func() {
		file.Close()
		profile = false
		threds = nil
		cycles = nil
	}()
	file.WriteString(fmt.Sprintf("Total Cycles: %d\nTotal Work: %d\nCycle\tThreads\tMemory\n", len(cycles), work))
	for i, m := range calcMem() {
		file.WriteString(fmt.Sprintf("%d:\t%d\t%d\n", i, threds[i], m))
	}
	profile = false
}
