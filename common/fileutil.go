package common

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func SlurpAsLines(filename string) []string {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return lines
}

func SlurpAsString(filename string) string {
	b := SlurpAsBytes(filename)
	str := string(b)
	return str
}

func SlurpAsBytes(filename string) []byte {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return b
}

func WriteStrings(lines []string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func WriteString(line string, filename string) error {
	return WriteStrings([]string{line}, filename)
}

func FileExists (filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func DirExists (filename string) bool {
	f, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	filemode := f.Mode()
	return filemode.IsDir()
}

func Which(filename string) string {
	path , err := exec.LookPath(filename)
	if err == nil {
		return path
	}
	return ""
}

func ExecExists(filename string) bool {
	_ , err := exec.LookPath(filename)
	return err == nil
}

