package dockerignore

import (
	"bufio"
	"io"
)

func Parse(r io.Reader) (ignores []string, err error) {
	scan := bufio.NewScanner(r)

	for scan.Scan() {
		line := scan.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		ignores = append(ignores, line)
	}

	err = scan.Err()
	return
}
