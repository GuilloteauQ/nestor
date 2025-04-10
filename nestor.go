package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
)

func createIfNotExist(folderName string) {
	_, err := os.Stat(folderName)
	if os.IsNotExist(err) {
		err = os.Mkdir(folderName, 0755)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Created folder '%s'\n", folderName)
		}
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

func store(result string, filenames []string) {
    fi, err := os.Lstat(result)
    if err != nil {
        log.Fatal(err)
    }
    if fi.Mode() & os.ModeSymlink != 0 {
        log.Fatal("Trying to store a symlink")
    }
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
	_, err = os.Stat(folderName)
	if os.IsNotExist(err) {
		if err := os.Mkdir(folderName, 0755); err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(folderName+"/info.json", jsonBytes, 0444); err != nil {
			log.Fatal(err)
		}
	}
	if err := os.Rename(result, folderName+"/data"); err != nil {
		log.Fatal(err)
	}
	if err := os.Chmod(folderName+"/data", 0444); err != nil {
		log.Fatal(err)
	}
	if err := os.Symlink(folderName+"/data", result); err != nil {
		log.Fatal(err)
	}
}

func importJSONInfo(filename string) map[string]interface{} {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	var payload map[string]interface{}
	err = json.Unmarshal(content, &payload)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
	return payload
}

func isSameInfo(filename string) bool {
	payload := importJSONInfo(".ne/store/" + filename + "/info.json")
	for key, value := range payload {
        hash := hashFile(key)
		if hex.EncodeToString(hash[:]) != value {
			return false
		}
	}
	return true
}

func check(filename string) bool {
	files, err := os.ReadDir(".ne/store")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.Name()[41:] == filename && isSameInfo(file.Name()) {
			return true
		}
	}
	return false
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		log.Fatalf("%s is not a regular file", src)
		return 0, err
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func get(filename string) {
	p, err := os.Readlink(filename)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(p)
	os.Remove(filename)
	copy(p, filename)
}

func main() {
	// TODO: Config file
	// TODO: find the store
	storeCmd := flag.NewFlagSet("store", flag.ExitOnError)
	var resultFlag = storeCmd.String("result", "foo", "help message for flag n")

	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
	var filename = checkCmd.String("file", "foo", "help message for flag n")

	createNeStore()

	switch os.Args[1] {
	case "store":
		storeCmd.Parse(os.Args[2:])
		store(*resultFlag, storeCmd.Args())
	case "get":
		get(os.Args[2])
	case "check":
		checkCmd.Parse(os.Args[2:])
		if check(*filename) {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	default:
		log.Fatalf("[ERROR] unknown subcommand '%s', see help for more details.", os.Args[1])
	}
	os.Exit(0)
}
