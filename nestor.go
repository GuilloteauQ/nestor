package main

import (
	"bufio"
    "crypto/sha1"
	"flag"
	"encoding/json"
	"encoding/hex"
    "log"
	"io"
    "os"
	"sort"
)

func createIfNotExist(folderName string) {
    _, err := os.Stat(folderName)
    if os.IsNotExist(err) {
        err = os.Mkdir(folderName, 0755)
        if err != nil { log.Fatal(err) } else { log.Printf("Created folder '%s'\n", folderName) }
    }
}

func createNeStore() {
    createIfNotExist(".ne")
    createIfNotExist(".ne/store")
}

func hashFile(filename string) [20]byte {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	bs := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(bs)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	return sha1.Sum(bs)
}
func main() {
	// TODO: check / set CLI subcommands 
	// set: store the file in the store
	// check: (maybe find another name) check if the file is in the store, if so check the hashes of the deps, return some exit code based on the result

	var resultFlag = flag.String("result", "foo", "help message for flag n")
	flag.Parse()
    filenames := flag.Args()
	log.Println("result: ", *resultFlag)
	log.Println("filenames: ", filenames)
    createNeStore()

	result := *resultFlag
	sort.Slice(filenames, func(i, j int) bool {
		return filenames[i] < filenames[j]
	})
    deps := make(map[string]string)
	for _, filename := range filenames {
		hash := hashFile(filename)
		deps[filename] = hex.EncodeToString(hash[:])
	}
	jsonBytes, err := json.MarshalIndent(deps, "", "   ")
    if err != nil {
        log.Fatal(err)
    }
	finalHash := sha1.Sum(jsonBytes)
	folderName := ".ne/store/" + hex.EncodeToString(finalHash[:]) + "-" + result
	createIfNotExist(folderName)
	if err := os.WriteFile(folderName + "/info.json", jsonBytes, 0444); err != nil {
		log.Fatal(err)
	}
	if err := os.Rename(result, folderName + "/data") ; err != nil { 
        log.Fatal(err) 
    } 
	if err := os.Chmod(folderName + "/data", 0444); err != nil {
        log.Fatal(err) 
    } 
	if err := os.Symlink(folderName + "/data", result); err != nil { 
        log.Fatal(err) 
    } 
}
