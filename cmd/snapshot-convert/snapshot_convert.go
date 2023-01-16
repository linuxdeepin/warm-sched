package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/linuxdeepin/warm-sched/core"
)

var _opts struct {
	csv2gob bool
	gob2csv bool
}

func init() {
	flag.BoolVar(&_opts.csv2gob, "csv2gob", false, "csv to gob")
	flag.BoolVar(&_opts.gob2csv, "gob2csv", false, "gob to csv")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if _opts.csv2gob && len(args) == 2 {
		err := csv2Gob(args[0], args[1])
		if err != nil {
			log.Fatal(err)
		}

	} else if _opts.gob2csv && len(args) == 2 {
		err := gob2csv(args[0], args[1])
		if err != nil {
			log.Fatal(err)
		}

	} else {
		flag.Usage()
		os.Exit(2)
	}
}

func csv2Gob(srcFile, dstFile string) error {
	log.Println("csv2gob", srcFile, dstFile)

	fh, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	defer func() {
		err := fh.Close()
		if err != nil {
			log.Println("WARN:", err)
		}
	}()

	csvReader := csv.NewReader(fh)
	csvReader.FieldsPerRecord = 3

	var dstSnap core.Snapshot

	for {
		record, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		var info core.FileInfo
		info.Name = record[0]
		fileSize, err := strconv.ParseUint(record[1], 10, 64)
		if err != nil {
			return err
		}
		info.FileSize = fileSize
		info.Mapping, err = parseMapping(record[2])
		if err != nil {
			return err
		}
		dstSnap.Infos = append(dstSnap.Infos, info)
	}

	err = core.StoreTo(dstFile, &dstSnap)
	if err != nil {
		return err
	}

	return nil
}

// 二进制 gob 到文本 csv 格式
func gob2csv(srcFile, dstFile string) error {
	log.Println("gob2csv", srcFile, dstFile)
	var srcSnap core.Snapshot
	err := core.LoadFrom(srcFile, &srcSnap)
	if err != nil {
		return err
	}

	fh, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer func() {
		err = fh.Close()
		if err != nil {
			log.Println("WARN:", err)
		}
	}()

	csvWriter := csv.NewWriter(fh)

	for _, info := range srcSnap.Infos {
		fileSize := strconv.FormatUint(info.FileSize, 10)
		mapping := formatMapping(info.Mapping)
		err = csvWriter.Write([]string{info.Name, fileSize, mapping})
		if err != nil {
			return err
		}
	}

	csvWriter.Flush()

	// 检查错误
	err = csvWriter.Error()
	if err != nil {
		return err
	}

	return nil
}

func formatMapping(mapping []core.PageRange) string {
	slice := make([]string, 0, len(mapping))
	for _, item := range mapping {
		itemStr := fmt.Sprintf("%d:%d", item.Offset, item.Count)
		slice = append(slice, itemStr)
	}
	return strings.Join(slice, ";")
}

func parseMapping(str string) ([]core.PageRange, error) {
	fields := strings.Split(str, ";")
	result := make([]core.PageRange, 0, len(fields))
	for _, field := range fields {
		var pageRange core.PageRange
		_, err := fmt.Sscanf(field, "%d:%d", &pageRange.Offset, &pageRange.Count)
		if err != nil {
			return nil, err
		}
		result = append(result, pageRange)
	}
	return result, nil
}
