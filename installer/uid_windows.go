//go:build windows

package installer

func UnixIDsFS(sys fs.FS, path string) (uid int, gid int, err error) {
	return -1, -1, nil
}

func UnixIDs(path string) (uid int, gid int, err error) {
	return IDFS(os.DirFS(path), ".")
}

func RunAs(f func() error, uid, gid int) error {
	return f()
}
