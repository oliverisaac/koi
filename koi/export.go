package koi

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func ExportCommand(input io.Reader, output io.Writer) (exitCode int, runError error) {
	inputContent, err := io.ReadAll(input)
	if err != nil {
		return 1, fmt.Errorf("failed to read input: %w", err)
	}

	if string(inputContent) == "" {
		return 1, fmt.Errorf("no input")
	}

	inputObject := make(map[string]interface{})

	err = yaml.Unmarshal(inputContent, &inputObject)
	if err != nil {
		return 1, fmt.Errorf("failed to unmarshal input: %w", err)
	}

	fields_to_remove := []string{
		"metadata.ownerReferences",
		"metadata.resourceVersion",
		"metadata.selfLink",
		"metadata.uid",
		"metadata.generation",
		"metadata.creationTimestamp",
		"metadata.managedFields",
		"metadata.generateName",
		"status",
		"spec.nodeName",
	}

	for _, field := range fields_to_remove {
		deletePathIfExists(inputObject, strings.Split(field, ".")...)
		deletePathIfExists(inputObject, strings.Split("items.[]."+field, ".")...)
	}

	encoder := yaml.NewEncoder(output)
	encoder.SetIndent(2)
	defer encoder.Close()
	err = encoder.Encode(inputObject)
	if err != nil {
		return 1, fmt.Errorf("failed to encode output: %w", err)
	}
	return 0, nil
}

func deletePathIfExists(inputObject interface{}, path ...string) {
	if len(path) == 0 {
		return
	}
	currKey := path[0]
	switch inputObject.(type) {
	case map[string]interface{}:
		asMap := inputObject.(map[string]interface{})
		if len(path) == 1 {
			delete(asMap, currKey)
		} else if next, ok := asMap[currKey]; ok {
			if len(path) > 2 {
				deletePathIfExists(next, path[1:]...)
			} else {
				nextKey := path[1]
				switch next.(type) {
				case map[string]interface{}:
					deletePathIfExists(asMap[path[0]], path[1:]...)
				case []interface{}:
					if nextKey == "[]" {
						delete(asMap, currKey)
					} else {
						nextInt, err := strconv.Atoi(nextKey)
						if err != nil {
							return
						}
						nextSlice := next.([]interface{})
						if nextInt >= 0 && nextInt < len(nextSlice) {
							nextSlice = append(nextSlice[:nextInt], nextSlice[nextInt+1:]...)
						}
					}
				}
			}
		}
	case []interface{}:
		asSice := inputObject.([]interface{})
		currKeyInt, err := strconv.Atoi(currKey)
		if err != nil {
			currKeyInt = -1
		}
		if len(path) == 1 && currKeyInt >= 0 && currKeyInt < len(asSice) {
			asSice = append(asSice[:currKeyInt], asSice[currKeyInt+1:]...)
			return
		} else if currKeyInt >= 0 && currKeyInt < len(asSice) {
			deletePathIfExists(asSice[currKeyInt], path[1:]...)
		} else if currKey == "[]" {
			for _, v := range asSice {
				deletePathIfExists(v, path[1:]...)
			}
		}
	}
}
