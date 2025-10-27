package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var oldGroupKeys = map[string]string{
	"DcomUsers":          "DISTRIBUTED COM USERS",
	"LocalAdmins":        "ADMINISTRATORS",
	"PSRemoteUsers":      "REMOTE MANAGEMENT USERS",
	"RemoteDesktopUsers": "REMOTE DESKTOP USERS",
}

func main() {
	inputFile := flag.String("i", "", "Path to input SharpHound zip file")
	outputFile := flag.String("o", "", "Path to output SharpHound zip file (optional)")

	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Error: No input file provided")
		flag.Usage()
		os.Exit(1)
	}

	unfixedComputers, err := extractFromZip(*inputFile)
	if err != nil {
		panic(err)
	}
	fixedComputers, err := fixComputers(unfixedComputers)
	if err != nil {
		panic(err)
	}
	err = generateBHZip(fixedComputers, *inputFile, *outputFile)
	if err != nil {
		panic(err)
	}
}

func extractFromZip(zipFile string) ([]byte, error) {
	bhFiles, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer bhFiles.Close()

	var cptrJSON []byte
	for _, bhFile := range bhFiles.File {
		if strings.HasSuffix(bhFile.Name, "_computers.json") {
			cptrFile, err := bhFile.Open()
			if err != nil {
				return nil, err
			}
			cptrJSON, err = io.ReadAll(cptrFile)
			cptrFile.Close()
			if err != nil {
				return nil, err
			}
		}
	}
	return cptrJSON, nil
}

func generateBHZip(fixedComputersContent []byte, originalZipPath string, newZipPath string) error {
	bhFiles, err := zip.OpenReader(originalZipPath)
	if err != nil {
		return err
	}
	defer bhFiles.Close()

	if newZipPath == "" {
		newZipPath = strings.Replace(originalZipPath, ".zip", "_Compatible.zip", 1)
	} else if !strings.HasSuffix(newZipPath, ".zip") {
		newZipPath = newZipPath + ".zip"
	}
	newPath, err := os.Create(newZipPath)
	if err != nil {
		return err
	}
	newZip := zip.NewWriter(newPath)
	defer newZip.Close()
	for _, file := range bhFiles.File {
		cFile, err := newZip.Create(file.Name)
		if err != nil {
			return err
		}
		if strings.HasSuffix(file.Name, "_computers.json") {
			_, err = cFile.Write(fixedComputersContent)
			if err != nil {
				return err
			}
			continue
		}
		readStream, err := file.Open()
		if err != nil {
			return err
		}
		defer readStream.Close()
		if _, err := io.Copy(cFile, readStream); err != nil {
			readStream.Close()
			return err
		}
		if err := readStream.Close(); err != nil {
			return err
		}
	}
	fmt.Printf("[+] Write %s\n", newZipPath)
	return nil
}

func fixComputers(computerJSONContent []byte) ([]byte, error) {
	var jsonData map[string]any
	if err := json.Unmarshal(computerJSONContent, &jsonData); err != nil {
		return nil, err
	}

	computers, ok := jsonData["data"].([]any)
	if !ok {
		return nil, fmt.Errorf("data has no list")
	}

	for _, cptr := range computers {
		cptrObj, ok := cptr.(map[string]any)
		if !ok {
			continue
		}
		localGroups, ok := cptrObj["LocalGroups"].([]any)
		if !ok {
			return nil, fmt.Errorf("no localgroups found")
		}

		for oKey, oPattern := range oldGroupKeys {
			if err := fixLocalGroup(cptrObj, localGroups, oKey, oPattern); err != nil {
				return nil, err
			}
		}

		props := cptrObj["Properties"].(map[string]any)
		props["highvalue"] = isHighValue(cptrObj)

	}
	fixedJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return nil, err
	}
	return fixedJSON, nil
}

func fixLocalGroup(computerObj map[string]any, localGroups []any, oldKey string, groupPattern string) error {
	computerObj[oldKey] = map[string]any{
		"Collected":     true,
		"FailureReason": nil,
		"Results":       []any{},
	}
	cptrName, _ := computerObj["Properties"].(map[string]any)["name"].(string)
	for _, lgItem := range localGroups {
		lg, ok := lgItem.(map[string]any)
		if !ok {
			return fmt.Errorf("no localgroups")
		}
		if lg["Name"] == fmt.Sprintf("%s@%s", groupPattern, cptrName) {
			computerObj[oldKey] = map[string]any{
				"Collected":     lg["Collected"],
				"FailureReason": lg["FailureReason"],
				"Results":       lg["Results"],
			}
			break
		}
	}
	return nil
}

func isHighValue(cptrObj map[string]any) bool {
	if isDC, ok := cptrObj["IsDC"].(bool); ok && isDC {
		return true
	}
	aces, ok := cptrObj["Aces"].([]any)
	if ok {
		for _, aceItem := range aces {
			ace, ok := aceItem.(map[string]any)
			if !ok {
				continue
			}
			if rName, _ := ace["RightName"].(string); rName == "GenericAll" || rName == "Owns" || rName == "WriteDacl" || rName == "WriteOwner" {
				return true
			}
		}
	}
	if localGroup, ok := cptrObj["LocalGroups"].([]any); ok {
		for _, groupItem := range localGroup {
			lg, ok := groupItem.(map[string]any)
			if !ok {
				continue
			}
			if lgName, _ := lg["Name"].(string); lgName != "" {
				lname := strings.ToLower(lgName)
				if strings.Contains(lname, "domain admin") || strings.Contains(lname, "enterprise admin") || strings.Contains(lname, "administrator") {
					return true
				}
			}
		}
	}
	return false
}
