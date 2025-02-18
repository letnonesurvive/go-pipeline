package main

import (
	"strconv"
)

const bufLen = 100

func SingleHash(in, out chan interface{}) {
	dataRaw := (<-in).(int)
	data := strconv.Itoa(dataRaw)

	crc32First := DataSignerCrc32(data)
	md5 := DataSignerMd5(data)
	crc32Second := DataSignerCrc32(md5)

	out <- crc32First + "~" + crc32Second
}

func ExecutePipeline(freeFlowJobs ...job) {

	tmp := make([]interface{}, 0, bufLen)

	for _, job := range freeFlowJobs {
		in := make(chan interface{}, bufLen)
		out := make(chan interface{}, bufLen)
		for _, value := range tmp {
			in <- value
		}
		tmp = tmp[:0] // clear the slice

		close(in)
		job(in, out)
		close(out)

		for value := range out {
			tmp = append(tmp, value)
		}
	}
}

func main() {

}

// сюда писать код
