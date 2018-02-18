package main

import (
	"bufio"
	"fmt"
	"github.com/datacharmer/dbdeployer/common"
	"os"
	"sort"
)

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
		vl := common.VersionToList(line)
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
