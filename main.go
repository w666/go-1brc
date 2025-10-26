package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
)

type measurement struct {
	min   float64
	max   float64
	sum   float64
	count int
}

func getChunk(file os.File, chunkSize int, startIndex int) ([]byte, int, bool) {
	buffer := make([]byte, chunkSize)
	bytesRead, err := file.ReadAt(buffer, int64(startIndex))
	if err != nil && err != io.EOF {
		fmt.Println(`failed to read file`, err)
		panic("failed to read file")
	}

	if bytesRead == 0 {
		fmt.Println(`Zero bytes read, end of file`)
		return buffer, 0, true
	}

	lastNewLineIndex := bytes.LastIndexByte(buffer, '\n')
	return buffer, lastNewLineIndex, false
}

func checkAndUpdateMeasurements(mu *sync.Mutex, measurements map[string]measurement, tmpMeasurements map[string]measurement) {
	mu.Lock()
	defer mu.Unlock()

	for key, tmpValue := range tmpMeasurements {
		value, isPresent := measurements[key]
		if isPresent {
			if value.min > tmpValue.max {
				value.min = tmpValue.min
			}
			if value.max < tmpValue.max {
				value.max = tmpValue.max
			}

			value.count = +tmpValue.count
			value.sum = +tmpValue.sum

			measurements[key] = value
		} else {
			measurements[key] = tmpValue
		}
	}
}

var newLineByte = []byte("\n")
var separator = []byte(";")

func parseFloatCustom(buffer []byte) float64 {
	index := 0
	isNegative := false
	var parsedValue float64

	if buffer[index] == '-' {
		isNegative = true
		index++
	}

	parsedValue = float64(buffer[index] - '0')
	index++

	if buffer[index] == '.' {
		index++
	} else {
		parsedValue = parsedValue*10 + float64(buffer[index]-'0')
		index++
	}

	// only one decimal digit
	parsedValue += float64(buffer[index]-'0') / 10

	if isNegative {
		parsedValue = -parsedValue
	}
	return parsedValue
}

func processChunk(variant int, sem chan int, wg *sync.WaitGroup, mu *sync.Mutex, buffer []byte, lastIndex int, measurements map[string]measurement) {
	defer wg.Done()
	defer func() { <-sem }()

	tmpMeasurements := make(map[string]measurement)

	if variant == 1 {

		startIndex := 0
		stopIndex := 0

		for {
			stopIndex = bytes.Index(buffer[startIndex:], newLineByte)

			if stopIndex == -1 {
				break
			}

			stopIndex = stopIndex + startIndex

			parts := bytes.Split(buffer[startIndex:stopIndex], separator)

			startIndex = stopIndex + 1

			if len(parts) < 2 {
				fmt.Println("last line empty")
				break
			}

			parsedValue := parseFloatCustom(parts[1])

			key := string(parts[0])
			value, isPresent := tmpMeasurements[key]

			if isPresent {
				if value.min > parsedValue {
					value.min = parsedValue
				} else if value.max < parsedValue {
					value.max = parsedValue
				}

				value.sum = value.sum + parsedValue
				value.count = value.count + 1

				tmpMeasurements[key] = value
			} else {
				tmpMeasurements[key] = measurement{
					min:   parsedValue,
					max:   parsedValue,
					sum:   parsedValue,
					count: 1,
				}
			}

		}
	}

	checkAndUpdateMeasurements(mu, measurements, tmpMeasurements)

}

func main() {

	f, err := os.Create("profiler")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	chunkSize := 16 * 1024 * 1024

	sem := make(chan int, runtime.NumCPU())

	measurements := make(map[string]measurement)

	file, err := os.Open("measurements.txt")
	if err != nil {
		fmt.Println("failed to open file", err)
		return
	}

	defer file.Close()

	lastIndex := 0
	skip := 0
	var buffer []byte
	var isEOF bool
	var wg = sync.WaitGroup{}
	var mu sync.Mutex

	for {
		sem <- 1
		skip += lastIndex + 1
		buffer, lastIndex, isEOF = getChunk(*file, chunkSize, skip)
		wg.Add(1)
		go processChunk(1, sem, &wg, &mu, buffer, lastIndex, measurements)
		if isEOF {
			break
		}
	}

	wg.Wait()

	isFirst := true
	for key, value := range measurements {
		if isFirst {
			fmt.Printf("%s=%v/%v/%v", key, value.min, value.sum/float64(value.count), value.max)
			isFirst = false
		} else {
			fmt.Printf(", %s=%v/%v/%v", key, value.min, math.Round(value.sum/float64(value.count)*10)/10, value.max)
		}
	}

}
