package hw10programoptimization

import (
	"archive/zip"
	"testing"

	"github.com/stretchr/testify/require"
)

// go test -bench=BenchmarkGetDomainStat -benchmem -count=10 > old/new.txt.
func BenchmarkGetDomainStat(b *testing.B) {
	r, err := zip.OpenReader("testdata/users.dat.zip")
	require.NoError(b, err)
	defer r.Close()

	require.Equal(b, 1, len(r.File))

	zf := r.File[0]

	for i := 0; i < b.N; i++ {
		rc, err := zf.Open()
		require.NoError(b, err)

		b.StartTimer()
		_, err = GetDomainStat(rc, "biz")
		b.StopTimer()
		require.NoError(b, err)

		rc.Close()
	}
}
