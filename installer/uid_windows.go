//go:build windows

package installer

func IDFS(sys fs.FS, path string) (int, int, error) {
	return -1, -1, nil
}

func ID(path string) (int, int, error) {
	return IDFS(os.DirFS(path), ".")
}

func RunAs(f func() error, uid, gid int) error {
	return f()
}
