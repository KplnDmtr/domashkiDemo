package main

import (

	// json "encoding/json"

	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	easy "gitlab.com/vk-golang/lectures/12_reflect/99_hw/optimization/generation"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	isAndroid := false
	isMSIE := false
	i := 0
	builder := strings.Builder{}
	builder.Grow(5000)
	builder.WriteString("found users:\n")
	seenBrowsers := make(map[string]bool, 500)
	uniqueBrowsers := 0

	user := easy.User{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		err = user.UnmarshalJSON(scanner.Bytes())
		if err != nil {
			panic(err)
		}

		isAndroid = false
		isMSIE = false
		for _, browser := range user.Browsers {
			if ok := strings.Contains(browser, "Android"); ok {
				isAndroid = true
				if _, ok := seenBrowsers[browser]; !ok {
					seenBrowsers[browser] = true
					uniqueBrowsers++
				}
			}
			if ok := strings.Contains(browser, "MSIE"); ok {
				isMSIE = true
				if _, ok := seenBrowsers[browser]; !ok {
					seenBrowsers[browser] = true
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			i++
			continue
		}
		builder.WriteString("[" + strconv.Itoa(i) + "] " + user.Name + " <" + strings.ReplaceAll(user.Email, "@", " [at] ") + ">\n")
		i++
	}
	builder.WriteString("\nTotal unique browsers ")
	builder.WriteString(strconv.Itoa(uniqueBrowsers))

	fmt.Fprintln(out, builder.String())
}
