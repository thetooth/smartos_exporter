package collector

import (
	"testing"
)

func TestParseMdbOutput(t *testing.T) {
	c, _ := NewGZFreeMemExporter()
	in := `
Page Summary                Pages                MB  %Tot
------------     ----------------  ----------------  ----
Kernel                    1624148              6344   10%
Boot pages                  75307               294    0%
ZFS File Data             4754787             18573   28%
VMM Memory                1335296              5216    8%
Anon                       107189               418    1%
Exec and libs               21626                84    0%
Page cache                  18824                73    0%
Free (cachelist)            56868               222    0%
Free (freelist)           8762424             34228   52%

Total                    16756469             65454
Physical                 16756468             65454
`
	err := c.parseMdbOutput(in)
	if err != nil {
		t.Error(err)
	}
}
