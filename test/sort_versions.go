package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
)

func VersionToList(version string) []int {
	// A valid version must be made of 3 integers
	re1 := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
	// Also valid version is 3 numbers with a prefix
	re2 := regexp.MustCompile(`^[^.0-9-]+(\d+)\.(\d+)\.(\d+)$`)
	verList1 := re1.FindAllStringSubmatch(version, -1)
	verList2 := re2.FindAllStringSubmatch(version, -1)
	verList := verList1
	//fmt.Printf("%#v\n", verList)
	if verList == nil {
		verList = verList2
	}
	if verList == nil {
		return []int{-1}
	}

	major, err1 := strconv.Atoi(verList[0][1])
	minor, err2 := strconv.Atoi(verList[0][2])
	rev, err3 := strconv.Atoi(verList[0][3])
	if err1 != nil || err2 != nil || err3 != nil {
		return []int{-1}
	}
	return []int{major, minor, rev}
}

/*
	This utility reads a list of versions (format x.x.xx)
	and sorts them in numerical order, taking into account
	all three fields, making sure that 5.7.9 comes before 5.7.10.
*/
func main() {

	scanner := bufio.NewScanner(os.Stdin)

	type version_list struct {
		text string
		mmr  []int
	}
	var vlist []version_list
	for scanner.Scan() {
		line := scanner.Text()
		vl := VersionToList(line)
		rec := version_list{
			text: line,
			mmr:  vl,
		}
		if vl[0] > 0 {
			vlist = append(vlist, rec)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	sort.Slice(vlist, func(i, j int) bool {
		return vlist[i].mmr[0] < vlist[j].mmr[0] ||
			(vlist[i].mmr[0] == vlist[j].mmr[0] && vlist[i].mmr[1] < vlist[j].mmr[1]) ||
			(vlist[i].mmr[0] == vlist[j].mmr[0] && vlist[i].mmr[1] == vlist[j].mmr[1] && vlist[i].mmr[2] < vlist[j].mmr[2])
	})
	for _, v := range vlist {
		fmt.Printf("%s\n", v.text)
	}
}
