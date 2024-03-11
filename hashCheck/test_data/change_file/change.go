package change_file

type Change interface {
	Me() (string, error)
}
