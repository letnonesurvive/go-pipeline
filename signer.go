package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

const bufLen = 100

func SingleHash(in, out chan interface{}) {

	for value := range in {
		data := strconv.Itoa(value.(int))

		crc32First := DataSignerCrc32(data)
		md5 := DataSignerMd5(data)
		crc32Second := DataSignerCrc32(md5)

		out <- crc32First + "~" + crc32Second
	}
}

func MultiHash(in, out chan interface{}) {

	for value := range in {
		data := value.(string)
		var res string
		for i := 0; i < 6; i++ {
			res += DataSignerCrc32(strconv.Itoa(i) + data)
		}
		out <- res
	}
}

func CombineResults(in, out chan interface{}) {

	var res string
	for value := range in {
		hashes := strings.Split(res, "_")
		if value.(string) <= hashes[0] || len(hashes[0]) == 0 {
			res = value.(string) + "_" + res
		} else if value.(string) >= hashes[len(hashes)-1] {
			res += "_" + value.(string)
		}
	}
	out <- res
}

func ExecutePipeline1(freeFlowJobs ...job) {

	tmp := make([]interface{}, 0, bufLen)

	wg := sync.WaitGroup{}
	for _, job := range freeFlowJobs {
		in := make(chan interface{}, bufLen)
		out := make(chan interface{}, bufLen)

		for _, value := range tmp {
			in <- value
		}
		tmp = tmp[:0] // clear the slice
		close(in)

		wg.Add(1)
		go func() {
			job(in, out)
			wg.Done()
		}()
		wg.Wait()

		close(out)
		for value := range out {
			tmp = append(tmp, value)
		}
	}
}

func ExecutePipeline(freeFlowJobs ...job) {

	wg := sync.WaitGroup{}

	in := make(chan interface{}, bufLen)
	out := make(chan interface{}, bufLen)

	for _, job := range freeFlowJobs {

		close(out)
		go func() {
			for value := range out {
				fmt.Println("Received ", value)
				in <- value
			}
			close(in)
		}()

		wg.Add(1)
		go func() {
			job(in, out)
			in = make(chan interface{}, bufLen)
			defer wg.Done()
		}()
		wg.Wait()

	}
}

func main() {

}

// сюда писать код
