package bosinit

const bashProfile = `
if [[ -f "${HOME}/.bashrc" ]]; then
  source "${HOME}/.bashrc"
fi
`

// BashProfile provides a better bash profile with color prompt.
var BashProfile = &WriteFile{
	Path:        "/home/rancher/.bash_profile",
	Permissions: FilePerm(0644),
	Owner:       "rancher",
	Content:     bashProfile,
}
