package archivetest

import (
	"archive/tar"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTar(t *testing.T) {
	data, err := NewTar(map[string]any{
		"dir": map[string]any{
			"file1": "# contents of dir/file1\n",
			"file2": []byte("# contents of dir/file2\n"),
			"subdir": &Dir{
				Perm: 0o700,
				Entries: map[string]any{
					"file": &File{
						Perm:     0o777,
						Contents: []byte("# contents of dir/subdir/file\n"),
					},
					"symlink": &Symlink{
						Target: "file",
					},
				},
			},
		},
	})
	require.NoError(t, err)

	tarReader := tar.NewReader(bytes.NewBuffer(data))

	header, err := tarReader.Next()
	require.NoError(t, err)
	assert.Equal(t, byte(tar.TypeDir), header.Typeflag)
	assert.Equal(t, "dir/", header.Name)
	assert.Equal(t, int64(0o777), header.Mode)

	header, err = tarReader.Next()
	require.NoError(t, err)
	assert.Equal(t, byte(tar.TypeReg), header.Typeflag)
	assert.Equal(t, "dir/file1", header.Name)
	assert.Equal(t, int64(len("# contents of dir/file1\n")), header.Size)
	assert.Equal(t, int64(0o666), header.Mode)

	header, err = tarReader.Next()
	require.NoError(t, err)
	assert.Equal(t, byte(tar.TypeReg), header.Typeflag)
	assert.Equal(t, "dir/file2", header.Name)
	assert.Equal(t, int64(len("# contents of dir/file2\n")), header.Size)
	assert.Equal(t, int64(0o666), header.Mode)

	header, err = tarReader.Next()
	require.NoError(t, err)
	assert.Equal(t, byte(tar.TypeDir), header.Typeflag)
	assert.Equal(t, "dir/subdir/", header.Name)
	assert.Equal(t, int64(0o700), header.Mode)

	header, err = tarReader.Next()
	require.NoError(t, err)
	assert.Equal(t, byte(tar.TypeReg), header.Typeflag)
	assert.Equal(t, "dir/subdir/file", header.Name)
	assert.Equal(t, int64(len("# contents of dir/subdir/file\n")), header.Size)
	assert.Equal(t, 0o777, int(header.Mode))

	header, err = tarReader.Next()
	require.NoError(t, err)
	assert.Equal(t, byte(tar.TypeSymlink), header.Typeflag)
	assert.Equal(t, "dir/subdir/symlink", header.Name)
	assert.Equal(t, "file", header.Linkname)
}
