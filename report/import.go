package report

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/dundee/gdu/v5/pkg/analyze"
)

// ReadAnalysis reads analysis report from JSON file and returns directory item
func ReadAnalysis(input io.Reader) (*analyze.Dir, error) {
	var data interface{}

	var buff bytes.Buffer
	buff.ReadFrom(input)
	json.Unmarshal(buff.Bytes(), &data)

	dataArray, ok := data.([]interface{})
	if !ok {
		return nil, errors.New("JSON file does not contain top level array")
	}
	if len(dataArray) < 4 {
		return nil, errors.New("Top level array must have at least 4 items")
	}

	items, ok := dataArray[3].([]interface{})
	if !ok {
		return nil, errors.New("Array of maps not found in the top level array on 4th position")
	}

	return processDir(items)
}

func processDir(items []interface{}) (*analyze.Dir, error) {
	dir := &analyze.Dir{
		File: &analyze.File{
			Flag: ' ',
		},
	}
	dirMap, ok := items[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("Directory item is not a map")
	}
	name, ok := dirMap["name"].(string)
	if !ok {
		return nil, errors.New("Directory name is not a string")
	}

	slashPos := strings.LastIndex(name, "/")
	if slashPos > -1 {
		dir.Name = name[slashPos+1:]
		dir.BasePath = name[:slashPos+1]
	} else {
		dir.Name = name
	}

	for _, v := range items[1:] {
		switch item := v.(type) {
		case map[string]interface{}:
			file := &analyze.File{}
			file.Name = item["name"].(string)
			file.Size = int64(item["asize"].(float64))
			file.Usage = int64(item["dsize"].(float64))
			file.Parent = dir
			file.Flag = ' '

			dir.Files.Append(file)
		case []interface{}:
			subdir, err := processDir(item)
			if err != nil {
				return nil, err
			}
			subdir.Parent = dir
			dir.Files.Append(subdir)
		}
	}

	return dir, nil
}
